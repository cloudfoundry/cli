package v7pushaction_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v7action"
	. "code.cloudfoundry.org/cli/actor/v7pushaction"
	"code.cloudfoundry.org/cli/actor/v7pushaction/v7pushactionfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Routes", func() {
	var (
		actor       *Actor
		fakeV2Actor *v7pushactionfakes.FakeV2Actor
		fakeV7Actor *v7pushactionfakes.FakeV7Actor
	)

	BeforeEach(func() {
		actor, fakeV2Actor, fakeV7Actor, _ = getTestPushActor()
	})

	Describe("CreateAndMapDefaultApplicationRoute", func() {
		var (
			orgGUID     string
			spaceGUID   string
			application v7action.Application

			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			orgGUID = "some-org-guid"
			spaceGUID = "some-space-guid"
			application = v7action.Application{Name: "some-app", GUID: "some-app-guid"}
		})

		JustBeforeEach(func() {
			warnings, executeErr = actor.CreateAndMapDefaultApplicationRoute(orgGUID, spaceGUID, application)
		})

		When("getting default domain errors", func() {
			BeforeEach(func() {
				fakeV7Actor.GetDefaultDomainReturns(
					v7action.Domain{},
					v7action.Warnings{"domain-warning"},
					errors.New("some-error"))
			})

			It("returns the error", func() {
				Expect(executeErr).To(MatchError("some-error"))
				Expect(warnings).To(ConsistOf("domain-warning"))
			})
		})

		When("getting organization domains succeeds", func() {
			BeforeEach(func() {
				fakeV7Actor.GetDefaultDomainReturns(
					v7action.Domain{
						GUID: "some-domain-guid",
						Name: "some-domain",
					},
					v7action.Warnings{"domain-warning"},
					nil,
				)
			})

			When("getting the application routes errors", func() {
				BeforeEach(func() {
					fakeV2Actor.GetApplicationRoutesReturns(
						[]v2action.Route{},
						v2action.Warnings{"route-warning"},
						errors.New("some-error"),
					)
				})

				It("returns the error", func() {
					Expect(executeErr).To(MatchError("some-error"))
					Expect(warnings).To(ConsistOf("domain-warning", "route-warning"))
				})
			})

			When("getting the application routes succeeds", func() {
				When("the route is already bound to the app", func() {
					BeforeEach(func() {
						fakeV2Actor.GetApplicationRoutesReturns(
							[]v2action.Route{
								{
									Host: "some-app",
									Domain: v2action.Domain{
										GUID: "some-domain-guid",
										Name: "some-domain",
									},
									GUID:      "some-route-guid",
									SpaceGUID: "some-space-guid",
								},
							},
							v2action.Warnings{"route-warning"},
							nil,
						)
					})

					It("returns any warnings", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(warnings).To(ConsistOf("domain-warning", "route-warning"))

						Expect(fakeV7Actor.GetDefaultDomainCallCount()).To(Equal(1), "Expected GetOrganizationDomains to be called once, but it was not")
						orgGUID := fakeV7Actor.GetDefaultDomainArgsForCall(0)
						Expect(orgGUID).To(Equal("some-org-guid"))

						Expect(fakeV2Actor.GetApplicationRoutesCallCount()).To(Equal(1), "Expected GetApplicationRoutes to be called once, but it was not")
						appGUID := fakeV2Actor.GetApplicationRoutesArgsForCall(0)
						Expect(appGUID).To(Equal("some-app-guid"))

						Expect(fakeV2Actor.CreateRouteCallCount()).To(Equal(0), "Expected CreateRoute to not be called but it was")
						Expect(fakeV2Actor.MapRouteToApplicationCallCount()).To(Equal(0), "Expected MapRouteToApplication to not be called but it was")
					})
				})

				When("the route isn't bound to the app", func() {
					When("finding route in space errors", func() {
						BeforeEach(func() {
							fakeV2Actor.FindRouteBoundToSpaceWithSettingsReturns(
								v2action.Route{},
								v2action.Warnings{"route-warning"},
								errors.New("some-error"),
							)
						})

						It("returns the error", func() {
							Expect(executeErr).To(MatchError("some-error"))
							Expect(warnings).To(ConsistOf("domain-warning", "route-warning"))
						})
					})

					When("the route exists", func() {
						BeforeEach(func() {
							fakeV2Actor.FindRouteBoundToSpaceWithSettingsReturns(
								v2action.Route{
									GUID: "some-route-guid",
									Host: "some-app",
									Domain: v2action.Domain{
										Name: "some-domain",
										GUID: "some-domain-guid",
									},
									SpaceGUID: "some-space-guid",
								},
								v2action.Warnings{"route-warning"},
								nil,
							)
						})

						When("the map command returns an error", func() {
							BeforeEach(func() {
								fakeV2Actor.MapRouteToApplicationReturns(
									v2action.Warnings{"map-warning"},
									errors.New("some-error"),
								)
							})

							It("returns the error", func() {
								Expect(executeErr).To(MatchError("some-error"))
								Expect(warnings).To(ConsistOf("domain-warning", "route-warning", "map-warning"))
							})
						})

						When("the map command succeeds", func() {
							BeforeEach(func() {
								fakeV2Actor.MapRouteToApplicationReturns(
									v2action.Warnings{"map-warning"},
									nil,
								)
							})

							It("maps the route to the app and returns any warnings", func() {
								Expect(executeErr).ToNot(HaveOccurred())
								Expect(warnings).To(ConsistOf("domain-warning", "route-warning", "map-warning"))

								Expect(fakeV2Actor.FindRouteBoundToSpaceWithSettingsCallCount()).To(Equal(1), "Expected FindRouteBoundToSpaceWithSettings to be called once, but it was not")
								spaceRoute := fakeV2Actor.FindRouteBoundToSpaceWithSettingsArgsForCall(0)
								Expect(spaceRoute).To(Equal(v2action.Route{
									Host: "some-app",
									Domain: v2action.Domain{
										Name: "some-domain",
										GUID: "some-domain-guid",
									},
									SpaceGUID: "some-space-guid",
								}))

								Expect(fakeV2Actor.MapRouteToApplicationCallCount()).To(Equal(1), "Expected MapRouteToApplication to be called once, but it was not")
								routeGUID, appGUID := fakeV2Actor.MapRouteToApplicationArgsForCall(0)
								Expect(routeGUID).To(Equal("some-route-guid"))
								Expect(appGUID).To(Equal("some-app-guid"))
							})
						})
					})

					When("the route does not exist", func() {
						BeforeEach(func() {
							fakeV2Actor.FindRouteBoundToSpaceWithSettingsReturns(
								v2action.Route{},
								v2action.Warnings{"route-warning"},
								actionerror.RouteNotFoundError{},
							)
						})

						When("the create route command errors", func() {
							BeforeEach(func() {
								fakeV2Actor.CreateRouteReturns(
									v2action.Route{},
									v2action.Warnings{"route-create-warning"},
									errors.New("some-error"),
								)
							})

							It("returns the error", func() {
								Expect(executeErr).To(MatchError("some-error"))
								Expect(warnings).To(ConsistOf("domain-warning", "route-warning", "route-create-warning"))
							})
						})

						When("the create route command succeeds", func() {
							BeforeEach(func() {
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
							})

							When("the map command errors", func() {
								BeforeEach(func() {
									fakeV2Actor.MapRouteToApplicationReturns(
										v2action.Warnings{"map-warning"},
										errors.New("some-error"),
									)
								})

								It("returns the error", func() {
									Expect(executeErr).To(MatchError("some-error"))
									Expect(warnings).To(ConsistOf("domain-warning", "route-warning", "route-create-warning", "map-warning"))
								})
							})
							When("the map command succeeds", func() {
								BeforeEach(func() {
									fakeV2Actor.MapRouteToApplicationReturns(
										v2action.Warnings{"map-warning"},
										nil,
									)
								})

								It("creates the route, maps it to the app, and returns any warnings", func() {
									Expect(executeErr).ToNot(HaveOccurred())
									Expect(warnings).To(ConsistOf("domain-warning", "route-warning", "route-create-warning", "map-warning"))

									Expect(fakeV2Actor.CreateRouteCallCount()).To(Equal(1), "Expected CreateRoute to be called once, but it was not")
									defaultRoute, shouldGeneratePort := fakeV2Actor.CreateRouteArgsForCall(0)
									Expect(defaultRoute).To(Equal(v2action.Route{
										Host: "some-app",
										Domain: v2action.Domain{
											Name: "some-domain",
											GUID: "some-domain-guid",
										},
										SpaceGUID: "some-space-guid",
									}))
									Expect(shouldGeneratePort).To(BeFalse())

									Expect(fakeV2Actor.FindRouteBoundToSpaceWithSettingsCallCount()).To(Equal(1), "Expected FindRouteBoundToSpaceWithSettings to be called once, but it was not")
									spaceRoute := fakeV2Actor.FindRouteBoundToSpaceWithSettingsArgsForCall(0)
									Expect(spaceRoute).To(Equal(v2action.Route{
										Host: "some-app",
										Domain: v2action.Domain{
											Name: "some-domain",
											GUID: "some-domain-guid",
										},
										SpaceGUID: "some-space-guid",
									}))

									Expect(fakeV2Actor.MapRouteToApplicationCallCount()).To(Equal(1), "Expected MapRouteToApplication to be called once, but it was not")
									routeGUID, appGUID := fakeV2Actor.MapRouteToApplicationArgsForCall(0)
									Expect(routeGUID).To(Equal("some-route-guid"))
									Expect(appGUID).To(Equal("some-app-guid"))
								})
							})
						})
					})
				})
			})
		})
	})
})
