package logs

import (
	"time"

	"github.com/cloudfoundry/loggregatorlib/logmessage"
	"github.com/gogo/protobuf/proto"
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
