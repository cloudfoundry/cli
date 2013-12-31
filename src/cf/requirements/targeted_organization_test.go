package requirements

import (
	"cf"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testassert "testhelpers/assert"
	testterm "testhelpers/terminal"
	"testing"
)

func TestTargetedOrgRequirement(t *testing.T) {
	ui := new(testterm.FakeUI)
	org := cf.OrganizationFields{}
	org.Name = "my-org"
	org.Guid = "my-org-guid"
	config := &configuration.Configuration{
		OrganizationFields: org,
	}

	req := newTargetedOrgRequirement(ui, config)
	success := req.Execute()
	assert.True(t, success)

	config.OrganizationFields = cf.OrganizationFields{}

	req = newTargetedOrgRequirement(ui, config)
	success = req.Execute()
	assert.False(t, success)
	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"FAILED"},
		{"No org targeted"},
	})
}
