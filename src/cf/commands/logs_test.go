package commands_test

import (
	"cf"
	. "cf/commands"
	"code.google.com/p/gogoprotobuf/proto"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
	"time"
)

func TestLogsFailWithUsage(t *testing.T) {
	reqFactory, logsRepo := getDefaultLogsDependencies()

	fakeUI := callLogs([]string{}, reqFactory, logsRepo)
	assert.True(t, fakeUI.FailedWithUsage)

	fakeUI = callLogs([]string{"foo"}, reqFactory, logsRepo)
	assert.False(t, fakeUI.FailedWithUsage)
}

func TestLogsRequirements(t *testing.T) {
	reqFactory, logsRepo := getDefaultLogsDependencies()

	reqFactory.LoginSuccess = true
	callLogs([]string{"my-app"}, reqFactory, logsRepo)
	assert.True(t, testhelpers.CommandDidPassRequirements)
	assert.Equal(t, reqFactory.ApplicationName, "my-app")

	reqFactory.LoginSuccess = false
	callLogs([]string{"my-app"}, reqFactory, logsRepo)
	assert.False(t, testhelpers.CommandDidPassRequirements)
}

func TestLogsTailsTheAppLogs(t *testing.T) {
	app := cf.Application{Name: "my-app", Guid: "my-app-guid"}

	///////////////
	currentTime := time.Now()
	messageType := logmessage.LogMessage_ERR
	sourceType := logmessage.LogMessage_DEA
	sourceId := "42"
	logMessage1 := &logmessage.LogMessage{
		Message:     []byte("Log Line 1"),
		AppId:       proto.String("my-app"),
		MessageType: &messageType,
		SourceType:  &sourceType,
		SourceId:    &sourceId,
		Timestamp:   proto.Int64(currentTime.UnixNano()),
	}

	otherSourceType := logmessage.LogMessage_ROUTER
	otherSourceId := "49"
	logMessage2 := &logmessage.LogMessage{
		Message:     []byte("Log Line 2"),
		AppId:       proto.String("my-app"),
		MessageType: &messageType,
		SourceType:  &otherSourceType,
		SourceId:    &otherSourceId,
		Timestamp:   proto.Int64(currentTime.UnixNano()),
	}

	/////////////////
	logs := []*logmessage.LogMessage{
		logMessage1,
		logMessage2,
	}

	reqFactory, logsRepo := getDefaultLogsDependencies()
	reqFactory.Application = app
	logsRepo.TailLogMessages = logs

	ui := callLogs([]string{"my-app"}, reqFactory, logsRepo)

	assert.Equal(t, reqFactory.ApplicationName, "my-app")
	assert.Equal(t, app, logsRepo.AppLogged)
	assert.Equal(t, len(ui.Outputs), 3)
	assert.Contains(t, ui.Outputs[0], "Connected")
	assert.Contains(t, ui.Outputs[1], "[DEA/42] Log Line 1")
	assert.Contains(t, ui.Outputs[2], "[ROUTER/49] Log Line 2")
}

func getDefaultLogsDependencies() (reqFactory *testhelpers.FakeReqFactory, logsRepo *testhelpers.FakeLogsRepository) {
	logsRepo = &testhelpers.FakeLogsRepository{}
	reqFactory = &testhelpers.FakeReqFactory{LoginSuccess: true}
	return
}

func callLogs(args []string, reqFactory *testhelpers.FakeReqFactory, logsRepo *testhelpers.FakeLogsRepository) (ui *testhelpers.FakeUI) {
	ui = new(testhelpers.FakeUI)
	ctxt := testhelpers.NewContext("logs", args)
	cmd := NewLogs(ui, logsRepo)
	testhelpers.RunCommand(cmd, ctxt, reqFactory)
	return
}
