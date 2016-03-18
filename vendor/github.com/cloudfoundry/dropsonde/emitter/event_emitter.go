package emitter

import (
	"fmt"

	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"
)

type EventEmitter interface {
	Emit(events.Event) error
	EmitEnvelope(*events.Envelope) error
	Close()
}

type eventEmitter struct {
	innerEmitter ByteEmitter
	origin       string
}

func NewEventEmitter(byteEmitter ByteEmitter, origin string) EventEmitter {
	return &eventEmitter{innerEmitter: byteEmitter, origin: origin}
}

func (e *eventEmitter) Emit(event events.Event) error {
	envelope, err := Wrap(event, e.origin)
	if err != nil {
		return fmt.Errorf("Wrap: %v", err)
	}

	return e.EmitEnvelope(envelope)
}

func (e *eventEmitter) EmitEnvelope(envelope *events.Envelope) error {
	data, err := proto.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("Marshal: %v", err)
	}

	return e.innerEmitter.Emit(data)
}

func (e *eventEmitter) Close() {
	e.innerEmitter.Close()
}
