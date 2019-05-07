package v7pushaction_test

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v7action"
	. "code.cloudfoundry.org/cli/actor/v7pushaction"
	"code.cloudfoundry.org/cli/actor/v7pushaction/v7pushactionfakes"
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UpdateWebProcessForApplication", func() {
	var (
		actor       *Actor
		fakeV2Actor *v7pushactionfakes.FakeV2Actor
		fakeV7Actor *v7pushactionfakes.FakeV7Actor

		paramPlan PushPlan

		warnings   Warnings
		executeErr error

		events []Event
	)

	BeforeEach(func() {
		actor, fakeV2Actor, fakeV7Actor, _ = getTestPushActor()

		paramPlan = PushPlan{
			Application: v7action.Application{
				GUID: "some-app-guid",
			},
		}

		fakeV7Actor.GetDefaultDomainReturns(
			v7action.Domain{
				GUID: "some-domain-guid",
				Name: "some-domain",
			},
			v7action.Warnings{"domain-warning"},
			nil,
		)
	})

	JustBeforeEach(func() {
		events = EventFollower(func(eventStream chan<- Event) {
			_, warnings, executeErr = actor.UpdateRoutesForApplication(paramPlan, eventStream, nil)
		})
	})

	When("creating a default route", func() {
		BeforeEach(func() {
			paramPlan.SkipRouteCreation = false
		})

		When("route creation and mapping is successful", func() {
			BeforeEach(func() {
				fakeV2Actor.FindRouteBoundToSpaceWithSettingsReturns(
					v2action.Route{},
					v2action.Warnings{"route-warning"},
					actionerror.RouteNotFoundError{},
				)

				fakeV2Actor.CreateRouteReturns(
					v2action.Route{
						GUID: "some-route-guid",
						Host: "some-app",
						Domain: v2action.Domain{
							Name: "some-domain",
							GUID: "some-domain-guid",
						},
						SpaceGUID: "some-space-guid",
					},
					v2action.Warnings{"route-create-warning"},
					nil,
				)

				fakeV2Actor.MapRouteToApplicationReturns(
					v2action.Warnings{"map-warning"},
					nil,
				)
			})

			It("creates the route, maps it to the app, and returns any warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("domain-warning", "route-warning", "route-create-warning", "map-warning"))
				Expect(events).To(ConsistOf(CreatingAndMappingRoutes, CreatedRoutes))
			})
		})

		When("route creation and mapping errors", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("some route error")
				fakeV7Actor.GetDefaultDomainReturns(
					v7action.Domain{
						GUID: "some-domain-guid",
						Name: "some-domain",
					},
					v7action.Warnings{"domain-warning"},
					expectedErr,
				)
			})

			It("returns errors and warnings", func() {
				Expect(executeErr).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("domain-warning"))
				Expect(events).To(ConsistOf(CreatingAndMappingRoutes))
			})
		})
	})

	When("skipping default route creation from manifest", func() {
		BeforeEach(func() {
			paramPlan.SkipRouteCreation = true
		})

		It("never attempts to create a route", func() {
			Expect(events).To(BeEmpty())
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(fakeV2Actor.GetApplicationRoutesCallCount()).To(BeZero())
			Expect(fakeV2Actor.CreateRouteCallCount()).To(BeZero())
		})
	})

	When("skipping default route creation from overrides", func() {
		BeforeEach(func() {
			paramPlan.SkipRouteCreation = false
			paramPlan.NoRouteFlag = true
		})

		It("never attempts to create a route", func() {
			Expect(events).To(BeEmpty())
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(fakeV2Actor.GetApplicationRoutesCallCount()).To(BeZero())
			Expect(fakeV2Actor.CreateRouteCallCount()).To(BeZero())
		})
	})
})
