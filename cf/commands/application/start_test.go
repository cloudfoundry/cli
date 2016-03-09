package application_test

import (
	"os"
	"time"

	. "github.com/cloudfoundry/cli/cf/commands/application"
	"github.com/cloudfoundry/cli/cf/trace/fakes"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/loggregatorlib/logmessage"

	testAppInstances "github.com/cloudfoundry/cli/cf/api/app_instances/fakes"
	testApplication "github.com/cloudfoundry/cli/cf/api/applications/fakes"
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	appCmdFakes "github.com/cloudfoundry/cli/cf/commands/application/fakes"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testlogs "github.com/cloudfoundry/cli/testhelpers/logs"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("start command", func() {
	var (
		ui                        *testterm.FakeUI
		configRepo                core_config.Repository
		defaultAppForStart        models.Application
		defaultInstanceResponses  [][]models.AppInstanceFields
		defaultInstanceErrorCodes []string
		requirementsFactory       *testreq.FakeReqFactory
		logMessages               []*logmessage.LogMessage
		logRepo                   *testapi.FakeLogsRepository
		appInstancesRepo          *testAppInstances.FakeAppInstancesRepository
		appRepo                   *testApplication.FakeApplicationRepository
		originalAppCommand        command_registry.Command
		deps                      command_registry.Dependency
		displayApp                *appCmdFakes.FakeAppDisplayer
	)

	updateCommandDependency := func(logsRepo api.LogsRepository) {
		deps.Ui = ui
		deps.Config = configRepo
		deps.RepoLocator = deps.RepoLocator.SetLogsRepository(logsRepo)
		deps.RepoLocator = deps.RepoLocator.SetApplicationRepository(appRepo)
		deps.RepoLocator = deps.RepoLocator.SetAppInstancesRepository(appInstancesRepo)

		//inject fake 'Start' into registry
		command_registry.Register(displayApp)

		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("start").SetDependency(deps, false))
	}

	getInstance := func(appGuid string) ([]models.AppInstanceFields, error) {
		var apiErr error
		var instances []models.AppInstanceFields

		if len(defaultInstanceResponses) > 0 {
			instances, defaultInstanceResponses = defaultInstanceResponses[0], defaultInstanceResponses[1:]
		}
		if len(defaultInstanceErrorCodes) > 0 {
			var errorCode string
			errorCode, defaultInstanceErrorCodes = defaultInstanceErrorCodes[0], defaultInstanceErrorCodes[1:]

			if errorCode != "" {
				apiErr = errors.NewHttpError(400, errorCode, "Error staging app")
			}
		}

		return instances, apiErr
	}

	AfterEach(func() {
		command_registry.Register(originalAppCommand)
	})

	BeforeEach(func() {
		deps = command_registry.NewDependency(new(fakes.FakePrinter))
		ui = new(testterm.FakeUI)
		requirementsFactory = &testreq.FakeReqFactory{}

		configRepo = testconfig.NewRepository()

		appInstancesRepo = &testAppInstances.FakeAppInstancesRepository{}
		appRepo = &testApplication.FakeApplicationRepository{}

		displayApp = &appCmdFakes.FakeAppDisplayer{}

		//save original command dependency and restore later
		originalAppCommand = command_registry.Commands.FindCommand("app")

		defaultInstanceErrorCodes = []string{"", ""}

		defaultAppForStart = models.Application{
			ApplicationFields: models.ApplicationFields{
				Name:          "my-app",
				Guid:          "my-app-guid",
				InstanceCount: 2,
				PackageState:  "STAGED",
			},
		}

		defaultAppForStart.Routes = []models.RouteSummary{
			models.RouteSummary{
				Host: "my-app",
				Domain: models.DomainFields{
					Name: "example.com",
				},
			},
		}

		instance1 := models.AppInstanceFields{
			State: models.InstanceStarting,
		}

		instance2 := models.AppInstanceFields{
			State: models.InstanceStarting,
		}

		instance3 := models.AppInstanceFields{
			State: models.InstanceRunning,
		}

		instance4 := models.AppInstanceFields{
			State: models.InstanceStarting,
		}

		defaultInstanceResponses = [][]models.AppInstanceFields{
			[]models.AppInstanceFields{instance1, instance2},
			[]models.AppInstanceFields{instance1, instance2},
			[]models.AppInstanceFields{instance3, instance4},
		}

		logRepo = &testapi.FakeLogsRepository{}
		logMessages = []*logmessage.LogMessage{}
		logRepo.TailLogsForStub = func(appGuid string, onConnect func()) (<-chan *logmessage.LogMessage, error) {
			c := make(chan *logmessage.LogMessage)

			onConnect()
			go func() {
				for _, log := range logMessages {
					c <- log
				}

				close(c)
			}()

			return c, nil
		}

	})

	callStart := func(args []string) bool {
		updateCommandDependency(logRepo)
		cmd := command_registry.Commands.FindCommand("start").(*Start)
		cmd.StagingTimeout = 100 * time.Millisecond
		cmd.StartupTimeout = 500 * time.Millisecond
		cmd.PingerThrottle = 10 * time.Millisecond
		command_registry.Register(cmd)
		return testcmd.RunCliCommandWithoutDependency("start", args, requirementsFactory)
	}

	callStartWithLoggingTimeout := func(args []string) (ui *testterm.FakeUI) {

		logRepoWithTimeout := &testapi.FakeLogsRepositoryWithTimeout{}

		updateCommandDependency(logRepoWithTimeout)

		cmd := command_registry.Commands.FindCommand("start").(*Start)
		cmd.LogServerConnectionTimeout = 100 * time.Millisecond
		cmd.StagingTimeout = 100 * time.Millisecond
		cmd.StartupTimeout = 200 * time.Millisecond
		cmd.PingerThrottle = 10 * time.Millisecond
		command_registry.Register(cmd)

		testcmd.RunCliCommandWithoutDependency("start", args, requirementsFactory)
		return
	}

	startAppWithInstancesAndErrors := func(app models.Application, requirementsFactory *testreq.FakeReqFactory) (*testterm.FakeUI, *testApplication.FakeApplicationRepository, *testAppInstances.FakeAppInstancesRepository) {
		appRepo.UpdateReturns(app, nil)
		appRepo.ReadReturns(app, nil)
		appRepo.GetAppReturns(app, nil)
		appInstancesRepo.GetInstancesStub = getInstance

		args := []string{"my-app"}

		requirementsFactory.Application = app
		callStart(args)
		return ui, appRepo, appInstancesRepo
	}

	It("fails requirements when not logged in", func() {
		requirementsFactory.LoginSuccess = false

		Expect(callStart([]string{"some-app-name"})).To(BeFalse())
	})

	It("fails requirements when a space is not targeted", func() {
		requirementsFactory.LoginSuccess = true
		requirementsFactory.TargetedSpaceSuccess = false

		Expect(callStart([]string{"some-app-name"})).To(BeFalse())
	})

	Describe("timeouts", func() {
		It("has sane default timeout values", func() {
			updateCommandDependency(logRepo)
			cmd := command_registry.Commands.FindCommand("start").(*Start)
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

			updateCommandDependency(logRepo)
			cmd := command_registry.Commands.FindCommand("start").(*Start)
			Expect(cmd.StagingTimeout).To(Equal(6 * time.Minute))
			Expect(cmd.StartupTimeout).To(Equal(3 * time.Minute))
		})

		Describe("when the staging timeout is zero seconds", func() {
			var (
				app models.Application
			)

			BeforeEach(func() {
				app = defaultAppForStart

				appRepo.UpdateReturns(app, nil)

				app.PackageState = "FAILED"
				app.StagingFailedReason = "BLAH, FAILED"
				appRepo.GetAppReturns(app, nil)

				requirementsFactory.LoginSuccess = true
				requirementsFactory.TargetedSpaceSuccess = true
				requirementsFactory.Application = app

				updateCommandDependency(logRepo)
				cmd := command_registry.Commands.FindCommand("start").(*Start)
				cmd.StagingTimeout = 0
				cmd.PingerThrottle = 1
				cmd.StartupTimeout = 1
				command_registry.Register(cmd)
			})

			It("can still respond to staging failures", func() {
				testcmd.RunCliCommandWithoutDependency("start", []string{"my-app"}, requirementsFactory)

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"my-app"},
					[]string{"FAILED"},
					[]string{"BLAH, FAILED"},
				))
			})
		})
	})

	Context("when logged in", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedSpaceSuccess = true
			configRepo = testconfig.NewRepositoryWithDefaults()
		})

		It("fails with usage when not provided exactly one arg", func() {
			callStart([]string{})
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires an argument"},
			))
		})

		It("uses proper org name and space name", func() {
			appRepo.ReadReturns(defaultAppForStart, nil)
			appRepo.GetAppReturns(defaultAppForStart, nil)
			appInstancesRepo.GetInstancesStub = getInstance

			updateCommandDependency(logRepo)
			cmd := command_registry.Commands.FindCommand("start").(*Start)
			cmd.PingerThrottle = 10 * time.Millisecond

			//defaultAppForStart.State = "started"
			cmd.ApplicationStart(defaultAppForStart, "some-org", "some-space")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"my-app", "some-org", "some-space", "my-user"},
				[]string{"OK"},
			))
		})

		It("starts an app, when given the app's name", func() {
			ui, appRepo, _ := startAppWithInstancesAndErrors(defaultAppForStart, requirementsFactory)

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"my-app", "my-org", "my-space", "my-user"},
				[]string{"OK"},
				[]string{"0 of 2 instances running", "2 starting"},
				[]string{"started"},
			))

			Expect(requirementsFactory.ApplicationName).To(Equal("my-app"))
			appGUID, _ := appRepo.UpdateArgsForCall(0)
			Expect(appGUID).To(Equal("my-app-guid"))
			Expect(displayApp.AppToDisplay).To(Equal(defaultAppForStart))
		})

		It("displays the command start command instead of the detected start command when set", func() {
			defaultAppForStart.Command = "command start command"
			defaultAppForStart.DetectedStartCommand = "detected start command"
			ui, appRepo, _ = startAppWithInstancesAndErrors(defaultAppForStart, requirementsFactory)

			Expect(appRepo.GetAppCallCount()).To(Equal(1))
			Expect(appRepo.GetAppArgsForCall(0)).To(Equal("my-app-guid"))
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"App my-app was started using this command `command start command`"},
			))
		})

		It("displays the detected start command when no other command is set", func() {
			defaultAppForStart.DetectedStartCommand = "detected start command"
			defaultAppForStart.Command = ""
			ui, appRepo, _ := startAppWithInstancesAndErrors(defaultAppForStart, requirementsFactory)

			Expect(appRepo.GetAppCallCount()).To(Equal(1))
			Expect(appRepo.GetAppArgsForCall(0)).To(Equal("my-app-guid"))
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"App my-app was started using this command `detected start command`"},
			))
		})

		It("handles timeouts gracefully", func() {
			requirementsFactory.Application = defaultAppForStart
			appRepo.UpdateReturns(defaultAppForStart, nil)
			appRepo.ReadReturns(defaultAppForStart, nil)

			callStartWithLoggingTimeout([]string{"my-app"})
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"timeout connecting to log server"},
			))
		})

		It("only displays staging logs when an app is starting", func() {
			requirementsFactory.Application = defaultAppForStart
			appRepo.UpdateReturns(defaultAppForStart, nil)
			appRepo.ReadReturns(defaultAppForStart, nil)

			currentTime := time.Now()
			wrongSourceName := "DEA"
			correctSourceName := "STG"

			logMessages = []*logmessage.LogMessage{
				testlogs.NewLogMessage("Log Line 1", defaultAppForStart.Guid, wrongSourceName, "1", logmessage.LogMessage_OUT, currentTime),
				testlogs.NewLogMessage("Log Line 2", defaultAppForStart.Guid, correctSourceName, "1", logmessage.LogMessage_OUT, currentTime),
				testlogs.NewLogMessage("Log Line 3", defaultAppForStart.Guid, correctSourceName, "1", logmessage.LogMessage_OUT, currentTime),
				testlogs.NewLogMessage("Log Line 4", defaultAppForStart.Guid, wrongSourceName, "1", logmessage.LogMessage_OUT, currentTime),
			}

			callStart([]string{"my-app"})

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
			logRepoClosed := make(chan struct{})

			logRepo.TailLogsForStub = func(appGuid string, onConnect func()) (<-chan *logmessage.LogMessage, error) {
				c := make(chan *logmessage.LogMessage)
				onConnect()

				go func() {
					c <- testlogs.NewLogMessage("Before close", appGuid, LogMessageTypeStaging, "1", logmessage.LogMessage_ERR, time.Now())

					<-logRepoClosed

					time.Sleep(50 * time.Millisecond)
					c <- testlogs.NewLogMessage("After close 1", appGuid, LogMessageTypeStaging, "1", logmessage.LogMessage_ERR, time.Now())
					c <- testlogs.NewLogMessage("After close 2", appGuid, LogMessageTypeStaging, "1", logmessage.LogMessage_ERR, time.Now())

					close(c)
				}()

				return c, nil
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

			defaultInstanceErrorCodes = []string{errors.NotStaged, errors.NotStaged, "", "", ""}
			defaultAppForStart.PackageState = "PENDING"
			ui, appRepo, _ := startAppWithInstancesAndErrors(defaultAppForStart, requirementsFactory)

			Expect(appRepo.GetAppArgsForCall(0)).To(Equal("my-app-guid"))

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Before close"},
				[]string{"After close 1"},
				[]string{"After close 2"},
				[]string{"my-app failed to stage within", "minutes"},
			))
		})

		It("displays an error message when staging fails", func() {
			defaultAppForStart.PackageState = "FAILED"
			defaultAppForStart.StagingFailedReason = "AWWW, FAILED"

			ui, _, _ := startAppWithInstancesAndErrors(defaultAppForStart, requirementsFactory)

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"my-app"},
				[]string{"FAILED"},
				[]string{"AWWW, FAILED"},
			))
		})

		It("displays an TIP about needing to push from source directory when staging fails with NoAppDetectedError", func() {
			defaultAppForStart.PackageState = "FAILED"
			defaultAppForStart.StagingFailedReason = "NoAppDetectedError"

			ui, _, _ := startAppWithInstancesAndErrors(defaultAppForStart, requirementsFactory)

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"my-app"},
				[]string{"FAILED"},
				[]string{"is executed from within the directory"},
			))
		})

		It("Display a TIP when starting the app timeout", func() {
			appInstance := models.AppInstanceFields{}
			appInstance.State = models.InstanceStarting
			appInstance2 := models.AppInstanceFields{}
			appInstance2.State = models.InstanceStarting
			appInstance3 := models.AppInstanceFields{}
			appInstance3.State = models.InstanceStarting
			appInstance4 := models.AppInstanceFields{}
			appInstance4.State = models.InstanceStarting
			defaultInstanceResponses = [][]models.AppInstanceFields{
				[]models.AppInstanceFields{appInstance, appInstance2},
				[]models.AppInstanceFields{appInstance3, appInstance4},
			}

			defaultInstanceErrorCodes = []string{"some error", ""}

			ui, _, _ := startAppWithInstancesAndErrors(defaultAppForStart, requirementsFactory)

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"TIP: Application must be listening on the right port."},
			))
		})

		It("prints a warning when failing to fetch instance count", func() {
			defaultInstanceResponses = [][]models.AppInstanceFields{}
			defaultInstanceErrorCodes = []string{"an-error"}

			ui, _, _ := startAppWithInstancesAndErrors(defaultAppForStart, requirementsFactory)

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"an-error"},
			))
		})

		Context("when an app instance is flapping", func() {
			It("fails and alerts the user", func() {
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

				ui, _, _ := startAppWithInstancesAndErrors(defaultAppForStart, requirementsFactory)

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"my-app"},
					[]string{"0 of 2 instances running", "1 starting", "1 failing"},
					[]string{"FAILED"},
					[]string{"Start unsuccessful"},
				))
			})
		})

		Context("when an app instance is crashed", func() {
			It("fails and alerts the user", func() {
				appInstance := models.AppInstanceFields{}
				appInstance.State = models.InstanceStarting
				appInstance2 := models.AppInstanceFields{}
				appInstance2.State = models.InstanceStarting
				appInstance3 := models.AppInstanceFields{}
				appInstance3.State = models.InstanceStarting
				appInstance4 := models.AppInstanceFields{}
				appInstance4.State = models.InstanceCrashed
				defaultInstanceResponses = [][]models.AppInstanceFields{
					[]models.AppInstanceFields{appInstance, appInstance2},
					[]models.AppInstanceFields{appInstance3, appInstance4},
				}

				defaultInstanceErrorCodes = []string{"", ""}

				ui, _, _ := startAppWithInstancesAndErrors(defaultAppForStart, requirementsFactory)

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"my-app"},
					[]string{"0 of 2 instances running", "1 starting", "1 crashed"},
					[]string{"FAILED"},
					[]string{"Start unsuccessful"},
				))
			})
		})

		Context("when an app instance is starting", func() {
			It("reports any additional details", func() {
				appInstance := models.AppInstanceFields{
					State: models.InstanceStarting,
				}
				appInstance2 := models.AppInstanceFields{
					State: models.InstanceStarting,
				}

				appInstance3 := models.AppInstanceFields{
					State: models.InstanceDown,
				}
				appInstance4 := models.AppInstanceFields{
					State:   models.InstanceStarting,
					Details: "no compatible cell",
				}

				appInstance5 := models.AppInstanceFields{
					State:   models.InstanceStarting,
					Details: "insufficient resources",
				}
				appInstance6 := models.AppInstanceFields{
					State:   models.InstanceStarting,
					Details: "no compatible cell",
				}

				appInstance7 := models.AppInstanceFields{
					State: models.InstanceRunning,
				}
				appInstance8 := models.AppInstanceFields{
					State: models.InstanceRunning,
				}

				defaultInstanceResponses = [][]models.AppInstanceFields{
					[]models.AppInstanceFields{appInstance, appInstance2},
					[]models.AppInstanceFields{appInstance3, appInstance4},
					[]models.AppInstanceFields{appInstance5, appInstance6},
					[]models.AppInstanceFields{appInstance7, appInstance8},
				}

				defaultInstanceErrorCodes = []string{"", ""}

				ui, _, _ := startAppWithInstancesAndErrors(defaultAppForStart, requirementsFactory)
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"my-app"},
					[]string{"0 of 2 instances running", "2 starting"},
					[]string{"0 of 2 instances running", "1 starting (no compatible cell)", "1 down"},
					[]string{"0 of 2 instances running", "2 starting (insufficient resources, no compatible cell)"},
					[]string{"2 of 2 instances running"},
					[]string{"App started"},
				))
			})
		})

		It("tells the user about the failure when waiting for the app to stage times out", func() {
			defaultInstanceErrorCodes = []string{errors.NotStaged, errors.NotStaged, errors.NotStaged}

			defaultAppForStart.PackageState = "PENDING"
			ui, _, _ := startAppWithInstancesAndErrors(defaultAppForStart, requirementsFactory)

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Starting", "my-app"},
				[]string{"FAILED"},
				[]string{"my-app failed to stage within", "minutes"},
			))
			Expect(ui.Outputs).ToNot(ContainSubstrings([]string{"instances running"}))
		})

		It("tells the user about the failure when starting the app fails", func() {
			app := models.Application{}
			app.Name = "my-app"
			app.Guid = "my-app-guid"
			appRepo.UpdateReturns(models.Application{}, errors.New("Error updating app."))
			appRepo.ReadReturns(app, nil)
			args := []string{"my-app"}
			requirementsFactory.Application = app
			callStart(args)

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"my-app"},
				[]string{"FAILED"},
				[]string{"Error updating app."},
			))
			appGUID, _ := appRepo.UpdateArgsForCall(0)
			Expect(appGUID).To(Equal("my-app-guid"))
		})

		It("warns the user when the app is already running", func() {
			app := models.Application{}
			app.Name = "my-app"
			app.Guid = "my-app-guid"
			app.State = "started"
			appRepo := &testApplication.FakeApplicationRepository{}
			appRepo.ReadReturns(app, nil)

			requirementsFactory.Application = app

			args := []string{"my-app"}
			callStart(args)

			Expect(ui.Outputs).To(ContainSubstrings([]string{"my-app", "is already started"}))

			Expect(appRepo.UpdateCallCount()).To(BeZero())
		})

		It("tells the user when connecting to the log server fails", func() {
			appRepo.ReadReturns(defaultAppForStart, nil)

			logRepo.TailLogsForReturns(nil, errors.New("Ooops"))

			requirementsFactory.Application = defaultAppForStart

			callStart([]string{"my-app"})

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"error tailing logs"},
				[]string{"Ooops"},
			))
		})
	})
})
