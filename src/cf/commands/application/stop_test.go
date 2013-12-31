package application_test

import (
	"cf"
	"cf/api"
	. "cf/commands/application"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
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
	appRepo := &testapi.FakeApplicationRepository{ReadApp: app}
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
	appRepo := &testapi.FakeApplicationRepository{ReadApp: app}
	args := []string{"my-app"}
	reqFactory := &testreq.FakeReqFactory{Application: app}
	ui := callStop(t, args, reqFactory, appRepo)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Stopping app", "my-app", "my-org", "my-space", "my-user"},
		{"OK"},
	})

	assert.Equal(t, reqFactory.ApplicationName, "my-app")
	assert.Equal(t, appRepo.UpdateAppGuid, "my-app-guid")
}

func TestStopApplicationWhenStopFails(t *testing.T) {
	app := cf.Application{}
	app.Name = "my-app"
	app.Guid = "my-app-guid"
	appRepo := &testapi.FakeApplicationRepository{ReadApp: app, UpdateErr: true}
	args := []string{"my-app"}
	reqFactory := &testreq.FakeReqFactory{Application: app}
	ui := callStop(t, args, reqFactory, appRepo)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Stopping", "my-app"},
		{"FAILED"},
		{"Error updating app."},
	})
	assert.Equal(t, appRepo.UpdateAppGuid, "my-app-guid")
}

func TestStopApplicationIsAlreadyStopped(t *testing.T) {
	app := cf.Application{}
	app.Name = "my-app"
	app.Guid = "my-app-guid"
	app.State = "stopped"
	appRepo := &testapi.FakeApplicationRepository{ReadApp: app}
	args := []string{"my-app"}
	reqFactory := &testreq.FakeReqFactory{Application: app}
	ui := callStop(t, args, reqFactory, appRepo)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"my-app", "is already stopped"},
	})
	assert.Equal(t, appRepo.UpdateAppGuid, "")
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

	appRepo := &testapi.FakeApplicationRepository{UpdateAppResult: expectedStoppedApp}
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
