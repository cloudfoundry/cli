package application_test

import (
	"cf"
	"cf/api"
	. "cf/commands/application"
	"cf/configuration"
	"code.google.com/p/gogoprotobuf/proto"
	"errors"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	"github.com/stretchr/testify/assert"
	"os"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
	"time"
)

var (
	defaultAppForStart        = cf.Application{}
	defaultInstanceReponses   = [][]cf.AppInstanceFields{}
	defaultInstanceErrorCodes = []string{"", ""}
)

func init() {
	defaultAppForStart.Name = "my-app"
	defaultAppForStart.Guid = "my-app-guid"
	defaultAppForStart.InstanceCount = 2

	domain := cf.DomainFields{}
	domain.Name = "example.com"

	route := cf.RouteSummary{}
	route.Host = "my-app"
	route.Domain = domain

	defaultAppForStart.Routes = []cf.RouteSummary{route}

	instance1 := cf.AppInstanceFields{}
	instance1.State = cf.InstanceStarting

	instance2 := cf.AppInstanceFields{}
	instance2.State = cf.InstanceStarting

	instance3 := cf.AppInstanceFields{}
	instance3.State = cf.InstanceRunning

	instance4 := cf.AppInstanceFields{}
	instance4.State = cf.InstanceStarting

	defaultInstanceReponses = [][]cf.AppInstanceFields{
		[]cf.AppInstanceFields{instance1, instance2},
		[]cf.AppInstanceFields{instance3, instance4},
	}
}

func callStart(args []string, config *configuration.Configuration, reqFactory *testreq.FakeReqFactory, appRepo api.ApplicationRepository, appInstancesRepo api.AppInstancesRepository, logRepo api.LogsRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("start", args)

	cmd := NewStart(ui, config, appRepo, appInstancesRepo, logRepo)
	cmd.StagingTimeout = 2 * time.Second
	cmd.StartupTimeout = 2 * time.Second
	cmd.PingerThrottle = 1 * time.Second

	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}

func startAppWithInstancesAndErrors(t *testing.T, app cf.Application, instances [][]cf.AppInstanceFields, errorCodes []string) (ui *testterm.FakeUI, appRepo *testapi.FakeApplicationRepository, appInstancesRepo *testapi.FakeAppInstancesRepo, reqFactory *testreq.FakeReqFactory) {
	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)
	space := cf.SpaceFields{}
	space.Name = "my-space"
	org := cf.OrganizationFields{}
	org.Name = "my-org"
	config := &configuration.Configuration{
		SpaceFields:             space,
		OrganizationFields:      org,
		AccessToken:             token,
		ApplicationStartTimeout: 2,
	}

	appRepo = &testapi.FakeApplicationRepository{
		FindByNameApp:   app,
		StartUpdatedApp: app,
	}
	appInstancesRepo = &testapi.FakeAppInstancesRepo{
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
	ui = callStart(args, config, reqFactory, appRepo, appInstancesRepo, logRepo)
	return
}

func TestStartCommandDefaultTimeouts(t *testing.T) {
	cmd := NewStart(new(testterm.FakeUI), &configuration.Configuration{}, &testapi.FakeApplicationRepository{}, &testapi.FakeAppInstancesRepo{}, &testapi.FakeLogsRepository{})
	assert.Equal(t, cmd.StagingTimeout, 20*time.Minute)
	assert.Equal(t, cmd.StartupTimeout, 5*time.Minute)
}

func TestStartCommandSetsTimeoutsFromEnv(t *testing.T) {
	oldStaging := os.Getenv("CF_STAGING_TIMEOUT")
	oldStart := os.Getenv("CF_STARTUP_TIMEOUT")
	defer func() {
		os.Setenv("CF_STAGING_TIMEOUT", oldStaging)
		os.Setenv("CF_STARTUP_TIMEOUT", oldStart)
	}()

	os.Setenv("CF_STAGING_TIMEOUT", "6")
	os.Setenv("CF_STARTUP_TIMEOUT", "3")
	cmd := NewStart(new(testterm.FakeUI), &configuration.Configuration{}, &testapi.FakeApplicationRepository{}, &testapi.FakeAppInstancesRepo{}, &testapi.FakeLogsRepository{})
	assert.Equal(t, cmd.StagingTimeout, 6*time.Minute)
	assert.Equal(t, cmd.StartupTimeout, 3*time.Minute)
}

func TestStartCommandFailsWithUsage(t *testing.T) {
	t.Parallel()

	config := &configuration.Configuration{}
	appRepo := &testapi.FakeApplicationRepository{}
	appInstancesRepo := &testapi.FakeAppInstancesRepo{
		GetInstancesResponses: [][]cf.AppInstanceFields{
			[]cf.AppInstanceFields{},
		},
		GetInstancesErrorCodes: []string{""},
	}
	logRepo := &testapi.FakeLogsRepository{}

	reqFactory := &testreq.FakeReqFactory{}

	ui := callStart([]string{}, config, reqFactory, appRepo, appInstancesRepo, logRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callStart([]string{"my-app"}, config, reqFactory, appRepo, appInstancesRepo, logRepo)
	assert.False(t, ui.FailedWithUsage)
}

func TestStartApplication(t *testing.T) {
	t.Parallel()

	ui, appRepo, _, reqFactory := startAppWithInstancesAndErrors(t, defaultAppForStart, defaultInstanceReponses, defaultInstanceErrorCodes)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"my-app", "my-org", "my-space", "my-user"},
		{"OK"},
		{"0 of 2 instances running (2 starting)"},
		{"Started", "my-app", "my-app.example.com"},
	})

	assert.Equal(t, reqFactory.ApplicationName, "my-app")
	assert.Equal(t, appRepo.StartAppGuid, "my-app-guid")
}

func TestStartApplicationWhenAppHasNoURL(t *testing.T) {
	t.Parallel()

	app := defaultAppForStart
	app.Routes = []cf.RouteSummary{}
	appInstance5 := cf.AppInstanceFields{}
	appInstance5.State = cf.InstanceRunning
	instances := [][]cf.AppInstanceFields{
		[]cf.AppInstanceFields{appInstance5},
	}

	errorCodes := []string{""}
	ui, appRepo, _, reqFactory := startAppWithInstancesAndErrors(t, app, instances, errorCodes)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"my-app"},
		{"OK"},
		{"Started"},
	})

	assert.Equal(t, reqFactory.ApplicationName, "my-app")
	assert.Equal(t, appRepo.StartAppGuid, "my-app-guid")
}

func TestStartApplicationWhenAppIsStillStaging(t *testing.T) {
	t.Parallel()
	appInstance6 := cf.AppInstanceFields{}
	appInstance6.State = cf.InstanceDown
	appInstance7 := cf.AppInstanceFields{}
	appInstance7.State = cf.InstanceStarting
	appInstance8 := cf.AppInstanceFields{}
	appInstance8.State = cf.InstanceStarting
	appInstance9 := cf.AppInstanceFields{}
	appInstance9.State = cf.InstanceStarting
	appInstance10 := cf.AppInstanceFields{}
	appInstance10.State = cf.InstanceRunning
	appInstance11 := cf.AppInstanceFields{}
	appInstance11.State = cf.InstanceRunning
	instances := [][]cf.AppInstanceFields{
		[]cf.AppInstanceFields{},
		[]cf.AppInstanceFields{},
		[]cf.AppInstanceFields{appInstance6, appInstance7},
		[]cf.AppInstanceFields{appInstance8, appInstance9},
		[]cf.AppInstanceFields{appInstance10, appInstance11},
	}

	errorCodes := []string{cf.APP_NOT_STAGED, cf.APP_NOT_STAGED, "", "", ""}

	ui, _, appInstancesRepo, _ := startAppWithInstancesAndErrors(t, defaultAppForStart, instances, errorCodes)

	assert.Equal(t, appInstancesRepo.GetInstancesAppGuid, "my-app-guid")

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Log Line 1"},
		{"Log Line 2"},
		{"0 of 2 instances running (1 starting, 1 down)"},
		{"0 of 2 instances running (2 starting)"},
	})
}

func TestStartApplicationWhenStagingFails(t *testing.T) {
	t.Parallel()

	instances := [][]cf.AppInstanceFields{[]cf.AppInstanceFields{}}
	errorCodes := []string{"170001"}

	ui, _, _, _ := startAppWithInstancesAndErrors(t, defaultAppForStart, instances, errorCodes)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"my-app"},
		{"OK"},
		{"FAILED"},
		{"Error staging app"},
	})
}

func TestStartApplicationWhenOneInstanceFlaps(t *testing.T) {
	t.Parallel()
	appInstance12 := cf.AppInstanceFields{}
	appInstance12.State = cf.InstanceStarting
	appInstance13 := cf.AppInstanceFields{}
	appInstance13.State = cf.InstanceStarting
	appInstance14 := cf.AppInstanceFields{}
	appInstance14.State = cf.InstanceStarting
	appInstance15 := cf.AppInstanceFields{}
	appInstance15.State = cf.InstanceFlapping
	instances := [][]cf.AppInstanceFields{
		[]cf.AppInstanceFields{appInstance12, appInstance13},
		[]cf.AppInstanceFields{appInstance14, appInstance15},
	}

	errorCodes := []string{"", ""}

	ui, _, _, _ := startAppWithInstancesAndErrors(t, defaultAppForStart, instances, errorCodes)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"my-app"},
		{"OK"},
		{"0 of 2 instances running (2 starting)"},
		{"FAILED"},
		{"Start unsuccessful"},
	})
}

func TestStartApplicationWhenStartTimesOut(t *testing.T) {
	t.Parallel()
	appInstance16 := cf.AppInstanceFields{}
	appInstance16.State = cf.InstanceStarting
	appInstance17 := cf.AppInstanceFields{}
	appInstance17.State = cf.InstanceStarting
	appInstance18 := cf.AppInstanceFields{}
	appInstance18.State = cf.InstanceStarting
	appInstance19 := cf.AppInstanceFields{}
	appInstance19.State = cf.InstanceDown
	appInstance20 := cf.AppInstanceFields{}
	appInstance20.State = cf.InstanceDown
	appInstance21 := cf.AppInstanceFields{}
	appInstance21.State = cf.InstanceDown
	instances := [][]cf.AppInstanceFields{
		[]cf.AppInstanceFields{appInstance16, appInstance17},
		[]cf.AppInstanceFields{appInstance18, appInstance19},
		[]cf.AppInstanceFields{appInstance20, appInstance21},
	}

	errorCodes := []string{"", "", ""}

	ui, _, _, _ := startAppWithInstancesAndErrors(t, defaultAppForStart, instances, errorCodes)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"my-app"},
		{"OK"},
		{"0 of 2 instances running (2 starting)"},
		{"0 of 2 instances running (1 starting, 1 down)"},
		{"0 of 2 instances running (2 down)"},
		{"FAILED"},
		{"Start app timeout"},
	})
}

func TestStartApplicationWhenStartFails(t *testing.T) {
	t.Parallel()

	config := &configuration.Configuration{}
	app := cf.Application{}
	app.Name = "my-app"
	app.Guid = "my-app-guid"
	appRepo := &testapi.FakeApplicationRepository{FindByNameApp: app, StartAppErr: true}
	appInstancesRepo := &testapi.FakeAppInstancesRepo{}
	logRepo := &testapi.FakeLogsRepository{}
	args := []string{"my-app"}
	reqFactory := &testreq.FakeReqFactory{Application: app}
	ui := callStart(args, config, reqFactory, appRepo, appInstancesRepo, logRepo)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"my-app"},
		{"FAILED"},
		{"Error starting application"},
	})
	assert.Equal(t, appRepo.StartAppGuid, "my-app-guid")
}

func TestStartApplicationIsAlreadyStarted(t *testing.T) {
	t.Parallel()

	config := &configuration.Configuration{}
	app := cf.Application{}
	app.Name = "my-app"
	app.Guid = "my-app-guid"
	app.State = "started"
	appRepo := &testapi.FakeApplicationRepository{FindByNameApp: app}
	appInstancesRepo := &testapi.FakeAppInstancesRepo{}
	logRepo := &testapi.FakeLogsRepository{}

	reqFactory := &testreq.FakeReqFactory{Application: app}

	args := []string{"my-app"}
	ui := callStart(args, config, reqFactory, appRepo, appInstancesRepo, logRepo)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"my-app", "is already started"},
	})

	assert.Equal(t, appRepo.StartAppGuid, "")
}

func TestStartApplicationWithLoggingFailure(t *testing.T) {
	t.Parallel()

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{Username: "my-user"})
	assert.NoError(t, err)
	space2 := cf.SpaceFields{}
	space2.Name = "my-space"
	org2 := cf.OrganizationFields{}
	org2.Name = "my-org"
	config := &configuration.Configuration{
		SpaceFields:             space2,
		OrganizationFields:      org2,
		AccessToken:             token,
		ApplicationStartTimeout: 2,
	}

	appRepo := &testapi.FakeApplicationRepository{FindByNameApp: defaultAppForStart}
	appInstancesRepo := &testapi.FakeAppInstancesRepo{
		GetInstancesResponses:  defaultInstanceReponses,
		GetInstancesErrorCodes: defaultInstanceErrorCodes,
	}

	logRepo := &testapi.FakeLogsRepository{
		TailLogErr: errors.New("Ooops"),
	}

	reqFactory := &testreq.FakeReqFactory{Application: defaultAppForStart}

	ui := callStart([]string{"my-app"}, config, reqFactory, appRepo, appInstancesRepo, logRepo)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		testassert.Line{"error tailing logs"},
		testassert.Line{"Ooops"},
	})
}
