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

   src/github.com/cloudfoundry/cli/cf/commands/application/delete_app_test.go
   src/github.com/cloudfoundry/cli/cf/terminal/ui_test.go
   src/github.com/cloudfoundry/loggregator_consumer/consumer_test.go
*/

package application_test

import (
	"time"

	"code.google.com/p/gogoprotobuf/proto"
	. "github.com/cloudfoundry/cli/cf/commands/application"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/terminal"
	testapi "github.com/cloudfoundry/cli/testhelpers/api"
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

var _ = Describe("logs command", func() {
	It("fails with usage when called without one argument", func() {
		requirementsFactory, logsRepo := getLogsDependencies()

		ui := callLogs([]string{}, requirementsFactory, logsRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())
	})

	It("fails requirements when not logged in", func() {
		requirementsFactory, logsRepo := getLogsDependencies()
		requirementsFactory.LoginSuccess = false

		callLogs([]string{"my-app"}, requirementsFactory, logsRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})

	It("TestLogsOutputsRecentLogs", func() {
		app := models.Application{}
		app.Name = "my-app"
		app.Guid = "my-app-guid"

		currentTime := time.Now()

		recentLogs := []*logmessage.LogMessage{
			testlogs.NewLogMessage("Log Line 1", app.Guid, "DEA", currentTime),
			testlogs.NewLogMessage("Log Line 2", app.Guid, "DEA", currentTime),
		}

		requirementsFactory, logsRepo := getLogsDependencies()
		requirementsFactory.Application = app
		logsRepo.RecentLogs = recentLogs

		ui := callLogs([]string{"--recent", "my-app"}, requirementsFactory, logsRepo)

		Expect(requirementsFactory.ApplicationName).To(Equal("my-app"))
		Expect(app.Guid).To(Equal(logsRepo.AppLoggedGuid))
		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Connected, dumping recent logs for app", "my-app", "my-org", "my-space", "my-user"},
			[]string{"Log Line 1"},
			[]string{"Log Line 2"},
		))
	})

	It("TestLogsEscapeFormattingVerbs", func() {
		app := models.Application{}
		app.Name = "my-app"
		app.Guid = "my-app-guid"

		recentLogs := []*logmessage.LogMessage{
			testlogs.NewLogMessage("hello%2Bworld%v", app.Guid, "DEA", time.Now()),
		}

		requirementsFactory, logsRepo := getLogsDependencies()
		requirementsFactory.Application = app
		logsRepo.RecentLogs = recentLogs

		ui := callLogs([]string{"--recent", "my-app"}, requirementsFactory, logsRepo)

		Expect(ui.Outputs).To(ContainSubstrings([]string{"hello%2Bworld%v"}))
	})

	It("TestLogsTailsTheAppLogs", func() {
		app := models.Application{}
		app.Name = "my-app"
		app.Guid = "my-app-guid"

		logs := []*logmessage.LogMessage{
			testlogs.NewLogMessage("Log Line 1", app.Guid, "DEA", time.Now()),
		}

		requirementsFactory, logsRepo := getLogsDependencies()
		requirementsFactory.Application = app
		logsRepo.TailLogMessages = logs

		ui := callLogs([]string{"my-app"}, requirementsFactory, logsRepo)

		Expect(requirementsFactory.ApplicationName).To(Equal("my-app"))
		Expect(app.Guid).To(Equal(logsRepo.AppLoggedGuid))
		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Connected, tailing logs for app", "my-app", "my-org", "my-space", "my-user"},
			[]string{"Log Line 1"},
		))
	})

	Context("when the loggregator server has an invalid cert", func() {
		var (
			requirementsFactory *testreq.FakeReqFactory
			logsRepo            *testapi.FakeLogsRepository
		)

		BeforeEach(func() {
			requirementsFactory, logsRepo = getLogsDependencies()
		})

		Context("when the skip-ssl-validation flag is not set", func() {
			It("fails and informs the user about the skip-ssl-validation flag", func() {
				logsRepo.TailLogErr = errors.NewInvalidSSLCert("https://example.com", "it don't work good")
				ui := callLogs([]string{"my-app"}, requirementsFactory, logsRepo)

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Received invalid SSL certificate", "https://example.com"},
					[]string{"TIP"},
				))
			})

			It("informs the user of the error when they include the --recent flag", func() {
				logsRepo.RecentLogErr = errors.NewInvalidSSLCert("https://example.com", "how does SSL work???")
				ui := callLogs([]string{"--recent", "my-app"}, requirementsFactory, logsRepo)

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Received invalid SSL certificate", "https://example.com"},
					[]string{"TIP"},
				))
			})
		})
	})

	Context("when the loggregator server has a valid cert", func() {
		var (
			flags               []string
			requirementsFactory *testreq.FakeReqFactory
			logsRepo            *testapi.FakeLogsRepository
		)

		BeforeEach(func() {
			requirementsFactory, logsRepo = getLogsDependencies()
			flags = []string{"my-app"}
		})

		It("tails logs", func() {
			ui := callLogs(flags, requirementsFactory, logsRepo)

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Connected, tailing logs for app", "my-org", "my-space", "my-user"},
			))
		})
	})

	Describe("Helpers", func() {
		date := time.Date(2014, 4, 4, 11, 39, 20, 5, time.UTC)

		createMessage := func(sourceId string, sourceName string, msgType logmessage.LogMessage_MessageType, date time.Time) *logmessage.LogMessage {
			timestamp := date.UnixNano()
			return &logmessage.LogMessage{
				Message:     []byte("Hello World!\n\r\n\r"),
				AppId:       proto.String("my-app-guid"),
				MessageType: &msgType,
				SourceId:    &sourceId,
				Timestamp:   &timestamp,
				SourceName:  &sourceName,
			}
		}

		Context("when the message comes from an app", func() {
			It("includes the instance index", func() {
				msg := createMessage("4", "App", logmessage.LogMessage_OUT, date)
				Expect(terminal.Decolorize(LogMessageOutput(msg, time.UTC))).To(Equal("2014-04-04T11:39:20.00+0000 [App/4]   OUT Hello World!"))
			})
		})

		Context("when the message comes from a cloudfoundry component", func() {
			It("doesn't include the instance index", func() {
				msg := createMessage("4", "DEA", logmessage.LogMessage_OUT, date)
				Expect(terminal.Decolorize(LogMessageOutput(msg, time.UTC))).To(Equal("2014-04-04T11:39:20.00+0000 [DEA]     OUT Hello World!"))
			})
		})

		Context("when the message was written to stderr", func() {
			It("shows the log type as 'ERR'", func() {
				msg := createMessage("4", "DEA", logmessage.LogMessage_ERR, date)
				Expect(terminal.Decolorize(LogMessageOutput(msg, time.UTC))).To(Equal("2014-04-04T11:39:20.00+0000 [DEA]     ERR Hello World!"))
			})
		})

		It("formats the time in the given time zone", func() {
			msg := createMessage("4", "DEA", logmessage.LogMessage_ERR, date)
			Expect(terminal.Decolorize(LogMessageOutput(msg, time.FixedZone("the-zone", 3*60*60)))).To(Equal("2014-04-04T14:39:20.00+0300 [DEA]     ERR Hello World!"))
		})
	})
})

func getLogsDependencies() (requirementsFactory *testreq.FakeReqFactory, logsRepo *testapi.FakeLogsRepository) {
	logsRepo = &testapi.FakeLogsRepository{}
	requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true}
	return
}

func callLogs(args []string, requirementsFactory *testreq.FakeReqFactory, logsRepo *testapi.FakeLogsRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)

	configRepo := testconfig.NewRepositoryWithDefaults()
	cmd := NewLogs(ui, configRepo, logsRepo)
	testcmd.RunCommand(cmd, args, requirementsFactory)
	return
}
