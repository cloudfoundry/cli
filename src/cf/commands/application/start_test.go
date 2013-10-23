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

var defaultAppForStart = cf.Application{
	Name:      "my-app",
	Guid:      "my-app-guid",
	Instances: 2,
	Urls:      []string{"http://my-app.example.com"},
}

func startAppWithInstancesAndErrors(t *testing.T, app cf.Application, instances [][]cf.ApplicationInstance, errorCodes []string) (ui *testterm.FakeUI, appRepo *testapi.FakeApplicationRepository, reqFactory *testreq.FakeReqFactory) {
	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)

	config := &configuration.Configuration{
		Space:                   cf.Space{Name: "my-space"},
		Organization:            cf.Organization{Name: "my-org"},
		AccessToken:             token,
		ApplicationStartTimeout: 2,
	}

	appRepo = &testapi.FakeApplicationRepository{
		FindByNameApp:          app,
		GetInstancesResponses:  instances,
		GetInstancesErrorCodes: errorCodes,
	}
	args := []string{"my-app"}
	reqFactory = &testreq.FakeReqFactory{Application: app}
	ui = callStart(args, config, reqFactory, appRepo)
	return
}

func TestStartCommandFailsWithUsage(t *testing.T) {
	t.Parallel()

	config := &configuration.Configuration{}
	appRepo := &testapi.FakeApplicationRepository{
		GetInstancesResponses: [][]cf.ApplicationInstance{
			[]cf.ApplicationInstance{},
		},
		GetInstancesErrorCodes: []string{""},
	}
	reqFactory := &testreq.FakeReqFactory{}

	ui := callStart([]string{}, config, reqFactory, appRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callStart([]string{"my-app"}, config, reqFactory, appRepo)
	assert.False(t, ui.FailedWithUsage)
}

func TestStartApplication(t *testing.T) {
	t.Parallel()

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

	errorCodes := []string{"", ""}
	ui, appRepo, reqFactory := startAppWithInstancesAndErrors(t, defaultAppForStart, instances, errorCodes)

	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[0], "my-org")
	assert.Contains(t, ui.Outputs[0], "my-space")
	assert.Contains(t, ui.Outputs[0], "my-user")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[3], "0 of 2 instances running (2 starting)")
	assert.Contains(t, ui.Outputs[4], "Started: app my-app available at http://my-app.example.com")

	assert.Equal(t, reqFactory.ApplicationName, "my-app")
	assert.Equal(t, appRepo.StartAppToStart.Guid, "my-app-guid")
}

func TestStartApplicationWhenAppHasNoURL(t *testing.T) {
	t.Parallel()

	app := defaultAppForStart
	app.Urls = []string{}

	instances := [][]cf.ApplicationInstance{
		[]cf.ApplicationInstance{
			cf.ApplicationInstance{State: cf.InstanceRunning},
		},
	}

	errorCodes := []string{""}
	ui, appRepo, reqFactory := startAppWithInstancesAndErrors(t, app, instances, errorCodes)

	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[3], "Started")

	assert.Equal(t, reqFactory.ApplicationName, "my-app")
	assert.Equal(t, appRepo.StartAppToStart.Guid, "my-app-guid")
}

func TestStartApplicationWhenAppIsStillStaging(t *testing.T) {
	t.Parallel()

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

	errorCodes := []string{cf.APP_NOT_STAGED, cf.APP_NOT_STAGED, "", "", ""}

	ui, _, _ := startAppWithInstancesAndErrors(t, defaultAppForStart, instances, errorCodes)

	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[3], "0 of 2 instances running (1 starting, 1 down)")
	assert.Contains(t, ui.Outputs[4], "0 of 2 instances running (2 starting)")
	assert.Contains(t, ui.Outputs[5], "Started: app my-app available at http://my-app.example.com")
}

func TestStartApplicationWhenStagingFails(t *testing.T) {
	t.Parallel()

	instances := [][]cf.ApplicationInstance{[]cf.ApplicationInstance{}}
	errorCodes := []string{"170001"}

	ui, _, _ := startAppWithInstancesAndErrors(t, defaultAppForStart, instances, errorCodes)

	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[3], "FAILED")
	assert.Contains(t, ui.Outputs[4], "Error staging app")
}

func TestStartApplicationWhenOneInstanceFlaps(t *testing.T) {
	t.Parallel()

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

	errorCodes := []string{"", ""}

	ui, _, _ := startAppWithInstancesAndErrors(t, defaultAppForStart, instances, errorCodes)

	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[3], "0 of 2 instances running (2 starting)")
	assert.Contains(t, ui.Outputs[4], "FAILED")
	assert.Contains(t, ui.Outputs[5], "Start unsuccessful")
}

func TestStartApplicationWhenStartTimesOut(t *testing.T) {
	t.Parallel()

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

	errorCodes := []string{"", "", ""}

	ui, _, _ := startAppWithInstancesAndErrors(t, defaultAppForStart, instances, errorCodes)

	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[3], "0 of 2 instances running (2 starting)")
	assert.Contains(t, ui.Outputs[4], "0 of 2 instances running (1 starting, 1 down)")
	assert.Contains(t, ui.Outputs[5], "0 of 2 instances running (2 down)")
	assert.Contains(t, ui.Outputs[6], "FAILED")
	assert.Contains(t, ui.Outputs[7], "Start app timeout")
}

func TestStartApplicationWhenStartFails(t *testing.T) {
	t.Parallel()

	config := &configuration.Configuration{}
	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}
	appRepo := &testapi.FakeApplicationRepository{FindByNameApp: app, StartAppErr: true}
	args := []string{"my-app"}
	reqFactory := &testreq.FakeReqFactory{Application: app}
	ui := callStart(args, config, reqFactory, appRepo)

	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[1], "FAILED")
	assert.Contains(t, ui.Outputs[2], "Error starting application")
	assert.Equal(t, appRepo.StartAppToStart.Guid, "my-app-guid")
}

func TestStartApplicationIsAlreadyStarted(t *testing.T) {
	t.Parallel()

	config := &configuration.Configuration{}
	app := cf.Application{Name: "my-app", Guid: "my-app-guid", State: "started"}
	appRepo := &testapi.FakeApplicationRepository{FindByNameApp: app}

	reqFactory := &testreq.FakeReqFactory{Application: app}

	args := []string{"my-app"}
	ui := callStart(args, config, reqFactory, appRepo)

	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[0], "is already started")
	assert.Equal(t, appRepo.StartAppToStart.Guid, "")
}

func callStart(args []string, config *configuration.Configuration, reqFactory *testreq.FakeReqFactory, appRepo api.ApplicationRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("start", args)

	cmd := NewStart(ui, config, appRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
