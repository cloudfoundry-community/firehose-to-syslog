package eventRouting

import (
	"strings"

	fevents "github.com/cloudfoundry-community/firehose-to-syslog/events"
)

//EventFilter Given an Event Filter out unwanted event
type EventFilter func(*fevents.Event) bool

//HasIgnoreField Filter out the event has ignored app filed
func HasIgnoreField(event *fevents.Event) bool {
	ignored, hasIgnoredField := event.Fields["cf_ignored_app"]
	return ignored == true && hasIgnoredField
}

//NotInCertainOrgs Filter out events not in certain orgs
func NotInCertainOrgs(orgFilters string) EventFilter {
	return func(event *fevents.Event) bool {
		orgName := event.Fields["cf_org_name"]
		if orgFilters == "" || orgName == nil {
			return false
		}
		orgs := strings.Split(orgFilters, "|")
		for _, org := range orgs {
			if org == orgName {
				return false
			}
		}
		return true
	}
}
