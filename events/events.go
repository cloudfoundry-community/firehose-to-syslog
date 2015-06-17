package events

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/cloudfoundry-community/firehose-to-syslog/caching"
	log "github.com/cloudfoundry-community/firehose-to-syslog/logging"
	"github.com/cloudfoundry/noaa/events"
	"strings"
)

type Event struct {
	Fields logrus.Fields
	Msg    string
	Type   string
}

var selectedEvents map[string]bool

func RouteEvents(in chan *events.Envelope) {
	for msg := range in {
		routeEvent(msg)
	}
}

func GetSelectedEvents() map[string]bool {
	return selectedEvents
}

func routeEvent(msg *events.Envelope) {

	eventType := msg.GetEventType()

	if selectedEvents[eventType.String()] {
		var event Event
		switch eventType {
		case events.Envelope_Heartbeat:
			event = Heartbeat(msg)
		case events.Envelope_HttpStart:
			event = HTTPStart(msg)
		case events.Envelope_HttpStop:
			event = HTTPStop(msg)
		case events.Envelope_HttpStartStop:
			event = HTTPStartStop(msg)
		case events.Envelope_LogMessage:
			event = LogMessage(msg)
		case events.Envelope_ValueMetric:
			event = ValueMetric(msg)
		case events.Envelope_CounterEvent:
			event = CounterEvent(msg)
		case events.Envelope_Error:
			event = ErrorEvent(msg)
		case events.Envelope_ContainerMetric:
			event = ContainerMetric(msg)
		}

		event.AnnotateWithAppData()
		event.ShipEvent()
	}
}

func SetupEventRouting(wantedEvents string) {
	selectedEvents = make(map[string]bool)
	for _, event := range strings.Split(wantedEvents, ",") {
		if isAuthorizedEvent(event) {
			selectedEvents[event] = true
			log.LogStd(fmt.Sprintf("Event Type [%s] is included in the fireshose!", event), false)
		} else {
			log.LogError(fmt.Sprintf("Rejected Event Name [%s] - See wanted/selected events arg. Check Your Command Line Arguments!", event), "")
		}
	}
	// If any event is not authorize we fallback to the default one
	if len(selectedEvents) < 1 {
		selectedEvents["LogMessage"] = true
	}
}

func isAuthorizedEvent(wantedEvent string) bool {
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
	return strings.Join(arrEvents, ", ")

}

func getAppInfo(appGUID string) caching.App {
	if app := caching.GetAppInfo(appGUID); app.Name != "" {
		return app
	}
	caching.GetAppByGUID(appGUID)
	return caching.GetAppInfo(appGUID)
}

func Heartbeat(msg *events.Envelope) Event {
	heartbeat := msg.GetHeartbeat()

	var avail uint64

	if heartbeat.GetSentCount() > 0 {
		avail = heartbeat.GetReceivedCount() / heartbeat.GetSentCount()
	}

	fields := logrus.Fields{
		"ctl_msg_id":     heartbeat.GetControlMessageIdentifier(),
		"error_count":    heartbeat.GetErrorCount(),
		"origin":         msg.GetOrigin(),
		"received_count": heartbeat.GetReceivedCount(),
		"sent_count":     heartbeat.GetSentCount(),
		"availability":   avail,
	}

	return Event{
		Fields: fields,
		Msg:    "",
		Type:   msg.GetEventType().String(),
	}
}

func HTTPStart(msg *events.Envelope) Event {
	httpStart := msg.GetHttpStart()

	fields := logrus.Fields{
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
	}

	return Event{
		Fields: fields,
		Msg:    "",
		Type:   msg.GetEventType().String(),
	}
}

func HTTPStop(msg *events.Envelope) Event {
	httpStop := msg.GetHttpStop()

	fields := logrus.Fields{
		"origin":         msg.GetOrigin(),
		"cf_app_id":      httpStop.GetApplicationId(),
		"content_length": httpStop.GetContentLength(),
		"peer_type":      httpStop.GetPeerType(),
		"request_id":     httpStop.GetRequestId(),
		"status_code":    httpStop.GetStatusCode(),
		"timestamp":      httpStop.GetTimestamp(),
		"uri":            httpStop.GetUri(),
	}

	return Event{
		Fields: fields,
		Msg:    "",
		Type:   msg.GetEventType().String(),
	}
}

func HTTPStartStop(msg *events.Envelope) Event {
	httpStartStop := msg.GetHttpStartStop()

	fields := logrus.Fields{
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
		"duration_ms":       (((httpStartStop.GetStopTimestamp() - httpStartStop.GetStartTimestamp()) / 1000) / 1000),
	}

	return Event{
		Fields: fields,
		Msg:    "",
		Type:   msg.GetEventType().String(),
	}
}

func LogMessage(msg *events.Envelope) Event {
	logMessage := msg.GetLogMessage()

	fields := logrus.Fields{
		"origin":          msg.GetOrigin(),
		"cf_app_id":       logMessage.GetAppId(),
		"timestamp":       logMessage.GetTimestamp(),
		"source_type":     logMessage.GetSourceType(),
		"message_type":    logMessage.GetMessageType().String(),
		"source_instance": logMessage.GetSourceInstance(),
	}

	return Event{
		Fields: fields,
		Msg:    string(logMessage.GetMessage()),
		Type:   msg.GetEventType().String(),
	}
}

func ValueMetric(msg *events.Envelope) Event {
	valMetric := msg.GetValueMetric()

	fields := logrus.Fields{
		"origin": msg.GetOrigin(),
		"name":   valMetric.GetName(),
		"unit":   valMetric.GetUnit(),
		"value":  valMetric.GetValue(),
	}

	return Event{
		Fields: fields,
		Msg:    "",
		Type:   msg.GetEventType().String(),
	}
}

func CounterEvent(msg *events.Envelope) Event {
	counterEvent := msg.GetCounterEvent()

	fields := logrus.Fields{
		"origin": msg.GetOrigin(),
		"name":   counterEvent.GetName(),
		"delta":  counterEvent.GetDelta(),
		"total":  counterEvent.GetTotal(),
	}

	return Event{
		Fields: fields,
		Msg:    "",
		Type:   msg.GetEventType().String(),
	}
}

func ErrorEvent(msg *events.Envelope) Event {
	errorEvent := msg.GetError()

	fields := logrus.Fields{
		"origin": msg.GetOrigin(),
		"code":   errorEvent.GetCode(),
		"delta":  errorEvent.GetSource(),
	}

	return Event{
		Fields: fields,
		Msg:    errorEvent.GetMessage(),
		Type:   msg.GetEventType().String(),
	}
}

func ContainerMetric(msg *events.Envelope) Event {
	containerMetric := msg.GetContainerMetric()

	fields := logrus.Fields{
		"origin":         msg.GetOrigin(),
		"cf_app_id":      containerMetric.GetApplicationId(),
		"cpu_percentage": containerMetric.GetCpuPercentage(),
		"disk_bytes":     containerMetric.GetDiskBytes(),
		"instance_index": containerMetric.GetInstanceIndex(),
		"memory_bytes":   containerMetric.GetMemoryBytes(),
	}

	return Event{
		Fields: fields,
		Msg:    "",
		Type:   msg.GetEventType().String(),
	}
}

func (e *Event) AnnotateWithAppData() {

	cfAppID := e.Fields["cf_app_id"]
	appGUID := ""
	if cfAppID != nil {
		appGUID = fmt.Sprintf("%s", cfAppID)
	}

	if cfAppID != nil && appGUID != "<nil>" && cfAppID != "" {
		appInfo := getAppInfo(appGUID)
		cfAppName := appInfo.Name
		cfSpaceID := appInfo.SpaceGUID
		cfSpaceName := appInfo.SpaceName
		cfOrgID := appInfo.OrgGUID
		cfOrgName := appInfo.OrgName

		if cfAppName != "" {
			e.Fields["cf_app_name"] = cfAppName
		}

		if cfSpaceID != "" {
			e.Fields["cf_space_id"] = cfSpaceID
		}

		if cfSpaceName != "" {
			e.Fields["cf_space_name"] = cfSpaceName
		}

		if cfOrgID != "" {
			e.Fields["cf_org_id"] = cfOrgID
		}

		if cfOrgName != "" {
			e.Fields["cf_org_name"] = cfOrgName
		}
		e.Fields["cf_origin"] = "firehose"
		e.Fields["event_type"] = e.Type
	}
}

func (e Event) ShipEvent() {

	defer func() {
		if r := recover(); r != nil {
			log.LogError("Recovered in event.Log()", r)
		}
	}()

	logrus.WithFields(e.Fields).Info(e.Msg)
}
