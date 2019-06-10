package ui

import (
	"code.cloudfoundry.org/cli/actor/loggingaction"
	"fmt"
	"github.com/fatih/color"
	"strings"
)

// LogTimestampFormat is the timestamp formatting for log lines.
const LogTimestampFormat = "2006-01-02T15:04:05.00-0700"

// DisplayLogMessage formats and outputs a given log message.
func (ui *UI) DisplayLogMessage(message loggingaction.LogMessage, displayHeader bool) {
	ui.terminalLock.Lock()
	defer ui.terminalLock.Unlock()

	var header string
	if displayHeader {
		time := message.Timestamp.In(ui.TimezoneLocation).Format(LogTimestampFormat)

		header = fmt.Sprintf("%s [%s/%s] %s ",
			time,
			message.SourceType,
			message.SourceInstance,
			message.MessageType,
		)
	}

	for _, line := range strings.Split(message.Message, "\n") {
		logLine := fmt.Sprintf("%s%s", header, strings.TrimRight(line, "\r\n"))
		if message.MessageType == "ERR" {
			logLine = ui.modifyColor(logLine, color.New(color.FgRed))
		}
		fmt.Fprintf(ui.Out, "   %s\n", logLine)
	}
}
