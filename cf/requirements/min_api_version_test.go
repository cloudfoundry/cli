package requirements_test

import (
	"github.com/blang/semver"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/requirements"

	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MinAPIVersionRequirement", func() {
	var (
		config      core_config.Repository
		requirement requirements.MinAPIVersionRequirement
	)

	BeforeEach(func() {
		config = testconfig.NewRepository()
		requiredVersion, err := semver.Make("1.2.3")
		Expect(err).NotTo(HaveOccurred())

		requirement = requirements.NewMinAPIVersionRequirement(config, "version-restricted-feature", requiredVersion)
	})

	Context("Execute", func() {
		Context("when the config's api version is greater than the required version", func() {
			BeforeEach(func() {
				config.SetApiVersion("1.2.4")
			})

			It("succeeds", func() {
				err := requirement.Execute()
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the config's api version is equal to the required version", func() {
			BeforeEach(func() {
				config.SetApiVersion("1.2.3")
			})

			It("succeeds", func() {
				err := requirement.Execute()
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the config's api version is less than the required version", func() {
			BeforeEach(func() {
				config.SetApiVersion("1.2.2")
			})

			It("errors", func() {
				err := requirement.Execute()
				Expect(err.Error()).To(ContainSubstring("version-restricted-feature requires CF API version 1.2.3+. Your target is 1.2.2."))
			})
		})

		Context("when the config's api version can not be parsed", func() {
			BeforeEach(func() {
				config.SetApiVersion("-")
			})

			It("errors", func() {
				err := requirement.Execute()
				Expect(err.Error()).To(ContainSubstring("Unable to parse CC API Version '-'"))
			})
		})

		Context("when the config's api version is empty", func() {
			BeforeEach(func() {
				config.SetApiVersion("")
			})

			It("errors", func() {
				err := requirement.Execute()
				Expect(err.Error()).To(ContainSubstring("Unable to determine CC API Version. Please log in again."))
			})
		})
	})
})
