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

func TestStopCommandFailsWithUsage(t *testing.T) {
	app := cf.Application{}
	app.Name = "my-app"
	app.Guid = "my-app-guid"
	appRepo := &testapi.FakeApplicationRepository{FindByNameApp: app}
	reqFactory := &testreq.FakeReqFactory{Application: app}

	ui := callStop(t, []string{}, reqFactory, appRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callStop(t, []string{"my-app"}, reqFactory, appRepo)
	assert.False(t, ui.FailedWithUsage)
}

func TestStopApplication(t *testing.T) {
	app := cf.Application{}
	app.Name = "my-app"
	app.Guid = "my-app-guid"
	appRepo := &testapi.FakeApplicationRepository{FindByNameApp: app}
	args := []string{"my-app"}
	reqFactory := &testreq.FakeReqFactory{Application: app}
	ui := callStop(t, args, reqFactory, appRepo)

	assert.Contains(t, ui.Outputs[0], "Stopping app")
	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[0], "my-org")
	assert.Contains(t, ui.Outputs[0], "my-space")
	assert.Contains(t, ui.Outputs[0], "my-user")
	assert.Contains(t, ui.Outputs[1], "OK")

	assert.Equal(t, reqFactory.ApplicationName, "my-app")
	assert.Equal(t, appRepo.StopAppGuid, "my-app-guid")
}

func TestStopApplicationWhenStopFails(t *testing.T) {
	app := cf.Application{}
	app.Name = "my-app"
	app.Guid = "my-app-guid"
	appRepo := &testapi.FakeApplicationRepository{FindByNameApp: app, StopAppErr: true}
	args := []string{"my-app"}
	reqFactory := &testreq.FakeReqFactory{Application: app}
	ui := callStop(t, args, reqFactory, appRepo)

	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[1], "FAILED")
	assert.Contains(t, ui.Outputs[2], "Error stopping application")
	assert.Equal(t, appRepo.StopAppGuid, "my-app-guid")
}

func TestStopApplicationIsAlreadyStopped(t *testing.T) {
	app := cf.Application{}
	app.Name = "my-app"
	app.Guid = "my-app-guid"
	app.State = "stopped"
	appRepo := &testapi.FakeApplicationRepository{FindByNameApp: app}
	args := []string{"my-app"}
	reqFactory := &testreq.FakeReqFactory{Application: app}
	ui := callStop(t, args, reqFactory, appRepo)

	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[0], "is already stopped")
	assert.Equal(t, appRepo.StopAppGuid, "")
}

func TestApplicationStopReturnsUpdatedApp(t *testing.T) {
	appToStop := cf.Application{}
	appToStop.Name = "my-app"
	appToStop.Guid = "my-app-guid"
	appToStop.State = "started"
	expectedStoppedApp := cf.Application{}
	expectedStoppedApp.Name = "my-stopped-app"
	expectedStoppedApp.Guid = "my-stopped-app-guid"
	expectedStoppedApp.State = "stopped"

	appRepo := &testapi.FakeApplicationRepository{StopUpdatedApp: expectedStoppedApp}
	config := &configuration.Configuration{}
	stopper := NewStop(new(testterm.FakeUI), config, appRepo)
	actualStoppedApp, err := stopper.ApplicationStop(appToStop)

	assert.NoError(t, err)
	assert.Equal(t, expectedStoppedApp, actualStoppedApp)
}

func TestApplicationStopReturnsUpdatedAppWhenAppIsAlreadyStopped(t *testing.T) {
	appToStop := cf.Application{}
	appToStop.Name = "my-app"
	appToStop.Guid = "my-app-guid"
	appToStop.State = "stopped"
	appRepo := &testapi.FakeApplicationRepository{}
	config := &configuration.Configuration{}
	stopper := NewStop(new(testterm.FakeUI), config, appRepo)
	updatedApp, err := stopper.ApplicationStop(appToStop)

	assert.NoError(t, err)
	assert.Equal(t, appToStop, updatedApp)
}

func callStop(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory, appRepo api.ApplicationRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("stop", args)

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

	cmd := NewStop(ui, config, appRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
