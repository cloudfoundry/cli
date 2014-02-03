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
		It("TestTargetedOrgRequirement", func() {

			ui := new(testterm.FakeUI)
			org := cf.OrganizationFields{}
			org.Name = "my-org"
			org.Guid = "my-org-guid"
			config := &configuration.Configuration{
				OrganizationFields: org,
			}

			req := newTargetedOrgRequirement(ui, config)
			success := req.Execute()
			assert.True(mr.T(), success)

			config.OrganizationFields = cf.OrganizationFields{}

			testassert.AssertPanic(mr.T(), testterm.FailedWasCalled, func() {
				newTargetedOrgRequirement(ui, config).Execute()
			})

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"FAILED"},
				{"No org targeted"},
			})
		})
	})
}
