package application_test

import (
	"cf"
	. "cf/commands/application"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestDeleteConfirmingWithY(t *testing.T) {
	ui, _, appRepo := deleteApp(t, "y", []string{"app-to-delete"})

	assert.Equal(t, appRepo.FindByNameName, "app-to-delete")
	assert.Equal(t, appRepo.DeletedApp.Name, "app-to-delete")
	assert.Equal(t, len(ui.Outputs), 2)
	assert.Contains(t, ui.Prompts[0], "Really delete")
	assert.Contains(t, ui.Outputs[0], "Deleting")
	assert.Contains(t, ui.Outputs[0], "app-to-delete")
	assert.Contains(t, ui.Outputs[0], "my-org")
	assert.Contains(t, ui.Outputs[0], "my-space")
	assert.Contains(t, ui.Outputs[0], "my-user")
	assert.Contains(t, ui.Outputs[1], "OK")
}

func TestDeleteConfirmingWithYes(t *testing.T) {
	ui, _, appRepo := deleteApp(t, "Yes", []string{"app-to-delete"})

	assert.Equal(t, appRepo.FindByNameName, "app-to-delete")
	assert.Equal(t, appRepo.DeletedApp.Name, "app-to-delete")
	assert.Equal(t, len(ui.Outputs), 2)
	assert.Contains(t, ui.Prompts[0], "Really delete")
	assert.Contains(t, ui.Outputs[0], "Deleting")
	assert.Contains(t, ui.Outputs[0], "app-to-delete")
	assert.Contains(t, ui.Outputs[0], "my-org")
	assert.Contains(t, ui.Outputs[0], "my-space")
	assert.Contains(t, ui.Outputs[0], "my-user")
	assert.Contains(t, ui.Outputs[1], "OK")
}

func TestDeleteWithForceOption(t *testing.T) {
	app := cf.Application{Name: "app-to-delete", Guid: "app-to-delete-guid"}
	reqFactory := &testreq.FakeReqFactory{}
	appRepo := &testapi.FakeApplicationRepository{FindByNameApp: app}

	ui := &testterm.FakeUI{}
	ctxt := testcmd.NewContext("delete", []string{"-f", "app-to-delete"})

	cmd := NewDeleteApp(ui, &configuration.Configuration{}, appRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)

	assert.Equal(t, appRepo.FindByNameName, "app-to-delete")
	assert.Equal(t, appRepo.DeletedApp, app)
	assert.Equal(t, len(ui.Prompts), 0)
	assert.Equal(t, len(ui.Outputs), 2)
	assert.Contains(t, ui.Outputs[0], "Deleting")
	assert.Contains(t, ui.Outputs[0], "app-to-delete")
	assert.Contains(t, ui.Outputs[1], "OK")
}

func TestDeleteAppThatDoesNotExist(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	appRepo := &testapi.FakeApplicationRepository{FindByNameNotFound: true}

	ui := &testterm.FakeUI{}
	ctxt := testcmd.NewContext("delete", []string{"-f", "app-to-delete"})

	cmd := NewDeleteApp(ui, &configuration.Configuration{}, appRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)

	assert.Equal(t, appRepo.FindByNameName, "app-to-delete")
	assert.Equal(t, appRepo.DeletedApp.Name, "")
	assert.Contains(t, ui.Outputs[0], "Deleting")
	assert.Contains(t, ui.Outputs[0], "app-to-delete")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[2], "app-to-delete")
	assert.Contains(t, ui.Outputs[2], "does not exist")
}

func TestDeleteCommandFailsWithUsage(t *testing.T) {
	ui, _, _ := deleteApp(t, "Yes", []string{})
	assert.True(t, ui.FailedWithUsage)

	ui, _, _ = deleteApp(t, "Yes", []string{"app-to-delete"})
	assert.False(t, ui.FailedWithUsage)
}

func deleteApp(t *testing.T, confirmation string, args []string) (ui *testterm.FakeUI, reqFactory *testreq.FakeReqFactory, appRepo *testapi.FakeApplicationRepository) {
	app := cf.Application{Name: "app-to-delete", Guid: "app-to-delete-guid"}
	reqFactory = &testreq.FakeReqFactory{}
	appRepo = &testapi.FakeApplicationRepository{FindByNameApp: app}
	ui = &testterm.FakeUI{
		Inputs: []string{confirmation},
	}

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)

	config := &configuration.Configuration{
		Space:        cf.Space{Name: "my-space"},
		Organization: cf.Organization{Name: "my-org"},
		AccessToken:  token,
	}

	ctxt := testcmd.NewContext("delete", args)
	cmd := NewDeleteApp(ui, config, appRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
