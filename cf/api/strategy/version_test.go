package strategy_test

import (
	. "github.com/cloudfoundry/cli/cf/api/strategy"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("api version", func() {
	Describe("parsing", func() {
		It("parses the major, minor and patch numbers", func() {
			version, err := ParseVersion("1.2.3")
			Expect(err).NotTo(HaveOccurred())
			Expect(version).To(Equal(Version{Major: 1, Minor: 2, Patch: 3}))
		})

		It("returns an error when there aren't three numbers", func() {
			_, err := ParseVersion("1.2")
			Expect(err).To(HaveOccurred())

			_, err = ParseVersion("1.2.3.4")
			Expect(err).To(HaveOccurred())
		})

		It("returns an error when there are non-digits in the version numbers", func() {
			_, err := ParseVersion("1.2.x")
			Expect(err).To(HaveOccurred())

			_, err = ParseVersion("1.x.2")
			Expect(err).To(HaveOccurred())

			_, err = ParseVersion("x.2.3")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("comparisons", func() {
		It("compares the major version", func() {
			Expect(Version{Major: 1, Minor: 2, Patch: 3}.LessThan(Version{Major: 2, Minor: 1, Patch: 1})).To(BeTrue())
			Expect(Version{Major: 2, Minor: 1, Patch: 1}.LessThan(Version{Major: 1, Minor: 3, Patch: 3})).To(BeFalse())
		})

		It("compares the minor version", func() {
			Expect(Version{Major: 1, Minor: 2, Patch: 3}.LessThan(Version{Major: 1, Minor: 3, Patch: 1})).To(BeTrue())
			Expect(Version{Major: 1, Minor: 3, Patch: 1}.LessThan(Version{Major: 1, Minor: 1, Patch: 100})).To(BeFalse())
		})

		It("compares the patch version", func() {
			Expect(Version{Major: 1, Minor: 2, Patch: 3}.LessThan(Version{Major: 1, Minor: 2, Patch: 42})).To(BeTrue())
			Expect(Version{Major: 1, Minor: 2, Patch: 42}.LessThan(Version{Major: 1, Minor: 2, Patch: 3})).To(BeFalse())
		})
	})
})
