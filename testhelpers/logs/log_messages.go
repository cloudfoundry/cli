package logs

import (
	"time"

	"github.com/cloudfoundry/cli/cf/api/logs"
	"github.com/cloudfoundry/loggregatorlib/logmessage"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"
)

func NewLogMessage(
	text string,
	appGUID string,
	sourceName string,
	sourceID string,
	messageType logmessage.LogMessage_MessageType,
	timestamp time.Time,
	drainURLs ...string,
) logs.Loggable {
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

func NewNoaaLogMessage(msgText, appGuid, sourceName string, timestamp time.Time) *events.LogMessage {
	messageType := events.LogMessage_ERR

	return &events.LogMessage{
		Message:     []byte(msgText),
		AppId:       proto.String(appGuid),
		MessageType: &messageType,
		SourceType:  proto.String(sourceName),
		Timestamp:   proto.Int64(timestamp.UnixNano()),
	}
}
