package command_test

import (
	"fmt"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	. "code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/util/ui"
	"code.cloudfoundry.org/cli/version"
	"github.com/blang/semver"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("version checks", func() {
	Describe("WarnIfCLIVersionBelowAPIDefinedMinimum", func() {
		var (
			testUI     *ui.UI
			fakeConfig *commandfakes.FakeConfig

			executeErr error

			apiVersion    string
			minCLIVersion string
			binaryVersion string
		)

		BeforeEach(func() {
			testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
			fakeConfig = new(commandfakes.FakeConfig)
		})

		JustBeforeEach(func() {
			executeErr = WarnIfCLIVersionBelowAPIDefinedMinimum(fakeConfig, apiVersion, testUI)
		})

		When("checking the cloud controller minimum version warning", func() {
			When("the CLI version is less than the recommended minimum", func() {
				BeforeEach(func() {
					apiVersion = ccversion.MinV2ClientVersion
					minCLIVersion = "1.0.0"
					fakeConfig.MinCLIVersionReturns(minCLIVersion)
					binaryVersion = "0.0.0"
					fakeConfig.BinaryVersionReturns(binaryVersion)
				})

				It("displays a recommendation to update the CLI version", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(testUI.Err).To(Say("Cloud Foundry API version %s requires CLI version %s. You are currently on version %s. To upgrade your CLI, please visit: https://github.com/cloudfoundry/cli#downloads", apiVersion, minCLIVersion, binaryVersion))
				})
			})
		})

		When("the CLI version is greater or equal to the recommended minimum", func() {
			BeforeEach(func() {
				apiVersion = "100.200.3"
				minCLIVersion = "1.0.0"
				fakeConfig.MinCLIVersionReturns(minCLIVersion)
				binaryVersion = "1.0.0"
				fakeConfig.BinaryVersionReturns(binaryVersion)
			})

			It("does not display a recommendation to update the CLI version", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(testUI.Err).NotTo(Say("Cloud Foundry API version %s requires CLI version %s. You are currently on version %s. To upgrade your CLI, please visit: https://github.com/cloudfoundry/cli#downloads", apiVersion, minCLIVersion, binaryVersion))
			})
		})

		When("an error is encountered while parsing the semver versions", func() {
			BeforeEach(func() {
				apiVersion = "100.200.3"
				minCLIVersion = "1.0.0"
				fakeConfig.MinCLIVersionReturns(minCLIVersion)
				fakeConfig.BinaryVersionReturns("&#%")
			})

			It("does not recommend to update the CLI version", func() {
				Expect(executeErr).To(MatchError("No Major.Minor.Patch elements found"))
				Expect(testUI.Err).NotTo(Say("Cloud Foundry API version %s requires CLI version %s. You are currently on version %s. To upgrade your CLI, please visit: https://github.com/cloudfoundry/cli#downloads", apiVersion, minCLIVersion, binaryVersion))
			})
		})

		Context("version contains string", func() {
			BeforeEach(func() {
				apiVersion = "100.200.3"
				minCLIVersion = "1.0.0"
				fakeConfig.MinCLIVersionReturns(minCLIVersion)
				binaryVersion = "1.0.0-alpha.5"
				fakeConfig.BinaryVersionReturns(binaryVersion)
			})

			It("parses the versions successfully and recommends to update the CLI version", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(testUI.Err).To(Say("Cloud Foundry API version %s requires CLI version %s. You are currently on version %s. To upgrade your CLI, please visit: https://github.com/cloudfoundry/cli#downloads", apiVersion, minCLIVersion, binaryVersion))
			})
		})

		Context("minimum version is empty", func() {
			BeforeEach(func() {
				apiVersion = "100.200.3"
				minCLIVersion = ""
				fakeConfig.MinCLIVersionReturns(minCLIVersion)
				binaryVersion = "1.0.0"
				fakeConfig.BinaryVersionReturns(binaryVersion)
			})

			It("does not return an error", func() {
				Expect(executeErr).ToNot(HaveOccurred())
			})
		})

		When("comparing the default versions", func() {
			BeforeEach(func() {
				apiVersion = "100.200.3"
				minCLIVersion = "1.2.3"
				fakeConfig.MinCLIVersionReturns(minCLIVersion)
				binaryVersion = version.DefaultVersion
				fakeConfig.BinaryVersionReturns(binaryVersion)
			})

			It("does not return an error or print a warning", func() {
				Expect(executeErr).ToNot(HaveOccurred())
			})
		})
	})

	Describe("WarnIfAPIVersionBelowSupportedMinimum", func() {
		var (
			testUI *ui.UI

			apiVersion string
		)

		BeforeEach(func() {
			testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		})

		When("checking the cloud controller minimum version warning", func() {
			When("checking for outdated API version", func() {
				When("the API version is older than the minimum supported API version", func() {
					BeforeEach(func() {
						min, err := semver.Make(ccversion.MinV2ClientVersion)
						Expect(err).ToNot(HaveOccurred())
						apiVersion = fmt.Sprintf("%d.%d.%d", min.Major, min.Minor-1, min.Patch)
					})

					It("outputs a warning telling the user to upgrade their CF version", func() {
						err := WarnIfAPIVersionBelowSupportedMinimum(apiVersion, testUI)
						Expect(err).ToNot(HaveOccurred())
						Expect(testUI.Err).To(Say("Your API version is no longer supported. Upgrade to a newer version of the API"))
					})
				})

				When("the API version is newer than the minimum supported API version", func() {
					BeforeEach(func() {
						min, err := semver.Make(ccversion.MinV2ClientVersion)
						Expect(err).ToNot(HaveOccurred())
						apiVersion = fmt.Sprintf("%d.%d.%d", min.Major, min.Minor+1, min.Patch)
					})

					It("continues silently", func() {
						err := WarnIfAPIVersionBelowSupportedMinimum(apiVersion, testUI)
						Expect(err).ToNot(HaveOccurred())
						Expect(testUI.Err).NotTo(Say("Your API version is no longer supported. Upgrade to a newer version of the API"))
					})
				})
			})
		})
	})
})
