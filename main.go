package main

import (
	"github.com/SpringerPE/firehose-to-syslog/firehose"
	"github.com/SpringerPE/firehose-to-syslog/logging"
	"github.com/SpringerPE/firehose-to-syslog/token"
	"github.com/cloudfoundry/noaa/events"
	"gopkg.in/alecthomas/kingpin.v1"
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
)

func RouteEvents(in chan *events.Envelope) {
	for msg := range in {
		switch msg.GetEventType() {
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

func main() {
	kingpin.Version("0.0.2 - ba541ca")
	kingpin.Parse()

	logging.SetupLogging(*syslogServer, *debug)
	token := token.GetToken(*uaaEndpoint, *firehoseUser, *firehosePassword, *skipSSLValidation)
	firehose := firehose.CreateFirehoseChan(*dopplerEndpoint, token, *subscriptionId, *skipSSLValidation)
	RouteEvents(firehose)
}
