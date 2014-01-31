package application_test

import (
	"cf"
	. "cf/commands/application"
	"cf/configuration"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"

	. "github.com/onsi/ginkgo"
	mr "github.com/tjarratt/mr_t"
	"time"
)

func getLogsDependencies() (reqFactory *testreq.FakeReqFactory, logsRepo *testapi.FakeLogsRepository) {
	logsRepo = &testapi.FakeLogsRepository{}
	reqFactory = &testreq.FakeReqFactory{LoginSuccess: true}
	return
}

func callLogs(t mr.TestingT, args []string, reqFactory *testreq.FakeReqFactory, logsRepo *testapi.FakeLogsRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("logs", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)
	space := cf.SpaceFields{}
	space.Name = "my-space"
	org := cf.OrganizationFields{}
	org.Name = "my-org"
	config := &configuration.Configuration{
		SpaceFields:        space,
		OrganizationFields: org,
		AccessToken:        token,
	}

	cmd := NewLogs(ui, config, logsRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestLogsFailWithUsage", func() {
			reqFactory, logsRepo := getLogsDependencies()

			ui := callLogs(mr.T(), []string{}, reqFactory, logsRepo)
			assert.True(mr.T(), ui.FailedWithUsage)

			ui = callLogs(mr.T(), []string{"foo"}, reqFactory, logsRepo)
			assert.False(mr.T(), ui.FailedWithUsage)
		})
		It("TestLogsRequirements", func() {

			reqFactory, logsRepo := getLogsDependencies()

			reqFactory.LoginSuccess = true
			callLogs(mr.T(), []string{"my-app"}, reqFactory, logsRepo)
			assert.True(mr.T(), testcmd.CommandDidPassRequirements)
			assert.Equal(mr.T(), reqFactory.ApplicationName, "my-app")

			reqFactory.LoginSuccess = false
			callLogs(mr.T(), []string{"my-app"}, reqFactory, logsRepo)
			assert.False(mr.T(), testcmd.CommandDidPassRequirements)
		})
		It("TestLogsOutputsRecentLogs", func() {

			app := cf.Application{}
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

			ui := callLogs(mr.T(), []string{"--recent", "my-app"}, reqFactory, logsRepo)

			assert.Equal(mr.T(), reqFactory.ApplicationName, "my-app")
			assert.Equal(mr.T(), app.Guid, logsRepo.AppLoggedGuid)
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Connected, dumping recent logs for app", "my-app", "my-org", "my-space", "my-user"},
				{"Log Line 1"},
				{"Log Line 2"},
			})
		})
		It("TestLogsEscapeFormattingVerbs", func() {

			app := cf.Application{}
			app.Name = "my-app"
			app.Guid = "my-app-guid"

			recentLogs := []*logmessage.Message{
				NewLogMessage("hello%2Bworld%v", app.Guid, "DEA", time.Now()),
			}

			reqFactory, logsRepo := getLogsDependencies()
			reqFactory.Application = app
			logsRepo.RecentLogs = recentLogs

			ui := callLogs(mr.T(), []string{"--recent", "my-app"}, reqFactory, logsRepo)

			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"hello%2Bworld%v"},
			})
		})
		It("TestLogsTailsTheAppLogs", func() {

			app := cf.Application{}
			app.Name = "my-app"
			app.Guid = "my-app-guid"

			logs := []*logmessage.Message{
				NewLogMessage("Log Line 1", app.Guid, "DEA", time.Now()),
			}

			reqFactory, logsRepo := getLogsDependencies()
			reqFactory.Application = app
			logsRepo.TailLogMessages = logs

			ui := callLogs(mr.T(), []string{"my-app"}, reqFactory, logsRepo)

			assert.Equal(mr.T(), reqFactory.ApplicationName, "my-app")
			assert.Equal(mr.T(), app.Guid, logsRepo.AppLoggedGuid)
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Connected, tailing logs for app", "my-app", "my-org", "my-space", "my-user"},
				{"Log Line 1"},
			})
		})
	})
}
