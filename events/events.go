package events

import (
	"github.com/cloudfoundry-community/firehose-to-syslog/logging"
	"github.com/cloudfoundry/noaa/events"
	"strings"
)

func RouteEvents(in chan *events.Envelope, selectedEvents map[string]bool) {
	for msg := range in {
		eventType := msg.GetEventType()
		if selectedEvents[eventType.String()] {
			switch eventType {
			case events.Envelope_Heartbeat:
				logging.Heartbeats(msg)
			case events.Envelope_HttpStart:
				logging.HttpStarts(msg)
			case events.Envelope_HttpStop:
				logging.HttpStops(msg)
			case events.Envelope_HttpStartStop:
				logging.HttpStartStops(msg)
			case events.Envelope_LogMessage:
				logging.LogMessages(msg)
			case events.Envelope_ValueMetric:
				logging.ValueMetrics(msg)
			case events.Envelope_CounterEvent:
				logging.CounterEvents(msg)
			case events.Envelope_Error:
				logging.ErrorEvents(msg)
			case events.Envelope_ContainerMetric:
				logging.ContainerMetrics(msg)
			}
		}
	}
}

func GetSelectedEvents(wantedEvents string) map[string]bool {
	selectedEvents := make(map[string]bool)
	for _, event := range strings.Split(wantedEvents, ",") {
		selectedEvents[event] = true
	}
	return selectedEvents
}
