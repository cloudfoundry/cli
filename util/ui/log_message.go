package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/fatih/color"
)

// LogTimestampFormat is the timestamp formatting for log lines.
const LogTimestampFormat = "2006-01-02T15:04:05.00-0700"

//go:generate counterfeiter . LogMessage

// LogMessage is a log response representing one to many joined lines of a log
// message.
type LogMessage interface {
	Message() string
	Type() string
	Timestamp() time.Time
	SourceType() string
	SourceInstance() string
}

// DisplayLogMessage formats and outputs a given log message.
func (ui *UI) DisplayLogMessage(message LogMessage, displayHeader bool) {
	ui.terminalLock.Lock()
	defer ui.terminalLock.Unlock()

	var header string
	if displayHeader {
		time := message.Timestamp().In(ui.TimezoneLocation).Format(LogTimestampFormat)

		header = fmt.Sprintf("%s [%s/%s] %s ",
			time,
			message.SourceType(),
			message.SourceInstance(),
			message.Type(),
		)
	}

	for _, line := range strings.Split(message.Message(), "\n") {
		logLine := fmt.Sprintf("%s%s", header, strings.TrimRight(line, "\r\n"))
		if message.Type() == "ERR" {
			logLine = ui.modifyColor(logLine, color.New(color.FgRed))
		}
		fmt.Fprintf(ui.Out, "   %s\n", logLine)
	}
}
