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
	"testing"
	"time"
)

func TestLogsFailWithUsage(t *testing.T) {
	reqFactory, logsRepo := getLogsDependencies()

	ui := callLogs(t, []string{}, reqFactory, logsRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callLogs(t, []string{"foo"}, reqFactory, logsRepo)
	assert.False(t, ui.FailedWithUsage)
}

func TestLogsRequirements(t *testing.T) {
	reqFactory, logsRepo := getLogsDependencies()

	reqFactory.LoginSuccess = true
	callLogs(t, []string{"my-app"}, reqFactory, logsRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)
	assert.Equal(t, reqFactory.ApplicationName, "my-app")

	reqFactory.LoginSuccess = false
	callLogs(t, []string{"my-app"}, reqFactory, logsRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)
}

func TestLogsOutputsRecentLogs(t *testing.T) {
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

	ui := callLogs(t, []string{"--recent", "my-app"}, reqFactory, logsRepo)

	assert.Equal(t, reqFactory.ApplicationName, "my-app")
	assert.Equal(t, app.Guid, logsRepo.AppLoggedGuid)

	assert.Equal(t, len(ui.Outputs), 3)
	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Connected, dumping recent logs for app", "my-app", "my-org", "my-space", "my-user"},
		{"Log Line 1"},
		{"Log Line 2"},
	})
}

func TestLogsEscapeFormattingVerbs(t *testing.T) {
	app := cf.Application{}
	app.Name = "my-app"
	app.Guid = "my-app-guid"

	recentLogs := []*logmessage.Message{
		NewLogMessage("hello%2Bworld%v", app.Guid, "DEA", time.Now()),
	}

	reqFactory, logsRepo := getLogsDependencies()
	reqFactory.Application = app
	logsRepo.RecentLogs = recentLogs

	ui := callLogs(t, []string{"--recent", "my-app"}, reqFactory, logsRepo)

	assert.Contains(t, ui.Outputs[1], "hello%2Bworld%v")
}

func TestLogsTailsTheAppLogs(t *testing.T) {
	app := cf.Application{}
	app.Name = "my-app"
	app.Guid = "my-app-guid"

	logs := []*logmessage.Message{
		NewLogMessage("Log Line 1", app.Guid, "DEA", time.Now()),
	}

	reqFactory, logsRepo := getLogsDependencies()
	reqFactory.Application = app
	logsRepo.TailLogMessages = logs

	ui := callLogs(t, []string{"my-app"}, reqFactory, logsRepo)

	assert.Equal(t, reqFactory.ApplicationName, "my-app")
	assert.Equal(t, app.Guid, logsRepo.AppLoggedGuid)

	assert.Equal(t, len(ui.Outputs), 2)
	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Connected, tailing logs for app", "my-app", "my-org", "my-space", "my-user"},
		{"Log Line 1"},
	})
}

func getLogsDependencies() (reqFactory *testreq.FakeReqFactory, logsRepo *testapi.FakeLogsRepository) {
	logsRepo = &testapi.FakeLogsRepository{}
	reqFactory = &testreq.FakeReqFactory{LoginSuccess: true}
	return
}

func callLogs(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory, logsRepo *testapi.FakeLogsRepository) (ui *testterm.FakeUI) {
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
