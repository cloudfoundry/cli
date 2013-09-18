package commands_test

import (
	"cf"
	"cf/api"
	. "cf/commands"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestSetEnvRequirements(t *testing.T) {
	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}
	appRepo := &testhelpers.FakeApplicationRepository{}
	args := []string{"my-app", "DATABASE_URL", "mysql://example.com/my-db"}

	reqFactory := &testhelpers.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: true}
	callSetEnv(args, reqFactory, appRepo)
	assert.True(t, testhelpers.CommandDidPassRequirements)

	reqFactory = &testhelpers.FakeReqFactory{Application: app, LoginSuccess: false, TargetedSpaceSuccess: true}
	callSetEnv(args, reqFactory, appRepo)
	assert.False(t, testhelpers.CommandDidPassRequirements)

	testhelpers.CommandDidPassRequirements = true

	reqFactory = &testhelpers.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: false}
	callSetEnv(args, reqFactory, appRepo)
	assert.False(t, testhelpers.CommandDidPassRequirements)
}

func TestRunWhenApplicationExists(t *testing.T) {
	app := cf.Application{Name: "my-app", Guid: "my-app-guid", EnvironmentVars: map[string]string{"foo": "bar"}}
	reqFactory := &testhelpers.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: true}
	appRepo := &testhelpers.FakeApplicationRepository{}

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

func TestRunWhenSettingTheEnvFails(t *testing.T) {
	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}
	reqFactory := &testhelpers.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: true}
	appRepo := &testhelpers.FakeApplicationRepository{
		AppByName: app,
		SetEnvErr: true,
	}

	args := []string{"does-not-exist", "DATABASE_URL", "mysql://example.com/my-db"}
	ui := callSetEnv(args, reqFactory, appRepo)

	assert.Contains(t, ui.Outputs[0], "Updating env variable")
	assert.Contains(t, ui.Outputs[1], "FAILED")
	assert.Contains(t, ui.Outputs[2], "Failed setting env")
}

func TestSetEnvFailsWithUsage(t *testing.T) {
	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}
	reqFactory := &testhelpers.FakeReqFactory{Application: app, LoginSuccess: true, TargetedSpaceSuccess: true}
	appRepo := &testhelpers.FakeApplicationRepository{AppByName: app}

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

func callSetEnv(args []string, reqFactory *testhelpers.FakeReqFactory, appRepo api.ApplicationRepository) (ui *testhelpers.FakeUI) {
	ui = new(testhelpers.FakeUI)
	ctxt := testhelpers.NewContext("set-env", args)

	cmd := NewSetEnv(ui, appRepo)
	testhelpers.RunCommand(cmd, ctxt, reqFactory)
	return
}
