package pushaction_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/pushaction"
	"code.cloudfoundry.org/cli/actor/pushaction/pushactionfakes"
	"code.cloudfoundry.org/cli/actor/v2action"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Routes", func() {
	var (
		actor       *Actor
		fakeV2Actor *pushactionfakes.FakeV2Actor
	)

	BeforeEach(func() {
		fakeV2Actor = new(pushactionfakes.FakeV2Actor)
		actor = NewActor(fakeV2Actor)
	})

	Describe("FindOrReturnEmptyRoute", func() {
		var (
			route v2action.Route

			returnedRoute v2action.Route
			warnings      Warnings
			executeErr    error
		)

		BeforeEach(func() {
			route = v2action.Route{
				Domain: v2action.Domain{
					Name: "some-domain.com",
					GUID: "some-domain-guid",
				},
				Host:      "some-host",
				SpaceGUID: "some-space-guid",
			}
		})

		JustBeforeEach(func() {
			returnedRoute, warnings, executeErr = actor.FindOrReturnPartialRoute(route)
		})

		Context("when the route exists", func() {
			var existingRoute v2action.Route

			BeforeEach(func() {
				fakeV2Actor.CheckRouteReturns(true, v2action.Warnings{"check-route-warnings"}, nil)

				existingRoute = route
				existingRoute.GUID = "route-guid"
			})

			Context("when the route exists in this space", func() {
				BeforeEach(func() {
					fakeV2Actor.GetRouteByHostAndDomainReturns(existingRoute, v2action.Warnings{"get-route-warnings"}, nil)
				})

				It("returns the existing route", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("check-route-warnings", "get-route-warnings"))
					Expect(returnedRoute).To(Equal(existingRoute))

					Expect(fakeV2Actor.CheckRouteCallCount()).To(Equal(1))
					Expect(fakeV2Actor.CheckRouteArgsForCall(0)).To(Equal(route))

					Expect(fakeV2Actor.GetRouteByHostAndDomainCallCount()).To(Equal(1))
					host, domainGUID := fakeV2Actor.GetRouteByHostAndDomainArgsForCall(0)
					Expect(host).To(Equal(route.Host))
					Expect(domainGUID).To(Equal(route.Domain.GUID))
				})
			})

			Context("when the route exists in a different space", func() {
				BeforeEach(func() {
					fakeV2Actor.GetRouteByHostAndDomainReturns(v2action.Route{}, v2action.Warnings{"get-route-warnings"}, v2action.RouteNotFoundError{})
				})

				It("returns a RouteInDifferentSpaceError and warnings", func() {
					Expect(executeErr).To(MatchError(RouteInDifferentSpaceError{Route: "some-host.some-domain.com"}))
					Expect(warnings).To(ConsistOf("check-route-warnings", "get-route-warnings"))
				})
			})

			Context("when the route lookup returns an error", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("nooooo")
					fakeV2Actor.GetRouteByHostAndDomainReturns(v2action.Route{}, v2action.Warnings{"get-route-warnings"}, expectedErr)
				})

				It("the error and warnings", func() {
					Expect(executeErr).To(MatchError(expectedErr))
					Expect(warnings).To(ConsistOf("check-route-warnings", "get-route-warnings"))
				})
			})
		})

		Context("when the route does not exist", func() {
			BeforeEach(func() {
				fakeV2Actor.CheckRouteReturns(false, v2action.Warnings{"check-route-warnings"}, nil)
			})

			It("returns route with an empty GUID", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("check-route-warnings"))
				Expect(returnedRoute).To(Equal(route))
			})
		})

		Context("when the route check errors", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("nooooo")
				fakeV2Actor.CheckRouteReturns(true, v2action.Warnings{"check-route-warnings"}, expectedErr)
			})

			It("the error and warnings", func() {
				Expect(executeErr).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("check-route-warnings"))
			})
		})
	})

	Describe("GetRouteWithDefaultDomain", func() {
		var (
			host    string
			orgGUID string

			defaultRoute v2action.Route
			warnings     Warnings
			executeErr   error

			domain v2action.Domain
		)

		BeforeEach(func() {
			host = "some-app"
			orgGUID = "some-org-guid"

			domain = v2action.Domain{
				Name: "private-domain.com",
				GUID: "some-private-domain-guid",
			}
		})

		JustBeforeEach(func() {
			defaultRoute, warnings, executeErr = actor.GetRouteWithDefaultDomain(host, orgGUID)
		})

		Context("when retrieving the domains is successful", func() {
			BeforeEach(func() {
				fakeV2Actor.GetOrganizationDomainsReturns(
					[]v2action.Domain{domain},
					v2action.Warnings{"private-domain-warnings", "shared-domain-warnings"},
					nil,
				)
			})

			Context("when retrieving the routes is successful", func() {
				BeforeEach(func() {
					// Assumes new route
					fakeV2Actor.CheckRouteReturns(false, v2action.Warnings{"get-route-warnings"}, nil)
				})

				It("returns the route and warnings", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("private-domain-warnings", "shared-domain-warnings", "get-route-warnings"))

					Expect(defaultRoute).To(Equal(v2action.Route{Domain: domain, Host: host}))

					Expect(fakeV2Actor.GetOrganizationDomainsCallCount()).To(Equal(1))
					Expect(fakeV2Actor.GetOrganizationDomainsArgsForCall(0)).To(Equal(orgGUID))

					Expect(fakeV2Actor.CheckRouteCallCount()).To(Equal(1))
					Expect(fakeV2Actor.CheckRouteArgsForCall(0)).To(Equal(v2action.Route{Domain: domain, Host: host}))
				})
			})

			Context("when retrieving the routes errors", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("whoops")
					fakeV2Actor.CheckRouteReturns(false, v2action.Warnings{"get-route-warnings"}, expectedErr)
				})

				It("returns errors and warnings", func() {
					Expect(executeErr).To(MatchError(expectedErr))
					Expect(warnings).To(ConsistOf("private-domain-warnings", "shared-domain-warnings", "get-route-warnings"))
				})
			})
		})

		Context("when retrieving the domains errors", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("whoops")
				fakeV2Actor.GetOrganizationDomainsReturns([]v2action.Domain{}, v2action.Warnings{"private-domain-warnings", "shared-domain-warnings"}, expectedErr)
			})

			It("returns errors and warnings", func() {
				Expect(executeErr).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("private-domain-warnings", "shared-domain-warnings"))
			})
		})
	})
})
