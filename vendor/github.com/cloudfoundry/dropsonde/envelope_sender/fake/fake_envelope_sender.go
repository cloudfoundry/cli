package fake

import (
	"sync"

	"github.com/cloudfoundry/sonde-go/events"
)

type FakeEnvelopeSender struct {
	envelopes []*events.Envelope
	sync.RWMutex
}

func NewFakeEnvelopeSender() *FakeEnvelopeSender {
	return &FakeEnvelopeSender{}
}

func (fms *FakeEnvelopeSender) SendEnvelope(envelope *events.Envelope) error {
	fms.Lock()
	defer fms.Unlock()
	fms.envelopes = append(fms.envelopes, envelope)

	return nil
}

func (fms *FakeEnvelopeSender) GetEnvelopes() []*events.Envelope {
	fms.RLock()
	defer fms.RUnlock()

	return fms.envelopes
}
