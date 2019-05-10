package v3action

import (
	"sort"
	"sync"
	"time"

	"code.cloudfoundry.org/cli/actor/actionerror"
	noaaErrors "github.com/cloudfoundry/noaa/errors"
	"github.com/cloudfoundry/sonde-go/events"
	log "github.com/sirupsen/logrus"
)

const StagingLog = "STG"

var flushInterval = 300 * time.Millisecond

type LogMessage struct {
	message        string
	messageType    events.LogMessage_MessageType
	timestamp      time.Time
	sourceType     string
	sourceInstance string
}

func (log LogMessage) Message() string {
	return log.message
}

func (log LogMessage) Type() string {
	if log.messageType == events.LogMessage_OUT {
		return "OUT"
	}
	return "ERR"
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

func NewLogMessage(message string, messageType int, timestamp time.Time, sourceType string, sourceInstance string) *LogMessage {
	return &LogMessage{
		message:        message,
		messageType:    events.LogMessage_MessageType(messageType),
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

func (actor Actor) GetStreamingLogs(appGUID string, client NOAAClient) (<-chan *LogMessage, <-chan error) {
	log.Info("Start Tailing Logs")

	ready := actor.setOnConnectBlocker(client)

	incomingLogStream, incomingErrStream := client.TailingLogs(appGUID, actor.Config.AccessToken())

	outgoingLogStream, outgoingErrStream := actor.blockOnConnect(ready)

	go actor.streamLogsBetween(incomingLogStream, incomingErrStream, outgoingLogStream, outgoingErrStream)

	return outgoingLogStream, outgoingErrStream
}

func (actor Actor) GetStreamingLogsForApplicationByNameAndSpace(appName string, spaceGUID string, client NOAAClient) (<-chan *LogMessage, <-chan error, Warnings, error) {
	app, allWarnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	if err != nil {
		return nil, nil, allWarnings, err
	}

	messages, logErrs := actor.GetStreamingLogs(app.GUID, client)

	return messages, logErrs, allWarnings, err
}

func (actor Actor) blockOnConnect(ready <-chan bool) (chan *LogMessage, chan error) {
	outgoingLogStream := make(chan *LogMessage)
	outgoingErrStream := make(chan error, 1)

	ticker := time.NewTicker(actor.Config.DialTimeout())

dance:
	for {
		select {
		case _, ok := <-ready:
			if !ok {
				break dance
			}
		case <-ticker.C:
			outgoingErrStream <- actionerror.NOAATimeoutError{}
			break dance
		}
	}

	return outgoingLogStream, outgoingErrStream
}

func (Actor) flushLogs(logs LogMessages, messages chan<- *LogMessage) LogMessages {
	sort.Stable(logs)
	for _, l := range logs {
		messages <- l
	}
	return LogMessages{}
}

func (Actor) setOnConnectBlocker(client NOAAClient) <-chan bool {
	ready := make(chan bool)
	var onlyRunOnInitialConnect sync.Once
	callOnConnectOrRetry := func() {
		onlyRunOnInitialConnect.Do(func() {
			close(ready)
		})
	}

	client.SetOnConnectCallback(callOnConnectOrRetry)

	return ready
}

func (actor Actor) streamLogsBetween(incomingLogStream <-chan *events.LogMessage, incomingErrStream <-chan error, outgoingLogStream chan<- *LogMessage, outgoingErrStream chan<- error) {
	log.Info("Processing Log Stream")

	defer close(outgoingLogStream)
	defer close(outgoingErrStream)

	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	var logsToBeSorted LogMessages
	var eventClosed, errClosed bool

dance:
	for {
		select {
		case event, ok := <-incomingLogStream:
			if !ok {
				if !errClosed {
					log.Debug("logging event stream closed")
				}
				eventClosed = true
				break
			}

			logsToBeSorted = append(logsToBeSorted, &LogMessage{
				message:        string(event.GetMessage()),
				messageType:    event.GetMessageType(),
				timestamp:      time.Unix(0, event.GetTimestamp()),
				sourceInstance: event.GetSourceInstance(),
				sourceType:     event.GetSourceType(),
			})
		case err, ok := <-incomingErrStream:
			if !ok {
				if !errClosed {
					log.Debug("logging error stream closed")
				}
				errClosed = true
				break
			}

			if _, ok := err.(noaaErrors.RetryError); ok {
				break
			}

			if err != nil {
				outgoingErrStream <- err
			}
		case <-ticker.C:
			log.Debug("processing logsToBeSorted")
			logsToBeSorted = actor.flushLogs(logsToBeSorted, outgoingLogStream)
			if eventClosed && errClosed {
				log.Debug("stopping log processing")
				break dance
			}
		}
	}
}
