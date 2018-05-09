package flag_test

import (
	. "code.cloudfoundry.org/cli/command/flag"
	flags "github.com/jessevdk/go-flags"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RoutePath", func() {
	var routePath RoutePath

	Describe("UnmarshalFlag", func() {
		BeforeEach(func() {
			routePath = RoutePath{}
		})

		Context("when passed a path beginning with a slash", func() {
			It("sets the path", func() {
				err := routePath.UnmarshalFlag("/banana")
				Expect(err).ToNot(HaveOccurred())
				Expect(routePath.Path).To(Equal("/banana"))
			})
		})

		Context("when passed a path that begins with a hyphen", func() {
			It("returns a error that an argument was expected", func() {
				Expect(routePath.UnmarshalFlag("-some-val")).To(MatchError(&flags.Error{
					Type:    flags.ErrExpectedArgument,
					Message: "expected argument for flag --route-path, but got option -some-val",
				}))
			})
		})

		Context("when passed a path that doesn't begin with a slash or hyphen", func() {
			It("prepends the path with a slash and sets it", func() {
				err := routePath.UnmarshalFlag("banana")
				Expect(err).ToNot(HaveOccurred())
				Expect(routePath.Path).To(Equal("/banana"))
			})
		})
	})
})
