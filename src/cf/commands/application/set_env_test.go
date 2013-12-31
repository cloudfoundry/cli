package application_test

import (
	"cf"
	"cf/api"
	. "cf/commands/application"
	"cf/configuration"
	"generic"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
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

	assert.Equal(t, len(ui.Outputs), 3)
	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{
			"Setting env variable",
			"DATABASE_URL",
			"mysql://example.com/my-db",
			"my-app",
			"my-org",
			"my-space",
			"my-user",
		},
		{"OK"},
		{"TIP"},
	})

	assert.Equal(t, reqFactory.ApplicationName, "my-app")
	assert.Equal(t, appRepo.UpdateAppGuid, app.Guid)

	envParams := appRepo.UpdateParams.Get("env").(generic.Map)
	assert.Equal(t, envParams.Get("DATABASE_URL").(string), "mysql://example.com/my-db")
	assert.Equal(t, envParams.Get("foo").(string), "bar")
}

func TestSetEnvWhenItAlreadyExists(t *testing.T) {
	app := cf.Application{}
	app.Name = "my-app"
	app.Guid = "my-app-guid"
	app.EnvironmentVars = map[string]string{"DATABASE_URL": "mysql://example.com/my-db"}
	reqFactory := &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: true}
	appRepo := &testapi.FakeApplicationRepository{}

	args := []string{"my-app", "DATABASE_URL", "mysql://example2.com/my-db"}
	ui := callSetEnv(t, args, reqFactory, appRepo)

	assert.Equal(t, len(ui.Outputs), 3)
	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{
			"Setting env variable",
			"DATABASE_URL",
			"mysql://example2.com/my-db",
			"my-app",
			"my-org",
			"my-space",
			"my-user",
		},
		{"OK"},
		{"TIP"},
	})

	assert.Equal(t, reqFactory.ApplicationName, "my-app")
	assert.Equal(t, appRepo.UpdateAppGuid, app.Guid)

	envParams := appRepo.UpdateParams.Get("env").(generic.Map)
	assert.Equal(t, envParams.Get("DATABASE_URL").(string), "mysql://example2.com/my-db")
}

func TestRunWhenSettingTheEnvFails(t *testing.T) {
	app := cf.Application{}
	app.Name = "my-app"
	app.Guid = "my-app-guid"
	reqFactory := &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: true}
	appRepo := &testapi.FakeApplicationRepository{
		ReadApp:   app,
		UpdateErr: true,
	}

	args := []string{"does-not-exist", "DATABASE_URL", "mysql://example.com/my-db"}
	ui := callSetEnv(t, args, reqFactory, appRepo)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Setting env variable"},
		{"FAILED"},
		{"Error updating app."},
	})
}

func TestSetEnvFailsWithUsage(t *testing.T) {
	app := cf.Application{}
	app.Name = "my-app"
	app.Guid = "my-app-guid"
	reqFactory := &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: true}
	appRepo := &testapi.FakeApplicationRepository{ReadApp: app}

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
