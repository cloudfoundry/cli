package requirements_test

import (
	"cf/models"
	. "cf/requirements"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
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
		assert.True(mr.T(), success)

		config.SetOrganizationFields(models.OrganizationFields{})

		testassert.AssertPanic(mr.T(), testterm.FailedWasCalled, func() {
			NewTargetedOrgRequirement(ui, config).Execute()
		})

		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"FAILED"},
			{"No org targeted"},
		})
	})
})
