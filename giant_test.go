package giant_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/clarktrimble/giant"
	"github.com/clarktrimble/giant/mock"
)

func TestGiant(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Giant Suite")
}

var _ = Describe("Giant", func() {

	var (
		ts  *mock.Server
		gnt *Giant
		ctx context.Context
		err error

		method string
		path   string
	)

	BeforeEach(func() {
		ts = mock.NewServer(`{"data": "thing2"}`)
		gnt = &Giant{
			Client:  http.Client{},
			BaseUri: ts.Server.URL,
		}
		ctx = context.Background()
	})

	Describe("sending object", func() {
		var (
			sndObj foo
			rcvObj foo
		)

		JustBeforeEach(func() {
			err = gnt.SendObject(ctx, method, path, sndObj, &rcvObj)
		})

		When("all is well", func() {
			BeforeEach(func() {
				method = "PUT"
				path = "/posts/"
				sndObj = foo{Data: "stuff"}
				rcvObj = foo{}
			})
			It("marshalls send and unmarshalls receive", func() {
				Expect(err).ToNot(HaveOccurred())

				Expect(ts.ContentHeader).To(Equal("application/json"))
				Expect(ts.Method).To(Equal("PUT"))
				Expect(ts.Path).To(Equal("/posts/"))
				Expect(ts.Body).To(Equal(`{"data":"stuff"}`))

				Expect(rcvObj).To(Equal(foo{Data: "thing2"}))
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
			})
			It("makes a matching request", func() {
				Expect(err).ToNot(HaveOccurred())
				Expect(ts.ContentHeader).To(Equal("application/json"))
				Expect(ts.Method).To(Equal("POST"))
				Expect(ts.Path).To(Equal("/posts/"))
				Expect(ts.Body).To(Equal(`{"data": "about a post"}`))
			})
		})

	})

})

// help

type foo struct {
	Data string `json:"data"`
}
