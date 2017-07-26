package events

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/cloudfoundry-community/firehose-to-syslog/caching"
	"github.com/cloudfoundry-community/firehose-to-syslog/utils"
	"github.com/cloudfoundry/sonde-go/events"
)

type Event struct {
	Fields map[string]interface{}
	Msg    string
	Type   string
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
		cf_ignored_app := appInfo.IgnoredApp

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

		e.Fields["cf_ignored_app"] = cf_ignored_app

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
	e.Fields["job_index"] = msg.GetIndex()
	e.Type = msg.GetEventType().String()

}
