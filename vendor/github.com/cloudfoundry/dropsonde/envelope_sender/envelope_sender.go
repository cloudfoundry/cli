package envelope_sender

import (
	"github.com/cloudfoundry/dropsonde/emitter"
	"github.com/cloudfoundry/sonde-go/events"
)

// A EnvelopeSender emits envelopes.
type EnvelopeSender interface {
	SendEnvelope(*events.Envelope) error
}

type envelopeSender struct {
	eventEmitter emitter.EventEmitter
}

// NewEnvelopeSender instantiates a envelopeSender with the given EventEmitter.
func NewEnvelopeSender(eventEmitter emitter.EventEmitter) EnvelopeSender {
	return &envelopeSender{eventEmitter: eventEmitter}
}

// SendEnvelope sends the given envelope.
// Returns an error if one occurs while sending the envelope.
func (ms *envelopeSender) SendEnvelope(envelope *events.Envelope) error {
	return ms.eventEmitter.EmitEnvelope(envelope)
}
