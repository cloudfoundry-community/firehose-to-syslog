package caching

import (
	"bytes"
	"encoding/gob"
	"time"

	"github.com/boltdb/bolt"
)

var (
	APP_BUCKET = []byte("AppBucketV2")
)

type BoltCacheStore struct {
	Path string

	db *bolt.DB
}

func (bcs *BoltCacheStore) Open() error {
	var err error
	bcs.db, err = bolt.Open(bcs.Path, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return err
	}
	return bcs.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(APP_BUCKET)
		return err
	})
}

func (bcs *BoltCacheStore) Close() error {
	return bcs.db.Close()
}

func (bcs *BoltCacheStore) Get(key string, rv interface{}) error {
	return bcs.db.View(func(tx *bolt.Tx) error {
		v := tx.Bucket(APP_BUCKET).Get([]byte(key))
		if len(v) == 0 {
			return ErrKeyNotFound
		}
		return gob.NewDecoder(bytes.NewReader(v)).Decode(rv)
	})
}

func (bcs *BoltCacheStore) Set(key string, val interface{}) error {
	b := &bytes.Buffer{}
	err := gob.NewEncoder(b).Encode(val)
	if err != nil {
		return err
	}
	return bcs.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(APP_BUCKET).Put([]byte(key), b.Bytes())
	})
}
