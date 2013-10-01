package commands_test

import (
	"cf"
	. "cf/commands"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestDeleteSpaceConfirmingWithY(t *testing.T) {
	ui, spaceRepo := deleteSpace("y", []string{"space-to-delete"})

	assert.Equal(t, spaceRepo.SpaceName, "space-to-delete")
	assert.Contains(t, ui.Prompts[0], "Really delete")

	assert.Contains(t, ui.Outputs[0], "Deleting")
	assert.Contains(t, ui.Outputs[0], "space-to-delete")
	assert.Equal(t, spaceRepo.DeletedSpace, spaceRepo.SpaceByName)
	assert.Contains(t, ui.Outputs[1], "OK")
}

func TestDeleteSpaceConfirmingWithYes(t *testing.T) {
	ui, spaceRepo := deleteSpace("Yes", []string{"space-to-delete"})

	assert.Equal(t, spaceRepo.SpaceName, "space-to-delete")
	assert.Contains(t, ui.Prompts[0], "Really delete")

	assert.Contains(t, ui.Outputs[0], "Deleting")
	assert.Contains(t, ui.Outputs[0], "space-to-delete")
	assert.Equal(t, spaceRepo.DeletedSpace, spaceRepo.SpaceByName)
	assert.Contains(t, ui.Outputs[1], "OK")
}

func TestDeleteSpaceWithForceOption(t *testing.T) {
	space := cf.Space{Name: "space-to-delete", Guid: "space-to-delete-guid"}
	reqFactory := &testhelpers.FakeReqFactory{}
	spaceRepo := &testhelpers.FakeSpaceRepository{SpaceByName: space}
	configRepo := &testhelpers.FakeConfigRepository{}

	ui := &testhelpers.FakeUI{}
	ctxt := testhelpers.NewContext("delete", []string{"-f", "space-to-delete"})

	cmd := NewDeleteSpace(ui, spaceRepo, configRepo)
	testhelpers.RunCommand(cmd, ctxt, reqFactory)

	assert.Equal(t, spaceRepo.SpaceName, "space-to-delete")
	assert.Equal(t, len(ui.Prompts), 0)
	assert.Contains(t, ui.Outputs[0], "Deleting")
	assert.Contains(t, ui.Outputs[0], "space-to-delete")
	assert.Equal(t, spaceRepo.DeletedSpace, spaceRepo.SpaceByName)
	assert.Contains(t, ui.Outputs[1], "OK")
}

func TestDeleteSpaceWhenSpaceDoesNotExist(t *testing.T) {
	reqFactory := &testhelpers.FakeReqFactory{}
	spaceRepo := &testhelpers.FakeSpaceRepository{}
	configRepo := &testhelpers.FakeConfigRepository{}

	ui := &testhelpers.FakeUI{}
	ctxt := testhelpers.NewContext("delete", []string{"-f", "space-to-delete"})

	cmd := NewDeleteSpace(ui, spaceRepo, configRepo)
	testhelpers.RunCommand(cmd, ctxt, reqFactory)

	assert.Equal(t, len(ui.Outputs), 3)
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[2], "space-to-delete")
	assert.Contains(t, ui.Outputs[2], "was already deleted.")
}

func TestDeleteSpaceWhenSpaceIsTargeted(t *testing.T) {
	space := cf.Space{Name: "space-to-delete", Guid: "space-to-delete-guid"}
	reqFactory := &testhelpers.FakeReqFactory{}
	spaceRepo := &testhelpers.FakeSpaceRepository{SpaceByName: space}
	configRepo := &testhelpers.FakeConfigRepository{}

	config, _ := configRepo.Get()
	config.Space = space
	configRepo.Save(config)

	ui := &testhelpers.FakeUI{}
	ctxt := testhelpers.NewContext("delete", []string{"-f", "space-to-delete"})

	cmd := NewDeleteSpace(ui, spaceRepo, configRepo)
	testhelpers.RunCommand(cmd, ctxt, reqFactory)

	config, _ = configRepo.Get()
	assert.Equal(t, config.HasSpace(), false)
}

func TestDeleteSpaceWhenSpaceNotTargeted(t *testing.T) {
	space := cf.Space{Name: "space-to-delete", Guid: "space-to-delete-guid"}
	reqFactory := &testhelpers.FakeReqFactory{}
	spaceRepo := &testhelpers.FakeSpaceRepository{SpaceByName: space}
	configRepo := &testhelpers.FakeConfigRepository{}

	config, _ := configRepo.Get()
	config.Space = cf.Space{Name: "do-not-delete", Guid: "do-not-delete-guid"}
	configRepo.Save(config)

	ui := &testhelpers.FakeUI{}
	ctxt := testhelpers.NewContext("delete", []string{"-f", "space-to-delete"})

	cmd := NewDeleteSpace(ui, spaceRepo, configRepo)
	testhelpers.RunCommand(cmd, ctxt, reqFactory)

	config, _ = configRepo.Get()
	assert.Equal(t, config.HasSpace(), true)
}

func TestDeleteSpaceCommandWith(t *testing.T) {
	ui, _ := deleteSpace("Yes", []string{})
	assert.True(t, ui.FailedWithUsage)

	ui, _ = deleteSpace("Yes", []string{"space-to-delete"})
	assert.False(t, ui.FailedWithUsage)
}

func TestDeleteSpaceCommandFailsWithUsage(t *testing.T) {
	ui, _ := deleteSpace("Yes", []string{})
	assert.True(t, ui.FailedWithUsage)

	ui, _ = deleteSpace("Yes", []string{"space-to-delete"})
	assert.False(t, ui.FailedWithUsage)
}

func deleteSpace(confirmation string, args []string) (ui *testhelpers.FakeUI, spaceRepo *testhelpers.FakeSpaceRepository) {
	space := cf.Space{Name: "space-to-delete", Guid: "space-to-delete-guid"}
	reqFactory := &testhelpers.FakeReqFactory{}
	spaceRepo = &testhelpers.FakeSpaceRepository{SpaceByName: space}
	configRepo := &testhelpers.FakeConfigRepository{}

	ui = &testhelpers.FakeUI{
		Inputs: []string{confirmation},
	}
	ctxt := testhelpers.NewContext("delete-space", args)
	cmd := NewDeleteSpace(ui, spaceRepo, configRepo)
	testhelpers.RunCommand(cmd, ctxt, reqFactory)
	return
}
