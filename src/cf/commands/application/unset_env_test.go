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

func TestUnsetEnvRequirements(t *testing.T) {
	app := cf.Application{}
	app.Name = "my-app"
	app.Guid = "my-app-guid"
	appRepo := &testapi.FakeApplicationRepository{}
	args := []string{"my-app", "DATABASE_URL"}

	reqFactory := &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: true}
	callUnsetEnv(t, args, reqFactory, appRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{Application: app, LoginSuccess: false, TargetedSpaceSuccess: true}
	callUnsetEnv(t, args, reqFactory, appRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: false}
	callUnsetEnv(t, args, reqFactory, appRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)
}

func TestUnsetEnvWhenApplicationExists(t *testing.T) {
	app := cf.Application{}
	app.Name = "my-app"
	app.Guid = "my-app-guid"
	app.EnvironmentVars = map[string]string{"foo": "bar", "DATABASE_URL": "mysql://example.com/my-db"}
	reqFactory := &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: true}
	appRepo := &testapi.FakeApplicationRepository{}

	args := []string{"my-app", "DATABASE_URL"}
	ui := callUnsetEnv(t, args, reqFactory, appRepo)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Removing env variable", "DATABASE_URL", "my-app", "my-org", "my-space", "my-user"},
		{"OK"},
	})

	assert.Equal(t, reqFactory.ApplicationName, "my-app")
	assert.Equal(t, appRepo.UpdateAppGuid, "my-app-guid")

	envParams := appRepo.UpdateParams.Get("env").(generic.Map)
	assert.Equal(t, envParams.Get("foo").(string), "bar")
}

func TestUnsetEnvWhenUnsettingTheEnvFails(t *testing.T) {
	app := cf.Application{}
	app.Name = "my-app"
	app.Guid = "my-app-guid"
	app.EnvironmentVars = map[string]string{"DATABASE_URL": "mysql://example.com/my-db"}
	reqFactory := &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: true}
	appRepo := &testapi.FakeApplicationRepository{
		ReadApp:   app,
		UpdateErr: true,
	}

	args := []string{"does-not-exist", "DATABASE_URL"}
	ui := callUnsetEnv(t, args, reqFactory, appRepo)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Removing env variable"},
		{"FAILED"},
		{"Error updating app."},
	})
}

func TestUnsetEnvWhenEnvVarDoesNotExist(t *testing.T) {
	app := cf.Application{}
	app.Name = "my-app"
	app.Guid = "my-app-guid"
	reqFactory := &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: true}
	appRepo := &testapi.FakeApplicationRepository{}

	args := []string{"my-app", "DATABASE_URL"}
	ui := callUnsetEnv(t, args, reqFactory, appRepo)

	assert.Equal(t, len(ui.Outputs), 3)
	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Removing env variable"},
		{"OK"},
		{"DATABASE_URL", "was not set."},
	})
}

func TestUnsetEnvFailsWithUsage(t *testing.T) {
	app := cf.Application{}
	app.Name = "my-app"
	app.Guid = "my-app-guid"
	reqFactory := &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: true}
	appRepo := &testapi.FakeApplicationRepository{ReadApp: app}

	args := []string{"my-app", "DATABASE_URL"}
	ui := callUnsetEnv(t, args, reqFactory, appRepo)
	assert.False(t, ui.FailedWithUsage)

	args = []string{"my-app"}
	ui = callUnsetEnv(t, args, reqFactory, appRepo)
	assert.True(t, ui.FailedWithUsage)

	args = []string{}
	ui = callUnsetEnv(t, args, reqFactory, appRepo)
	assert.True(t, ui.FailedWithUsage)
}

func callUnsetEnv(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory, appRepo api.ApplicationRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("unset-env", args)

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

	cmd := NewUnsetEnv(ui, config, appRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
