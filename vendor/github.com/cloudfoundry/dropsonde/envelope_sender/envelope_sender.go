package envelope_sender

import "github.com/cloudfoundry/sonde-go/events"

type EnvelopeEmitter interface {
	EmitEnvelope(*events.Envelope) error
}

// A EnvelopeSender emits envelopes.
type EnvelopeSender struct {
	emitter EnvelopeEmitter
}

// NewEnvelopeSender instantiates a EnvelopeSender with the given EventEmitter.
func NewEnvelopeSender(emitter EnvelopeEmitter) *EnvelopeSender {
	return &EnvelopeSender{emitter: emitter}
}

// SendEnvelope sends the given envelope.
// Returns an error if one occurs while sending the envelope.
func (ms *EnvelopeSender) SendEnvelope(envelope *events.Envelope) error {
	return ms.emitter.EmitEnvelope(envelope)
}
