package caching_test

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	. "github.com/cloudfoundry-community/firehose-to-syslog/caching"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	cfclient "github.com/cloudfoundry-community/go-cfclient"
)

type mockAppClient struct {
	lock sync.RWMutex
	apps map[string]cfclient.App
	n    int
}

func newMockAppClient(n int) *mockAppClient {
	apps := getApps(n)
	return &mockAppClient{
		apps: apps,
		n:    n,
	}
}

func (m *mockAppClient) AppByGuid(guid string) (cfclient.App, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	app, ok := m.apps[guid]
	if ok {
		return app, nil
	}
	return app, errors.New("No such app")
}

func (m *mockAppClient) ListOrgs() ([]cfclient.Org, error) { return []cfclient.Org{}, nil }
func (m *mockAppClient) OrgSpaces(org string) ([]cfclient.Space, error) {
	return []cfclient.Space{}, nil
}

func (m *mockAppClient) CreateApp(appID, spaceID, orgID string) {
	m.lock.Lock()
	defer m.lock.Unlock()

	app := cfclient.App{
		Guid: appID,
		Name: appID,
		SpaceData: cfclient.SpaceResource{
			Entity: cfclient.Space{
				Guid: spaceID,
				Name: spaceID,
				OrgData: cfclient.OrgResource{
					Entity: cfclient.Org{
						Guid: orgID,
						Name: orgID,
					},
				},
			},
		},
	}

	m.apps[appID] = app
}

func getApps(n int) map[string]cfclient.App {
	apps := make(map[string]cfclient.App, n)
	for i := 0; i < n; i++ {
		app := cfclient.App{
			Guid: fmt.Sprintf("cf_app_id_%d", i),
			Name: fmt.Sprintf("cf_app_name_%d", i),
			SpaceData: cfclient.SpaceResource{
				Entity: cfclient.Space{
					Guid: fmt.Sprintf("cf_space_id_%d", i%50),
					Name: fmt.Sprintf("cf_space_name_%d", i%50),
					OrgData: cfclient.OrgResource{
						Entity: cfclient.Org{
							Guid: fmt.Sprintf("cf_org_id_%d", i%100),
							Name: fmt.Sprintf("cf_org_name_%d", i%100),
						},
					},
				},
			},
		}
		apps[app.Guid] = app
	}
	return apps
}

// Little bit hacky test.
// To preload in cache we just look for every apps first.
func (m *mockAppClient) ingestApps(cache Caching) {
	for _, app := range m.apps {
		cache.GetApp(app.Guid)
	}
}

var _ = Describe("Caching", func() {
	var (
		boltdbPath         = "/tmp/boltdb"
		ignoreMissingApps  = true
		cacheInvalidateTTL = 3 * time.Second
		n                  = 10

		nilApp *App = nil

		config = &CachingBoltConfig{
			Path:               boltdbPath,
			IgnoreMissingApps:  ignoreMissingApps,
			CacheInvalidateTTL: cacheInvalidateTTL,
			RequestBySec:       50,
		}

		client *mockAppClient = nil
		cache  *CachingBolt   = nil
		gerr   error          = nil
	)

	BeforeEach(func() {
		os.Remove(boltdbPath)
		client = newMockAppClient(n)
		cache, gerr = NewCachingBolt(client, config)

		Ω(gerr).ShouldNot(HaveOccurred())

		gerr = cache.Open()
		Ω(gerr).ShouldNot(HaveOccurred())
		client.ingestApps(cache)
	})

	AfterEach(func() {
		gerr = cache.Close()
		Ω(gerr).ShouldNot(HaveOccurred())

		time.Sleep(1 * time.Second)
		gerr = os.Remove(boltdbPath)
		Ω(gerr).ShouldNot(HaveOccurred())
	})

	Context("Get app good case", func() {
		It("Have 10 apps", func() {
			client.ingestApps(cache)
			apps, err := cache.GetAllApps()

			Ω(err).ShouldNot(HaveOccurred())

			Expect(apps).NotTo(Equal(nil))
			Expect(len(apps)).To(Equal(n))
		})

		It("Expect app", func() {
			guid := "cf_app_id_0"
			app, err := cache.GetApp(guid)
			Ω(err).ShouldNot(HaveOccurred())

			Expect(app).NotTo(Equal(nil))
			Expect(app.Guid).To(Equal(guid))
		})
	})

	Context("Get app bad case", func() {
		It("Expect no app", func() {
			guid := "cf_app_id_not_exists"
			app, err := cache.GetApp(guid)
			Ω(err).Should(HaveOccurred())
			Expect(app).To(Equal(nilApp))

			// We ignore missing apps, so for the second time query, we already
			// recorded the missing app, so nil, err is expected to return
			app, err = cache.GetApp(guid)
			Ω(err).Should(HaveOccurred())
			Expect(app).To(Equal(nilApp))
		})
	})

	Context("Cache invalidation", func() {
		It("Expect new app", func() {
			id := fmt.Sprintf("id_%d", time.Now().UnixNano())
			client.CreateApp(id, id, id)

			// Sleep for CacheInvalidateTTL interval to make sure the cache
			// invalidation happens
			time.Sleep(cacheInvalidateTTL + 1)

			app, err := cache.GetApp(id)
			Ω(err).ShouldNot(HaveOccurred())
			Expect(app).NotTo(Equal(nilApp))
			Expect(app.Guid).To(Equal(id))

			apps, err := cache.GetAllApps()
			Ω(err).ShouldNot(HaveOccurred())

			Expect(apps).NotTo(Equal(nil))
			Expect(len(apps)).To(Equal(n + 1))
		})
	})

	Context("NewCachingBolt error", func() {
		It("Expect error", func() {
			dup := *config
			dup.Path = fmt.Sprintf("/not-exists-%d/boltdb", time.Now().UnixNano())
			bcache, err := NewCachingBolt(client, &dup)
			Ω(err).ShouldNot(HaveOccurred())

			err = bcache.Open()
			Ω(err).Should(HaveOccurred())
		})
	})

	Context("Load from existing boltdb", func() {
		It("Expect 10 apps from existing boltdb", func() {
			dup := *config
			dup.Path = fmt.Sprintf("/tmp/%d", time.Now().UnixNano())
			bcache, err := NewCachingBolt(client, &dup)

			Ω(err).ShouldNot(HaveOccurred())

			err = bcache.Open()
			Ω(err).ShouldNot(HaveOccurred())
			client.ingestApps(bcache)

			defer os.Remove(dup.Path)
			time.Sleep(time.Second)
			bcache.Close()

			// Load from existing db
			bcache, err = NewCachingBolt(client, &dup)
			Ω(err).ShouldNot(HaveOccurred())

			err = bcache.Open()
			Ω(err).ShouldNot(HaveOccurred())

			apps, err := bcache.GetAllApps()
			Ω(err).ShouldNot(HaveOccurred())

			Expect(apps).NotTo(Equal(nil))
			Expect(len(apps)).To(Equal(n))
		})
	})
})
