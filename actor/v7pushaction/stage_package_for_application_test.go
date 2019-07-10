package v7pushaction_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/v7action"
	. "code.cloudfoundry.org/cli/actor/v7pushaction"
	"code.cloudfoundry.org/cli/actor/v7pushaction/v7pushactionfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("StagePackageForApplication", func() {
	var (
		actor       *Actor
		fakeV7Actor *v7pushactionfakes.FakeV7Actor

		paramPlan PushPlan

		returnedPushPlan PushPlan
		warnings         Warnings
		executeErr       error

		events []Event
	)

	BeforeEach(func() {
		actor, fakeV7Actor, _ = getTestPushActor()

		paramPlan = PushPlan{
			Application: v7action.Application{
				GUID: "some-app-guid",
			},
			PackageGUID: "some-pkg-guid",
		}
	})

	JustBeforeEach(func() {
		events = EventFollower(func(eventStream chan<- Event) {
			returnedPushPlan, warnings, executeErr = actor.StagePackageForApplication(paramPlan, eventStream, nil)
		})
	})

	Describe("staging package", func() {
		It("stages the application using the package guid", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(fakeV7Actor.StageApplicationPackageCallCount()).To(Equal(1))
			Expect(fakeV7Actor.StageApplicationPackageArgsForCall(0)).To(Equal("some-pkg-guid"))
		})

		When("staging is successful", func() {
			BeforeEach(func() {
				fakeV7Actor.StageApplicationPackageReturns(v7action.Build{GUID: "some-build-guid"}, v7action.Warnings{"some-staging-warning"}, nil)
			})

			It("returns a polling build event and warnings", func() {
				Expect(events).To(ConsistOf(StartingStaging, PollingBuild, StagingComplete))
				Expect(ConsistOf("some-staging-warning"))
			})
		})

		When("staging errors", func() {
			BeforeEach(func() {
				fakeV7Actor.StageApplicationPackageReturns(v7action.Build{}, v7action.Warnings{"some-staging-warning"}, errors.New("ahhh, i failed"))
			})

			It("returns errors and warnings", func() {
				Expect(events).To(ConsistOf(StartingStaging))
				Expect(ConsistOf("some-staging-warning"))
				Expect(executeErr).To(MatchError("ahhh, i failed"))
			})
		})
	})

	Describe("polling build", func() {
		When("the the polling is successful", func() {
			BeforeEach(func() {
				fakeV7Actor.PollBuildReturns(v7action.Droplet{GUID: "some-droplet-guid"}, v7action.Warnings{"some-poll-build-warning"}, nil)
			})

			It("returns a staging complete event and warnings", func() {
				Expect(events).To(ConsistOf(StartingStaging, PollingBuild, StagingComplete))
				Expect(warnings).To(ConsistOf("some-poll-build-warning"))

				Expect(fakeV7Actor.PollBuildCallCount()).To(Equal(1))
			})

			It("sets the droplet GUID on push plan", func() {
				Expect(returnedPushPlan.DropletGUID).To(Equal("some-droplet-guid"))
			})
		})

		When("the the polling returns an error", func() {
			var someErr error

			BeforeEach(func() {
				someErr = errors.New("I AM A BANANA")
				fakeV7Actor.PollBuildReturns(v7action.Droplet{}, v7action.Warnings{"some-poll-build-warning"}, someErr)
			})

			It("returns errors and warnings", func() {
				Expect(events).To(ConsistOf(StartingStaging, PollingBuild))
				Expect(warnings).To(ConsistOf("some-poll-build-warning"))
				Expect(executeErr).To(MatchError(someErr))
			})
		})
	})
})
