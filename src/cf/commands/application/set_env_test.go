package application_test

import (
	"cf"
	"cf/api"
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

func TestSetEnvRequirements(t *testing.T) {
	app := cf.Application{}
	app.Name = "my-app"
	app.Guid = "my-app-guid"
	appRepo := &testapi.FakeApplicationRepository{}
	args := []string{"my-app", "DATABASE_URL", "mysql://example.com/my-db"}

	reqFactory := &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: true}
	callSetEnv(t, args, reqFactory, appRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{Application: app, LoginSuccess: false, TargetedSpaceSuccess: true}
	callSetEnv(t, args, reqFactory, appRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)

	testcmd.CommandDidPassRequirements = true

	reqFactory = &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: false}
	callSetEnv(t, args, reqFactory, appRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)
}

func TestRunWhenApplicationExists(t *testing.T) {
	app := cf.Application{}
	app.Name = "my-app"
	app.Guid = "my-app-guid"
	app.EnvironmentVars = map[string]string{"foo": "bar"}
	reqFactory := &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: true}
	appRepo := &testapi.FakeApplicationRepository{}

	args := []string{"my-app", "DATABASE_URL", "mysql://example.com/my-db"}
	ui := callSetEnv(t, args, reqFactory, appRepo)

	assert.Contains(t, ui.Outputs[0], "Setting env variable")
	assert.Contains(t, ui.Outputs[0], "DATABASE_URL")
	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[0], "my-org")
	assert.Contains(t, ui.Outputs[0], "my-space")
	assert.Contains(t, ui.Outputs[0], "my-user")
	assert.Contains(t, ui.Outputs[1], "OK")

	assert.Equal(t, reqFactory.ApplicationName, "my-app")
	assert.Equal(t, appRepo.SetEnvAppGuid, app.Guid)
	assert.Equal(t, appRepo.SetEnvVars, map[string]string{
		"DATABASE_URL": "mysql://example.com/my-db",
		"foo":          "bar",
	})
}

func TestSetEnvWhenItAlreadyExists(t *testing.T) {
	app := cf.Application{}
	app.Name = "my-app"
	app.Guid = "my-app-guid"
	app.EnvironmentVars = map[string]string{"DATABASE_URL": "mysql://example.com/my-db"}
	reqFactory := &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: true}
	appRepo := &testapi.FakeApplicationRepository{}

	args := []string{"my-app", "DATABASE_URL", "mysql://example.com/my-db"}
	ui := callSetEnv(t, args, reqFactory, appRepo)

	assert.Equal(t, len(ui.Outputs), 3)
	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[0], "DATABASE_URL")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[2], "DATABASE_URL")
	assert.Contains(t, ui.Outputs[2], "was already set.")

}

func TestRunWhenSettingTheEnvFails(t *testing.T) {
	app := cf.Application{}
	app.Name = "my-app"
	app.Guid = "my-app-guid"
	reqFactory := &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: true}
	appRepo := &testapi.FakeApplicationRepository{
		FindByNameApp: app,
		SetEnvErr:     true,
	}

	args := []string{"does-not-exist", "DATABASE_URL", "mysql://example.com/my-db"}
	ui := callSetEnv(t, args, reqFactory, appRepo)

	assert.Contains(t, ui.Outputs[0], "Setting env variable")
	assert.Contains(t, ui.Outputs[1], "FAILED")
	assert.Contains(t, ui.Outputs[2], "Failed setting env")
}

func TestSetEnvFailsWithUsage(t *testing.T) {
	app := cf.Application{}
	app.Name = "my-app"
	app.Guid = "my-app-guid"
	reqFactory := &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: true}
	appRepo := &testapi.FakeApplicationRepository{FindByNameApp: app}

	args := []string{"my-app", "DATABASE_URL", "..."}
	ui := callSetEnv(t, args, reqFactory, appRepo)
	assert.False(t, ui.FailedWithUsage)

	args = []string{"my-app", "DATABASE_URL"}
	ui = callSetEnv(t, args, reqFactory, appRepo)
	assert.True(t, ui.FailedWithUsage)

	args = []string{"my-app"}
	ui = callSetEnv(t, args, reqFactory, appRepo)
	assert.True(t, ui.FailedWithUsage)

	args = []string{}
	ui = callSetEnv(t, args, reqFactory, appRepo)
	assert.True(t, ui.FailedWithUsage)
}

func callSetEnv(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory, appRepo api.ApplicationRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("set-env", args)

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

	cmd := NewSetEnv(ui, config, appRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
