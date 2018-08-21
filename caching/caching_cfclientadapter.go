package caching

import (
	"fmt"
	"io"
	"net/http"

	cfclient "github.com/cloudfoundry-community/go-cfclient"
)

type CFClientAdapter struct {
	CF *cfclient.Client
}

func (c *CFClientAdapter) DoGet(url string) (io.ReadCloser, error) {
	resp, err := c.CF.DoRequestWithoutRedirects(c.CF.NewRequest(http.MethodGet, url))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("bad status code: %s", resp.Status)
	}

	return resp.Body, nil
}
