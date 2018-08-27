package command_test

import (
	. "code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/version"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Minimum Version Check", func() {
	Describe("MinimumCFAPIVersionCheck", func() {
		minimumVersion := "1.0.0"
		Context("current version is greater than min", func() {
			It("does not return an error", func() {
				currentVersion := "1.0.1"
				err := MinimumCCAPIVersionCheck(currentVersion, minimumVersion)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("current version is less than min", func() {
			It("does return an error", func() {
				currentVersion := "1.0.0-alpha.5"
				err := MinimumCCAPIVersionCheck(currentVersion, minimumVersion)
				Expect(err).To(MatchError(translatableerror.MinimumCFAPIVersionNotMetError{
					CurrentVersion: currentVersion,
					MinimumVersion: minimumVersion,
				}))
			})

			When("a custom command is provided", func() {
				currentVersion := "1.0.0-alpha.5"
				It("sets the command on the MinimumAPIVersionNotMetError", func() {
					err := MinimumCCAPIVersionCheck(currentVersion, minimumVersion, "some-command")
					Expect(err).To(MatchError(translatableerror.MinimumCFAPIVersionNotMetError{
						Command:        "some-command",
						CurrentVersion: currentVersion,
						MinimumVersion: minimumVersion,
					}))
				})
			})
		})

		Context("current version is the default version", func() {
			It("does not return an error", func() {
				err := MinimumCCAPIVersionCheck(version.DefaultVersion, minimumVersion)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("minimum version is empty", func() {
			It("does not return an error", func() {
				err := MinimumCCAPIVersionCheck("2.0.0", "")
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	Describe("MinimumUAAAPIVersionCheck", func() {
		minimumVersion := "1.0.0"
		Context("current version is greater than min", func() {
			It("does not return an error", func() {
				currentVersion := "1.0.1"
				err := MinimumUAAAPIVersionCheck(currentVersion, minimumVersion)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("current version is less than min", func() {
			It("does return an error", func() {
				currentVersion := "1.0.0-alpha.5"
				err := MinimumUAAAPIVersionCheck(currentVersion, minimumVersion)
				Expect(err).To(MatchError(translatableerror.MinimumUAAAPIVersionNotMetError{
					MinimumVersion: minimumVersion,
				}))
			})

			When("a custom command is provided", func() {
				currentVersion := "1.0.0-alpha.5"
				It("sets the command on the MinimumAPIVersionNotMetError", func() {
					err := MinimumUAAAPIVersionCheck(currentVersion, minimumVersion, "some-command")
					Expect(err).To(MatchError(translatableerror.MinimumUAAAPIVersionNotMetError{
						Command:        "some-command",
						MinimumVersion: minimumVersion,
					}))
				})
			})
		})

		Context("current version is the default version", func() {
			It("does not return an error", func() {
				err := MinimumUAAAPIVersionCheck(version.DefaultVersion, minimumVersion)
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("minimum version is empty", func() {
			It("does not return an error", func() {
				err := MinimumUAAAPIVersionCheck("2.0.0", "")
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})
