package application_test

import (
	"cf"
	"cf/api"
	. "cf/commands/application"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestSetEnvRequirements(t *testing.T) {
	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}
	appRepo := &testapi.FakeApplicationRepository{}
	args := []string{"my-app", "DATABASE_URL", "mysql://example.com/my-db"}

	reqFactory := &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: true}
	callSetEnv(args, reqFactory, appRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{Application: app, LoginSuccess: false, TargetedSpaceSuccess: true}
	callSetEnv(args, reqFactory, appRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)

	testcmd.CommandDidPassRequirements = true

	reqFactory = &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: false}
	callSetEnv(args, reqFactory, appRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)
}

func TestRunWhenApplicationExists(t *testing.T) {
	app := cf.Application{Name: "my-app", Guid: "my-app-guid", EnvironmentVars: map[string]string{"foo": "bar"}}
	reqFactory := &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: true}
	appRepo := &testapi.FakeApplicationRepository{}

	args := []string{"my-app", "DATABASE_URL", "mysql://example.com/my-db"}
	ui := callSetEnv(args, reqFactory, appRepo)

	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[0], "DATABASE_URL")
	assert.Contains(t, ui.Outputs[1], "OK")

	assert.Equal(t, reqFactory.ApplicationName, "my-app")
	assert.Equal(t, appRepo.SetEnvApp, app)
	assert.Equal(t, appRepo.SetEnvVars, map[string]string{
		"DATABASE_URL": "mysql://example.com/my-db",
		"foo":          "bar",
	})
}

func TestSetEnvWhenItAlreadyExists(t *testing.T) {
	app := cf.Application{Name: "my-app", Guid: "my-app-guid", EnvironmentVars: map[string]string{"DATABASE_URL": "mysql://example.com/my-db"}}
	reqFactory := &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: true}
	appRepo := &testapi.FakeApplicationRepository{}

	args := []string{"my-app", "DATABASE_URL", "mysql://example.com/my-db"}
	ui := callSetEnv(args, reqFactory, appRepo)

	assert.Equal(t, len(ui.Outputs), 3)
	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[0], "DATABASE_URL")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[2], "DATABASE_URL")
	assert.Contains(t, ui.Outputs[2], "was already set.")

}

func TestRunWhenSettingTheEnvFails(t *testing.T) {
	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}
	reqFactory := &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: true}
	appRepo := &testapi.FakeApplicationRepository{
		FindByNameApp: app,
		SetEnvErr:     true,
	}

	args := []string{"does-not-exist", "DATABASE_URL", "mysql://example.com/my-db"}
	ui := callSetEnv(args, reqFactory, appRepo)

	assert.Contains(t, ui.Outputs[0], "Updating env variable")
	assert.Contains(t, ui.Outputs[1], "FAILED")
	assert.Contains(t, ui.Outputs[2], "Failed setting env")
}

func TestSetEnvFailsWithUsage(t *testing.T) {
	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}
	reqFactory := &testreq.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: true}
	appRepo := &testapi.FakeApplicationRepository{FindByNameApp: app}

	args := []string{"my-app", "DATABASE_URL", "..."}
	ui := callSetEnv(args, reqFactory, appRepo)
	assert.False(t, ui.FailedWithUsage)

	args = []string{"my-app", "DATABASE_URL"}
	ui = callSetEnv(args, reqFactory, appRepo)
	assert.True(t, ui.FailedWithUsage)

	args = []string{"my-app"}
	ui = callSetEnv(args, reqFactory, appRepo)
	assert.True(t, ui.FailedWithUsage)

	args = []string{}
	ui = callSetEnv(args, reqFactory, appRepo)
	assert.True(t, ui.FailedWithUsage)
}

func callSetEnv(args []string, reqFactory *testreq.FakeReqFactory, appRepo api.ApplicationRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("set-env", args)

	cmd := NewSetEnv(ui, appRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
