package v2action

import (
	"sort"
	"time"

	"github.com/cloudfoundry/noaa"
	noaaErrors "github.com/cloudfoundry/noaa/errors"
	"github.com/cloudfoundry/sonde-go/events"
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

func (Actor) GetStreamingLogs(appGUID string, client NOAAClient, config Config) (<-chan *LogMessage, <-chan error) {
	// Do not pass in token because client should have a TokenRefresher set
	eventStream, errStream := client.TailingLogs(appGUID, "")

	messages := make(chan *LogMessage)
	errs := make(chan error)

	go func() {
		defer close(messages)
		defer close(errs)

		ticker := time.NewTicker(flushInterval)
		defer ticker.Stop()

		var logs LogMessages

	dance:
		for {
			select {
			case event, ok := <-eventStream:
				if !ok {
					break dance
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
					break dance
				}

				if _, ok := err.(noaaErrors.RetryError); ok {
					break
				}

				if err != nil {
					errs <- err
				}
			case <-ticker.C:
				sort.Stable(logs)
				for _, l := range logs {
					messages <- l
				}

				logs = logs[0:0]
			}
		}
	}()

	return messages, errs
}

func (actor Actor) GetRecentLogsForApplicationByNameAndSpace(appName string, spaceGUID string, client NOAAClient, config Config) ([]LogMessage, Warnings, error) {
	app, allWarnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	if err != nil {
		return nil, allWarnings, err
	}

	noaaMessages, err := client.RecentLogs(app.GUID, "")
	if err != nil {
		return nil, allWarnings, err
	}

	noaaMessages = noaa.SortRecent(noaaMessages)

	var logMessages []LogMessage

	for _, message := range noaaMessages {
		logMessages = append(logMessages, LogMessage{
			message:        string(message.GetMessage()),
			messageType:    message.GetMessageType(),
			timestamp:      time.Unix(0, message.GetTimestamp()),
			sourceType:     message.GetSourceType(),
			sourceInstance: message.GetSourceInstance(),
		})
	}

	return logMessages, allWarnings, nil
}

func (actor Actor) GetStreamingLogsForApplicationByNameAndSpace(appName string, spaceGUID string, client NOAAClient, config Config) (<-chan *LogMessage, <-chan error, Warnings, error) {
	app, allWarnings, err := actor.GetApplicationByNameAndSpace(appName, spaceGUID)
	if err != nil {
		return nil, nil, allWarnings, err
	}

	messages, logErrs := actor.GetStreamingLogs(app.GUID, client, config)

	return messages, logErrs, allWarnings, err
}
