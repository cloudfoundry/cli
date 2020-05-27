package v7pushaction_test

import (
	"errors"
	"fmt"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	. "code.cloudfoundry.org/cli/actor/v7pushaction"
	"code.cloudfoundry.org/cli/actor/v7pushaction/v7pushactionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/resources"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func streamsDrainedAndClosed(eventStream <-chan *PushEvent) bool {
	for range eventStream {
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

		eventStream <-chan *PushEvent

		expectedPlan PushPlan
	)

	successfulChangeAppFunc := func(pushPlan PushPlan, eStream chan<- *PushEvent, progressBar ProgressBar) (PushPlan, Warnings, error) {
		defer GinkgoRecover()

		Expect(pushPlan).To(Equal(plan))
		Expect(eStream).ToNot(BeNil())
		Expect(progressBar).To(Equal(fakeProgressBar))

		pushPlan.Application.GUID = "successful-app-guid"
		successfulChangeAppFuncCallCount++
		return pushPlan, nil, nil
	}

	warningChangeAppFunc := func(pushPlan PushPlan, eventStream chan<- *PushEvent, progressBar ProgressBar) (PushPlan, Warnings, error) {
		pushPlan.Application.GUID = "warning-app-guid"
		warningChangeAppFuncCallCount++
		return pushPlan, Warnings{"warning-1", "warning-2"}, nil
	}

	errorChangeAppFunc := func(pushPlan PushPlan, eventStream chan<- *PushEvent, progressBar ProgressBar) (PushPlan, Warnings, error) {
		pushPlan.Application.GUID = "error-app-guid"
		errorChangeAppFuncCallCount++
		return pushPlan, nil, errors.New("some error")
	}

	BeforeEach(func() {
		actor, _, _ = getTestPushActor()

		successfulChangeAppFuncCallCount = 0
		warningChangeAppFuncCallCount = 0
		errorChangeAppFuncCallCount = 0

		fakeProgressBar = new(v7pushactionfakes.FakeProgressBar)
		plan = PushPlan{
			Application: resources.Application{
				Name: "some-app",
				GUID: "some-app-guid",
			},
			SpaceGUID: "some-space-guid",
		}

		expectedPlan = plan
	})

	AfterEach(func() {
		Eventually(streamsDrainedAndClosed(eventStream)).Should(BeTrue())
	})

	JustBeforeEach(func() {
		eventStream = actor.Actualize(plan, fakeProgressBar)
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
				expectedPlan.Application.GUID = "successful-app-guid"
				Eventually(eventStream).Should(Receive(Equal(&PushEvent{Plan: expectedPlan})))

				expectedPlan.Application.GUID = "warning-app-guid"
				Eventually(eventStream).Should(Receive(Equal(&PushEvent{Plan: expectedPlan, Warnings: Warnings{"warning-1", "warning-2"}})))

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
				expectedPlan.Application.GUID = "error-app-guid"
				Eventually(eventStream).Should(Receive(Equal(&PushEvent{Plan: expectedPlan, Err: errors.New("some error")})))

				Expect(successfulChangeAppFuncCallCount).To(Equal(0))
				Expect(warningChangeAppFuncCallCount).To(Equal(0))
				Expect(errorChangeAppFuncCallCount).To(Equal(1))
			})
		})
	})
})
