package logs

import (
	"fmt"
	"strings"
	"time"

	"code.cloudfoundry.org/cli/actor/sharedaction"
)

type logCacheMessage struct {
	msg sharedaction.LogMessage
}

func NewLogCacheMessage(m sharedaction.LogMessage) *logCacheMessage {
	return &logCacheMessage{
		msg: m,
	}
}

func (m *logCacheMessage) ToSimpleLog() string {
	return m.msg.Message()
}

func (m *logCacheMessage) GetSourceName() string {
	return m.msg.SourceType()
}

func (m *logCacheMessage) ToLog(loc *time.Location) string {

	logMsg := m.msg

	sourceName := logMsg.SourceType()
	sourceID := logMsg.SourceInstance()
	t := logMsg.Timestamp()
	timeFormat := "2006-01-02T15:04:05.00-0700"
	timeString := t.In(loc).Format(timeFormat)

	var logHeader string

	if sourceID == "" {
		logHeader = fmt.Sprintf("%s [%s]", timeString, sourceName)
	} else {
		logHeader = fmt.Sprintf("%s [%s/%s]", timeString, sourceName, sourceID)
	}

	// coloredLogHeader := terminal.LogSysHeaderColor(logHeader)

	// // Calculate padding
	// longestHeader := fmt.Sprintf("%s  [HEALTH/10] ", timeFormat)
	// expectedHeaderLength := utf8.RuneCountInString(longestHeader)
	// headerPadding := strings.Repeat(" ", max(0, expectedHeaderLength - utf8.RuneCountInString(logHeader)))

	// logHeader += headerPadding
	// coloredLogHeader += headerPadding

	msgText := logMsg.Message()
	msgText = strings.TrimRight(msgText, "\r\n")

	// msgLines := strings.Split(msgText, "\n")
	// contentPadding := strings.Repeat(" ", utf8.RuneCountInString(logHeader))
	// coloringFunc := terminal.LogStdoutColor

	logType := logMsg.Type()
	// if logType == "ERR" {
	// coloringFunc = terminal.LogStderrColor
	// } else {
	// logType = "OUT"
	// }

	// logContent := fmt.Sprintf("%s %s", logType, msgLines[0])
	// for _, msgLine := range msgLines[1:] {
	//     logContent = fmt.Sprintf("%s\n%s%s", logContent, contentPadding, msgLine)
	// }

	// logContent = coloringFunc(logContent)

	// return fmt.Sprintf("%s%s", coloredLogHeader, logContent)
	return fmt.Sprintf("%s %s %s", logHeader, logType, msgText)
}
