package authclient_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestAuthclient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Authclient Suite")
}
