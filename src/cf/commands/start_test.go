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

func startAppWithInstancesAndErrors(instances [][]cf.ApplicationInstance, errors []bool) (ui *testhelpers.FakeUI, appRepo *testhelpers.FakeApplicationRepository) {
	config := &configuration.Configuration{}
	app := cf.Application{
		Name:      "my-app",
		Guid:      "my-app-guid",
		Instances: 2,
	}

	appRepo = &testhelpers.FakeApplicationRepository{
		AppByName:             app,
		GetInstancesResponses: instances,
		GetInstancesErrors:    errors,
	}
	args := []string{"my-app"}

	ui = callStart(args, config, appRepo)
	return
}

func TestStartApplication(t *testing.T) {
	instances := [][]cf.ApplicationInstance{
		[]cf.ApplicationInstance{
			cf.ApplicationInstance{State: "running"},
			cf.ApplicationInstance{State: "starting"},
		},
		[]cf.ApplicationInstance{
			cf.ApplicationInstance{State: "running"},
			cf.ApplicationInstance{State: "running"},
		},
	}

	errors := []bool{false, false}
	ui, appRepo := startAppWithInstancesAndErrors(instances, errors)

	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[3], "1 of 2 instances running (1 running, 1 starting)")
	assert.Contains(t, ui.Outputs[4], "2 of 2 instances running")

	assert.Equal(t, appRepo.AppName, "my-app")
	assert.Equal(t, appRepo.StartedApp.Guid, "my-app-guid")
}

func TestStartApplicationWhenAppIsStillStaging(t *testing.T) {
	instances := [][]cf.ApplicationInstance{
		[]cf.ApplicationInstance{},
		[]cf.ApplicationInstance{},
		[]cf.ApplicationInstance{
			cf.ApplicationInstance{State: "down"},
			cf.ApplicationInstance{State: "starting"},
		},
		[]cf.ApplicationInstance{
			cf.ApplicationInstance{State: "starting"},
			cf.ApplicationInstance{State: "running"},
		},
		[]cf.ApplicationInstance{
			cf.ApplicationInstance{State: "running"},
			cf.ApplicationInstance{State: "running"},
		},
	}

	errors := []bool{true, true, false, false, false}

	ui, _ := startAppWithInstancesAndErrors(instances, errors)

	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[3], "0 of 2 instances running (1 starting, 1 down)")
	assert.Contains(t, ui.Outputs[4], "1 of 2 instances running (1 running, 1 starting)")
	assert.Contains(t, ui.Outputs[5], "2 of 2 instances running")
}

func TestStartApplicationWhenOneInstanceFlaps ( t *testing.T) {
	instances := [][]cf.ApplicationInstance{
		[]cf.ApplicationInstance{
			cf.ApplicationInstance{State: "starting"},
			cf.ApplicationInstance{State: "starting"},
		},
		[]cf.ApplicationInstance{
			cf.ApplicationInstance{State: "starting"},
			cf.ApplicationInstance{State: "flapping"},
		},
	}

	errors := []bool{ false, false}

	ui, _ := startAppWithInstancesAndErrors(instances, errors)

	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[3], "0 of 2 instances running (2 starting)")
	assert.Contains(t, ui.Outputs[4], "FAILED")
	assert.Contains(t, ui.Outputs[5], "Start unsuccessful")
}

func TestStartApplicationWhenAppDoesNotExist(t *testing.T) {
	config := &configuration.Configuration{}
	appRepo := &testhelpers.FakeApplicationRepository{AppByNameErr: true}
	args := []string{"i-do-not-exist"}
	ui := callStart(args, config, appRepo)

	assert.Contains(t, ui.Outputs[0], "FAILED")
	assert.Contains(t, ui.Outputs[1], "i-do-not-exist")
}

func TestStartApplicationWhenStartFails(t *testing.T) {
	config := &configuration.Configuration{}
	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}
	appRepo := &testhelpers.FakeApplicationRepository{AppByName: app, StartAppErr: true}
	args := []string{"my-app"}
	ui := callStart(args, config, appRepo)

	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[1], "FAILED")
	assert.Contains(t, ui.Outputs[2], "Error starting application")
	assert.Equal(t, appRepo.StartedApp.Guid, "my-app-guid")
}

func TestStartApplicationIsAlreadyStarted(t *testing.T) {
	config := &configuration.Configuration{}
	app := cf.Application{Name: "my-app", Guid: "my-app-guid", State: "started"}
	appRepo := &testhelpers.FakeApplicationRepository{AppByName: app}
	args := []string{"my-app"}
	ui := callStart(args, config, appRepo)

	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[0], "is already started")
	assert.Equal(t, appRepo.StartedApp.Guid, "")
}

func callStart(args []string, config *configuration.Configuration, appRepo api.ApplicationRepository) (ui *testhelpers.FakeUI) {
	context := testhelpers.NewContext("start", args)
	ui = new(testhelpers.FakeUI)
	s := NewStart(ui, config, appRepo)
	s.Run(context)

	return
}
