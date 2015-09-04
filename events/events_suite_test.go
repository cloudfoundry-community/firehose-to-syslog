package events_test

import (
	"github.com/Sirupsen/logrus"
	"github.com/cloudfoundry-community/firehose-to-syslog/events"
	. "github.com/cloudfoundry/sonde-go/events"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"testing"
)

func TestEvents(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Events Suite")
}

var _ = Describe("Events", func() {
	Describe("GetSelectedEvents", func() {
		Context("called with a empty list", func() {
			It("should return a hash of only the default event", func() {
				expected := map[string]bool{"LogMessage": true}
				events.SetupEventRouting("")
				Expect(events.GetSelectedEvents()).To(Equal(expected))
			})
		})

		Context("called with a list of bogus event names", func() {
			It("should err out", func() {
				err := events.SetupEventRouting("bogus,bogus1")
				Expect(err).To(HaveOccurred())
			})
		})

		Context("called with a list of real event names", func() {
			It("should return a hash of events", func() {
				expected := map[string]bool{
					"HttpStartStop": true,
					"CounterEvent":  true,
				}
				events.SetupEventRouting("HttpStartStop,CounterEvent")
				Expect(events.GetSelectedEvents()).To(Equal(expected))
			})
		})
	})

	Describe("GetTotalCountOfSelectedEvents", func() {
		Context("called after 10 events have been routed", func() {
			var expected = uint64(10)
			BeforeEach(func() {
				msgChan := make(chan *Envelope)
				go func() {
					defer close(msgChan)
					for i := 0; i < int(expected); i++ {
						msg := &Envelope{EventType: Envelope_LogMessage.Enum()}
						msgChan <- msg
					}
				}()
				logrus.SetOutput(ioutil.Discard)
				events.SetupEventRouting("")
				events.RouteEvents(msgChan, map[string]string{})
			})
			It("should return a total of 10", func() {
				Expect(events.GetTotalCountOfSelectedEvents()).To(Equal(expected))
			})
		})
	})

	Describe("Constructing a Event from a LogMessage", func() {
		var eventType Envelope_EventType = 5
		var messageType LogMessage_MessageType = 1
		var posixStart int64 = 1
		origin := "yomomma__0"
		sourceType := "Kehe"
		logMsg := "Help, I'm a rock! Help, I'm a rock! Help, I'm a cop! Help, I'm a cop!"
		sourceInstance := ">9000"
		appID := "eea38ba5-53a5-4173-9617-b442d35ec2fd"

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

		Context("given a envelope", func() {
			It("should give us what we want", func() {
				event := events.LogMessage(envelope)
				Expect(event.Fields["origin"]).To(Equal(origin))
				Expect(event.Fields["cf_app_id"]).To(Equal(appID))
				Expect(event.Fields["timestamp"]).To(Equal(posixStart))
				Expect(event.Fields["source_type"]).To(Equal(sourceType))
				Expect(event.Fields["message_type"]).To(Equal("OUT"))
				Expect(event.Fields["source_instance"]).To(Equal(sourceInstance))
				Expect(event.Msg).To(Equal(logMsg))
			})
		})
	})

	Describe("AnnotateWithAppData", func() {
		Context("called with Fields set to empty map", func() {
			It("should do nothing", func() {
				event := events.Event{}
				wanted := events.Event{}
				event.AnnotateWithAppData()
				Expect(event).To(Equal(wanted))
			})
		})

		Context("called with Fields set to logrus.Fields", func() {
			It("should do nothing", func() {
				event := events.Event{logrus.Fields{}, "", "log"}
				wanted := events.Event{logrus.Fields{}, "", "log"}
				event.AnnotateWithAppData()
				Expect(event).To(Equal(wanted))
			})
		})

		Context("called with empty cf_app_id", func() {
			It("should do nothing", func() {
				event := events.Event{logrus.Fields{"cf_app_id": ""}, "", "log"}
				wanted := events.Event{logrus.Fields{"cf_app_id": ""}, "", "log"}
				event.AnnotateWithAppData()
				Expect(event).To(Equal(wanted))
			})
		})
	})
})
