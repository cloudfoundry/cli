package commands_test

import (
	"cf"
	. "cf/commands"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestDeleteSpaceConfirmingWithY(t *testing.T) {
	ui, reqFactory, spaceRepo := deleteSpace("y", []string{"space-to-delete"})

	assert.Equal(t, reqFactory.SpaceName, "space-to-delete")
	assert.Contains(t, ui.Prompts[0], "Really delete")

	assert.Contains(t, ui.Outputs[0], "Deleting")
	assert.Equal(t, spaceRepo.DeletedSpace, reqFactory.Space)
	assert.Contains(t, ui.Outputs[1], "OK")
}

func TestDeleteSpaceConfirmingWithYes(t *testing.T) {
	ui, reqFactory, spaceRepo := deleteSpace("Yes", []string{"space-to-delete"})

	assert.Equal(t, reqFactory.SpaceName, "space-to-delete")
	assert.Contains(t, ui.Prompts[0], "Really delete")

	assert.Contains(t, ui.Outputs[0], "Deleting")
	assert.Equal(t, spaceRepo.DeletedSpace, reqFactory.Space)
	assert.Contains(t, ui.Outputs[1], "OK")
}

func TestDeleteSpaceWithForceOption(t *testing.T) {
	space := cf.Space{Name: "space-to-delete", Guid: "space-to-delete-guid"}
	reqFactory := &testhelpers.FakeReqFactory{Space: space}
	spaceRepo := &testhelpers.FakeSpaceRepository{}

	ui := &testhelpers.FakeUI{}
	ctxt := testhelpers.NewContext("delete", []string{"-f", "space-to-delete"})

	cmd := NewDeleteSpace(ui, spaceRepo)
	testhelpers.RunCommand(cmd, ctxt, reqFactory)

	assert.Equal(t, reqFactory.SpaceName, "space-to-delete")
	assert.Equal(t, len(ui.Prompts), 0)
	assert.Contains(t, ui.Outputs[0], "Deleting")
	assert.Contains(t, ui.Outputs[0], "space-to-delete")
	assert.Equal(t, spaceRepo.DeletedSpace, reqFactory.Space)
	assert.Contains(t, ui.Outputs[1], "OK")
}

func TestDeleteSpaceCommandFailsWithUsage(t *testing.T) {
	ui, _, _ := deleteSpace("Yes", []string{})
	assert.True(t, ui.FailedWithUsage)

	ui, _, _ = deleteSpace("Yes", []string{"space-to-delete"})
	assert.False(t, ui.FailedWithUsage)
}

func deleteSpace(confirmation string, args []string) (ui *testhelpers.FakeUI, reqFactory *testhelpers.FakeReqFactory, spaceRepo *testhelpers.FakeSpaceRepository) {
	space := cf.Space{Name: "space-to-delete", Guid: "space-to-delete-guid"}
	reqFactory = &testhelpers.FakeReqFactory{Space: space}
	spaceRepo = &testhelpers.FakeSpaceRepository{}
	ui = &testhelpers.FakeUI{
		Inputs: []string{confirmation},
	}
	ctxt := testhelpers.NewContext("delete-space", args)
	cmd := NewDeleteSpace(ui, spaceRepo)
	testhelpers.RunCommand(cmd, ctxt, reqFactory)
	return
}
