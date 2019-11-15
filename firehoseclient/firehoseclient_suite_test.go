package firehoseclient_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestFirehoseclient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Firehoseclient Suite")
}
