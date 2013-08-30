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

func startAppWithInstancesAndErrors(instances [][]cf.ApplicationInstance, errorCodes []int) (ui *testhelpers.FakeUI, appRepo *testhelpers.FakeApplicationRepository) {
	config := &configuration.Configuration{ApplicationStartTimeout: 2}
	app := cf.Application{
		Name:      "my-app",
		Guid:      "my-app-guid",
		Instances: 2,
		Urls:      []string{"http://my-app.example.com"},
	}

	appRepo = &testhelpers.FakeApplicationRepository{
		AppByName:              app,
		GetInstancesResponses:  instances,
		GetInstancesErrorCodes: errorCodes,
	}
	args := []string{"my-app"}

	ui = callStart(args, config, appRepo)
	return
}

func TestStartApplication(t *testing.T) {
	instances := [][]cf.ApplicationInstance{
		[]cf.ApplicationInstance{
			cf.ApplicationInstance{State: cf.InstanceStarting},
			cf.ApplicationInstance{State: cf.InstanceStarting},
		},
		[]cf.ApplicationInstance{
			cf.ApplicationInstance{State: cf.InstanceRunning},
			cf.ApplicationInstance{State: cf.InstanceStarting},
		},
	}

	errorCodes := []int{0, 0}
	ui, appRepo := startAppWithInstancesAndErrors(instances, errorCodes)

	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[3], "0 of 2 instances running (2 starting)")
	assert.Contains(t, ui.Outputs[4], "Start successful! App my-app available at http://my-app.example.com")

	assert.Equal(t, appRepo.AppName, "my-app")
	assert.Equal(t, appRepo.StartedApp.Guid, "my-app-guid")
}

func TestStartApplicationWhenAppIsStillStaging(t *testing.T) {
	instances := [][]cf.ApplicationInstance{
		[]cf.ApplicationInstance{},
		[]cf.ApplicationInstance{},
		[]cf.ApplicationInstance{
			cf.ApplicationInstance{State: cf.InstanceDown},
			cf.ApplicationInstance{State: cf.InstanceStarting},
		},
		[]cf.ApplicationInstance{
			cf.ApplicationInstance{State: cf.InstanceStarting},
			cf.ApplicationInstance{State: cf.InstanceStarting},
		},
		[]cf.ApplicationInstance{
			cf.ApplicationInstance{State: cf.InstanceRunning},
			cf.ApplicationInstance{State: cf.InstanceRunning},
		},
	}

	errorCodes := []int{170002, 170002, 0, 0, 0}

	ui, _ := startAppWithInstancesAndErrors(instances, errorCodes)

	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[3], "0 of 2 instances running (1 starting, 1 down)")
	assert.Contains(t, ui.Outputs[4], "0 of 2 instances running (2 starting)")
	assert.Contains(t, ui.Outputs[5], "Start successful! App my-app available at http://my-app.example.com")
}

func TestStartApplicationWhenStagingFails(t *testing.T) {
	instances := [][]cf.ApplicationInstance{[]cf.ApplicationInstance{}}
	errorCodes := []int{170001}

	ui, _ := startAppWithInstancesAndErrors(instances, errorCodes)

	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[3], "FAILED")
	assert.Contains(t, ui.Outputs[4], "Error staging app")
}

func TestStartApplicationWhenOneInstanceFlaps(t *testing.T) {
	instances := [][]cf.ApplicationInstance{
		[]cf.ApplicationInstance{
			cf.ApplicationInstance{State: cf.InstanceStarting},
			cf.ApplicationInstance{State: cf.InstanceStarting},
		},
		[]cf.ApplicationInstance{
			cf.ApplicationInstance{State: cf.InstanceStarting},
			cf.ApplicationInstance{State: cf.InstanceFlapping},
		},
	}

	errorCodes := []int{0, 0}

	ui, _ := startAppWithInstancesAndErrors(instances, errorCodes)

	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[3], "0 of 2 instances running (2 starting)")
	assert.Contains(t, ui.Outputs[4], "FAILED")
	assert.Contains(t, ui.Outputs[5], "Start unsuccessful")
}

func TestStartApplicationWhenStartTimesOut(t *testing.T) {
	instances := [][]cf.ApplicationInstance{
		[]cf.ApplicationInstance{
			cf.ApplicationInstance{State: cf.InstanceStarting},
			cf.ApplicationInstance{State: cf.InstanceStarting},
		},
		[]cf.ApplicationInstance{
			cf.ApplicationInstance{State: cf.InstanceStarting},
			cf.ApplicationInstance{State: cf.InstanceDown},
		},
		[]cf.ApplicationInstance{
			cf.ApplicationInstance{State: cf.InstanceDown},
			cf.ApplicationInstance{State: cf.InstanceDown},
		},
	}

	errorCodes := []int{0, 0, 0}

	ui, _ := startAppWithInstancesAndErrors(instances, errorCodes)

	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[3], "0 of 2 instances running (2 starting)")
	assert.Contains(t, ui.Outputs[4], "0 of 2 instances running (1 starting, 1 down)")
	assert.Contains(t, ui.Outputs[5], "0 of 2 instances running (2 down)")
	assert.Contains(t, ui.Outputs[6], "FAILED")
	assert.Contains(t, ui.Outputs[7], "Start app timeout")
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
