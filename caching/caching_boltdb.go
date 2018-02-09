package caching

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/boltdb/bolt"
	"github.com/cloudfoundry-community/firehose-to-syslog/logging"
	cfclient "github.com/cloudfoundry-community/go-cfclient"
	json "github.com/mailru/easyjson"
	rate "go.uber.org/ratelimit"
)

const (
	APP_BUCKET = "AppBucket"
)

type CachingBoltConfig struct {
	Path               string
	IgnoreMissingApps  bool
	CacheInvalidateTTL time.Duration
	RequestBySec       int
}

type CachingBolt struct {
	appClient   AppClient
	appdb       *bolt.DB
	lock        sync.RWMutex
	cache       map[string]*App
	missingApps map[string]struct{}

	closing chan struct{}
	wg      sync.WaitGroup
	config  *CachingBoltConfig
}

func NewCachingBolt(client AppClient, config *CachingBoltConfig) (*CachingBolt, error) {
	return &CachingBolt{
		appClient:   client,
		cache:       make(map[string]*App),
		missingApps: make(map[string]struct{}),
		closing:     make(chan struct{}),
		config:      config,
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
// save remote API and ignore missing app, then a nil app info and an error
// will be returned.
func (c *CachingBolt) GetApp(appGuid string) (*App, error) {
	app, err := c.getAppFromCache(appGuid)
	if err != nil {
		return nil, err
	}

	// Find in cache
	if app != nil {
		return app, nil
	}

	// First time seeing app
	app, err = c.getAppFromRemote(appGuid)
	if err != nil {
		if c.config.IgnoreMissingApps {
			// Record this missing app
			c.lock.Lock()
			c.missingApps[appGuid] = struct{}{}
			c.lock.Unlock()
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

func (c *CachingBolt) getAppFromCache(appGuid string) (*App, error) {
	c.lock.RLock()
	if app, ok := c.cache[appGuid]; ok {
		// in in-memory cache
		c.lock.RUnlock()
		return app, nil
	}

	_, alreadyMissed := c.missingApps[appGuid]
	if c.config.IgnoreMissingApps && alreadyMissed {
		// already missed
		c.lock.RUnlock()
		return nil, errors.New("App was missed and ignored")
	}
	c.lock.RUnlock()

	// Didn't find in cache and it is not missed or we are not ignoring missed app
	return nil, nil
}

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

func (c *CachingBolt) createBucket() error {
	return c.appdb.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(APP_BUCKET))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
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

func (c *CachingBolt) addRecord(app *App) {
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

func (c *CachingBolt) fillDatabase(apps map[string]*App) {
	for _, app := range apps {
		c.addRecord(app)
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
	cfApp, err := c.appClient.GetAppByGuidNoInlineCall(appGuid)
	if err != nil {
		return nil, err
	}

	app := c.fromPCFApp(&cfApp)
	c.addRecord(app)
	return app, nil
}

func (c *CachingBolt) isOptOut(envVar map[string]interface{}) bool {
	if val, ok := envVar["F2S_DISABLE_LOGGING"]; ok && val == "true" {
		return true
	}
	return false
}

//Inspired by cf exporter
func (c *CachingBolt) getAllAppsFromRemote() (map[string]*App, error) {
	logging.LogStd("Retrieving Apps from Remote ...", false)
	apps := make(map[string]*App)
	// Listing Orgs
	organizations, err := c.appClient.ListOrgs()
	if err != nil {
		logging.LogError("Error while listing organization: %v", err)
		return apps, err
	}
	var wg = &sync.WaitGroup{}

	//Adding Global Rate limitings 50sec
	rl := rate.New(c.config.RequestBySec, rate.WithoutSlack)

	for _, organization := range organizations {
		wg.Add(1)
		go func(organization cfclient.Org, rl rate.Limiter) {
			defer wg.Done()
			rl.Take()
			err := c.getOrgSpaces(&organization, rl)
			if err != nil {
				logging.LogError("Error gettings Org", err)
			}
		}(organization, rl)
	}

	wg.Wait()
	logging.LogStd(fmt.Sprintf("Found [%d] Apps!", len(apps)), false)
	apps, err = c.getAllAppsFromBoltDB()
	if err != nil {
		logging.LogError("Error while gettings all apps from BoltDB: %v", err)
		return apps, err
	}
	return apps, nil
}
func (c *CachingBolt) getOrgSpaces(organization *cfclient.Org, l rate.Limiter) error {
	spaces, err := c.appClient.OrgSpaces(organization.Guid)
	if err != nil {
		logging.LogError("Error while listing spaces for organization `%s`: %v", err)
		return err
	}

	var wg = &sync.WaitGroup{}
	errChannel := make(chan error, len(spaces))

	for _, space := range spaces {
		wg.Add(1)
		go func(space cfclient.Space) {
			defer wg.Done()
			l.Take()
			err := c.getSpaceSummary(organization, &space)
			if err != nil {
				errChannel <- err
			}
		}(space)
	}

	wg.Wait()
	close(errChannel)

	return <-errChannel
}

func (c *CachingBolt) getSpaceSummary(organization *cfclient.Org, space *cfclient.Space) error {
	spaceSummary, err := space.Summary()
	if err != nil {
		logging.LogError("Error while getting summary for space `%s`: %v", err)
		return err
	}

	for _, application := range spaceSummary.Apps {
		app := c.fromPCFAppSummary(&application, organization, space)
		c.addRecord(app)
	}
	return nil
}

func (c *CachingBolt) fromPCFAppSummary(app *cfclient.AppSummary, organization *cfclient.Org, space *cfclient.Space) *App {
	return &App{
		app.Name,
		app.Guid,
		space.Name,
		space.Guid,
		organization.Name,
		organization.Guid,
		c.isOptOut(app.Environment),
	}
}
