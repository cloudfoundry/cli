package logs

import (
	"code.google.com/p/gogoprotobuf/proto"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
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
