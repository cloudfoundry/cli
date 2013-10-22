package commands_test

import (
	"cf"
	. "cf/commands"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testterm "testhelpers/terminal"
	"testing"
)

func TestStacks(t *testing.T) {
	stacks := []cf.Stack{
		cf.Stack{Name: "Stack-1", Description: "Stack 1 Description"},
		cf.Stack{Name: "Stack-2", Description: "Stack 2 Description"},
	}
	stackRepo := &testapi.FakeStackRepository{
		FindAllStacks: stacks,
	}

	ui := callStacks(t, stackRepo)

	assert.Equal(t, len(ui.Outputs), 6)
	assert.Contains(t, ui.Outputs[0], "Getting stacks in org")
	assert.Contains(t, ui.Outputs[0], "my-org")
	assert.Contains(t, ui.Outputs[0], "my-space")
	assert.Contains(t, ui.Outputs[0], "my-user")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[4], "Stack-1")
	assert.Contains(t, ui.Outputs[4], "Stack 1 Description")
	assert.Contains(t, ui.Outputs[5], "Stack-2")
	assert.Contains(t, ui.Outputs[5], "Stack 2 Description")
}

func callStacks(t *testing.T, stackRepo *testapi.FakeStackRepository) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}

	ctxt := testcmd.NewContext("stacks", []string{})

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)

	config := &configuration.Configuration{
		Space:        cf.Space{Name: "my-space"},
		Organization: cf.Organization{Name: "my-org"},
		AccessToken:  token,
	}

	cmd := NewStacks(ui, config, stackRepo)
	testcmd.RunCommand(cmd, ctxt, nil)

	return
}
