/*
                       WARNING WARNING WARNING

                Attention all potential contributors

   This testfile is not in the best state. We've been slowly transitioning
   from the built in "testing" package to using Ginkgo. As you can see, we've
   changed the format, but a lot of the setup, test body, descriptions, etc
   are either hardcoded, completely lacking, or misleading.

   For example:

   Describe("Testing with ginkgo"...)      // This is not a great description
   It("TestDoesSoemthing"...)              // This is a horrible description

   Describe("create-user command"...       // Describe the actual object under test
   It("creates a user when provided ..."   // this is more descriptive

   For good examples of writing Ginkgo tests for the cli, refer to

   src/cf/commands/application/delete_app_test.go
   src/cf/terminal/ui_test.go
   src/github.com/cloudfoundry/loggregator_consumer/consumer_test.go
*/

package application_test

import (
	"cf/api"
	. "cf/commands/application"
	"cf/configuration"
	"cf/errors"
	"cf/models"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"os"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"time"
)

var _ = Describe("start command", func() {
	var (
		defaultAppForStart        = models.Application{}
		defaultInstanceReponses   = [][]models.AppInstanceFields{}
		defaultInstanceErrorCodes = []string{"", ""}
		requirementsFactory       *testreq.FakeReqFactory
	)

	BeforeEach(func() {
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

		defaultInstanceReponses = [][]models.AppInstanceFields{
			[]models.AppInstanceFields{instance1, instance2},
			[]models.AppInstanceFields{instance1, instance2},
			[]models.AppInstanceFields{instance3, instance4},
		}
	})

	It("has sane default timeout values", func() {
		cmd := NewStart(new(testterm.FakeUI), testconfig.NewRepository(), &testcmd.FakeAppDisplayer{}, &testapi.FakeApplicationRepository{}, &testapi.FakeAppInstancesRepo{}, &testapi.FakeLogsRepository{})
		Expect(cmd.StagingTimeout).To(Equal(15 * time.Minute))
		Expect(cmd.StartupTimeout).To(Equal(5 * time.Minute))
	})

	It("fails requirements when not logged in", func() {
		requirementsFactory.LoginSuccess = false
		cmd := NewStart(new(testterm.FakeUI), testconfig.NewRepository(), &testcmd.FakeAppDisplayer{}, &testapi.FakeApplicationRepository{}, &testapi.FakeAppInstancesRepo{}, &testapi.FakeLogsRepository{})
		testcmd.RunCommand(cmd, testcmd.NewContext("start", []string{"some-app-name"}), requirementsFactory)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
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
		cmd := NewStart(new(testterm.FakeUI), testconfig.NewRepository(), &testcmd.FakeAppDisplayer{}, &testapi.FakeApplicationRepository{}, &testapi.FakeAppInstancesRepo{}, &testapi.FakeLogsRepository{})
		Expect(cmd.StagingTimeout).To(Equal(6 * time.Minute))
		Expect(cmd.StartupTimeout).To(Equal(3 * time.Minute))
	})

	Context("when logged in", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
		})

		It("fails with usage when provided with no args", func() {
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

			ui := callStart([]string{}, config, requirementsFactory, displayApp, appRepo, appInstancesRepo, logRepo)
			Expect(ui.FailedWithUsage).To(BeTrue())
		})

		It("starts an app, when given the app's name", func() {
			displayApp := &testcmd.FakeAppDisplayer{}
			ui, appRepo, _ := startAppWithInstancesAndErrors(displayApp, defaultAppForStart, defaultInstanceReponses, defaultInstanceErrorCodes, requirementsFactory)

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"my-app", "my-org", "my-space", "my-user"},
				{"OK"},
				{"0 of 2 instances running", "2 starting"},
				{"Started"},
			})

			Expect(requirementsFactory.ApplicationName).To(Equal("my-app"))
			Expect(appRepo.UpdateAppGuid).To(Equal("my-app-guid"))
			Expect(displayApp.AppToDisplay).To(Equal(defaultAppForStart))
		})

		It("only displays staging logs when an app is starting", func() {
			displayApp := &testcmd.FakeAppDisplayer{}
			requirementsFactory.Application = defaultAppForStart
			appRepo := &testapi.FakeApplicationRepository{
				UpdateAppResult: defaultAppForStart,
			}
			appRepo.ReadReturns.App = defaultAppForStart

			appInstancesRepo := &testapi.FakeAppInstancesRepo{
				GetInstancesResponses:  defaultInstanceReponses,
				GetInstancesErrorCodes: defaultInstanceErrorCodes,
			}

			currentTime := time.Now()
			wrongSourceName := "DEA"
			correctSourceName := "STG"

			logRepo := &testapi.FakeLogsRepository{
				TailLogMessages: []*logmessage.LogMessage{
					NewLogMessage("Log Line 1", defaultAppForStart.Guid, wrongSourceName, currentTime),
					NewLogMessage("Log Line 2", defaultAppForStart.Guid, correctSourceName, currentTime),
					NewLogMessage("Log Line 3", defaultAppForStart.Guid, correctSourceName, currentTime),
					NewLogMessage("Log Line 4", defaultAppForStart.Guid, wrongSourceName, currentTime),
				},
			}

			ui := callStart([]string{"my-app"}, testconfig.NewRepository(), requirementsFactory, displayApp, appRepo, appInstancesRepo, logRepo)

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"Log Line 2"},
				{"Log Line 3"},
			})
			testassert.SliceDoesNotContain(ui.Outputs, testassert.Lines{
				{"Log Line 1"},
				{"Log Line 4"},
			})
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

			errorCodes := []string{errors.APP_NOT_STAGED, errors.APP_NOT_STAGED, "", "", ""}

			ui, _, appInstancesRepo := startAppWithInstancesAndErrors(displayApp, defaultAppForStart, instances, errorCodes, requirementsFactory)

			Expect(appInstancesRepo.GetInstancesAppGuid).To(Equal("my-app-guid"))

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"Log Line 1"},
				{"Log Line 2"},
				{"0 of 2 instances running", "2 starting"},
			})
		})

		It("displays an error message when staging fails", func() {
			displayApp := &testcmd.FakeAppDisplayer{}
			instances := [][]models.AppInstanceFields{[]models.AppInstanceFields{}}
			errorCodes := []string{"170001"}

			ui, _, _ := startAppWithInstancesAndErrors(displayApp, defaultAppForStart, instances, errorCodes, requirementsFactory)

			testassert.SliceContains(ui.Outputs, testassert.Lines{
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

			ui, _, _ := startAppWithInstancesAndErrors(displayApp, defaultAppForStart, instances, errorCodes, requirementsFactory)

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"my-app"},
				{"OK"},
				{"0 of 2 instances running", "1 starting", "1 failing"},
				{"FAILED"},
				{"Start unsuccessful"},
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
			instances := [][]models.AppInstanceFields{
				[]models.AppInstanceFields{appInstance, appInstance2},
				[]models.AppInstanceFields{appInstance3, appInstance4},
				[]models.AppInstanceFields{appInstance5, appInstance6},
			}

			errorCodes := []string{errors.APP_NOT_STAGED, errors.APP_NOT_STAGED, errors.APP_NOT_STAGED}

			ui, _, _ := startAppWithInstancesAndErrors(displayApp, defaultAppForStart, instances, errorCodes, requirementsFactory)

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"Starting", "my-app"},
				{"OK"},
				{"FAILED"},
				{"Start app timeout"},
			})
			testassert.SliceDoesNotContain(ui.Outputs, testassert.Lines{
				{"instances running"},
			})
		})

		It("tells the user about the failure when starting the app fails", func() {
			config := testconfig.NewRepository()
			displayApp := &testcmd.FakeAppDisplayer{}
			app := models.Application{}
			app.Name = "my-app"
			app.Guid = "my-app-guid"
			appRepo := &testapi.FakeApplicationRepository{UpdateErr: true}
			appRepo.ReadReturns.App = app
			appInstancesRepo := &testapi.FakeAppInstancesRepo{}
			logRepo := &testapi.FakeLogsRepository{}
			args := []string{"my-app"}
			requirementsFactory.Application = app
			ui := callStart(args, config, requirementsFactory, displayApp, appRepo, appInstancesRepo, logRepo)

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"my-app"},
				{"FAILED"},
				{"Error updating app."},
			})
			Expect(appRepo.UpdateAppGuid).To(Equal("my-app-guid"))
		})

		It("warns the user when the app is already running", func() {
			displayApp := &testcmd.FakeAppDisplayer{}
			config := testconfig.NewRepository()
			app := models.Application{}
			app.Name = "my-app"
			app.Guid = "my-app-guid"
			app.State = "started"
			appRepo := &testapi.FakeApplicationRepository{}
			appRepo.ReadReturns.App = app
			appInstancesRepo := &testapi.FakeAppInstancesRepo{}
			logRepo := &testapi.FakeLogsRepository{}

			requirementsFactory.Application = app

			args := []string{"my-app"}
			ui := callStart(args, config, requirementsFactory, displayApp, appRepo, appInstancesRepo, logRepo)

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"my-app", "is already started"},
			})

			Expect(appRepo.UpdateAppGuid).To(Equal(""))
		})

		It("tells the user when connecting to the log server fails", func() {
			configRepo := testconfig.NewRepositoryWithDefaults()
			displayApp := &testcmd.FakeAppDisplayer{}

			appRepo := &testapi.FakeApplicationRepository{}
			appRepo.ReadReturns.App = defaultAppForStart
			appInstancesRepo := &testapi.FakeAppInstancesRepo{
				GetInstancesResponses:  defaultInstanceReponses,
				GetInstancesErrorCodes: defaultInstanceErrorCodes,
			}

			logRepo := &testapi.FakeLogsRepository{
				TailLogErr: errors.New("Ooops"),
			}

			requirementsFactory.Application = defaultAppForStart

			ui := callStart([]string{"my-app"}, configRepo, requirementsFactory, displayApp, appRepo, appInstancesRepo, logRepo)

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				testassert.Line{"error tailing logs"},
				testassert.Line{"Ooops"},
			})
		})
	})
})

func callStart(args []string, config configuration.Reader, requirementsFactory *testreq.FakeReqFactory, displayApp ApplicationDisplayer, appRepo api.ApplicationRepository, appInstancesRepo api.AppInstancesRepository, logRepo api.LogsRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("start", args)

	cmd := NewStart(ui, config, displayApp, appRepo, appInstancesRepo, logRepo)
	cmd.StagingTimeout = 50 * time.Millisecond
	cmd.StartupTimeout = 50 * time.Millisecond
	cmd.PingerThrottle = 50 * time.Millisecond

	testcmd.RunCommand(cmd, ctxt, requirementsFactory)
	return
}

func startAppWithInstancesAndErrors(displayApp ApplicationDisplayer, app models.Application, instances [][]models.AppInstanceFields, errorCodes []string, requirementsFactory *testreq.FakeReqFactory) (ui *testterm.FakeUI, appRepo *testapi.FakeApplicationRepository, appInstancesRepo *testapi.FakeAppInstancesRepo) {
	configRepo := testconfig.NewRepositoryWithDefaults()
	appRepo = &testapi.FakeApplicationRepository{
		UpdateAppResult: app,
	}
	appRepo.ReadReturns.App = app
	appInstancesRepo = &testapi.FakeAppInstancesRepo{
		GetInstancesResponses:  instances,
		GetInstancesErrorCodes: errorCodes,
	}

	logRepo := &testapi.FakeLogsRepository{
		TailLogMessages: []*logmessage.LogMessage{
			NewLogMessage("Log Line 1", app.Guid, LogMessageTypeStaging, time.Now()),
			NewLogMessage("Log Line 2", app.Guid, LogMessageTypeStaging, time.Now()),
		},
	}

	args := []string{"my-app"}

	requirementsFactory.Application = app
	ui = callStart(args, configRepo, requirementsFactory, displayApp, appRepo, appInstancesRepo, logRepo)
	return
}
