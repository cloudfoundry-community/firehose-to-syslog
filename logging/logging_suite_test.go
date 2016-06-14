package logging

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/Sirupsen/logrus"
	"testing"
)

func TestEvents(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Logging Suite")
}

var _ = Describe("Logging", func() {
	Describe("SetupLogging", func() {
		Context("called with a Text formatter", func() {
			It("should set the logging formatter as TextFormatter", func() {
				expected := &logrus.TextFormatter{}
				Expect(GetLogFormatter("text")).To(Equal(expected))
			})
		})

		Context("called with a JSON formatter", func() {
			It("should set the logging formatter as JSONFormatter", func() {
				expected := &logrus.JSONFormatter{}
				Expect(GetLogFormatter("json")).To(Equal(expected))
			})
		})

		Context("called with a nil formatter", func() {
			It("should set the logging formatter as JSONFormatter", func() {
				expected := &logrus.JSONFormatter{}
				Expect(GetLogFormatter("")).To(Equal(expected))
			})
		})
	})
})
