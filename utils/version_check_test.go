package utils_test

import (
	. "github.com/cloudfoundry/cli/utils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Version Check", func() {
	Context("GreaterThanOrEqual", func() {
		var (
			curVersion Version
		)

		It("returns true if current version is greater or equal to targeted version", func() {
			curVersion = NewVersion("6.12.0")
			Ω(curVersion.GreaterThanOrEqual(NewVersion("6.7.0"))).To(BeTrue())
			Ω(curVersion.GreaterThanOrEqual(NewVersion("6.12.0"))).To(BeTrue())
			Ω(curVersion.GreaterThanOrEqual(NewVersion("5.9.0"))).To(BeTrue())
		})

		It("returns false if current version is less than targeted version", func() {
			curVersion = NewVersion("6.12.0")
			Ω(curVersion.GreaterThanOrEqual(NewVersion("7.5.0"))).To(BeFalse())
			Ω(curVersion.GreaterThanOrEqual(NewVersion("6.15.0"))).To(BeFalse())
			Ω(curVersion.GreaterThanOrEqual(NewVersion("6.12.0.1"))).To(BeFalse())
			Ω(curVersion.GreaterThanOrEqual(NewVersion("6.12.0.0.1"))).To(BeFalse())
		})

		It("returns false if current version has less digits than targeted version", func() {
			curVersion = NewVersion("6.12.0")
			Ω(curVersion.GreaterThanOrEqual(NewVersion("6.12.0.0"))).To(BeTrue())
		})

		It("returns true if current version has more digits than targeted version", func() {
			curVersion = NewVersion("6.12.0.0")
			Ω(curVersion.GreaterThanOrEqual(NewVersion("6.12.0"))).To(BeTrue())
		})
	})
})
