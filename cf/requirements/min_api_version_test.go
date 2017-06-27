package requirements_test

import (
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/requirements"
	"github.com/blang/semver"

	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("MinAPIVersionRequirement", func() {
	var (
		config      coreconfig.Repository
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
				config.SetAPIVersion("1.2.4")
			})

			It("succeeds", func() {
				err := requirement.Execute()
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the config's api version is equal to the required version", func() {
			BeforeEach(func() {
				config.SetAPIVersion("1.2.3")
			})

			It("succeeds", func() {
				err := requirement.Execute()
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the config's api version is less than the required version", func() {
			BeforeEach(func() {
				config.SetAPIVersion("1.2.2")
			})

			It("errors", func() {
				err := requirement.Execute()
				Expect(err.Error()).To(ContainSubstring("version-restricted-feature requires CF API version 1.2.3 or higher. Your target is 1.2.2."))
			})
		})

		Context("when the config's api version can not be parsed", func() {
			BeforeEach(func() {
				config.SetAPIVersion("-")
			})

			It("errors", func() {
				err := requirement.Execute()
				Expect(err.Error()).To(ContainSubstring("Unable to parse CC API Version '-'"))
			})
		})

		Context("when the config's api version is empty", func() {
			BeforeEach(func() {
				config.SetAPIVersion("")
			})

			It("errors", func() {
				err := requirement.Execute()
				Expect(err.Error()).To(ContainSubstring("Unable to determine CC API Version. Please log in again."))
			})
		})
	})
})
