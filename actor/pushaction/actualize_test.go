package pushaction_test

import (
	"errors"
	"time"

	. "code.cloudfoundry.org/cli/actor/pushaction"
	"code.cloudfoundry.org/cli/actor/pushaction/pushactionfakes"
	"code.cloudfoundry.org/cli/actor/v3action"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

func actualizedStreamsDrainedAndClosed(
	configStream <-chan PushState,
	eventStream <-chan Event,
	warningsStream <-chan Warnings,
	errorStream <-chan error,
) bool {
	var configStreamClosed, eventStreamClosed, warningsStreamClosed, errorStreamClosed bool
	for {
		select {
		case _, ok := <-configStream:
			if !ok {
				configStreamClosed = true
			}
		case _, ok := <-eventStream:
			if !ok {
				eventStreamClosed = true
			}
		case _, ok := <-warningsStream:
			if !ok {
				warningsStreamClosed = true
			}
		case _, ok := <-errorStream:
			if !ok {
				errorStreamClosed = true
			}
		}
		if configStreamClosed && eventStreamClosed && warningsStreamClosed && errorStreamClosed {
			break
		}
	}
	return true
}

// TODO: for refactor: We can use the following style of code to validate that
// each event is received in a specific order

// Expect(nextEvent()).Should(Equal(SettingUpApplication))
// Expect(nextEvent()).Should(Equal(CreatingApplication))
// Expect(nextEvent()).Should(Equal(...))
// Expect(nextEvent()).Should(Equal(...))
// Expect(nextEvent()).Should(Equal(...))
func getNextEvent(c <-chan PushState, e <-chan Event, w <-chan Warnings) func() Event {
	timeOut := time.Tick(500 * time.Millisecond)

	return func() Event {
		for {
			select {
			case <-c:
			case event, ok := <-e:
				if ok {
					return event
				}
				return ""
			case <-w:
			case <-timeOut:
				return ""
			}
		}
	}
}

var _ = Describe("Actualize", func() {
	var (
		actor           *Actor
		fakeV2Actor     *pushactionfakes.FakeV2Actor
		fakeV3Actor     *pushactionfakes.FakeV3Actor
		fakeSharedActor *pushactionfakes.FakeSharedActor

		state           PushState
		fakeProgressBar *pushactionfakes.FakeProgressBar

		stateStream    <-chan PushState
		eventStream    <-chan Event
		warningsStream <-chan Warnings
		errorStream    <-chan error
	)

	BeforeEach(func() {
		fakeV2Actor = new(pushactionfakes.FakeV2Actor)
		fakeV3Actor = new(pushactionfakes.FakeV3Actor)
		fakeSharedActor = new(pushactionfakes.FakeSharedActor)
		actor = NewActor(fakeV2Actor, fakeV3Actor, fakeSharedActor)

		fakeProgressBar = new(pushactionfakes.FakeProgressBar)
		state = PushState{
			Application: v3action.Application{
				Name: "some-app",
			},
			SpaceGUID: "some-space-guid",
		}
	})

	AfterEach(func() {
		Eventually(actualizedStreamsDrainedAndClosed(stateStream, eventStream, warningsStream, errorStream)).Should(BeTrue())
	})

	JustBeforeEach(func() {
		stateStream, eventStream, warningsStream, errorStream = actor.Actualize(state, fakeProgressBar)
	})

	Describe("application creation", func() {
		Context("when the application exists", func() {
			BeforeEach(func() {
				state.Application.GUID = "some-app-guid"
			})

			It("returns a skipped app creation event", func() {
				Eventually(getNextEvent(stateStream, eventStream, warningsStream)).Should(Equal(SkipingApplicationCreation))

				Eventually(stateStream).Should(Receive(MatchFields(IgnoreExtras,
					Fields{
						"Application": Equal(v3action.Application{
							Name: "some-app",
							GUID: "some-app-guid",
						}),
					})))

				Consistently(fakeV3Actor.CreateApplicationInSpaceCallCount).Should(Equal(0))
			})
		})

		Context("when the application does not exist", func() {
			Context("when the creation is successful", func() {
				var expectedApp v3action.Application

				BeforeEach(func() {
					expectedApp = v3action.Application{
						GUID: "some-app-guid",
						Name: "some-app",
					}

					fakeV3Actor.CreateApplicationInSpaceReturns(expectedApp, v3action.Warnings{"some-app-warnings"}, nil)
				})

				It("returns an app created event, warnings, and updated state", func() {
					Eventually(warningsStream).Should(Receive(ConsistOf("some-app-warnings")))
					Eventually(getNextEvent(stateStream, eventStream, warningsStream)).Should(Equal(CreatedApplication))
					Eventually(stateStream).Should(Receive(MatchFields(IgnoreExtras,
						Fields{
							"Application": Equal(expectedApp),
						})))
				})

				It("creates the application", func() {
					Eventually(fakeV3Actor.CreateApplicationInSpaceCallCount).Should(Equal(1))
					passedApp, passedSpaceGUID := fakeV3Actor.CreateApplicationInSpaceArgsForCall(0)
					Expect(passedApp).To(Equal(state.Application))
					Expect(passedSpaceGUID).To(Equal(state.SpaceGUID))
				})
			})

			Context("when the creation errors", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("SPICY!!")

					fakeV3Actor.CreateApplicationInSpaceReturns(v3action.Application{}, v3action.Warnings{"some-app-warnings"}, expectedErr)
				})

				It("returns warnings and error", func() {
					Eventually(warningsStream).Should(Receive(ConsistOf("some-app-warnings")))
					Eventually(errorStream).Should(Receive(MatchError(expectedErr)))
				})
			})
		})
	})

	Context("when all operations are finished", func() {
		It("returns a complete event", func() {
			Eventually(getNextEvent(stateStream, eventStream, warningsStream)).Should(Equal(Complete))
		})
	})
})
