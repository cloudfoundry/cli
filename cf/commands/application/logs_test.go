package application_test

import (
	"time"

	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/terminal"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testlogs "github.com/cloudfoundry/cli/testhelpers/logs"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	"github.com/cloudfoundry/loggregatorlib/logmessage"

	. "github.com/cloudfoundry/cli/cf/commands/application"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("logs command", func() {
	var (
		ui                  *testterm.FakeUI
		logsRepo            *testapi.FakeLogsRepository
		requirementsFactory *testreq.FakeReqFactory
		configRepo          core_config.Repository
		deps                command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.RepoLocator = deps.RepoLocator.SetLogsRepository(logsRepo)
		deps.Config = configRepo
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("logs").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		logsRepo = &testapi.FakeLogsRepository{}
		requirementsFactory = &testreq.FakeReqFactory{}
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCliCommand("logs", args, requirementsFactory, updateCommandDependency, false)
	}

	Describe("requirements", func() {
		It("fails with usage when called without one argument", func() {
			requirementsFactory.LoginSuccess = true

			runCommand()
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires an argument"},
			))
		})

		It("fails requirements when not logged in", func() {
			Expect(runCommand("my-app")).To(BeFalse())
		})
		It("fails if a space is not targeted", func() {
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedSpaceSuccess = false
			Expect(runCommand("--recent", "my-app")).To(BeFalse())
		})

	})

	Context("when logged in", func() {
		var (
			app models.Application
		)

		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedSpaceSuccess = true

			app = models.Application{}
			app.Name = "my-app"
			app.Guid = "my-app-guid"

			currentTime := time.Now()
			recentLogs := []*logmessage.LogMessage{
				testlogs.NewLogMessage("Log Line 1", app.Guid, "DEA", "1", logmessage.LogMessage_ERR, currentTime),
				testlogs.NewLogMessage("Log Line 2", app.Guid, "DEA", "1", logmessage.LogMessage_ERR, currentTime),
			}

			appLogs := []*logmessage.LogMessage{
				testlogs.NewLogMessage("Log Line 1", app.Guid, "DEA", "1", logmessage.LogMessage_ERR, time.Now()),
			}

			requirementsFactory.Application = app
			logsRepo.RecentLogsForReturns(recentLogs, nil)
			logsRepo.TailLogsForStub = func(appGuid string, onConnect func(), onMessage func(*logmessage.LogMessage)) error {
				onConnect()
				for _, log := range appLogs {
					onMessage(log)
				}
				return nil
			}

		})

		It("shows the recent logs when the --recent flag is provided", func() {
			runCommand("--recent", "my-app")

			Expect(requirementsFactory.ApplicationName).To(Equal("my-app"))
			Expect(app.Guid).To(Equal(logsRepo.RecentLogsForArgsForCall(0)))
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Connected, dumping recent logs for app", "my-app", "my-org", "my-space", "my-user"},
				[]string{"Log Line 1"},
				[]string{"Log Line 2"},
			))
		})

		Context("when the log messages contain format string identifiers", func() {
			BeforeEach(func() {
				logsRepo.RecentLogsForReturns([]*logmessage.LogMessage{
					testlogs.NewLogMessage("hello%2Bworld%v", app.Guid, "DEA", "1", logmessage.LogMessage_ERR, time.Now()),
				}, nil)
			})

			It("does not treat them as format strings", func() {
				runCommand("--recent", "my-app")
				Expect(ui.Outputs).To(ContainSubstrings([]string{"hello%2Bworld%v"}))
			})
		})

		It("tails the app's logs when no flags are given", func() {
			runCommand("my-app")

			Expect(requirementsFactory.ApplicationName).To(Equal("my-app"))
			appGuid, _, _ := logsRepo.TailLogsForArgsForCall(0)
			Expect(app.Guid).To(Equal(appGuid))
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Connected, tailing logs for app", "my-app", "my-org", "my-space", "my-user"},
				[]string{"Log Line 1"},
			))
		})

		Context("when the loggregator server has an invalid cert", func() {
			Context("when the skip-ssl-validation flag is not set", func() {
				It("fails and informs the user about the skip-ssl-validation flag", func() {
					logsRepo.TailLogsForReturns(errors.NewInvalidSSLCert("https://example.com", "it don't work good"))
					runCommand("my-app")

					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Received invalid SSL certificate", "https://example.com"},
						[]string{"TIP"},
					))
				})

				It("informs the user of the error when they include the --recent flag", func() {
					logsRepo.RecentLogsForReturns(nil, errors.NewInvalidSSLCert("https://example.com", "how does SSL work???"))
					runCommand("--recent", "my-app")

					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Received invalid SSL certificate", "https://example.com"},
						[]string{"TIP"},
					))
				})
			})
		})

		Context("when the loggregator server has a valid cert", func() {
			It("tails logs", func() {
				runCommand("my-app")
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Connected, tailing logs for app", "my-org", "my-space", "my-user"},
				))
			})
		})

		Describe("Helpers", func() {
			var date time.Time

			BeforeEach(func() {
				date = time.Date(2014, 4, 4, 11, 39, 20, 5, time.UTC)
			})

			Context("when the message comes", func() {
				It("include the instance index", func() {
					msg := testlogs.NewLogMessage("Hello World!", app.Guid, "DEA", "4", logmessage.LogMessage_OUT, date)
					Expect(terminal.Decolorize(LogMessageOutput(msg, time.UTC))).To(Equal("2014-04-04T11:39:20.00+0000 [DEA/4]      OUT Hello World!"))
				})

				It("doesn't include the instance index if sourceID is empty", func() {
					msg := testlogs.NewLogMessage("Hello World!", app.Guid, "DEA", "", logmessage.LogMessage_OUT, date)
					Expect(terminal.Decolorize(LogMessageOutput(msg, time.UTC))).To(Equal("2014-04-04T11:39:20.00+0000 [DEA]        OUT Hello World!"))
				})
			})

			Context("when the message was written to stderr", func() {
				It("shows the log type as 'ERR'", func() {
					msg := testlogs.NewLogMessage("Hello World!", app.Guid, "DEA", "4", logmessage.LogMessage_ERR, date)
					Expect(terminal.Decolorize(LogMessageOutput(msg, time.UTC))).To(Equal("2014-04-04T11:39:20.00+0000 [DEA/4]      ERR Hello World!"))
				})
			})

			It("formats the time in the given time zone", func() {
				msg := testlogs.NewLogMessage("Hello World!", app.Guid, "DEA", "4", logmessage.LogMessage_ERR, date)
				Expect(terminal.Decolorize(LogMessageOutput(msg, time.FixedZone("the-zone", 3*60*60)))).To(Equal("2014-04-04T14:39:20.00+0300 [DEA/4]      ERR Hello World!"))
			})
		})
	})
})
