package application_test

import (
	"cf"
	"cf/api"
	. "cf/commands/application"
	"cf/configuration"
	"cf/models"
	"errors"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	mr "github.com/tjarratt/mr_t"
	"os"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"time"
)

var _ = Describe("Testing with ginkgo", func() {
	var (
		defaultAppForStart        = models.Application{}
		defaultInstanceReponses   = [][]models.AppInstanceFields{}
		defaultInstanceErrorCodes = []string{"", ""}
		defaultStartTimeout       = 50 * time.Millisecond
	)

	BeforeEach(func() {
		defaultAppForStart.Name = "my-app"
		defaultAppForStart.Guid = "my-app-guid"
		defaultAppForStart.InstanceCount = 2

		domain := models.DomainFields{}
		domain.Name = "example.com"

		route := models.RouteSummary{}
		route.Host = "my-app"
		route.Domain = domain

		defaultAppForStart.Routes = []models.RouteSummary{route}

		instance1 := models.AppInstanceFields{}
		instance1.State = models.InstanceStarting

		instance2 := models.AppInstanceFields{}
		instance2.State = models.InstanceStarting

		instance3 := models.AppInstanceFields{}
		instance3.State = models.InstanceRunning

		instance4 := models.AppInstanceFields{}
		instance4.State = models.InstanceStarting

		defaultInstanceReponses = [][]models.AppInstanceFields{
			[]models.AppInstanceFields{instance1, instance2},
			[]models.AppInstanceFields{instance1, instance2},
			[]models.AppInstanceFields{instance3, instance4},
		}
	})

	It("TestStartCommandDefaultTimeouts", func() {
		cmd := NewStart(new(testterm.FakeUI), testconfig.NewRepository(), &testcmd.FakeAppDisplayer{}, &testapi.FakeApplicationRepository{}, &testapi.FakeAppInstancesRepo{}, &testapi.FakeLogsRepository{})
		Expect(cmd.StagingTimeout).To(Equal(15 * time.Minute))
		Expect(cmd.StartupTimeout).To(Equal(5 * time.Minute))
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
		cmd := NewStart(new(testterm.FakeUI), testconfig.NewRepository(), &testcmd.FakeAppDisplayer{}, &testapi.FakeApplicationRepository{}, &testapi.FakeAppInstancesRepo{}, &testapi.FakeLogsRepository{})
		Expect(cmd.StagingTimeout).To(Equal(6 * time.Minute))
		Expect(cmd.StartupTimeout).To(Equal(3 * time.Minute))
	})

	It("TestStartCommandFailsWithUsage", func() {
		config := testconfig.NewRepository()
		displayApp := &testcmd.FakeAppDisplayer{}
		appRepo := &testapi.FakeApplicationRepository{}
		appInstancesRepo := &testapi.FakeAppInstancesRepo{
			GetInstancesResponses: [][]models.AppInstanceFields{
				[]models.AppInstanceFields{},
			},
			GetInstancesErrorCodes: []string{""},
		}
		logRepo := &testapi.FakeLogsRepository{}

		reqFactory := &testreq.FakeReqFactory{}

		ui := callStart([]string{}, config, reqFactory, displayApp, appRepo, appInstancesRepo, logRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callStart([]string{"my-app"}, config, reqFactory, displayApp, appRepo, appInstancesRepo, logRepo)
		Expect(ui.FailedWithUsage).To(BeFalse())
	})

	It("TestStartApplication", func() {
		displayApp := &testcmd.FakeAppDisplayer{}
		ui, appRepo, _, reqFactory := startAppWithInstancesAndErrors(displayApp, defaultAppForStart, defaultInstanceReponses, defaultInstanceErrorCodes, defaultStartTimeout)

		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"my-app", "my-org", "my-space", "my-user"},
			{"OK"},
			{"0 of 2 instances running", "2 starting"},
			{"Started"},
		})

		Expect(reqFactory.ApplicationName).To(Equal("my-app"))
		Expect(appRepo.UpdateAppGuid).To(Equal("my-app-guid"))
		Expect(displayApp.AppToDisplay).To(Equal(defaultAppForStart))
	})

	It("TestStartApplicationOnlyShowsCurrentStagingLogs", func() {
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

		ui := callStart([]string{"my-app"}, testconfig.NewRepository(), reqFactory, displayApp, appRepo, appInstancesRepo, logRepo)

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
		displayApp := &testcmd.FakeAppDisplayer{}
		app := defaultAppForStart
		app.Routes = []models.RouteSummary{}
		appInstance := models.AppInstanceFields{}
		appInstance.State = models.InstanceRunning
		instances := [][]models.AppInstanceFields{
			[]models.AppInstanceFields{appInstance},
			[]models.AppInstanceFields{appInstance},
		}

		errorCodes := []string{""}
		ui, appRepo, _, reqFactory := startAppWithInstancesAndErrors(displayApp, app, instances, errorCodes, defaultStartTimeout)

		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"my-app"},
			{"OK"},
			{"Started"},
		})

		Expect(reqFactory.ApplicationName).To(Equal("my-app"))
		Expect(appRepo.UpdateAppGuid).To(Equal("my-app-guid"))
	})

	It("TestStartApplicationWhenAppIsStillStaging", func() {
		displayApp := &testcmd.FakeAppDisplayer{}
		appInstance := models.AppInstanceFields{}
		appInstance.State = models.InstanceDown
		appInstance2 := models.AppInstanceFields{}
		appInstance2.State = models.InstanceStarting
		appInstance3 := models.AppInstanceFields{}
		appInstance3.State = models.InstanceStarting
		appInstance4 := models.AppInstanceFields{}
		appInstance4.State = models.InstanceStarting
		appInstance5 := models.AppInstanceFields{}
		appInstance5.State = models.InstanceRunning
		appInstance6 := models.AppInstanceFields{}
		appInstance6.State = models.InstanceRunning
		instances := [][]models.AppInstanceFields{
			[]models.AppInstanceFields{},
			[]models.AppInstanceFields{},
			[]models.AppInstanceFields{appInstance, appInstance2},
			[]models.AppInstanceFields{appInstance3, appInstance4},
			[]models.AppInstanceFields{appInstance5, appInstance6},
		}

		errorCodes := []string{cf.APP_NOT_STAGED, cf.APP_NOT_STAGED, "", "", ""}

		ui, _, appInstancesRepo, _ := startAppWithInstancesAndErrors(displayApp, defaultAppForStart, instances, errorCodes, defaultStartTimeout)

		Expect(appInstancesRepo.GetInstancesAppGuid).To(Equal("my-app-guid"))

		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Log Line 1"},
			{"Log Line 2"},
			{"0 of 2 instances running", "2 starting"},
		})
	})

	XIt("TestStartApplicationWhenStagingFails", func() {
		// TODO: fix this flakey test
		displayApp := &testcmd.FakeAppDisplayer{}
		instances := [][]models.AppInstanceFields{[]models.AppInstanceFields{}}
		errorCodes := []string{"170001"}

		ui, _, _, _ := startAppWithInstancesAndErrors(displayApp, defaultAppForStart, instances, errorCodes, defaultStartTimeout)

		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"my-app"},
			{"OK"},
			{"FAILED"},
			{"Error staging app"},
		})
	})

	It("TestStartApplicationWhenOneInstanceFlaps", func() {
		displayApp := &testcmd.FakeAppDisplayer{}
		appInstance := models.AppInstanceFields{}
		appInstance.State = models.InstanceStarting
		appInstance2 := models.AppInstanceFields{}
		appInstance2.State = models.InstanceStarting
		appInstance3 := models.AppInstanceFields{}
		appInstance3.State = models.InstanceStarting
		appInstance4 := models.AppInstanceFields{}
		appInstance4.State = models.InstanceFlapping
		instances := [][]models.AppInstanceFields{
			[]models.AppInstanceFields{appInstance, appInstance2},
			[]models.AppInstanceFields{appInstance3, appInstance4},
		}

		errorCodes := []string{"", ""}

		ui, _, _, _ := startAppWithInstancesAndErrors(displayApp, defaultAppForStart, instances, errorCodes, defaultStartTimeout)

		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"my-app"},
			{"OK"},
			{"0 of 2 instances running", "1 starting", "1 failing"},
			{"FAILED"},
			{"Start unsuccessful"},
		})
	})

	It("TestStartApplicationWhenStartTimesOut", func() {
		displayApp := &testcmd.FakeAppDisplayer{}
		appInstance := models.AppInstanceFields{}
		appInstance.State = models.InstanceStarting
		appInstance2 := models.AppInstanceFields{}
		appInstance2.State = models.InstanceStarting
		appInstance3 := models.AppInstanceFields{}
		appInstance3.State = models.InstanceStarting
		appInstance4 := models.AppInstanceFields{}
		appInstance4.State = models.InstanceDown
		appInstance5 := models.AppInstanceFields{}
		appInstance5.State = models.InstanceDown
		appInstance6 := models.AppInstanceFields{}
		appInstance6.State = models.InstanceDown
		instances := [][]models.AppInstanceFields{
			[]models.AppInstanceFields{appInstance, appInstance2},
			[]models.AppInstanceFields{appInstance3, appInstance4},
			[]models.AppInstanceFields{appInstance5, appInstance6},
		}

		errorCodes := []string{cf.APP_NOT_STAGED, cf.APP_NOT_STAGED, cf.APP_NOT_STAGED}

		ui, _, _, _ := startAppWithInstancesAndErrors(displayApp, defaultAppForStart, instances, errorCodes, 0)

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
		config := testconfig.NewRepository()
		displayApp := &testcmd.FakeAppDisplayer{}
		app := models.Application{}
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
		Expect(appRepo.UpdateAppGuid).To(Equal("my-app-guid"))
	})

	It("TestStartApplicationIsAlreadyStarted", func() {
		displayApp := &testcmd.FakeAppDisplayer{}
		config := testconfig.NewRepository()
		app := models.Application{}
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

		Expect(appRepo.UpdateAppGuid).To(Equal(""))
	})

	It("TestStartApplicationWithLoggingFailure", func() {
		configRepo := testconfig.NewRepositoryWithDefaults()
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

		ui := callStart([]string{"my-app"}, configRepo, reqFactory, displayApp, appRepo, appInstancesRepo, logRepo)

		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			testassert.Line{"error tailing logs"},
			testassert.Line{"Ooops"},
		})
	})
})

func callStart(args []string, config configuration.Reader, reqFactory *testreq.FakeReqFactory, displayApp ApplicationDisplayer, appRepo api.ApplicationRepository, appInstancesRepo api.AppInstancesRepository, logRepo api.LogsRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("start", args)

	cmd := NewStart(ui, config, displayApp, appRepo, appInstancesRepo, logRepo)
	cmd.StagingTimeout = 50 * time.Millisecond
	cmd.StartupTimeout = 50 * time.Millisecond
	cmd.PingerThrottle = 50 * time.Millisecond

	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}

func startAppWithInstancesAndErrors(displayApp ApplicationDisplayer, app models.Application, instances [][]models.AppInstanceFields, errorCodes []string, startTimeout time.Duration) (ui *testterm.FakeUI, appRepo *testapi.FakeApplicationRepository, appInstancesRepo *testapi.FakeAppInstancesRepo, reqFactory *testreq.FakeReqFactory) {
	configRepo := testconfig.NewRepositoryWithDefaults()
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
	ui = callStart(args, configRepo, reqFactory, displayApp, appRepo, appInstancesRepo, logRepo)
	return
}
