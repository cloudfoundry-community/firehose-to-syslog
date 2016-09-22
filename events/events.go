package events

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"regexp"
	"sort"

	"github.com/Sirupsen/logrus"
	"github.com/cloudfoundry-community/firehose-to-syslog/caching"
	"github.com/cloudfoundry-community/firehose-to-syslog/logging"
	"github.com/cloudfoundry-community/firehose-to-syslog/utils"
	"github.com/cloudfoundry/sonde-go/events"
)

type MsgLine struct {
	Timestamp int64
	Msg       string
}
type MsgLines []MsgLine

func (slice MsgLines) Len() int {
	return len(slice)
}

func (slice MsgLines) Less(i, j int) bool {
	return slice[i].Timestamp < slice[j].Timestamp
}

func (slice MsgLines) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

type Event struct {
	Fields map[string]interface{}
	Msg    string
	Type   string
}

type MultiLineMessage struct {
	Event
	MsgLines MsgLines
}

func HttpStart(msg *events.Envelope) *Event {
	httpStart := msg.GetHttpStart()

	fields := logrus.Fields{
		"cf_app_id":         utils.FormatUUID(httpStart.GetApplicationId()),
		"instance_id":       httpStart.GetInstanceId(),
		"instance_index":    httpStart.GetInstanceIndex(),
		"method":            httpStart.GetMethod().String(),
		"parent_request_id": utils.FormatUUID(httpStart.GetParentRequestId()),
		"peer_type":         httpStart.GetPeerType().String(),
		"request_id":        utils.FormatUUID(httpStart.GetRequestId()),
		"remote_addr":       httpStart.GetRemoteAddress(),
		"timestamp":         httpStart.GetTimestamp(),
		"uri":               httpStart.GetUri(),
		"user_agent":        httpStart.GetUserAgent(),
	}

	return &Event{
		Fields: fields,
		Msg:    "",
	}
}

func HttpStop(msg *events.Envelope) *Event {
	httpStop := msg.GetHttpStop()

	fields := logrus.Fields{
		"cf_app_id":      utils.FormatUUID(httpStop.GetApplicationId()),
		"content_length": httpStop.GetContentLength(),
		"peer_type":      httpStop.GetPeerType().String(),
		"request_id":     utils.FormatUUID(httpStop.GetRequestId()),
		"status_code":    httpStop.GetStatusCode(),
		"timestamp":      httpStop.GetTimestamp(),
		"uri":            httpStop.GetUri(),
	}

	return &Event{
		Fields: fields,
		Msg:    "",
	}
}

func HttpStartStop(msg *events.Envelope) *Event {
	httpStartStop := msg.GetHttpStartStop()

	fields := logrus.Fields{
		"cf_app_id":       utils.FormatUUID(httpStartStop.GetApplicationId()),
		"content_length":  httpStartStop.GetContentLength(),
		"instance_id":     httpStartStop.GetInstanceId(),
		"instance_index":  httpStartStop.GetInstanceIndex(),
		"method":          httpStartStop.GetMethod().String(),
		"peer_type":       httpStartStop.GetPeerType().String(),
		"remote_addr":     httpStartStop.GetRemoteAddress(),
		"request_id":      utils.FormatUUID(httpStartStop.GetRequestId()),
		"start_timestamp": httpStartStop.GetStartTimestamp(),
		"status_code":     httpStartStop.GetStatusCode(),
		"stop_timestamp":  httpStartStop.GetStopTimestamp(),
		"uri":             httpStartStop.GetUri(),
		"user_agent":      httpStartStop.GetUserAgent(),
		"duration_ms":     (((httpStartStop.GetStopTimestamp() - httpStartStop.GetStartTimestamp()) / 1000) / 1000),
		"forwarded":       httpStartStop.GetForwarded(),
	}

	return &Event{
		Fields: fields,
		Msg:    "",
	}
}

func LogMessage(msg *events.Envelope) *Event {
	logMessage := msg.GetLogMessage()

	fields := logrus.Fields{
		"cf_app_id":       logMessage.GetAppId(),
		"timestamp":       logMessage.GetTimestamp(),
		"source_type":     logMessage.GetSourceType(),
		"message_type":    logMessage.GetMessageType().String(),
		"source_instance": logMessage.GetSourceInstance(),
	}

	return &Event{
		Fields: fields,
		Msg:    string(logMessage.GetMessage()),
	}
}

func ValueMetric(msg *events.Envelope) *Event {
	valMetric := msg.GetValueMetric()

	fields := logrus.Fields{
		"name":  valMetric.GetName(),
		"unit":  valMetric.GetUnit(),
		"value": valMetric.GetValue(),
	}

	return &Event{
		Fields: fields,
		Msg:    "",
	}
}

func CounterEvent(msg *events.Envelope) *Event {
	counterEvent := msg.GetCounterEvent()

	fields := logrus.Fields{
		"name":  counterEvent.GetName(),
		"delta": counterEvent.GetDelta(),
		"total": counterEvent.GetTotal(),
	}

	return &Event{
		Fields: fields,
		Msg:    "",
	}
}

func ErrorEvent(msg *events.Envelope) *Event {
	errorEvent := msg.GetError()

	fields := logrus.Fields{
		"code":   errorEvent.GetCode(),
		"source": errorEvent.GetSource(),
	}

	return &Event{
		Fields: fields,
		Msg:    errorEvent.GetMessage(),
	}
}

func ContainerMetric(msg *events.Envelope) *Event {
	containerMetric := msg.GetContainerMetric()

	fields := logrus.Fields{
		"cf_app_id":          containerMetric.GetApplicationId(),
		"cpu_percentage":     containerMetric.GetCpuPercentage(),
		"disk_bytes":         containerMetric.GetDiskBytes(),
		"disk_bytes_quota":   containerMetric.GetDiskBytesQuota(),
		"instance_index":     containerMetric.GetInstanceIndex(),
		"memory_bytes":       containerMetric.GetMemoryBytes(),
		"memory_bytes_quota": containerMetric.GetMemoryBytesQuota(),
	}

	return &Event{
		Fields: fields,
		Msg:    "",
	}
}

func (e *Event) AnnotateWithAppData(caching caching.Caching) {
	cf_app_id := e.Fields["cf_app_id"]
	appGuid := fmt.Sprintf("%s", cf_app_id)

	if cf_app_id != nil && appGuid != "<nil>" && cf_app_id != "" {
		appInfo := caching.GetAppInfoCache(appGuid)
		cf_app_name := appInfo.Name
		cf_space_id := appInfo.SpaceGuid
		cf_space_name := appInfo.SpaceName
		cf_org_id := appInfo.OrgGuid
		cf_org_name := appInfo.OrgName

		if cf_app_name != "" {
			e.Fields["cf_app_name"] = cf_app_name
		}

		if cf_space_id != "" {
			e.Fields["cf_space_id"] = cf_space_id
		}

		if cf_space_name != "" {
			e.Fields["cf_space_name"] = cf_space_name
		}

		if cf_org_id != "" {
			e.Fields["cf_org_id"] = cf_org_id
		}

		if cf_org_name != "" {
			e.Fields["cf_org_name"] = cf_org_name
		}
	}
}

func (e *Event) AnnotateWithMetaData(extraFields map[string]string) {
	e.Fields["cf_origin"] = "firehose"
	e.Fields["event_type"] = e.Type
	for k, v := range extraFields {
		e.Fields[k] = v
	}
}

func (e *Event) AnnotateWithEnveloppeData(msg *events.Envelope) {
	e.Fields["origin"] = msg.GetOrigin()
	e.Fields["deployment"] = msg.GetDeployment()
	e.Fields["ip"] = msg.GetIp()
	e.Fields["job"] = msg.GetJob()
	e.Fields["index"] = msg.GetIndex()
	e.Type = msg.GetEventType().String()
}

func (e *Event) CachePartialEventMessage(caching caching.Caching, log logging.Logging) bool {

	var err error
	var msg *MultiLineMessage

	cfAppID := e.Fields["cf_app_id"]
	cfAppIndex := e.Fields["index"]
	timestamp := e.Fields["timestamp"].(int64)

	isStackTraceStart, err := regexp.MatchString("^([a-z]\\w+\\.)+([a-zA-Z]\\w+Exception:).*", e.Msg)
	if err != nil {
		logging.LogError("Error checking if log message is a multi-line stacktrace!\n", err)
		return false
	}

	isStackTraceMiddle, err := regexp.MatchString("^\\tat ([a-z]\\w+\\.)+([a-zA-Z][\\$_0-9a-zA-Z]+)\\.([a-z]\\w+)\\(.*\\)", e.Msg)
	if err != nil {
		logging.LogError("Error checking if log message is a multi-line stacktrace!\n", err)
		return false
	}

	msgBytes := caching.GetMultiLineMessage(cfAppID.(string), cfAppIndex.(string))
	if len(msgBytes) > 0 {
		msg = &MultiLineMessage{}
		cachedBuffer := bytes.NewBuffer(msgBytes)
		if err = gob.NewDecoder(cachedBuffer).Decode(msg); err != nil {
			logging.LogError("Error deserializing multiline message.\n", err)
			return false
		}
		sort.Sort(msg.MsgLines)
	} else {
		msg = nil
	}

	if isStackTraceStart || isStackTraceMiddle {

		var buffer bytes.Buffer

		if msg == nil {
			msg = e.createMultiLineMessage(isStackTraceStart)
		} else {
			numCachedLines := len(msg.MsgLines)
			if numCachedLines > 0 && isStackTraceStart && msg.MsgLines[numCachedLines-1].Timestamp < timestamp {
				if msg.Fields["timestamp"] != nil {
					log.ShipEvents(msg.Fields, msg.MergeLines())
				}
				caching.DeleteMultiLineMessage(cfAppID.(string), cfAppIndex.(string))
			}
			e.addToMultiLineMessage(isStackTraceStart, msg)
		}

		if err = gob.NewEncoder(&buffer).Encode(msg); err != nil {
			logging.LogError("Error serializing multiline message.\n", err)
			return false
		}

		caching.PutMultiLineMessage(cfAppID.(string), cfAppIndex.(string), buffer.Bytes())
		return true
	}

	if msg != nil {
		if ts := msg.Fields["timestamp"]; ts != nil && ts.(int64) < timestamp {
			log.ShipEvents(msg.Fields, msg.MergeLines())
			caching.DeleteMultiLineMessage(cfAppID.(string), cfAppIndex.(string))
		}
	}

	return false
}

func (e *Event) createMultiLineMessage(isStackTraceStart bool) *MultiLineMessage {

	if isStackTraceStart {
		return &MultiLineMessage{*e, make([]MsgLine, 0)}
	}

	msgLines := []MsgLine{
		{
			Timestamp: e.Fields["timestamp"].(int64),
			Msg:       e.Msg,
		},
	}
	return &MultiLineMessage{
		MsgLines: msgLines,
	}
}

func (e *Event) addToMultiLineMessage(isStackTraceStart bool, msg *MultiLineMessage) {

	if isStackTraceStart {
		msg.Fields = e.Fields
		msg.Msg = e.Msg
		msg.Type = e.Type
	} else {
		msg.MsgLines = append(msg.MsgLines, MsgLine{e.Fields["timestamp"].(int64), e.Msg})
	}
}

func (msg *MultiLineMessage) MergeLines() string {

	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("%s\n", msg.Msg))
	for _, msgLine := range msg.MsgLines {
		buffer.WriteString(fmt.Sprintf("%s\n", msgLine.Msg))
	}
	return buffer.String()
}
