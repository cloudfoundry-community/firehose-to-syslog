package caching

import (
	"fmt"
	"sync"
	"time"

	"github.com/boltdb/bolt"
	"github.com/cloudfoundry-community/firehose-to-syslog/logging"
	cfclient "github.com/cloudfoundry-community/go-cfclient"
	json "github.com/mailru/easyjson"
)

const (
	MISSING_APP_BUCKET = "MissingAppBucket"
	APP_BUCKET         = "AppBucket"
)

type CachingBoltConfig struct {
	Path               string
	IgnoreMissingApps  bool
	MissingAppsTTL     time.Duration
	CacheInvalidateTTL time.Duration
}

type CachingBolt struct {
	appClient AppClient
	appdb     *bolt.DB
	lock      sync.RWMutex
	cache     map[string]*App
	closing   chan struct{}
	wg        sync.WaitGroup
	config    *CachingBoltConfig
}

func NewCachingBolt(client AppClient, config *CachingBoltConfig) (*CachingBolt, error) {
	return &CachingBolt{
		appClient: client,
		closing:   make(chan struct{}),
		config:    config,
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

	if err := c.createBucket(); err != nil {
		logging.LogError("Fail to create bucket: ", err)
		return err
	}

	if c.config.CacheInvalidateTTL != time.Duration(0) {
		c.invalidateCache()
	}

	if c.config.IgnoreMissingApps {
		err := c.createMissingAppBucket(c.config.MissingAppsTTL)
		if err != nil {
			return err
		}
	}

	return c.populateCache()
}

func (c *CachingBolt) populateCache() error {
	apps, err := c.getAllAppsFromBoltDB()
	if err != nil {
		return err
	}

	if len(apps) == 0 {
		// populate from remote
		apps, err = c.getAllAppsFromRemote()
		if err != nil {
			return err
		}
	}

	c.cache = apps

	return nil
}

func (c *CachingBolt) Close() error {
	close(c.closing)

	// Wait for background goroutine exit
	c.wg.Wait()

	return c.appdb.Close()
}

// GetAppInfo tries first get app info from cache. If caches doesn't have this
// app info (cache miss), it issues API to retrieve the app info from remote
// if the app is not already missing and clients don't ignore the missing app
// info, and then add the app info to the cache
// On the other hand, if the app is already missing and clients want to
// save remote API and ignore missing app, then a nil app info and nil error
// will be returned. Clients probably need check the nil-ness of returned
// App info and error
func (c *CachingBolt) GetApp(appGuid string) (*App, error) {
	c.lock.RLock()
	if app, ok := c.cache[appGuid]; ok {
		// in in-memory cache
		c.lock.RUnlock()
		return app, nil
	}
	c.lock.RUnlock()

	if c.config.IgnoreMissingApps && c.alreadyMissed(appGuid) {
		return nil, nil
	}

	// First time seeing app
	app, err := c.getAppFromRemote(appGuid)
	if err != nil {
		if c.config.IgnoreMissingApps {
			c.recordMissingApp(appGuid)
		}
		return nil, err
	}

	// Add to in-memory cache
	c.lock.Lock()
	c.cache[app.Guid] = app
	c.lock.Unlock()

	return app, nil
}

// GetAllApps returns all apps info
func (c *CachingBolt) GetAllApps() (map[string]*App, error) {
	c.lock.RLock()
	apps := make(map[string]*App, len(c.cache))
	for _, app := range c.cache {
		dup := *app
		apps[dup.Guid] = &dup
	}
	c.lock.RUnlock()

	return apps, nil
}

// getAllAppsFromBoltDB get all app information from boltdb
func (c *CachingBolt) getAllAppsFromBoltDB() (map[string]*App, error) {
	var allData [][]byte
	c.appdb.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(APP_BUCKET))
		b.ForEach(func(guid []byte, v []byte) error {
			allData = append(allData, v)
			return nil
		})
		return nil
	})

	apps := make(map[string]*App, len(allData))
	for i := range allData {
		var app App
		err := json.Unmarshal(allData[i], &app)
		if err != nil {
			return nil, err
		}
		apps[app.Guid] = &app
	}

	return apps, nil
}

func (c *CachingBolt) getAllAppsFromRemote() (map[string]*App, error) {
	logging.LogStd("Retrieving Apps for Cache...", false)

	cfApps, err := c.appClient.ListApps()
	if err != nil {
		return nil, err
	}

	apps := make(map[string]*App, len(cfApps))
	for i := range cfApps {
		logging.LogStd(fmt.Sprintf("App [%s] Found...", cfApps[i].Name), false)
		app := c.fromPCFApp(&cfApps[i])
		apps[app.Guid] = app
	}

	c.fillDatabase(apps)
	logging.LogStd(fmt.Sprintf("Found [%d] Apps!", len(apps)), false)

	return apps, nil
}

func (c *CachingBolt) createBucket() error {
	return c.appdb.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(APP_BUCKET))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
}

func (c *CachingBolt) createMissingAppBucket(ttl time.Duration) error {
	err := c.appdb.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(MISSING_APP_BUCKET))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

	if err != nil {
		return err
	}

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()

		ticker := time.NewTicker(ttl)
		for {
			select {
			case <-ticker.C:
				c.appdb.Update(func(tx *bolt.Tx) error {
					err := tx.DeleteBucket([]byte(MISSING_APP_BUCKET))
					if err != nil {
						return fmt.Errorf("clear bucket: %s", err)
					}
					return nil
				})
			case <-c.closing:
				return
			}
		}
	}()

	return nil
}

// invalidateCache perodically fetches a full copy apps info from remote
// and update boltdb and in-memory cache
func (c *CachingBolt) invalidateCache() {
	ticker := time.NewTicker(c.config.CacheInvalidateTTL)

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()

		for {
			select {
			case <-ticker.C:
				// continue
				apps, err := c.getAllAppsFromRemote()
				if err == nil {
					c.lock.Lock()
					c.cache = apps
					c.lock.Unlock()
				}
			case <-c.closing:
				return
			}
		}
	}()
}

func (c *CachingBolt) fillDatabase(apps map[string]*App) {
	for _, app := range apps {
		c.appdb.Update(func(tx *bolt.Tx) error {
			serialize, err := json.Marshal(app)
			if err != nil {
				return fmt.Errorf("Error Marshaling data: %s", err)
			}

			b := tx.Bucket([]byte(APP_BUCKET))
			if err := b.Put([]byte(app.Guid), serialize); err != nil {
				return fmt.Errorf("Error inserting data: %s", err)
			}
			return nil
		})
	}
}

func (c *CachingBolt) fromPCFApp(app *cfclient.App) *App {
	return &App{
		app.Name,
		app.Guid,
		app.SpaceData.Entity.Name,
		app.SpaceData.Entity.Guid,
		app.SpaceData.Entity.OrgData.Entity.Name,
		app.SpaceData.Entity.OrgData.Entity.Guid,
		c.isOptOut(app.Environment),
	}
}

func (c *CachingBolt) getAppFromRemote(appGuid string) (*App, error) {
	cfApp, err := c.appClient.AppByGuid(appGuid)
	if err != nil {
		return nil, err
	}

	app := c.fromPCFApp(&cfApp)
	c.fillDatabase(map[string]*App{app.Guid: app})

	return app, nil
}

func (c *CachingBolt) alreadyMissed(appGuid string) bool {
	var d []byte

	c.appdb.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(MISSING_APP_BUCKET))
		if b == nil {
			return nil
		}

		d = b.Get([]byte(appGuid))
		return nil
	})

	return d != nil
}

func (c *CachingBolt) recordMissingApp(appGuid string) error {
	return c.appdb.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(MISSING_APP_BUCKET))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		err = b.Put([]byte(appGuid), []byte("0"))

		if err != nil {
			return fmt.Errorf("Error inserting data: %s", err)
		}
		return nil
	})
}

func (c *CachingBolt) isOptOut(envVar map[string]interface{}) bool {
	if val, ok := envVar["F2S_DISABLE_LOGGING"]; ok && val == "true" {
		return true
	}
	return false
}
