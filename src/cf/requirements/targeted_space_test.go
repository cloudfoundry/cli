package requirements

import (
	"cf"
	"cf/configuration"
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
			org := cf.OrganizationFields{}
			org.Name = "my-org"
			org.Guid = "my-org-guid"
			space := cf.SpaceFields{}
			space.Name = "my-space"
			space.Guid = "my-space-guid"
			config := &configuration.Configuration{
				OrganizationFields: org,

				SpaceFields: space,
			}

			req := newTargetedSpaceRequirement(ui, config)
			success := req.Execute()
			assert.True(mr.T(), success)

			config.SpaceFields = cf.SpaceFields{}

			req = newTargetedSpaceRequirement(ui, config)
			success = req.Execute()
			assert.False(mr.T(), success)
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"FAILED"},
				{"No space targeted"},
			})

			ui.ClearOutputs()
			config.OrganizationFields = cf.OrganizationFields{}

			req = newTargetedSpaceRequirement(ui, config)
			success = req.Execute()
			assert.False(mr.T(), success)
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"FAILED"},
				{"No org and space targeted"},
			})
		})
	})
}
