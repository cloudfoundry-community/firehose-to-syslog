package caching_test

import (
	"github.com/deejross/firehose-to-syslog/caching"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCaching(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Caching Suite")
}

var _ = Describe("Caching", func() {
	Describe("IsNeeded", func() {
		Context("When user wants events that need to be decorated with app details", func() {
			It("returns true", func() {
				events := "ValueMetric,Error,LogMessage"
				Expect(caching.IsNeeded(events)).To(BeTrue())

				events = "HttpStartStop,ValueMetric"
				Expect(caching.IsNeeded(events)).To(BeTrue())

				events = "ContainerMetric,ValueMetric"
				Expect(caching.IsNeeded(events)).To(BeTrue())
			})
		})

		Context("When logs dont need to be decorated with app details", func() {
			It("returns false", func() {
				events := "ValueMetric,CounterEvent,Error"
				Expect(caching.IsNeeded(events)).To(BeFalse())
			})
		})
	})
})
