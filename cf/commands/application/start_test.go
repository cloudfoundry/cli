package application_test

import (
	"os"
	"time"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/api/app_instances"
	testAppInstanaces "github.com/cloudfoundry/cli/cf/api/app_instances/fakes"
	"github.com/cloudfoundry/cli/cf/api/applications"
	testApplication "github.com/cloudfoundry/cli/cf/api/applications/fakes"
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	. "github.com/cloudfoundry/cli/cf/commands/application"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testlogs "github.com/cloudfoundry/cli/testhelpers/logs"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

var _ = Describe("start command", func() {
	var (
		ui                        *testterm.FakeUI
		defaultAppForStart        = models.Application{}
		defaultInstanceResponses  = [][]models.AppInstanceFields{}
		defaultInstanceErrorCodes = []string{"", ""}
		requirementsFactory       *testreq.FakeReqFactory
		logsForTail               []*logmessage.LogMessage
		logRepo                   *testapi.FakeLogsRepository
	)

	getInstance := func(appGuid string) (instances []models.AppInstanceFields, apiErr error) {
		if len(defaultInstanceResponses) > 0 {
			instances = defaultInstanceResponses[0]
			if len(defaultInstanceResponses) > 1 {
				defaultInstanceResponses = defaultInstanceResponses[1:]
			}
		}
		if len(defaultInstanceErrorCodes) > 0 {
			errorCode := defaultInstanceErrorCodes[0]
			if len(defaultInstanceErrorCodes) > 1 {
				defaultInstanceErrorCodes = defaultInstanceErrorCodes[1:]
			}
			if errorCode != "" {
				apiErr = errors.NewHttpError(400, errorCode, "Error staging app")
			}
		}
		return
	}

	BeforeEach(func() {
		ui = new(testterm.FakeUI)
		requirementsFactory = &testreq.FakeReqFactory{}

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

		defaultInstanceResponses = [][]models.AppInstanceFields{
			[]models.AppInstanceFields{instance1, instance2},
			[]models.AppInstanceFields{instance1, instance2},
			[]models.AppInstanceFields{instance3, instance4},
		}

		logsForTail = []*logmessage.LogMessage{}
		logRepo = new(testapi.FakeLogsRepository)
		logRepo.TailLogsForStub = func(appGuid string, onConnect func(), onMessage func(*logmessage.LogMessage)) error {
			onConnect()
			for _, log := range logsForTail {
				onMessage(log)
			}
			return nil
		}
	})

	callStart := func(args []string, config core_config.Reader, requirementsFactory *testreq.FakeReqFactory, displayApp ApplicationDisplayer, appRepo applications.ApplicationRepository, appInstancesRepo app_instances.AppInstancesRepository, logRepo api.LogsRepository) (ui *testterm.FakeUI) {
		ui = new(testterm.FakeUI)

		cmd := NewStart(ui, config, displayApp, appRepo, appInstancesRepo, logRepo)
		cmd.StagingTimeout = 100 * time.Millisecond
		cmd.StartupTimeout = 200 * time.Millisecond
		cmd.PingerThrottle = 50 * time.Millisecond

		testcmd.RunCommand(cmd, args, requirementsFactory)
		return
	}

	startAppWithInstancesAndErrors := func(displayApp ApplicationDisplayer, app models.Application, requirementsFactory *testreq.FakeReqFactory) (ui *testterm.FakeUI, appRepo *testApplication.FakeApplicationRepository, appInstancesRepo *testAppInstanaces.FakeAppInstancesRepository) {
		configRepo := testconfig.NewRepositoryWithDefaults()
		appRepo = &testApplication.FakeApplicationRepository{
			UpdateAppResult: app,
		}
		appRepo.ReadReturns.App = app
		appInstancesRepo = &testAppInstanaces.FakeAppInstancesRepository{}
		appInstancesRepo.GetInstancesStub = getInstance

		logsForTail = []*logmessage.LogMessage{
			testlogs.NewLogMessage("Log Line 1", app.Guid, LogMessageTypeStaging, time.Now()),
			testlogs.NewLogMessage("Log Line 2", app.Guid, LogMessageTypeStaging, time.Now()),
		}

		args := []string{"my-app"}

		requirementsFactory.Application = app
		ui = callStart(args, configRepo, requirementsFactory, displayApp, appRepo, appInstancesRepo, logRepo)
		return
	}

	It("fails requirements when not logged in", func() {
		requirementsFactory.LoginSuccess = false
		cmd := NewStart(new(testterm.FakeUI), testconfig.NewRepository(), &testcmd.FakeAppDisplayer{}, &testApplication.FakeApplicationRepository{}, &testAppInstanaces.FakeAppInstancesRepository{}, &testapi.FakeLogsRepository{})

		Expect(testcmd.RunCommand(cmd, []string{"some-app-name"}, requirementsFactory)).To(BeFalse())
	})

	Describe("timeouts", func() {
		It("has sane default timeout values", func() {
			cmd := NewStart(new(testterm.FakeUI), testconfig.NewRepository(), &testcmd.FakeAppDisplayer{}, &testApplication.FakeApplicationRepository{}, &testAppInstanaces.FakeAppInstancesRepository{}, &testapi.FakeLogsRepository{})
			Expect(cmd.StagingTimeout).To(Equal(15 * time.Minute))
			Expect(cmd.StartupTimeout).To(Equal(5 * time.Minute))
		})

		It("can read timeout values from environment variables", func() {
			oldStaging := os.Getenv("CF_STAGING_TIMEOUT")
			oldStart := os.Getenv("CF_STARTUP_TIMEOUT")
			defer func() {
				os.Setenv("CF_STAGING_TIMEOUT", oldStaging)
				os.Setenv("CF_STARTUP_TIMEOUT", oldStart)
			}()

			os.Setenv("CF_STAGING_TIMEOUT", "6")
			os.Setenv("CF_STARTUP_TIMEOUT", "3")
			cmd := NewStart(new(testterm.FakeUI), testconfig.NewRepository(), &testcmd.FakeAppDisplayer{}, &testApplication.FakeApplicationRepository{}, &testAppInstanaces.FakeAppInstancesRepository{}, &testapi.FakeLogsRepository{})
			Expect(cmd.StagingTimeout).To(Equal(6 * time.Minute))
			Expect(cmd.StartupTimeout).To(Equal(3 * time.Minute))
		})

		Describe("when the staging timeout is zero seconds", func() {
			var (
				app models.Application
				cmd *Start
			)

			BeforeEach(func() {
				app = defaultAppForStart

				instances := []models.AppInstanceFields{models.AppInstanceFields{}}
				appRepo := &testApplication.FakeApplicationRepository{
					UpdateAppResult: app,
				}
				appRepo.ReadReturns.App = app
				appInstancesRepo := &testAppInstanaces.FakeAppInstancesRepository{}
				appInstancesRepo.GetInstancesReturns(instances, errors.New("Error staging app"))

				requirementsFactory.LoginSuccess = true
				requirementsFactory.Application = app
				config := testconfig.NewRepository()
				displayApp := &testcmd.FakeAppDisplayer{}

				cmd = NewStart(ui, config, displayApp, appRepo, appInstancesRepo, logRepo)
				cmd.StagingTimeout = 1
				cmd.PingerThrottle = 1
				cmd.StartupTimeout = 1
			})

			It("can still respond to staging failures", func() {
				testcmd.RunCommand(cmd, []string{"my-app"}, requirementsFactory)

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"my-app"},
					[]string{"FAILED"},
					[]string{"Error staging app"},
				))
			})
		})
	})

	Context("when logged in", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
		})

		It("fails with usage when provided with no args", func() {
			config := testconfig.NewRepository()
			displayApp := &testcmd.FakeAppDisplayer{}
			appRepo := &testApplication.FakeApplicationRepository{}
			appInstancesRepo := &testAppInstanaces.FakeAppInstancesRepository{}
			logRepo := &testapi.FakeLogsRepository{}

			ui := callStart([]string{}, config, requirementsFactory, displayApp, appRepo, appInstancesRepo, logRepo)
			Expect(ui.FailedWithUsage).To(BeTrue())
		})

		It("uses uses proper org name and space name", func() {
			config := testconfig.NewRepositoryWithDefaults()
			displayApp := &testcmd.FakeAppDisplayer{}
			appRepo := &testApplication.FakeApplicationRepository{}
			appInstancesRepo := &testAppInstanaces.FakeAppInstancesRepository{}

			appRepo.ReadReturns.App = defaultAppForStart
			appInstancesRepo = &testAppInstanaces.FakeAppInstancesRepository{}
			appInstancesRepo.GetInstancesStub = getInstance

			cmd := NewStart(ui, config, displayApp, appRepo, appInstancesRepo, logRepo)
			cmd.StagingTimeout = 100 * time.Millisecond
			cmd.StartupTimeout = 200 * time.Millisecond
			cmd.PingerThrottle = 50 * time.Millisecond
			cmd.ApplicationStart(defaultAppForStart, "some-org", "some-space")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"my-app", "some-org", "some-space", "my-user"},
				[]string{"OK"},
			))
		})

		It("starts an app, when given the app's name", func() {
			displayApp := &testcmd.FakeAppDisplayer{}
			ui, appRepo, _ := startAppWithInstancesAndErrors(displayApp, defaultAppForStart, requirementsFactory)

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"my-app", "my-org", "my-space", "my-user"},
				[]string{"OK"},
				[]string{"0 of 2 instances running", "2 starting"},
				[]string{"started"},
			))

			Expect(requirementsFactory.ApplicationName).To(Equal("my-app"))
			Expect(appRepo.UpdateAppGuid).To(Equal("my-app-guid"))
			Expect(displayApp.AppToDisplay).To(Equal(defaultAppForStart))
		})

		It("only displays staging logs when an app is starting", func() {
			displayApp := &testcmd.FakeAppDisplayer{}
			requirementsFactory.Application = defaultAppForStart
			appRepo := &testApplication.FakeApplicationRepository{
				UpdateAppResult: defaultAppForStart,
			}
			appRepo.ReadReturns.App = defaultAppForStart

			appInstancesRepo := &testAppInstanaces.FakeAppInstancesRepository{}

			currentTime := time.Now()
			wrongSourceName := "DEA"
			correctSourceName := "STG"

			logsForTail = []*logmessage.LogMessage{
				testlogs.NewLogMessage("Log Line 1", defaultAppForStart.Guid, wrongSourceName, currentTime),
				testlogs.NewLogMessage("Log Line 2", defaultAppForStart.Guid, correctSourceName, currentTime),
				testlogs.NewLogMessage("Log Line 3", defaultAppForStart.Guid, correctSourceName, currentTime),
				testlogs.NewLogMessage("Log Line 4", defaultAppForStart.Guid, wrongSourceName, currentTime),
			}

			ui := callStart([]string{"my-app"}, testconfig.NewRepository(), requirementsFactory, displayApp, appRepo, appInstancesRepo, logRepo)

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Log Line 2"},
				[]string{"Log Line 3"},
			))
			Expect(ui.Outputs).ToNot(ContainSubstrings(
				[]string{"Log Line 1"},
				[]string{"Log Line 4"},
			))
		})

		It("gracefully handles starting an app that is still staging", func() {
			displayApp := &testcmd.FakeAppDisplayer{}

			logRepoClosed := make(chan struct{})

			logRepo.TailLogsForStub = func(appGuid string, onConnect func(), onMessage func(*logmessage.LogMessage)) error {
				onConnect()
				onMessage(testlogs.NewLogMessage("Before close", appGuid, LogMessageTypeStaging, time.Now()))

				<-logRepoClosed

				time.Sleep(50 * time.Millisecond)
				onMessage(testlogs.NewLogMessage("After close 1", appGuid, LogMessageTypeStaging, time.Now()))
				onMessage(testlogs.NewLogMessage("After close 2", appGuid, LogMessageTypeStaging, time.Now()))

				return nil
			}

			logRepo.CloseStub = func() {
				close(logRepoClosed)
			}

			defaultInstanceResponses = [][]models.AppInstanceFields{
				[]models.AppInstanceFields{},
				[]models.AppInstanceFields{},
				[]models.AppInstanceFields{{State: models.InstanceDown}, {State: models.InstanceStarting}},
				[]models.AppInstanceFields{{State: models.InstanceStarting}, {State: models.InstanceStarting}},
				[]models.AppInstanceFields{{State: models.InstanceRunning}, {State: models.InstanceRunning}},
			}

			defaultInstanceErrorCodes = []string{errors.APP_NOT_STAGED, errors.APP_NOT_STAGED, "", "", ""}

			ui, _, appInstancesRepo := startAppWithInstancesAndErrors(displayApp, defaultAppForStart, requirementsFactory)

			Expect(appInstancesRepo.GetInstancesArgsForCall(0)).To(Equal("my-app-guid"))

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Before close"},
				[]string{"After close 1"},
				[]string{"After close 2"},
				[]string{"0 of 2 instances running", "2 starting"},
			))
		})

		It("displays an error message when staging fails", func() {
			displayApp := &testcmd.FakeAppDisplayer{}
			defaultInstanceResponses = [][]models.AppInstanceFields{[]models.AppInstanceFields{}}
			defaultInstanceErrorCodes = []string{"170001"}

			ui, _, _ := startAppWithInstancesAndErrors(displayApp, defaultAppForStart, requirementsFactory)

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"my-app"},
				[]string{"FAILED"},
				[]string{"Error staging app"},
			))
		})

		Context("when an app instance is flapping", func() {
			It("fails and alerts the user", func() {
				displayApp := &testcmd.FakeAppDisplayer{}
				appInstance := models.AppInstanceFields{}
				appInstance.State = models.InstanceStarting
				appInstance2 := models.AppInstanceFields{}
				appInstance2.State = models.InstanceStarting
				appInstance3 := models.AppInstanceFields{}
				appInstance3.State = models.InstanceStarting
				appInstance4 := models.AppInstanceFields{}
				appInstance4.State = models.InstanceFlapping
				defaultInstanceResponses = [][]models.AppInstanceFields{
					[]models.AppInstanceFields{appInstance, appInstance2},
					[]models.AppInstanceFields{appInstance3, appInstance4},
				}

				defaultInstanceErrorCodes = []string{"", ""}

				ui, _, _ := startAppWithInstancesAndErrors(displayApp, defaultAppForStart, requirementsFactory)

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"my-app"},
					[]string{"0 of 2 instances running", "1 starting", "1 failing"},
					[]string{"FAILED"},
					[]string{"Start unsuccessful"},
				))
			})
		})

		It("tells the user about the failure when waiting for the app to start times out", func() {
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
			defaultInstanceResponses = [][]models.AppInstanceFields{
				[]models.AppInstanceFields{appInstance, appInstance2},
				[]models.AppInstanceFields{appInstance3, appInstance4},
				[]models.AppInstanceFields{appInstance5, appInstance6},
			}

			defaultInstanceErrorCodes = []string{errors.APP_NOT_STAGED, errors.APP_NOT_STAGED, errors.APP_NOT_STAGED}

			ui, _, _ := startAppWithInstancesAndErrors(displayApp, defaultAppForStart, requirementsFactory)

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Starting", "my-app"},
				[]string{"FAILED"},
				[]string{"Start app timeout"},
			))
			Expect(ui.Outputs).ToNot(ContainSubstrings([]string{"instances running"}))
		})

		It("tells the user about the failure when starting the app fails", func() {
			config := testconfig.NewRepository()
			displayApp := &testcmd.FakeAppDisplayer{}
			app := models.Application{}
			app.Name = "my-app"
			app.Guid = "my-app-guid"
			appRepo := &testApplication.FakeApplicationRepository{UpdateErr: true}
			appRepo.ReadReturns.App = app
			appInstancesRepo := &testAppInstanaces.FakeAppInstancesRepository{}
			args := []string{"my-app"}
			requirementsFactory.Application = app
			ui := callStart(args, config, requirementsFactory, displayApp, appRepo, appInstancesRepo, logRepo)

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"my-app"},
				[]string{"FAILED"},
				[]string{"Error updating app."},
			))
			Expect(appRepo.UpdateAppGuid).To(Equal("my-app-guid"))
		})

		It("warns the user when the app is already running", func() {
			displayApp := &testcmd.FakeAppDisplayer{}
			config := testconfig.NewRepository()
			app := models.Application{}
			app.Name = "my-app"
			app.Guid = "my-app-guid"
			app.State = "started"
			appRepo := &testApplication.FakeApplicationRepository{}
			appRepo.ReadReturns.App = app
			appInstancesRepo := &testAppInstanaces.FakeAppInstancesRepository{}

			requirementsFactory.Application = app

			args := []string{"my-app"}
			ui := callStart(args, config, requirementsFactory, displayApp, appRepo, appInstancesRepo, logRepo)

			Expect(ui.Outputs).To(ContainSubstrings([]string{"my-app", "is already started"}))

			Expect(appRepo.UpdateAppGuid).To(Equal(""))
		})

		It("tells the user when connecting to the log server fails", func() {
			configRepo := testconfig.NewRepositoryWithDefaults()
			displayApp := &testcmd.FakeAppDisplayer{}

			appRepo := &testApplication.FakeApplicationRepository{}
			appRepo.ReadReturns.App = defaultAppForStart
			appInstancesRepo := &testAppInstanaces.FakeAppInstancesRepository{}

			logRepo.TailLogsForReturns(errors.New("Ooops"))

			requirementsFactory.Application = defaultAppForStart

			ui := callStart([]string{"my-app"}, configRepo, requirementsFactory, displayApp, appRepo, appInstancesRepo, logRepo)

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"error tailing logs"},
				[]string{"Ooops"},
			))
		})
	})
})
