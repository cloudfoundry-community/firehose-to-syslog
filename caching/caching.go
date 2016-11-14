package caching

import (
	"regexp"
	"time"
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
	CreateBucket()
	PerformPoollingCaching(time.Duration)
	GetAppByGuid(string) []App
	GetAllApp() []App
	GetAppInfo(string) App
	GetAppInfoCache(string) App
	Close()
}

func IsNeeded(wantedEvents string) bool {
	r := regexp.MustCompile("LogMessage|HttpStart|HttpStop|HttpStartStop|ContainerMetric")
	return r.MatchString(wantedEvents)
}
