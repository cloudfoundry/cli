package v7pushaction_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/v7action"
	. "code.cloudfoundry.org/cli/actor/v7pushaction"
	"code.cloudfoundry.org/cli/actor/v7pushaction/v7pushactionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/resources"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("StopApplication", func() {
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
			Application: resources.Application{
				GUID: "some-app-guid",
			},
		}
	})

	JustBeforeEach(func() {
		events = EventFollower(func(eventStream chan<- *PushEvent) {
			_, warnings, executeErr = actor.StopApplication(paramPlan, eventStream, nil)
		})
	})

	When("The app is running", func() {
		BeforeEach(func() {
			fakeV7Actor.StopApplicationReturns(v7action.Warnings{"some-stopping-warning"}, nil)
			paramPlan.Application.State = constant.ApplicationStarted
		})

		When("Stopping the app succeeds", func() {
			It("Uploads a package and exits", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("some-stopping-warning"))
				Expect(events).To(ConsistOf(StoppingApplication, StoppingApplicationComplete))

				Expect(fakeV7Actor.StopApplicationCallCount()).To(Equal(1))
				Expect(fakeV7Actor.StopApplicationArgsForCall(0)).To(Equal("some-app-guid"))
				Expect(fakeV7Actor.StageApplicationPackageCallCount()).To(BeZero())
			})
		})

		When("Stopping the app fails", func() {
			BeforeEach(func() {
				fakeV7Actor.StopApplicationReturns(v7action.Warnings{"some-stopping-warning"}, errors.New("bummer"))
			})

			It("returns errors and warnings", func() {
				Expect(executeErr).To(MatchError("bummer"))
				Expect(warnings).To(ConsistOf("some-stopping-warning"))
				Expect(events).To(ConsistOf(StoppingApplication))
			})
		})
	})
})
