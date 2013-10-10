package commands_test

import (
	"cf"
	. "cf/commands"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
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

	ui := callStacks(stackRepo)

	assert.Equal(t, len(ui.Outputs), 5)
	assert.Contains(t, ui.Outputs[0], "Getting stacks")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[3], "Stack-1")
	assert.Contains(t, ui.Outputs[3], "Stack 1 Description")
	assert.Contains(t, ui.Outputs[4], "Stack-2")
	assert.Contains(t, ui.Outputs[4], "Stack 2 Description")
}

func callStacks(stackRepo *testapi.FakeStackRepository) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}

	ctxt := testcmd.NewContext("stacks", []string{})
	cmd := NewStacks(ui, stackRepo)
	testcmd.RunCommand(cmd, ctxt, nil)

	return
}
