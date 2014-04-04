package application

import (
	"cf/models"
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

func NewLogMessage(msgText, appGuid, sourceName string, timestamp time.Time) *logmessage.LogMessage {
	messageType := logmessage.LogMessage_ERR

	return &logmessage.LogMessage{
		Message:     []byte(msgText),
		AppId:       proto.String(appGuid),
		MessageType: &messageType,
		SourceName:  proto.String(sourceName),
		Timestamp:   proto.Int64(timestamp.UnixNano()),
	}
}

func simpleLogMessageOutput(logMsg *logmessage.LogMessage) (msgText string) {
	msgText = string(logMsg.GetMessage())
	reg, err := regexp.Compile("[\n\r]+$")
	if err != nil {
		return
	}
	msgText = reg.ReplaceAllString(msgText, "")
	return
}

func LogMessageOutput(msg *logmessage.LogMessage) string {
	logHeader, coloredLogHeader := extractLogHeader(msg)
	logContent := extractLogContent(msg, logHeader)

	return fmt.Sprintf("%s%s", coloredLogHeader, logContent)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func extractLogHeader(msg *logmessage.LogMessage) (logHeader, coloredLogHeader string) {
	logMsg := msg
	sourceName := logMsg.GetSourceName()
	sourceID := logMsg.GetSourceId()
	t := time.Unix(0, logMsg.GetTimestamp())
	timeFormat := TIMESTAMP_FORMAT
	timeString := t.Format(timeFormat)

	logHeader = fmt.Sprintf("%s [%s]", timeString, sourceName)
	coloredLogHeader = terminal.LogSysHeaderColor(logHeader)

	if sourceName == "App" {
		logHeader = fmt.Sprintf("%s [%s/%s]", timeString, sourceName, sourceID)
		coloredLogHeader = terminal.LogAppHeaderColor(logHeader)
	}

	// Calculate padding
	longestHeader := fmt.Sprintf("%s  [App/0]  ", timeFormat)
	expectedHeaderLength := len(longestHeader)
	padding := strings.Repeat(" ", max(0, expectedHeaderLength-len(logHeader)))

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

func coloredAppState(app models.ApplicationFields) string {
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

func coloredAppInstances(app models.ApplicationFields) string {
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

func coloredInstanceState(instance models.AppInstanceFields) (colored string) {
	state := string(instance.State)
	switch state {
	case "started", "running":
		colored = "running"
	case "stopped":
		colored = terminal.StoppedColor("stopped")
	case "flapping":
		colored = terminal.CrashedColor("crashing")
	case "down":
		colored = terminal.CrashedColor("down")
	case "starting":
		colored = terminal.AdvisoryColor("starting")
	default:
		colored = terminal.WarningColor(state)
	}

	return
}
