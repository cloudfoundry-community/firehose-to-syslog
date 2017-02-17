package eventRouting

import (
	"sort"
	"strings"
	"time"

	"github.com/cloudfoundry/sonde-go/events"
)

type EventRouting interface {
	GetSelectedEvents() map[string]bool
	RouteEvent(msg *events.Envelope)
	SetupEventRouting(wantedEvents string) error
	SetExtraFields(extraEventsString string)
	GetTotalCountOfSelectedEvents() uint64
	GetSelectedEventsCount() map[string]uint64
	LogEventTotals(logTotalsTime time.Duration)
}

func IsAuthorizedEvent(wantedEvent string) bool {
	for _, authorizeEvent := range events.Envelope_EventType_name {
		if wantedEvent == authorizeEvent {
			return true
		}
	}
	return false
}

func GetListAuthorizedEventEvents() (authorizedEvents string) {
	arrEvents := []string{}
	for _, listEvent := range events.Envelope_EventType_name {
		arrEvents = append(arrEvents, listEvent)
	}
	sort.Strings(arrEvents)
	return strings.Join(arrEvents, ", ")
}
