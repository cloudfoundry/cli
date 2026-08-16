package v7pushaction_test

import (
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	. "code.cloudfoundry.org/cli/v9/actor/v7pushaction"
	"code.cloudfoundry.org/cli/v9/actor/v7pushaction/v7pushactionfakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPushAction(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "V7 Push Actions Suite")
}

var _ = BeforeSuite(func() {
	// Suppress log output during tests. This equates with setting to panic-level severity in older logging libraries
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
	SetDefaultEventuallyTimeout(3 * time.Second)
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

func EventFollower(wrapper func(eventStream chan<- *PushEvent)) []Event {
	eventStream := make(chan *PushEvent)
	closed := make(chan bool)

	var events []Event

	go func() {
		for {
			event, ok := <-eventStream
			if !ok {
				close(closed)
				return
			}
			events = append(events, event.Event)
		}
	}()

	wrapper(eventStream)
	close(eventStream)

	<-closed
	return events
}
