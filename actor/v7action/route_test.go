package v7action_test

import (
	"errors"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"

	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Route Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		actor, fakeCloudControllerClient, _, _, _, _ = NewTestActor()
	})

	Describe("CreateRoute", func() {
		var (
			warnings   Warnings
			executeErr error
			path       string
		)

		JustBeforeEach(func() {
			_, warnings, executeErr = actor.CreateRoute("space-guid", "domain-name", "hostname", path)
		})

		When("the API layer calls are successful", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetDomainsReturns(
					[]ccv3.Domain{
						{Name: "domain-name", GUID: "domain-guid"},
					},
					ccv3.Warnings{"get-domains-warning"},
					nil,
				)

				fakeCloudControllerClient.CreateRouteReturns(
					ccv3.Route{GUID: "route-guid", SpaceGUID: "space-guid", DomainGUID: "domain-guid", Host: "hostname", Path: "path-name"},
					ccv3.Warnings{"create-warning-1", "create-warning-2"},
					nil)
			})

			When("the input path starts with '/'", func() {
				BeforeEach(func() {
					path = "/path-name"
				})

				It("returns the route with '/<path>' and prints warnings", func() {
					Expect(warnings).To(ConsistOf("create-warning-1", "create-warning-2", "get-domains-warning"))
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(fakeCloudControllerClient.CreateRouteCallCount()).To(Equal(1))
					passedRoute := fakeCloudControllerClient.CreateRouteArgsForCall(0)

					Expect(passedRoute).To(Equal(
						ccv3.Route{
							SpaceGUID:  "space-guid",
							DomainGUID: "domain-guid",
							Host:       "hostname",
							Path:       "/path-name",
						},
					))
				})
			})
		})

		When("the API call to get the domain returns an error", func() {
			When("the cc client returns an RouteNotUniqueError", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetDomainsReturns(
						[]ccv3.Domain{
							{Name: "domain-name", GUID: "domain-guid"},
						},
						ccv3.Warnings{"get-domains-warning"},
						nil,
					)

					fakeCloudControllerClient.GetOrganizationsReturns(
						[]ccv3.Organization{
							{Name: "org-name", GUID: "org-guid"},
						},
						ccv3.Warnings{"get-orgs-warning"},
						nil,
					)

					fakeCloudControllerClient.GetSpacesReturns(
						[]ccv3.Space{
							{Name: "space-name", GUID: "space-guid"},
						},
						ccv3.Warnings{"get-spaces-warning"},
						nil,
					)

					fakeCloudControllerClient.CreateRouteReturns(
						ccv3.Route{},
						ccv3.Warnings{"create-route-warning"},
						ccerror.RouteNotUniqueError{
							UnprocessableEntityError: ccerror.UnprocessableEntityError{Message: "some cool error"},
						},
					)
				})

				It("returns the RouteAlreadyExistsError and warnings", func() {
					Expect(executeErr).To(MatchError(actionerror.RouteAlreadyExistsError{
						Err: ccerror.RouteNotUniqueError{
							UnprocessableEntityError: ccerror.UnprocessableEntityError{Message: "some cool error"},
						},
					}))
					Expect(warnings).To(ConsistOf("get-domains-warning", "create-route-warning"))
				})
			})

			When("the cc client returns a different error", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetDomainsReturns(
						[]ccv3.Domain{},
						ccv3.Warnings{"domain-warning-1", "domain-warning-2"},
						errors.New("api-domains-error"),
					)
				})

				It("it returns an error and prints warnings", func() {
					Expect(warnings).To(ConsistOf("domain-warning-1", "domain-warning-2"))
					Expect(executeErr).To(MatchError("api-domains-error"))

					Expect(fakeCloudControllerClient.GetDomainsCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.CreateRouteCallCount()).To(Equal(0))
				})
			})
		})
	})

	Describe("GetRoutesBySpace", func() {
		var (
			routes     []Route
			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			fakeCloudControllerClient.GetDomainsReturns(
				[]ccv3.Domain{
					{Name: "domain1-name", GUID: "domain1-guid"},
					{Name: "domain2-name", GUID: "domain2-guid"},
				},
				ccv3.Warnings{"get-domains-warning"},
				nil,
			)

			fakeCloudControllerClient.GetSpacesReturns(
				[]ccv3.Space{
					{Name: "space-name", GUID: "space-guid"},
				},
				ccv3.Warnings{"get-spaces-warning"},
				nil,
			)

			fakeCloudControllerClient.GetRoutesReturns(
				[]ccv3.Route{
					{GUID: "route1-guid", SpaceGUID: "space-guid", DomainGUID: "domain1-guid", Host: "hostname", URL: "hostname.domain1-name"},
					{GUID: "route2-guid", SpaceGUID: "space-guid", DomainGUID: "domain2-guid", Path: "/my-path", URL: "domain2-name/my-path"},
					{GUID: "route3-guid", SpaceGUID: "space-guid", DomainGUID: "domain1-guid", URL: "domain1-name"},
				},
				ccv3.Warnings{"get-route-warning-1", "get-route-warning-2"},
				nil,
			)
		})

		JustBeforeEach(func() {
			routes, warnings, executeErr = actor.GetRoutesBySpace("space-guid")
		})

		When("the API layer calls are successful", func() {
			It("returns the routes and warnings", func() {
				Expect(routes).To(Equal([]Route{
					{GUID: "route1-guid", SpaceGUID: "space-guid", DomainGUID: "domain1-guid", Host: "hostname", DomainName: "domain1-name", SpaceName: "space-name", URL: "hostname.domain1-name"},
					{GUID: "route2-guid", SpaceGUID: "space-guid", DomainGUID: "domain2-guid", Path: "/my-path", DomainName: "domain2-name", SpaceName: "space-name", URL: "domain2-name/my-path"},
					{GUID: "route3-guid", SpaceGUID: "space-guid", DomainGUID: "domain1-guid", DomainName: "domain1-name", SpaceName: "space-name", URL: "domain1-name"},
				}))
				Expect(warnings).To(ConsistOf("get-route-warning-1", "get-route-warning-2", "get-spaces-warning"))
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(fakeCloudControllerClient.GetSpacesCallCount()).To(Equal(1))
				query := fakeCloudControllerClient.GetSpacesArgsForCall(0)
				Expect(query).To(HaveLen(1))
				Expect(query[0].Key).To(Equal(ccv3.GUIDFilter))
				Expect(query[0].Values).To(ConsistOf("space-guid"))

				Expect(fakeCloudControllerClient.GetRoutesCallCount()).To(Equal(1))
				query = fakeCloudControllerClient.GetRoutesArgsForCall(0)
				Expect(query).To(HaveLen(1))
				Expect(query[0].Key).To(Equal(ccv3.SpaceGUIDFilter))
				Expect(query[0].Values).To(ConsistOf("space-guid"))
			})
		})

		When("getting routes fails", func() {
			var err = errors.New("failed to get route")

			BeforeEach(func() {
				fakeCloudControllerClient.GetRoutesReturns(
					nil,
					ccv3.Warnings{"get-route-warning-1", "get-route-warning-2"},
					err)
			})

			It("returns the error and any warnings", func() {
				Expect(executeErr).To(Equal(err))
				Expect(warnings).To(ConsistOf("get-route-warning-1", "get-route-warning-2"))
			})
		})

		When("getting spaces fails", func() {
			var err = errors.New("failed to get spaces")

			BeforeEach(func() {
				fakeCloudControllerClient.GetSpacesReturns(
					nil,
					ccv3.Warnings{"get-spaces-warning"},
					err,
				)
			})

			It("returns the error and any warnings", func() {
				Expect(executeErr).To(Equal(err))
				Expect(warnings).To(ConsistOf("get-route-warning-1", "get-route-warning-2", "get-spaces-warning"))
			})
		})

	})

	Describe("GetRouteByNameAndSpace", func() {
		BeforeEach(func() {
			fakeCloudControllerClient.GetDomainsReturns(
				[]ccv3.Domain{
					{Name: "domain-name", GUID: "domain-guid"},
				},
				ccv3.Warnings{"get-domains-warning"},
				nil,
			)

			fakeCloudControllerClient.GetSpacesReturns(
				[]ccv3.Space{
					{Name: "space-name", GUID: "space-guid"},
				},
				ccv3.Warnings{"get-spaces-warning"},
				nil,
			)

			fakeCloudControllerClient.GetRoutesReturns(
				[]ccv3.Route{
					{GUID: "route1-guid", SpaceGUID: "space-guid", DomainGUID: "domain-guid", Host: "hostname", URL: "hostname.domain-name", Path: "/the-path"},
				},
				ccv3.Warnings{"get-route-warning-1", "get-route-warning-2"},
				nil,
			)
		})

		When("the route does not have a host", func() {
			It("returns the route and warnings", func() {
				routeGUID, warnings, executeErr := actor.GetRouteGUID("hostname.domain-name", "space-guid")
				Expect(routeGUID).To(Equal("route1-guid"))
				Expect(warnings).To(ConsistOf("get-domains-warning", "get-route-warning-1", "get-route-warning-2"))
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(fakeCloudControllerClient.GetDomainsCallCount()).To(Equal(1))
				query := fakeCloudControllerClient.GetDomainsArgsForCall(0)
				Expect(query).To(HaveLen(1))
				Expect(query[0].Key).To(Equal(ccv3.NameFilter))
				Expect(query[0].Values).To(ConsistOf("hostname.domain-name"))

				Expect(fakeCloudControllerClient.GetRoutesCallCount()).To(Equal(1))
				query = fakeCloudControllerClient.GetRoutesArgsForCall(0)
				Expect(query).To(HaveLen(4))
				Expect(query[0].Key).To(Equal(ccv3.SpaceGUIDFilter))
				Expect(query[0].Values).To(ConsistOf("space-guid"))
				Expect(query[1].Key).To(Equal(ccv3.DomainGUIDFilter))
				Expect(query[1].Values).To(ConsistOf("domain-guid"))
				Expect(query[2].Key).To(Equal(ccv3.HostsFilter))
				Expect(query[2].Values).To(ConsistOf(""))
				Expect(query[3].Key).To(Equal(ccv3.PathsFilter))
				Expect(query[3].Values).To(ConsistOf(""))
			})
		})

		When("the route has a host defined", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetDomainsReturnsOnCall(
					0,
					[]ccv3.Domain{},
					ccv3.Warnings{"get-domains-warning-1"},
					nil,
				)

				fakeCloudControllerClient.GetDomainsReturnsOnCall(
					1,
					[]ccv3.Domain{
						{Name: "domain-name", GUID: "domain-guid"},
					},
					ccv3.Warnings{"get-domains-warning-2"},
					nil,
				)

				fakeCloudControllerClient.GetRoutesReturns(
					[]ccv3.Route{
						{GUID: "route1-guid", SpaceGUID: "space-guid", DomainGUID: "domain-guid", Host: "hostname", URL: "hostname.domain-name", Path: "/the-path"},
					},
					ccv3.Warnings{"get-route-warning-1", "get-route-warning-2"},
					nil,
				)
			})
			It("returns the route and warnings", func() {
				routeGUID, warnings, executeErr := actor.GetRouteGUID("hostname.domain-name/the-path", "space-guid")
				Expect(routeGUID).To(Equal("route1-guid"))
				Expect(warnings).To(ConsistOf("get-domains-warning-1", "get-domains-warning-2", "get-route-warning-1", "get-route-warning-2"))
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(fakeCloudControllerClient.GetDomainsCallCount()).To(Equal(2))
				query := fakeCloudControllerClient.GetDomainsArgsForCall(0)
				Expect(query).To(HaveLen(1))
				Expect(query[0].Key).To(Equal(ccv3.NameFilter))
				Expect(query[0].Values).To(ConsistOf("hostname.domain-name"))
				query = fakeCloudControllerClient.GetDomainsArgsForCall(1)
				Expect(query).To(HaveLen(1))
				Expect(query[0].Key).To(Equal(ccv3.NameFilter))
				Expect(query[0].Values).To(ConsistOf("domain-name"))

				Expect(fakeCloudControllerClient.GetRoutesCallCount()).To(Equal(1))
				query = fakeCloudControllerClient.GetRoutesArgsForCall(0)
				Expect(query).To(HaveLen(4))
				Expect(query[0].Key).To(Equal(ccv3.SpaceGUIDFilter))
				Expect(query[0].Values).To(ConsistOf("space-guid"))
				Expect(query[1].Key).To(Equal(ccv3.DomainGUIDFilter))
				Expect(query[1].Values).To(ConsistOf("domain-guid"))
				Expect(query[2].Key).To(Equal(ccv3.HostsFilter))
				Expect(query[2].Values).To(ConsistOf("hostname"))
				Expect(query[3].Key).To(Equal(ccv3.PathsFilter))
				Expect(query[3].Values).To(ConsistOf("/the-path"))
			})
		})

		When("invalid domain cannot be found", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetDomainsReturns(
					[]ccv3.Domain{},
					ccv3.Warnings{"get-domains-warning"},
					nil,
				)
			})

			It("returns the error and any warnings", func() {
				_, warnings, executeErr := actor.GetRouteGUID("unsplittabledomain/the-path", "space-guid")
				Expect(warnings).To(ConsistOf("get-domains-warning"))
				Expect(executeErr).To(MatchError(actionerror.DomainNotFoundError{Name: "unsplittabledomain"}))
			})
		})

		When("the route does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetRoutesReturns(
					[]ccv3.Route{},
					ccv3.Warnings{"get-route-warning-1", "get-route-warning-2"},
					nil,
				)
			})

			It("returns the error and any warnings", func() {
				_, warnings, executeErr := actor.GetRouteGUID("unsplittabledomain/the-path", "space-guid")
				Expect(warnings).To(ConsistOf("get-domains-warning", "get-route-warning-1", "get-route-warning-2"))
				Expect(executeErr).To(MatchError(actionerror.RouteNotFoundError{Host: "", DomainName: "unsplittabledomain", Path: "/the-path"}))
			})
		})

		When("getting domain fails", func() {
			var err = errors.New("failed to get domain")

			BeforeEach(func() {
				fakeCloudControllerClient.GetDomainsReturns(
					nil,
					ccv3.Warnings{"get-domains-warning"},
					err,
				)
			})

			It("returns the error and any warnings", func() {
				_, warnings, executeErr := actor.GetRouteGUID("hostname.domain-name/the-path", "space-guid")
				Expect(warnings).To(ConsistOf("get-domains-warning"))
				Expect(executeErr).To(Equal(err))
			})
		})

		When("getting route fails", func() {
			var err = errors.New("failed to get route")

			BeforeEach(func() {
				fakeCloudControllerClient.GetRoutesReturns(
					nil,
					ccv3.Warnings{"get-route-warning-1", "get-route-warning-2"},
					err)
			})

			It("returns the error and any warnings", func() {
				_, warnings, executeErr := actor.GetRouteGUID("hostname.domain-name/the-path", "space-guid")
				Expect(executeErr).To(Equal(err))
				Expect(warnings).To(ConsistOf("get-domains-warning", "get-route-warning-1", "get-route-warning-2"))
			})
		})
	})

	Describe("GetRoutesByOrg", func() {
		var (
			routes     []Route
			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			fakeCloudControllerClient.GetDomainsReturns(
				[]ccv3.Domain{
					{Name: "domain1-name", GUID: "domain1-guid"},
					{Name: "domain2-name", GUID: "domain2-guid"},
				},
				ccv3.Warnings{"get-domains-warning"},
				nil,
			)

			fakeCloudControllerClient.GetSpacesReturns(
				[]ccv3.Space{
					{Name: "space1-name", GUID: "space1-guid"},
					{Name: "space2-name", GUID: "space2-guid"},
				},
				ccv3.Warnings{"get-spaces-warning"},
				nil,
			)

			fakeCloudControllerClient.GetRoutesReturns(
				[]ccv3.Route{
					{GUID: "route1-guid", SpaceGUID: "space1-guid", URL: "hostname.domain1-name", DomainGUID: "domain1-guid", Host: "hostname"},
					{GUID: "route2-guid", SpaceGUID: "space2-guid", URL: "domain2-name/my-path", DomainGUID: "domain2-guid", Path: "/my-path"},
					{GUID: "route3-guid", SpaceGUID: "space1-guid", URL: "domain1-name", DomainGUID: "domain1-guid"},
				},
				ccv3.Warnings{"get-route-warning-1", "get-route-warning-2"},
				nil,
			)
		})

		JustBeforeEach(func() {
			routes, warnings, executeErr = actor.GetRoutesByOrg("org-guid")
		})

		When("the API layer calls are successful", func() {
			It("returns the routes and warnings", func() {
				Expect(routes).To(Equal([]Route{
					{
						GUID:       "route1-guid",
						SpaceGUID:  "space1-guid",
						DomainGUID: "domain1-guid",
						Host:       "hostname",
						DomainName: "domain1-name",
						SpaceName:  "space1-name",
						URL:        "hostname.domain1-name",
					},
					{
						GUID:       "route2-guid",
						SpaceGUID:  "space2-guid",
						DomainGUID: "domain2-guid",
						Path:       "/my-path",
						DomainName: "domain2-name",
						SpaceName:  "space2-name",
						URL:        "domain2-name/my-path",
					},
					{
						GUID:       "route3-guid",
						SpaceGUID:  "space1-guid",
						DomainGUID: "domain1-guid",
						DomainName: "domain1-name",
						SpaceName:  "space1-name",
						URL:        "domain1-name",
					},
				}))
				Expect(warnings).To(ConsistOf("get-route-warning-1", "get-route-warning-2", "get-spaces-warning"))
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(fakeCloudControllerClient.GetRoutesCallCount()).To(Equal(1))
				query := fakeCloudControllerClient.GetRoutesArgsForCall(0)
				Expect(query).To(HaveLen(1))
				Expect(query[0].Key).To(Equal(ccv3.OrganizationGUIDFilter))
				Expect(query[0].Values).To(ConsistOf("org-guid"))

				Expect(fakeCloudControllerClient.GetSpacesCallCount()).To(Equal(1))
				query = fakeCloudControllerClient.GetSpacesArgsForCall(0)
				Expect(query).To(HaveLen(1))
				Expect(query[0].Key).To(Equal(ccv3.GUIDFilter))
				Expect(query[0].Values).To(ConsistOf("space1-guid", "space2-guid"))
			})
		})

		When("getting routes fails", func() {
			var err = errors.New("failed to get route")

			BeforeEach(func() {
				fakeCloudControllerClient.GetRoutesReturns(
					nil,
					ccv3.Warnings{"get-route-warning-1", "get-route-warning-2"},
					err)
			})

			It("returns the error and any warnings", func() {
				Expect(executeErr).To(Equal(err))
				Expect(warnings).To(ConsistOf("get-route-warning-1", "get-route-warning-2"))
			})
		})

		When("getting spaces fails", func() {
			var err = errors.New("failed to get spaces")

			BeforeEach(func() {
				fakeCloudControllerClient.GetSpacesReturns(
					nil,
					ccv3.Warnings{"get-spaces-warning"},
					err,
				)
			})

			It("returns the error and any warnings", func() {
				Expect(executeErr).To(Equal(err))
				Expect(warnings).To(ConsistOf("get-route-warning-1", "get-route-warning-2", "get-spaces-warning"))
			})
		})
	})

	Describe("GetRouteSummaries", func() {
		var (
			routes         []Route
			routeSummaries []RouteSummary
			warnings       Warnings
			executeErr     error
		)

		BeforeEach(func() {
			routes = []Route{
				{GUID: "route-guid-1"},
				{GUID: "route-guid-2"},
				{GUID: "route-guid-3"},
			}

			fakeCloudControllerClient.GetRouteDestinationsReturnsOnCall(0,
				[]ccv3.RouteDestination{
					{
						App: ccv3.RouteDestinationApp{GUID: "app-guid-1"},
					},
				},
				ccv3.Warnings{"get-destinations-warning-1"},
				nil,
			)

			fakeCloudControllerClient.GetRouteDestinationsReturnsOnCall(1,
				[]ccv3.RouteDestination{
					{
						App: ccv3.RouteDestinationApp{GUID: "app-guid-1"},
					},
					{
						App: ccv3.RouteDestinationApp{GUID: "app-guid-2"},
					},
				},
				ccv3.Warnings{"get-destinations-warning-2"},
				nil,
			)

			fakeCloudControllerClient.GetRouteDestinationsReturnsOnCall(2,
				nil,
				ccv3.Warnings{"get-destinations-warning-3"},
				nil,
			)

			fakeCloudControllerClient.GetApplicationsReturns(
				[]ccv3.Application{
					{
						GUID: "app-guid-1",
						Name: "app-name-1",
					},
					{
						GUID: "app-guid-2",
						Name: "app-name-2",
					},
				},
				ccv3.Warnings{"get-apps-warning"},
				nil,
			)
		})

		JustBeforeEach(func() {
			routeSummaries, warnings, executeErr = actor.GetRouteSummaries(routes)
		})

		When("the API layer calls are successful", func() {
			It("returns the routes and warnings", func() {
				Expect(routeSummaries).To(Equal([]RouteSummary{
					{
						Route:    Route{GUID: "route-guid-1"},
						AppNames: []string{"app-name-1"},
					},
					{
						Route:    Route{GUID: "route-guid-2"},
						AppNames: []string{"app-name-1", "app-name-2"},
					},
					{
						Route:    Route{GUID: "route-guid-3"},
						AppNames: nil,
					},
				}))
				Expect(warnings).To(ConsistOf("get-destinations-warning-1", "get-destinations-warning-2", "get-destinations-warning-3", "get-apps-warning"))
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(fakeCloudControllerClient.GetRouteDestinationsCallCount()).To(Equal(3))
				Expect(fakeCloudControllerClient.GetRouteDestinationsArgsForCall(0)).To(Equal("route-guid-1"))
				Expect(fakeCloudControllerClient.GetRouteDestinationsArgsForCall(1)).To(Equal("route-guid-2"))
				Expect(fakeCloudControllerClient.GetRouteDestinationsArgsForCall(2)).To(Equal("route-guid-3"))

				Expect(fakeCloudControllerClient.GetApplicationsCallCount()).To(Equal(1))
				query := fakeCloudControllerClient.GetApplicationsArgsForCall(0)
				Expect(query).To(ConsistOf(
					ccv3.Query{Key: ccv3.GUIDFilter, Values: []string{"app-guid-1", "app-guid-2"}},
				))
			})
		})

		When("getting route destinations fails for one route", func() {
			var err = errors.New("failed to get route destinations")

			BeforeEach(func() {
				fakeCloudControllerClient.GetRouteDestinationsReturnsOnCall(1,
					nil,
					ccv3.Warnings{"get-destinations-warning-2"},
					err,
				)
			})

			It("returns the error and any warnings", func() {
				Expect(executeErr).To(Equal(err))
				Expect(warnings).To(ConsistOf("get-destinations-warning-1", "get-destinations-warning-2"))
			})
		})

		When("getting apps fails", func() {
			var err = errors.New("failed to get apps")

			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationsReturns(
					nil,
					ccv3.Warnings{"get-apps-warning"},
					err,
				)
			})

			It("returns the error and any warnings", func() {
				Expect(executeErr).To(Equal(err))
				Expect(warnings).To(ConsistOf("get-destinations-warning-1", "get-destinations-warning-2", "get-destinations-warning-3", "get-apps-warning"))
			})
		})
	})

	Describe("GetRouteDestinations", func() {
		var (
			routeGUID    string
			destinations []RouteDestination

			executeErr error
			warnings   Warnings
		)

		JustBeforeEach(func() {
			destinations, warnings, executeErr = actor.GetRouteDestinations(routeGUID)
		})

		BeforeEach(func() {
			routeGUID = "route-guid"
		})

		When("the cloud controller client errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetRouteDestinationsReturns(
					nil,
					ccv3.Warnings{"get-destinations-warning"},
					errors.New("get-destinations-error"),
				)
			})

			It("returns the error and warnings", func() {
				Expect(executeErr).To(MatchError(errors.New("get-destinations-error")))
				Expect(warnings).To(ConsistOf("get-destinations-warning"))
			})
		})

		When("the cloud controller client succeeds", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetRouteDestinationsReturns(
					[]ccv3.RouteDestination{
						{GUID: "destination-guid-1", App: ccv3.RouteDestinationApp{GUID: "app-guid-1"}},
						{GUID: "destination-guid-2", App: ccv3.RouteDestinationApp{GUID: "app-guid-2"}},
					},
					ccv3.Warnings{"get-destinations-warning"},
					nil,
				)
			})

			It("returns the destinations and warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("get-destinations-warning"))
				Expect(destinations).To(ConsistOf(
					RouteDestination{GUID: "destination-guid-1", App: RouteDestinationApp{GUID: "app-guid-1"}},
					RouteDestination{GUID: "destination-guid-2", App: RouteDestinationApp{GUID: "app-guid-2"}},
				))
			})
		})
	})

	Describe("GetRouteDestinationByAppGUID", func() {
		var (
			routeGUID   = "route-guid"
			appGUID     = "app-guid"
			destination RouteDestination

			executeErr error
			warnings   Warnings
		)

		JustBeforeEach(func() {
			destination, warnings, executeErr = actor.GetRouteDestinationByAppGUID(routeGUID, appGUID)
		})

		When("the cloud controller client errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetRouteDestinationsReturns(
					nil,
					ccv3.Warnings{"get-destinations-warning"},
					errors.New("get-destinations-error"),
				)
			})

			It("returns the error and warnings", func() {
				Expect(destination).To(Equal(RouteDestination{}))
				Expect(executeErr).To(MatchError(errors.New("get-destinations-error")))
				Expect(warnings).To(ConsistOf("get-destinations-warning"))
			})
		})

		When("the cloud controller client succeeds with a matching app", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetRouteDestinationsReturns(
					[]ccv3.RouteDestination{
						{
							GUID: "destination-guid-1",
							App:  ccv3.RouteDestinationApp{GUID: appGUID, Process: struct{ Type string }{Type: "worker"}},
						},
						{
							GUID: "destination-guid-2",
							App:  ccv3.RouteDestinationApp{GUID: appGUID, Process: struct{ Type string }{Type: constant.ProcessTypeWeb}},
						},
						{
							GUID: "destination-guid-3",
							App:  ccv3.RouteDestinationApp{GUID: "app-guid-2", Process: struct{ Type string }{Type: constant.ProcessTypeWeb}},
						},
					},
					ccv3.Warnings{"get-destinations-warning"},
					nil,
				)
			})

			It("returns the matching destination and warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("get-destinations-warning"))
				Expect(destination).To(Equal(RouteDestination{
					GUID: "destination-guid-2",
					App:  RouteDestinationApp{GUID: appGUID, Process: struct{ Type string }{Type: constant.ProcessTypeWeb}},
				}))
			})
		})

		When("the cloud controller client succeeds without a matching app", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetRouteDestinationsReturns(
					[]ccv3.RouteDestination{
						{
							GUID: "destination-guid-1",
							App:  ccv3.RouteDestinationApp{GUID: appGUID, Process: struct{ Type string }{Type: "worker"}},
						},
						{
							GUID: "destination-guid-2",
							App:  ccv3.RouteDestinationApp{GUID: "app-guid-2", Process: struct{ Type string }{Type: constant.ProcessTypeWeb}},
						},
						{
							GUID: "destination-guid-3",
							App:  ccv3.RouteDestinationApp{GUID: "app-guid-3", Process: struct{ Type string }{Type: constant.ProcessTypeWeb}},
						},
					},
					ccv3.Warnings{"get-destinations-warning"},
					nil,
				)
			})

			It("returns an error and warnings", func() {
				Expect(destination).To(Equal(RouteDestination{}))
				Expect(executeErr).To(MatchError(actionerror.RouteDestinationNotFoundError{
					AppGUID:     appGUID,
					ProcessType: constant.ProcessTypeWeb,
					RouteGUID:   routeGUID,
				}))
				Expect(warnings).To(ConsistOf("get-destinations-warning"))
			})
		})
	})

	Describe("DeleteRoute", func() {
		When("deleting a route", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetDomainsReturns(
					[]ccv3.Domain{
						{GUID: "domain-guid"},
					},
					ccv3.Warnings{"get-domains-warning"},
					nil,
				)
				fakeCloudControllerClient.GetRoutesReturns(
					[]ccv3.Route{
						{GUID: "route-guid"},
					},
					ccv3.Warnings{"get-routes-warning"},
					nil,
				)
				fakeCloudControllerClient.DeleteRouteReturns(
					ccv3.JobURL("https://jobs/job_guid"),
					ccv3.Warnings{"delete-warning"},
					nil,
				)
			})

			It("delegates to the cloud controller client", func() {
				warnings, executeErr := actor.DeleteRoute("domain.com", "hostname", "/path")
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("get-domains-warning", "get-routes-warning", "delete-warning"))

				// Get the domain
				Expect(fakeCloudControllerClient.GetDomainsCallCount()).To(Equal(1))
				query := fakeCloudControllerClient.GetDomainsArgsForCall(0)
				Expect(query).To(ConsistOf([]ccv3.Query{
					{Key: ccv3.NameFilter, Values: []string{"domain.com"}},
				}))

				// Get the route based on the domain GUID
				Expect(fakeCloudControllerClient.GetRoutesCallCount()).To(Equal(1))
				query = fakeCloudControllerClient.GetRoutesArgsForCall(0)
				Expect(query).To(ConsistOf([]ccv3.Query{
					{Key: "domain_guids", Values: []string{"domain-guid"}},
					{Key: "hosts", Values: []string{"hostname"}},
					{Key: "paths", Values: []string{"/path"}},
				}))

				// Delete the route asynchronously
				Expect(fakeCloudControllerClient.DeleteRouteCallCount()).To(Equal(1))
				passedRouteGuid := fakeCloudControllerClient.DeleteRouteArgsForCall(0)
				Expect(passedRouteGuid).To(Equal("route-guid"))

				// Poll the delete job
				Expect(fakeCloudControllerClient.PollJobCallCount()).To(Equal(1))
				responseJobUrl := fakeCloudControllerClient.PollJobArgsForCall(0)
				Expect(responseJobUrl).To(Equal(ccv3.JobURL("https://jobs/job_guid")))
			})

			It("only passes in queries that are not blank", func() {
				_, err := actor.DeleteRoute("domain.com", "", "")
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeCloudControllerClient.GetDomainsCallCount()).To(Equal(1))
				query := fakeCloudControllerClient.GetDomainsArgsForCall(0)
				Expect(query).To(ConsistOf([]ccv3.Query{
					{Key: ccv3.NameFilter, Values: []string{"domain.com"}},
				}))

				Expect(fakeCloudControllerClient.GetRoutesCallCount()).To(Equal(1))
				query = fakeCloudControllerClient.GetRoutesArgsForCall(0)
				Expect(query).To(ConsistOf([]ccv3.Query{
					{Key: "domain_guids", Values: []string{"domain-guid"}},
					{Key: "hosts", Values: []string{""}},
					{Key: "paths", Values: []string{""}},
				}))

				Expect(fakeCloudControllerClient.DeleteRouteCallCount()).To(Equal(1))
				passedRouteGuid := fakeCloudControllerClient.DeleteRouteArgsForCall(0)

				Expect(passedRouteGuid).To(Equal("route-guid"))
			})

			When("getting domains fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetDomainsReturns(
						nil,
						ccv3.Warnings{"get-domains-warning"},
						errors.New("get-domains-error"),
					)
				})

				It("returns the error", func() {
					warnings, err := actor.DeleteRoute("domain.com", "hostname", "path")
					Expect(err).To(MatchError("get-domains-error"))
					Expect(warnings).To(ConsistOf("get-domains-warning"))
				})
			})

			When("getting routes fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetRoutesReturns(
						nil,
						ccv3.Warnings{"get-routes-warning"},
						errors.New("get-routes-error"),
					)
				})

				It("returns the error", func() {
					warnings, err := actor.DeleteRoute("domain.com", "hostname", "path")
					Expect(err).To(MatchError("get-routes-error"))
					Expect(warnings).To(ConsistOf("get-domains-warning", "get-routes-warning"))
				})
			})

			When("deleting route fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.DeleteRouteReturns(
						"",
						ccv3.Warnings{"delete-route-warning"},
						errors.New("delete-route-error"),
					)
				})

				It("returns the error", func() {
					warnings, err := actor.DeleteRoute("domain.com", "hostname", "path")
					Expect(err).To(MatchError("delete-route-error"))
					Expect(warnings).To(ConsistOf("get-domains-warning", "get-routes-warning", "delete-route-warning"))
				})
			})

			When("polling the job fails", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.PollJobReturns(
						ccv3.Warnings{"poll-job-warning"},
						errors.New("async-route-delete-error"),
					)
				})

				It("returns the error", func() {
					warnings, err := actor.DeleteRoute("domain.com", "hostname", "path")
					Expect(err).To(MatchError("async-route-delete-error"))
					Expect(warnings).To(ConsistOf(
						"get-domains-warning",
						"get-routes-warning",
						"delete-warning",
						"poll-job-warning",
					))
				})
			})

			When("no routes are returned", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetRoutesReturns(
						[]ccv3.Route{},
						ccv3.Warnings{"get-routes-warning"},
						nil,
					)
				})

				It("returns the error", func() {
					warnings, err := actor.DeleteRoute("domain.com", "hostname", "/path")
					Expect(err).To(Equal(actionerror.RouteNotFoundError{
						DomainName: "domain.com",
						Host:       "hostname",
						Path:       "/path",
					}))
					Expect(warnings).To(ConsistOf("get-domains-warning", "get-routes-warning"))
				})
			})
		})
	})

	Describe("GetRouteByAttributes", func() {
		var (
			domainName = "some-domain.com"
			domainGUID = "domain-guid"
			hostname   = "hostname"
			path       = "/path"

			executeErr error
			warnings   Warnings
			route      Route
		)

		JustBeforeEach(func() {
			route, warnings, executeErr = actor.GetRouteByAttributes(domainName, domainGUID, hostname, path)
		})

		When("The cc client errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetRoutesReturns(nil, ccv3.Warnings{"get-routes-warning"}, errors.New("scooby"))
			})

			It("returns and empty route, warnings, and the error", func() {
				Expect(fakeCloudControllerClient.GetRoutesCallCount()).To(Equal(1))
				actualQueries := fakeCloudControllerClient.GetRoutesArgsForCall(0)
				Expect(actualQueries).To(ConsistOf(
					ccv3.Query{Key: ccv3.DomainGUIDFilter, Values: []string{domainGUID}},
					ccv3.Query{Key: ccv3.HostsFilter, Values: []string{hostname}},
					ccv3.Query{Key: ccv3.PathsFilter, Values: []string{path}},
				))

				Expect(warnings).To(ConsistOf("get-routes-warning"))
				Expect(executeErr).To(MatchError(errors.New("scooby")))
			})

		})

		When("the cc client succeeds and a route is found", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetRoutesReturns([]ccv3.Route{{
					DomainGUID: domainGUID,
					Host:       hostname,
					Path:       path,
				}}, ccv3.Warnings{"get-routes-warning"}, nil)
			})

			It("returns the route and the warnings", func() {
				Expect(fakeCloudControllerClient.GetRoutesCallCount()).To(Equal(1))
				actualQueries := fakeCloudControllerClient.GetRoutesArgsForCall(0)
				Expect(actualQueries).To(ConsistOf(
					ccv3.Query{Key: ccv3.DomainGUIDFilter, Values: []string{domainGUID}},
					ccv3.Query{Key: ccv3.HostsFilter, Values: []string{hostname}},
					ccv3.Query{Key: ccv3.PathsFilter, Values: []string{path}},
				))

				Expect(warnings).To(ConsistOf("get-routes-warning"))
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(route).To(Equal(Route{
					DomainGUID: domainGUID,
					Host:       hostname,
					Path:       path,
				}))
			})
		})

		When("the cc client succeeds and a route is not found", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetRoutesReturns([]ccv3.Route{}, ccv3.Warnings{"get-routes-warning"}, nil)
			})

			It("returns the route and the warnings", func() {
				Expect(fakeCloudControllerClient.GetRoutesCallCount()).To(Equal(1))
				actualQueries := fakeCloudControllerClient.GetRoutesArgsForCall(0)
				Expect(actualQueries).To(ConsistOf(
					ccv3.Query{Key: ccv3.DomainGUIDFilter, Values: []string{domainGUID}},
					ccv3.Query{Key: ccv3.HostsFilter, Values: []string{hostname}},
					ccv3.Query{Key: ccv3.PathsFilter, Values: []string{path}},
				))

				Expect(warnings).To(ConsistOf("get-routes-warning"))
				Expect(executeErr).To(MatchError(actionerror.RouteNotFoundError{
					DomainName: domainName,
					DomainGUID: domainGUID,
					Host:       hostname,
					Path:       path,
				}))
			})
		})
	})

	Describe("MapRoute", func() {
		var (
			routeGUID string
			appGUID   string

			executeErr error
			warnings   Warnings
		)

		JustBeforeEach(func() {
			warnings, executeErr = actor.MapRoute(routeGUID, appGUID)
		})

		BeforeEach(func() {
			routeGUID = "route-guid"
			appGUID = "app-guid"
		})

		When("the cloud controller client errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.MapRouteReturns(ccv3.Warnings{"map-route-warning"}, errors.New("map-route-error"))
			})

			It("returns the error and warnings", func() {
				Expect(executeErr).To(MatchError(errors.New("map-route-error")))
				Expect(warnings).To(ConsistOf("map-route-warning"))
			})
		})

		When("the cloud controller client succeeds", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.MapRouteReturns(ccv3.Warnings{"map-route-warning"}, nil)
			})

			It("returns the error and warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("map-route-warning"))
			})
		})
	})

	Describe("UnmapRoute", func() {
		var (
			routeGUID       string
			destinationGUID string

			executeErr error
			warnings   Warnings
		)

		JustBeforeEach(func() {
			warnings, executeErr = actor.UnmapRoute(routeGUID, destinationGUID)
		})

		BeforeEach(func() {
			routeGUID = "route-guid"
			destinationGUID = "destination-guid"
		})

		When("the cloud controller client errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.UnmapRouteReturns(ccv3.Warnings{"unmap-route-warning"}, errors.New("unmap-route-error"))
			})

			It("returns the error and warnings", func() {
				Expect(executeErr).To(MatchError(errors.New("unmap-route-error")))
				Expect(warnings).To(ConsistOf("unmap-route-warning"))
			})
		})

		When("the cloud controller client succeeds", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.UnmapRouteReturns(ccv3.Warnings{"unmap-route-warning"}, nil)
			})

			It("returns the error and warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("unmap-route-warning"))
			})
		})
	})

	Describe("DeleteOrphanedRoutes", func() {
		var (
			spaceGUID string

			warnings   Warnings
			executeErr error
		)
		BeforeEach(func() {
			spaceGUID = "space-guid"
		})

		JustBeforeEach(func() {
			warnings, executeErr = actor.DeleteOrphanedRoutes(spaceGUID)
		})

		When("the cloud controller client succeeds", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.DeleteOrphanedRoutesReturns(
					ccv3.JobURL("job"),
					ccv3.Warnings{"delete-orphaned-routes-warning"},
					nil,
				)
			})

			It("deletes orphaned routes", func() {
				Expect(fakeCloudControllerClient.DeleteOrphanedRoutesCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.DeleteOrphanedRoutesArgsForCall(0)).To(Equal(spaceGUID))
			})

			When("polling the job succeeds", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.PollJobReturns(ccv3.Warnings{"poll-job-warning"}, nil)
				})
				It("returns the error and warnings", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("delete-orphaned-routes-warning", "poll-job-warning"))

					Expect(fakeCloudControllerClient.PollJobCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.PollJobArgsForCall(0)).To(Equal(ccv3.JobURL("job")))
				})
			})

			When("polling the job errors", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.PollJobReturns(ccv3.Warnings{"poll-job-warning"}, errors.New("poll-error"))
				})
				It("returns the error and warnings", func() {
					Expect(executeErr).To(MatchError("poll-error"))
					Expect(warnings).To(ConsistOf("delete-orphaned-routes-warning", "poll-job-warning"))

					Expect(fakeCloudControllerClient.PollJobCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.PollJobArgsForCall(0)).To(Equal(ccv3.JobURL("job")))
				})
			})
		})

		When("the cloud controller client error", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.DeleteOrphanedRoutesReturns(
					ccv3.JobURL(""),
					ccv3.Warnings{"delete-orphaned-routes-warning"},
					errors.New("orphaned-error"),
				)
			})

			It("returns the error and warnings", func() {
				Expect(executeErr).To(MatchError("orphaned-error"))
				Expect(warnings).To(ConsistOf("delete-orphaned-routes-warning"))

				Expect(fakeCloudControllerClient.DeleteOrphanedRoutesCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.DeleteOrphanedRoutesArgsForCall(0)).To(Equal(spaceGUID))

				Expect(fakeCloudControllerClient.PollJobCallCount()).To(Equal(0))
			})
		})
	})

	Describe("GetApplicationRoutes", func() {
		var (
			appGUID string

			routes     []Route
			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			appGUID = "some-app-guid"
		})

		JustBeforeEach(func() {
			routes, warnings, executeErr = actor.GetApplicationRoutes(appGUID)
		})

		When("getting routes fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationRoutesReturns(
					[]ccv3.Route{},
					ccv3.Warnings{"get-application-routes-warning"},
					errors.New("application-routes-error"),
				)
			})

			It("returns the warnings and error", func() {
				Expect(executeErr).To(MatchError("application-routes-error"))
				Expect(warnings).To(ConsistOf("get-application-routes-warning"))

				Expect(fakeCloudControllerClient.GetApplicationRoutesCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetApplicationRoutesArgsForCall(0)).To(Equal(appGUID))
			})
		})

		When("getting routes succeeds", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpacesReturns(
					[]ccv3.Space{
						{
							GUID: "routes-space-guid",
							Name: "space-name",
						},
					},
					ccv3.Warnings{"get-spaces-warning"},
					nil,
				)
				fakeCloudControllerClient.GetApplicationRoutesReturns(
					[]ccv3.Route{
						{
							GUID:       "some-route-guid",
							URL:        "some-url.sh",
							SpaceGUID:  "routes-space-guid",
							DomainGUID: "routes-domain-guid",
						},
					},
					ccv3.Warnings{"get-application-routes-warning"},
					nil,
				)
			})

			It("returns the warnings and routes", func() {
				Expect(executeErr).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("get-spaces-warning", "get-application-routes-warning"))

				Expect(fakeCloudControllerClient.GetApplicationRoutesCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetApplicationRoutesArgsForCall(0)).To(Equal(appGUID))

				Expect(routes).To(ConsistOf(
					Route{
						GUID:       "some-route-guid",
						URL:        "some-url.sh",
						SpaceGUID:  "routes-space-guid",
						DomainGUID: "routes-domain-guid",
						SpaceName:  "space-name",
						DomainName: "some-url.sh",
					},
				))
			})

			When("no routes are returned", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationRoutesReturns(
						[]ccv3.Route{},
						ccv3.Warnings{"get-application-routes-warning"},
						nil,
					)
				})

				It("returns an empty list", func() {
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(warnings).To(ConsistOf("get-application-routes-warning"))
					Expect(routes).To(HaveLen(0))
				})
			})

			When("getting spaces fails", func() {
				var err = errors.New("failed to get spaces")

				BeforeEach(func() {
					fakeCloudControllerClient.GetSpacesReturns(
						nil,
						ccv3.Warnings{"get-spaces-warning"},
						err,
					)
				})

				It("returns the error and any warnings", func() {
					Expect(executeErr).To(Equal(err))
					Expect(warnings).To(ConsistOf("get-application-routes-warning", "get-spaces-warning"))

					Expect(fakeCloudControllerClient.GetSpacesCallCount()).To(Equal(1))
					queries := fakeCloudControllerClient.GetSpacesArgsForCall(0)
					Expect(queries).To(ConsistOf(ccv3.Query{Key: ccv3.GUIDFilter, Values: []string{"routes-space-guid"}}))
				})
			})
		})
	})
})
