package caching

import (
	"errors"
	"io"
	"regexp"
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

//go:generate counterfeiter . Caching
type Caching interface {
	FillCache() error
	GetApp(string) (*App, error)
}

//go:generate counterfeiter . CFSimpleClient
type CFSimpleClient interface {
	DoGet(url string) (io.ReadCloser, error)
}

func IsNeeded(wantedEvents string) bool {
	r := regexp.MustCompile("LogMessage|HttpStart|HttpStop|HttpStartStop|ContainerMetric")
	return r.MatchString(wantedEvents)
}
