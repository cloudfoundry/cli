package application_test

import (
	"cf"
	. "cf/commands/application"
	"github.com/stretchr/testify/assert"
	testcmd "testhelpers/commands"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestEnvRequirements(t *testing.T) {
	reqFactory := getEnvDependencies()

	reqFactory.LoginSuccess = true
	callEnv([]string{"my-app"}, reqFactory)
	assert.True(t, testcmd.CommandDidPassRequirements)
	assert.Equal(t, reqFactory.ApplicationName, "my-app")

	reqFactory.LoginSuccess = false
	callEnv([]string{"my-app"}, reqFactory)
	assert.False(t, testcmd.CommandDidPassRequirements)
}

func TestEnvFailsWithUsage(t *testing.T) {
	reqFactory := getEnvDependencies()
	ui := callEnv([]string{}, reqFactory)

	assert.True(t, ui.FailedWithUsage)
	assert.False(t, testcmd.CommandDidPassRequirements)
}

func TestEnvListsEnvironmentVariables(t *testing.T) {
	reqFactory := getEnvDependencies()
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
	reqFactory := getEnvDependencies()
	reqFactory.Application.EnvironmentVars = map[string]string{}

	ui := callEnv([]string{"my-app"}, reqFactory)

	assert.Contains(t, ui.Outputs[0], "Getting env variables for")
	assert.Contains(t, ui.Outputs[0], "my-app")

	assert.Contains(t, ui.Outputs[1], "OK")

	assert.Contains(t, ui.Outputs[2], "No env variables exist")
}

func callEnv(args []string, reqFactory *testreq.FakeReqFactory) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}
	ctxt := testcmd.NewContext("env", args)
	cmd := NewEnv(ui)
	testcmd.RunCommand(cmd, ctxt, reqFactory)

	return
}

func getEnvDependencies() (reqFactory *testreq.FakeReqFactory) {
	reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, Application: cf.Application{Name: "my-app"}}
	return
}
