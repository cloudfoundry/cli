package sharedaction

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	logcache "code.cloudfoundry.org/go-log-cache"
	"code.cloudfoundry.org/go-log-cache/rpc/logcache_v1"
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"github.com/sirupsen/logrus"
)

const (
	StagingLog      = "STG"
	RecentLogsLines = 1000

	retryCount    = 5
	retryInterval = time.Millisecond * 250
	walkDelay     = time.Second * 2
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

// (4/16/2020)
// We have decided not to fully unit test this functionality and wanted to document why
// This logic is very highly coupled to the go-log-cache dependancy we consume.  We struggled with mocks to try and unit test this
// We also thought of other ways to refactor this that might have allowed us to better unit test this, we believe while this would add some value
// It would not be valuable enough to justify the effort.  We are confident that the Log Egress team has adequate testing in the go-log-cache client
// And are confident in our integration tests to ensure we are able to stream logs.

func GetStreamingLogs(appGUID string, client LogCacheClient) (<-chan LogMessage, <-chan error, context.CancelFunc) {

	logrus.Info("Start Tailing Logs")

	outgoingLogStream := make(chan LogMessage, 1000)
	outgoingErrStream := make(chan error, 1000)
	ctx, cancelFunc := context.WithCancel(context.Background())
	go func() {
		defer close(outgoingLogStream)
		defer close(outgoingErrStream)

		ts := latestEnvelopeTimestamp(client, outgoingErrStream, ctx, appGUID)

		// if the context was cancelled we may not have seen an envelope
		if ts.IsZero() {
			return
		}

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
			logcache.WithWalkDelay(walkDelay),
			logcache.WithWalkStartTime(ts),
			logcache.WithWalkEnvelopeTypes(logcache_v1.EnvelopeType_LOG),
			logcache.WithWalkBackoff(newCliRetryBackoff(retryInterval, retryCount)),
			logcache.WithWalkLogger(log.New(channelWriter{
				errChannel: outgoingErrStream,
			}, "", 0)),
		)
	}()

	return outgoingLogStream, outgoingErrStream, cancelFunc
}

func latestEnvelopeTimestamp(client LogCacheClient, errs chan error, ctx context.Context, sourceID string) time.Time {

	// Fetching the most recent timestamp could be implemented with client.Read directly rather than using logcache.Walk
	// We use Walk because we want the extra retry behavior provided through Walk

	// Wrap client.Read in our own function to allow us to specify our own read options
	// https://github.com/cloudfoundry/go-log-cache/issues/27
	r := func(ctx context.Context, sourceID string, _ time.Time, opts ...logcache.ReadOption) ([]*loggregator_v2.Envelope, error) {
		os := []logcache.ReadOption{
			logcache.WithLimit(1),
			logcache.WithDescending(),
		}
		for _, o := range opts {
			os = append(os, o)
		}
		return client.Read(ctx, sourceID, time.Time{}, os...)
	}

	var timestamp time.Time

	logcache.Walk(
		ctx,
		sourceID,
		logcache.Visitor(func(envelopes []*loggregator_v2.Envelope) bool {
			timestamp = time.Unix(0, envelopes[0].Timestamp)
			return false
		}),
		r,
		logcache.WithWalkDelay(walkDelay),
		logcache.WithWalkBackoff(newCliRetryBackoff(retryInterval, retryCount)),
		logcache.WithWalkLogger(log.New(channelWriter{
			errChannel: errs,
		}, "", 0)),
	)

	return timestamp
}

func GetRecentLogs(appGUID string, client LogCacheClient) ([]LogMessage, error) {
	logLineRequestCount := RecentLogsLines
	var envelopes []*loggregator_v2.Envelope
	var err error

	for logLineRequestCount >= 1 {
		envelopes, err = client.Read(
			context.Background(),
			appGUID,
			time.Time{},
			logcache.WithEnvelopeTypes(logcache_v1.EnvelopeType_LOG),
			logcache.WithLimit(logLineRequestCount),
			logcache.WithDescending(),
		)
		if err == nil || err.Error() != "unexpected status code 429" {
			break
		}
		logLineRequestCount /= 2
	}
	if err != nil {
		return nil, fmt.Errorf("Failed to retrieve logs from Log Cache: %s", err)
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
