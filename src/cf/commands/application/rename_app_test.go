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

func TestRenameAppFailsWithUsage(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	appRepo := &testapi.FakeApplicationRepository{}

	ui := callRename(t, []string{}, reqFactory, appRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callRename(t, []string{"foo"}, reqFactory, appRepo)
	assert.True(t, ui.FailedWithUsage)
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

	assert.Equal(t, appRepo.UpdateAppGuid, app.Guid)
	assert.Equal(t, appRepo.UpdateParams.Get("name"), "my-new-app")
	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Renaming app", "my-app", "my-new-app", "my-org", "my-space", "my-user"},
		{"OK"},
	})
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
