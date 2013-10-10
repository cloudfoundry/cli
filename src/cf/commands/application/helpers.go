package application

import (
	"cf"
	"cf/terminal"
	"errors"
	"fmt"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	BYTE     = 1.0
	KILOBYTE = 1024 * BYTE
	MEGABYTE = 1024 * KILOBYTE
	GIGABYTE = 1024 * MEGABYTE
	TERABYTE = 1024 * GIGABYTE

	TIMESTAMP_FORMAT = "2006-01-02T15:04:05.00-0700"
)

func byteSize(bytes uint64) string {
	unit := ""
	value := float32(bytes)

	switch {
	case bytes >= TERABYTE:
		unit = "T"
		value = value / TERABYTE
	case bytes >= GIGABYTE:
		unit = "G"
		value = value / GIGABYTE
	case bytes >= MEGABYTE:
		unit = "M"
		value = value / MEGABYTE
	case bytes >= KILOBYTE:
		unit = "K"
		value = value / KILOBYTE
	case bytes == 0:
		return "0"
	}

	stringValue := fmt.Sprintf("%.1f", value)
	stringValue = strings.TrimRight(stringValue, ".0")
	return fmt.Sprintf("%s%s", stringValue, unit)
}

func bytesFromString(s string) (bytes uint64, err error) {
	unit := string(s[len(s)-1])
	stringValue := s[0 : len(s)-1]

	value, err := strconv.ParseUint(stringValue, 10, 0)
	if err != nil {
		return
	}

	switch unit {
	case "T":
		bytes = value * TERABYTE
	case "G":
		bytes = value * GIGABYTE
	case "M":
		bytes = value * MEGABYTE
	case "K":
		bytes = value * KILOBYTE
	}

	if bytes == 0 {
		err = errors.New("Could not parse byte string")
	}

	return
}

func logMessageOutput(appName string, msg *logmessage.Message) string {
	logHeader, coloredLogHeader := extractLogHeader(appName, msg)
	logMsg := msg.GetLogMessage()
	logContent := extractLogContent(logMsg, logHeader)

	return fmt.Sprintf("%s%s", coloredLogHeader, logContent)
}

func extractLogHeader(appName string, msg *logmessage.Message) (logHeader, coloredLogHeader string) {
	logMsg := msg.GetLogMessage()
	sourceType := msg.GetShortSourceTypeName()
	sourceId := logMsg.GetSourceId()
	t := time.Unix(0, logMsg.GetTimestamp())
	timeFormat := TIMESTAMP_FORMAT
	timeString := t.Format(timeFormat)

	logHeader = fmt.Sprintf("%s %s [%s]", timeString, appName, sourceType)
	coloredLogHeader = terminal.LogSysHeaderColor(logHeader)

	if logMsg.GetSourceType() == logmessage.LogMessage_WARDEN_CONTAINER {
		logHeader = fmt.Sprintf("%s %s [%s/%s]", timeString, appName, sourceType, sourceId)
		coloredLogHeader = terminal.LogAppHeaderColor(logHeader)
	}

	// Calculate padding
	longestHeader := fmt.Sprintf("%s %s [Executor] ", timeFormat, appName)
	expectedHeaderLength := len(longestHeader)
	padding := strings.Repeat(" ", expectedHeaderLength-len(logHeader))

	logHeader = logHeader + padding
	coloredLogHeader = coloredLogHeader + padding

	return
}

func extractLogContent(logMsg *logmessage.LogMessage, logHeader string) (logContent string) {
	msgText := logMsg.GetMessage()
	msgLines := strings.Split(string(msgText), "\n")
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

func MapStr(args interface{}) []string {
	r := reflect.ValueOf(args)
	rval := make([]string, r.Len())
	for i := 0; i < r.Len(); i++ {
		rval[i] = r.Index(i).Interface().(fmt.Stringer).String()
	}
	return rval

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
