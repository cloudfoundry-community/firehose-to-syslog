package uaago_test

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"

	"io/ioutil"

	"github.com/cloudfoundry-incubator/uaago"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type request struct {
	Request *http.Request
	Body    []byte
}

var _ = Describe("Client", func() {
	Context("GetOauthToken", func() {
		Context("with https", func() {
			var testServer *httptest.Server
			BeforeEach(func() {
				testServer = httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
					if validRequest(request) {
						authValue := request.Header.Get("Authorization")
						authValueBytes := "Basic " + base64.StdEncoding.EncodeToString([]byte("myusername:mypassword"))
						if authValueBytes == authValue {
							jsonData := []byte(`
						{
							"access_token":"good-token",
							"token_type":"bearer",
							"expires_in":599,
							"scope":"cloud_controller.write doppler.firehose",
							"jti":"28edda5c-4e37-4a63-9ba3-b32f48530a51"
						}
						`)
							writer.Write(jsonData)
							return
						}
					}
					writer.WriteHeader(http.StatusUnauthorized)
				}))
			})
			AfterEach(func() {
				testServer.Close()
			})
			It("Should get a valid oauth token from the given UAA", func() {
				client, err := uaago.NewClient(testServer.URL)
				Expect(err).ToNot(HaveOccurred())

				token, err := client.GetAuthToken("myusername", "mypassword", true)
				Expect(err).ToNot(HaveOccurred())
				Expect(token).To(Equal("bearer good-token"))
			})
		})
	})

	Context("GetOauthToken With Expires_in", func() {
		Context("with https", func() {
			var testServer *httptest.Server
			BeforeEach(func() {
				testServer = httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
					if validRequest(request) {
						authValue := request.Header.Get("Authorization")
						authValueBytes := "Basic " + base64.StdEncoding.EncodeToString([]byte("myusername:mypassword"))
						if authValueBytes == authValue {
							jsonData := []byte(`
						{
							"access_token":"good-token",
							"token_type":"bearer",
							"expires_in":598,
							"scope":"cloud_controller.write doppler.firehose",
							"jti":"28edda5c-4e37-4a63-9ba3-b32f48530a51"
						}
						`)
							writer.Write(jsonData)
							return
						}
					}
					writer.WriteHeader(http.StatusUnauthorized)
				}))
			})
			AfterEach(func() {
				testServer.Close()
			})
			It("Should get a valid oauth token and expires_in from the given UAA", func() {
				client, err := uaago.NewClient(testServer.URL)
				Expect(err).ToNot(HaveOccurred())

				token, expiresIn, err := client.GetAuthTokenWithExpiresIn("myusername", "mypassword", true)
				Expect(err).ToNot(HaveOccurred())
				Expect(token).To(Equal("bearer good-token"))
				Expect(expiresIn).To(Equal(598))
			})
		})

		Context("with invalid expires_in", func() {
			var testServer *httptest.Server
			BeforeEach(func() {
				testServer = httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
					if validRequest(request) {
						authValue := request.Header.Get("Authorization")
						authValueBytes := "Basic " + base64.StdEncoding.EncodeToString([]byte("myusername:mypassword"))
						if authValueBytes == authValue {
							jsonData := []byte(`
						{
							"access_token":"good-token",
							"token_type":"bearer",
							"expires_in":"invalid",
							"scope":"cloud_controller.write doppler.firehose",
							"jti":"28edda5c-4e37-4a63-9ba3-b32f48530a51"
						}
						`)
							writer.Write(jsonData)
							return
						}
					}
					writer.WriteHeader(http.StatusUnauthorized)
				}))
			})
			AfterEach(func() {
				testServer.Close()
			})
			It("Should get a valid oauth token and expires_in from the given UAA", func() {
				client, err := uaago.NewClient(testServer.URL)
				Expect(err).ToNot(HaveOccurred())

				token, expiresIn, err := client.GetAuthTokenWithExpiresIn("myusername", "mypassword", true)
				Expect(err).To(HaveOccurred())
				Expect(token).To(Equal(""))
				Expect(expiresIn).To(Equal(-1))
			})
		})

		Context("without expires_in", func() {
			var testServer *httptest.Server
			BeforeEach(func() {
				testServer = httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
					if validRequest(request) {
						authValue := request.Header.Get("Authorization")
						authValueBytes := "Basic " + base64.StdEncoding.EncodeToString([]byte("myusername:mypassword"))
						if authValueBytes == authValue {
							jsonData := []byte(`
						{
							"access_token":"good-token",
							"token_type":"bearer",
							"scope":"cloud_controller.write doppler.firehose",
							"jti":"28edda5c-4e37-4a63-9ba3-b32f48530a51"
						}
						`)
							writer.Write(jsonData)
							return
						}
					}
					writer.WriteHeader(http.StatusUnauthorized)
				}))
			})
			AfterEach(func() {
				testServer.Close()
			})
			It("Should get a valid oauth token missing the expires_in from the given UAA", func() {
				client, err := uaago.NewClient(testServer.URL)
				Expect(err).ToNot(HaveOccurred())

				token, expiresIn, err := client.GetAuthTokenWithExpiresIn("myusername", "mypassword", false)
				Expect(err).ToNot(HaveOccurred())
				Expect(token).To(Equal("bearer good-token"))
				Expect(expiresIn).To(Equal(0))
			})
		})
	})

	Context("TokenIsAuthorized", func() {
		var (
			uaaTestServer     *httptest.Server
			uaaRequests       = make(chan *request, 10)
			uaaResponseBodies = make(chan string, 10)
			client            *uaago.Client

			basicAuthUser = "some-basic-auth-user"
			basicAuthPass = "some-basic-auth-pass"
		)

		BeforeEach(func() {
			uaaTestServer = httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
				body, _ := ioutil.ReadAll(req.Body)
				req.Body.Close()
				uaaRequests <- &request{Request: req, Body: body}

				writer.WriteHeader(http.StatusOK)
				writer.Write([]byte(<-uaaResponseBodies))
			}))
			client, _ = uaago.NewClient(uaaTestServer.URL)
		})

		AfterEach(func() {
			uaaTestServer.Close()
		})

		It("talks to UAA", func() {
			uaaResponseBodies <- "some client_id"
			client.TokenIsAuthorized(basicAuthUser, basicAuthPass, "some token", "some client_id", false)
			var req *request
			Eventually(uaaRequests).Should(Receive(&req))

			Expect(req.Request.Method).To(Equal("POST"))
			Expect(req.Request.URL).To(ContainSubstring("/check_token"))

			Expect(string(req.Body)).To(ContainSubstring("some token"))

			authUser, authPass, ok := req.Request.BasicAuth()
			Expect(ok).To(BeTrue())
			Expect(authUser).To(Equal(basicAuthUser))
			Expect(authPass).To(Equal(basicAuthPass))
		})

		It("returns an error for non-200 http responses", func() {
			unauthServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(401)
			}))
			client, _ = uaago.NewClient(unauthServer.URL)
			defer unauthServer.Close()

			_, err := client.TokenIsAuthorized("some-user", "some-password", "some-token", "some-client-id", false)
			Expect(err).To(HaveOccurred())
		})

		Context("valid: client_id=ingestor", func() {
			It("returns true", func() {
				uaaResponseBodies <- "ingestor"
				isValid, err := client.TokenIsAuthorized(basicAuthUser, basicAuthPass, "some token", "ingestor", false)

				Expect(err).ToNot(HaveOccurred())
				Expect(isValid).To(BeTrue())
			})
		})

		Context("invalid: client_id=foo", func() {
			It("returns false", func() {
				uaaResponseBodies <- "foo"
				isValid, err := client.TokenIsAuthorized(basicAuthUser, basicAuthPass, "some token", "ingestor", false)

				Expect(err).ToNot(HaveOccurred())
				Expect(isValid).To(BeFalse())
			})
		})
	})

	Context("GetRefreshToken", func() {
		Context("with https", func() {
			var testServer *httptest.Server
			It("should get a valid refresh and access token from the given UAA", func() {
				testServer = httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
					if validRequest(request) {
						Expect(request.PostForm.Get("refresh_token")).To(Equal("refresh-token"))
						Expect(request.PostForm.Get("client_id")).To(Equal("some-client-id"))
						Expect(request.PostForm.Get("grant_type")).To(Equal("refresh_token"))
						jsonData := []byte(`
						{
							"access_token":"some-access-token",
							"token_type":"bearer",
							"refresh_token" : "some-new-token-r",
							"expires_in" : 43199,
							"scope":"cloud_controller.write doppler.firehose",
							"jti":"28edda5c-4e37-4a63-9ba3-b32f48530a51"
						}
						`)
						writer.Write(jsonData)
						return
					}
				}))

				client, err := uaago.NewClient(testServer.URL)
				Expect(err).ToNot(HaveOccurred())

				refreshToken, accessToken, err := client.GetRefreshToken("some-client-id", "refresh-token", true)
				Expect(err).ToNot(HaveOccurred())
				Expect(refreshToken).To(Equal("some-new-token-r"))
				Expect(accessToken).To(Equal("bearer some-access-token"))

				testServer.Close()
			})

			It("should return an error for non-200 response from UAA", func() {
				unauthServer := httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
					writer.WriteHeader(401)
				}))
				client, _ := uaago.NewClient(unauthServer.URL)
				defer unauthServer.Close()

				_, _, err := client.GetRefreshToken("some-client-id", "invalid-token", true)
				Expect(err).To(HaveOccurred())
			})

			It("should return an error if refresh_token is not populated", func() {
				testServer = httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
					jsonData := []byte(`
						{
							"access_token":"good-token",
							"token_type":"bearer",
							"expires_in" : 43199,
							"scope":"cloud_controller.write doppler.firehose",
							"jti":"28edda5c-4e37-4a63-9ba3-b32f48530a51"
						}
						`)
					writer.Write(jsonData)
				}))

				client, err := uaago.NewClient(testServer.URL)
				Expect(err).ToNot(HaveOccurred())

				_, _, err = client.GetRefreshToken("some-client-id", "refresh-token", true)
				Expect(err).To(MatchError("Missing refresh_token in response body"))
			})

			It("should return an error if token_type is not populated", func() {
				testServer = httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
					jsonData := []byte(`
						{
							"access_token":"good-token",
							"refresh_token": "refresh-token",
							"expires_in" : 43199,
							"scope":"cloud_controller.write doppler.firehose",
							"jti":"28edda5c-4e37-4a63-9ba3-b32f48530a51"
						}
						`)
					writer.Write(jsonData)
				}))

				client, err := uaago.NewClient(testServer.URL)
				Expect(err).ToNot(HaveOccurred())

				_, _, err = client.GetRefreshToken("some-client-id", "refresh-token", true)
				Expect(err).To(MatchError("Missing token_type in response body"))
			})

			It("should return an error if access_token is not populated", func() {
				testServer = httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
					jsonData := []byte(`
						{
							"token_type":"bearer",
							"refresh_token" : "some-new-token-r",
							"expires_in" : 43199,
							"scope":"cloud_controller.write doppler.firehose",
							"jti":"28edda5c-4e37-4a63-9ba3-b32f48530a51"
						}
						`)
					writer.Write(jsonData)
				}))

				client, err := uaago.NewClient(testServer.URL)
				Expect(err).ToNot(HaveOccurred())

				_, _, err = client.GetRefreshToken("some-client-id", "refresh-token", true)
				Expect(err).To(MatchError("Missing access_token in response body"))
			})
		})
	})
})

func validRequest(request *http.Request) bool {
	isPost := request.Method == "POST"
	correctPath := request.URL.Path == "/oauth/token"
	correctType := request.Header.Get("content-type") == "application/x-www-form-urlencoded"
	request.ParseForm()
	hasClientId := len(request.PostForm.Get("client_id")) > 0
	hasGrantType := len(request.PostForm.Get("grant_type")) > 0

	return isPost && correctPath && correctType && hasClientId && hasGrantType
}
