package requirements_test

import (
	"cf/models"
	. "cf/requirements"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	testassert "testhelpers/assert"
	testconfig "testhelpers/configuration"
	testterm "testhelpers/terminal"
)

var _ = Describe("Testing with ginkgo", func() {
	It("TestTargetedOrgRequirement", func() {
		ui := new(testterm.FakeUI)
		org := models.OrganizationFields{}
		org.Name = "my-org"
		org.Guid = "my-org-guid"
		config := testconfig.NewRepositoryWithDefaults()

		req := NewTargetedOrgRequirement(ui, config)
		success := req.Execute()
		Expect(success).To(BeTrue())

		config.SetOrganizationFields(models.OrganizationFields{})

		testassert.AssertPanic(testterm.FailedWasCalled, func() {
			NewTargetedOrgRequirement(ui, config).Execute()
		})

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"FAILED"},
			{"No org targeted"},
		})
	})
})
