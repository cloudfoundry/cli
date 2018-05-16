package versioncheck_test

import (
	. "code.cloudfoundry.org/cli/actor/versioncheck"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("IsMinimumAPIVersionMet", func() {

	var (
		executeErr       error
		minimumVersion   string
		currentVersion   string
		minAPIVersionMet bool
	)

	BeforeEach(func() {
		currentVersion = "1.0.0"
		minimumVersion = "1.0.0"
	})

	JustBeforeEach(func() {
		minAPIVersionMet, executeErr = IsMinimumAPIVersionMet(currentVersion, minimumVersion)
	})

	Context("if the current version is not a valid semver", func() {
		BeforeEach(func() {
			currentVersion = "asfd"
		})

		It("raises an error", func() {
			Expect(executeErr).To(HaveOccurred())
		})
	})

	Context("if the minimum version is not a valid semver", func() {
		BeforeEach(func() {
			minimumVersion = "asfd"
		})

		It("raises an error", func() {
			Expect(executeErr).To(HaveOccurred())
		})
	})

	Context("minimum version is empty", func() {
		BeforeEach(func() {
			minimumVersion = ""
		})

		It("returns true with no errors", func() {
			Expect(minAPIVersionMet).To(Equal(true))
			Expect(executeErr).ToNot(HaveOccurred())
		})
	})

	Context("current version is less than min", func() {
		BeforeEach(func() {
			currentVersion = "3.22.0"
			minimumVersion = "3.25.0"
		})

		It("returns true", func() {
			Expect(minAPIVersionMet).To(Equal(false))
			Expect(executeErr).ToNot(HaveOccurred())
		})
	})

	Context("current version is greater than or equal to min", func() {
		BeforeEach(func() {
			currentVersion = "3.26.0"
			minimumVersion = "3.25.0"
		})

		It("returns true", func() {
			Expect(minAPIVersionMet).To(Equal(true))
			Expect(executeErr).ToNot(HaveOccurred())
		})
	})
})
