package requirements_test

import (
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/models"

	testassert "github.com/cloudfoundry/cli/testhelpers/assert"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/requirements"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TargetedOrganizationRequirement", func() {
	var (
		ui     *testterm.FakeUI
		config configuration.ReadWriter
	)

	BeforeEach(func() {
		ui = new(testterm.FakeUI)
		config = testconfig.NewRepositoryWithDefaults()
	})

	Context("when the user has an org targeted", func() {
		It("succeeds", func() {
			req := NewTargetedOrgRequirement(ui, config)
			success := req.Execute()
			Expect(success).To(BeTrue())
		})
	})

	Context("when the user does not have an org targeted", func() {
		It("fails", func() {
			config.SetOrganizationFields(models.OrganizationFields{})

			testassert.AssertPanic(testterm.QuietPanic, func() {
				NewTargetedOrgRequirement(ui, config).Execute()
			})

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"FAILED"},
				[]string{"No org targeted"},
			))
		})
	})
})
