package rando_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/clarktrimble/giant/rando"
)

func TestRando(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Rando Suite")
}

var _ = Describe("Rando", func() {

	var (
		length int = 9
		id     string
	)

	JustBeforeEach(func() {
		id = Rando(length)
	})

	Describe("generating a random string", func() {
		When("all is well", func() {
			It("returns a string of requested length", func() {
				Expect(id).To(HaveLen(length))
			})
		})
	})

	// Todo: check more, perhaps with fuzz?

})
