package commands_test

import (
	"cf"
	"cf/api"
	. "cf/commands"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestRunWhenApplicationExists(t *testing.T) {
	config := &configuration.Configuration{}
	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}
	appRepo := &testhelpers.FakeApplicationRepository{AppByName: app}

	args := []string{"my-app", "DATABASE_URL", "mysql://example.com/my-db"}
	ui := callSetEnv(args, config, appRepo)

	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[0], "DATABASE_URL")
	assert.Contains(t, ui.Outputs[1], "OK")

	assert.Equal(t, appRepo.AppName, "my-app")
	assert.Equal(t, appRepo.SetEnvApp, app)
	assert.Equal(t, appRepo.SetEnvName, "DATABASE_URL")
	assert.Equal(t, appRepo.SetEnvValue, "mysql://example.com/my-db")
}

func TestRunWhenAppDoesNotExist(t *testing.T) {
	config := &configuration.Configuration{}
	appRepo := &testhelpers.FakeApplicationRepository{AppByNameErr: true}

	args := []string{"does-not-exist", "DATABASE_URL", "mysql://example.com/my-db"}
	ui := callSetEnv(args, config, appRepo)

	assert.Contains(t, ui.Outputs[0], "FAILED")
	assert.Contains(t, ui.Outputs[1], "App does not exist.")
}

func TestRunWhenSettingTheEnvFails(t *testing.T) {
	config := &configuration.Configuration{}
	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}
	appRepo := &testhelpers.FakeApplicationRepository{
		AppByName: app,
		SetEnvErr: true,
	}

	args := []string{"does-not-exist", "DATABASE_URL", "mysql://example.com/my-db"}
	ui := callSetEnv(args, config, appRepo)

	assert.Contains(t, ui.Outputs[0], "Updating env variable")
	assert.Contains(t, ui.Outputs[1], "FAILED")
	assert.Contains(t, ui.Outputs[2], "Failed setting env")
}

func TestRunReportsWhenArgumentsAreMissing(t *testing.T) {
	config := &configuration.Configuration{}
	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}
	appRepo := &testhelpers.FakeApplicationRepository{
		AppByName: app,
	}

	args := []string{"my-app", "DATABASE_URL"}
	ui := callSetEnv(args, config, appRepo)

	assert.Contains(t, ui.Outputs[0], "FAILED")
	assert.Contains(t, ui.Outputs[1], "Please enter app name, variable name and value.")

	args = []string{"my-app"}
	ui = callSetEnv(args, config, appRepo)

	assert.Contains(t, ui.Outputs[0], "FAILED")
	assert.Contains(t, ui.Outputs[1], "Please enter app name, variable name and value.")

	args = []string{}
	ui = callSetEnv(args, config, appRepo)

	assert.Contains(t, ui.Outputs[0], "FAILED")
	assert.Contains(t, ui.Outputs[1], "Please enter app name, variable name and value.")
}

func callSetEnv(args []string, config *configuration.Configuration, appRepo api.ApplicationRepository) (ui *testhelpers.FakeUI) {
	context := testhelpers.NewContext("set-env", args)
	ui = new(testhelpers.FakeUI)
	se := NewSetEnv(ui, config, appRepo)
	se.Run(context)

	return
}
