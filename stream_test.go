package giant

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("StreamLines", func() {

	var (
		server *httptest.Server
		gnt    *Giant
		ctx    context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
	})

	AfterEach(func() {
		if server != nil {
			server.Close()
		}
	})

	When("server streams lines", func() {
		BeforeEach(func() {
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintln(w, `{"event":"start"}`)
				fmt.Fprintln(w, `{"event":"stop"}`)
			}))
			gnt = &Giant{
				Client:  http.Client{},
				BaseUri: server.URL,
			}
		})

		It("iterates over lines", func() {
			lines, err := gnt.StreamLines(ctx, "/events")
			Expect(err).ToNot(HaveOccurred())

			var received []string
			for data := range lines {
				received = append(received, string(data))
			}

			Expect(received).To(Equal([]string{
				`{"event":"start"}`,
				`{"event":"stop"}`,
			}))
		})
	})
})
