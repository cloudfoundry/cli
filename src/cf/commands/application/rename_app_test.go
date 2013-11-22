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

func TestRenameAppFailsWithUsage(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	appRepo := &testapi.FakeApplicationRepository{}

	fakeUI := callRename(t, []string{}, reqFactory, appRepo)
	assert.True(t, fakeUI.FailedWithUsage)

	fakeUI = callRename(t, []string{"foo"}, reqFactory, appRepo)
	assert.True(t, fakeUI.FailedWithUsage)
}

func TestRenameRequirements(t *testing.T) {
	appRepo := &testapi.FakeApplicationRepository{}

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	callRename(t, []string{"my-app", "my-new-app"}, reqFactory, appRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)
	assert.Equal(t, reqFactory.ApplicationName, "my-app")
}

func TestRenameRun(t *testing.T) {
	appRepo := &testapi.FakeApplicationRepository{}
	app := cf.Application{}
	app.Name = "my-app"
	app.Guid = "my-app-guid"
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, Application: app}
	ui := callRename(t, []string{"my-app", "my-new-app"}, reqFactory, appRepo)

	assert.Contains(t, ui.Outputs[0], "Renaming app ")
	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[0], "my-new-app")
	assert.Contains(t, ui.Outputs[0], "my-org")
	assert.Contains(t, ui.Outputs[0], "my-space")
	assert.Contains(t, ui.Outputs[0], "my-user")
	assert.Equal(t, appRepo.RenameAppGuid, app.Guid)
	assert.Equal(t, appRepo.RenameNewName, "my-new-app")
	assert.Contains(t, ui.Outputs[1], "OK")
}

func callRename(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory, appRepo *testapi.FakeApplicationRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("rename", args)

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

	cmd := NewRenameApp(ui, config, appRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
