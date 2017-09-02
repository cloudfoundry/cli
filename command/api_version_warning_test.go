package command_test

import (
	. "code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/commandfakes"
	"code.cloudfoundry.org/cli/util/ui"
	"code.cloudfoundry.org/cli/version"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("WarnAPIVersionCheck", func() {
	var (
		testUI        *ui.UI
		fakeConfig    *commandfakes.FakeConfig
		apiVersion    string
		minCLIVersion string
		binaryVersion string
	)

	BeforeEach(func() {
		testUI = ui.NewTestUI(nil, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)

		apiVersion = "1.2.3"
		fakeConfig.APIVersionReturns(apiVersion)
		minCLIVersion = "1.0.0"
		fakeConfig.MinCLIVersionReturns(minCLIVersion)
	})

	Context("when checking the cloud controller minimum version warning", func() {
		Context("when the CLI version is less than the recommended minimum", func() {
			BeforeEach(func() {
				binaryVersion = "0.0.0"
				fakeConfig.BinaryVersionReturns(binaryVersion)
			})

			It("displays a recommendation to update the CLI version", func() {
				err := WarnAPIVersionCheck(fakeConfig, testUI)
				Expect(testUI.Err).To(Say("Cloud Foundry API version %s requires CLI version %s. You are currently on version %s. To upgrade your CLI, please visit: https://github.com/cloudfoundry/cli#downloads", apiVersion, minCLIVersion, binaryVersion))
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	Context("when the CLI version is greater or equal to the recommended minimum", func() {
		BeforeEach(func() {
			binaryVersion = "1.0.0"
			fakeConfig.BinaryVersionReturns(binaryVersion)
		})

		It("does not display a recommendation to update the CLI version", func() {
			err := WarnAPIVersionCheck(fakeConfig, testUI)
			Expect(err).ToNot(HaveOccurred())
			Expect(testUI.Err).NotTo(Say("Cloud Foundry API version %s requires CLI version %s. You are currently on version %s. To upgrade your CLI, please visit: https://github.com/cloudfoundry/cli#downloads", apiVersion, minCLIVersion, binaryVersion))
		})
	})

	Context("when an error is encountered while parsing the semver versions", func() {
		BeforeEach(func() {
			fakeConfig.BinaryVersionReturns("&#%")
		})

		It("does not recommend to update the CLI version", func() {
			err := WarnAPIVersionCheck(fakeConfig, testUI)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("No Major.Minor.Patch elements found"))
			Expect(testUI.Err).NotTo(Say("Cloud Foundry API version %s requires CLI version %s.", apiVersion, minCLIVersion))
		})
	})

	Context("version contains string", func() {
		BeforeEach(func() {
			minCLIVersion = "1.0.0"
			fakeConfig.MinCLIVersionReturns(minCLIVersion)
			binaryVersion = "1.0.0-alpha.5"
			fakeConfig.BinaryVersionReturns(binaryVersion)
		})

		It("does not return an error", func() {
			err := WarnAPIVersionCheck(fakeConfig, testUI)
			Expect(err).ToNot(HaveOccurred())
			Expect(testUI.Err).To(Say("Cloud Foundry API version %s requires CLI version %s. You are currently on version %s. To upgrade your CLI, please visit: https://github.com/cloudfoundry/cli#downloads", apiVersion, minCLIVersion, binaryVersion))
		})
	})

	Context("minimum version is empty", func() {
		BeforeEach(func() {
			minCLIVersion = ""
			fakeConfig.MinCLIVersionReturns(minCLIVersion)
			binaryVersion = "1.0.0"
			fakeConfig.BinaryVersionReturns(binaryVersion)
		})

		It("does not return an error", func() {
			err := WarnAPIVersionCheck(fakeConfig, testUI)
			Expect(err).ToNot(HaveOccurred())
			Expect(testUI.Err).NotTo(Say("Cloud Foundry API version %s requires CLI version %s. You are currently on version %s. To upgrade your CLI, please visit: https://github.com/cloudfoundry/cli#downloads", apiVersion, minCLIVersion, binaryVersion))
		})
	})

	Context("when comparing the default versions", func() {
		BeforeEach(func() {
			minCLIVersion = "1.2.3"
			fakeConfig.MinCLIVersionReturns(minCLIVersion)
			binaryVersion = version.DefaultVersion
			fakeConfig.BinaryVersionReturns(binaryVersion)
		})

		It("does not return an error", func() {
			err := WarnAPIVersionCheck(fakeConfig, testUI)
			Expect(err).ToNot(HaveOccurred())
			Expect(testUI.Err).NotTo(Say("Cloud Foundry API version %s requires CLI version %s. You are currently on version %s. To upgrade your CLI, please visit: https://github.com/cloudfoundry/cli#downloads", apiVersion, minCLIVersion, binaryVersion))
		})
	})
})
