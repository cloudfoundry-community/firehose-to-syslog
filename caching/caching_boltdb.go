package caching

import (
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/cloudfoundry-community/firehose-to-syslog/logging"
	cfClient "github.com/cloudfoundry-community/go-cfclient"
	json "github.com/mailru/easyjson"
	"log"
	"os"
	"time"
	"bytes"
)

const (
	MISSING_APPS_BUCKET = "MissingAppsBucket"
	APP_BUCKET          = "AppBucket"
)

type CachingBolt struct {
	GcfClient *cfClient.Client
	Appdb     *bolt.DB
}

func NewCachingBolt(gcfClientSet *cfClient.Client, boltDatabasePath string) Caching {

	//Use bolt for in-memory  - file caching
	db, err := bolt.Open(boltDatabasePath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Fatal("Error opening bolt db: ", err)
		os.Exit(1)

	}

	cachingBolt := &CachingBolt{
		GcfClient: gcfClientSet,
		Appdb:     db,
	}

	cachingBolt.createMissingAppsBucket()

	return cachingBolt
}

func (c *CachingBolt) CreateBucket() {
	c.Appdb.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(APP_BUCKET))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
}

func (c *CachingBolt) createMissingAppsBucket() {
	c.Appdb.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(MISSING_APPS_BUCKET))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

	go func() {
		for range time.Tick(1 * time.Hour) {
			c.Appdb.Update(func(tx *bolt.Tx) error {
				err := tx.DeleteBucket([]byte(MISSING_APPS_BUCKET))
				if err != nil {
					return fmt.Errorf("clear bucket: %s", err)
				}
				return nil
			})
		}
	}()
}

func (c *CachingBolt) PerformPoollingCaching(tickerTime time.Duration) {
	// Ticker Pooling the CC every X sec
	ccPooling := time.NewTicker(tickerTime)

	var apps []App
	go func() {
		for range ccPooling.C {
			apps = c.GetAllApp()
		}
	}()

}

func (c *CachingBolt) fillDatabase(listApps []App) {
	for _, app := range listApps {
		c.Appdb.Update(func(tx *bolt.Tx) error {
			b, err := tx.CreateBucketIfNotExists([]byte(APP_BUCKET))
			if err != nil {
				return fmt.Errorf("create bucket: %s", err)
			}

			serialize, err := json.Marshal(app)

			if err != nil {
				return fmt.Errorf("Error Marshaling data: %s", err)
			}
			err = b.Put([]byte(app.Guid), serialize)

			if err != nil {
				return fmt.Errorf("Error inserting data: %s", err)
			}
			return nil
		})

	}

}

func (c *CachingBolt) GetAppByGuid(appGuid string) []App {
	var apps []App
	app, err := c.GcfClient.AppByGuid(appGuid)
	if err != nil {
		return apps
	}
	apps = append(apps, App{
		app.Name,
		app.Guid,
		app.SpaceData.Entity.Name,
		app.SpaceData.Entity.Guid,
		app.SpaceData.Entity.OrgData.Entity.Name,
		app.SpaceData.Entity.OrgData.Entity.Guid,
		c.isOptOut(app.Environment),
	})
	c.fillDatabase(apps)
	return apps

}

func (c *CachingBolt) GetAllApp() []App {

	logging.LogStd("Retrieving Apps for Cache...", false)
	var apps []App

	defer func() {
		if r := recover(); r != nil {
			logging.LogError("Recovered in caching.GetAllApp()", r)
		}
	}()

	cfApps, err := c.GcfClient.ListApps()
	if err != nil {
		return apps
	}

	for _, app := range cfApps {
		logging.LogStd(fmt.Sprintf("App [%s] Found...", app.Name), false)
		apps = append(apps, App{
			app.Name,
			app.Guid,
			app.SpaceData.Entity.Name,
			app.SpaceData.Entity.Guid,
			app.SpaceData.Entity.OrgData.Entity.Name,
			app.SpaceData.Entity.OrgData.Entity.Guid,
			c.isOptOut(app.Environment),
		})
	}

	c.fillDatabase(apps)
	logging.LogStd(fmt.Sprintf("Found [%d] Apps!", len(apps)), false)

	return apps
}

func (c *CachingBolt) GetAppInfo(appGuid string) App {

	var d []byte
	var app App
	c.Appdb.View(func(tx *bolt.Tx) error {
		logging.LogStd(fmt.Sprintf("Looking for App %s in Cache!\n", appGuid), false)
		b := tx.Bucket([]byte(APP_BUCKET))
		d = b.Get([]byte(appGuid))
		return nil
	})
	err := json.Unmarshal([]byte(d), &app)
	if err != nil {
		return App{}
	}
	return app
}

func (c *CachingBolt) alreadyMissed(appGuid string) bool {
	var d []byte

	c.Appdb.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(MISSING_APPS_BUCKET))
		d = b.Get([]byte(appGuid))
		return nil
	})

	return bytes.Equal(d, []byte("0"))
}

func (c *CachingBolt) recordMissingApp(appGuid string) {
	c.Appdb.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(MISSING_APPS_BUCKET))
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

func (c *CachingBolt) Close() {
	c.Appdb.Close()
}

func (c *CachingBolt) isOptOut(envVar map[string]interface{}) bool {
	if val, ok := envVar["F2S_DISABLE_LOGGING"]; ok != false && val == "true" {
		return true
	}
	return false
}

func (c *CachingBolt) GetAppInfoCache(appGuid string) App {
	if app := c.GetAppInfo(appGuid); app.Name != "" {
		return app
	}

	if c.alreadyMissed(appGuid) {
		return App{}
	}

	// First time seeing app
	apps := c.GetAppByGuid(appGuid)
	if apps[0].Name == "" {
		c.recordMissingApp(appGuid)
		return App{}
	}
	return apps[0]
}
