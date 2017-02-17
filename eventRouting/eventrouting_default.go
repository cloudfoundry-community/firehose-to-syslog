package eventRouting

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/cloudfoundry-community/firehose-to-syslog/caching"
	fevents "github.com/cloudfoundry-community/firehose-to-syslog/events"
	"github.com/cloudfoundry-community/firehose-to-syslog/extrafields"
	"github.com/cloudfoundry-community/firehose-to-syslog/logging"
	"github.com/cloudfoundry/sonde-go/events"
)

type EventRoutingDefault struct {
	CachingClient       caching.Caching
	selectedEvents      map[string]bool
	selectedEventsCount map[string]uint64
	mutex               *sync.Mutex
	log                 logging.Logging
	ExtraFields         map[string]string
}

func NewEventRouting(caching caching.Caching, logging logging.Logging) EventRouting {
	return &EventRoutingDefault{
		CachingClient:       caching,
		selectedEvents:      make(map[string]bool),
		selectedEventsCount: make(map[string]uint64),
		log:                 logging,
		mutex:               &sync.Mutex{},
		ExtraFields:         make(map[string]string),
	}
}

func (e *EventRoutingDefault) GetSelectedEvents() map[string]bool {
	return e.selectedEvents
}

func (e *EventRoutingDefault) RouteEvent(msg *events.Envelope) {

	eventType := msg.GetEventType()

	if e.selectedEvents[eventType.String()] {
		var event *fevents.Event
		switch eventType {
		case events.Envelope_HttpStartStop:
			event = fevents.HttpStartStop(msg)
		case events.Envelope_LogMessage:
			event = fevents.LogMessage(msg)
		case events.Envelope_ValueMetric:
			event = fevents.ValueMetric(msg)
		case events.Envelope_CounterEvent:
			event = fevents.CounterEvent(msg)
		case events.Envelope_Error:
			event = fevents.ErrorEvent(msg)
		case events.Envelope_ContainerMetric:
			event = fevents.ContainerMetric(msg)
		}

		event.AnnotateWithEnveloppeData(msg)

		event.AnnotateWithMetaData(e.ExtraFields)
		if _, hasAppId := event.Fields["cf_app_id"]; hasAppId {
			event.AnnotateWithAppData(e.CachingClient)
		}

		e.mutex.Lock()
		//We do not ship Event
		if ignored, hasIgnoredField := event.Fields["cf_ignored_app"]; ignored == true && hasIgnoredField {
			e.selectedEventsCount["ignored_app_message"]++
		} else {
			e.log.ShipEvents(event.Fields, event.Msg)
			e.selectedEventsCount[eventType.String()]++

		}
		e.mutex.Unlock()
	}
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

func (e *EventRoutingDefault) GetTotalCountOfSelectedEvents() uint64 {
	var total = uint64(0)
	for _, count := range e.GetSelectedEventsCount() {
		total += count
	}
	return total
}

func (e *EventRoutingDefault) GetSelectedEventsCount() map[string]uint64 {
	return e.selectedEventsCount
}

func (e *EventRoutingDefault) LogEventTotals(logTotalsTime time.Duration) {
	firehoseEventTotals := time.NewTicker(logTotalsTime)
	count := uint64(0)
	startTime := time.Now()
	totalTime := startTime

	go func() {
		for range firehoseEventTotals.C {
			elapsedTime := time.Since(startTime).Seconds()
			totalElapsedTime := time.Since(totalTime).Seconds()
			startTime = time.Now()
			event, lastCount := e.getEventTotals(totalElapsedTime, elapsedTime, count)
			count = lastCount
			e.log.ShipEvents(event.Fields, event.Msg)
		}
	}()
}

func (e *EventRoutingDefault) getEventTotals(totalElapsedTime float64, elapsedTime float64, lastCount uint64) (*fevents.Event, uint64) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	totalCount := e.GetTotalCountOfSelectedEvents()
	sinceLastTime := float64(int(elapsedTime*10)) / 10
	fields := logrus.Fields{
		"total_count":   totalCount,
		"by_sec_Events": int((totalCount - lastCount) / uint64(sinceLastTime)),
	}

	for eventtype, count := range e.GetSelectedEventsCount() {
		fields[eventtype] = count
	}

	event := &fevents.Event{
		Type:   "firehose_to_syslog_stats",
		Msg:    "Statistic for firehose to syslog",
		Fields: fields,
	}
	event.AnnotateWithMetaData(map[string]string{})
	return event, totalCount
}
