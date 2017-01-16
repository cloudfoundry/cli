package emitter

import (
	"fmt"

	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"
)

type ByteEmitter interface {
	Emit([]byte) error
	Close()
}

type EventEmitter struct {
	innerEmitter ByteEmitter
	origin       string
}

func NewEventEmitter(byteEmitter ByteEmitter, origin string) *EventEmitter {
	return &EventEmitter{innerEmitter: byteEmitter, origin: origin}
}

func (e *EventEmitter) Origin() string {
	return e.origin
}

func (e *EventEmitter) Emit(event events.Event) error {
	envelope, err := Wrap(event, e.origin)
	if err != nil {
		return fmt.Errorf("Wrap: %v", err)
	}

	return e.EmitEnvelope(envelope)
}

func (e *EventEmitter) EmitEnvelope(envelope *events.Envelope) error {
	data, err := proto.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("Marshal: %v", err)
	}

	return e.innerEmitter.Emit(data)
}

func (e *EventEmitter) Close() {
	e.innerEmitter.Close()
}
