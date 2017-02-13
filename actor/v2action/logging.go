package v2action

import (
	"time"

	"github.com/cloudfoundry/sonde-go/events"
)

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

func (actor Actor) GetStreamingLogs(appGUID string, client NOAAClient) (<-chan *LogMessage, <-chan error) {
	// Do not pass in token because client should have a TokenRefresher set
	eventStream, errStream := client.TailingLogs(appGUID, "")

	messages := make(chan *LogMessage)
	errs := make(chan error)

	go func() {
		defer close(messages)
		defer close(errs)

	dance:
		for {
			select {
			case event, ok := <-eventStream:
				if !ok {
					break dance
				}

				messages <- &LogMessage{
					message:        string(event.GetMessage()),
					messageType:    event.GetMessageType(),
					timestamp:      time.Unix(0, event.GetTimestamp()),
					sourceInstance: event.GetSourceInstance(),
					sourceType:     event.GetSourceType(),
				}
			case err, ok := <-errStream:
				if !ok {
					break dance
				}
				errs <- err
			}
		}
	}()

	return messages, errs
}
