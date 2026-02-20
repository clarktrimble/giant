package giant

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/clarktrimble/giant/logrt"
	"github.com/clarktrimble/launch"
)

//go:generate moq -pkg giant -out mock_test.go ./logger Logger

func TestGiant(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Giant Suite")
}

var _ = Describe("Giant", func() {

	Describe("creating a client with trippers", func() {
		var (
			cfg *Config
			lgr *LoggerMock
			gnt *Giant
		)

		JustBeforeEach(func() {
			gnt = cfg.NewWithTrippers(lgr)
		})

		When("all is well", func() {
			BeforeEach(func() {
				cfg = &Config{
					BaseUri:       "https://api.open-meteo.com",
					Headers:       []string{"X-Authorization-Token", "this-is-secret", "bargle"},
					RedactHeaders: []string{"X-Authorization-TOKEN"},
				}

				lgr = &LoggerMock{}
			})

			It("creates a client with LogRt as first transport", func() {
				Expect(gnt.BaseUri).To(Equal("https://api.open-meteo.com"))
				Expect(gnt.Headers).To(Equal(map[string]string{"X-Authorization-Token": "this-is-secret"}))

				logRt, ok := gnt.Client.Transport.(*logrt.LogRt)
				Expect(ok).To(BeTrue())
				Expect(logRt.RedactHeaders).To(Equal(map[string]bool{"Authorization": true, "X-Authorization-Token": true}))
			})
		})

		When("OAuth2 is configured", func() {
			var (
				authServer *oauthTestServer
				apiServer  *oauthTestServer
			)

			BeforeEach(func() {
				authServer = newOauthTestServer("test-token-123")
				apiServer = newOauthTestServer("")

				lgr = &loggerMock{
					InfoFunc:       func(ctx context.Context, msg string, kv ...any) {},
					DebugFunc:      func(ctx context.Context, msg string, kv ...any) {},
					WithFieldsFunc: func(ctx context.Context, kv ...any) context.Context { return ctx },
				}
			})

			AfterEach(func() {
				authServer.Server.Close()
				apiServer.Server.Close()
			})

			When("OAuth2 has its own BaseUri", func() {
				BeforeEach(func() {
					cfg = &Config{
						BaseUri: apiServer.Server.URL,
						OAuth2: &OAuth2Config{
							BaseUri:      authServer.Server.URL,
							TokenPath:    "/oauth/token",
							ClientID:     "test-client",
							ClientSecret: launch.Redact("test-secret"),
						},
					}
				})

				It("fetches token from auth server and sets Authorization header", func() {
					ctx := context.Background()
					_, err := gnt.SendJson(ctx, "GET", "/api/resource", nil)

					Expect(err).ToNot(HaveOccurred())
					Expect(authServer.TokenRequested).To(BeTrue())
					Expect(apiServer.AuthHeader).To(Equal("Bearer test-token-123"))
				})
			})

			When("OAuth2 BaseUri is empty", func() {
				BeforeEach(func() {
					// single server handles both token and api
					authServer = newOauthTestServer("fallback-token-456")

					cfg = &Config{
						BaseUri: authServer.Server.URL,
						OAuth2: &OAuth2Config{
							TokenPath:    "/oauth/token",
							ClientID:     "test-client",
							ClientSecret: launch.Redact("test-secret"),
						},
					}
				})

				It("uses Config.BaseUri for token endpoint", func() {
					ctx := context.Background()
					_, err := gnt.SendJson(ctx, "GET", "/api/resource", nil)

					Expect(err).ToNot(HaveOccurred())
					Expect(authServer.TokenRequested).To(BeTrue())
					Expect(authServer.AuthHeader).To(Equal("Bearer fallback-token-456"))
				})
			})
		})
	})

	Describe("making requests", func() {
		var (
			ts  *testServer
			gnt *Giant
			ctx context.Context
			err error

			method string
			path   string
		)

		BeforeEach(func() {
			ts = newTestServer(`{"data": "thing2"}`)
			gnt = &Giant{
				Client:  http.Client{},
				BaseUri: ts.Server.URL,
			}
			ctx = context.Background()
		})

		Describe("sending object", func() {
			var (
				sndObj foo
				rcvObj *foo
			)

			BeforeEach(func() {
				method = "PUT"
				path = "/posts/"
				sndObj = foo{Data: "stuff"}
			})

			JustBeforeEach(func() {
				err = gnt.SendObject(ctx, method, path, sndObj, rcvObj)
			})

			When("both send and receive are specified", func() {
				BeforeEach(func() {
					rcvObj = &foo{}
				})

				It("marshalls send and unmarshalls receive", func() {
					Expect(err).ToNot(HaveOccurred())

					Expect(ts.ContentHeader).To(Equal("application/json"))
					Expect(ts.Method).To(Equal("PUT"))
					Expect(ts.Path).To(Equal("/posts/"))
					Expect(ts.Body).To(Equal(`{"data":"stuff"}`))

					Expect(rcvObj).To(Equal(&foo{Data: "thing2"}))
				})
			})

			When("rcvObj is nil", func() {
				BeforeEach(func() {
					rcvObj = nil
				})
				It("skips unmarshalling receive", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(rcvObj).To(BeNil())
				})
			})
		})

		Describe("sending json", func() {
			var (
				body io.Reader
				data []byte
			)

			JustBeforeEach(func() {
				data, err = gnt.SendJson(ctx, method, path, body)
			})

			When("all is well", func() {
				BeforeEach(func() {
					method = "PATCH"
					path = "/posts/"
					body = bytes.NewBufferString(`{"data": "about a post"}`)
				})
				It("sends a request and returns data", func() {
					Expect(err).ToNot(HaveOccurred())

					Expect(ts.ContentHeader).To(Equal("application/json"))
					Expect(ts.Method).To(Equal("PATCH"))
					Expect(ts.Path).To(Equal("/posts/"))
					Expect(ts.Body).To(Equal(`{"data": "about a post"}`))

					Expect(string(data)).To(Equal(`{"data": "thing2"}`))
				})
			})
		})

		Describe("sending a request", func() {

			var (
				rq       Request
				response *http.Response
			)

			JustBeforeEach(func() {
				response, err = gnt.Send(ctx, rq)
			})

			When("request is blank", func() {
				BeforeEach(func() {
					rq = Request{}
				})
				It("makes a default request and returns response", func() {
					Expect(err).ToNot(HaveOccurred())

					Expect(ts.ContentHeader).To(Equal(""))
					Expect(ts.Method).To(Equal("GET"))
					Expect(ts.Path).To(Equal("/"))
					Expect(ts.Body).To(Equal(""))

					body, err := io.ReadAll(response.Body)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(body)).To(Equal(`{"data": "thing2"}`))
				})
			})

			When("request is fully specified", func() {
				BeforeEach(func() {
					rq = Request{
						Method: "POST",
						Path:   "/posts/",
						Body:   bytes.NewBufferString(`{"data": "about a post"}`),
						Headers: map[string]string{
							"Content-Type": "application/json",
							"Accept":       "application/json",
						},
					}
					gnt.Headers = map[string]string{
						"ForThe": "Win",
					}
				})
				It("makes a matching request", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(ts.ContentHeader).To(Equal("application/json"))
					Expect(ts.FtwHeader).To(Equal("Win"))
					Expect(ts.Method).To(Equal("POST"))
					Expect(ts.Path).To(Equal("/posts/"))
					Expect(ts.Body).To(Equal(`{"data": "about a post"}`))
				})
			})

		})
	})

})

// help

type foo struct {
	Data string `json:"data"`
}

type testServer struct {
	Server        *httptest.Server
	ContentHeader string
	FtwHeader     string
	Method        string
	Path          string
	Body          string
}

func newTestServer(responseBody string) (ts *testServer) {

	ts = &testServer{}

	ts.Server = httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

		body, err := io.ReadAll(request.Body)
		Expect(err).ToNot(HaveOccurred())

		ts.ContentHeader = request.Header.Get("Content-Type")
		ts.FtwHeader = request.Header.Get("ForThe")
		ts.Method = request.Method
		ts.Path = request.RequestURI
		ts.Body = string(body)

		fmt.Fprint(writer, responseBody)
	}))

	return
}

type oauthTestServer struct {
	Server         *httptest.Server
	Token          string
	TokenRequested bool
	AuthHeader     string
}

func newOauthTestServer(token string) (ots *oauthTestServer) {

	ots = &oauthTestServer{Token: token}

	ots.Server = httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

		if request.URL.Path == "/oauth/token" {
			ots.TokenRequested = true
			resp := map[string]string{"access_token": ots.Token}
			writer.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(writer).Encode(resp)
			Expect(err).ToNot(HaveOccurred())
			return
		}

		ots.AuthHeader = request.Header.Get("Authorization")
		fmt.Fprint(writer, `{"status": "ok"}`)
	}))

	return
}
