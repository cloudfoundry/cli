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
	It("TestSpaceRequirement", func() {
		ui := new(testterm.FakeUI)
		org := models.OrganizationFields{}
		org.Name = "my-org"
		org.Guid = "my-org-guid"
		space := models.SpaceFields{}
		space.Name = "my-space"
		space.Guid = "my-space-guid"
		config := testconfig.NewRepositoryWithDefaults()

		req := NewTargetedSpaceRequirement(ui, config)
		success := req.Execute()
		Expect(success).To(BeTrue())

		config.SetSpaceFields(models.SpaceFields{})

		testassert.AssertPanic(testterm.FailedWasCalled, func() {
			NewTargetedSpaceRequirement(ui, config).Execute()
		})

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"FAILED"},
			{"No space targeted"},
		})

		ui.ClearOutputs()
		config.SetOrganizationFields(models.OrganizationFields{})

		testassert.AssertPanic(testterm.FailedWasCalled, func() {
			NewTargetedSpaceRequirement(ui, config).Execute()
		})

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"FAILED"},
			{"No org and space targeted"},
		})
	})
})
