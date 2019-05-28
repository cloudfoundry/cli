package v7pushaction_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/v7action"
	. "code.cloudfoundry.org/cli/actor/v7pushaction"
	"code.cloudfoundry.org/cli/actor/v7pushaction/v7pushactionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RestartApplication", func() {
	var (
		actor       *Actor
		fakeV7Actor *v7pushactionfakes.FakeV7Actor

		paramPlan PushPlan

		warnings   Warnings
		executeErr error

		events []Event
	)

	BeforeEach(func() {
		actor, _, fakeV7Actor, _ = getTestPushActor()

		paramPlan = PushPlan{
			Application: v7action.Application{
				GUID: "some-app-guid",
			},
		}
	})

	JustBeforeEach(func() {
		events = EventFollower(func(eventStream chan<- Event) {
			_, warnings, executeErr = actor.RestartApplication(paramPlan, eventStream, nil)
		})
	})

	When("The app is running", func() {
		BeforeEach(func() {
			fakeV7Actor.RestartApplicationReturns(v7action.Warnings{"some-restarting-warning"}, nil)
			paramPlan.Application.State = constant.ApplicationStarted
		})

		When("Restarting the app succeeds", func() {
			It("Uploads a package and exits", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("some-restarting-warning"))
				Expect(events).To(ConsistOf(RestartingApplication, RestartingApplicationComplete))

				Expect(fakeV7Actor.RestartApplicationCallCount()).To(Equal(1))
				Expect(fakeV7Actor.RestartApplicationArgsForCall(0)).To(Equal("some-app-guid"))
				Expect(fakeV7Actor.StageApplicationPackageCallCount()).To(BeZero())
			})
		})

		When("Restarting the app fails", func() {
			BeforeEach(func() {
				fakeV7Actor.RestartApplicationReturns(v7action.Warnings{"some-restarting-warning"}, errors.New("bummer"))
			})

			It("returns errors and warnings", func() {
				Expect(executeErr).To(MatchError("bummer"))
				Expect(warnings).To(ConsistOf("some-restarting-warning"))
				Expect(events).To(ConsistOf(RestartingApplication))
			})
		})
	})
})
