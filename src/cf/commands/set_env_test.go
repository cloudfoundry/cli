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

func TestSetEnvRequirements(t *testing.T) {
	config := &configuration.Configuration{}
	appRepo := &testhelpers.FakeApplicationRepository{}
	args := []string{"my-app", "DATABASE_URL", "mysql://example.com/my-db"}

	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true, SpaceSuccess: true}
	callSetEnv(args, config, reqFactory, appRepo)
	assert.True(t, testhelpers.CommandDidPassRequirements)

	reqFactory = &testhelpers.FakeReqFactory{LoginSuccess: false, SpaceSuccess: true}
	callSetEnv(args, config, reqFactory, appRepo)
	assert.False(t, testhelpers.CommandDidPassRequirements)

	testhelpers.CommandDidPassRequirements = true

	reqFactory = &testhelpers.FakeReqFactory{LoginSuccess: true, SpaceSuccess: false}
	callSetEnv(args, config, reqFactory, appRepo)
	assert.False(t, testhelpers.CommandDidPassRequirements)
}

func TestRunWhenApplicationExists(t *testing.T) {
	config := &configuration.Configuration{}
	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}
	reqFactory := &testhelpers.FakeReqFactory{Application: app, LoginSuccess: true, SpaceSuccess: true}
	appRepo := &testhelpers.FakeApplicationRepository{}

	args := []string{"my-app", "DATABASE_URL", "mysql://example.com/my-db"}
	ui := callSetEnv(args, config, reqFactory, appRepo)

	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[0], "DATABASE_URL")
	assert.Contains(t, ui.Outputs[1], "OK")

	assert.Equal(t, reqFactory.ApplicationName, "my-app")
	assert.Equal(t, appRepo.SetEnvApp, app)
	assert.Equal(t, appRepo.SetEnvName, "DATABASE_URL")
	assert.Equal(t, appRepo.SetEnvValue, "mysql://example.com/my-db")
}

func TestRunWhenSettingTheEnvFails(t *testing.T) {
	config := &configuration.Configuration{}
	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}
	reqFactory := &testhelpers.FakeReqFactory{Application: app, LoginSuccess: true, SpaceSuccess: true}
	appRepo := &testhelpers.FakeApplicationRepository{
		AppByName: app,
		SetEnvErr: true,
	}

	args := []string{"does-not-exist", "DATABASE_URL", "mysql://example.com/my-db"}
	ui := callSetEnv(args, config, reqFactory, appRepo)

	assert.Contains(t, ui.Outputs[0], "Updating env variable")
	assert.Contains(t, ui.Outputs[1], "FAILED")
	assert.Contains(t, ui.Outputs[2], "Failed setting env")
}

func TestSetEnvFailsWithUsage(t *testing.T) {
	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}
	reqFactory := &testhelpers.FakeReqFactory{Application: app, LoginSuccess: true, SpaceSuccess: true}
	appRepo := &testhelpers.FakeApplicationRepository{AppByName: app}
	config := &configuration.Configuration{}

	args := []string{"my-app", "DATABASE_URL", "..."}
	ui := callSetEnv(args, config, reqFactory, appRepo)
	assert.False(t, ui.FailedWithUsage)

	args = []string{"my-app", "DATABASE_URL"}
	ui = callSetEnv(args, config, reqFactory, appRepo)
	assert.True(t, ui.FailedWithUsage)

	args = []string{"my-app"}
	ui = callSetEnv(args, config, reqFactory, appRepo)
	assert.True(t, ui.FailedWithUsage)

	args = []string{}
	ui = callSetEnv(args, config, reqFactory, appRepo)
	assert.True(t, ui.FailedWithUsage)
}

func callSetEnv(args []string, config *configuration.Configuration, reqFactory *testhelpers.FakeReqFactory, appRepo api.ApplicationRepository) (ui *testhelpers.FakeUI) {
	ui = new(testhelpers.FakeUI)
	ctxt := testhelpers.NewContext("set-env", args)

	cmd := NewSetEnv(ui, config, appRepo)
	testhelpers.RunCommand(cmd, ctxt, reqFactory)
	return
}
