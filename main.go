package main

import (
	"crypto/tls"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/Sirupsen/logrus/hooks/syslog"
	"github.com/SpringerPE/firehose-to-syslog/token"
	"github.com/cloudfoundry/noaa"
	"github.com/cloudfoundry/noaa/events"
	"gopkg.in/alecthomas/kingpin.v1"
	"io/ioutil"
	"log/syslog"
	"os"
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
	allEvents         = kingpin.Flag("all-events", "Process all events. If this is unset we will only consume log messages from the routing layer and application logs.").Bool()
)

func CreateFirehoseChan(DopplerEndpoint string, Token string, subId string, skipSSLValidation bool) chan *events.Envelope {
	connection := noaa.NewConsumer(DopplerEndpoint, &tls.Config{InsecureSkipVerify: skipSSLValidation}, nil)
	msgChan := make(chan *events.Envelope)
	go func() {
		errorChan := make(chan error)
		defer close(msgChan)
		defer close(errorChan)

		go connection.Firehose(subId, Token, msgChan, errorChan, nil)

		for err := range errorChan {
			fmt.Fprintf(os.Stderr, "%v\n", err.Error())
		}
	}()
	return msgChan
}

func FilterEvent(in chan *events.Envelope, eventType string) chan *events.Envelope {
	evts := make(chan *events.Envelope)
	go func() {
		defer close(evts)
		for msg := range in {
			switch msg.GetEventType().String() {
			case eventType:
				evts <- msg
			}
		}
	}()
	return evts
}

func ValueMetrics(in chan *events.Envelope) {
	for msg := range in {
		valMetric := msg.GetValueMetric()

		valueName := valMetric.GetName()
		valueUnit := valMetric.GetUnit()
		value := valMetric.GetValue()

		log.WithFields(log.Fields{
			"name":       valueName,
			"unit":       valueUnit,
			"value":      value,
			"event_type": "ValueMetric",
		}).Info("")
	}
}

func ContainerMetrics(in chan *events.Envelope) {
	for msg := range in {
		contMetric := msg.GetContainerMetric()

		log.WithFields(log.Fields{
			"cf_app_id":      contMetric.GetApplicationId(),
			"cpu_percentage": contMetric.GetCpuPercentage(),
			"disk_bytes":     contMetric.GetDiskBytes(),
			"instance_index": contMetric.GetInstanceIndex(),
			"memory_bytes":   contMetric.GetMemoryBytes(),
			"event_type":     "ContainerMetric",
		}).Info("")
	}
}

func Heartbeats(in chan *events.Envelope) {
	for msg := range in {
		heartbeat := msg.GetHeartbeat()

		log.WithFields(log.Fields{
			"error_count":    heartbeat.GetErrorCount(),
			"received_count": heartbeat.GetReceivedCount(),
			"sent_count":     heartbeat.GetSentCount(),
			"event_type":     "Heartbeat",
		}).Info("")
	}
}

func LogMessages(in chan *events.Envelope) {
	for msg := range in {
		logmsg := msg.GetLogMessage()
		app_id := logmsg.GetAppId()

		log.WithFields(log.Fields{
			"cf_app_id":       app_id,
			"timestamp":       logmsg.GetTimestamp(),
			"source_type":     logmsg.GetSourceType(),
			"message_type":    logmsg.GetMessageType().String(),
			"source_instance": logmsg.GetSourceInstance(),
			"event_type":      "LogMessage",
		}).Info(string(logmsg.GetMessage()))
	}
}

func setupLogging(syslogServer string, debug bool) {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stderr)
	if !debug {
		log.SetOutput(ioutil.Discard)
	}
	if syslogServer != "" {
		hook, err := logrus_syslog.NewSyslogHook("tcp", syslogServer, syslog.LOG_INFO, "doppler")
		if err != nil {
			log.Error("Unable to connect to syslog server.")
		} else {
			log.AddHook(hook)
		}
	}
}

func CounterEvents(in chan *events.Envelope) {
	for msg := range in {
		evt := msg.GetCounterEvent()

		log.WithFields(log.Fields{
			"name":       evt.GetName(),
			"delta":      evt.GetDelta(),
			"total":      evt.GetTotal(),
			"event_type": "CounterEvent",
		}).Info("")
	}
}

func ErrorEvents(in chan *events.Envelope) {
	for msg := range in {
		evt := msg.GetError()

		log.WithFields(log.Fields{
			"code":       evt.GetCode(),
			"delta":      evt.GetSource(),
			"event_type": "ErrorEvent",
		}).Info(evt.GetMessage())
	}
}

func main() {
	kingpin.Version("0.0.1 - 2a962c4")
	kingpin.Parse()

	setupLogging(*syslogServer, *debug)

	token := token.GetToken(*uaaEndpoint, *firehoseUser, *firehosePassword, *skipSSLValidation)

	firehose := CreateFirehoseChan(*dopplerEndpoint, token, *subscriptionId, *skipSSLValidation)

	if *allEvents {
		go Heartbeats(FilterEvent(firehose, "Heartbeat"))
		// httpstart := FilterEvent(firehose, "HttpStart")
		// httpstop := FilterEvent(firehose, "HttpStop")
		// httpstartstop := FilterEvent(firehose, "HttpStartStop")
		go LogMessages(FilterEvent(firehose, "LogMessage"))
		go ValueMetrics(FilterEvent(firehose, "ValueMetric"))
		go CounterEvents(FilterEvent(firehose, "CounterEvent"))
		go ErrorEvents(FilterEvent(firehose, "Error"))
		ContainerMetrics(FilterEvent(firehose, "ContainerMetric"))
	} else {
		LogMessages(FilterEvent(firehose, "LogMessage"))
	}

}
