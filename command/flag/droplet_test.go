package flag_test

import (
	. "code.cloudfoundry.org/cli/command/flag"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Droplet", func() {
	var droplet Droplet

	Describe("UnmarshalFlag", func() {
		BeforeEach(func() {
			droplet = Droplet{}
		})

		When("passed a path beginning with a slash", func() {
			It("sets the path", func() {
				err := droplet.UnmarshalFlag("/banana")
				Expect(err).ToNot(HaveOccurred())
				Expect(droplet.Path).To(Equal("/banana"))
			})
		})

		When("passed a path that doesn't begin with a slash", func() {
			It("prepends the path with a slash and sets it", func() {
				err := droplet.UnmarshalFlag("banana")
				Expect(err).ToNot(HaveOccurred())
				Expect(droplet.Path).To(Equal("/banana"))
			})
		})
	})
})
