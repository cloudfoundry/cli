package flag_test

import (
	. "code.cloudfoundry.org/cli/command/flag"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("V7RoutePath", func() {
	var routePath V7RoutePath

	Describe("UnmarshalFlag", func() {
		BeforeEach(func() {
			routePath = V7RoutePath{}
		})

		When("passed a path beginning with a slash", func() {
			It("sets the path", func() {
				err := routePath.UnmarshalFlag("/banana")
				Expect(err).ToNot(HaveOccurred())
				Expect(routePath.Path).To(Equal("/banana"))
			})
		})

		When("passed a path that doesn't begin with a slash", func() {
			It("prepends the path with a slash and sets it", func() {
				err := routePath.UnmarshalFlag("banana")
				Expect(err).ToNot(HaveOccurred())
				Expect(routePath.Path).To(Equal("/banana"))
			})
		})

		When("passed an empty string", func() {
			It("leaves the string as empty", func() {
				err := routePath.UnmarshalFlag("")
				Expect(err).ToNot(HaveOccurred())
				Expect(routePath.Path).To(Equal(""))
			})
		})
	})
})
