package requirements

import (
	"cf"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testterm "testhelpers/terminal"
	"testing"
)

func TestSpaceRequirement(t *testing.T) {
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
	assert.True(t, success)

	config.SpaceFields = cf.SpaceFields{}

	req = newTargetedSpaceRequirement(ui, config)
	success = req.Execute()
	assert.False(t, success)
	assert.Contains(t, ui.Outputs[0], "FAILED")
	assert.Contains(t, ui.Outputs[1], "No space targeted")

	ui.ClearOutputs()
	config.OrganizationFields = cf.OrganizationFields{}

	req = newTargetedSpaceRequirement(ui, config)
	success = req.Execute()
	assert.False(t, success)
	assert.Contains(t, ui.Outputs[0], "FAILED")
	assert.Contains(t, ui.Outputs[1], "No org and space targeted")
}
