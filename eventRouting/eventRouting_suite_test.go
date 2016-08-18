package eventRouting_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestEventRouting(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "EventRouting Suite")
}
