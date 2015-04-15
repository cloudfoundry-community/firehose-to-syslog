package caching

import (
	"fmt"
	"github.com/boltdb/bolt"
	cfClient "github.com/cloudfoundry-community/go-cfclient"
)

type App struct {
	Name string
	Guid string
}

var gcfClient *cfClient.Client
var appdb *bolt.DB

func FillDatabase(listApps []App) {

	for _, app := range listApps {
		appdb.Update(func(tx *bolt.Tx) error {
			b, err := tx.CreateBucketIfNotExists([]byte("AppBucket"))
			if err != nil {
				return fmt.Errorf("create bucket: %s", err)
			}

			err = b.Put([]byte(app.Guid), []byte(app.Name))

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
	apps = append(apps, App{app.Name, app.Guid})
	fmt.Println("Looking for ", appGuid)
	FillDatabase(apps)
	return apps

}

func GetAllApp() []App {
	var apps []App
	for _, app := range gcfClient.ListApps() {
		apps = append(apps, App{app.Name, app.Guid})
	}
	FillDatabase(apps)
	return apps
}

func GetAppName(appGuid string) string {
	var d []byte
	appdb.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("AppBucket"))
		d = b.Get([]byte(appGuid))
		return nil
	})
	return string(d[:])
}

func SetCfClient(cfClient *cfClient.Client) {
	gcfClient = cfClient

}

func SetAppDb(db *bolt.DB) {
	appdb = db
}
