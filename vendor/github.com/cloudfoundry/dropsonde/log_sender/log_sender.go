package log_sender

import (
	"bufio"
	"io"
	"strings"
	"time"
	"unicode/utf8"

	"fmt"
	"syscall"

	"github.com/cloudfoundry/dropsonde/metrics"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"
)

const (
	maxTagLen = 256
	maxTags   = 10
)

type EventEmitter interface {
	Emit(events.Event) error
	EmitEnvelope(*events.Envelope) error
	Origin() string
}

type LogChainer interface {
	SetTimestamp(t int64) LogChainer
	SetTag(key, value string) LogChainer
	SetAppId(id string) LogChainer
	SetSourceType(s string) LogChainer
	SetSourceInstance(s string) LogChainer
	Send() error
}

// A LogSender emits log events.
type LogSender struct {
	eventEmitter EventEmitter
}

// NewLogSender instantiates a LogSender with the given EventEmitter.
func NewLogSender(eventEmitter EventEmitter) *LogSender {
	return &LogSender{
		eventEmitter: eventEmitter,
	}
}

// SendAppLog sends a log message with the given appid and log message
// with a message type of std out.
// Returns an error if one occurs while sending the event.
func (l *LogSender) SendAppLog(appID, message, sourceType, sourceInstance string) error {
	metrics.BatchIncrementCounter("logSenderTotalMessagesRead")
	return l.eventEmitter.Emit(makeLogMessage(appID, message, sourceType, sourceInstance, events.LogMessage_OUT))
}

// SendAppErrorLog sends a log error message with the given appid and log message
// with a message type of std err.
// Returns an error if one occurs while sending the event.
func (l *LogSender) SendAppErrorLog(appID, message, sourceType, sourceInstance string) error {
	metrics.BatchIncrementCounter("logSenderTotalMessagesRead")
	return l.eventEmitter.Emit(makeLogMessage(appID, message, sourceType, sourceInstance, events.LogMessage_ERR))
}

// ScanLogStream sends a log message with the given meta-data for each line from reader.
// Restarts on read errors and continues until EOF.
func (l *LogSender) ScanLogStream(appID, sourceType, sourceInstance string, reader io.Reader) {
	l.scanLogStream(appID, sourceType, sourceInstance, l.SendAppLog, reader)
}

// ScanErrorLogStream sends a log error message with the given meta-data for each line from reader.
// Restarts on read errors and continues until EOF.
func (l *LogSender) ScanErrorLogStream(appID, sourceType, sourceInstance string, reader io.Reader) {
	l.scanLogStream(appID, sourceType, sourceInstance, l.SendAppErrorLog, reader)
}

// LogMessage creates a log message that can be manipulated via cascading calls
// and then sent.
func (l *LogSender) LogMessage(message []byte, msgType events.LogMessage_MessageType) LogChainer {
	return logChainer{
		emitter: l.eventEmitter,
		envelope: &events.Envelope{
			Origin:    proto.String(l.eventEmitter.Origin()),
			EventType: events.Envelope_LogMessage.Enum(),
			LogMessage: &events.LogMessage{
				Message:     message,
				MessageType: msgType.Enum(),
			},
		},
	}
}

func (l *LogSender) scanLogStream(appID, sourceType, sourceInstance string, sender func(string, string, string, string) error, reader io.Reader) {
	for {
		err := sendScannedLines(appID, sourceType, sourceInstance, bufio.NewScanner(reader), sender)
		if l.isMessageTooLong(err, appID, sourceType, sourceInstance) {
			continue
		}

		return
	}
}

func (l *LogSender) isMessageTooLong(err error, appID string, sourceType string, sourceInstance string) bool {
	if err == nil {
		return false
	}

	if err == bufio.ErrTooLong {
		l.SendAppErrorLog(appID, "Dropped log message: message too long (>64K without a newline)", sourceType, sourceInstance)
		return true
	}

	if strings.Contains(err.Error(), syscall.EMSGSIZE.Error()) {
		l.SendAppErrorLog(appID, fmt.Sprintf("Dropped log message: message could not fit in UDP packet"), sourceType, sourceInstance)
		return true
	}

	return false
}

func makeLogMessage(appID, message, sourceType, sourceInstance string, messageType events.LogMessage_MessageType) *events.LogMessage {
	return &events.LogMessage{
		Message:        []byte(message),
		AppId:          proto.String(appID),
		MessageType:    &messageType,
		SourceType:     &sourceType,
		SourceInstance: &sourceInstance,
		Timestamp:      proto.Int64(time.Now().UnixNano()),
	}
}

func sendScannedLines(appID, sourceType, sourceInstance string, scanner *bufio.Scanner, send func(string, string, string, string) error) error {
	for scanner.Scan() {
		line := scanner.Text()

		if len(strings.TrimSpace(line)) == 0 {
			continue
		}

		err := send(appID, line, sourceType, sourceInstance)
		if err != nil {
			return err
		}
	}
	return scanner.Err()
}

type envelopeEmitter interface {
	EmitEnvelope(*events.Envelope) error
}

type logChainer struct {
	emitter  envelopeEmitter
	envelope *events.Envelope
	err      error
}

func (c logChainer) SetTimestamp(t int64) LogChainer {
	c.envelope.LogMessage.Timestamp = proto.Int64(t)
	return c
}

func (c logChainer) SetTag(key, value string) LogChainer {
	if utf8.RuneCountInString(key) > maxTagLen || utf8.RuneCountInString(value) > maxTagLen {
		return logChainer{
			err: fmt.Errorf("Tag exceeds max length of %d", maxTagLen),
		}
	}

	if c.envelope.Tags == nil {
		c.envelope.Tags = make(map[string]string)
	}
	c.envelope.Tags[key] = value
	if len(c.envelope.Tags) > maxTags {
		return logChainer{
			err: fmt.Errorf("Too many tags. Max of %d", maxTags),
		}
	}
	return c
}

func (c logChainer) SetAppId(id string) LogChainer {
	c.envelope.LogMessage.AppId = proto.String(id)
	return c
}

func (c logChainer) SetSourceType(s string) LogChainer {
	c.envelope.LogMessage.SourceType = proto.String(s)
	return c
}

func (c logChainer) SetSourceInstance(s string) LogChainer {
	c.envelope.LogMessage.SourceInstance = proto.String(s)
	return c
}

// Send sends the log message with the envelope timestamp set to now and the
// log message timestamp set to now if none was provided by SetTimestamp.
func (c logChainer) Send() error {
	if c.err != nil {
		return c.err
	}

	metrics.BatchIncrementCounter("logSenderTotalMessagesRead")

	c.envelope.Timestamp = proto.Int64(time.Now().UnixNano())

	if c.envelope.LogMessage.Timestamp == nil {
		c.envelope.LogMessage.Timestamp = proto.Int64(time.Now().UnixNano())
	}
	return c.emitter.EmitEnvelope(c.envelope)
}
