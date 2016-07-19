package uihelpers

import (
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"code.cloudfoundry.org/cli/cf/terminal"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
)

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func ExtractLogHeader(msg *logmessage.LogMessage, loc *time.Location) (logHeader, coloredLogHeader string) {
	logMsg := msg
	sourceName := logMsg.GetSourceName()
	sourceID := logMsg.GetSourceId()
	t := time.Unix(0, logMsg.GetTimestamp())
	timeFormat := "2006-01-02T15:04:05.00-0700"
	timeString := t.In(loc).Format(timeFormat)

	if sourceID == "" {
		logHeader = fmt.Sprintf("%s [%s]", timeString, sourceName)
	} else {
		logHeader = fmt.Sprintf("%s [%s/%s]", timeString, sourceName, sourceID)
	}

	coloredLogHeader = terminal.LogSysHeaderColor(logHeader)

	// Calculate padding
	longestHeader := fmt.Sprintf("%s  [HEALTH/10] ", timeFormat)
	expectedHeaderLength := utf8.RuneCountInString(longestHeader)
	padding := strings.Repeat(" ", max(0, expectedHeaderLength-utf8.RuneCountInString(logHeader)))

	logHeader = logHeader + padding
	coloredLogHeader = coloredLogHeader + padding

	return
}

var newLinesPattern = regexp.MustCompile("[\n\r]+$")

func ExtractLogContent(logMsg *logmessage.LogMessage, logHeader string) (logContent string) {
	msgText := string(logMsg.GetMessage())
	msgText = newLinesPattern.ReplaceAllString(msgText, "")

	msgLines := strings.Split(msgText, "\n")
	padding := strings.Repeat(" ", utf8.RuneCountInString(logHeader))
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
