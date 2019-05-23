package v7pushaction_test

import (
	"errors"
	"fmt"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	. "code.cloudfoundry.org/cli/actor/v7pushaction"
	"code.cloudfoundry.org/cli/actor/v7pushaction/v7pushactionfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func actualizedStreamsDrainedAndClosed(
	configStream <-chan PushPlan,
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

func buildV3Resource(name string) sharedaction.V3Resource {
	return sharedaction.V3Resource{
		FilePath:    name,
		Checksum:    ccv3.Checksum{Value: fmt.Sprintf("checksum-%s", name)},
		SizeInBytes: 6,
	}
}

var _ = Describe("Actualize", func() {
	var (
		actor *Actor

		plan            PushPlan
		fakeProgressBar *v7pushactionfakes.FakeProgressBar

		successfulChangeAppFuncCallCount int
		warningChangeAppFuncCallCount    int
		errorChangeAppFuncCallCount      int

		planStream     <-chan PushPlan
		eventStream    <-chan Event
		warningsStream <-chan Warnings
		errorStream    <-chan error

		expectedPlan PushPlan
	)

	successfulChangeAppFunc := func(pushPlan PushPlan, eStream chan<- Event, progressBar ProgressBar) (PushPlan, Warnings, error) {
		defer GinkgoRecover()

		Expect(pushPlan).To(Equal(plan))
		Expect(eStream).ToNot(BeNil())
		Expect(progressBar).To(Equal(fakeProgressBar))

		pushPlan.Application.GUID = "successful-app-guid"
		successfulChangeAppFuncCallCount++
		return pushPlan, nil, nil
	}

	warningChangeAppFunc := func(pushPlan PushPlan, eventStream chan<- Event, progressBar ProgressBar) (PushPlan, Warnings, error) {
		pushPlan.Application.GUID = "warning-app-guid"
		warningChangeAppFuncCallCount++
		return pushPlan, Warnings{"warning-1", "warning-2"}, nil
	}

	errorChangeAppFunc := func(pushPlan PushPlan, eventStream chan<- Event, progressBar ProgressBar) (PushPlan, Warnings, error) {
		pushPlan.Application.GUID = "error-app-guid"
		errorChangeAppFuncCallCount++
		return pushPlan, nil, errors.New("some error")
	}

	BeforeEach(func() {
		actor, _, _, _ = getTestPushActor()

		successfulChangeAppFuncCallCount = 0
		warningChangeAppFuncCallCount = 0
		errorChangeAppFuncCallCount = 0

		fakeProgressBar = new(v7pushactionfakes.FakeProgressBar)
		plan = PushPlan{
			Application: v7action.Application{
				Name: "some-app",
				GUID: "some-app-guid",
			},
			SpaceGUID: "some-space-guid",
		}

		expectedPlan = plan
	})

	AfterEach(func() {
		Eventually(actualizedStreamsDrainedAndClosed(planStream, eventStream, warningsStream, errorStream)).Should(BeTrue())
	})

	JustBeforeEach(func() {
		planStream, eventStream, warningsStream, errorStream = actor.Actualize(plan, fakeProgressBar)
	})

	Describe("ChangeApplicationSequence", func() {
		When("none of the ChangeApplicationSequence return errors", func() {
			BeforeEach(func() {
				actor.ChangeApplicationSequence = func(plan PushPlan) []ChangeApplicationFunc {
					return []ChangeApplicationFunc{
						successfulChangeAppFunc,
						warningChangeAppFunc,
					}
				}
			})

			It("iterates over the actor's ChangeApplicationSequence", func() {
				Eventually(warningsStream).Should(Receive(BeNil()))
				expectedPlan.Application.GUID = "successful-app-guid"
				Eventually(planStream).Should(Receive(Equal(expectedPlan)))

				Eventually(warningsStream).Should(Receive(ConsistOf("warning-1", "warning-2")))
				expectedPlan.Application.GUID = "warning-app-guid"
				Eventually(planStream).Should(Receive(Equal(expectedPlan)))

				Eventually(eventStream).Should(Receive(Equal(Complete)))

				Expect(successfulChangeAppFuncCallCount).To(Equal(1))
				Expect(warningChangeAppFuncCallCount).To(Equal(1))
				Expect(errorChangeAppFuncCallCount).To(Equal(0))
			})
		})

		When("the ChangeApplicationSequence return errors", func() {
			BeforeEach(func() {
				actor.ChangeApplicationSequence = func(plan PushPlan) []ChangeApplicationFunc {
					return []ChangeApplicationFunc{
						errorChangeAppFunc,
						successfulChangeAppFunc,
					}
				}
			})

			It("iterates over the actor's ChangeApplicationSequence", func() {
				Eventually(warningsStream).Should(Receive(BeNil()))
				Eventually(errorStream).Should(Receive(MatchError("some error")))

				Expect(successfulChangeAppFuncCallCount).To(Equal(0))
				Expect(warningChangeAppFuncCallCount).To(Equal(0))
				Expect(errorChangeAppFuncCallCount).To(Equal(1))

				Consistently(eventStream).ShouldNot(Receive(Equal(Complete)))
			})
		})
	})
})
