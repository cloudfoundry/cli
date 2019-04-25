package v7pushaction_test

import (
	"code.cloudfoundry.org/cli/actor/v7action"
	. "code.cloudfoundry.org/cli/actor/v7pushaction"
	"code.cloudfoundry.org/cli/actor/v7pushaction/v7pushactionfakes"
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SetDropletForApplication", func() {
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
			DropletGUID: "some-droplet-guid",
		}
	})

	JustBeforeEach(func() {
		events = EventFollower(func(eventStream chan<- Event) {
			_, warnings, executeErr = actor.SetDropletForApplication(paramPlan, eventStream, nil)
		})
	})

	When("setting the droplet is successful", func() {
		BeforeEach(func() {
			fakeV7Actor.SetApplicationDropletReturns(v7action.Warnings{"some-set-droplet-warning"}, nil)
		})

		It("returns a SetDropletComplete event and warnings", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(warnings).To(ConsistOf("some-set-droplet-warning"))
			Expect(events).To(ConsistOf(SettingDroplet, SetDropletComplete))

			Expect(fakeV7Actor.SetApplicationDropletCallCount()).To(Equal(1))
			appGUID, dropletGUID := fakeV7Actor.SetApplicationDropletArgsForCall(0)
			Expect(appGUID).To(Equal("some-app-guid"))
			Expect(dropletGUID).To(Equal("some-droplet-guid"))
		})
	})

	When("setting the droplet errors", func() {
		BeforeEach(func() {
			fakeV7Actor.SetApplicationDropletReturns(v7action.Warnings{"some-set-droplet-warning"}, errors.New("the climate is arid"))
		})

		It("returns an error and warnings", func() {
			Expect(executeErr).To(MatchError("the climate is arid"))
			Expect(warnings).To(ConsistOf("some-set-droplet-warning"))
			Expect(events).To(ConsistOf(SettingDroplet))
		})
	})
})
