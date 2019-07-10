package v7pushaction_test

import (
	"os"
	"time"

	"testing"

	. "code.cloudfoundry.org/cli/actor/v7pushaction"
	"code.cloudfoundry.org/cli/actor/v7pushaction/v7pushactionfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	log "github.com/sirupsen/logrus"
)

func TestPushAction(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "V7 Push Actions Suite")
}

var _ = BeforeEach(func() {
	SetDefaultEventuallyTimeout(3 * time.Second)
	log.SetLevel(log.PanicLevel)
})

func getCurrentDir() string {
	pwd, err := os.Getwd()
	Expect(err).NotTo(HaveOccurred())
	return pwd
}

func getTestPushActor() (*Actor, *v7pushactionfakes.FakeV7Actor, *v7pushactionfakes.FakeSharedActor) {
	fakeV7Actor := new(v7pushactionfakes.FakeV7Actor)
	fakeSharedActor := new(v7pushactionfakes.FakeSharedActor)
	actor := NewActor(fakeV7Actor, fakeSharedActor)
	return actor, fakeV7Actor, fakeSharedActor
}

func EventFollower(wrapper func(eventStream chan<- Event)) []Event {
	eventStream := make(chan Event)
	closed := make(chan bool)

	var events []Event

	go func() {
		for {
			event, ok := <-eventStream
			if !ok {
				close(closed)
				return
			}
			events = append(events, event)
		}
	}()

	wrapper(eventStream)
	close(eventStream)

	<-closed
	return events
}
