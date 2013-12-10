package application

import (
	"cf"
	"cf/terminal"
	"code.google.com/p/gogoprotobuf/proto"
	"fmt"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	"regexp"
	"strings"
	"time"
)

const (
	TIMESTAMP_FORMAT = "2006-01-02T15:04:05.00-0700"
)

func NewLogMessage(msgText, appGuid, sourceName string, timestamp time.Time) (msg *logmessage.Message) {
	messageType := logmessage.LogMessage_ERR
	sourceType := logmessage.LogMessage_UNKNOWN

	logMsg := logmessage.LogMessage{
		Message:     []byte(msgText),
		AppId:       proto.String(appGuid),
		MessageType: &messageType,
		SourceType:  &sourceType,
		SourceName:  proto.String(sourceName),
		Timestamp:   proto.Int64(timestamp.UnixNano()),
	}
	data, _ := proto.Marshal(&logMsg)
	msg, _ = logmessage.ParseMessage(data)

	return
}

func simpleLogMessageOutput(msg *logmessage.Message) (msgText string) {
	logMsg := msg.GetLogMessage()
	msgText = string(logMsg.GetMessage())
	reg, err := regexp.Compile("[\n\r]+$")
	if err != nil {
		return
	}
	msgText = reg.ReplaceAllString(msgText, "")
	return
}

func logMessageOutput(msg *logmessage.Message) string {
	logHeader, coloredLogHeader := extractLogHeader(msg)
	logMsg := msg.GetLogMessage()
	logContent := extractLogContent(logMsg, logHeader)

	return fmt.Sprintf("%s%s", coloredLogHeader, logContent)
}

func extractLogHeader(msg *logmessage.Message) (logHeader, coloredLogHeader string) {
	logMsg := msg.GetLogMessage()
	sourceType := msg.GetShortSourceTypeName()
	sourceId := logMsg.GetSourceId()
	t := time.Unix(0, logMsg.GetTimestamp())
	timeFormat := TIMESTAMP_FORMAT
	timeString := t.Format(timeFormat)

	logHeader = fmt.Sprintf("%s [%s]", timeString, sourceType)
	coloredLogHeader = terminal.LogSysHeaderColor(logHeader)

	if logMsg.GetSourceType() == logmessage.LogMessage_WARDEN_CONTAINER {
		logHeader = fmt.Sprintf("%s [%s/%s]", timeString, sourceType, sourceId)
		coloredLogHeader = terminal.LogAppHeaderColor(logHeader)
	}

	// Calculate padding
	longestHeader := fmt.Sprintf("%s  [App/0]  ", timeFormat)
	expectedHeaderLength := len(longestHeader)
	padding := strings.Repeat(" ", expectedHeaderLength-len(logHeader))

	logHeader = logHeader + padding
	coloredLogHeader = coloredLogHeader + padding

	return
}

func extractLogContent(logMsg *logmessage.LogMessage, logHeader string) (logContent string) {
	msgText := string(logMsg.GetMessage())
	reg, err := regexp.Compile("[\n\r]+$")
	if err == nil {
		msgText = reg.ReplaceAllString(msgText, "")
	}

	msgLines := strings.Split(msgText, "\n")
	padding := strings.Repeat(" ", len(logHeader))
	coloringFunc := terminal.LogStdoutColor
	logType := "OUT"

	if logMsg.GetMessageType() == logmessage.LogMessage_ERR {
		coloringFunc = terminal.LogStderrColor
		logType = "ERR"
	}

	logContent = fmt.Sprintf("%s %s", logType, msgLines[0])
	for _, msgLine := range msgLines[1:] {
		logContent = fmt.Sprintf("%s\n%s%s", logContent, padding, msgLine)
	}
	logContent = coloringFunc(logContent)

	return
}

func envVarFound(varName string, existingEnvVars map[string]string) (found bool) {
	for name, _ := range existingEnvVars {
		if name == varName {
			found = true
			return
		}
	}
	return
}

func coloredAppState(app cf.ApplicationFields) string {
	appState := strings.ToLower(app.State)

	if app.RunningInstances == 0 {
		if appState == "stopped" {
			return appState
		} else {
			return terminal.CrashedColor(appState)
		}
	}

	if app.RunningInstances < app.InstanceCount {
		return terminal.WarningColor(appState)
	}

	return appState
}

func coloredAppInstaces(app cf.ApplicationFields) string {
	healthString := fmt.Sprintf("%d/%d", app.RunningInstances, app.InstanceCount)

	if app.RunningInstances == 0 {
		if strings.ToLower(app.State) == "stopped" {
			return healthString
		} else {
			return terminal.CrashedColor(healthString)
		}
	}

	if app.RunningInstances < app.InstanceCount {
		return terminal.WarningColor(healthString)
	}

	return healthString
}

func coloredInstanceState(instance cf.AppInstanceFields) (colored string) {
	state := string(instance.State)
	switch state {
	case "started", "running":
		colored = terminal.StartedColor("running")
	case "stopped":
		colored = terminal.StoppedColor("stopped")
	case "flapping":
		colored = terminal.WarningColor("crashing")
	case "starting":
		colored = terminal.AdvisoryColor("starting")
	default:
		colored = terminal.FailureColor(state)
	}

	return
}
