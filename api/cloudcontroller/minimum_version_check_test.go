package cloudcontroller_test

import (
	. "code.cloudfoundry.org/cli/api/cloudcontroller"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Minimum Version Check", func() {
	Describe("MinimumAPIVersionCheck", func() {
		minimumVersion := "1.0.0"
		Context("current version is greater than min", func() {
			It("does not return an error", func() {
				currentVersion := "1.0.1"
				err := MinimumAPIVersionCheck(currentVersion, minimumVersion)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("current version is less than min", func() {
			It("does return an error", func() {
				currentVersion := "1.0.0-alpha.5"
				err := MinimumAPIVersionCheck(currentVersion, minimumVersion)
				Expect(err).To(MatchError(ccerror.MinimumAPIVersionNotMetError{
					CurrentVersion: currentVersion,
					MinimumVersion: minimumVersion,
				}))
			})
		})

		Context("minimum version is empty", func() {
			It("does not return an error", func() {
				err := MinimumAPIVersionCheck("2.0.0", "")
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})
