package flag_test

import (
	. "code.cloudfoundry.org/cli/command/flag"
	flags "github.com/jessevdk/go-flags"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Command", func() {
	var command Command

	BeforeEach(func() {
		command = Command{}
	})

	Describe("UnmarshalFlag", func() {
		Context("when the value provided to the command flag is valid", func() {
			It("unmarshals into a filtered string", func() {
				err := command.UnmarshalFlag("default")
				Expect(err).ToNot(HaveOccurred())
				Expect(command.IsSet).To(BeTrue())
				Expect(command.Value).To(BeEmpty())
			})
		})

		Context("when the value provided to the command flag starts with a '-'", func() {
			It("returns a ErrExpectedArgument error that an argument for command was expected", func() {
				Expect(command.UnmarshalFlag("-some-val")).To(MatchError(&flags.Error{
					Type:    flags.ErrExpectedArgument,
					Message: "expected argument for flag -c, but got option -some-val",
				}))
			})
		})
	})
})
