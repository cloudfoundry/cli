package application_test

import (
	. "cf/commands/application"
	"cf/models"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"time"
)

func init() {
	Describe("Testing with ginkgo", func() {
		It("TestLogsFailWithUsage", func() {
			reqFactory, logsRepo := getLogsDependencies()

			ui := callLogs([]string{}, reqFactory, logsRepo)
			assert.True(mr.T(), ui.FailedWithUsage)

			ui = callLogs([]string{"foo"}, reqFactory, logsRepo)
			assert.False(mr.T(), ui.FailedWithUsage)
		})
		It("TestLogsRequirements", func() {

			reqFactory, logsRepo := getLogsDependencies()

			reqFactory.LoginSuccess = true
			callLogs([]string{"my-app"}, reqFactory, logsRepo)
			assert.True(mr.T(), testcmd.CommandDidPassRequirements)
			assert.Equal(mr.T(), reqFactory.ApplicationName, "my-app")

			reqFactory.LoginSuccess = false
			callLogs([]string{"my-app"}, reqFactory, logsRepo)
			assert.False(mr.T(), testcmd.CommandDidPassRequirements)
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

			assert.Equal(mr.T(), reqFactory.ApplicationName, "my-app")
			assert.Equal(mr.T(), app.Guid, logsRepo.AppLoggedGuid)
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

			assert.Equal(mr.T(), reqFactory.ApplicationName, "my-app")
			assert.Equal(mr.T(), app.Guid, logsRepo.AppLoggedGuid)
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Connected, tailing logs for app", "my-app", "my-org", "my-space", "my-user"},
				{"Log Line 1"},
			})
		})
	})
}

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
