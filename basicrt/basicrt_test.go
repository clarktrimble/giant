package basicrt

import (
	"io"
	"net/http"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestBasicRt(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "BasicRt Suite")
}

var _ = Describe("BasicRt", func() {

	Describe("tripperware", func() {

		var (
			rt      *BasicRt
			request *http.Request
			err     error
		)

		JustBeforeEach(func() {
			_, err = rt.RoundTrip(request)
		})

		Describe("adding basic auth header", func() {

			When("all is well", func() {
				BeforeEach(func() {
					rt = New("top", "secret")
					rt.Wrap(&testRt{
						Status: 201,
					})

					request, err = http.NewRequest("PUT", "https://boxworld.org/cardboard", nil)
					Expect(err).ToNot(HaveOccurred())
				})

				It("sets the header", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(request.Header["Authorization"]).To(Equal([]string{"Basic dG9wOnNlY3JldA=="}))
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
