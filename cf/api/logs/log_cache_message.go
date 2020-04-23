package logs

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"code.cloudfoundry.org/cli/actor/sharedaction"
)

//go:generate counterfeiter . ColorLogger
type ColorLogger interface {
	LogSysHeaderColor(string) string
	LogStdoutColor(string) string
	LogStderrColor(string) string
}

type LogCacheMessage struct {
	colorLogger ColorLogger
	msg         sharedaction.LogMessage
}

func NewLogCacheMessage(c ColorLogger, m sharedaction.LogMessage) *LogCacheMessage {
	return &LogCacheMessage{
		colorLogger: c,
		msg:         m,
	}
}

func (m *LogCacheMessage) ToSimpleLog() string {
	return m.msg.Message()
}

func (m *LogCacheMessage) GetSourceName() string {
	return m.msg.SourceType()
}

func (m *LogCacheMessage) ToLog(loc *time.Location) string {
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

	coloredLogHeader := m.colorLogger.LogSysHeaderColor(logHeader)

	// Calculate padding
	longestHeader := fmt.Sprintf("%s  [HEALTH/10] ", timeFormat)
	expectedHeaderLength := utf8.RuneCountInString(longestHeader)
	headerPadding := strings.Repeat(" ", max(0, expectedHeaderLength-utf8.RuneCountInString(logHeader)))

	logHeader += headerPadding
	coloredLogHeader += headerPadding

	msgText := logMsg.Message()
	msgText = strings.TrimRight(msgText, "\r\n")

	msgLines := strings.Split(msgText, "\n")
	contentPadding := strings.Repeat(" ", utf8.RuneCountInString(logHeader))

	logType := logMsg.Type()
	if logType != "ERR" {
		logType = "OUT"
	}

	logContent := fmt.Sprintf("%s %s", logType, msgLines[0])
	for _, msgLine := range msgLines[1:] {
		logContent = fmt.Sprintf("%s\n%s%s", logContent, contentPadding, msgLine)
	}

	if logType == "ERR" {
		logContent = m.colorLogger.LogStderrColor(logContent)
	} else {
		logContent = m.colorLogger.LogStdoutColor(logContent)
	}

	return fmt.Sprintf("%s%s", coloredLogHeader, logContent)
}
