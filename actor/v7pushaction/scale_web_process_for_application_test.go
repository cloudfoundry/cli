package v7pushaction_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/v7action"
	. "code.cloudfoundry.org/cli/actor/v7pushaction"
	"code.cloudfoundry.org/cli/actor/v7pushaction/v7pushactionfakes"
	"code.cloudfoundry.org/cli/types"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("ScaleWebProcessForApplication", func() {
	var (
		actor       *Actor
		fakeV7Actor *v7pushactionfakes.FakeV7Actor

		paramPlan PushPlan

		warnings   Warnings
		executeErr error

		events []Event
	)

	BeforeEach(func() {
		actor, fakeV7Actor, _ = getTestPushActor()

		paramPlan = PushPlan{
			Application: v7action.Application{
				GUID: "some-app-guid",
			},
		}
	})

	JustBeforeEach(func() {
		events = EventFollower(func(eventStream chan<- Event) {
			_, warnings, executeErr = actor.ScaleWebProcessForApplication(paramPlan, eventStream, nil)
		})
	})

	When("a scale override is passed", func() {
		When("the scale is successful", func() {
			var memory types.NullUint64

			BeforeEach(func() {
				paramPlan.Application.GUID = "some-app-guid"

				paramPlan.ScaleWebProcessNeedsUpdate = true
				memory = types.NullUint64{IsSet: true, Value: 2048}
				paramPlan.ScaleWebProcess = v7action.Process{
					MemoryInMB: memory,
				}

				fakeV7Actor.ScaleProcessByApplicationReturns(v7action.Warnings{"scaling-warnings"}, nil)
				fakeV7Actor.UpdateApplicationReturns(
					v7action.Application{
						Name: "some-app",
						GUID: paramPlan.Application.GUID,
					},
					v7action.Warnings{"some-app-update-warnings"},
					nil)
			})

			It("returns warnings and continues", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("scaling-warnings"))
				Expect(events).To(ConsistOf(ScaleWebProcess, ScaleWebProcessComplete))

				Expect(fakeV7Actor.ScaleProcessByApplicationCallCount()).To(Equal(1))
				passedAppGUID, passedProcess := fakeV7Actor.ScaleProcessByApplicationArgsForCall(0)
				Expect(passedAppGUID).To(Equal("some-app-guid"))
				Expect(passedProcess).To(MatchFields(IgnoreExtras,
					Fields{
						"MemoryInMB": Equal(memory),
					}))
			})
		})

		When("the scale errors", func() {
			var expectedErr error

			BeforeEach(func() {
				paramPlan.ScaleWebProcessNeedsUpdate = true

				expectedErr = errors.New("nopes")
				fakeV7Actor.ScaleProcessByApplicationReturns(v7action.Warnings{"scaling-warnings"}, expectedErr)
			})

			It("returns warnings and an error", func() {
				Expect(executeErr).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("scaling-warnings"))
				Expect(events).To(ConsistOf(ScaleWebProcess))
			})
		})
	})

	When("a scale override is not provided", func() {
		It("should not scale the application", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(events).To(BeEmpty())
			Expect(fakeV7Actor.ScaleProcessByApplicationCallCount()).To(Equal(0))
		})
	})
})
