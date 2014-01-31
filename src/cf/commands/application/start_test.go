package application_test

import (
	"cf"
	"cf/api"
	. "cf/commands/application"
	"cf/configuration"
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

	. "github.com/onsi/ginkgo"
	mr "github.com/tjarratt/mr_t"
	"time"
)

var (
	defaultAppForStart        = cf.Application{}
	defaultInstanceReponses   = [][]cf.AppInstanceFields{}
	defaultInstanceErrorCodes = []string{"", ""}
	defaultStartTimeout       = 50 * time.Millisecond
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
		[]cf.AppInstanceFields{instance1, instance2},
		[]cf.AppInstanceFields{instance3, instance4},
	}
}

func callStart(args []string, config *configuration.Configuration, reqFactory *testreq.FakeReqFactory, displayApp ApplicationDisplayer, appRepo api.ApplicationRepository, appInstancesRepo api.AppInstancesRepository, logRepo api.LogsRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("start", args)

	cmd := NewStart(ui, config, displayApp, appRepo, appInstancesRepo, logRepo)
	cmd.StagingTimeout = 5 * time.Millisecond
	cmd.StartupTimeout = config.ApplicationStartTimeout
	cmd.PingerThrottle = 5 * time.Millisecond

	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}

func startAppWithInstancesAndErrors(t mr.TestingT, displayApp ApplicationDisplayer, app cf.Application, instances [][]cf.AppInstanceFields, errorCodes []string, startTimeout time.Duration) (ui *testterm.FakeUI, appRepo *testapi.FakeApplicationRepository, appInstancesRepo *testapi.FakeAppInstancesRepo, reqFactory *testreq.FakeReqFactory) {
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
		ApplicationStartTimeout: startTimeout,
	}

	appRepo = &testapi.FakeApplicationRepository{
		ReadApp:         app,
		UpdateAppResult: app,
	}
	appInstancesRepo = &testapi.FakeAppInstancesRepo{
		GetInstancesResponses:  instances,
		GetInstancesErrorCodes: errorCodes,
	}

	logRepo := &testapi.FakeLogsRepository{
		TailLogMessages: []*logmessage.Message{
			NewLogMessage("Log Line 1", app.Guid, LogMessageTypeStaging, time.Now()),
			NewLogMessage("Log Line 2", app.Guid, LogMessageTypeStaging, time.Now()),
		},
	}

	args := []string{"my-app"}
	reqFactory = &testreq.FakeReqFactory{Application: app}
	ui = callStart(args, config, reqFactory, displayApp, appRepo, appInstancesRepo, logRepo)
	return
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestStartCommandDefaultTimeouts", func() {

			cmd := NewStart(new(testterm.FakeUI), &configuration.Configuration{}, &testcmd.FakeAppDisplayer{}, &testapi.FakeApplicationRepository{}, &testapi.FakeAppInstancesRepo{}, &testapi.FakeLogsRepository{})
			assert.Equal(mr.T(), cmd.StagingTimeout, 15*time.Minute)
			assert.Equal(mr.T(), cmd.StartupTimeout, 5*time.Minute)
		})
		It("TestStartCommandSetsTimeoutsFromEnv", func() {

			oldStaging := os.Getenv("CF_STAGING_TIMEOUT")
			oldStart := os.Getenv("CF_STARTUP_TIMEOUT")
			defer func() {
				os.Setenv("CF_STAGING_TIMEOUT", oldStaging)
				os.Setenv("CF_STARTUP_TIMEOUT", oldStart)
			}()

			os.Setenv("CF_STAGING_TIMEOUT", "6")
			os.Setenv("CF_STARTUP_TIMEOUT", "3")
			cmd := NewStart(new(testterm.FakeUI), &configuration.Configuration{}, &testcmd.FakeAppDisplayer{}, &testapi.FakeApplicationRepository{}, &testapi.FakeAppInstancesRepo{}, &testapi.FakeLogsRepository{})
			assert.Equal(mr.T(), cmd.StagingTimeout, 6*time.Minute)
			assert.Equal(mr.T(), cmd.StartupTimeout, 3*time.Minute)
		})
		It("TestStartCommandFailsWithUsage", func() {
			mr.T().Parallel()

			config := &configuration.Configuration{}
			displayApp := &testcmd.FakeAppDisplayer{}
			appRepo := &testapi.FakeApplicationRepository{}
			appInstancesRepo := &testapi.FakeAppInstancesRepo{
				GetInstancesResponses: [][]cf.AppInstanceFields{
					[]cf.AppInstanceFields{},
				},
				GetInstancesErrorCodes: []string{""},
			}
			logRepo := &testapi.FakeLogsRepository{}

			reqFactory := &testreq.FakeReqFactory{}

			ui := callStart([]string{}, config, reqFactory, displayApp, appRepo, appInstancesRepo, logRepo)
			assert.True(mr.T(), ui.FailedWithUsage)

			ui = callStart([]string{"my-app"}, config, reqFactory, displayApp, appRepo, appInstancesRepo, logRepo)
			assert.False(mr.T(), ui.FailedWithUsage)
		})
		It("TestStartApplication", func() {
			mr.T().Parallel()

			displayApp := &testcmd.FakeAppDisplayer{}
			ui, appRepo, _, reqFactory := startAppWithInstancesAndErrors(mr.T(), displayApp, defaultAppForStart, defaultInstanceReponses, defaultInstanceErrorCodes, defaultStartTimeout)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"my-app", "my-org", "my-space", "my-user"},
				{"OK"},
				{"0 of 2 instances running", "2 starting"},
				{"Started"},
			})

			assert.Equal(mr.T(), reqFactory.ApplicationName, "my-app")
			assert.Equal(mr.T(), appRepo.UpdateAppGuid, "my-app-guid")
			assert.Equal(mr.T(), displayApp.AppToDisplay, defaultAppForStart)
		})
		It("TestStartApplicationOnlyShowsCurrentStagingLogs", func() {
			mr.T().Parallel()

			displayApp := &testcmd.FakeAppDisplayer{}
			reqFactory := &testreq.FakeReqFactory{Application: defaultAppForStart}
			appRepo := &testapi.FakeApplicationRepository{
				ReadApp:         defaultAppForStart,
				UpdateAppResult: defaultAppForStart,
			}
			appInstancesRepo := &testapi.FakeAppInstancesRepo{
				GetInstancesResponses:  defaultInstanceReponses,
				GetInstancesErrorCodes: defaultInstanceErrorCodes,
			}

			currentTime := time.Now()
			wrongSourceName := "DEA"
			correctSourceName := "STG"

			logRepo := &testapi.FakeLogsRepository{
				TailLogMessages: []*logmessage.Message{
					NewLogMessage("Log Line 1", defaultAppForStart.Guid, wrongSourceName, currentTime),
					NewLogMessage("Log Line 2", defaultAppForStart.Guid, correctSourceName, currentTime),
					NewLogMessage("Log Line 3", defaultAppForStart.Guid, correctSourceName, currentTime),
					NewLogMessage("Log Line 4", defaultAppForStart.Guid, wrongSourceName, currentTime),
				},
			}

			ui := callStart([]string{"my-app"}, &configuration.Configuration{}, reqFactory, displayApp, appRepo, appInstancesRepo, logRepo)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Log Line 2"},
				{"Log Line 3"},
			})
			testassert.SliceDoesNotContain(mr.T(), ui.Outputs, testassert.Lines{
				{"Log Line 1"},
				{"Log Line 4"},
			})
		})
		It("TestStartApplicationWhenAppHasNoURL", func() {
			mr.T().Parallel()

			displayApp := &testcmd.FakeAppDisplayer{}
			app := defaultAppForStart
			app.Routes = []cf.RouteSummary{}
			appInstance := cf.AppInstanceFields{}
			appInstance.State = cf.InstanceRunning
			instances := [][]cf.AppInstanceFields{
				[]cf.AppInstanceFields{appInstance},
			}

			errorCodes := []string{""}
			ui, appRepo, _, reqFactory := startAppWithInstancesAndErrors(mr.T(), displayApp, app, instances, errorCodes, defaultStartTimeout)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"my-app"},
				{"OK"},
				{"Started"},
			})

			assert.Equal(mr.T(), reqFactory.ApplicationName, "my-app")
			assert.Equal(mr.T(), appRepo.UpdateAppGuid, "my-app-guid")
		})
		It("TestStartApplicationWhenAppIsStillStaging", func() {
			mr.T().Parallel()

			displayApp := &testcmd.FakeAppDisplayer{}
			appInstance := cf.AppInstanceFields{}
			appInstance.State = cf.InstanceDown
			appInstance2 := cf.AppInstanceFields{}
			appInstance2.State = cf.InstanceStarting
			appInstance3 := cf.AppInstanceFields{}
			appInstance3.State = cf.InstanceStarting
			appInstance4 := cf.AppInstanceFields{}
			appInstance4.State = cf.InstanceStarting
			appInstance5 := cf.AppInstanceFields{}
			appInstance5.State = cf.InstanceRunning
			appInstance6 := cf.AppInstanceFields{}
			appInstance6.State = cf.InstanceRunning
			instances := [][]cf.AppInstanceFields{
				[]cf.AppInstanceFields{},
				[]cf.AppInstanceFields{},
				[]cf.AppInstanceFields{appInstance, appInstance2},
				[]cf.AppInstanceFields{appInstance3, appInstance4},
				[]cf.AppInstanceFields{appInstance5, appInstance6},
			}

			errorCodes := []string{cf.APP_NOT_STAGED, cf.APP_NOT_STAGED, "", "", ""}

			ui, _, appInstancesRepo, _ := startAppWithInstancesAndErrors(mr.T(), displayApp, defaultAppForStart, instances, errorCodes, defaultStartTimeout)

			assert.Equal(mr.T(), appInstancesRepo.GetInstancesAppGuid, "my-app-guid")

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Log Line 1"},
				{"Log Line 2"},
				{"0 of 2 instances running", "2 starting"},
			})
		})
		It("TestStartApplicationWhenStagingFails", func() {
			mr.T().Skip()
			mr.T().Parallel()

			displayApp := &testcmd.FakeAppDisplayer{}
			instances := [][]cf.AppInstanceFields{[]cf.AppInstanceFields{}}
			errorCodes := []string{"170001"}

			ui, _, _, _ := startAppWithInstancesAndErrors(mr.T(), displayApp, defaultAppForStart, instances, errorCodes, defaultStartTimeout)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"my-app"},
				{"OK"},
				{"FAILED"},
				{"Error staging app"},
			})
		})
		It("TestStartApplicationWhenOneInstanceFlaps", func() {
			mr.T().Parallel()

			displayApp := &testcmd.FakeAppDisplayer{}
			appInstance := cf.AppInstanceFields{}
			appInstance.State = cf.InstanceStarting
			appInstance2 := cf.AppInstanceFields{}
			appInstance2.State = cf.InstanceStarting
			appInstance3 := cf.AppInstanceFields{}
			appInstance3.State = cf.InstanceStarting
			appInstance4 := cf.AppInstanceFields{}
			appInstance4.State = cf.InstanceFlapping
			instances := [][]cf.AppInstanceFields{
				[]cf.AppInstanceFields{appInstance, appInstance2},
				[]cf.AppInstanceFields{appInstance3, appInstance4},
			}

			errorCodes := []string{"", ""}

			ui, _, _, _ := startAppWithInstancesAndErrors(mr.T(), displayApp, defaultAppForStart, instances, errorCodes, defaultStartTimeout)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"my-app"},
				{"OK"},
				{"0 of 2 instances running", "1 starting", "1 failing"},
				{"FAILED"},
				{"Start unsuccessful"},
			})
		})
		It("TestStartApplicationWhenStartTimesOut", func() {
			mr.T().Parallel()

			displayApp := &testcmd.FakeAppDisplayer{}
			appInstance := cf.AppInstanceFields{}
			appInstance.State = cf.InstanceStarting
			appInstance2 := cf.AppInstanceFields{}
			appInstance2.State = cf.InstanceStarting
			appInstance3 := cf.AppInstanceFields{}
			appInstance3.State = cf.InstanceStarting
			appInstance4 := cf.AppInstanceFields{}
			appInstance4.State = cf.InstanceDown
			appInstance5 := cf.AppInstanceFields{}
			appInstance5.State = cf.InstanceDown
			appInstance6 := cf.AppInstanceFields{}
			appInstance6.State = cf.InstanceDown
			instances := [][]cf.AppInstanceFields{
				[]cf.AppInstanceFields{appInstance, appInstance2},
				[]cf.AppInstanceFields{appInstance3, appInstance4},
				[]cf.AppInstanceFields{appInstance5, appInstance6},
			}

			errorCodes := []string{"500", "500", "500"}

			ui, _, _, _ := startAppWithInstancesAndErrors(mr.T(), displayApp, defaultAppForStart, instances, errorCodes, 0)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Starting", "my-app"},
				{"OK"},
				{"FAILED"},
				{"Start app timeout"},
			})
			testassert.SliceDoesNotContain(mr.T(), ui.Outputs, testassert.Lines{
				{"instances running"},
			})
		})
		It("TestStartApplicationWhenStartFails", func() {
			mr.T().Parallel()

			config := &configuration.Configuration{}
			displayApp := &testcmd.FakeAppDisplayer{}
			app := cf.Application{}
			app.Name = "my-app"
			app.Guid = "my-app-guid"
			appRepo := &testapi.FakeApplicationRepository{ReadApp: app, UpdateErr: true}
			appInstancesRepo := &testapi.FakeAppInstancesRepo{}
			logRepo := &testapi.FakeLogsRepository{}
			args := []string{"my-app"}
			reqFactory := &testreq.FakeReqFactory{Application: app}
			ui := callStart(args, config, reqFactory, displayApp, appRepo, appInstancesRepo, logRepo)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"my-app"},
				{"FAILED"},
				{"Error updating app."},
			})
			assert.Equal(mr.T(), appRepo.UpdateAppGuid, "my-app-guid")
		})
		It("TestStartApplicationIsAlreadyStarted", func() {
			mr.T().Parallel()

			displayApp := &testcmd.FakeAppDisplayer{}
			config := &configuration.Configuration{}
			app := cf.Application{}
			app.Name = "my-app"
			app.Guid = "my-app-guid"
			app.State = "started"
			appRepo := &testapi.FakeApplicationRepository{ReadApp: app}
			appInstancesRepo := &testapi.FakeAppInstancesRepo{}
			logRepo := &testapi.FakeLogsRepository{}

			reqFactory := &testreq.FakeReqFactory{Application: app}

			args := []string{"my-app"}
			ui := callStart(args, config, reqFactory, displayApp, appRepo, appInstancesRepo, logRepo)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"my-app", "is already started"},
			})

			assert.Equal(mr.T(), appRepo.UpdateAppGuid, "")
		})
		It("TestStartApplicationWithLoggingFailure", func() {
			mr.T().Parallel()

			token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{Username: "my-user"})
			assert.NoError(mr.T(), err)
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

			displayApp := &testcmd.FakeAppDisplayer{}

			appRepo := &testapi.FakeApplicationRepository{ReadApp: defaultAppForStart}
			appInstancesRepo := &testapi.FakeAppInstancesRepo{
				GetInstancesResponses:  defaultInstanceReponses,
				GetInstancesErrorCodes: defaultInstanceErrorCodes,
			}

			logRepo := &testapi.FakeLogsRepository{
				TailLogErr: errors.New("Ooops"),
			}

			reqFactory := &testreq.FakeReqFactory{Application: defaultAppForStart}

			ui := callStart([]string{"my-app"}, config, reqFactory, displayApp, appRepo, appInstancesRepo, logRepo)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				testassert.Line{"error tailing logs"},
				testassert.Line{"Ooops"},
			})
		})
	})
}
