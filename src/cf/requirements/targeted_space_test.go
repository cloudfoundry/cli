package requirements

import (
	"cf"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testassert "testhelpers/assert"
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

	testassert.AssertPanic(t, testterm.FailedWasCalled, func() {
		newTargetedSpaceRequirement(ui, config).Execute()
	})

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"FAILED"},
		{"No space targeted"},
	})

	ui.ClearOutputs()
	config.OrganizationFields = cf.OrganizationFields{}

	testassert.AssertPanic(t, testterm.FailedWasCalled, func() {
		newTargetedSpaceRequirement(ui, config).Execute()
	})

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"FAILED"},
		{"No org and space targeted"},
	})
}
