package application_test

import (
	"cf"
	. "cf/commands/application"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestDeleteConfirmingWithY(t *testing.T) {
	ui, _, appRepo := deleteApp(t, "y", []string{"app-to-delete"})

	assert.Equal(t, appRepo.ReadName, "app-to-delete")
	assert.Equal(t, appRepo.DeletedAppGuid, "app-to-delete-guid")
	testassert.SliceContains(t, ui.Prompts, testassert.Lines{
		{"Really delete"},
	})
	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Deleting", "app-to-delete", "my-org", "my-space", "my-user"},
		{"OK"},
	})
}

func TestDeleteConfirmingWithYes(t *testing.T) {
	ui, _, appRepo := deleteApp(t, "Yes", []string{"app-to-delete"})

	assert.Equal(t, appRepo.ReadName, "app-to-delete")
	assert.Equal(t, appRepo.DeletedAppGuid, "app-to-delete-guid")

	testassert.SliceContains(t, ui.Prompts, testassert.Lines{
		{"Really delete", "app-to-delete"},
	})
	assert.Equal(t, len(ui.Outputs), 2)
	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Deleting", "app-to-delete", "my-org", "my-space", "my-user"},
		{"OK"},
	})
}

func TestDeleteWithForceOption(t *testing.T) {
	app := cf.Application{}
	app.Name = "app-to-delete"
	app.Guid = "app-to-delete-guid"

	reqFactory := &testreq.FakeReqFactory{}
	appRepo := &testapi.FakeApplicationRepository{ReadApp: app}

	ui := &testterm.FakeUI{}
	ctxt := testcmd.NewContext("delete", []string{"-f", "app-to-delete"})

	cmd := NewDeleteApp(ui, &configuration.Configuration{}, appRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)

	assert.Equal(t, appRepo.ReadName, "app-to-delete")
	assert.Equal(t, appRepo.DeletedAppGuid, "app-to-delete-guid")
	assert.Equal(t, len(ui.Prompts), 0)
	assert.Equal(t, len(ui.Outputs), 2)
	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Deleting", "app-to-delete"},
		{"OK"},
	})
}

func TestDeleteAppThatDoesNotExist(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	appRepo := &testapi.FakeApplicationRepository{ReadNotFound: true}

	ui := &testterm.FakeUI{}
	ctxt := testcmd.NewContext("delete", []string{"-f", "app-to-delete"})

	cmd := NewDeleteApp(ui, &configuration.Configuration{}, appRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)

	assert.Equal(t, appRepo.ReadName, "app-to-delete")
	assert.Equal(t, appRepo.DeletedAppGuid, "")

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Deleting", "app-to-delete"},
		{"OK"},
		{"app-to-delete", "does not exist"},
	})
}

func TestDeleteCommandFailsWithUsage(t *testing.T) {
	ui, _, _ := deleteApp(t, "Yes", []string{})
	assert.True(t, ui.FailedWithUsage)

	ui, _, _ = deleteApp(t, "Yes", []string{"app-to-delete"})
	assert.False(t, ui.FailedWithUsage)
}

func deleteApp(t *testing.T, confirmation string, args []string) (ui *testterm.FakeUI, reqFactory *testreq.FakeReqFactory, appRepo *testapi.FakeApplicationRepository) {

	app := cf.Application{}
	app.Name = "app-to-delete"
	app.Guid = "app-to-delete-guid"

	reqFactory = &testreq.FakeReqFactory{}
	appRepo = &testapi.FakeApplicationRepository{ReadApp: app}
	ui = &testterm.FakeUI{
		Inputs: []string{confirmation},
	}

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)

	org := cf.OrganizationFields{}
	org.Name = "my-org"
	space := cf.SpaceFields{}
	space.Name = "my-space"
	config := &configuration.Configuration{
		SpaceFields:        space,
		OrganizationFields: org,
		AccessToken:        token,
	}

	ctxt := testcmd.NewContext("delete", args)
	cmd := NewDeleteApp(ui, config, appRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
