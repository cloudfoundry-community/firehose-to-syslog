package authclient

import (
	"crypto/tls"
	"net/http"
)

type tokenFetcher interface {
	GetAuthToken(clientID, secret string, skipCertVerify bool) (string, error)
}

type AuthClient struct {
	tokenFetcher   tokenFetcher
	clientID       string
	secret         string
	skipCertVerify bool
	httpClient     *http.Client
}

func NewHttp(tf tokenFetcher, clientID, secret string, skipCertVerify bool) *AuthClient {
	return &AuthClient{
		tokenFetcher:   tf,
		clientID:       clientID,
		secret:         secret,
		skipCertVerify: skipCertVerify,
		httpClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: skipCertVerify,
				},
			},
		},
	}
}

func (c *AuthClient) Do(req *http.Request) (*http.Response, error) {
	token, err := c.tokenFetcher.GetAuthToken(
		c.clientID,
		c.secret,
		c.skipCertVerify,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", token)
	return c.httpClient.Do(req)
}
