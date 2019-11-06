package loggingaction

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	logcache "code.cloudfoundry.org/log-cache/pkg/client"
	"code.cloudfoundry.org/log-cache/pkg/rpc/logcache_v1"
	"github.com/sirupsen/logrus"
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

type channelWriter struct {
	errChannel chan error
}

func (c channelWriter) Write(bytes []byte) (n int, err error) {
	c.errChannel <- errors.New(strings.Trim(string(bytes), "\n"))

	return len(bytes), nil
}

func GetStreamingLogs(appGUID string, client LogCacheClient) (<-chan LogMessage, <-chan error, context.CancelFunc) {
	logrus.Info("Start Tailing Logs")

	outgoingLogStream := make(chan LogMessage, 1000)
	outgoingErrStream := make(chan error, 1000)
	ctx, cancelFunc := context.WithCancel(context.Background())
	go func() {
		defer close(outgoingLogStream)
		defer close(outgoingErrStream)

		logcache.Walk(
			ctx,
			appGUID,
			logcache.Visitor(func(envelopes []*loggregator_v2.Envelope) bool {
				logMessages := convertEnvelopesToLogMessages(envelopes)
				for _, logMessage := range logMessages {
					select {
					case <-ctx.Done():
						return false
					default:
						outgoingLogStream <- logMessage
					}
				}

				return true
			}),
			client.Read,
			logcache.WithWalkStartTime(time.Now().Add(-5*time.Second)),
			logcache.WithWalkEnvelopeTypes(logcache_v1.EnvelopeType_LOG),
			logcache.WithWalkBackoff(logcache.NewAlwaysRetryBackoff(250*time.Millisecond)),
			logcache.WithWalkLogger(log.New(channelWriter{
				errChannel: outgoingErrStream,
			}, "", 0)),
		)
	}()

	return outgoingLogStream, outgoingErrStream, cancelFunc
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
