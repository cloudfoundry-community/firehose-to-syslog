package caching_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCaching(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Caching Suite")
}
