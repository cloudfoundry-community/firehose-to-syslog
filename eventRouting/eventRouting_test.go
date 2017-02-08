package eventRouting_test

import (
	. "github.com/cloudfoundry-community/firehose-to-syslog/caching/cachingfakes"
	. "github.com/cloudfoundry-community/firehose-to-syslog/eventRouting"
	. "github.com/cloudfoundry-community/firehose-to-syslog/logging/loggingfakes"
	. "github.com/cloudfoundry/sonde-go/events"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Events", func() {

	var eventRouting EventRouting

	BeforeEach(func() {
		logging := new(FakeLogging)
		caching := new(FakeCaching)
		eventRouting = NewEventRouting(caching, logging)
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

	Context("called after 10 events have been routed", func() {
		var expected = uint64(10)
		BeforeEach(func() {
			for i := 0; i < int(expected); i++ {
				eventRouting.RouteEvent(&Envelope{EventType: Envelope_LogMessage.Enum()})
			}
		})

		It("should return a total of 10", func() {
			Expect(eventRouting.GetTotalCountOfSelectedEvents()).To(Equal(expected))
		})
	})


	Context("GetListAuthorizedEventEvents", func() {
		It("should return right list of authorized events", func() {
			Expect(GetListAuthorizedEventEvents()).To(Equal("ContainerMetric, CounterEvent, Error, HttpStartStop, LogMessage, ValueMetric"))
		})
	})

})
