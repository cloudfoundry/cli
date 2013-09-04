package commands_test

import (
	"cf"
	"cf/api"
	. "cf/commands"
	"cf/configuration"
	"cf/requirements"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestStopApplication(t *testing.T) {
	config := &configuration.Configuration{}
	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}
	appRepo := &testhelpers.FakeApplicationRepository{AppByName: app}
	args := []string{"my-app"}
	reqFactory := &testhelpers.FakeReqFactory{Application: app}
	ui := callStop(args, config, reqFactory, appRepo)

	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[1], "OK")

	assert.Equal(t, reqFactory.ApplicationName, "my-app")
	assert.Equal(t, appRepo.StoppedApp.Guid, "my-app-guid")
}

func TestStopApplicationWhenStopFails(t *testing.T) {
	config := &configuration.Configuration{}
	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}
	appRepo := &testhelpers.FakeApplicationRepository{AppByName: app, StopAppErr: true}
	args := []string{"my-app"}
	reqFactory := &testhelpers.FakeReqFactory{Application: app}
	ui := callStop(args, config, reqFactory, appRepo)

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
	reqFactory := &testhelpers.FakeReqFactory{Application: app}
	ui := callStop(args, config, reqFactory, appRepo)

	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[0], "is already stopped")
	assert.Equal(t, appRepo.StoppedApp.Guid, "")
}

func callStop(args []string, config *configuration.Configuration, reqFactory requirements.Factory, appRepo api.ApplicationRepository) (ui *testhelpers.FakeUI) {
	ui = new(testhelpers.FakeUI)
	ctxt := testhelpers.NewContext("stop", args)

	cmd := NewStop(ui, config, appRepo)
	cmd.GetRequirements(reqFactory, ctxt)
	cmd.Run(ctxt)

	return
}
