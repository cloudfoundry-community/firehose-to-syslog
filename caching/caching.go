package caching

import (
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	log "github.com/cloudfoundry-community/firehose-to-syslog/logging"
	cfClient "github.com/cloudfoundry-community/go-cfclient"
)

// App struct that holds metadata about an application.
type App struct {
	Name      string
	GUID      string
	SpaceName string
	SpaceGUID string
	OrgName   string
	OrgGUID   string
}

var gcfClient *cfClient.Client
var appdb *bolt.DB

// Setup setups the neccecary stuff for caching.
func Setup(cfClient *cfClient.Client, db *bolt.DB) {
	appdb = db
	gcfClient = cfClient
	createBucket()
}

func createBucket() {
	appdb.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("AppBucket"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil

	})
}

func fillDatabase(listApps []App) {
	for _, app := range listApps {
		appdb.Update(func(tx *bolt.Tx) error {
			b, err := tx.CreateBucketIfNotExists([]byte("AppBucket"))
			if err != nil {
				return fmt.Errorf("create bucket: %s", err)
			}

			serialize, err := json.Marshal(app)

			if err != nil {
				return fmt.Errorf("Error Marshaling data: %s", err)
			}
			err = b.Put([]byte(app.GUID), serialize)

			if err != nil {
				return fmt.Errorf("Error inserting data: %s", err)
			}
			return nil
		})
	}
}

// GetAppByGUID fetches and stored metadata about a app in our cache, and returns the App
func GetAppByGUID(appGUID string) []App {
	var apps []App
	app := gcfClient.AppByGuid(appGUID)
	apps = append(apps, App{app.Name, app.Guid, app.SpaceData.Entity.Name, app.SpaceData.Entity.Guid, app.SpaceData.Entity.OrgData.Entity.Name, app.SpaceData.Entity.OrgData.Entity.Guid})
	fillDatabase(apps)
	return apps
}

// Fill gets all the apps with the gcfClient and sticks them in our cache
func Fill() {
	log.LogStd("Retrieving Apps for Cache...", false)
	var apps []App

	defer func() {
		if r := recover(); r != nil {
			log.LogError("Recovered in caching.GetAllApp()", r)
		}
	}()

	for _, app := range gcfClient.ListApps() {
		log.LogStd(fmt.Sprintf("App [%s] Found...", app.Name), false)
		apps = append(apps, App{app.Name, app.Guid, app.SpaceData.Entity.Name, app.SpaceData.Entity.Guid, app.SpaceData.Entity.OrgData.Entity.Name, app.SpaceData.Entity.OrgData.Entity.Guid})
	}

	fillDatabase(apps)

	log.LogStd(fmt.Sprintf("Found [%d] Apps!", len(apps)), false)
}

// GetAppInfo returns a App
func GetAppInfo(appGUID string) App {
	defer func() {
		if r := recover(); r != nil {
			log.LogError(fmt.Sprintf("Recovered from panic retrieving App Info for App Guid: %s", appGUID), r)
		}
	}()

	var d []byte
	var app App
	appdb.View(func(tx *bolt.Tx) error {
		log.LogStd(fmt.Sprintf("Looking for App %s in Cache!\n", appGUID), false)
		b := tx.Bucket([]byte("AppBucket"))
		d = b.Get([]byte(appGUID))
		return nil
	})
	err := json.Unmarshal([]byte(d), &app)
	if err != nil {
		return App{}
	}
	return app
}
