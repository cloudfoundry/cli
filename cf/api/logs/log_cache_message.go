package logs

import (
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
	return ""
}

func (m *logCacheMessage) ToLog(loc *time.Location) string {
	return ""
}
