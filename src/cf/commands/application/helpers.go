package application

import (
	"cf"
	"cf/terminal"
	"fmt"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	"regexp"
	"strings"
	"time"
)

const (
	TIMESTAMP_FORMAT = "2006-01-02T15:04:05.00-0700"
)

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
	longestHeader := fmt.Sprintf("%s [Executor] ", timeFormat)
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
	logType := "STDOUT"

	if logMsg.GetMessageType() == logmessage.LogMessage_ERR {
		coloringFunc = terminal.LogStderrColor
		logType = "STDERR"
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

func coloredAppState(app cf.Application) string {
	appState := strings.ToLower(app.State)

	if app.RunningInstances == 0 {
		if appState == "stopped" {
			return appState
		} else {
			return terminal.CrashedColor(appState)
		}
	}

	if app.RunningInstances < app.Instances {
		return terminal.WarningColor(appState)
	}

	return appState
}

func coloredAppInstaces(app cf.Application) string {
	healthString := fmt.Sprintf("%d/%d", app.RunningInstances, app.Instances)

	if app.RunningInstances == 0 {
		if strings.ToLower(app.State) == "stopped" {
			return healthString
		} else {
			return terminal.CrashedColor(healthString)
		}
	}

	if app.RunningInstances < app.Instances {
		return terminal.WarningColor(healthString)
	}

	return healthString
}

func coloredInstanceState(instance cf.ApplicationInstance) (colored string) {
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
