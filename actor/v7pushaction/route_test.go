package v7pushaction_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"

	"code.cloudfoundry.org/cli/actor/v7action"
	. "code.cloudfoundry.org/cli/actor/v7pushaction"
	"code.cloudfoundry.org/cli/actor/v7pushaction/v7pushactionfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Routes", func() {
	var (
		actor       *Actor
		fakeV7Actor *v7pushactionfakes.FakeV7Actor
	)

	BeforeEach(func() {
		actor, fakeV7Actor, _ = getTestPushActor()
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

			When("getting route fails", func() {
				BeforeEach(func() {
					fakeV7Actor.GetRouteByAttributesReturns(
						v7action.Route{},
						v7action.Warnings{"route-warning"},
						errors.New("route-error"),
					)
				})

				It("returns error and warnings", func() {
					Expect(executeErr).To(MatchError(errors.New("route-error")))
					Expect(warnings).To(ConsistOf("domain-warning", "route-warning"))

					Expect(fakeV7Actor.GetDefaultDomainCallCount()).To(Equal(1))
					orgGUID := fakeV7Actor.GetDefaultDomainArgsForCall(0)
					Expect(orgGUID).To(Equal("some-org-guid"))

					Expect(fakeV7Actor.CreateRouteCallCount()).To(Equal(0))
					Expect(fakeV7Actor.MapRouteCallCount()).To(Equal(0))
				})
			})

			When("the route already exists", func() {
				BeforeEach(func() {
					fakeV7Actor.GetRouteByAttributesReturns(
						v7action.Route{GUID: "route-guid"},
						v7action.Warnings{"route-warning"},
						nil,
					)
				})

				When("mapping the route succeeds", func() {
					BeforeEach(func() {
						fakeV7Actor.MapRouteReturns(
							v7action.Warnings{"map-route-warning"},
							nil,
						)
					})

					It("returns any warnings", func() {
						Expect(executeErr).NotTo(HaveOccurred())
						Expect(warnings).To(ConsistOf("domain-warning", "route-warning", "map-route-warning"))

						Expect(fakeV7Actor.GetDefaultDomainCallCount()).To(Equal(1))
						orgGUID := fakeV7Actor.GetDefaultDomainArgsForCall(0)
						Expect(orgGUID).To(Equal("some-org-guid"))

						Expect(fakeV7Actor.CreateRouteCallCount()).To(Equal(0))
						Expect(fakeV7Actor.MapRouteCallCount()).To(Equal(1))
						routeGUID, appGUID := fakeV7Actor.MapRouteArgsForCall(0)
						Expect(routeGUID).To(Equal("route-guid"))
						Expect(appGUID).To(Equal("some-app-guid"))
					})

				})

				When("mapping the route fails", func() {
					BeforeEach(func() {
						fakeV7Actor.MapRouteReturns(
							v7action.Warnings{"map-route-warning"},
							errors.New("map-route-error"),
						)
					})

					It("returns error and warnings", func() {
						Expect(executeErr).To(MatchError(errors.New("map-route-error")))
						Expect(warnings).To(ConsistOf("domain-warning", "route-warning", "map-route-warning"))

						Expect(fakeV7Actor.GetDefaultDomainCallCount()).To(Equal(1))
						orgGUID := fakeV7Actor.GetDefaultDomainArgsForCall(0)
						Expect(orgGUID).To(Equal("some-org-guid"))

						Expect(fakeV7Actor.CreateRouteCallCount()).To(Equal(0))
						Expect(fakeV7Actor.MapRouteCallCount()).To(Equal(1))
						routeGUID, appGUID := fakeV7Actor.MapRouteArgsForCall(0)
						Expect(routeGUID).To(Equal("route-guid"))
						Expect(appGUID).To(Equal("some-app-guid"))
					})
				})
			})

			When("the route doest *not* exist", func() {
				BeforeEach(func() {
					fakeV7Actor.GetRouteByAttributesReturns(
						v7action.Route{GUID: "route-guid"},
						v7action.Warnings{"route-warning"},
						actionerror.RouteNotFoundError{},
					)
				})

				When("creating the route succeeds", func() {
					BeforeEach(func() {
						fakeV7Actor.CreateRouteReturns(
							v7action.Route{GUID: "route-guid"},
							v7action.Warnings{"create-route-warning"},
							nil,
						)
					})

					When("mapping the route succeeds", func() {
						BeforeEach(func() {
							fakeV7Actor.MapRouteReturns(
								v7action.Warnings{"map-route-warning"},
								nil,
							)
						})

						It("returns any warnings", func() {
							Expect(executeErr).NotTo(HaveOccurred())
							Expect(warnings).To(ConsistOf("domain-warning", "route-warning", "create-route-warning", "map-route-warning"))

							Expect(fakeV7Actor.GetDefaultDomainCallCount()).To(Equal(1))
							orgGUID := fakeV7Actor.GetDefaultDomainArgsForCall(0)
							Expect(orgGUID).To(Equal("some-org-guid"))

							Expect(fakeV7Actor.CreateRouteCallCount()).To(Equal(1))
							spaceGUID, domainName, host, path := fakeV7Actor.CreateRouteArgsForCall(0)
							Expect(spaceGUID).To(Equal("some-space-guid"))
							Expect(domainName).To(Equal("some-domain"))
							Expect(host).To(Equal("some-app"))
							Expect(path).To(Equal(""))
							Expect(fakeV7Actor.MapRouteCallCount()).To(Equal(1))
							routeGUID, appGUID := fakeV7Actor.MapRouteArgsForCall(0)
							Expect(routeGUID).To(Equal("route-guid"))
							Expect(appGUID).To(Equal("some-app-guid"))
						})

					})

					When("mapping the route fails", func() {
						BeforeEach(func() {
							fakeV7Actor.MapRouteReturns(
								v7action.Warnings{"map-route-warning"},
								errors.New("map-route-error"),
							)
						})

						It("returns error and warnings", func() {
							Expect(executeErr).To(MatchError(errors.New("map-route-error")))
							Expect(warnings).To(ConsistOf("domain-warning", "route-warning", "create-route-warning", "map-route-warning"))

							Expect(fakeV7Actor.GetDefaultDomainCallCount()).To(Equal(1))
							orgGUID := fakeV7Actor.GetDefaultDomainArgsForCall(0)
							Expect(orgGUID).To(Equal("some-org-guid"))

							Expect(fakeV7Actor.CreateRouteCallCount()).To(Equal(1))
							spaceGUID, domainName, host, path := fakeV7Actor.CreateRouteArgsForCall(0)
							Expect(spaceGUID).To(Equal("some-space-guid"))
							Expect(domainName).To(Equal("some-domain"))
							Expect(host).To(Equal("some-app"))
							Expect(path).To(Equal(""))
							Expect(fakeV7Actor.MapRouteCallCount()).To(Equal(1))
						})
					})
				})

				When("creating the route fails", func() {
					BeforeEach(func() {
						fakeV7Actor.CreateRouteReturns(
							v7action.Route{},
							v7action.Warnings{"create-route-warning"},
							errors.New("create-route-error"),
						)
					})

					It("returns any warnings", func() {
						Expect(executeErr).To(MatchError(errors.New("create-route-error")))
						Expect(warnings).To(ConsistOf("domain-warning", "route-warning", "create-route-warning"))

						Expect(fakeV7Actor.GetDefaultDomainCallCount()).To(Equal(1))
						orgGUID := fakeV7Actor.GetDefaultDomainArgsForCall(0)
						Expect(orgGUID).To(Equal("some-org-guid"))

						Expect(fakeV7Actor.CreateRouteCallCount()).To(Equal(1))
						spaceGUID, domainName, host, path := fakeV7Actor.CreateRouteArgsForCall(0)
						Expect(spaceGUID).To(Equal("some-space-guid"))
						Expect(domainName).To(Equal("some-domain"))
						Expect(host).To(Equal("some-app"))
						Expect(path).To(Equal(""))
						Expect(fakeV7Actor.MapRouteCallCount()).To(Equal(0))
					})
				})
			})
		})
	})
})
