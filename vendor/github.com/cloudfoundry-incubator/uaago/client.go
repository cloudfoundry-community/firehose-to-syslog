package uaago

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type Client struct {
	uaaUrl *url.URL
}

func NewClient(uaaUrl string) (*Client, error) {
	if len(uaaUrl) == 0 {
		return nil, fmt.Errorf("client: missing url")
	}

	parsedURL, err := url.Parse(uaaUrl)
	if err != nil {
		return nil, err
	}

	return &Client{
		uaaUrl: parsedURL,
	}, nil
}

func (c *Client) GetAuthToken(username, password string, insecureSkipVerify bool) (string, error) {
	token, _, err := c.GetAuthTokenWithExpiresIn(username, password, insecureSkipVerify)
	return token, err
}

func (c *Client) GetAuthTokenWithExpiresIn(username, password string, insecureSkipVerify bool) (string, int, error) {
	data := url.Values{
		"client_id":  {username},
		"grant_type": {"client_credentials"},
	}

	request, err := http.NewRequest("POST", fmt.Sprintf("%s/oauth/token", c.uaaUrl), strings.NewReader(data.Encode()))
	if err != nil {
		return "", -1, err
	}
	request.SetBasicAuth(username, password)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient(insecureSkipVerify).Do(request)
	if err != nil {
		return "", -1, err
	}

	if resp.StatusCode != http.StatusOK {
		return "", -1, fmt.Errorf("Received a status code %v", resp.Status)
	}

	jsonData := make(map[string]interface{})
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&jsonData)

	expiresIn := 0
	if value, ok := jsonData["expires_in"]; ok {
		asFloat, err := strconv.ParseFloat(fmt.Sprintf("%f", value), 64)
		if err != nil {
			return "", -1, err
		}
		expiresIn = int(asFloat)
	}

	return fmt.Sprintf("%s %s", jsonData["token_type"], jsonData["access_token"]), expiresIn, err
}

func (c *Client) TokenIsAuthorized(username, password, token, client_id string, insecureSkipVerify bool) (bool, error) {

	request, err := http.NewRequest("POST", fmt.Sprintf("%s/check_token", c.uaaUrl.String()), strings.NewReader("token="+token))
	if err != nil {
		return false, err
	}
	request.SetBasicAuth(username, password)
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient(insecureSkipVerify).Do(request)
	if err != nil {
		return false, err
	}
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("Received a status code %v", resp.Status)
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	if strings.Contains(string(responseBody), client_id) {
		return true, nil
	}

	return false, nil
}

func (c *Client) httpClient(insecureSkipVerify bool) *http.Client {
	config := &tls.Config{InsecureSkipVerify: insecureSkipVerify}
	tr := &http.Transport{TLSClientConfig: config}
	return &http.Client{Transport: tr}
}
