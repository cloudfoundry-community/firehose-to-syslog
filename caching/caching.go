package caching

import (
	"errors"
	"net/http"
	"regexp"

	cfclient "github.com/cloudfoundry-community/go-cfclient"
)

type App struct {
	Name       string
	Guid       string
	SpaceName  string
	SpaceGuid  string
	OrgName    string
	OrgGuid    string
	IgnoredApp bool
}

var (
	// ErrKeyNotFound is returned if value is not found
	ErrKeyNotFound = errors.New("key not found")
)

// CacheStore provides a mechanism to persist data
// After it has been opened, Get / Set are threadsafe
type CacheStore interface {
	// Open initializes the store
	Open() error

	// Close closes the store
	Close() error

	// Get looks up key, and decodes it into rv.
	// Returns ErrKeyNotFound if not found
	Get(key string, rv interface{}) error

	// Set encodes the value and stores it
	Set(key string, value interface{}) error
}

type Caching interface {
	FillCache() error
	GetApp(string) (*App, error)
}

type CFSimpleClient interface {
	DoGet(url string) (*http.Response, error)
}

type AppClient interface {
	AppByGuid(appGuid string) (cfclient.App, error)
	ListOrgs() ([]cfclient.Org, error)
	OrgSpaces(guid string) ([]cfclient.Space, error)
	GetAppByGuidNoInlineCall(appGuid string) (cfclient.App, error)
}

func IsNeeded(wantedEvents string) bool {
	r := regexp.MustCompile("LogMessage|HttpStart|HttpStop|HttpStartStop|ContainerMetric")
	return r.MatchString(wantedEvents)
}
