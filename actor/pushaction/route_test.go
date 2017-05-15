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

	Describe("BindRoutes", func() {
		var (
			config ApplicationConfig

			returnedConfig ApplicationConfig
			boundRoutes    bool
			warnings       Warnings
			executeErr     error
		)

		BeforeEach(func() {
			config = ApplicationConfig{
				DesiredApplication: v2action.Application{
					GUID: "some-app-guid",
				},
			}
		})

		JustBeforeEach(func() {
			returnedConfig, boundRoutes, warnings, executeErr = actor.BindRoutes(config)
		})

		Context("when routes need to be bound to the application", func() {
			BeforeEach(func() {
				config.CurrentRoutes = []v2action.Route{
					{GUID: "some-route-guid-2", Host: "some-route-2"},
				}
				config.DesiredRoutes = []v2action.Route{
					{GUID: "some-route-guid-1", Host: "some-route-1", Domain: v2action.Domain{Name: "some-domain.com"}},
					{GUID: "some-route-guid-2", Host: "some-route-2"},
					{GUID: "some-route-guid-3", Host: "some-route-3"},
				}
			})

			Context("when the binding is successful", func() {
				BeforeEach(func() {
					fakeV2Actor.BindRouteToApplicationReturns(v2action.Warnings{"bind-route-warning"}, nil)
				})

				It("only creates the routes that do not exist", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("bind-route-warning", "bind-route-warning"))
					Expect(boundRoutes).To(BeTrue())

					Expect(returnedConfig.CurrentRoutes).To(Equal(config.DesiredRoutes))

					Expect(fakeV2Actor.BindRouteToApplicationCallCount()).To(Equal(2))

					routeGUID, appGUID := fakeV2Actor.BindRouteToApplicationArgsForCall(0)
					Expect(routeGUID).To(Equal("some-route-guid-1"))
					Expect(appGUID).To(Equal("some-app-guid"))

					routeGUID, appGUID = fakeV2Actor.BindRouteToApplicationArgsForCall(1)
					Expect(routeGUID).To(Equal("some-route-guid-3"))
					Expect(appGUID).To(Equal("some-app-guid"))
				})
			})

			Context("when the binding errors", func() {
				Context("when the route is bound in another space", func() {
					BeforeEach(func() {
						fakeV2Actor.BindRouteToApplicationReturns(v2action.Warnings{"bind-route-warning"}, v2action.RouteInDifferentSpaceError{})
					})

					It("sends the RouteInDifferentSpaceError (with a guid set) and warnings and returns true", func() {
						Expect(executeErr).To(MatchError(v2action.RouteInDifferentSpaceError{Route: "some-route-1.some-domain.com"}))
						Expect(warnings).To(ConsistOf("bind-route-warning"))
					})
				})

				Context("generic error", func() {
					var expectedErr error
					BeforeEach(func() {
						expectedErr = errors.New("oh my")
						fakeV2Actor.BindRouteToApplicationReturns(v2action.Warnings{"bind-route-warning"}, expectedErr)
					})

					It("sends the warnings and errors and returns true", func() {
						Expect(executeErr).To(MatchError(expectedErr))
						Expect(warnings).To(ConsistOf("bind-route-warning"))
					})
				})
			})
		})

		Context("when no routes need to be bound", func() {
			It("returns false", func() {
				Expect(executeErr).ToNot(HaveOccurred())
			})
		})
	})

	Describe("CreateRoutes", func() {
		var (
			config ApplicationConfig

			returnedConfig ApplicationConfig
			createdRoutes  bool
			warnings       Warnings
			executeErr     error
		)

		BeforeEach(func() {
			config = ApplicationConfig{}
		})

		JustBeforeEach(func() {
			returnedConfig, createdRoutes, warnings, executeErr = actor.CreateRoutes(config)
		})

		Describe("when routes need to be created", func() {
			BeforeEach(func() {
				config.DesiredRoutes = []v2action.Route{
					{GUID: "", Host: "some-route-1"},
					{GUID: "some-route-guid-2", Host: "some-route-2"},
					{GUID: "", Host: "some-route-3"},
				}
			})

			Context("when the creation is successful", func() {
				BeforeEach(func() {
					fakeV2Actor.CreateRouteReturnsOnCall(0, v2action.Route{GUID: "some-route-guid-1", Host: "some-route-1"}, v2action.Warnings{"create-route-warning"}, nil)
					fakeV2Actor.CreateRouteReturnsOnCall(1, v2action.Route{GUID: "some-route-guid-3", Host: "some-route-3"}, v2action.Warnings{"create-route-warning"}, nil)
				})

				It("only creates the routes that do not exist", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("create-route-warning", "create-route-warning"))
					Expect(createdRoutes).To(BeTrue())
					Expect(returnedConfig.DesiredRoutes).To(Equal([]v2action.Route{
						{GUID: "some-route-guid-1", Host: "some-route-1"},
						{GUID: "some-route-guid-2", Host: "some-route-2"},
						{GUID: "some-route-guid-3", Host: "some-route-3"},
					}))

					Expect(fakeV2Actor.CreateRouteCallCount()).To(Equal(2))
					Expect(fakeV2Actor.CreateRouteArgsForCall(0)).To(Equal(v2action.Route{Host: "some-route-1"}))
					Expect(fakeV2Actor.CreateRouteArgsForCall(1)).To(Equal(v2action.Route{Host: "some-route-3"}))
				})
			})

			Context("when the creation errors", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("oh my")
					fakeV2Actor.CreateRouteReturns(
						v2action.Route{},
						v2action.Warnings{"create-route-warning"},
						expectedErr)
				})

				It("sends the warnings and errors and returns true", func() {
					Expect(executeErr).To(MatchError(expectedErr))
					Expect(warnings).To(ConsistOf("create-route-warning"))
				})
			})
		})

		Context("when no routes are created", func() {
			BeforeEach(func() {
				config.DesiredRoutes = []v2action.Route{
					{GUID: "some-route-guid-1", Host: "some-route-1"},
					{GUID: "some-route-guid-2", Host: "some-route-2"},
					{GUID: "some-route-guid-3", Host: "some-route-3"},
				}
			})

			It("returns false", func() {
				Expect(createdRoutes).To(BeFalse())
			})
		})
	})

	Describe("GetRouteWithDefaultDomain", func() {
		var (
			host        string
			orgGUID     string
			spaceGUID   string
			knownRoutes []v2action.Route

			defaultRoute v2action.Route
			warnings     Warnings
			executeErr   error

			domain v2action.Domain
		)

		BeforeEach(func() {
			host = "some-app"
			orgGUID = "some-org-guid"
			spaceGUID = "some-space-guid"
			knownRoutes = nil

			domain = v2action.Domain{
				Name: "private-domain.com",
				GUID: "some-private-domain-guid",
			}
		})

		JustBeforeEach(func() {
			defaultRoute, warnings, executeErr = actor.GetRouteWithDefaultDomain(host, orgGUID, spaceGUID, knownRoutes)
		})

		Context("when retrieving the domains is successful", func() {
			BeforeEach(func() {
				fakeV2Actor.GetOrganizationDomainsReturns(
					[]v2action.Domain{domain},
					v2action.Warnings{"private-domain-warnings", "shared-domain-warnings"},
					nil,
				)
			})

			Context("when the route exists", func() {
				BeforeEach(func() {
					// Assumes new route
					fakeV2Actor.FindRouteBoundToSpaceWithSettingsReturns(v2action.Route{
						Domain:    domain,
						GUID:      "some-route-guid",
						Host:      host,
						SpaceGUID: spaceGUID,
					}, v2action.Warnings{"get-route-warnings"}, nil)
				})

				It("returns the route and warnings", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("private-domain-warnings", "shared-domain-warnings", "get-route-warnings"))

					Expect(defaultRoute).To(Equal(v2action.Route{
						Domain:    domain,
						GUID:      "some-route-guid",
						Host:      host,
						SpaceGUID: spaceGUID,
					}))

					Expect(fakeV2Actor.GetOrganizationDomainsCallCount()).To(Equal(1))
					Expect(fakeV2Actor.GetOrganizationDomainsArgsForCall(0)).To(Equal(orgGUID))

					Expect(fakeV2Actor.FindRouteBoundToSpaceWithSettingsCallCount()).To(Equal(1))
					Expect(fakeV2Actor.FindRouteBoundToSpaceWithSettingsArgsForCall(0)).To(Equal(v2action.Route{Domain: domain, Host: host, SpaceGUID: spaceGUID}))
				})

				Context("when the route has been found", func() {
					BeforeEach(func() {
						knownRoutes = []v2action.Route{{
							Domain:    domain,
							GUID:      "some-route-guid",
							Host:      host,
							SpaceGUID: spaceGUID,
						}}
					})

					It("should return the known route and warnings", func() {
						Expect(executeErr).ToNot(HaveOccurred())
						Expect(warnings).To(ConsistOf("private-domain-warnings", "shared-domain-warnings"))

						Expect(defaultRoute).To(Equal(v2action.Route{
							Domain:    domain,
							GUID:      "some-route-guid",
							Host:      host,
							SpaceGUID: spaceGUID,
						}))

						Expect(fakeV2Actor.FindRouteBoundToSpaceWithSettingsCallCount()).To(Equal(0))
					})
				})
			})

			Context("when the route does not exist", func() {
				BeforeEach(func() {
					fakeV2Actor.FindRouteBoundToSpaceWithSettingsReturns(v2action.Route{}, v2action.Warnings{"get-route-warnings"}, v2action.RouteNotFoundError{})
				})

				It("returns a partial route", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("private-domain-warnings", "shared-domain-warnings", "get-route-warnings"))

					Expect(defaultRoute).To(Equal(v2action.Route{Domain: domain, Host: host, SpaceGUID: spaceGUID}))
				})
			})

			Context("when retrieving the routes errors", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("whoops")
					fakeV2Actor.FindRouteBoundToSpaceWithSettingsReturns(v2action.Route{}, v2action.Warnings{"get-route-warnings"}, expectedErr)
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
