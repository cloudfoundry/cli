package flag_test

import (
	flags "github.com/jessevdk/go-flags"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "code.cloudfoundry.org/cli/command/flag"
)

var _ = Describe("Positive Integer", func() {
	var (
		posInt PositiveInteger
	)

	Describe("UnmarshalFlag", func() {
		BeforeEach(func() {
			posInt = PositiveInteger{}
		})

		When("passed a positive integer", func() {
			It("sets the value", func() {
				err := posInt.UnmarshalFlag("42")
				Expect(err).ToNot(HaveOccurred())
				Expect(posInt.Value).To(BeEquivalentTo(42))
			})
		})

		When("passed a non-positive integer", func() {
			It("it returns an error", func() {
				err := posInt.UnmarshalFlag("0")
				Expect(err).To(MatchError(&flags.Error{
					Type:    flags.ErrMarshal,
					Message: `Value must be greater than or equal to 1.`,
				}))
			})
		})
	})

})
