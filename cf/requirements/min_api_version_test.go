package requirements_test

import (
	"github.com/blang/semver"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/requirements"

	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MinAPIVersionRequirement", func() {
	var (
		ui          *testterm.FakeUI
		config      core_config.Repository
		requirement requirements.MinAPIVersionRequirement
	)

	BeforeEach(func() {
		ui = new(testterm.FakeUI)
		config = testconfig.NewRepository()
		requiredVersion, err := semver.Make("1.2.3")
		Expect(err).NotTo(HaveOccurred())

		requirement = requirements.NewMinAPIVersionRequirement(ui, config, "version-restricted-feature", requiredVersion)
	})

	Context("Execute", func() {
		Context("when the config's api version is greater than the required version", func() {
			BeforeEach(func() {
				config.SetApiVersion("1.2.4")
			})

			It("returns true", func() {
				Expect(requirement.Execute()).To(BeTrue())
			})

			It("does not print anything", func() {
				requirement.Execute()
				Expect(ui.Outputs).To(BeEmpty())
			})
		})

		Context("when the config's api version is equal to the required version", func() {
			BeforeEach(func() {
				config.SetApiVersion("1.2.3")
			})

			It("returns true", func() {
				Expect(requirement.Execute()).To(BeTrue())
			})

			It("does not print anything", func() {
				requirement.Execute()
				Expect(ui.Outputs).To(BeEmpty())
			})
		})

		Context("when the config's api version is less than the required version", func() {
			BeforeEach(func() {
				config.SetApiVersion("1.2.2")
			})

			It("panics and prints a message", func() {
				Expect(func() { requirement.Execute() }).To(Panic())
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"version-restricted-feature requires CF API version 1.2.3+. Your target is 1.2.2."},
				))
			})
		})

		Context("when the config's api version can not be parsed", func() {
			BeforeEach(func() {
				config.SetApiVersion("-")
			})

			It("panics and prints a message", func() {
				Expect(func() { requirement.Execute() }).To(Panic())
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"Unable to parse CC API Version '-'"},
				))
			})
		})

		Context("when the config's api version is empty", func() {
			BeforeEach(func() {
				config.SetApiVersion("")
			})

			It("panics and prints a message", func() {
				Expect(func() { requirement.Execute() }).To(Panic())
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"Unable to determine CC API Version. Please log in again."},
				))
			})
		})
	})
})
