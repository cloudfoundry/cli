package requirements_test

import (
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"

	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"

	. "github.com/cloudfoundry/cli/cf/requirements"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TargetedOrganizationRequirement", func() {
	var (
		config core_config.ReadWriter
	)

	BeforeEach(func() {
		config = testconfig.NewRepositoryWithDefaults()
	})

	Context("when the user has an org targeted", func() {
		It("succeeds", func() {
			req := NewTargetedOrgRequirement(config)
			err := req.Execute()
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("when the user does not have an org targeted", func() {
		It("errors", func() {
			config.SetOrganizationFields(models.OrganizationFields{})

			err := NewTargetedOrgRequirement(config).Execute()

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("No org targeted"))
		})
	})
})
