package application_test

import (
	"cf"
	. "cf/commands/application"
	"cf/configuration"
	"code.google.com/p/gogoprotobuf/proto"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
	"time"
)

func TestLogsFailWithUsage(t *testing.T) {
	reqFactory, logsRepo := getLogsDependencies()

	fakeUI := callLogs(t, []string{}, reqFactory, logsRepo)
	assert.True(t, fakeUI.FailedWithUsage)

	fakeUI = callLogs(t, []string{"foo"}, reqFactory, logsRepo)
	assert.False(t, fakeUI.FailedWithUsage)
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
	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}

	currentTime := time.Now()
	messageType := logmessage.LogMessage_ERR
	sourceType := logmessage.LogMessage_DEA
	logMessage1 := logmessage.LogMessage{
		Message:     []byte("Log Line 1"),
		AppId:       proto.String("my-app"),
		MessageType: &messageType,
		SourceType:  &sourceType,
		Timestamp:   proto.Int64(currentTime.UnixNano()),
	}

	logMessage2 := logmessage.LogMessage{
		Message:     []byte("Log Line 2"),
		AppId:       proto.String("my-app"),
		MessageType: &messageType,
		SourceType:  &sourceType,
		Timestamp:   proto.Int64(currentTime.UnixNano()),
	}

	recentLogs := []logmessage.LogMessage{
		logMessage1,
		logMessage2,
	}

	reqFactory, logsRepo := getLogsDependencies()
	reqFactory.Application = app
	logsRepo.RecentLogs = recentLogs

	ui := callLogs(t, []string{"--recent", "my-app"}, reqFactory, logsRepo)

	assert.Equal(t, reqFactory.ApplicationName, "my-app")
	assert.Equal(t, app, logsRepo.AppLogged)
	assert.Equal(t, len(ui.Outputs), 3)
	assert.Contains(t, ui.Outputs[0], "Connected, dumping recent logs for app")
	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[0], "my-org")
	assert.Contains(t, ui.Outputs[0], "my-space")
	assert.Contains(t, ui.Outputs[0], "my-user")
	assert.Contains(t, ui.Outputs[1], "Log Line 1")
	assert.Contains(t, ui.Outputs[2], "Log Line 2")
}

func TestLogsTailsTheAppLogs(t *testing.T) {
	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}

	currentTime := time.Now()
	messageType := logmessage.LogMessage_ERR
	deaSourceType := logmessage.LogMessage_DEA
	deaSourceId := "42"
	deaLogMessage := logmessage.LogMessage{
		Message:     []byte("Log Line 1"),
		AppId:       proto.String("my-app"),
		MessageType: &messageType,
		SourceType:  &deaSourceType,
		SourceId:    &deaSourceId,
		Timestamp:   proto.Int64(currentTime.UnixNano()),
	}

	logs := []logmessage.LogMessage{deaLogMessage}

	reqFactory, logsRepo := getLogsDependencies()
	reqFactory.Application = app
	logsRepo.TailLogMessages = logs

	ui := callLogs(t, []string{"my-app"}, reqFactory, logsRepo)

	assert.Equal(t, reqFactory.ApplicationName, "my-app")
	assert.Equal(t, app, logsRepo.AppLogged)
	assert.Equal(t, len(ui.Outputs), 2)
	assert.Contains(t, ui.Outputs[0], "Connected, tailing logs for app")
	assert.Contains(t, ui.Outputs[0], "my-app")
	assert.Contains(t, ui.Outputs[0], "my-org")
	assert.Contains(t, ui.Outputs[0], "my-space")
	assert.Contains(t, ui.Outputs[0], "my-user")
	assert.Contains(t, ui.Outputs[1], "Log Line 1")
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

	config := &configuration.Configuration{
		Space:        cf.Space{Name: "my-space"},
		Organization: cf.Organization{Name: "my-org"},
		AccessToken:  token,
	}

	cmd := NewLogs(ui, config, logsRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
