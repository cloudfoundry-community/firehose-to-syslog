package logrus_syslog_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestLogrusSyslogNg(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "LogrusSyslogNg Suite")
}
