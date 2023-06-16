package logrt_test

import (
	"io"
	"net/http"
	"net/url"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/clarktrimble/giant/logrt"
	"github.com/clarktrimble/giant/mock"
)

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

			lgr *mock.Logger
			err error
		)

		JustBeforeEach(func() {
			response, err = rt.RoundTrip(request)
		})

		Describe("logging request and response", func() {

			When("all is well", func() {
				BeforeEach(func() {
					lgr = mock.NewLogger()
					rt = &LogRt{
						Logger: lgr,
					}
					rt.Wrap(&mock.TestRt{
						Status: 200,
					})

					request, err = http.NewRequest("PUT", "https://boxworld.org/cardboard", nil)
					Expect(err).ToNot(HaveOccurred())
				})
				It("logs the request and the response", func() {

					Expect(err).ToNot(HaveOccurred())
					Expect(lgr.Logged).To(HaveLen(2))

					Expect(lgr.Logged[0]).To(HaveKey("request_id"))
					Expect(lgr.Logged[0]["request_id"]).To(HaveLen(7))
					rid := lgr.Logged[0]["request_id"]

					Expect(lgr.Logged[0]).To(Equal(map[string]any{
						"body":       "",
						"headers":    http.Header{},
						"host":       "boxworld.org",
						"method":     "PUT",
						"msg":        "sending request",
						"path":       "/cardboard",
						"query":      url.Values{},
						"request_id": rid,
						"scheme":     "https",
					}))

					Expect(lgr.Logged[1]).To(HaveKey("elapsed"))
					lgr.Logged[1]["elapsed"] = 0

					Expect(lgr.Logged[1]).To(Equal(map[string]any{
						"body":       `{"ima": "pc"}`,
						"elapsed":    0,
						"headers":    http.Header(nil),
						"msg":        "received response",
						"path":       "/cardboard",
						"request_id": rid,
						"status":     200,
					}))

					body, err := io.ReadAll(response.Body)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(body)).To(Equal(`{"ima": "pc"}`))
				})
			})
		})

	})

})
