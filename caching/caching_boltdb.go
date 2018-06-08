package caching

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/boltdb/bolt"
	"github.com/cloudfoundry-community/firehose-to-syslog/logging"
	cfclient "github.com/cloudfoundry-community/go-cfclient"
	uuid "github.com/satori/go.uuid"
)

var (
	APP_BUCKET = []byte("AppBucketV2")
)

type entity struct {
	Name             string                 `json:"name"`
	SpaceGUID        string                 `json:"space_guid"`
	OrganizationGUID string                 `json:"organization_guid"`
	Environment      map[string]interface{} `json:"environment_json"`
	TTL              time.Time
}

func (e *entity) appIsOptOut() bool {
	return e.Environment["F2S_DISABLE_LOGGING"] == "true"
}

type CachingBoltConfig struct {
	Path               string
	IgnoreMissingApps  bool
	CacheInvalidateTTL time.Duration
}

type CachingBolt struct {
	client *cfclient.Client
	appdb  *bolt.DB

	config *CachingBoltConfig
}

func NewCachingBolt(client *cfclient.Client, config *CachingBoltConfig) (*CachingBolt, error) {
	return &CachingBolt{
		client: client,
		config: config,
	}, nil
}

func (c *CachingBolt) Open() error {
	// Open bolt db
	db, err := bolt.Open(c.config.Path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		logging.LogError("Fail to open boltdb: ", err)
		return err
	}
	c.appdb = db

	err = c.appdb.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(APP_BUCKET)
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
	if err != nil {
		logging.LogError("Fail to create bucket: ", err)
		return err
	}

	return nil
}

func (c *CachingBolt) Close() error {
	return c.appdb.Close()
}

var (
	errNotFound = errors.New("not found")
)

// entityType *must* be checked for safety by caller
// guid will be validated as a guid by this function
func (c *CachingBolt) getEntity(entityType, guid string) (*entity, error) {
	// First verify the GUID is in fact that - else we could become a confused deputy due to path construction issues
	u, err := uuid.FromString(guid)
	if err != nil {
		return nil, err
	}
	uuid := u.String()
	key := []byte(fmt.Sprintf("%s/%s", entityType, uuid))

	// Check if we have it already
	var rv entity
	err = c.appdb.View(func(tx *bolt.Tx) error {
		v := tx.Bucket(APP_BUCKET).Get(key)
		if len(v) == 0 {
			return errNotFound
		}
		return gob.NewDecoder(bytes.NewReader(v)).Decode(&rv)
	})
	switch err {
	case nil:
		if rv.TTL.Before(time.Now()) {
			return &rv, nil
		}
		// else continue
	case errNotFound:
		// continue
	default:
		return nil, err
	}

	// Fetch from remote
	nv, err := c.fetchEntityFromAPI(entityType, uuid)
	if err != nil {
		if entityType == "apps" && c.config.IgnoreMissingApps {
			nv = &entity{}
		} else {
			return nil, err
		}
	}

	// Set TTL to value between 75% and 125% of desired amount. This is to spread out cache invalidations
	nv.TTL = time.Now().Add(time.Duration(float64(c.config.CacheInvalidateTTL.Nanoseconds()) * (0.75 + (rand.Float64() / 2.0))))
	b := &bytes.Buffer{}
	err = gob.NewEncoder(b).Encode(nv)
	if err != nil {
		return nil, err
	}

	// Write to DB
	err = c.appdb.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(APP_BUCKET).Put(key, b.Bytes())
	})
	if err != nil {
		return nil, err
	}

	return nv, nil
}

// both entityType and guid must have been validated by the caller
func (c *CachingBolt) fetchEntityFromAPI(entityType, guid string) (*entity, error) {
	resp, err := c.client.DoRequestWithoutRedirects(c.client.NewRequest(http.MethodGet, fmt.Sprintf("/v2/%s/%s", entityType, guid)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status code: %s", resp.Status)
	}

	var md struct {
		Entity entity `json:"entity"`
	}
	err = json.NewDecoder(resp.Body).Decode(&md)
	if err != nil {
		return nil, err
	}

	return &md.Entity, nil
}

func (c *CachingBolt) GetApp(appGuid string) (*App, error) {
	app, err := c.getEntity("apps", appGuid)
	if err != nil {
		if c.config.IgnoreMissingApps {
			app = &entity{}
		} else {
			return nil, err
		}
	}

	space, err := c.getEntity("spaces", app.SpaceGUID)
	if err != nil {
		if c.config.IgnoreMissingApps {
			space = &entity{}
		} else {
			return nil, err
		}
	}

	org, err := c.getEntity("organizations", space.OrganizationGUID)
	if err != nil {
		if c.config.IgnoreMissingApps {
			org = &entity{}
		} else {
			return nil, err
		}
	}

	return &App{
		Guid:       appGuid,
		Name:       app.Name,
		SpaceGuid:  app.SpaceGUID,
		SpaceName:  space.Name,
		OrgGuid:    space.OrganizationGUID,
		OrgName:    org.Name,
		IgnoredApp: app.appIsOptOut(),
	}, nil
}
