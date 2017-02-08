package uaatokenrefresher

import (
	"fmt"

	"github.com/cloudfoundry-community/firehose-to-syslog/logging"
	"github.com/cloudfoundry-incubator/uaago"
)

type UAATokenRefresher struct {
	url               string
	clientID          string
	clientSecret      string
	skipSSLValidation bool
	client            *uaago.Client
}

func NewUAATokenRefresher(authEndpoint string,
	clientID string,
	clientSecret string,
	skipSSLValidation bool,
) (*UAATokenRefresher, error) {
	client, err := uaago.NewClient(authEndpoint)
	if err != nil {
		return &UAATokenRefresher{}, err
	}

	return &UAATokenRefresher{
		url:               authEndpoint,
		clientID:          clientID,
		clientSecret:      clientSecret,
		skipSSLValidation: skipSSLValidation,
		client:            client,
	}, nil
}

func (uaa *UAATokenRefresher) RefreshAuthToken() (string, error) {
	authToken, err := uaa.client.GetAuthToken(uaa.clientID, uaa.clientSecret, uaa.skipSSLValidation)
	if err != nil {
		logging.LogStd(fmt.Sprint("Error getting oauth token: %s. Please check your Client ID and Secret.", err.Error()), false)
		return "", err
	}

	return authToken, nil
}
