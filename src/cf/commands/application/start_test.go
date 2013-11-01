package application_test

import (
	"cf"
	"cf/api"
	. "cf/commands/application"
	"cf/configuration"
	"code.google.com/p/gogoprotobuf/proto"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
	"time"
)

var defaultAppForStart = cf.Application{
	Name:      "my-app",
	Guid:      "my-app-guid",
	Instances: 2,
	Routes: []cf.Route{
		{Host: "my-app", Domain: cf.Domain{Name: "example.com"}},
	},
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

	currentTime := time.Now()
	messageType := logmessage.LogMessage_ERR
	sourceType := logmessage.LogMessage_DEA
	logMessage1 := logmessage.LogMessage{
		Message:     []byte("Log Line 1"),
		AppId:       proto.String(app.Guid),
		MessageType: &messageType,
		SourceType:  &sourceType,
		Timestamp:   proto.Int64(currentTime.UnixNano()),
	}

	logMessage2 := logmessage.LogMessage{
		Message:     []byte("Log Line 2"),
		AppId:       proto.String(app.Guid),
		MessageType: &messageType,
		SourceType:  &sourceType,
		Timestamp:   proto.Int64(currentTime.UnixNano()),
	}

	logRepo := &testapi.FakeLogsRepository{
		TailLogMessages: []logmessage.LogMessage{
			logMessage1,
			logMessage2,
		},
	}

	args := []string{"my-app"}
	reqFactory = &testreq.FakeReqFactory{Application: app}
	ui = callStart(args, config, reqFactory, appRepo, logRepo)
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
	logRepo := &testapi.FakeLogsRepository{}

	reqFactory := &testreq.FakeReqFactory{}

	ui := callStart([]string{}, config, reqFactory, appRepo, logRepo) //
	assert.True(t, ui.FailedWithUsage)

	ui = callStart([]string{"my-app"}, config, reqFactory, appRepo, logRepo)
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
	assert.Contains(t, ui.Outputs[6], "0 of 2 instances running (2 starting)")
	assert.Contains(t, ui.Outputs[7], "Started")
	assert.Contains(t, ui.Outputs[7], "my-app")
	assert.Contains(t, ui.Outputs[7], "my-app.example.com")

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
	assert.Contains(t, ui.Outputs[6], "Started")

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

	assert.Contains(t, ui.Outputs[2], "Staging")
	assert.Contains(t, ui.Outputs[3], "Log Line 1")
	assert.Contains(t, ui.Outputs[4], "Log Line 2")

	assert.Contains(t, ui.Outputs[6], "0 of 2 instances running (1 starting, 1 down)")
	assert.Contains(t, ui.Outputs[7], "0 of 2 instances running (2 starting)")
}

func TestStartApplicationWhenStagingFails(t *testing.T) {
	t.Parallel()

	instances := [][]cf.ApplicationInstance{[]cf.ApplicationInstance{}}
	errorCodes := []string{"170001"}

	ui, _, _ := startAppWithInstancesAndErrors(t, defaultAppForStart, instances, errorCodes)

	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[6], "FAILED")
	assert.Contains(t, ui.Outputs[7], "Error staging app")
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
	assert.Contains(t, ui.Outputs[6], "0 of 2 instances running (2 starting)")
	assert.Contains(t, ui.Outputs[7], "FAILED")
	assert.Contains(t, ui.Outputs[8], "Start unsuccessful")
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
	assert.Contains(t, ui.Outputs[6], "0 of 2 instances running (2 starting)")
	assert.Contains(t, ui.Outputs[7], "0 of 2 instances running (1 starting, 1 down)")
	assert.Contains(t, ui.Outputs[8], "0 of 2 instances running (2 down)")
	assert.Contains(t, ui.Outputs[9], "FAILED")
	assert.Contains(t, ui.Outputs[10], "Start app timeout")
}

func TestStartApplicationWhenStartFails(t *testing.T) {
	t.Parallel()

	config := &configuration.Configuration{}
	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}
	appRepo := &testapi.FakeApplicationRepository{FindByNameApp: app, StartAppErr: true}
	logRepo := &testapi.FakeLogsRepository{}
	args := []string{"my-app"}
	reqFactory := &testreq.FakeReqFactory{Application: app}
	ui := callStart(args, config, reqFactory, appRepo, logRepo)

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
	logRepo := &testapi.FakeLogsRepository{}

	reqFactory := &testreq.FakeReqFactory{Application: app}

	args := []string{"my-app"}
	ui := callStart(args, config, reqFactory, appRepo, logRepo)

	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[0], "is already started")
	assert.Equal(t, appRepo.StartAppToStart.Guid, "")
}

func callStart(args []string, config *configuration.Configuration, reqFactory *testreq.FakeReqFactory, appRepo api.ApplicationRepository, logRepo api.LogsRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("start", args)

	cmd := NewStart(ui, config, appRepo, logRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
