package logs

import (
	"time"

	"code.cloudfoundry.org/cli/cf/api/logs"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	"github.com/gogo/protobuf/proto"
)

func NewLogMessage(text string, appGUID string, sourceName string, sourceID string, messageType logmessage.LogMessage_MessageType, timestamp time.Time, drainURLs ...string) logs.Loggable {
	return logs.NewLoggregatorLogMessage(
		&logmessage.LogMessage{
			Message:     []byte(text),
			MessageType: &messageType,
			Timestamp:   proto.Int64(timestamp.UnixNano()),
			AppId:       proto.String(appGUID),
			SourceId:    proto.String(sourceID),
			DrainUrls:   drainURLs,
			SourceName:  proto.String(sourceName),
		},
	)
}
