package application_test

import (
	. "cf/commands/application"
	"cf/errors"
	"cf/models"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"time"
)

var _ = Describe("logs command", func() {
	It("fails with usage when called without one argument", func() {
		reqFactory, logsRepo := getLogsDependencies()

		ui := callLogs([]string{}, reqFactory, logsRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())
	})

	It("fails requirements when not logged in", func() {
		reqFactory, logsRepo := getLogsDependencies()
		reqFactory.LoginSuccess = false

		callLogs([]string{"my-app"}, reqFactory, logsRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})

	It("TestLogsOutputsRecentLogs", func() {
		app := models.Application{}
		app.Name = "my-app"
		app.Guid = "my-app-guid"

		currentTime := time.Now()

		recentLogs := []*logmessage.LogMessage{
			NewLogMessage("Log Line 1", app.Guid, "DEA", currentTime),
			NewLogMessage("Log Line 2", app.Guid, "DEA", currentTime),
		}

		reqFactory, logsRepo := getLogsDependencies()
		reqFactory.Application = app
		logsRepo.RecentLogs = recentLogs

		ui := callLogs([]string{"--recent", "my-app"}, reqFactory, logsRepo)

		Expect(reqFactory.ApplicationName).To(Equal("my-app"))
		Expect(app.Guid).To(Equal(logsRepo.AppLoggedGuid))
		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Connected, dumping recent logs for app", "my-app", "my-org", "my-space", "my-user"},
			{"Log Line 1"},
			{"Log Line 2"},
		})
	})

	It("TestLogsEscapeFormattingVerbs", func() {
		app := models.Application{}
		app.Name = "my-app"
		app.Guid = "my-app-guid"

		recentLogs := []*logmessage.LogMessage{
			NewLogMessage("hello%2Bworld%v", app.Guid, "DEA", time.Now()),
		}

		reqFactory, logsRepo := getLogsDependencies()
		reqFactory.Application = app
		logsRepo.RecentLogs = recentLogs

		ui := callLogs([]string{"--recent", "my-app"}, reqFactory, logsRepo)

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"hello%2Bworld%v"},
		})
	})

	It("TestLogsTailsTheAppLogs", func() {
		app := models.Application{}
		app.Name = "my-app"
		app.Guid = "my-app-guid"

		logs := []*logmessage.LogMessage{
			NewLogMessage("Log Line 1", app.Guid, "DEA", time.Now()),
		}

		reqFactory, logsRepo := getLogsDependencies()
		reqFactory.Application = app
		logsRepo.TailLogMessages = logs

		ui := callLogs([]string{"my-app"}, reqFactory, logsRepo)

		Expect(reqFactory.ApplicationName).To(Equal("my-app"))
		Expect(app.Guid).To(Equal(logsRepo.AppLoggedGuid))
		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Connected, tailing logs for app", "my-app", "my-org", "my-space", "my-user"},
			{"Log Line 1"},
		})
	})

	Context("when the loggregator server has an invalid cert", func() {
		var (
			reqFactory *testreq.FakeReqFactory
			logsRepo   *testapi.FakeLogsRepository
		)

		BeforeEach(func() {
			reqFactory, logsRepo = getLogsDependencies()
		})

		Context("when the skip-ssl-validation flag is not set", func() {
			It("fails and informs the user about the skip-ssl-validation flag", func() {
				logsRepo.TailLogErr = errors.NewInvalidSSLCert("https://example.com", "it don't work good")
				ui := callLogs([]string{"my-app"}, reqFactory, logsRepo)

				testassert.SliceContains(ui.Outputs, testassert.Lines{
					{"Received invalid SSL certificate", "https://example.com"},
					{"TIP"},
				})
			})

			It("informs the user of the error when they include the --recent flag", func() {
				logsRepo.RecentLogErr = errors.NewInvalidSSLCert("https://example.com", "how does SSL work???")
				ui := callLogs([]string{"--recent", "my-app"}, reqFactory, logsRepo)

				testassert.SliceContains(ui.Outputs, testassert.Lines{
					{"Received invalid SSL certificate", "https://example.com"},
					{"TIP"},
				})
			})
		})
	})

	Context("when the loggregator server has a valid cert", func() {
		var (
			flags      []string
			reqFactory *testreq.FakeReqFactory
			logsRepo   *testapi.FakeLogsRepository
		)

		BeforeEach(func() {
			reqFactory, logsRepo = getLogsDependencies()
			flags = []string{"my-app"}
		})

		It("tails logs", func() {
			ui := callLogs(flags, reqFactory, logsRepo)

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"Connected, tailing logs for app", "my-org", "my-space", "my-user"},
			})
		})
	})
})

func getLogsDependencies() (reqFactory *testreq.FakeReqFactory, logsRepo *testapi.FakeLogsRepository) {
	logsRepo = &testapi.FakeLogsRepository{}
	reqFactory = &testreq.FakeReqFactory{LoginSuccess: true}
	return
}

func callLogs(args []string, reqFactory *testreq.FakeReqFactory, logsRepo *testapi.FakeLogsRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("logs", args)

	configRepo := testconfig.NewRepositoryWithDefaults()
	cmd := NewLogs(ui, configRepo, logsRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
