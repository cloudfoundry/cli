package fake

import (
	"sync"

	"github.com/cloudfoundry/sonde-go/events"
)

type Message struct {
	Event  events.Event
	Origin string
}

type FakeEventEmitter struct {
	ReturnError error
	messages    []Message
	envelopes   []*events.Envelope
	Origin      string
	isClosed    bool
	sync.RWMutex
}

func NewFakeEventEmitter(origin string) *FakeEventEmitter {
	return &FakeEventEmitter{Origin: origin}
}

func (f *FakeEventEmitter) Emit(e events.Event) error {

	f.Lock()
	defer f.Unlock()

	if f.ReturnError != nil {
		err := f.ReturnError
		f.ReturnError = nil
		return err
	}

	f.messages = append(f.messages, Message{e, f.Origin})
	return nil
}

func (f *FakeEventEmitter) EmitEnvelope(e *events.Envelope) error {

	f.Lock()
	defer f.Unlock()

	if f.ReturnError != nil {
		err := f.ReturnError
		f.ReturnError = nil
		return err
	}

	f.envelopes = append(f.envelopes, e)
	return nil
}

func (f *FakeEventEmitter) GetMessages() (messages []Message) {
	f.Lock()
	defer f.Unlock()

	messages = make([]Message, len(f.messages))
	copy(messages, f.messages)
	return
}

func (f *FakeEventEmitter) GetEnvelopes() (envelopes []*events.Envelope) {
	f.Lock()
	defer f.Unlock()

	envelopes = make([]*events.Envelope, len(f.envelopes))
	copy(envelopes, f.envelopes)
	return
}

func (f *FakeEventEmitter) GetEvents() []events.Event {
	messages := f.GetMessages()
	events := []events.Event{}
	for _, msg := range messages {
		events = append(events, msg.Event)
	}
	return events
}

func (f *FakeEventEmitter) Close() {
	f.Lock()
	defer f.Unlock()
	f.isClosed = true
}

func (f *FakeEventEmitter) IsClosed() bool {
	f.RLock()
	defer f.RUnlock()
	return f.isClosed
}

func (f *FakeEventEmitter) Reset() {
	f.Lock()
	defer f.Unlock()

	f.isClosed = false
	f.messages = []Message{}
	f.ReturnError = nil
}
