package requirements_test

import (
	"cf/configuration"
	"cf/models"
	. "cf/requirements"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	testassert "testhelpers/assert"
	testterm "testhelpers/terminal"
)

func init() {
	Describe("Testing with ginkgo", func() {
		It("TestSpaceRequirement", func() {
			ui := new(testterm.FakeUI)
			org := models.OrganizationFields{}
			org.Name = "my-org"
			org.Guid = "my-org-guid"
			space := models.SpaceFields{}
			space.Name = "my-space"
			space.Guid = "my-space-guid"
			config := &configuration.Configuration{
				OrganizationFields: org,

				SpaceFields: space,
			}

			req := NewTargetedSpaceRequirement(ui, config)
			success := req.Execute()
			assert.True(mr.T(), success)

			config.SpaceFields = models.SpaceFields{}

			testassert.AssertPanic(mr.T(), testterm.FailedWasCalled, func() {
				NewTargetedSpaceRequirement(ui, config).Execute()
			})

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"FAILED"},
				{"No space targeted"},
			})

			ui.ClearOutputs()
			config.OrganizationFields = models.OrganizationFields{}

			testassert.AssertPanic(mr.T(), testterm.FailedWasCalled, func() {
				NewTargetedSpaceRequirement(ui, config).Execute()
			})

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"FAILED"},
				{"No org and space targeted"},
			})
		})
	})
}
