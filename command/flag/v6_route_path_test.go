package flag_test

import (
	. "code.cloudfoundry.org/cli/command/flag"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("V6RoutePath", func() {
	var routePath V6RoutePath

	Describe("UnmarshalFlag", func() {
		BeforeEach(func() {
			routePath = V6RoutePath{}
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
	})
})
