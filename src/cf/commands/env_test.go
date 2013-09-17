package commands_test

import (
	"cf"
	. "cf/commands"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestEnvRequirements(t *testing.T) {
	reqFactory := getDefaultEnvDependencies()

	reqFactory.LoginSuccess = true
	callEnv([]string{"my-app"}, reqFactory)
	assert.True(t, testhelpers.CommandDidPassRequirements)
	assert.Equal(t, reqFactory.ApplicationName, "my-app")

	reqFactory.LoginSuccess = false
	callEnv([]string{"my-app"}, reqFactory)
	assert.False(t, testhelpers.CommandDidPassRequirements)
}

func TestEnvFailsWithUsage(t *testing.T) {
	reqFactory := getDefaultEnvDependencies()
	ui := callEnv([]string{}, reqFactory)

	assert.True(t, ui.FailedWithUsage)
	assert.False(t, testhelpers.CommandDidPassRequirements)
}

func TestEnvListsEnvironmentVariables(t *testing.T) {
	reqFactory := getDefaultEnvDependencies()
	reqFactory.Application.EnvironmentVars = map[string]string{
		"my-key":  "my-value",
		"my-key2": "my-value2",
	}

	ui := callEnv([]string{"my-app"}, reqFactory)

	assert.Contains(t, ui.Outputs[0], "Getting env variables for")
	assert.Contains(t, ui.Outputs[0], "my-app")

	assert.Contains(t, ui.Outputs[1], "OK")

	assert.Contains(t, ui.Outputs[2], "my-key")
	assert.Contains(t, ui.Outputs[2], "my-value")
	assert.Contains(t, ui.Outputs[3], "my-key2")
	assert.Contains(t, ui.Outputs[3], "my-value2")
}

func TestEnvShowsEmptyMessage(t *testing.T) {
	reqFactory := getDefaultEnvDependencies()
	reqFactory.Application.EnvironmentVars = map[string]string{}

	ui := callEnv([]string{"my-app"}, reqFactory)

	assert.Contains(t, ui.Outputs[0], "Getting env variables for")
	assert.Contains(t, ui.Outputs[0], "my-app")

	assert.Contains(t, ui.Outputs[1], "OK")

	assert.Contains(t, ui.Outputs[2], "No env variables exist")
}

func callEnv(args []string, reqFactory *testhelpers.FakeReqFactory) (ui *testhelpers.FakeUI) {
	ui = &testhelpers.FakeUI{}
	ctxt := testhelpers.NewContext("env", args)
	cmd := NewEnv(ui)
	testhelpers.RunCommand(cmd, ctxt, reqFactory)

	return
}

func getDefaultEnvDependencies() (reqFactory *testhelpers.FakeReqFactory) {
	reqFactory = &testhelpers.FakeReqFactory{LoginSuccess: true, Application: cf.Application{Name: "my-app"}}
	return
}
