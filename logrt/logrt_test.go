package logrt

import (
	"context"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

//go:generate moq -out mock_test.go . logger

func TestLogRt(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "LogRt Suite")
}

var _ = Describe("LogRt", func() {

	Describe("tripperware", func() {

		var (
			rt       *LogRt
			request  *http.Request
			response *http.Response
			ctx      context.Context

			lgr *loggerMock
			err error
		)

		BeforeEach(func() {
			lgr = &loggerMock{
				DebugFunc: func(ctx context.Context, msg string, kv ...any) {},
				WithFieldsFunc: func(ctx context.Context, kv ...any) context.Context {
					return ctx
				},
			}

			rt = New(lgr, []string{"X-Authorization-Token"}, false)
			rt.Wrap(&testRt{
				Status: 200,
			})

			rand.Seed(1) //nolint:staticcheck // predictable request_id

			request, err = http.NewRequest("PUT", "https://boxworld.org/cardboard", nil)
			Expect(err).ToNot(HaveOccurred())
			request.Header.Set("content-type", "application/json")
			request.Header.Set("X-Authorization-Token", "this-is-secret")
			request.Header.Set("Authorization", "this-is-also-secret")

			ctx = request.Context()
		})

		JustBeforeEach(func() {
			response, err = rt.RoundTrip(request)
		})

		Describe("logging request and response", func() {

			When("all is well", func() {
				It("logs the request and the response", func() {

					Expect(err).ToNot(HaveOccurred())

					wfc := lgr.WithFieldsCalls()
					Expect(wfc).To(HaveLen(1))
					Expect(wfc[0].Ctx).To(Equal(ctx))
					Expect(wfc[0].Kv).To(HaveLen(2))
					Expect(wfc[0].Kv[0]).To(Equal("request_id"))
					Expect(wfc[0].Kv[1]).To(HaveLen(7))

					ic := lgr.DebugCalls()
					Expect(ic).To(HaveLen(2))
					Expect(ic[0].Msg).To(Equal("sending request"))
					Expect(ic[0].Kv).To(HaveExactElements(
						"method",
						"PUT",
						"scheme",
						"https",
						"host",
						"boxworld.org",
						"path",
						"/cardboard",
						"headers",
						http.Header{
							"Content-Type":          []string{"application/json"},
							"X-Authorization-Token": []string{"--redacted--"},
							"Authorization":         []string{"--redacted--"},
						},
						"query",
						url.Values{},
						"body",
						"",
					))

					Expect(ic[1].Msg).To(Equal("received response"))
					// check elapsed and set to zero for next check
					Expect(ic[1].Kv[5]).To(BeNumerically(">", 0))
					ic[1].Kv[5] = 0

					Expect(ic[1].Kv).To(HaveExactElements(
						"status",
						200,
						"headers",
						http.Header(nil),
						"elapsed",
						0,
						"path",
						"/cardboard",
						"body",
						`{"ima": "pc"}`,
					))

					body, err := io.ReadAll(response.Body)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(body)).To(Equal(`{"ima": "pc"}`))
				})
			})

			When("skipping body", func() {
				BeforeEach(func() {
					rt.SkipBody = true
				})

				It("logs the request and the response sans body", func() {

					// copy from "all is well" above, minus body fields :/

					Expect(err).ToNot(HaveOccurred())

					wfc := lgr.WithFieldsCalls()
					Expect(wfc).To(HaveLen(1))
					Expect(wfc[0].Ctx).To(Equal(ctx))
					Expect(wfc[0].Kv).To(HaveLen(2))
					Expect(wfc[0].Kv[0]).To(Equal("request_id"))
					Expect(wfc[0].Kv[1]).To(HaveLen(7))

					ic := lgr.DebugCalls()
					Expect(ic).To(HaveLen(2))
					Expect(ic[0].Msg).To(Equal("sending request"))
					Expect(ic[0].Kv).To(HaveExactElements(
						"method",
						"PUT",
						"scheme",
						"https",
						"host",
						"boxworld.org",
						"path",
						"/cardboard",
						"headers",
						http.Header{
							"Content-Type":          []string{"application/json"},
							"X-Authorization-Token": []string{"--redacted--"},
							"Authorization":         []string{"--redacted--"},
						},
						"query",
						url.Values{},
					))

					Expect(ic[1].Msg).To(Equal("received response"))
					// check elapsed and set to zero for next check
					Expect(ic[1].Kv[5]).To(BeNumerically(">", 0))
					ic[1].Kv[5] = 0

					Expect(ic[1].Kv).To(HaveExactElements(
						"status",
						200,
						"headers",
						http.Header(nil),
						"elapsed",
						0,
						"path",
						"/cardboard",
					))

					body, err := io.ReadAll(response.Body)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(body)).To(Equal(`{"ima": "pc"}`))
				})
			})
		})

	})
})

type testRt struct {
	Status int
}

func (rt *testRt) RoundTrip(request *http.Request) (response *http.Response, err error) {

	response = &http.Response{
		StatusCode: rt.Status,
		Body:       io.NopCloser(strings.NewReader(`{"ima": "pc"}`)),
		Request:    request,
	}

	return
}
