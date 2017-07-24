package caching

import (
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

//go:generate counterfeiter . Caching

type Caching interface {
	Open() error
	Close() error

	GetAllApps() (map[string]*App, error)
	GetApp(string) (*App, error)
}

type AppClient interface {
	AppByGuid(appGuid string) (cfclient.App, error)
	ListApps() ([]cfclient.App, error)
}

func IsNeeded(wantedEvents string) bool {
	r := regexp.MustCompile("LogMessage|HttpStart|HttpStop|HttpStartStop|ContainerMetric")
	return r.MatchString(wantedEvents)
}
