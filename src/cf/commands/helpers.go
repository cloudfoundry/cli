package commands

import (
	"cf/terminal"
	"fmt"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	"strings"
	"time"
)

const (
	BYTE     = 1.0
	KILOBYTE = 1024 * BYTE
	MEGABYTE = 1024 * KILOBYTE
	GIGABYTE = 1024 * MEGABYTE
	TERABYTE = 1024 * GIGABYTE
)

func byteSize(bytes int) string {
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
	}

	stringValue := fmt.Sprintf("%.1f", value)
	stringValue = strings.TrimRight(stringValue, ".0")
	return fmt.Sprintf("%s%s", stringValue, unit)
}

func coloredState(state string) (colored string) {
	switch state {
	case "started", "running":
		colored = terminal.SuccessColor("running")
	case "stopped":
		colored = terminal.StoppedColor("stopped")
	case "flapping":
		colored = terminal.WarningColor("flapping")
	case "starting":
		colored = terminal.AdvisoryColor("starting")
	default:
		colored = terminal.FailureColor(state)
	}

	return
}

func logMessageOutput(appName string, lm *logmessage.LogMessage) string {
	sourceTypeNames := map[logmessage.LogMessage_SourceType]string{
		logmessage.LogMessage_CLOUD_CONTROLLER: "API",
		logmessage.LogMessage_ROUTER:           "Router",
		logmessage.LogMessage_UAA:              "UAA",
		logmessage.LogMessage_DEA:              "Executor",
		logmessage.LogMessage_WARDEN_CONTAINER: "App",
	}

	sourceType, _ := sourceTypeNames[*lm.SourceType]
	sourceId := "?"
	if lm.SourceId != nil {
		sourceId = *lm.SourceId
	}
	msg := lm.GetMessage()

	t := time.Unix(0, *lm.Timestamp)
	timeString := t.Format("Jan 2 15:04:05")

	channel := ""
	if lm.MessageType != nil && *lm.MessageType == logmessage.LogMessage_ERR {
		channel = "STDERR "
	}

	return fmt.Sprintf("%s %s %s/%s %s%s", timeString, appName, sourceType, sourceId, channel, msg)
}
