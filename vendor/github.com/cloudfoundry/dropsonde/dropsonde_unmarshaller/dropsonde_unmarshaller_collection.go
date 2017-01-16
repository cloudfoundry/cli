package dropsonde_unmarshaller

import (
	"log"
	"sync"

	"github.com/cloudfoundry/sonde-go/events"
)

// A DropsondeUnmarshallerCollection is a collection of DropsondeUnmarshaller instances.
type DropsondeUnmarshallerCollection struct {
	unmarshallers []*DropsondeUnmarshaller
}

// NewDropsondeUnmarshallerCollection instantiates a DropsondeUnmarshallerCollection,
// creates the specified number of DropsondeUnmarshaller instances.
func NewDropsondeUnmarshallerCollection(size int) *DropsondeUnmarshallerCollection {
	var unmarshallers []*DropsondeUnmarshaller
	for i := 0; i < size; i++ {
		unmarshallers = append(unmarshallers, NewDropsondeUnmarshaller())
	}

	log.Printf("dropsondeUnmarshallerCollection: created %v unmarshallers", size)

	return &DropsondeUnmarshallerCollection{
		unmarshallers: unmarshallers,
	}
}

// Returns the number of unmarshallers in its collection.
func (u *DropsondeUnmarshallerCollection) Size() int {
	return len(u.unmarshallers)
}

// Run calls Run on each marshaller in its collection.
// This is done in separate go routines.
func (u *DropsondeUnmarshallerCollection) Run(inputChan <-chan []byte, outputChan chan<- *events.Envelope, waitGroup *sync.WaitGroup) {
	for _, unmarshaller := range u.unmarshallers {
		go func(um *DropsondeUnmarshaller) {
			defer waitGroup.Done()
			um.Run(inputChan, outputChan)
		}(unmarshaller)
	}
}
