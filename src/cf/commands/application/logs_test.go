package application_test

import (
	. "cf/commands/application"
	"cf/models"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	mr "github.com/tjarratt/mr_t"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"time"
)

var _ = Describe("Testing with ginkgo", func() {
	It("TestLogsFailWithUsage", func() {
		reqFactory, logsRepo := getLogsDependencies()

		ui := callLogs([]string{}, reqFactory, logsRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callLogs([]string{"foo"}, reqFactory, logsRepo)
		Expect(ui.FailedWithUsage).To(BeFalse())
	})
	It("TestLogsRequirements", func() {

		reqFactory, logsRepo := getLogsDependencies()

		reqFactory.LoginSuccess = true
		callLogs([]string{"my-app"}, reqFactory, logsRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())
		Expect(reqFactory.ApplicationName).To(Equal("my-app"))

		reqFactory.LoginSuccess = false
		callLogs([]string{"my-app"}, reqFactory, logsRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})
	It("TestLogsOutputsRecentLogs", func() {

		app := models.Application{}
		app.Name = "my-app"
		app.Guid = "my-app-guid"

		currentTime := time.Now()

		recentLogs := []*logmessage.Message{
			NewLogMessage("Log Line 1", app.Guid, "DEA", currentTime),
			NewLogMessage("Log Line 2", app.Guid, "DEA", currentTime),
		}

		reqFactory, logsRepo := getLogsDependencies()
		reqFactory.Application = app
		logsRepo.RecentLogs = recentLogs

		ui := callLogs([]string{"--recent", "my-app"}, reqFactory, logsRepo)

		Expect(reqFactory.ApplicationName).To(Equal("my-app"))
		Expect(app.Guid).To(Equal(logsRepo.AppLoggedGuid))
		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Connected, dumping recent logs for app", "my-app", "my-org", "my-space", "my-user"},
			{"Log Line 1"},
			{"Log Line 2"},
		})
	})
	It("TestLogsEscapeFormattingVerbs", func() {

		app := models.Application{}
		app.Name = "my-app"
		app.Guid = "my-app-guid"

		recentLogs := []*logmessage.Message{
			NewLogMessage("hello%2Bworld%v", app.Guid, "DEA", time.Now()),
		}

		reqFactory, logsRepo := getLogsDependencies()
		reqFactory.Application = app
		logsRepo.RecentLogs = recentLogs

		ui := callLogs([]string{"--recent", "my-app"}, reqFactory, logsRepo)

		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"hello%2Bworld%v"},
		})
	})
	It("TestLogsTailsTheAppLogs", func() {

		app := models.Application{}
		app.Name = "my-app"
		app.Guid = "my-app-guid"

		logs := []*logmessage.Message{
			NewLogMessage("Log Line 1", app.Guid, "DEA", time.Now()),
		}

		reqFactory, logsRepo := getLogsDependencies()
		reqFactory.Application = app
		logsRepo.TailLogMessages = logs

		ui := callLogs([]string{"my-app"}, reqFactory, logsRepo)

		Expect(reqFactory.ApplicationName).To(Equal("my-app"))
		Expect(app.Guid).To(Equal(logsRepo.AppLoggedGuid))
		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Connected, tailing logs for app", "my-app", "my-org", "my-space", "my-user"},
			{"Log Line 1"},
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
