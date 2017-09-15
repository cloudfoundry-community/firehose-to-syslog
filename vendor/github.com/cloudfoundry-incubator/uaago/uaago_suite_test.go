package uaago_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestUaaGo(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "UaaGo Suite")
}
