package shared_test

import (
	"code.cloudfoundry.org/cli/command/v7/shared"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("version checker", func() {
	var (
		warning          string
		executeErr       error
		currentCCVersion string
	)

	Context("CheckCCAPIVersion", func() {
		BeforeEach(func() {
			currentCCVersion = "3.84.0"
		})

		JustBeforeEach(func() {
			warning, executeErr = shared.CheckCCAPIVersion(currentCCVersion)
		})

		It("does not return a warning", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(warning).To(Equal(""))
		})

		When("the current version is less than the minimum version", func() {
			BeforeEach(func() {
				currentCCVersion = "3.83.0"
			})

			It("does return a warning", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warning).To(Equal("Warning: Your targeted API's version (3.83.0) is less than the minimum supported API version (3.84.0). Some commands may not function correctly."))
			})
		})

		When("the API version is empty", func() {
			BeforeEach(func() {
				currentCCVersion = ""
			})
			It("returns an error", func() {
				Expect(executeErr).To(HaveOccurred())
			})
		})

	})
})
