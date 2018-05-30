package eventRouting_test

import (
	. "github.com/cloudfoundry-community/firehose-to-syslog/caching"
	. "github.com/cloudfoundry-community/firehose-to-syslog/caching/cachingfakes"
	. "github.com/cloudfoundry-community/firehose-to-syslog/eventRouting"
	"github.com/cloudfoundry-community/firehose-to-syslog/logging"
	. "github.com/cloudfoundry-community/firehose-to-syslog/logging/loggingfakes"
	. "github.com/cloudfoundry-community/firehose-to-syslog/stats"
	"github.com/cloudfoundry/sonde-go/events"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Events", func() {

	var eventRouting EventRouting
	var stats *Stats
	var caching *FakeCaching
	var filters []EventFilter
	var logging logging.Logging

	BeforeEach(func() {
		logging = new(FakeLogging)
		caching = new(FakeCaching)
		stats = &Stats{}
		filters = make([]EventFilter, 0)
		filters = append(filters, HasIgnoreField)
		eventRouting = NewEventRouting(caching, logging, stats, filters)
		eventRouting.SetupEventRouting("")

	})

	Context("called with a empty list", func() {
		It("should return a hash of only the default event", func() {
			expected := map[string]bool{"LogMessage": true}
			Expect(eventRouting.GetSelectedEvents()).To(Equal(expected))
		})
	})

	Context("called with a list of bogus event names", func() {
		It("should err out", func() {
			err := eventRouting.SetupEventRouting("bogus,bogus1")
			Expect(err).To(HaveOccurred())
		})
	})

	Context("called with a list of real event names", func() {
		It("should return a hash of events", func() {
			expected := map[string]bool{
				"HttpStartStop": true,
				"CounterEvent":  true,
			}
			eventRouting.SetupEventRouting("HttpStartStop,CounterEvent")
			Expect(eventRouting.GetSelectedEvents()).To(Equal(expected))
		})
	})

	// Context("called after 10 events have been routed", func() {
	// 	var expected = uint64(10)
	// 	BeforeEach(func() {
	// 		for i := 0; i < int(expected); i++ {
	// 			eventRouting.RouteEvent(&Envelope{EventType: Envelope_LogMessage.Enum()})
	// 		}
	// 	})
	//
	// 	It("should return a total of 10", func() {
	// 		Expect(eventRouting.GetTotalCountOfSelectedEvents()).To(Equal(expected))
	// 	})
	// })

	Context("GetListAuthorizedEventEvents", func() {
		It("should return right list of authorized events", func() {
			Expect(GetListAuthorizedEventEvents()).To(Equal("ContainerMetric, CounterEvent, Error, HttpStartStop, LogMessage, ValueMetric"))
		})
	})

	Context("Test Filter", func() {
		var testMsg *events.Envelope
		var logMessage *events.LogMessage
		var eventType events.Envelope_EventType
		BeforeEach(func() {
			eventType = events.Envelope_LogMessage
			logMessage = &events.LogMessage{}
			testMsg = &events.Envelope{
				EventType:  &eventType,
				LogMessage: logMessage,
			}
			stats.ConsumeLogMessage = uint64(0)
			stats.Publish = uint64(0)
			stats.Ignored = uint64(0)
		})
		Context("When the log message is consumed and published", func() {
			BeforeEach(func() {
				eventRouting.RouteEvent(testMsg)
			})
			It("Should have one more Log Message consumed", func() {
				Ω(stats.ConsumeLogMessage).Should(Equal(uint64(1)))
			})
			It("Should have one more event published", func() {
				Ω(stats.Publish).Should(Equal(uint64(1)))
			})
			It("Should have 0 event ignored", func() {
				Ω(stats.Ignored).Should(Equal(uint64(0)))
			})
		})
		Context("When the log message has cf_ignored_app field", func() {
			BeforeEach(func() {
				var appId = "Fake App"
				logMessage = &events.LogMessage{AppId: &appId}
				testMsg = &events.Envelope{
					EventType:  &eventType,
					LogMessage: logMessage,
				}
				caching.GetAppStub = func(string) (*App, error) {
					return &App{IgnoredApp: true}, nil
				}
				eventRouting.RouteEvent(testMsg)
			})
			It("Should have one more Log Message consumed", func() {
				Ω(stats.ConsumeLogMessage).Should(Equal(uint64(1)))
			})
			It("Should have one more event published", func() {
				Ω(stats.Publish).Should(Equal(uint64(0)))
			})
			It("Should have 0 event ignored", func() {
				Ω(stats.Ignored).Should(Equal(uint64(1)))
			})
		})
	})

	Context("When we specify which orgs log message we publish", func() {
		BeforeEach(func() {
			eventType := events.Envelope_LogMessage
			var appId1 = "Fake App1"
			logMessage1 := &events.LogMessage{AppId: &appId1}
			testMsg1 := &events.Envelope{
				EventType:  &eventType,
				LogMessage: logMessage1,
			}

			var appId2 = "Fake App2"
			logMessage2 := &events.LogMessage{AppId: &appId2}
			testMsg2 := &events.Envelope{
				EventType:  &eventType,
				LogMessage: logMessage2,
			}

			var appId3 = "Fake App3"
			logMessage3 := &events.LogMessage{AppId: &appId3}
			testMsg3 := &events.Envelope{
				EventType:  &eventType,
				LogMessage: logMessage3,
			}
			filters = append(filters, NotInCertainOrgs("org1|org3"))
			eventRouting = NewEventRouting(caching, logging, stats, filters)

			stubValue := map[string]string{
				appId1: "org1",
				appId2: "org2",
				appId3: "org3",
			}
			caching.GetAppStub = func(appId string) (*App, error) {
				return &App{OrgName: stubValue[appId]}, nil
			}

			eventRouting.RouteEvent(testMsg1)
			eventRouting.RouteEvent(testMsg2)
			eventRouting.RouteEvent(testMsg3)
		})
		It("Should have three Log Message consumed", func() {
			Ω(stats.ConsumeLogMessage).Should(Equal(uint64(3)))
		})
		It("Should have two events published", func() {
			Ω(stats.Publish).Should(Equal(uint64(2)))
		})
		It("Should have one event ignored", func() {
			Ω(stats.Ignored).Should(Equal(uint64(1)))
		})
	})
})
