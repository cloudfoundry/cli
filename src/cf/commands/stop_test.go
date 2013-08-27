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

func TestStopApplication(t *testing.T) {
	config := &configuration.Configuration{}
	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}
	appRepo := &testhelpers.FakeApplicationRepository{AppByName: app}
	args := []string{"my-app"}
	ui := callStop(args, config, appRepo)

	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[1], "OK")

	assert.Equal(t, appRepo.AppName, "my-app")
	assert.Equal(t, appRepo.StoppedApp.Guid, "my-app-guid")
}

func TestStopApplicationWhenAppDoesNotExist(t *testing.T) {
	config := &configuration.Configuration{}
	appRepo := &testhelpers.FakeApplicationRepository{AppByNameErr: true}
	args := []string{"i-do-not-exist"}
	ui := callStop(args, config, appRepo)

	assert.Contains(t, ui.Outputs[0], "FAILED")
	assert.Contains(t, ui.Outputs[1], "i-do-not-exist")
}

func TestStopApplicationWhenStopFails(t *testing.T) {
	config := &configuration.Configuration{}
	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}
	appRepo := &testhelpers.FakeApplicationRepository{AppByName: app, StopAppErr: true}
	args := []string{"my-app"}
	ui := callStop(args, config, appRepo)

	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[1], "FAILED")
	assert.Contains(t, ui.Outputs[2], "Error stopping application")
	assert.Equal(t, appRepo.StoppedApp.Guid, "my-app-guid")
}

func TestStopApplicationIsAlreadyStopped(t *testing.T) {
	config := &configuration.Configuration{}
	app := cf.Application{Name: "my-app", Guid: "my-app-guid", State: "stopped"}
	appRepo := &testhelpers.FakeApplicationRepository{AppByName: app}
	args := []string{"my-app"}
	ui := callStop(args, config, appRepo)

	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[0], "is already stopped")
	assert.Equal(t, appRepo.StoppedApp.Guid, "")
}

func callStop(args []string, config *configuration.Configuration, appRepo api.ApplicationRepository) (ui *testhelpers.FakeUI) {
	context := testhelpers.NewContext("stop", args)
	ui = new(testhelpers.FakeUI)
	s := NewStop(ui, config, appRepo)
	s.Run(context)

	return
}
