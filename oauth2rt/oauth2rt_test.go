package oauth2rt

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Todo: test from _test bro
// Todo: dry this stuff up!!
// Todo: mock logger properly

type nopLogger struct{}

func (l *nopLogger) Info(ctx context.Context, msg string, kv ...any)  {}
func (l *nopLogger) Debug(ctx context.Context, msg string, kv ...any) {}

func TestOAuth2Rt(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "OAuth2Rt Suite")
}

var _ = Describe("OAuth2Rt", func() {

	Describe("tripperware", func() {

		var (
			rt       *OAuth2Rt
			mock     *mockRt
			request  *http.Request
			response *http.Response
			err      error
		)

		JustBeforeEach(func() {
			response, err = rt.RoundTrip(request)
		})

		Describe("adding bearer token", func() {

			When("all is well", func() {
				BeforeEach(func() {
					mock = &mockRt{
						TokenResponse: `{"access_token": "test-token-123"}`,
						APIStatus:     200,
					}

					rt = &OAuth2Rt{
						BaseUri:      "https://api.example.com",
						TokenPath:    "/api/oauth",
						ClientID:     "my-client",
						ClientSecret: "my-secret",
						Logger:       &nopLogger{},
					}
					rt.Wrap(mock)

					request, err = http.NewRequest("GET", "https://api.example.com/data", nil)
					Expect(err).ToNot(HaveOccurred())
				})

				It("fetches token and sets bearer header", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(response.StatusCode).To(Equal(200))
					Expect(mock.LastAuthHeader).To(Equal("Bearer test-token-123"))
					Expect(mock.TokenRequests).To(Equal(1))
				})
			})

			When("token is cached", func() {
				BeforeEach(func() {
					mock = &mockRt{
						TokenResponse: `{"access_token": "test-token-123"}`,
						APIStatus:     200,
					}

					rt = &OAuth2Rt{
						BaseUri:      "https://api.example.com",
						TokenPath:    "/api/oauth",
						ClientID:     "my-client",
						ClientSecret: "my-secret",
						Logger:       &nopLogger{},
						token:        "cached-token",
					}
					rt.Wrap(mock)

					request, err = http.NewRequest("GET", "https://api.example.com/data", nil)
					Expect(err).ToNot(HaveOccurred())
				})

				It("uses cached token without fetching", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(mock.LastAuthHeader).To(Equal("Bearer cached-token"))
					Expect(mock.TokenRequests).To(Equal(0))
				})
			})

			When("api returns 401", func() {
				BeforeEach(func() {
					mock = &mockRt{
						TokenResponse:  `{"access_token": "fresh-token"}`,
						APIStatus:      401,
						RetryAPIStatus: 200,
					}

					rt = &OAuth2Rt{
						BaseUri:      "https://api.example.com",
						TokenPath:    "/api/oauth",
						ClientID:     "my-client",
						ClientSecret: "my-secret",
						Logger:       &nopLogger{},
						token:        "stale-token",
					}
					rt.Wrap(mock)

					request, err = http.NewRequest("GET", "https://api.example.com/data", nil)
					Expect(err).ToNot(HaveOccurred())
				})

				It("clears cache, fetches new token, and retries", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(response.StatusCode).To(Equal(200))
					Expect(mock.LastAuthHeader).To(Equal("Bearer fresh-token"))
					Expect(mock.TokenRequests).To(Equal(1))
					Expect(mock.APIRequests).To(Equal(2))
				})
			})

			When("api returns 401 on POST with body", func() {
				BeforeEach(func() {
					mock = &mockRt{
						TokenResponse:  `{"access_token": "fresh-token"}`,
						APIStatus:      401,
						RetryAPIStatus: 200,
					}

					rt = &OAuth2Rt{
						BaseUri:      "https://api.example.com",
						TokenPath:    "/api/oauth",
						ClientID:     "my-client",
						ClientSecret: "my-secret",
						Logger:       &nopLogger{},
						token:        "stale-token",
					}
					rt.Wrap(mock)

					request, err = http.NewRequest("POST", "https://api.example.com/data", strings.NewReader(`{"foo":"bar"}`))
					Expect(err).ToNot(HaveOccurred())
				})

				It("preserves body on retry", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(response.StatusCode).To(Equal(200))
					Expect(mock.APIRequests).To(Equal(2))
					Expect(mock.LastBody).To(Equal(`{"foo":"bar"}`))
				})
			})
		})
	})
})

// mockRt simulates both token endpoint and API responses
type mockRt struct {
	TokenResponse  string
	APIStatus      int
	RetryAPIStatus int

	TokenRequests  int
	APIRequests    int
	LastAuthHeader string
	LastBody       string
}

func (rt *mockRt) RoundTrip(req *http.Request) (*http.Response, error) {

	// token endpoint
	if req.URL.Path == "/api/oauth" {
		rt.TokenRequests++
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(rt.TokenResponse)),
			Request:    req,
		}, nil
	}

	// api endpoint
	rt.APIRequests++
	rt.LastAuthHeader = req.Header.Get("Authorization")
	if req.Body != nil {
		body, _ := io.ReadAll(req.Body)
		rt.LastBody = string(body)
	}

	status := rt.APIStatus
	if rt.APIRequests > 1 && rt.RetryAPIStatus != 0 {
		status = rt.RetryAPIStatus
	}

	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(`{"data": "ok"}`)),
		Request:    req,
	}, nil
}
