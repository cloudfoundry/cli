// Package dropsonde_marshaller provides a tool for marshalling Envelopes
// to Protocol Buffer messages.
//
// Use
//
// Instantiate a Marshaller and run it:
//
//		marshaller := dropsonde_marshaller.NewDropsondeMarshaller(logger)
//		inputChan := make(chan *events.Envelope) // or use a channel provided by some other source
//		outputChan := make(chan []byte)
//		go marshaller.Run(inputChan, outputChan)
//
// The marshaller self-instruments, counting the number of messages
// processed and the number of errors. These can be accessed through the Emit
// function on the marshaller.
package dropsonde_marshaller

import (
	"unicode"
	"unicode/utf8"

	"github.com/cloudfoundry/dropsonde/metrics"
	"github.com/cloudfoundry/gosteno"
	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"
)

var metricNames map[events.Envelope_EventType]string

func init() {
	metricNames = make(map[events.Envelope_EventType]string)
	for eventType, eventName := range events.Envelope_EventType_name {
		r, n := utf8.DecodeRuneInString(eventName)
		modifiedName := string(unicode.ToLower(r)) + eventName[n:]
		metricName := "dropsondeMarshaller." + modifiedName + "Received"
		metricNames[events.Envelope_EventType(eventType)] = metricName
	}
}

// A DropsondeMarshaller is an self-instrumenting tool for converting dropsonde
// Envelopes to binary (Protocol Buffer) messages.
type DropsondeMarshaller interface {
	Run(inputChan <-chan *events.Envelope, outputChan chan<- []byte)
}

// NewDropsondeMarshaller instantiates a DropsondeMarshaller and logs to the
// provided logger.
func NewDropsondeMarshaller(logger *gosteno.Logger) DropsondeMarshaller {
	messageCounts := make(map[events.Envelope_EventType]*uint64)
	for key := range events.Envelope_EventType_name {
		var count uint64
		messageCounts[events.Envelope_EventType(key)] = &count
	}
	return &dropsondeMarshaller{
		logger:        logger,
		messageCounts: messageCounts,
	}
}

type dropsondeMarshaller struct {
	logger            *gosteno.Logger
	messageCounts     map[events.Envelope_EventType]*uint64
	marshalErrorCount uint64
}

// Run reads Envelopes from inputChan, marshals them to Protocol Buffer format,
// and emits the binary messages onto outputChan. It operates one message at a
// time, and will block if outputChan is not read.
func (u *dropsondeMarshaller) Run(inputChan <-chan *events.Envelope, outputChan chan<- []byte) {
	for message := range inputChan {

		messageBytes, err := proto.Marshal(message)
		if err != nil {
			u.logger.Errorf("dropsondeMarshaller: marshal error %v", err)
			metrics.BatchIncrementCounter("dropsondeMarshaller.marshalErrors")
			continue
		}

		u.incrementMessageCount(message.GetEventType())
		outputChan <- messageBytes
	}
}

func (u *dropsondeMarshaller) incrementMessageCount(eventType events.Envelope_EventType) {
	metricName := metricNames[eventType]
	if metricName == "" {
		metricName = "dropsondeMarshaller.unknownEventTypeReceived"
	}

	metrics.BatchIncrementCounter(metricName)
}
