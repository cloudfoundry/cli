package loggingaction

import (
	"context"
	"time"

	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	logcache "code.cloudfoundry.org/log-cache/pkg/client"
	"code.cloudfoundry.org/log-cache/pkg/rpc/logcache_v1"
)

const (
	StagingLog      = "STG"
	RecentLogsLines = 100
)

type LogMessage struct {
	Message        string
	MessageType    string
	Timestamp      time.Time
	SourceType     string
	SourceInstance string
}

func (l LogMessage) Staging() bool {
	return l.SourceType == StagingLog
}

func convertEnvelopesToLogMessages(envelopes []*loggregator_v2.Envelope) []LogMessage {
	var logMessages []LogMessage
	for _, envelope := range envelopes {
		logEnvelope, ok := envelope.GetMessage().(*loggregator_v2.Envelope_Log)
		if !ok {
			continue
		}
		log := logEnvelope.Log

		logMessages = append(logMessages, LogMessage{
			Message:        string(log.Payload),
			MessageType:    loggregator_v2.Log_Type_name[int32(log.Type)],
			Timestamp:      time.Unix(0, envelope.GetTimestamp()),
			SourceType:     envelope.GetTags()["source_type"],
			SourceInstance: envelope.GetInstanceId(),
		})
	}
	return logMessages
}

func GetRecentLogs(appGUID string, client LogCacheClient) ([]LogMessage, error) {
	envelopes, err := client.Read(
		context.Background(),
		appGUID,
		time.Time{},
		logcache.WithEnvelopeTypes(logcache_v1.EnvelopeType_LOG),
		logcache.WithLimit(RecentLogsLines),
		logcache.WithDescending(),
	)
	if err != nil {
		return nil, err
	}

	logMessages := convertEnvelopesToLogMessages(envelopes)
	var reorderedLogMessages []LogMessage
	for i := len(logMessages) - 1; i >= 0; i-- {
		reorderedLogMessages = append(reorderedLogMessages, logMessages[i])
	}

	return reorderedLogMessages, nil
}
