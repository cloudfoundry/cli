package requirements

import (
	"cf"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testterm "testhelpers/terminal"
	"testing"
)

func TestTargetedOrgRequirement(t *testing.T) {
	ui := new(testterm.FakeUI)
	config := &configuration.Configuration{
		Organization: cf.Organization{
			Name: "my-org",
			Guid: "my-org-guid",
		},
	}

	req := newTargetedOrgRequirement(ui, config)
	success := req.Execute()
	assert.True(t, success)

	config.Organization = cf.Organization{}

	req = newTargetedOrgRequirement(ui, config)
	success = req.Execute()
	assert.False(t, success)
	assert.Contains(t, ui.Outputs[0], "FAILED")
	assert.Contains(t, ui.Outputs[1], "No org targeted")
}
