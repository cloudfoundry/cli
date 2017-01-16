package emitter

import (
	"errors"
	"time"

	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"
)

var ErrorMissingOrigin = errors.New("Event not emitted due to missing origin information")
var ErrorUnknownEventType = errors.New("Cannot create envelope for unknown event type")

func Wrap(event events.Event, origin string) (*events.Envelope, error) {
	if origin == "" {
		return nil, ErrorMissingOrigin
	}

	envelope := &events.Envelope{Origin: proto.String(origin), Timestamp: proto.Int64(time.Now().UnixNano())}

	switch event := event.(type) {
	case *events.HttpStartStop:
		envelope.EventType = events.Envelope_HttpStartStop.Enum()
		envelope.HttpStartStop = event
	case *events.ValueMetric:
		envelope.EventType = events.Envelope_ValueMetric.Enum()
		envelope.ValueMetric = event
	case *events.CounterEvent:
		envelope.EventType = events.Envelope_CounterEvent.Enum()
		envelope.CounterEvent = event
	case *events.LogMessage:
		envelope.EventType = events.Envelope_LogMessage.Enum()
		envelope.LogMessage = event
	case *events.ContainerMetric:
		envelope.EventType = events.Envelope_ContainerMetric.Enum()
		envelope.ContainerMetric = event
	default:
		return nil, ErrorUnknownEventType
	}

	return envelope, nil
}
