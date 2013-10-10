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

func TestStopCommandFailsWithUsage(t *testing.T) {
	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}
	appRepo := &testapi.FakeApplicationRepository{FindByNameApp: app}
	reqFactory := &testreq.FakeReqFactory{Application: app}

	ui := callStop([]string{}, reqFactory, appRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callStop([]string{"my-app"}, reqFactory, appRepo)
	assert.False(t, ui.FailedWithUsage)
}

func TestStopApplication(t *testing.T) {
	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}
	appRepo := &testapi.FakeApplicationRepository{FindByNameApp: app}
	args := []string{"my-app"}
	reqFactory := &testreq.FakeReqFactory{Application: app}
	ui := callStop(args, reqFactory, appRepo)

	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[1], "OK")

	assert.Equal(t, reqFactory.ApplicationName, "my-app")
	assert.Equal(t, appRepo.StopAppToStop.Guid, "my-app-guid")
}

func TestStopApplicationWhenStopFails(t *testing.T) {
	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}
	appRepo := &testapi.FakeApplicationRepository{FindByNameApp: app, StopAppErr: true}
	args := []string{"my-app"}
	reqFactory := &testreq.FakeReqFactory{Application: app}
	ui := callStop(args, reqFactory, appRepo)

	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[1], "FAILED")
	assert.Contains(t, ui.Outputs[2], "Error stopping application")
	assert.Equal(t, appRepo.StopAppToStop.Guid, "my-app-guid")
}

func TestStopApplicationIsAlreadyStopped(t *testing.T) {
	app := cf.Application{Name: "my-app", Guid: "my-app-guid", State: "stopped"}
	appRepo := &testapi.FakeApplicationRepository{FindByNameApp: app}
	args := []string{"my-app"}
	reqFactory := &testreq.FakeReqFactory{Application: app}
	ui := callStop(args, reqFactory, appRepo)

	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[0], "is already stopped")
	assert.Equal(t, appRepo.StopAppToStop.Guid, "")
}

func TestApplicationStopReturnsUpdatedApp(t *testing.T) {
	appToStop := cf.Application{Name: "my-app", Guid: "my-app-guid", State: "started"}
	expectedStoppedApp := cf.Application{Name: "my-stopped-app", Guid: "my-stopped-app-guid", State: "stopped"}

	appRepo := &testapi.FakeApplicationRepository{StopUpdatedApp: expectedStoppedApp}
	stopper := NewStop(new(testterm.FakeUI), appRepo)
	actualStoppedApp, err := stopper.ApplicationStop(appToStop)

	assert.NoError(t, err)
	assert.Equal(t, expectedStoppedApp, actualStoppedApp)
}

func TestApplicationStopReturnsUpdatedAppWhenAppIsAlreadyStopped(t *testing.T) {
	appToStop := cf.Application{Name: "my-app", Guid: "my-app-guid", State: "stopped"}
	appRepo := &testapi.FakeApplicationRepository{}
	stopper := NewStop(new(testterm.FakeUI), appRepo)
	updatedApp, err := stopper.ApplicationStop(appToStop)

	assert.NoError(t, err)
	assert.Equal(t, appToStop, updatedApp)
}

func callStop(args []string, reqFactory *testreq.FakeReqFactory, appRepo api.ApplicationRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("stop", args)

	cmd := NewStop(ui, appRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
