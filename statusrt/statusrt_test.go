package statusrt

import (
	"io"
	"net/http"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestStatusRt(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "StatusRt Suite")
}

var _ = Describe("StatusRt", func() {

	Describe("tripperware", func() {

		var (
			rt       *StatusRt
			request  *http.Request
			response *http.Response

			err error
		)

		JustBeforeEach(func() {
			response, err = rt.RoundTrip(request)
		})

		Describe("detecting non-200 statuses", func() {

			When("status is in the 200's", func() {
				BeforeEach(func() {
					rt = &StatusRt{}
					rt.Wrap(&testRt{
						Status: 201,
					})

					request, err = http.NewRequest("PUT", "https://boxworld.org/cardboard", nil)
					Expect(err).ToNot(HaveOccurred())
				})
				It("passes thru the response", func() {

					Expect(err).ToNot(HaveOccurred())

					body, err := io.ReadAll(response.Body)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(body)).To(Equal(`{"ima": "pc"}`))
				})
			})

			When("status is _not_ in the 200's", func() {
				BeforeEach(func() {
					rt = &StatusRt{}
					rt.Wrap(&testRt{
						Status: 404,
					})

					request, err = http.NewRequest("PUT", "https://boxworld.org/cardboard", nil)
					Expect(err).ToNot(HaveOccurred())
				})
				It("returns an error", func() {

					Expect(err).To(HaveOccurred())
					Expect(response.StatusCode).To(Equal(404))
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
