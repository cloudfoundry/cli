package commands_test

import (
	"cf"
	"cf/api"
	. "cf/commands"
	"cf/configuration"
	"code.google.com/p/gogoprotobuf/proto"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
	"time"
)

func TestRecentLogsWithAppName(t *testing.T) {
	config := &configuration.Configuration{}
	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}
	appRepo := &testhelpers.FakeApplicationRepository{AppByName: app}

	///////////////
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

	/////////////////
	recentLogs := []logmessage.LogMessage{
		logMessage1,
		logMessage2,
	}

	logsRepo := &testhelpers.FakeLogsRepository{RecentLogs: recentLogs}
	reqFactory := &testhelpers.FakeReqFactory{Application: app}

	ui := callRecentLogs([]string{"my-app"}, config, reqFactory, appRepo, logsRepo)

	assert.Equal(t, reqFactory.ApplicationName, "my-app")
	assert.Equal(t, app, logsRepo.AppLogged)
	assert.Equal(t, len(ui.Outputs), 3)
	assert.Contains(t, ui.Outputs[0], "Connected, dumping recent logs...")
	assert.Contains(t, ui.Outputs[1], "Log Line 1")
	assert.Contains(t, ui.Outputs[2], "Log Line 2")
}

func TestRecentLogsWithoutAppNameShowsUsage(t *testing.T) {
	config := &configuration.Configuration{}
	appRepo := &testhelpers.FakeApplicationRepository{}
	logsRepo := &testhelpers.FakeLogsRepository{}
	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}

	reqFactory := &testhelpers.FakeReqFactory{Application: app}

	ui := callRecentLogs([]string{}, config, reqFactory, appRepo, logsRepo)
	assert.True(t, ui.FailedWithUsage)
	ui = callRecentLogs([]string{"my-app"}, config, reqFactory, appRepo, logsRepo)
	assert.False(t, ui.FailedWithUsage)
}

func callRecentLogs(args []string, config *configuration.Configuration, reqFactory *testhelpers.FakeReqFactory, appRepo api.ApplicationRepository, logsRepo api.LogsRepository) (ui *testhelpers.FakeUI) {
	ui = new(testhelpers.FakeUI)
	ctxt := testhelpers.NewContext("logs", args)

	cmd := NewRecentLogs(ui, appRepo, logsRepo)
	testhelpers.RunCommand(cmd, ctxt, reqFactory)
	return
}
