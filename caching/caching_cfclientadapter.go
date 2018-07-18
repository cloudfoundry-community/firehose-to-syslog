package caching

import (
	"net/http"

	cfclient "github.com/cloudfoundry-community/go-cfclient"
)

type CFClientAdapter struct {
	CF *cfclient.Client
}

func (c *CFClientAdapter) DoGet(url string) (*http.Response, error) {
	return c.CF.DoRequestWithoutRedirects(c.CF.NewRequest(http.MethodGet, url))
}
