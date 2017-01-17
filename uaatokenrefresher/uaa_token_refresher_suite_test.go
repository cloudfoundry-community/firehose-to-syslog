package uaatokenrefresher_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestUaaTokenRefresher(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "UaaTokenRefresher Suite")
}
