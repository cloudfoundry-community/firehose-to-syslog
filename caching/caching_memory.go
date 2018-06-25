package caching

import (
	"bytes"
	"encoding/gob"
	"sync"
)

type MemoryCacheStore struct {
	mu    sync.RWMutex
	cache map[string][]byte
}

func (mcs *MemoryCacheStore) Open() error {
	mcs.cache = make(map[string][]byte)
	return nil
}

func (mcs *MemoryCacheStore) Close() error {
	mcs.cache = nil
	return nil
}

func (mcs *MemoryCacheStore) Get(key string, rv interface{}) error {
	mcs.mu.RLock()
	v := mcs.cache[key]
	mcs.mu.RUnlock()

	if len(v) == 0 {
		return ErrKeyNotFound
	}
	return gob.NewDecoder(bytes.NewReader(v)).Decode(rv)
}

func (mcs *MemoryCacheStore) Set(key string, val interface{}) error {
	b := &bytes.Buffer{}
	err := gob.NewEncoder(b).Encode(val)
	if err != nil {
		return err
	}
	mcs.mu.Lock()
	mcs.cache[key] = b.Bytes()
	mcs.mu.Unlock()

	return nil
}
