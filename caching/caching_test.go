package caching_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"time"

	. "github.com/cloudfoundry-community/firehose-to-syslog/caching"
	. "github.com/cloudfoundry-community/firehose-to-syslog/caching/cachingfakes"
	cfclient "github.com/cloudfoundry-community/go-cfclient"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Caching", func() {
	var (
		ignoreMissingApps  = false
		cacheInvalidateTTL = 3 * time.Second

		nilApp *App = nil

		config = &CacheLazyFillConfig{
			IgnoreMissingApps:  ignoreMissingApps,
			CacheInvalidateTTL: cacheInvalidateTTL,
			StripAppSuffixes:   []string{"-blue"},
		}

		client *FakeCFSimpleClient = nil
		store  *MemoryCacheStore   = nil
		cache  Caching             = nil
		gerr   error               = nil
	)

	BeforeEach(func() {
		client = new(FakeCFSimpleClient)
		client.DoGetStub = func(url string) (io.ReadCloser, error) {
			rv, ok := map[string]interface{}{
				"/v2/apps/d5945122-a4e3-11e8-9fc4-5f40e2ef7d66": &cfclient.AppResource{
					Meta: cfclient.Meta{
						Guid: "d5945122-a4e3-11e8-9fc4-5f40e2ef7d66",
					},
					Entity: cfclient.App{
						Name:      "bar",
						SpaceGuid: "e65ff164-a4e3-11e8-9577-7f76276dbf08",
					},
				},
				"/v2/spaces/e65ff164-a4e3-11e8-9577-7f76276dbf08": &cfclient.SpaceResource{
					Meta: cfclient.Meta{
						Guid: "e65ff164-a4e3-11e8-9577-7f76276dbf08",
					},
					Entity: cfclient.Space{
						Name:             "spacebar",
						OrganizationGuid: "27bd9bc4-a4e5-11e8-9374-d3160760cc79",
					},
				},
				"/v2/spaces/ce219ae2-a4e5-11e8-a91e-a32572b0821d": &cfclient.SpaceResource{
					Meta: cfclient.Meta{
						Guid: "ce219ae2-a4e5-11e8-a91e-a32572b0821d",
					},
					Entity: cfclient.Space{
						Name:             "spacefoo",
						OrganizationGuid: "05ef4d02-a4e6-11e8-875e-2f10395401c5",
					},
				},
				"/v2/organizations/27bd9bc4-a4e5-11e8-9374-d3160760cc79": &cfclient.OrgResource{
					Meta: cfclient.Meta{
						Guid: "27bd9bc4-a4e5-11e8-9374-d3160760cc79",
					},
					Entity: cfclient.Org{
						Name: "orgbar",
					},
				},
				"/v2/organizations/05ef4d02-a4e6-11e8-875e-2f10395401c5": &cfclient.OrgResource{
					Meta: cfclient.Meta{
						Guid: "05ef4d02-a4e6-11e8-875e-2f10395401c5",
					},
					Entity: cfclient.Org{
						Name: "orgfoo",
					},
				},
				"/v2/apps?results-per-page=100": &cfclient.AppResponse{
					Count: 1,
					Pages: 1,
					Resources: []cfclient.AppResource{
						{
							Meta: cfclient.Meta{
								Guid: "6b2a73ae-a4d2-11e8-ba76-c744cb2813fd",
							},
							Entity: cfclient.App{
								Name:      "foo-blue",
								SpaceGuid: "ce219ae2-a4e5-11e8-a91e-a32572b0821d",
							},
						},
					},
				},
			}[url]
			if !ok {
				return nil, fmt.Errorf("not found: %s", url)
			}

			rb := &bytes.Buffer{}
			err := json.NewEncoder(rb).Encode(rv)
			if err != nil {
				return nil, err
			}

			return ioutil.NopCloser(bytes.NewReader(rb.Bytes())), nil
		}

		store = &MemoryCacheStore{}
		cache = NewCacheLazyFill(client, store, config)

		gerr = store.Open()
		Ω(gerr).ShouldNot(HaveOccurred())
	})

	AfterEach(func() {
		gerr = store.Close()
		Ω(gerr).ShouldNot(HaveOccurred())
	})

	Context("Get app good case", func() {
		It("Expect app", func() {
			err := cache.FillCache()
			Ω(err).ShouldNot(HaveOccurred())

			guid := "6b2a73ae-a4d2-11e8-ba76-c744cb2813fd"
			app, err := cache.GetApp(guid)
			Ω(err).ShouldNot(HaveOccurred())

			Expect(app).NotTo(Equal(nil))
			Expect(app.Name).To(Equal("foo"))
			Expect(app.Guid).To(Equal(guid))
			Expect(app.SpaceName).To(Equal("spacefoo"))
			Expect(app.OrgName).To(Equal("orgfoo"))
		})
	})

	Context("Get app bad case", func() {
		It("Expect no app", func() {
			guid := "ca709384-a4d2-11e8-ad77-972bea7ff1b7" // does not exist
			app, err := cache.GetApp(guid)
			Ω(err).Should(HaveOccurred())
			Expect(app).To(Equal(nilApp))
		})
	})

	Context("Get app without pre-fill", func() {
		It("Expect app", func() {
			guid := "d5945122-a4e3-11e8-9fc4-5f40e2ef7d66"
			app, err := cache.GetApp(guid)
			Ω(err).ShouldNot(HaveOccurred())

			Expect(app).NotTo(Equal(nil))
			Expect(app.Name).To(Equal("bar"))
			Expect(app.Guid).To(Equal(guid))
			Expect(app.SpaceName).To(Equal("spacebar"))
			Expect(app.OrgName).To(Equal("orgbar"))
		})
	})
})
