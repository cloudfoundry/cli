package requirements_test

import (
	"cf"
	"cf/configuration"
	. "cf/requirements"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestSpaceRequirement(t *testing.T) {
	ui := new(testhelpers.FakeUI)
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

	req := NewSpaceRequirement(ui, config)
	err := req.Execute()
	assert.NoError(t, err)

	config.Space = cf.Space{}

	req = NewSpaceRequirement(ui, config)
	err = req.Execute()
	assert.Error(t, err)
	assert.Contains(t, ui.Outputs[0], "FAILED")
	assert.Contains(t, ui.Outputs[1], "No space targeted")

	ui.ClearOutputs()
	config.Organization = cf.Organization{}

	req = NewSpaceRequirement(ui, config)
	err = req.Execute()
	assert.Error(t, err)
	assert.Contains(t, ui.Outputs[0], "FAILED")
	assert.Contains(t, ui.Outputs[1], "No org and space targeted")
}
