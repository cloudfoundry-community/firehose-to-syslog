package filters_test

import (
	. "github.com/eljuanchosf/firehose-to-syslog/filters"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
)

func TestEvents(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Filters Suite")
}

var _ = Describe("Filters", func() {
	Describe("ParseFilters", func() {
		Context("called with a empty string", func() {
			It("should return a empty hash", func() {
				expected := map[string][]string{}
				Expect(ParseFilters("")).To(Equal(expected))
			})
		})

		Context("called with values", func() {
			It("should return a hash with the filters we want", func() {
				expected := map[string][]string{"cf_org_name": {"org1","org2"}, "cf_app_name": {"wakawaka"}}
				filters := "org_name:org1,org2|app_name:wakawaka"
				Expect(ParseFilters(filters)).To(Equal(expected))
			})
		})

		Context("called with filters with weird whitespace", func() {
			It("should return a hash with the filters we want", func() {
				expected := map[string][]string{"cf_org_name": {"org1","org2"}, "cf_app_name": {"wakawaka"}}
				filters := "    org_name:    org1, org2|\n   app_name:    wakawaka"
				Expect(ParseFilters(filters)).To(Equal(expected))
			})
		})
	})
})
