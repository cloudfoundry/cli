package application_test

import (
	"strings"
	"time"

	"github.com/cloudfoundry/cli/cf/api/logs"
	"github.com/cloudfoundry/cli/cf/api/logs/logsfakes"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/requirements"
	"github.com/cloudfoundry/cli/cf/requirements/requirementsfakes"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testlogs "github.com/cloudfoundry/cli/testhelpers/logs"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	"github.com/cloudfoundry/loggregatorlib/logmessage"

	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("logs command", func() {
	var (
		ui                  *testterm.FakeUI
		logsRepo            *logsfakes.FakeRepository
		requirementsFactory *requirementsfakes.FakeFactory

		configRepo coreconfig.Repository
		deps       commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.RepoLocator = deps.RepoLocator.SetLogsRepository(logsRepo)
		deps.Config = configRepo
		commandregistry.Commands.SetCommand(
			commandregistry.Commands.FindCommand("logs").SetDependency(deps, pluginCall),
		)
	}

	runCommand := func(args ...string) bool {
		return testcmd.RunCLICommand("logs", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		logsRepo = &logsfakes.FakeRepository{}
		requirementsFactory = &requirementsfakes.FakeFactory{}
		configRepo = testconfig.NewRepositoryWithDefaults()
	})

	Describe("requirements", func() {
		It("fails with usage when called without one argument", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})

			runCommand()
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires an argument"},
			))
		})

		It("fails requirements when not logged in", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Failing{})

			Expect(runCommand("my-app")).To(BeFalse())
		})

		It("fails if a space is not targeted", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
			requirementsFactory.NewTargetedSpaceRequirementReturns(requirements.Failing{
				Message: "not targeting space",
			})
			Expect(runCommand("--recent", "my-app")).To(BeFalse())
		})

	})

	Context("when logged in", func() {
		var (
			app        models.Application
			recentLogs []logs.Loggable
			appLogs    []logs.Loggable
		)

		BeforeEach(func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
			requirementsFactory.NewTargetedSpaceRequirementReturns(requirements.Passing{})

			app = models.Application{}
			app.Name = "my-app"
			app.GUID = "my-app-guid"

			recentLogs = []logs.Loggable{
				testlogs.NewLogMessage("Log Line 1", app.GUID, "DEA", "1", logmessage.LogMessage_ERR, time.Now()),
				testlogs.NewLogMessage("Log Line 2", app.GUID, "DEA", "1", logmessage.LogMessage_ERR, time.Now()),
			}
			appLogs = []logs.Loggable{
				testlogs.NewLogMessage("Log Line 1", app.GUID, "DEA", "1", logmessage.LogMessage_ERR, time.Now()),
			}

			applicationReq := new(requirementsfakes.FakeApplicationRequirement)
			applicationReq.GetApplicationReturns(app)
			requirementsFactory.NewApplicationRequirementReturns(applicationReq)
		})

		JustBeforeEach(func() {
			logsRepo.RecentLogsForReturns(recentLogs, nil)
			logsRepo.TailLogsForStub = func(appGUID string, onConnect func(), logChan chan<- logs.Loggable, errChan chan<- error) {
				onConnect()
				go func() {
					for _, log := range appLogs {
						logChan <- log
					}
					close(logChan)
					close(errChan)
				}()
			}
		})

		It("shows the recent logs when the --recent flag is provided", func() {
			runCommand("--recent", "my-app")

			Expect(app.GUID).To(Equal(logsRepo.RecentLogsForArgsForCall(0)))
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Connected, dumping recent logs for app", "my-app", "my-org", "my-space", "my-user"},
				[]string{"Log Line 1"},
				[]string{"Log Line 2"},
			))
		})

		Context("with unicode runes present", func() {
			BeforeEach(func() {
				recentLogs = []logs.Loggable{
					testlogs.NewLogMessage("Unicode Line \u2713\u2028\u0000Log Line 2", app.GUID, "DEA", "1", logmessage.LogMessage_ERR, time.Now()),
				}
				appLogs = []logs.Loggable{
					testlogs.NewLogMessage("Unicode Line \u2713\u2028\u0000LLog Line 2", app.GUID, "DEA", "1", logmessage.LogMessage_ERR, time.Now()),
				}
			})

			It("replaces rune provided by --newline with \\n when the --recent flag is provided", func() {
				runCommand("--newline", "2028", "--recent", "my-app")

				Expect(strings.Join(ui.Outputs(), "\n")).To(ContainSubstring(
					"Unicode Line \u2713\n\u0000",
				))
			})

			It("replaces rune provided by --newline with \\n when the --recent flag is not provided", func() {
				runCommand("--newline", "2713", "my-app")

				Expect(strings.Join(ui.Outputs(), "\n")).To(ContainSubstring(
					"Unicode Line \n\u2028\u0000",
				))
			})

			It("doesn't replace null when the flag isn't specified", func() {
				runCommand("--recent", "my-app")

				Expect(strings.Join(ui.Outputs(), "\n")).To(ContainSubstring(
					"Unicode Line \u2713\u2028\u0000",
				))
			})

			DescribeTable("handles parse errors", func(newline string) {
				Expect(runCommand("--newline", newline, "--recent", "my-app")).To(BeFalse())
			},
				Entry("hex prefix", "0x2028"),
				Entry("non-numeric", "LINE SEPARATOR"),
			)
		})

		Context("when the log messages contain format string identifiers", func() {
			BeforeEach(func() {
				recentLogs = []logs.Loggable{
					testlogs.NewLogMessage("hello%2Bworld%v", app.GUID, "DEA", "1", logmessage.LogMessage_ERR, time.Now()),
				}
			})

			It("does not treat them as format strings", func() {
				runCommand("--recent", "my-app")
				Expect(ui.Outputs()).To(ContainSubstrings([]string{"hello%2Bworld%v"}))
			})
		})

		It("tails the app's logs when no flags are given", func() {
			runCommand("my-app")

			appGUID, _, _, _ := logsRepo.TailLogsForArgsForCall(0)
			Expect(app.GUID).To(Equal(appGUID))
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Connected, tailing logs for app", "my-app", "my-org", "my-space", "my-user"},
				[]string{"Log Line 1"},
			))
		})

		Context("when the loggregator server has an invalid cert", func() {
			Context("when the skip-ssl-validation flag is not set", func() {
				It("fails and informs the user about the skip-ssl-validation flag", func() {
					logsRepo.TailLogsForStub = func(appGUID string, onConnect func(), logChan chan<- logs.Loggable, errChan chan<- error) {
						errChan <- errors.NewInvalidSSLCert("https://example.com", "it don't work good")
					}
					runCommand("my-app")

					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"Received invalid SSL certificate", "https://example.com"},
						[]string{"TIP"},
					))
				})

				It("informs the user of the error when they include the --recent flag", func() {
					logsRepo.RecentLogsForReturns(nil, errors.NewInvalidSSLCert("https://example.com", "how does SSL work???"))
					runCommand("--recent", "my-app")

					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"Received invalid SSL certificate", "https://example.com"},
						[]string{"TIP"},
					))
				})
			})
		})

		Context("when the loggregator server has a valid cert", func() {
			It("tails logs", func() {
				runCommand("my-app")
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Connected, tailing logs for app", "my-org", "my-space", "my-user"},
				))
			})
		})
	})
})
