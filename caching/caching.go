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

func FillDatabase(db *bolt.DB, cfClient *cfClient.Client) {

	for _, app := range cfClient.ListApps() {
		db.Update(func(tx *bolt.Tx) error {
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

func GetAppName(appGuid string, db *bolt.DB) string {
	var d []byte
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("AppBucket"))
		d = b.Get([]byte(appGuid))
		return nil
	})
	return string(d[:])
}
