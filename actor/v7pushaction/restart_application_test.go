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
		actor, fakeV7Actor, _ = getTestPushActor()

		paramPlan = PushPlan{
			Application: v7action.Application{
				GUID: "some-app-guid",
			},
		}
	})

	JustBeforeEach(func() {
		events = EventFollower(func(eventStream chan<- *PushEvent) {
			_, warnings, executeErr = actor.RestartApplication(paramPlan, eventStream, nil)
		})
	})

	It("Restarts the app", func() {
		Expect(fakeV7Actor.RestartApplicationCallCount()).To(Equal(1))
		Expect(fakeV7Actor.RestartApplicationArgsForCall(0)).To(Equal("some-app-guid"))
	})

	When("Restarting the app succeeds", func() {
		BeforeEach(func() {
			fakeV7Actor.PollStartCalls(func(s string, b bool, handleInstanceDetails func(string)) (warnings v7action.Warnings, err error) {
				handleInstanceDetails("Instances starting...")
				return nil, nil
			})

			fakeV7Actor.RestartApplicationReturns(v7action.Warnings{"some-restarting-warning"}, nil)
			paramPlan.Application.State = constant.ApplicationStarted
		})

		When("the noWait flag is set", func() {
			BeforeEach(func() {
				paramPlan.NoWait = true
			})

			It("calls PollStart with true", func() {
				Expect(fakeV7Actor.PollStartCallCount()).To(Equal(1))
				actualAppGUID, givenNoWait, _ := fakeV7Actor.PollStartArgsForCall(0)
				Expect(givenNoWait).To(Equal(true))
				Expect(actualAppGUID).To(Equal("some-app-guid"))
			})
		})

		It("calls pollStart", func() {
			Expect(fakeV7Actor.PollStartCallCount()).To(Equal(1))
			actualAppGUID, givenNoWait, _ := fakeV7Actor.PollStartArgsForCall(0)
			Expect(givenNoWait).To(Equal(false))
			Expect(actualAppGUID).To(Equal("some-app-guid"))
			Expect(events).To(ConsistOf(RestartingApplication, InstanceDetails, RestartingApplicationComplete))
		})

		When("pollStart errors", func() {
			BeforeEach(func() {
				fakeV7Actor.PollStartReturns(
					v7action.Warnings{"poll-start-warning"},
					errors.New("poll-start-error"),
				)
			})

			It("returns errors and warnings", func() {
				Expect(executeErr).To(MatchError("poll-start-error"))
				Expect(warnings).To(ConsistOf("some-restarting-warning", "poll-start-warning"))
			})

		})

		When("pollStart succeeds", func() {
			BeforeEach(func() {
				fakeV7Actor.PollStartReturns(
					v7action.Warnings{"poll-start-warning"},
					nil,
				)
			})

			It("Uploads a package and exits", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("some-restarting-warning", "poll-start-warning"))
				Expect(events).To(ConsistOf(RestartingApplication, RestartingApplicationComplete))
			})
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
