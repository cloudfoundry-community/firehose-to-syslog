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

func FilterEvents(in chan *events.Envelope) {
	for msg := range in {
		switch msg.GetEventType() {
		case events.Envelope_Heartbeat:
			Heartbeats(msg)
		case events.Envelope_HttpStart:
			HttpStarts(msg)
		case events.Envelope_HttpStop:
			HttpStops(msg)
		case events.Envelope_HttpStartStop:
			HttpStartStops(msg)
		case events.Envelope_LogMessage:
			LogMessages(msg)
		case events.Envelope_ValueMetric:
			ValueMetrics(msg)
		case events.Envelope_CounterEvent:
			CounterEvents(msg)
		case events.Envelope_Error:
			ErrorEvents(msg)
		case events.Envelope_ContainerMetric:
			ContainerMetrics(msg)
		}
	}
}

func Heartbeats(msg *events.Envelope) {
	metric := msg.GetHeartbeat()

	log.WithFields(log.Fields{
		"ctl_msg_id":     metric.GetControlMessageIdentifier(),
		"error_count":    metric.GetErrorCount(),
		"event_type":     msg.GetEventType(),
		"origin":         msg.GetOrigin(),
		"received_count": metric.GetReceivedCount(),
		"sent_count":     metric.GetSentCount(),
	}).Info("")
}

func HttpStarts(msg *events.Envelope) {
	metric := msg.GetHttpStart()

	log.WithFields(log.Fields{
		"event_type":        msg.GetEventType(),
		"origin":            msg.GetOrigin(),
		"cf_app_id":         metric.GetApplicationId(),
		"instance_id":       metric.GetInstanceId(),
		"instance_index":    metric.GetInstanceIndex(),
		"method":            metric.GetMethod(),
		"parent_request_id": metric.GetParentRequestId(),
		"peer_type":         metric.GetPeerType(),
		"request_id":        metric.GetRequestId(),
		"remote_addr":       metric.GetRemoteAddress(),
		"timestamp":         metric.GetTimestamp(),
		"uri":               metric.GetUri(),
		"user_agent":        metric.GetUserAgent(),
	}).Info("")
}

func HttpStops(msg *events.Envelope) {
	metric := msg.GetHttpStop()

	log.WithFields(log.Fields{
		"event_type":     msg.GetEventType(),
		"origin":         msg.GetOrigin(),
		"cf_app_id":      metric.GetApplicationId(),
		"content_length": metric.GetContentLength(),
		"peer_type":      metric.GetPeerType(),
		"request_id":     metric.GetRequestId(),
		"status_code":    metric.GetStatusCode(),
		"timestamp":      metric.GetTimestamp(),
		"uri":            metric.GetUri(),
	}).Info("")
}

func HttpStartStops(msg *events.Envelope) {
	metric := msg.GetHttpStartStop()

	log.WithFields(log.Fields{
		"event_type":        msg.GetEventType(),
		"origin":            msg.GetOrigin(),
		"cf_app_id":         metric.GetApplicationId(),
		"content_length":    metric.GetContentLength(),
		"instance_id":       metric.GetInstanceId(),
		"instance_index":    metric.GetInstanceIndex(),
		"method":            metric.GetMethod(),
		"parent_request_id": metric.GetParentRequestId(),
		"peer_type":         metric.GetPeerType(),
		"remote_addr":       metric.GetRemoteAddress(),
		"request_id":        metric.GetRequestId(),
		"start_timestamp":   metric.GetStartTimestamp(),
		"status_code":       metric.GetStatusCode(),
		"stop_timestamp":    metric.GetStopTimestamp(),
		"uri":               metric.GetUri(),
		"user_agent":        metric.GetUserAgent(),
	}).Info("")
}

func LogMessages(msg *events.Envelope) {
	logmsg := msg.GetLogMessage()
	app_id := logmsg.GetAppId()

	log.WithFields(log.Fields{
		"event_type":      msg.GetEventType(),
		"origin":          msg.GetOrigin(),
		"cf_app_id":       app_id,
		"timestamp":       logmsg.GetTimestamp(),
		"source_type":     logmsg.GetSourceType(),
		"message_type":    logmsg.GetMessageType().String(),
		"source_instance": logmsg.GetSourceInstance(),
	}).Info(string(logmsg.GetMessage()))
}

func ValueMetrics(msg *events.Envelope) {
	valMetric := msg.GetValueMetric()
	valueName := valMetric.GetName()
	valueUnit := valMetric.GetUnit()
	value := valMetric.GetValue()

	log.WithFields(log.Fields{
		"event_type": msg.GetEventType(),
		"origin":     msg.GetOrigin(),
		"name":       valueName,
		"unit":       valueUnit,
		"value":      value,
	}).Info("")
}

func CounterEvents(msg *events.Envelope) {
	evt := msg.GetCounterEvent()

	log.WithFields(log.Fields{
		"event_type": msg.GetEventType(),
		"origin":     msg.GetOrigin(),
		"name":       evt.GetName(),
		"delta":      evt.GetDelta(),
		"total":      evt.GetTotal(),
	}).Info("")
}

func ErrorEvents(msg *events.Envelope) {
	evt := msg.GetError()

	log.WithFields(log.Fields{
		"event_type": msg.GetEventType(),
		"origin":     msg.GetOrigin(),
		"code":       evt.GetCode(),
		"delta":      evt.GetSource(),
	}).Info(evt.GetMessage())
}

func ContainerMetrics(msg *events.Envelope) {
	contMetric := msg.GetContainerMetric()

	log.WithFields(log.Fields{
		"event_type":     msg.GetEventType(),
		"origin":         msg.GetOrigin(),
		"cf_app_id":      contMetric.GetApplicationId(),
		"cpu_percentage": contMetric.GetCpuPercentage(),
		"disk_bytes":     contMetric.GetDiskBytes(),
		"instance_index": contMetric.GetInstanceIndex(),
		"memory_bytes":   contMetric.GetMemoryBytes(),
	}).Info("")
}

func main() {
	kingpin.Version("0.0.2 - ba541ca")
	kingpin.Parse()

	setupLogging(*syslogServer, *debug)

	token := token.GetToken(*uaaEndpoint, *firehoseUser, *firehosePassword, *skipSSLValidation)

	firehose := CreateFirehoseChan(*dopplerEndpoint, token, *subscriptionId, *skipSSLValidation)

	FilterEvents(firehose)
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
