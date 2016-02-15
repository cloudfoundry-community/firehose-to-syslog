package utils_test

import (
	. "github.com/cloudfoundry-community/firehose-to-syslog/utils"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing Utils packages", func() {
	Describe("UUID Formated", func() {
		Context("Called with proper UUID", func() {
			It("Should return formated String", func() {
				uuid := &events.UUID{High: proto.Uint64(0), Low: proto.Uint64(0)}
				Expect(FormatUUID(uuid)).To(Equal(("00000000-0000-0000-0000-000000000000")))
			})

		})
	})
	Describe("Concat String ", func() {
		Context("Called with String Map", func() {
			It("Should return Concat string", func() {
				Expect(ConcatFormat([]string{"foo", "bar"})).To(Equal(("foo.bar")))
			})
			It("Should return Proper string", func() {
				Expect(ConcatFormat([]string{"foo   ", "bar"})).To(Equal(("foo.bar")))
			})

		})
	})

})
