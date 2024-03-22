package flag_test

import (
	flags "github.com/jessevdk/go-flags"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "code.cloudfoundry.org/cli/command/flag"
)

var _ = Describe("Timeout", func() {
	var (
		timeout Timeout
	)

	Describe("UnmarshalFlag", func() {
		BeforeEach(func() {
			timeout = Timeout{}
		})

		When("passed a positive integer", func() {
			It("sets the value", func() {
				err := timeout.UnmarshalFlag("42")
				Expect(err).ToNot(HaveOccurred())
				Expect(timeout.Value).To(BeEquivalentTo(42))
				Expect(timeout.IsSet).To(BeTrue())
			})
		})

		When("passed a non-positive integer", func() {
			It("it returns an error", func() {
				err := timeout.UnmarshalFlag("0")
				Expect(err).To(MatchError(&flags.Error{
					Type:    flags.ErrRequired,
					Message: `Timeout must be an integer greater than or equal to 1`,
				}))
			})
		})
	})
})
