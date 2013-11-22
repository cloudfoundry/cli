package application_test

import (
	"cf"
	. "cf/commands/application"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestEnvRequirements(t *testing.T) {
	reqFactory := getEnvDependencies()

	reqFactory.LoginSuccess = true
	callEnv(t, []string{"my-app"}, reqFactory)
	assert.True(t, testcmd.CommandDidPassRequirements)
	assert.Equal(t, reqFactory.ApplicationName, "my-app")

	reqFactory.LoginSuccess = false
	callEnv(t, []string{"my-app"}, reqFactory)
	assert.False(t, testcmd.CommandDidPassRequirements)
}

func TestEnvFailsWithUsage(t *testing.T) {
	reqFactory := getEnvDependencies()
	ui := callEnv(t, []string{}, reqFactory)

	assert.True(t, ui.FailedWithUsage)
	assert.False(t, testcmd.CommandDidPassRequirements)
}

func TestEnvListsEnvironmentVariables(t *testing.T) {
	reqFactory := getEnvDependencies()
	reqFactory.Application.EnvironmentVars = map[string]string{
		"my-key":  "my-value",
		"my-key2": "my-value2",
	}

	ui := callEnv(t, []string{"my-app"}, reqFactory)

	assert.Contains(t, ui.Outputs[0], "Getting env variables for app")
	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[0], "my-org")
	assert.Contains(t, ui.Outputs[0], "my-space")
	assert.Contains(t, ui.Outputs[0], "my-user")

	assert.Contains(t, ui.Outputs[1], "OK")

	assert.Contains(t, ui.Outputs[3], "my-key")
	assert.Contains(t, ui.Outputs[3], "my-value")
	assert.Contains(t, ui.Outputs[4], "my-key2")
	assert.Contains(t, ui.Outputs[4], "my-value2")
}

func TestEnvShowsEmptyMessage(t *testing.T) {
	reqFactory := getEnvDependencies()
	reqFactory.Application.EnvironmentVars = map[string]string{}

	ui := callEnv(t, []string{"my-app"}, reqFactory)

	assert.Contains(t, ui.Outputs[0], "Getting env variables for")
	assert.Contains(t, ui.Outputs[0], "my-app")

	assert.Contains(t, ui.Outputs[1], "OK")

	assert.Contains(t, ui.Outputs[3], "No env variables exist")
}

func callEnv(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}
	ctxt := testcmd.NewContext("env", args)

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

	cmd := NewEnv(ui, config)
	testcmd.RunCommand(cmd, ctxt, reqFactory)

	return
}

func getEnvDependencies() (reqFactory *testreq.FakeReqFactory) {
	app := cf.Application{}
	app.Name = "my-app"
	reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, Application: app}
	return
}
