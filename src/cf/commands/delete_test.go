package commands_test

import (
	"cf"
	. "cf/commands"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestDeleteConfirmingWithY(t *testing.T) {
	ui, _, appRepo := deleteApp("y", []string{"app-to-delete"})

	assert.Equal(t, appRepo.AppName, "app-to-delete")
	assert.Equal(t, len(ui.Outputs), 2)
	assert.Contains(t, ui.Prompts[0], "Really delete")
	assert.Contains(t, ui.Outputs[0], "app-to-delete")
	assert.Contains(t, ui.Outputs[0], "Deleting")
	assert.Contains(t, ui.Outputs[1], "OK")
}

func TestDeleteConfirmingWithYes(t *testing.T) {
	ui, _, appRepo := deleteApp("Yes", []string{"app-to-delete"})

	assert.Equal(t, appRepo.AppName, "app-to-delete")
	assert.Equal(t, len(ui.Outputs), 2)
	assert.Contains(t, ui.Prompts[0], "Really delete")
	assert.Contains(t, ui.Outputs[0], "app-to-delete")
	assert.Contains(t, ui.Outputs[0], "Deleting")
	assert.Contains(t, ui.Outputs[1], "OK")
}

func TestDeleteWithForceOption(t *testing.T) {
	app := cf.Application{Name: "app-to-delete", Guid: "app-to-delete-guid"}
	reqFactory := &testhelpers.FakeReqFactory{}
	appRepo := &testhelpers.FakeApplicationRepository{AppByName: app}

	ui := &testhelpers.FakeUI{}
	ctxt := testhelpers.NewContext("delete", []string{"-f", "app-to-delete"})

	cmd := NewDelete(ui, appRepo)
	testhelpers.RunCommand(cmd, ctxt, reqFactory)

	assert.Equal(t, appRepo.AppName, "app-to-delete")
	assert.Equal(t, len(ui.Prompts), 0)
	assert.Equal(t, len(ui.Outputs), 2)
	assert.Contains(t, ui.Outputs[0], "Deleting")
	assert.Contains(t, ui.Outputs[0], "app-to-delete")
	assert.Contains(t, ui.Outputs[1], "OK")
}

func TestDeleteAppThatDoesNotExist(t *testing.T) {
	reqFactory := &testhelpers.FakeReqFactory{}
	appRepo := &testhelpers.FakeApplicationRepository{DidNotFindByName: true}

	ui := &testhelpers.FakeUI{}
	ctxt := testhelpers.NewContext("delete", []string{"-f", "app-to-delete"})

	cmd := NewDelete(ui, appRepo)
	testhelpers.RunCommand(cmd, ctxt, reqFactory)

	assert.Equal(t, appRepo.AppName, "app-to-delete")
	assert.Contains(t, ui.Outputs[0], "Deleting")
	assert.Contains(t, ui.Outputs[0], "app-to-delete")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[2], "app-to-delete")
	assert.Contains(t, ui.Outputs[2], "does not exist")
}

func TestDeleteCommandFailsWithUsage(t *testing.T) {
	ui, _, _ := deleteApp("Yes", []string{})
	assert.True(t, ui.FailedWithUsage)

	ui, _, _ = deleteApp("Yes", []string{"app-to-delete"})
	assert.False(t, ui.FailedWithUsage)
}

func deleteApp(confirmation string, args []string) (ui *testhelpers.FakeUI, reqFactory *testhelpers.FakeReqFactory, appRepo *testhelpers.FakeApplicationRepository) {
	app := cf.Application{Name: "app-to-delete", Guid: "app-to-delete-guid"}
	reqFactory = &testhelpers.FakeReqFactory{}
	appRepo = &testhelpers.FakeApplicationRepository{AppByName: app}
	ui = &testhelpers.FakeUI{
		Inputs: []string{confirmation},
	}
	ctxt := testhelpers.NewContext("delete", args)
	cmd := NewDelete(ui, appRepo)
	testhelpers.RunCommand(cmd, ctxt, reqFactory)
	return
}
