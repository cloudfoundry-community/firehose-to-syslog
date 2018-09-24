package logging

import (
	"encoding/json"
	"fmt"
	"github.com/stvp/go-udp-testing"
	"math/rand"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
)

func TestEvents(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Logging Suite")
}

// Can't use ginkgo because go-udp-testing requires to be passed testing.T
func TestJSONCeeLoggingFormatter(t *testing.T) {
	RegisterTestingT(t)

	port := rand.Intn(65535 - 1080) + 1081
	listeningUDPSocket := fmt.Sprintf("localhost:%d", port)

	udp.SetAddr(listeningUDPSocket)
	result := udp.ReceiveString(t, func() {
		logging := NewLogging(listeningUDPSocket, "udp", "json-cee", "", false, true)
		logging.Connect()
		logging.ShipEvents(nil, "msg field content")
	})

	Ω(result).Should(ContainSubstring("@cee:"))

	jsonPayload := strings.Split(result, "@cee:")[1]
	var payload map[string]interface{}
	json.Unmarshal([]byte(jsonPayload), &payload)
	Ω(payload["msg"]).Should(Equal("msg field content"))
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
