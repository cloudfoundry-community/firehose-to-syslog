package main

import (
	"github.com/cloudfoundry-community/firehose-to-syslog/firehose"
	"github.com/cloudfoundry-community/firehose-to-syslog/logging"
	"github.com/cloudfoundry-community/firehose-to-syslog/token"
	"github.com/cloudfoundry/noaa/events"
	"gopkg.in/alecthomas/kingpin.v1"
	"strings"
)

var (
	debug             = kingpin.Flag("debug", "Enable debug mode. This disables forwarding to syslog").Bool()
	uaaEndpoint       = kingpin.Flag("uaa-endpoint", "UAA endpoint.").Required().String()
	dopplerEndpoint   = kingpin.Flag("doppler-endpoint", "UAA endpoint.").Required().String()
	syslogServer      = kingpin.Flag("syslog-server", "Syslog server.").String()
	subscriptionId    = kingpin.Flag("subscription-id", "Id for the subscription.").Default("firehose").String()
	firehoseUser      = kingpin.Flag("firehose-user", "User with firehose permissions.").Default("doppler").String()
	firehosePassword  = kingpin.Flag("firehose-password", "Password for firehose user.").Default("doppler").String()
	skipSSLValidation = kingpin.Flag("skip-ssl-validation", "Please don't").Bool()
	wantedEvents      = kingpin.Flag("events", "Comma seperated list of events you would like. Valid options are Heartbeat, HttpStart, HttpStop, HttpStartStop, LogMessage, ValueMetric, CounterEvent, Error, ContainerMetric").Default("LogMessage").String()
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

func getSelectedEvents(wantedEvents string) map[string]bool {
	selectedEvents := make(map[string]bool)
	for _, event := range strings.Split(wantedEvents, ",") {
		selectedEvents[event] = true
	}
	return selectedEvents
}

func main() {
	kingpin.Version("0.0.2 - ba541ca")
	kingpin.Parse()

	selectedEvents := getSelectedEvents(*wantedEvents)
	logging.SetupLogging(*syslogServer, *debug)
	token := token.GetToken(*uaaEndpoint, *firehoseUser, *firehosePassword, *skipSSLValidation)
	firehose := firehose.CreateFirehoseChan(*dopplerEndpoint, token, *subscriptionId, *skipSSLValidation)
	RouteEvents(firehose, selectedEvents)
}
