package authclient_test

import (
	"errors"
	"github.com/cloudfoundry-community/firehose-to-syslog/authclient"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
	"net/http/httptest"
)

var _ = Describe("Http", func() {
	It("inserts an auth header for any calls it makes", func (){
		var token string
		h := http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
			token = r.Header.Get("Authorization")
		})
		s := httptest.NewServer(h)


		tf := newStubTokenFetcher()
		tf.token = "test-token"

		c := authclient.NewHttp(tf, "test-client", "test-secret", true)

		req, err := http.NewRequest(http.MethodGet, s.URL, nil)
		Expect(err).ToNot(HaveOccurred())
		c.Do(req)

		Expect(token).To(Equal("test-token"))
		Expect(tf.skipCertVerify).To(BeTrue())
	})

	It("fails when the token fetcher has an error", func() {
		s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

		tf := newStubTokenFetcher()
		tf.err = errors.New("an Error")
		c := authclient.NewHttp(tf, "test-client", "test-secret", true)

		req, err := http.NewRequest(http.MethodGet, s.URL, nil)
		Expect(err).ToNot(HaveOccurred())

		_, err = c.Do(req)
		Expect(err).To(HaveOccurred())
	})
})

type stubTokenFetcher struct {
	token string
	err error
	skipCertVerify bool
}

func newStubTokenFetcher() *stubTokenFetcher {
	return &stubTokenFetcher{}
}

func (s *stubTokenFetcher) GetAuthToken(username, password string, skipCertVerify bool) (string, error) {
	s.skipCertVerify = skipCertVerify
	return s.token, s.err
}
