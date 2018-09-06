package v3action

import (
	"sort"
	"time"

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
	// Do not pass in token because client should have a TokenRefresher set
	eventStream, errStream := client.TailingLogs(appGUID, "")

	messages := make(chan *LogMessage)
	errs := make(chan error)

	go func() {
		log.Info("Processing Log Stream")

		defer close(messages)
		defer close(errs)

		ticker := time.NewTicker(flushInterval)
		defer ticker.Stop()

		var logs LogMessages
		var eventClosed, errClosed bool

	dance:
		for {
			select {
			case event, ok := <-eventStream:
				if !ok {
					if !errClosed {
						log.Debug("logging event stream closed")
					}
					eventClosed = true
					break
				}

				logs = append(logs, &LogMessage{
					message:        string(event.GetMessage()),
					messageType:    event.GetMessageType(),
					timestamp:      time.Unix(0, event.GetTimestamp()),
					sourceInstance: event.GetSourceInstance(),
					sourceType:     event.GetSourceType(),
				})
			case err, ok := <-errStream:
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
					errs <- err
				}
			case <-ticker.C:
				log.Debug("processing cached logs")
				logs = actor.flushLogs(logs, messages)
				if eventClosed && errClosed {
					log.Debug("stopping log processing")
					break dance
				}
			}
		}
	}()

	return messages, errs
}

func (actor Actor) GetStreamingLogsForApplicationByNameAndSpace(appName string, spaceGUID string, client NOAAClient) (<-chan *LogMessage, <-chan error, Warnings, error) {
	app, allWarnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	if err != nil {
		return nil, nil, allWarnings, err
	}

	messages, logErrs := actor.GetStreamingLogs(app.GUID, client)

	return messages, logErrs, allWarnings, err
}

func (Actor) flushLogs(logs LogMessages, messages chan<- *LogMessage) LogMessages {
	sort.Stable(logs)
	for _, l := range logs {
		messages <- l
	}
	return LogMessages{}
}
