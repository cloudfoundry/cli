package v7pushaction_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/v7action"
	. "code.cloudfoundry.org/cli/actor/v7pushaction"
	"code.cloudfoundry.org/cli/actor/v7pushaction/v7pushactionfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UnmapRoutesFromApplication", func() {
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
			ApplicationRoutes: []v7action.Route{
				{GUID: "route1-guid"},
				{GUID: "route2-guid"},
				{GUID: "route3-guid"},
			},
		}

		fakeV7Actor.GetRouteDestinationByAppGUIDReturnsOnCall(0,
			v7action.RouteDestination{GUID: "destination1-guid"},
			v7action.Warnings{"get-route1-destination-warning"},
			nil,
		)
		fakeV7Actor.GetRouteDestinationByAppGUIDReturnsOnCall(1,
			v7action.RouteDestination{GUID: "destination2-guid"},
			v7action.Warnings{"get-route2-destination-warning"},
			nil,
		)

		fakeV7Actor.GetRouteDestinationByAppGUIDReturnsOnCall(2,
			v7action.RouteDestination{GUID: "destination3-guid"},
			v7action.Warnings{"get-route3-destination-warning"},
			nil,
		)

		fakeV7Actor.UnmapRouteReturnsOnCall(0,
			v7action.Warnings{"unmap-route1-warning"},
			nil,
		)
		fakeV7Actor.UnmapRouteReturnsOnCall(1,
			v7action.Warnings{"unmap-route2-warning"},
			nil,
		)
		fakeV7Actor.UnmapRouteReturnsOnCall(2,
			v7action.Warnings{"unmap-route3-warning"},
			nil,
		)
	})

	JustBeforeEach(func() {
		events = EventFollower(func(eventStream chan<- *PushEvent) {
			_, warnings, executeErr = actor.UnmapRoutesFromApplication(paramPlan, eventStream, nil)
		})
	})

	It("returns errors and warnings", func() {
		Expect(executeErr).NotTo(HaveOccurred())
		Expect(warnings).To(ConsistOf(
			"get-route1-destination-warning",
			"unmap-route1-warning",
			"get-route2-destination-warning",
			"unmap-route2-warning",
			"get-route3-destination-warning",
			"unmap-route3-warning",
		))
		Expect(events).To(ConsistOf(UnmappingRoutes))
	})

	When("getting routes destination fails", func() {
		BeforeEach(func() {
			fakeV7Actor.GetRouteDestinationByAppGUIDReturnsOnCall(0,
				v7action.RouteDestination{},
				v7action.Warnings{"get-route1-destination-warning"},
				errors.New("get-route1-destination-error"),
			)
		})

		It("returns errors and warnings", func() {
			Expect(executeErr).To(MatchError("get-route1-destination-error"))
			Expect(warnings).To(ConsistOf(
				"get-route1-destination-warning",
			))
			Expect(events).To(ConsistOf(UnmappingRoutes))
		})
	})

	When("unmapping route fails", func() {
		BeforeEach(func() {
			fakeV7Actor.UnmapRouteReturnsOnCall(1,
				v7action.Warnings{"unmap-route2-warning"},
				errors.New("unmap-route2-error"),
			)
		})

		It("returns errors and warnings", func() {
			Expect(executeErr).To(MatchError("unmap-route2-error"))
			Expect(warnings).To(ConsistOf(
				"get-route1-destination-warning",
				"unmap-route1-warning",
				"get-route2-destination-warning",
				"unmap-route2-warning",
			))
			Expect(events).To(ConsistOf(UnmappingRoutes))
		})
	})
})
