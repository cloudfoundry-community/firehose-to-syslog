package events_test

import (
	"testing"

	. "github.com/cloudfoundry/sonde-go/events"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestEvents(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Events Suite")
}

func CreateLogMessage() (msg *Envelope) {
	var eventType Envelope_EventType = 5
	var messageType LogMessage_MessageType = 1
	var posixStart int64 = 1
	var origin string = "yomomma__0"
	var sourceType string = "Kehe"
	var logMsg string = "Help, I'm a rock! Help, I'm a rock! Help, I'm a cop! Help, I'm a cop!"
	var sourceInstance string = ">9000"
	var appID string = "eea38ba5-53a5-4173-9617-b442d35ec2fd"

	logMessage := LogMessage{
		Message:        []byte(logMsg),
		AppId:          &appID,
		Timestamp:      &posixStart,
		SourceType:     &sourceType,
		MessageType:    &messageType,
		SourceInstance: &sourceInstance,
	}

	envelope := &Envelope{
		EventType:  &eventType,
		Origin:     &origin,
		LogMessage: &logMessage,
	}
	return envelope
}
