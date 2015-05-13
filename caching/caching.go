package caching

import (
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	cfClient "github.com/cloudfoundry-community/go-cfclient"
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
	var apps []App
	for _, app := range gcfClient.ListApps() {
		apps = append(apps, App{app.Name, app.Guid, app.SpaceData.Entity.Name, app.SpaceData.Entity.Guid, app.SpaceData.Entity.OrgData.Entity.Name, app.SpaceData.Entity.OrgData.Entity.Guid})
	}
	FillDatabase(apps)
	return apps
}

func GetAppInfo(appGuid string) App {
	var d []byte
	var app App
	appdb.View(func(tx *bolt.Tx) error {
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
