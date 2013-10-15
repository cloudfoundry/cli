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
	config := &configuration.Configuration{
		Organization: cf.Organization{
			Name: "my-org",
			Guid: "my-org-guid",
		},
		Space: cf.Space{
			Name: "my-space",
			Guid: "my-space-guid",
		},
	}

	req := newTargetedSpaceRequirement(ui, config)
	success := req.Execute()
	assert.True(t, success)

	config.Space = cf.Space{}

	req = newTargetedSpaceRequirement(ui, config)
	success = req.Execute()
	assert.False(t, success)
	assert.Contains(t, ui.Outputs[0], "FAILED")
	assert.Contains(t, ui.Outputs[1], "No space targeted")

	ui.ClearOutputs()
	config.Organization = cf.Organization{}

	req = newTargetedSpaceRequirement(ui, config)
	success = req.Execute()
	assert.False(t, success)
	assert.Contains(t, ui.Outputs[0], "FAILED")
	assert.Contains(t, ui.Outputs[1], "No org and space targeted")
}
