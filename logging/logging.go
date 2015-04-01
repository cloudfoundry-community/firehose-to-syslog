package logging

import (
	log "github.com/Sirupsen/logrus"
	"github.com/Sirupsen/logrus/hooks/syslog"
	"github.com/cloudfoundry/noaa/events"
	"io/ioutil"
	"log/syslog"
	"os"
)

func SetupLogging(syslogServer string, debug bool) {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)
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

func Heartbeats(msg *events.Envelope) {
	heartbeat := msg.GetHeartbeat()

	log.WithFields(log.Fields{
		"ctl_msg_id":     heartbeat.GetControlMessageIdentifier(),
		"error_count":    heartbeat.GetErrorCount(),
		"event_type":     msg.GetEventType(),
		"origin":         msg.GetOrigin(),
		"received_count": heartbeat.GetReceivedCount(),
		"sent_count":     heartbeat.GetSentCount(),
	}).Info("")
}

func HttpStarts(msg *events.Envelope) {
	httpStart := msg.GetHttpStart()

	log.WithFields(log.Fields{
		"event_type":        msg.GetEventType(),
		"origin":            msg.GetOrigin(),
		"cf_app_id":         httpStart.GetApplicationId(),
		"instance_id":       httpStart.GetInstanceId(),
		"instance_index":    httpStart.GetInstanceIndex(),
		"method":            httpStart.GetMethod(),
		"parent_request_id": httpStart.GetParentRequestId(),
		"peer_type":         httpStart.GetPeerType(),
		"request_id":        httpStart.GetRequestId(),
		"remote_addr":       httpStart.GetRemoteAddress(),
		"timestamp":         httpStart.GetTimestamp(),
		"uri":               httpStart.GetUri(),
		"user_agent":        httpStart.GetUserAgent(),
	}).Info("")
}

func HttpStops(msg *events.Envelope) {
	httpStop := msg.GetHttpStop()

	log.WithFields(log.Fields{
		"event_type":     msg.GetEventType(),
		"origin":         msg.GetOrigin(),
		"cf_app_id":      httpStop.GetApplicationId(),
		"content_length": httpStop.GetContentLength(),
		"peer_type":      httpStop.GetPeerType(),
		"request_id":     httpStop.GetRequestId(),
		"status_code":    httpStop.GetStatusCode(),
		"timestamp":      httpStop.GetTimestamp(),
		"uri":            httpStop.GetUri(),
	}).Info("")
}

func HttpStartStops(msg *events.Envelope) {
	httpStartStop := msg.GetHttpStartStop()

	log.WithFields(log.Fields{
		"event_type":        msg.GetEventType(),
		"origin":            msg.GetOrigin(),
		"cf_app_id":         httpStartStop.GetApplicationId(),
		"content_length":    httpStartStop.GetContentLength(),
		"instance_id":       httpStartStop.GetInstanceId(),
		"instance_index":    httpStartStop.GetInstanceIndex(),
		"method":            httpStartStop.GetMethod(),
		"parent_request_id": httpStartStop.GetParentRequestId(),
		"peer_type":         httpStartStop.GetPeerType(),
		"remote_addr":       httpStartStop.GetRemoteAddress(),
		"request_id":        httpStartStop.GetRequestId(),
		"start_timestamp":   httpStartStop.GetStartTimestamp(),
		"status_code":       httpStartStop.GetStatusCode(),
		"stop_timestamp":    httpStartStop.GetStopTimestamp(),
		"uri":               httpStartStop.GetUri(),
		"user_agent":        httpStartStop.GetUserAgent(),
	}).Info("")
}

func LogMessages(msg *events.Envelope) {
	logMessage := msg.GetLogMessage()

	log.WithFields(log.Fields{
		"event_type":      msg.GetEventType(),
		"origin":          msg.GetOrigin(),
		"cf_app_id":       logMessage.GetAppId(),
		"timestamp":       logMessage.GetTimestamp(),
		"source_type":     logMessage.GetSourceType(),
		"message_type":    logMessage.GetMessageType().String(),
		"source_instance": logMessage.GetSourceInstance(),
	}).Info(string(logMessage.GetMessage()))
}

func ValueMetrics(msg *events.Envelope) {
	valMetric := msg.GetValueMetric()

	log.WithFields(log.Fields{
		"event_type": msg.GetEventType(),
		"origin":     msg.GetOrigin(),
		"name":       valMetric.GetName(),
		"unit":       valMetric.GetUnit(),
		"value":      valMetric.GetValue(),
	}).Info("")
}

func CounterEvents(msg *events.Envelope) {
	counterEvent := msg.GetCounterEvent()

	log.WithFields(log.Fields{
		"event_type": msg.GetEventType(),
		"origin":     msg.GetOrigin(),
		"name":       counterEvent.GetName(),
		"delta":      counterEvent.GetDelta(),
		"total":      counterEvent.GetTotal(),
	}).Info("")
}

func ErrorEvents(msg *events.Envelope) {
	errorEvent := msg.GetError()

	log.WithFields(log.Fields{
		"event_type": msg.GetEventType(),
		"origin":     msg.GetOrigin(),
		"code":       errorEvent.GetCode(),
		"delta":      errorEvent.GetSource(),
	}).Info(errorEvent.GetMessage())
}

func ContainerMetrics(msg *events.Envelope) {
	containerMetric := msg.GetContainerMetric()

	log.WithFields(log.Fields{
		"event_type":     msg.GetEventType(),
		"origin":         msg.GetOrigin(),
		"cf_app_id":      containerMetric.GetApplicationId(),
		"cpu_percentage": containerMetric.GetCpuPercentage(),
		"disk_bytes":     containerMetric.GetDiskBytes(),
		"instance_index": containerMetric.GetInstanceIndex(),
		"memory_bytes":   containerMetric.GetMemoryBytes(),
	}).Info("")
}
