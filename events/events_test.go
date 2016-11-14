package events_test

import (
	. "github.com/cloudfoundry-community/firehose-to-syslog/caching"
	. "github.com/cloudfoundry-community/firehose-to-syslog/caching/cachingfakes"
	fevents "github.com/cloudfoundry-community/firehose-to-syslog/events"
	. "github.com/cloudfoundry/sonde-go/events"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Events", func() {
	var caching *FakeCaching
	var event *fevents.Event
	var msg *Envelope
	BeforeEach(func() {
		caching = new(FakeCaching)
		msg = CreateLogMessage()
		event = fevents.LogMessage(msg)
		event.AnnotateWithEnveloppeData(msg)
	})

	Context("given a envelope", func() {
		It("should give us what we want", func() {
			Expect(event.Fields["origin"]).To(Equal("yomomma__0"))
			Expect(event.Fields["cf_app_id"]).To(Equal("eea38ba5-53a5-4173-9617-b442d35ec2fd"))
			Expect(event.Fields["timestamp"]).To(Equal(int64(1)))
			Expect(event.Fields["source_type"]).To(Equal("Kehe"))
			Expect(event.Fields["message_type"]).To(Equal("OUT"))
			Expect(event.Fields["source_instance"]).To(Equal(">9000"))
			Expect(event.Msg).To(Equal("Help, I'm a rock! Help, I'm a rock! Help, I'm a cop! Help, I'm a cop!"))
		})
	})
	Context("given metadata", func() {
		It("Should give us the right metadata", func() {
			event.AnnotateWithMetaData(map[string]string{"extra": "field"})
			Expect(event.Fields["cf_origin"]).To(Equal("firehose"))
			Expect(event.Fields["event_type"]).To(Equal(event.Type))
			Expect(event.Fields["extra"]).To(Equal("field"))

		})

	})

	Context("given Application Metadata", func() {
		It("Should give us the right Application metadata", func() {
			caching.GetAppInfoCacheStub = func(appid string) App {
				Expect(appid).To(Equal("eea38ba5-53a5-4173-9617-b442d35ec2fd"))
				return App{
					Name:       "App-Name",
					Guid:       appid,
					SpaceName:  "Space-Name",
					SpaceGuid:  "Space-Guid",
					OrgName:    "Org-Name",
					OrgGuid:    "Org-Guid",
					IgnoredApp: true,
				}
			}
			event.AnnotateWithAppData(caching)
			Expect(event.Fields["cf_app_name"]).To(Equal("App-Name"))
			Expect(event.Fields["cf_space_id"]).To(Equal("Space-Guid"))
			Expect(event.Fields["cf_space_name"]).To(Equal("Space-Name"))
			Expect(event.Fields["cf_org_id"]).To(Equal("Org-Guid"))
			Expect(event.Fields["cf_org_name"]).To(Equal("Org-Name"))
			Expect(event.Fields["cf_ignored_app"]).To(Equal(true))

		})

	})

})
