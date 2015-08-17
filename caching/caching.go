package caching

import (
	"fmt"
	"github.com/boltdb/bolt"
	log "github.com/cloudfoundry-community/firehose-to-syslog/logging"
	cfClient "github.com/cloudfoundry-community/go-cfclient"
	json "github.com/pquerna/ffjson/ffjson"
)

type App struct {
	Name      string
	Guid      string
	SpaceName string
	SpaceGuid string
	OrgName   string
	OrgGuid   string
}

var gcfClient *cfClient.Client
var appdb *bolt.DB

func CreateBucket() {
	appdb.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("AppBucket"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil

	})

}

func FillDatabase(listApps []App) {

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
			err = b.Put([]byte(app.Guid), serialize)

			if err != nil {
				return fmt.Errorf("Error inserting data: %s", err)
			}
			return nil
		})

	}

}

func GetAppByGuid(appGuid string) []App {
	var apps []App
	app := gcfClient.AppByGuid(appGuid)
	apps = append(apps, App{app.Name, app.Guid, app.SpaceData.Entity.Name, app.SpaceData.Entity.Guid, app.SpaceData.Entity.OrgData.Entity.Name, app.SpaceData.Entity.OrgData.Entity.Guid})
	FillDatabase(apps)
	return apps

}

func GetAllApp() []App {

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

	FillDatabase(apps)

	log.LogStd(fmt.Sprintf("Found [%d] Apps!", len(apps)), false)

	return apps
}

func GetAppInfo(appGuid string) App {

	defer func() {
		if r := recover(); r != nil {
			log.LogError(fmt.Sprintf("Recovered from panic retrieving App Info for App Guid: %s", appGuid), r)
		}
	}()

	var d []byte
	var app App
	appdb.View(func(tx *bolt.Tx) error {
		log.LogStd(fmt.Sprintf("Looking for App %s in Cache!\n", appGuid), false)
		b := tx.Bucket([]byte("AppBucket"))
		d = b.Get([]byte(appGuid))
		return nil
	})
	err := json.Unmarshal([]byte(d), &app)
	if err != nil {
		return App{}
	}
	return app
}

func SetCfClient(cfClient *cfClient.Client) {
	gcfClient = cfClient

}

func SetAppDb(db *bolt.DB) {
	appdb = db
}
