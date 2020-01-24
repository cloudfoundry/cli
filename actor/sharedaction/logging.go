package sharedaction

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

	retryCount    = 5
	retryInterval = time.Millisecond * 250
)

type LogMessage struct {
	message        string
	messageType    string
	timestamp      time.Time
	sourceType     string
	sourceInstance string
}

func (log LogMessage) Message() string {
	return log.message
}

func (log LogMessage) Type() string {
	return log.messageType
}

func (log LogMessage) Staging() bool {
	return log.sourceType == StagingLog
}

func (log LogMessage) Timestamp() time.Time {
	return log.timestamp
}

func (log LogMessage) SourceType() string {
	return log.sourceType
}

func (log LogMessage) SourceInstance() string {
	return log.sourceInstance
}

func NewLogMessage(message string, messageType string, timestamp time.Time, sourceType string, sourceInstance string) *LogMessage {
	return &LogMessage{
		message:        message,
		messageType:    messageType,
		timestamp:      timestamp,
		sourceType:     sourceType,
		sourceInstance: sourceInstance,
	}
}

type LogMessages []*LogMessage

func (lm LogMessages) Len() int { return len(lm) }

func (lm LogMessages) Less(i, j int) bool {
	return lm[i].timestamp.Before(lm[j].timestamp)
}

func (lm LogMessages) Swap(i, j int) {
	lm[i], lm[j] = lm[j], lm[i]
}

type channelWriter struct {
	errChannel chan error
}

func (c channelWriter) Write(bytes []byte) (n int, err error) {
	c.errChannel <- errors.New(strings.Trim(string(bytes), "\n"))

	return len(bytes), nil
}

// cliRetryBackoff returns true for OnErr after sleeping the given interval for a limited number of times,
// and returns true for OnEmpty always.
// Basically: retry x times on connection failures, and wait forever for logs to show up.
type cliRetryBackoff struct {
	interval time.Duration
	maxCount int
	count    int
}

func newCliRetryBackoff(interval time.Duration, maxCount int) *cliRetryBackoff {
	return &cliRetryBackoff{
		interval: interval,
		maxCount: maxCount,
	}
}

func (b *cliRetryBackoff) OnErr(error) bool {
	b.count++
	if b.count >= b.maxCount {
		return false
	}

	time.Sleep(b.interval)
	return true
}

func (b *cliRetryBackoff) OnEmpty() bool {
	return true
}

func (b *cliRetryBackoff) Reset() {
	b.count = 0
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
						outgoingLogStream <- *logMessage
					}
				}

				return true
			}),
			client.Read,
			logcache.WithWalkStartTime(time.Now().Add(-5*time.Second)),
			logcache.WithWalkEnvelopeTypes(logcache_v1.EnvelopeType_LOG),
			logcache.WithWalkBackoff(newCliRetryBackoff(retryInterval, retryCount)),
			logcache.WithWalkLogger(log.New(channelWriter{
				errChannel: outgoingErrStream,
			}, "", 0)),
		)
	}()

	return outgoingLogStream, outgoingErrStream, cancelFunc
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
		reorderedLogMessages = append(reorderedLogMessages, *logMessages[i])
	}

	return reorderedLogMessages, nil
}

func convertEnvelopesToLogMessages(envelopes []*loggregator_v2.Envelope) []*LogMessage {
	var logMessages []*LogMessage
	for _, envelope := range envelopes {
		logEnvelope, ok := envelope.GetMessage().(*loggregator_v2.Envelope_Log)
		if !ok {
			continue
		}
		log := logEnvelope.Log

		logMessages = append(logMessages, NewLogMessage(
			string(log.Payload),
			loggregator_v2.Log_Type_name[int32(log.Type)],
			time.Unix(0, envelope.GetTimestamp()),
			envelope.GetTags()["source_type"],
			envelope.GetInstanceId(),
		))
	}
	return logMessages
}
