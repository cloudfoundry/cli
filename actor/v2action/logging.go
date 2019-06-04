package v2action

import (
	"context"
	"time"

	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	logcache "code.cloudfoundry.org/log-cache/pkg/client"
	"code.cloudfoundry.org/log-cache/pkg/rpc/logcache_v1"
	"github.com/cloudfoundry/sonde-go/events"
	log "github.com/sirupsen/logrus"
)

const (
	StagingLog      = "STG"
	RecentLogsLines = 100
)

var flushInterval = 300 * time.Millisecond

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

//TODO this is only used in tests
func NewLogMessage(message string, messageType int, timestamp time.Time, sourceType string, sourceInstance string) *LogMessage {
	return &LogMessage{
		message:        message,
		messageType:    events.LogMessage_MessageType_name[int32(events.LogMessage_MessageType(messageType))],
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

func (actor Actor) GetStreamingLogs(appGUID string, client LogCacheClient) (<-chan *LogMessage, <-chan error, context.CancelFunc) {
	log.Info("Start Tailing Logs")

	outgoingLogStream := make(chan *LogMessage, 1000)
	outgoingErrStream := make(chan error, 1000)
	ctx, cancelFunc := context.WithCancel(context.Background())
	go logcache.Walk(
		ctx,
		appGUID,
		logcache.Visitor(func(envelopes []*loggregator_v2.Envelope) bool {
			logMessages := convertEnvelopesToLogMessages(envelopes)
			for _, logMessage := range logMessages {
				select {
				case <-ctx.Done():
					return false
				default:
					outgoingLogStream <- &logMessage
				}
			}

			return true
		}),
		client.Read,
		logcache.WithWalkStartTime(time.Unix(0, 0)), //TODO
		//TODO logcache.WithWalkEnvelopeTypes(),
		logcache.WithWalkBackoff(logcache.NewAlwaysRetryBackoff(250*time.Millisecond)),
	)

	go func() {
		<-ctx.Done()
		close(outgoingLogStream)
		close(outgoingErrStream)
	}()

	return outgoingLogStream, outgoingErrStream, cancelFunc
}

func (actor Actor) GetRecentLogsForApplicationByNameAndSpace(appName string, spaceGUID string, client LogCacheClient) ([]LogMessage, Warnings, error) {
	app, allWarnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	if err != nil {
		return nil, allWarnings, err
	}

	envelopes, err := client.Read(
		context.Background(),
		app.GUID,
		time.Time{},
		logcache.WithEnvelopeTypes(logcache_v1.EnvelopeType_LOG),
		logcache.WithLimit(RecentLogsLines),
		logcache.WithDescending(),
	)

	if err != nil {
		return nil, allWarnings, err
	}

	logMessages := convertEnvelopesToLogMessages(envelopes)
	return logMessages, allWarnings, nil
}

func convertEnvelopesToLogMessages(envelopes []*loggregator_v2.Envelope) []LogMessage {
	var logMessages []LogMessage
	for i := len(envelopes) - 1; i >= 0; i-- {
		envelope := envelopes[i]
		logEnvelope, ok := envelope.GetMessage().(*loggregator_v2.Envelope_Log)
		if !ok {
			continue
		}
		log := logEnvelope.Log

		logMessages = append(logMessages, LogMessage{
			message:        string(log.Payload),
			messageType:    loggregator_v2.Log_Type_name[int32(log.Type)],
			timestamp:      time.Unix(0, envelope.GetTimestamp()),
			sourceType:     envelope.GetTags()["source_type"], //TODO magical constant
			sourceInstance: envelope.GetInstanceId(),
		})
	}
	return logMessages
}

func (actor Actor) GetStreamingLogsForApplicationByNameAndSpace(appName string, spaceGUID string, client LogCacheClient) (<-chan *LogMessage, <-chan error, Warnings, error, context.CancelFunc) {
	app, allWarnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	if err != nil {
		return nil, nil, allWarnings, err, func() {}
	}

	messages, logErrs, cancel := actor.GetStreamingLogs(app.GUID, nil)

	return messages, logErrs, allWarnings, err, cancel
}
