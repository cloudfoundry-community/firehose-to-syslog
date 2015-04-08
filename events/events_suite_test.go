package events_test

import (
	"github.com/cloudfoundry-community/firehose-to-syslog/events"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
				Expect(events.GetSelectedEvents("")).To(Equal(expected))
			})
		})

		Context("called with a list of bogus event names", func() {
			It("should return a hash of only the default event", func() {
				expected := map[string]bool{"LogMessage": true}
				Expect(events.GetSelectedEvents("bogus,bogus1")).To(Equal(expected))
			})
		})

		Context("called with a list of both real and bogus event names", func() {
			It("should return a hash of only the real events", func() {
				expected := map[string]bool{
					"HttpStartStop": true,
					"CounterEvent":  true,
				}
				Expect(events.GetSelectedEvents("bogus,HttpStartStop,bogus1,CounterEvent")).To(Equal(expected))
			})
		})
	})
})
