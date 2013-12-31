package commands_test

import (
	"cf"
	. "cf/commands"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testterm "testhelpers/terminal"
	"testing"
)

func TestStacks(t *testing.T) {
	stack1 := cf.Stack{}
	stack1.Name = "Stack-1"
	stack1.Description = "Stack 1 Description"

	stack2 := cf.Stack{}
	stack2.Name = "Stack-2"
	stack2.Description = "Stack 2 Description"

	stackRepo := &testapi.FakeStackRepository{
		FindAllStacks: []cf.Stack{stack1, stack2},
	}

	ui := callStacks(t, stackRepo)

	assert.Equal(t, len(ui.Outputs), 6)
	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Getting stacks in org", "my-org", "my-space", "my-user"},
		{"OK"},
		{"Stack-1", "Stack 1 Description"},
		{"Stack-2", "Stack 2 Description"},
	})
}

func callStacks(t *testing.T, stackRepo *testapi.FakeStackRepository) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}

	ctxt := testcmd.NewContext("stacks", []string{})

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)

	space := cf.SpaceFields{}
	space.Name = "my-space"

	org := cf.OrganizationFields{}
	org.Name = "my-org"

	config := &configuration.Configuration{
		SpaceFields:        space,
		OrganizationFields: org,
		AccessToken:        token,
	}

	cmd := NewStacks(ui, config, stackRepo)
	testcmd.RunCommand(cmd, ctxt, nil)

	return
}
