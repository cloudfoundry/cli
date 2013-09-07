package commands_test

import (
	"cf"
	. "cf/commands"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestStacks(t *testing.T) {
	stacks := []cf.Stack{
		cf.Stack{Name: "Stack-1", Description: "Stack 1 Description"},
		cf.Stack{Name: "Stack-2", Description: "Stack 2 Description"},
	}
	stackRepo := &testhelpers.FakeStackRepository{
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

func callStacks(stackRepo *testhelpers.FakeStackRepository) (ui *testhelpers.FakeUI) {
	ui = &testhelpers.FakeUI{}

	ctxt := testhelpers.NewContext("stacks", []string{})
	cmd := NewStacks(ui, stackRepo)
	testhelpers.RunCommand(cmd, ctxt, nil)

	return
}
