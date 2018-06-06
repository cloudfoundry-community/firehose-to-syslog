package eventRouting

import (
	"fmt"
	"os"
	"strings"

	"github.com/cloudfoundry-community/firehose-to-syslog/caching"
	fevents "github.com/cloudfoundry-community/firehose-to-syslog/events"
	"github.com/cloudfoundry-community/firehose-to-syslog/extrafields"
	"github.com/cloudfoundry-community/firehose-to-syslog/logging"
	"github.com/cloudfoundry-community/firehose-to-syslog/stats"
	"github.com/cloudfoundry/sonde-go/events"
)

type EventRoutingDefault struct {
	CachingClient  caching.Caching
	selectedEvents map[string]bool
	Stats          *stats.Stats
	log            logging.Logging
	ExtraFields    map[string]string
	eventFilters   []EventFilter
}

func NewEventRouting(caching caching.Caching, logging logging.Logging, stats *stats.Stats, filters []EventFilter) EventRouting {
	return &EventRoutingDefault{
		CachingClient:  caching,
		selectedEvents: make(map[string]bool),
		log:            logging,
		Stats:          stats,
		ExtraFields:    make(map[string]string),
		eventFilters:   filters,
	}
}

func (e *EventRoutingDefault) GetSelectedEvents() map[string]bool {
	return e.selectedEvents
}

func (e *EventRoutingDefault) RouteEvent(msg *events.Envelope) {

	var event *fevents.Event
	switch msg.GetEventType() {
	case events.Envelope_HttpStartStop:
		event = fevents.HttpStartStop(msg)
		e.Stats.Inc(stats.ConsumeHttpStartStop)
	case events.Envelope_LogMessage:
		event = fevents.LogMessage(msg)
		e.Stats.Inc(stats.ConsumeLogMessage)
	case events.Envelope_ValueMetric:
		event = fevents.ValueMetric(msg)
		e.Stats.Inc(stats.ConsumeValueMetric)
	case events.Envelope_CounterEvent:
		event = fevents.CounterEvent(msg)
		e.Stats.Inc(stats.ConsumeCounterEvent)
	case events.Envelope_Error:
		event = fevents.ErrorEvent(msg)
		e.Stats.Inc(stats.ConsumeError)
	case events.Envelope_ContainerMetric:
		event = fevents.ContainerMetric(msg)
		e.Stats.Inc(stats.ConsumeContainerMetric)
	}

	event.AnnotateWithEnveloppeData(msg)

	event.AnnotateWithMetaData(e.ExtraFields)
	if _, hasAppId := event.Fields["cf_app_id"]; hasAppId {
		event.AnnotateWithAppData(e.CachingClient)
		//We do not ship Event for now event only concern app type of stream
		for _, filter := range e.eventFilters {
			if filter(event) {
				e.Stats.Inc(stats.Ignored)
				return
			}
		}
	}

	e.log.ShipEvents(event.Fields, event.Msg)
	e.Stats.Inc(stats.Publish)

}

func (e *EventRoutingDefault) SetupEventRouting(wantedEvents string) error {
	e.selectedEvents = make(map[string]bool)
	if wantedEvents == "" {
		e.selectedEvents["LogMessage"] = true
	} else {
		for _, event := range strings.Split(wantedEvents, ",") {
			if IsAuthorizedEvent(strings.TrimSpace(event)) {
				e.selectedEvents[strings.TrimSpace(event)] = true
				logging.LogStd(fmt.Sprintf("Event Type [%s] is included in the fireshose!", event), false)
			} else {
				return fmt.Errorf("Rejected Event Name [%s] - Valid events: %s", event, GetListAuthorizedEventEvents())
			}
		}
	}
	return nil
}

func (e *EventRoutingDefault) SetExtraFields(extraEventsString string) {
	// Parse extra fields from cmd call
	extraFields, err := extrafields.ParseExtraFields(extraEventsString)
	if err != nil {
		logging.LogError("Error parsing extra fields: ", err)
		os.Exit(1)
	}
	e.ExtraFields = extraFields
}
