package caching

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	uuid "github.com/satori/go.uuid"
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

type CacheLazyFillConfig struct {
	// IgnoreMissingApps no error if an app can't be found
	IgnoreMissingApps bool

	// CacheInvalidateTTL is the approx TTL for cached data. Code will randomly pick between 0.75x and 1.2
	CacheInvalidateTTL time.Duration
	StripAppSuffixes   []string
}

type CacheLazyFill struct {
	client CFSimpleClient
	store  CacheStore

	config *CacheLazyFillConfig
}

func NewCacheLazyFill(client CFSimpleClient, store CacheStore, config *CacheLazyFillConfig) *CacheLazyFill {
	return &CacheLazyFill{
		client: client,
		store:  store,
		config: config,
	}
}

// FillCache communicates with the server to enumerate *all* applications and fills
// the cache
func (c *CacheLazyFill) FillCache() error {
	allApps, err := c.fetchEntityListFromAPI("apps")
	if err != nil {
		return err
	}

	// Populate bolt with all apps
	for guid, app := range allApps {
		// Canonicalise the guid
		u, err := uuid.FromString(guid)
		if err != nil {
			return err
		}
		uuid := u.String()

		// Save our app out
		err = c.normaliseAndSaveEntityToDatabase("apps", uuid, app)
		if err != nil {
			return err
		}

		// Fetch and poulate space and org
		_, _, err = c.getSpaceAndOrg(app.SpaceGUID)
		if err != nil {
			return err
		}
	}

	return nil
}

// getEntity looks up the entity in the cache, and if not found, or TTL expired, fetches from the server
// entityType *must* be checked for safety by caller
// guid will be validated as a guid by this function
// apps are treated specially, in that if IgnoreMissingApps is set, then an error will result in an empty object returned.
// Also for apps, we will strip anything that matches StripAppSuffixes from the name before storing.
func (c *CacheLazyFill) getEntity(entityType, guid string) (*entity, error) {
	// Canonicalise guid
	u, err := uuid.FromString(guid)
	if err != nil {
		return nil, err
	}
	uuid := u.String()

	// Check if we have it already
	var rv entity
	err = c.store.Get(makeCacheStorageKey(entityType, uuid), &rv)
	switch err {
	case nil:
		if rv.TTL.After(time.Now()) {
			return &rv, nil
		}
		// else continue
	case ErrKeyNotFound:
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

	// Save it out
	err = c.normaliseAndSaveEntityToDatabase(entityType, uuid, nv)
	if err != nil {
		return nil, err
	}

	return nv, nil
}

// makeCacheStorageKey returns the key for the entity in the bolt bucket
// entityType is "apps" or "spaces" or "orgs" - caller must validate
// uuid  must be validated by caller
func makeCacheStorageKey(entityType, uuid string) string {
	return fmt.Sprintf("%s/%s", entityType, uuid)
}

// normaliseAndSaveEntityToDatabase saves the entity to the cache, stripping app name prefixes if enabled
// entityType is "apps" or "spaces" or "orgs" - caller must validate
// uuid  must be validated by caller
// nv may be modified by this function
func (c *CacheLazyFill) normaliseAndSaveEntityToDatabase(entityType, uuid string, nv *entity) error {
	// Strip name suffixes if applicable. This is intended for blue green deployments,
	// so that things like -venerable can be stripped from renamed apps
	if entityType == "apps" {
		for _, suffix := range c.config.StripAppSuffixes {
			if strings.HasSuffix(nv.Name, suffix) {
				nv.Name = nv.Name[:len(nv.Name)-len(suffix)]
				break
			}
		}
	}

	// Set TTL to value between 75% and 125% of desired amount. This is to spread out cache invalidations
	nv.TTL = time.Now().Add(time.Duration(float64(c.config.CacheInvalidateTTL.Nanoseconds()) * (0.75 + (rand.Float64() / 2.0))))

	// Write to DB
	return c.store.Set(makeCacheStorageKey(entityType, uuid), nv)
}

// fetchEntityListFromAPI fetches a full list of all such entities from the server
// entityType must have been validated by the caller
func (c *CacheLazyFill) fetchEntityListFromAPI(entityType string) (map[string]*entity, error) {
	url := fmt.Sprintf("/v2/%s?results-per-page=100", entityType)
	rv := make(map[string]*entity)
	for {
		var md struct {
			NextURL   string `json:"next_url"`
			Resources []*struct {
				Metadata struct {
					GUID string `json:"guid"`
				} `json:"metadata"`
				Entity *entity `json:"entity"`
			} `json:"resources"`
		}
		err := c.makeRequestAndDecodeJSON(url, &md)
		if err != nil {
			return nil, err
		}

		for _, r := range md.Resources {
			rv[r.Metadata.GUID] = r.Entity
		}

		if md.NextURL == "" {
			// we're done!
			return rv, nil
		}

		url = md.NextURL
	}
}

func (c *CacheLazyFill) makeRequestAndDecodeJSON(url string, rv interface{}) error {
	resp, err := c.client.DoGet(url)
	if err != nil {
		return err
	}
	defer resp.Close()

	err = json.NewDecoder(resp).Decode(rv)
	if err != nil {
		return err
	}

	return err
}

// both entityType and guid must have been validated by the caller
func (c *CacheLazyFill) fetchEntityFromAPI(entityType, guid string) (*entity, error) {
	var md struct {
		Entity *entity `json:"entity"`
	}
	err := c.makeRequestAndDecodeJSON(fmt.Sprintf("/v2/%s/%s", entityType, guid), &md)
	if err != nil {
		return nil, err
	}
	return md.Entity, nil
}

func (c *CacheLazyFill) getSpaceAndOrg(spaceGuid string) (*entity, *entity, error) {
	space, err := c.getEntity("spaces", spaceGuid)
	if err != nil {
		if c.config.IgnoreMissingApps {
			space = &entity{}
		} else {
			return nil, nil, err
		}
	}

	org, err := c.getEntity("organizations", space.OrganizationGUID)
	if err != nil {
		if c.config.IgnoreMissingApps {
			org = &entity{}
		} else {
			return nil, nil, err
		}
	}

	return space, org, nil
}

func (c *CacheLazyFill) GetApp(appGuid string) (*App, error) {
	app, err := c.getEntity("apps", appGuid)
	if err != nil {
		if c.config.IgnoreMissingApps {
			app = &entity{}
		} else {
			return nil, err
		}
	}

	space, org, err := c.getSpaceAndOrg(app.SpaceGUID)
	if err != nil {
		return nil, err
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
