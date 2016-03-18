// Package envelopes provides a simple API for sending dropsonde envelopes
// through the dropsonde system.
//
// Use
//
// See the documentation for package dropsonde for configuration details.
//
// Importing package dropsonde and initializing will initial this package.
// To send envelopes use
//
//		envelopes.SendEnvelope(envelope)
//
package envelopes

import (
	"github.com/cloudfoundry/dropsonde/envelope_sender"

	"github.com/cloudfoundry/sonde-go/events"
)

var envelopeSender envelope_sender.EnvelopeSender

// Initialize prepares the envelopes package for use with the automatic Emitter.
func Initialize(es envelope_sender.EnvelopeSender) {
	envelopeSender = es
}

// SendEnvelope sends the given Envelope.
func SendEnvelope(envelope *events.Envelope) error {
	if envelopeSender == nil {
		return nil
	}
	return envelopeSender.SendEnvelope(envelope)
}
