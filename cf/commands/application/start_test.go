package application_test

import (
	"os"
	"time"

	. "code.cloudfoundry.org/cli/cf/commands/application"
	"code.cloudfoundry.org/cli/cf/commands/application/applicationfakes"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"
	"code.cloudfoundry.org/cli/cf/trace/tracefakes"

	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/models"

	"code.cloudfoundry.org/cli/cf/api/appinstances/appinstancesfakes"
	"code.cloudfoundry.org/cli/cf/api/applications/applicationsfakes"
	"code.cloudfoundry.org/cli/cf/api/logs"
	"code.cloudfoundry.org/cli/cf/api/logs/logsfakes"
	testcmd "code.cloudfoundry.org/cli/util/testhelpers/commands"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"

	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"

	"sync"

	"sync/atomic"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("start command", func() {
	var (
		ui                        *testterm.FakeUI
		configRepo                coreconfig.Repository
		defaultAppForStart        models.Application
		defaultInstanceResponses  [][]models.AppInstanceFields
		defaultInstanceErrorCodes []string
		requirementsFactory       *requirementsfakes.FakeFactory
		logMessages               atomic.Value
		logRepo                   *logsfakes.FakeRepository

		appInstancesRepo   *appinstancesfakes.FakeAppInstancesRepository
		appRepo            *applicationsfakes.FakeRepository
		originalAppCommand commandregistry.Command
		deps               commandregistry.Dependency
		displayApp         *applicationfakes.FakeAppDisplayer
	)

	updateCommandDependency := func(logsRepo logs.Repository) {
		deps.UI = ui
		deps.Config = configRepo
		deps.RepoLocator = deps.RepoLocator.SetLogsRepository(logsRepo)
		deps.RepoLocator = deps.RepoLocator.SetApplicationRepository(appRepo)
		deps.RepoLocator = deps.RepoLocator.SetAppInstancesRepository(appInstancesRepo)

		//inject fake 'Start' into registry
		commandregistry.Register(displayApp)

		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("start").SetDependency(deps, false))
	}

	getInstance := func(appGUID string) ([]models.AppInstanceFields, error) {
		var apiErr error
		var instances []models.AppInstanceFields

		if len(defaultInstanceResponses) > 0 {
			instances, defaultInstanceResponses = defaultInstanceResponses[0], defaultInstanceResponses[1:]
		}
		if len(defaultInstanceErrorCodes) > 0 {
			var errorCode string
			errorCode, defaultInstanceErrorCodes = defaultInstanceErrorCodes[0], defaultInstanceErrorCodes[1:]

			if errorCode != "" {
				apiErr = errors.NewHTTPError(400, errorCode, "Error staging app")
			}
		}

		return instances, apiErr
	}

	AfterEach(func() {
		commandregistry.Register(originalAppCommand)
	})

	BeforeEach(func() {
		deps = commandregistry.NewDependency(os.Stdout, new(tracefakes.FakePrinter), "")
		ui = new(testterm.FakeUI)
		requirementsFactory = new(requirementsfakes.FakeFactory)

		configRepo = testconfig.NewRepository()

		appInstancesRepo = new(appinstancesfakes.FakeAppInstancesRepository)
		appRepo = new(applicationsfakes.FakeRepository)

		displayApp = new(applicationfakes.FakeAppDisplayer)

		//save original command dependency and restore later
		originalAppCommand = commandregistry.Commands.FindCommand("app")

		defaultInstanceErrorCodes = []string{"", ""}

		defaultAppForStart = models.Application{
			ApplicationFields: models.ApplicationFields{
				Name:          "my-app",
				GUID:          "my-app-guid",
				InstanceCount: 2,
				PackageState:  "STAGED",
			},
		}

		defaultAppForStart.Routes = []models.RouteSummary{
			{
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
			{instance1, instance2},
			{instance1, instance2},
			{instance3, instance4},
		}

		logRepo = new(logsfakes.FakeRepository)
		logMessages.Store([]logs.Loggable{})

		closeWait := sync.WaitGroup{}
		closeWait.Add(1)

		logRepo.TailLogsForStub = func(appGUID string, onConnect func(), logChan chan<- logs.Loggable, errChan chan<- error) {
			onConnect()

			go func() {
				for _, log := range logMessages.Load().([]logs.Loggable) {
					logChan <- log
				}

				closeWait.Wait()
				close(logChan)
			}()
		}

		logRepo.CloseStub = func() {
			closeWait.Done()
		}
	})

	callStart := func(args []string) bool {
		updateCommandDependency(logRepo)
		cmd := commandregistry.Commands.FindCommand("start").(*Start)
		cmd.StagingTimeout = 100 * time.Millisecond
		cmd.StartupTimeout = 500 * time.Millisecond
		cmd.PingerThrottle = 10 * time.Millisecond
		commandregistry.Register(cmd)
		return testcmd.RunCLICommandWithoutDependency("start", args, requirementsFactory, ui)
	}

	callStartWithLoggingTimeout := func(args []string) bool {

		logRepoWithTimeout := logsfakes.FakeRepository{}
		updateCommandDependency(&logRepoWithTimeout)

		cmd := commandregistry.Commands.FindCommand("start").(*Start)
		cmd.LogServerConnectionTimeout = 100 * time.Millisecond
		cmd.StagingTimeout = 100 * time.Millisecond
		cmd.StartupTimeout = 200 * time.Millisecond
		cmd.PingerThrottle = 10 * time.Millisecond
		commandregistry.Register(cmd)

		return testcmd.RunCLICommandWithoutDependency("start", args, requirementsFactory, ui)
	}

	startAppWithInstancesAndErrors := func(app models.Application, requirementsFactory *requirementsfakes.FakeFactory) (*testterm.FakeUI, *applicationsfakes.FakeRepository, *appinstancesfakes.FakeAppInstancesRepository) {
		appRepo.UpdateReturns(app, nil)
		appRepo.ReadReturns(app, nil)
		appRepo.GetAppReturns(app, nil)
		appInstancesRepo.GetInstancesStub = getInstance

		args := []string{"my-app"}

		applicationReq := new(requirementsfakes.FakeApplicationRequirement)
		applicationReq.GetApplicationReturns(app)
		requirementsFactory.NewApplicationRequirementReturns(applicationReq)
		callStart(args)
		return ui, appRepo, appInstancesRepo
	}

	It("fails requirements when not logged in", func() {
		requirementsFactory.NewLoginRequirementReturns(requirements.Failing{Message: "not logged in"})

		Expect(callStart([]string{"some-app-name"})).To(BeFalse())
	})

	It("fails requirements when a space is not targeted", func() {
		requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
		requirementsFactory.NewTargetedSpaceRequirementReturns(requirements.Failing{Message: "not targeting space"})

		Expect(callStart([]string{"some-app-name"})).To(BeFalse())
	})

	Describe("timeouts", func() {
		It("has sane default timeout values", func() {
			updateCommandDependency(logRepo)
			cmd := commandregistry.Commands.FindCommand("start").(*Start)
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
			cmd := commandregistry.Commands.FindCommand("start").(*Start)
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

				requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
				requirementsFactory.NewTargetedSpaceRequirementReturns(requirements.Passing{})
				applicationReq := new(requirementsfakes.FakeApplicationRequirement)
				applicationReq.GetApplicationReturns(app)
				requirementsFactory.NewApplicationRequirementReturns(applicationReq)

				updateCommandDependency(logRepo)
				cmd := commandregistry.Commands.FindCommand("start").(*Start)
				cmd.StagingTimeout = 0
				cmd.PingerThrottle = 1
				cmd.StartupTimeout = 1
				commandregistry.Register(cmd)
			})

			It("can still respond to staging failures", func() {
				testcmd.RunCLICommandWithoutDependency("start", []string{"my-app"}, requirementsFactory, ui)

				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"my-app"},
					[]string{"FAILED"},
					[]string{"BLAH, FAILED"},
				))
			})
		})

		Context("when the timeout happens exactly when the connection is established", func() {
			var startWait *sync.WaitGroup

			BeforeEach(func() {
				requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
				requirementsFactory.NewTargetedSpaceRequirementReturns(requirements.Passing{})
				configRepo = testconfig.NewRepositoryWithDefaults()
				logRepo.TailLogsForStub = func(appGUID string, onConnect func(), logChan chan<- logs.Loggable, errChan chan<- error) {
					startWait.Wait()
					onConnect()
				}
			})

			It("times out gracefully", func() {
				updateCommandDependency(logRepo)
				cmd := commandregistry.Commands.FindCommand("start").(*Start)
				cmd.LogServerConnectionTimeout = 10 * time.Millisecond
				startWait = new(sync.WaitGroup)
				startWait.Add(1)
				doneWait := new(sync.WaitGroup)
				doneWait.Add(1)
				cmd.TailStagingLogs(defaultAppForStart, make(chan bool, 1), startWait, doneWait)
			})
		})

		Context("when the noaa library reconnects", func() {
			var app models.Application
			BeforeEach(func() {
				app = defaultAppForStart
				app.PackageState = "FAILED"
				app.StagingFailedReason = "BLAH, FAILED"

				requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
				requirementsFactory.NewTargetedSpaceRequirementReturns(requirements.Passing{})
				applicationReq := new(requirementsfakes.FakeApplicationRequirement)
				applicationReq.GetApplicationReturns(app)
				requirementsFactory.NewApplicationRequirementReturns(applicationReq)

				appRepo.GetAppReturns(app, nil)
				appRepo.UpdateReturns(app, nil)

				cmd := commandregistry.Commands.FindCommand("start").(*Start)
				cmd.StagingTimeout = 1
				cmd.PingerThrottle = 1
				cmd.StartupTimeout = 1
				commandregistry.Register(cmd)

				logRepo.TailLogsForStub = func(appGUID string, onConnect func(), logChan chan<- logs.Loggable, errChan chan<- error) {
					onConnect()
					onConnect()
					onConnect()
				}
				updateCommandDependency(logRepo)
			})

			It("it doesn't cause a negative wait group - github#1019", func() {
				testcmd.RunCLICommandWithoutDependency("start", []string{"my-app"}, requirementsFactory, ui)
			})
		})
	})

	Context("when logged in", func() {
		BeforeEach(func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
			requirementsFactory.NewTargetedSpaceRequirementReturns(requirements.Passing{})
			configRepo = testconfig.NewRepositoryWithDefaults()
		})

		It("fails with usage when not provided exactly one arg", func() {
			callStart([]string{})
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires an argument"},
			))
		})

		It("uses proper org name and space name", func() {
			appRepo.ReadReturns(defaultAppForStart, nil)
			appRepo.GetAppReturns(defaultAppForStart, nil)
			appInstancesRepo.GetInstancesStub = getInstance

			updateCommandDependency(logRepo)
			cmd := commandregistry.Commands.FindCommand("start").(*Start)
			cmd.PingerThrottle = 10 * time.Millisecond

			//defaultAppForStart.State = "started"
			cmd.ApplicationStart(defaultAppForStart, "some-org", "some-space")

			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"my-app", "some-org", "some-space", "my-user"},
				[]string{"OK"},
			))
		})

		It("starts an app, when given the app's name", func() {
			ui, appRepo, _ := startAppWithInstancesAndErrors(defaultAppForStart, requirementsFactory)

			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"my-app", "my-org", "my-space", "my-user"},
				[]string{"OK"},
				[]string{"0 of 2 instances running", "2 starting"},
				[]string{"started"},
			))

			appGUID, _ := appRepo.UpdateArgsForCall(0)
			Expect(appGUID).To(Equal("my-app-guid"))
			Expect(displayApp.AppToDisplay).To(Equal(defaultAppForStart))
		})

		Context("when app instance count is zero", func() {
			var zeroInstanceApp models.Application
			BeforeEach(func() {
				zeroInstanceApp = models.Application{
					ApplicationFields: models.ApplicationFields{
						Name:          "my-app",
						GUID:          "my-app-guid",
						InstanceCount: 0,
						PackageState:  "STAGED",
					},
				}
				defaultInstanceResponses = [][]models.AppInstanceFields{{}}
			})

			It("exit without polling for the app, and warns the user", func() {
				ui, _, _ := startAppWithInstancesAndErrors(zeroInstanceApp, requirementsFactory)

				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"my-app", "my-org", "my-space", "my-user"},
					[]string{"OK"},
					[]string{"App state changed to started, but note that it has 0 instances."},
				))

				Expect(appRepo.UpdateCallCount()).To(Equal(1))
				appGuid, appParams := appRepo.UpdateArgsForCall(0)
				Expect(appGuid).To(Equal(zeroInstanceApp.GUID))
				startedState := "started"
				Expect(appParams).To(Equal(models.AppParams{State: &startedState}))

				zeroInstanceApp.State = startedState
				ui, _, _ = startAppWithInstancesAndErrors(zeroInstanceApp, requirementsFactory)
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"App my-app is already started"},
				))
				Expect(appRepo.UpdateCallCount()).To(Equal(1))
			})
		})
		It("displays the command start command instead of the detected start command when set", func() {
			defaultAppForStart.Command = "command start command"
			defaultAppForStart.DetectedStartCommand = "detected start command"
			ui, appRepo, _ = startAppWithInstancesAndErrors(defaultAppForStart, requirementsFactory)

			Expect(appRepo.GetAppCallCount()).To(Equal(1))
			Expect(appRepo.GetAppArgsForCall(0)).To(Equal("my-app-guid"))
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"App my-app was started using this command `command start command`"},
			))
		})

		It("displays the detected start command when no other command is set", func() {
			defaultAppForStart.DetectedStartCommand = "detected start command"
			defaultAppForStart.Command = ""
			ui, appRepo, _ := startAppWithInstancesAndErrors(defaultAppForStart, requirementsFactory)

			Expect(appRepo.GetAppCallCount()).To(Equal(1))
			Expect(appRepo.GetAppArgsForCall(0)).To(Equal("my-app-guid"))
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"App my-app was started using this command `detected start command`"},
			))
		})

		It("handles timeouts gracefully", func() {
			applicationReq := new(requirementsfakes.FakeApplicationRequirement)
			applicationReq.GetApplicationReturns(defaultAppForStart)
			requirementsFactory.NewApplicationRequirementReturns(applicationReq)
			appRepo.UpdateReturns(defaultAppForStart, nil)
			appRepo.ReadReturns(defaultAppForStart, nil)

			callStartWithLoggingTimeout([]string{"my-app"})
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"timeout connecting to log server"},
			))
		})

		It("only displays staging logs when an app is starting", func() {
			applicationReq := new(requirementsfakes.FakeApplicationRequirement)
			applicationReq.GetApplicationReturns(defaultAppForStart)
			requirementsFactory.NewApplicationRequirementReturns(applicationReq)
			appRepo.UpdateReturns(defaultAppForStart, nil)
			appRepo.ReadReturns(defaultAppForStart, nil)

			message1 := logsfakes.FakeLoggable{}
			message1.ToSimpleLogReturns("Log Line 1")

			message2 := logsfakes.FakeLoggable{}
			message2.GetSourceNameReturns("STG")
			message2.ToSimpleLogReturns("Log Line 2")

			message3 := logsfakes.FakeLoggable{}
			message3.GetSourceNameReturns("STG")
			message3.ToSimpleLogReturns("Log Line 3")

			message4 := logsfakes.FakeLoggable{}
			message4.ToSimpleLogReturns("Log Line 4")

			logMessages.Store([]logs.Loggable{
				&message1,
				&message2,
				&message3,
				&message4,
			})

			callStart([]string{"my-app"})

			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Log Line 2"},
				[]string{"Log Line 3"},
			))
			Expect(ui.Outputs()).ToNot(ContainSubstrings(
				[]string{"Log Line 1"},
				[]string{"Log Line 4"},
			))
		})

		It("gracefully handles starting an app that is still staging", func() {
			closeWait := sync.WaitGroup{}
			closeWait.Add(1)

			logRepo.TailLogsForStub = func(appGUID string, onConnect func(), logChan chan<- logs.Loggable, errChan chan<- error) {
				onConnect()

				go func() {
					message1 := logsfakes.FakeLoggable{}
					message1.ToSimpleLogReturns("Before close")
					message1.GetSourceNameReturns("STG")

					logChan <- &message1

					closeWait.Wait()

					message2 := logsfakes.FakeLoggable{}
					message2.ToSimpleLogReturns("After close 1")
					message2.GetSourceNameReturns("STG")

					message3 := logsfakes.FakeLoggable{}
					message3.ToSimpleLogReturns("After close 2")
					message3.GetSourceNameReturns("STG")

					logChan <- &message2
					logChan <- &message3

					close(logChan)
				}()
			}

			logRepo.CloseStub = func() {
				closeWait.Done()
			}

			defaultInstanceResponses = [][]models.AppInstanceFields{
				{},
				{},
				{{State: models.InstanceDown}, {State: models.InstanceStarting}},
				{{State: models.InstanceStarting}, {State: models.InstanceStarting}},
				{{State: models.InstanceRunning}, {State: models.InstanceRunning}},
			}

			defaultInstanceErrorCodes = []string{errors.NotStaged, errors.NotStaged, "", "", ""}
			defaultAppForStart.PackageState = "PENDING"
			ui, appRepo, _ := startAppWithInstancesAndErrors(defaultAppForStart, requirementsFactory)

			Expect(appRepo.GetAppArgsForCall(0)).To(Equal("my-app-guid"))

			Expect(ui.Outputs()).To(ContainSubstrings(
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

			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"my-app"},
				[]string{"FAILED"},
				[]string{"AWWW, FAILED"},
			))
		})

		It("displays an TIP about needing to push from source directory when staging fails with NoAppDetectedError", func() {
			defaultAppForStart.PackageState = "FAILED"
			defaultAppForStart.StagingFailedReason = "NoAppDetectedError"

			ui, _, _ := startAppWithInstancesAndErrors(defaultAppForStart, requirementsFactory)

			Expect(ui.Outputs()).To(ContainSubstrings(
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
				{appInstance, appInstance2},
				{appInstance3, appInstance4},
			}

			defaultInstanceErrorCodes = []string{"some error", ""}

			ui, _, _ := startAppWithInstancesAndErrors(defaultAppForStart, requirementsFactory)

			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"TIP: Application must be listening on the right port."},
			))
		})

		It("prints a warning when failing to fetch instance count", func() {
			defaultInstanceResponses = [][]models.AppInstanceFields{}
			defaultInstanceErrorCodes = []string{"an-error"}

			ui, _, _ := startAppWithInstancesAndErrors(defaultAppForStart, requirementsFactory)

			Expect(ui.Outputs()).To(ContainSubstrings(
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
					{appInstance, appInstance2},
					{appInstance3, appInstance4},
				}

				defaultInstanceErrorCodes = []string{"", ""}

				ui, _, _ := startAppWithInstancesAndErrors(defaultAppForStart, requirementsFactory)

				Expect(ui.Outputs()).To(ContainSubstrings(
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
					{appInstance, appInstance2},
					{appInstance3, appInstance4},
				}

				defaultInstanceErrorCodes = []string{"", ""}

				ui, _, _ := startAppWithInstancesAndErrors(defaultAppForStart, requirementsFactory)

				Expect(ui.Outputs()).To(ContainSubstrings(
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
					{appInstance, appInstance2},
					{appInstance3, appInstance4},
					{appInstance5, appInstance6},
					{appInstance7, appInstance8},
				}

				defaultInstanceErrorCodes = []string{"", ""}

				ui, _, _ := startAppWithInstancesAndErrors(defaultAppForStart, requirementsFactory)
				Expect(ui.Outputs()).To(ContainSubstrings(
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

			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Starting", "my-app"},
				[]string{"FAILED"},
				[]string{"my-app failed to stage within", "minutes"},
			))
			Expect(ui.Outputs()).ToNot(ContainSubstrings([]string{"instances running"}))
		})

		It("tells the user about the failure when starting the app fails", func() {
			app := models.Application{}
			app.Name = "my-app"
			app.GUID = "my-app-guid"
			appRepo.UpdateReturns(models.Application{}, errors.New("Error updating app."))
			appRepo.ReadReturns(app, nil)
			args := []string{"my-app"}
			applicationReq := new(requirementsfakes.FakeApplicationRequirement)
			applicationReq.GetApplicationReturns(app)
			requirementsFactory.NewApplicationRequirementReturns(applicationReq)
			callStart(args)

			Expect(ui.Outputs()).To(ContainSubstrings(
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
			app.GUID = "my-app-guid"
			app.State = "started"
			appRepo := new(applicationsfakes.FakeRepository)
			appRepo.ReadReturns(app, nil)

			applicationReq := new(requirementsfakes.FakeApplicationRequirement)
			applicationReq.GetApplicationReturns(app)
			requirementsFactory.NewApplicationRequirementReturns(applicationReq)

			args := []string{"my-app"}
			callStart(args)

			Expect(ui.Outputs()).To(ContainSubstrings([]string{"my-app", "is already started"}))

			Expect(appRepo.UpdateCallCount()).To(BeZero())
		})

		It("tells the user when connecting to the log server fails", func() {
			appRepo.ReadReturns(defaultAppForStart, nil)

			logRepo.TailLogsForStub = func(appGUID string, onConnect func(), logChan chan<- logs.Loggable, errChan chan<- error) {
				errChan <- errors.New("Ooops")
			}

			applicationReq := new(requirementsfakes.FakeApplicationRequirement)
			applicationReq.GetApplicationReturns(defaultAppForStart)
			requirementsFactory.NewApplicationRequirementReturns(applicationReq)

			callStart([]string{"my-app"})

			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"error tailing logs"},
				[]string{"Ooops"},
			))
		})
	})
})
