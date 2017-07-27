package caching

type CachingEmpty struct{}

func NewCachingEmpty() Caching {
	return &CachingEmpty{}
}

func (c *CachingEmpty) Open() error {
	return nil
}

func (c *CachingEmpty) Close() error {
	return nil
}

func (c *CachingEmpty) GetAllApps() (map[string]*App, error) {
	return nil, nil
}

func (c *CachingEmpty) GetApp(appGuid string) (*App, error) {
	return nil, nil
}
