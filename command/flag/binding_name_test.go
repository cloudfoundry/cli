package flag_test

import (
	. "code.cloudfoundry.org/cli/command/flag"
	flags "github.com/jessevdk/go-flags"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BindingName", func() {
	var bindingName BindingName

	BeforeEach(func() {
		bindingName = BindingName{}
	})

	When("the value provided to the --binding-name flag is the empty string", func() {
		It("returns a ErrMarshal error that the binding name must be greater than 1 character long", func() {
			Expect(bindingName.UnmarshalFlag("")).To(MatchError(&flags.Error{
				Type:    flags.ErrMarshal,
				Message: "--binding-name must be at least 1 character in length",
			}))
		})
	})

	When("the value provided to the --binding-name flag is greater than 0 characters long", func() {
		It("stores the binding name and does not return an error", func() {
			err := bindingName.UnmarshalFlag("some-name")
			Expect(err).NotTo(HaveOccurred())
			Expect(bindingName.Value).To(Equal("some-name"))
		})
	})
})
