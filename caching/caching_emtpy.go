package caching

import (
	"time"
)

type CachingEmpty struct{}

func NewCachingEmpty() Caching {
	return &CachingEmpty{}
}

func (c *CachingEmpty) CreateBucket() {}

func (c *CachingEmpty) PerformPoollingCaching(tickerTime time.Duration) {}

func (c *CachingEmpty) fillDatabase(listApps []App) {}

func (c *CachingEmpty) GetAppByGuid(appGuid string) []App {
	return []App{}
}

func (c *CachingEmpty) GetAllApp() []App {
	return []App{}
}

func (c *CachingEmpty) GetAppInfo(appGuid string) App {
	return App{}
}

func (c *CachingEmpty) Close() {}

func (c *CachingEmpty) GetAppInfoCache(appGuid string) App {
	return App{}
}
