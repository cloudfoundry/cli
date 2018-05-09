package flag_test

import (
	. "code.cloudfoundry.org/cli/command/flag"
	flags "github.com/jessevdk/go-flags"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Buildpack", func() {
	var buildpack Buildpack

	BeforeEach(func() {
		buildpack = Buildpack{}
	})

	Describe("UnmarshalFlag", func() {
		Context("when the value provided to the buildpack flag is valid", func() {
			It("unmarshals into a filtered string and does not return an error", func() {
				err := buildpack.UnmarshalFlag("default")
				Expect(err).ToNot(HaveOccurred())
				Expect(buildpack.IsSet).To(BeTrue())
				Expect(buildpack.Value).To(BeEmpty())
			})
		})

		Context("when the value provided to the buildpack flag starts with a '-'", func() {
			It("returns a ErrExpectedArgument error that an argument for buildpack was expected", func() {
				Expect(buildpack.UnmarshalFlag("-some-val")).To(MatchError(&flags.Error{
					Type:    flags.ErrExpectedArgument,
					Message: "expected argument for flag --buildpack, but got option -some-val",
				}))
			})
		})
	})
})
