package v7action_test

import (
	"errors"
	"fmt"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/resources"
	"code.cloudfoundry.org/cli/util/batcher"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Route Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		actor, fakeCloudControllerClient, _, _, _, _, _ = NewTestActor()
	})

	Describe("CreateRoute", func() {
		var (
			warnings   Warnings
			executeErr error
			hostname   string
			path       string
			port       int
		)

		BeforeEach(func() {
			hostname = ""
			path = ""
			port = 0
		})

		JustBeforeEach(func() {
			_, warnings, executeErr = actor.CreateRoute("space-guid", "domain-name", hostname, path, port)
		})

		When("the API layer calls are successful", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetDomainsReturns(
					[]resources.Domain{
						{Name: "domain-name", GUID: "domain-guid"},
					},
					ccv3.Warnings{"get-domains-warning"},
					nil,
				)

				fakeCloudControllerClient.CreateRouteReturns(
					resources.Route{GUID: "route-guid", SpaceGUID: "space-guid", DomainGUID: "domain-guid", Host: "hostname", Path: "path-name"},
					ccv3.Warnings{"create-warning-1", "create-warning-2"},
					nil)
			})

			When("a host and path are given", func() {
				BeforeEach(func() {
					hostname = "hostname"
					path = "/path-name"
				})

				It("returns the route with '/<path>' and prints warnings", func() {
					Expect(warnings).To(ConsistOf("create-warning-1", "create-warning-2", "get-domains-warning"))
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(fakeCloudControllerClient.CreateRouteCallCount()).To(Equal(1))
					passedRoute := fakeCloudControllerClient.CreateRouteArgsForCall(0)

					Expect(passedRoute).To(Equal(
						resources.Route{
							SpaceGUID:  "space-guid",
							DomainGUID: "domain-guid",
							Host:       hostname,
							Path:       path,
						},
					))
				})
			})

			When("a port is given", func() {
				BeforeEach(func() {
					port = 1234
				})

				It("returns the route with port and prints warnings", func() {
					Expect(warnings).To(ConsistOf("create-warning-1", "create-warning-2", "get-domains-warning"))
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(fakeCloudControllerClient.CreateRouteCallCount()).To(Equal(1))
					passedRoute := fakeCloudControllerClient.CreateRouteArgsForCall(0)

					Expect(passedRoute).To(Equal(
						resources.Route{
							SpaceGUID:  "space-guid",
							DomainGUID: "domain-guid",
							Port:       1234,
						},
					))
				})
			})
		})

		When("the API call to get the domain returns an error", func() {
			When("the cc client returns an RouteNotUniqueError", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetDomainsReturns(
						[]resources.Domain{
							{Name: "domain-name", GUID: "domain-guid"},
						},
						ccv3.Warnings{"get-domains-warning"},
						nil,
					)

					fakeCloudControllerClient.GetOrganizationsReturns(
						[]resources.Organization{
							{Name: "org-name", GUID: "org-guid"},
						},
						ccv3.Warnings{"get-orgs-warning"},
						nil,
					)

					fakeCloudControllerClient.GetSpacesReturns(
						[]resources.Space{
							{Name: "space-name", GUID: "space-guid"},
						},
						ccv3.IncludedResources{},
						ccv3.Warnings{"get-spaces-warning"},
						nil,
					)

					fakeCloudControllerClient.CreateRouteReturns(
						resources.Route{},
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
						[]resources.Domain{},
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
			routes     []resources.Route
			warnings   Warnings
			labels     string
			executeErr error
		)

		BeforeEach(func() {
			labels = ""
			fakeCloudControllerClient.GetDomainsReturns(
				[]resources.Domain{
					{Name: "domain1-name", GUID: "domain1-guid"},
					{Name: "domain2-name", GUID: "domain2-guid"},
				},
				ccv3.Warnings{"get-domains-warning"},
				nil,
			)

			fakeCloudControllerClient.GetRoutesReturns(
				[]resources.Route{
					{GUID: "route1-guid", SpaceGUID: "space-guid", DomainGUID: "domain1-guid", Host: "hostname", URL: "hostname.domain1-name", Destinations: []resources.RouteDestination{}},
					{GUID: "route2-guid", SpaceGUID: "space-guid", DomainGUID: "domain2-guid", Path: "/my-path", URL: "domain2-name/my-path", Destinations: []resources.RouteDestination{}},
					{GUID: "route3-guid", SpaceGUID: "space-guid", DomainGUID: "domain1-guid", URL: "domain1-name", Destinations: []resources.RouteDestination{}},
				},
				ccv3.Warnings{"get-route-warning-1", "get-route-warning-2"},
				nil,
			)
		})

		JustBeforeEach(func() {
			routes, warnings, executeErr = actor.GetRoutesBySpace("space-guid", labels)
		})

		When("the API layer calls are successful", func() {
			It("returns the routes and warnings", func() {
				Expect(routes).To(Equal([]resources.Route{
					{GUID: "route1-guid", SpaceGUID: "space-guid", DomainGUID: "domain1-guid", Host: "hostname", URL: "hostname.domain1-name", Destinations: []resources.RouteDestination{}},
					{GUID: "route2-guid", SpaceGUID: "space-guid", DomainGUID: "domain2-guid", Path: "/my-path", URL: "domain2-name/my-path", Destinations: []resources.RouteDestination{}},
					{GUID: "route3-guid", SpaceGUID: "space-guid", DomainGUID: "domain1-guid", URL: "domain1-name", Destinations: []resources.RouteDestination{}},
				}))
				Expect(warnings).To(ConsistOf("get-route-warning-1", "get-route-warning-2"))
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(fakeCloudControllerClient.GetRoutesCallCount()).To(Equal(1))
				query := fakeCloudControllerClient.GetRoutesArgsForCall(0)
				Expect(query).To(HaveLen(1))
				Expect(query[0].Key).To(Equal(ccv3.SpaceGUIDFilter))
				Expect(query[0].Values).To(ConsistOf("space-guid"))
			})

			When("a label selector is provided", func() {
				BeforeEach(func() {
					labels = "ink=blink"
				})

				It("passes a label selector query", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(fakeCloudControllerClient.GetRoutesCallCount()).To(Equal(1))
					expectedQuery := []ccv3.Query{
						{Key: ccv3.SpaceGUIDFilter, Values: []string{"space-guid"}},
						{Key: ccv3.LabelSelectorFilter, Values: []string{"ink=blink"}},
					}
					actualQuery := fakeCloudControllerClient.GetRoutesArgsForCall(0)
					Expect(actualQuery).To(Equal(expectedQuery))
				})
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
	})

	Describe("GetRoute", func() {
		BeforeEach(func() {
			fakeCloudControllerClient.GetDomainsReturns(
				[]resources.Domain{{Name: "domain-name", GUID: "domain-guid"}},
				ccv3.Warnings{"get-domains-warning"},
				nil,
			)
		})

		When("the route has a host and a path", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetDomainsReturnsOnCall(
					0,
					[]resources.Domain{},
					ccv3.Warnings{"get-domains-warning-1"},
					nil,
				)

				fakeCloudControllerClient.GetDomainsReturnsOnCall(
					1,
					[]resources.Domain{
						{Name: "domain-name", GUID: "domain-guid"},
					},
					ccv3.Warnings{"get-domains-warning-2"},
					nil,
				)

				fakeCloudControllerClient.GetRoutesReturns(
					[]resources.Route{
						{
							GUID:       "route1-guid",
							SpaceGUID:  "space-guid",
							DomainGUID: "domain-guid",
							Host:       "hostname",
							URL:        "hostname.domain-name",
							Path:       "/the-path",
						},
					},
					ccv3.Warnings{"get-route-warning-1", "get-route-warning-2"},
					nil,
				)
			})

			It("returns the route and warnings", func() {
				route, warnings, executeErr := actor.GetRoute("hostname.domain-name/the-path", "space-guid")
				Expect(route.GUID).To(Equal("route1-guid"))
				Expect(warnings).To(ConsistOf("get-route-warning-1", "get-route-warning-2", "get-domains-warning-1", "get-domains-warning-2"))
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(fakeCloudControllerClient.GetDomainsCallCount()).To(Equal(2))
				query := fakeCloudControllerClient.GetDomainsArgsForCall(0)
				Expect(query[0].Key).To(Equal(ccv3.NameFilter))
				Expect(query[0].Values).To(ConsistOf("hostname.domain-name"))

				query = fakeCloudControllerClient.GetDomainsArgsForCall(1)
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

		When("the route does not have a host", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetRoutesReturns(
					[]resources.Route{
						{
							GUID:       "route1-guid",
							SpaceGUID:  "space-guid",
							DomainGUID: "domain-guid",
							Host:       "",
							URL:        "domain-name",
							Path:       "",
						},
					},
					ccv3.Warnings{},
					nil,
				)
			})

			It("returns the route and warnings", func() {
				_, _, executeErr := actor.GetRoute("domain-name", "space-guid")
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(fakeCloudControllerClient.GetDomainsCallCount()).To(Equal(1))
				query := fakeCloudControllerClient.GetDomainsArgsForCall(0)
				Expect(query[0].Key).To(Equal(ccv3.NameFilter))
				Expect(query[0].Values).To(ConsistOf("domain-name"))

				Expect(fakeCloudControllerClient.GetRoutesCallCount()).To(Equal(1))
				query = fakeCloudControllerClient.GetRoutesArgsForCall(0)
				Expect(query[2].Key).To(Equal(ccv3.HostsFilter))
				Expect(query[2].Values).To(ConsistOf(""))
			})
		})

		When("the route has a port", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetDomainsReturns(
					[]resources.Domain{{
						Name:      "domain-name",
						GUID:      "domain-guid",
						Protocols: []string{"tcp"},
					}},
					ccv3.Warnings{"get-domains-warning"},
					nil,
				)

				fakeCloudControllerClient.GetRoutesReturns(
					[]resources.Route{
						{
							GUID:       "route1-guid",
							SpaceGUID:  "space-guid",
							DomainGUID: "domain-guid",
							Host:       "",
							URL:        "domain-name",
							Path:       "",
						},
					},
					ccv3.Warnings{},
					nil,
				)
			})

			It("returns the route and warnings", func() {
				_, _, executeErr := actor.GetRoute("domain-name:8080", "space-guid")
				Expect(executeErr).ToNot(HaveOccurred())

				query := fakeCloudControllerClient.GetDomainsArgsForCall(0)
				Expect(query[0].Key).To(Equal(ccv3.NameFilter))
				Expect(query[0].Values).To(ConsistOf("domain-name"))

				Expect(fakeCloudControllerClient.GetRoutesCallCount()).To(Equal(1))
				query = fakeCloudControllerClient.GetRoutesArgsForCall(0)
				Expect(query[4].Key).To(Equal(ccv3.PortsFilter))
				Expect(query[4].Values).To(ConsistOf("8080"))
			})
		})

		When("invalid domain cannot be found", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetDomainsReturns(
					[]resources.Domain{},
					ccv3.Warnings{"get-domains-warning"},
					nil,
				)
			})

			It("returns the error and any warnings", func() {
				_, warnings, executeErr := actor.GetRoute("unsplittabledomain/the-path", "space-guid")
				Expect(warnings).To(ConsistOf("get-domains-warning"))
				Expect(executeErr).To(MatchError(actionerror.DomainNotFoundError{Name: "unsplittabledomain"}))
			})
		})

		When("the route does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetDomainsReturns(
					[]resources.Domain{{
						Name:      "domain-name",
						GUID:      "domain-guid",
						Protocols: []string{"http"},
					}},
					ccv3.Warnings{"get-domains-warning"},
					nil,
				)

				fakeCloudControllerClient.GetRoutesReturns(
					[]resources.Route{},
					ccv3.Warnings{"get-route-warning-1", "get-route-warning-2"},
					nil,
				)
			})

			It("returns the error and any warnings", func() {
				_, warnings, executeErr := actor.GetRoute("unsplittabledomain/the-path", "space-guid")
				Expect(warnings).To(ConsistOf("get-domains-warning", "get-route-warning-1", "get-route-warning-2"))
				Expect(executeErr).To(MatchError(actionerror.RouteNotFoundError{Host: "", DomainName: "domain-name", Path: "/the-path"}))
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
				_, warnings, executeErr := actor.GetRoute("hostname.domain-name/the-path", "space-guid")
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
				_, warnings, executeErr := actor.GetRoute("hostname.domain-name/the-path", "space-guid")
				Expect(executeErr).To(Equal(err))
				Expect(warnings).To(ConsistOf("get-domains-warning", "get-route-warning-1", "get-route-warning-2"))
			})
		})
	})

	Describe("GetRoutesByOrg", func() {
		var (
			routes     []resources.Route
			warnings   Warnings
			executeErr error
			labels     string
		)

		BeforeEach(func() {
			labels = ""

			fakeCloudControllerClient.GetRoutesReturns(
				[]resources.Route{
					{GUID: "route1-guid", SpaceGUID: "space1-guid", URL: "hostname.domain1-name", DomainGUID: "domain1-guid", Host: "hostname"},
					{GUID: "route2-guid", SpaceGUID: "space2-guid", URL: "domain2-name/my-path", DomainGUID: "domain2-guid", Path: "/my-path"},
					{GUID: "route3-guid", SpaceGUID: "space1-guid", URL: "domain1-name", DomainGUID: "domain1-guid"},
				},
				ccv3.Warnings{"get-route-warning-1", "get-route-warning-2"},
				nil,
			)
		})

		JustBeforeEach(func() {
			routes, warnings, executeErr = actor.GetRoutesByOrg("org-guid", labels)
		})

		When("the API layer calls are successful", func() {
			It("returns the routes and warnings", func() {
				Expect(routes).To(Equal([]resources.Route{
					{
						GUID:       "route1-guid",
						SpaceGUID:  "space1-guid",
						DomainGUID: "domain1-guid",
						Host:       "hostname",
						URL:        "hostname.domain1-name",
					},
					{
						GUID:       "route2-guid",
						SpaceGUID:  "space2-guid",
						DomainGUID: "domain2-guid",
						Path:       "/my-path",
						URL:        "domain2-name/my-path",
					},
					{
						GUID:       "route3-guid",
						SpaceGUID:  "space1-guid",
						DomainGUID: "domain1-guid",
						URL:        "domain1-name",
					},
				}))
				Expect(warnings).To(ConsistOf("get-route-warning-1", "get-route-warning-2"))
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(fakeCloudControllerClient.GetRoutesCallCount()).To(Equal(1))
				query := fakeCloudControllerClient.GetRoutesArgsForCall(0)
				Expect(query).To(HaveLen(1))
				Expect(query[0].Key).To(Equal(ccv3.OrganizationGUIDFilter))
				Expect(query[0].Values).To(ConsistOf("org-guid"))
			})

			When("a label selector is provided", func() {
				BeforeEach(func() {
					labels = "env=prod"
				})

				It("converts it into a query key", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(fakeCloudControllerClient.GetRoutesCallCount()).To(Equal(1))
					expectedQuery := []ccv3.Query{
						{Key: ccv3.OrganizationGUIDFilter, Values: []string{"org-guid"}},
						{Key: ccv3.LabelSelectorFilter, Values: []string{"env=prod"}},
					}
					actualQuery := fakeCloudControllerClient.GetRoutesArgsForCall(0)
					Expect(actualQuery).To(Equal(expectedQuery))
				})
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
	})

	Describe("GetRouteSummaries", func() {
		var (
			routes         []resources.Route
			routeSummaries []RouteSummary
			warnings       Warnings
			executeErr     error
		)

		BeforeEach(func() {
			routes = []resources.Route{
				{
					GUID: "route-guid-1",
					Destinations: []resources.RouteDestination{
						{
							App: resources.RouteDestinationApp{
								GUID: "app-guid-1",
							},
						},
					},
				},
				{
					GUID: "route-guid-2",
					Destinations: []resources.RouteDestination{
						{
							App: resources.RouteDestinationApp{
								GUID: "app-guid-1",
							},
						},
						{
							App: resources.RouteDestinationApp{
								GUID: "app-guid-2",
							},
						},
					},
				},
				{
					GUID:         "route-guid-3",
					Destinations: []resources.RouteDestination{},
					DomainGUID:   "domain-guid-1",
					Port:         1028,
					URL:          "domain-1.com:1028",
				},
			}

			fakeCloudControllerClient.GetApplicationsReturns(
				[]resources.Application{
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
						Route: resources.Route{
							GUID: "route-guid-1",
							Destinations: []resources.RouteDestination{
								{
									App: resources.RouteDestinationApp{GUID: "app-guid-1"},
								},
							},
						},
						AppNames: []string{"app-name-1"},
					},
					{
						Route: resources.Route{
							GUID: "route-guid-2",
							Destinations: []resources.RouteDestination{
								{
									App: resources.RouteDestinationApp{GUID: "app-guid-1"},
								},
								{
									App: resources.RouteDestinationApp{GUID: "app-guid-2"},
								},
							},
						},
						AppNames: []string{"app-name-1", "app-name-2"},
					},
					{
						Route: resources.Route{
							GUID:         "route-guid-3",
							Destinations: []resources.RouteDestination{},
							DomainGUID:   "domain-guid-1",
							Port:         1028,
							URL:          "domain-1.com:1028",
						},
						AppNames:   nil,
						DomainName: "domain-1.com",
					},
				}))
				Expect(warnings).To(ConsistOf("get-apps-warning"))
				Expect(executeErr).ToNot(HaveOccurred())

				Expect(fakeCloudControllerClient.GetApplicationsCallCount()).To(Equal(1))
				query := fakeCloudControllerClient.GetApplicationsArgsForCall(0)
				Expect(query).To(ConsistOf(
					ccv3.Query{Key: ccv3.GUIDFilter, Values: []string{"app-guid-1", "app-guid-2"}},
				))
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
				Expect(warnings).To(ConsistOf("get-apps-warning"))
			})
		})

		When("there are many routes", func() {
			const batches = 10

			BeforeEach(func() {
				var (
					manySpaces []resources.Space
					manyApps   []resources.Application
				)

				routes = nil

				for i := 0; i < batcher.BatchSize*batches; i++ {
					routes = append(routes, resources.Route{
						GUID: fmt.Sprintf("route-guid-%d", i),
						Destinations: []resources.RouteDestination{
							{
								App: resources.RouteDestinationApp{
									GUID: fmt.Sprintf("fake-app-guid-%d", i),
								},
							},
						},
						SpaceGUID: fmt.Sprintf("fake-space-guid-%d", i),
						URL:       fmt.Sprintf("fake-url-%d/fake-path-%d:%d", i, i, i),
						Host:      fmt.Sprintf("fake-host-%d", i),
						Path:      fmt.Sprintf("fake-path-%d", i),
						Port:      i,
					})

					manySpaces = append(manySpaces, resources.Space{
						GUID: fmt.Sprintf("fake-space-guid-%d", i),
						Name: fmt.Sprintf("fake-space-name-%d", i),
					})

					manyApps = append(manyApps, resources.Application{
						GUID: fmt.Sprintf("fake-app-guid-%d", i),
						Name: fmt.Sprintf("fake-app-name-%d", i),
					})
				}

				for b := 0; b < batches; b++ {
					fakeCloudControllerClient.GetSpacesReturnsOnCall(
						b,
						manySpaces[:batcher.BatchSize],
						ccv3.IncludedResources{},
						ccv3.Warnings{"get-space-warning"},
						nil,
					)
					manySpaces = manySpaces[batcher.BatchSize:]

					fakeCloudControllerClient.GetApplicationsReturnsOnCall(
						b,
						manyApps[:batcher.BatchSize],
						ccv3.Warnings{"get-apps-warning"},
						nil,
					)
					manyApps = manyApps[batcher.BatchSize:]
				}
			})

			It("makes multiple different calls to get spaces", func() {
				Expect(fakeCloudControllerClient.GetSpacesCallCount()).To(Equal(10))
				Expect(fakeCloudControllerClient.GetSpacesArgsForCall(0)).
					NotTo(Equal(fakeCloudControllerClient.GetSpacesArgsForCall(1)))
			})

			It("makes multiple different calls to get apps", func() {
				Expect(fakeCloudControllerClient.GetApplicationsCallCount()).To(Equal(10))
				Expect(fakeCloudControllerClient.GetApplicationsArgsForCall(0)).
					NotTo(Equal(fakeCloudControllerClient.GetApplicationsArgsForCall(1)))
			})
		})
	})

	Describe("GetRouteDestinations", func() {
		var (
			routeGUID    string
			destinations []resources.RouteDestination

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
					[]resources.RouteDestination{
						{GUID: "destination-guid-1", App: resources.RouteDestinationApp{GUID: "app-guid-1"}},
						{GUID: "destination-guid-2", App: resources.RouteDestinationApp{GUID: "app-guid-2"}},
					},
					ccv3.Warnings{"get-destinations-warning"},
					nil,
				)
			})

			It("returns the destinations and warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("get-destinations-warning"))
				Expect(destinations).To(ConsistOf(
					resources.RouteDestination{GUID: "destination-guid-1", App: resources.RouteDestinationApp{GUID: "app-guid-1"}},
					resources.RouteDestination{GUID: "destination-guid-2", App: resources.RouteDestinationApp{GUID: "app-guid-2"}},
				))
			})
		})
	})

	Describe("GetRouteDestinationByAppGUID", func() {
		var (
			routeGUID   = "route-guid"
			appGUID     = "app-guid"
			route       = resources.Route{GUID: routeGUID}
			destination resources.RouteDestination

			executeErr error
		)

		JustBeforeEach(func() {
			destination, executeErr = actor.GetRouteDestinationByAppGUID(route, appGUID)
		})

		When("the cloud controller client succeeds with a matching app", func() {
			BeforeEach(func() {
				route.Destinations = []resources.RouteDestination{
					{
						GUID: "destination-guid-1",
						App:  resources.RouteDestinationApp{GUID: appGUID, Process: struct{ Type string }{Type: "worker"}},
					},
					{
						GUID: "destination-guid-2",
						App:  resources.RouteDestinationApp{GUID: appGUID, Process: struct{ Type string }{Type: constant.ProcessTypeWeb}},
					},
					{
						GUID: "destination-guid-3",
						App:  resources.RouteDestinationApp{GUID: "app-guid-2", Process: struct{ Type string }{Type: constant.ProcessTypeWeb}},
					},
				}
			})

			It("returns the matching destination and warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(destination).To(Equal(resources.RouteDestination{
					GUID: "destination-guid-2",
					App:  resources.RouteDestinationApp{GUID: appGUID, Process: struct{ Type string }{Type: constant.ProcessTypeWeb}},
				}))
			})
		})

		When("the cloud controller client succeeds without a matching app", func() {
			BeforeEach(func() {
				route.Destinations = []resources.RouteDestination{
					{
						GUID: "destination-guid-1",
						App:  resources.RouteDestinationApp{GUID: appGUID, Process: struct{ Type string }{Type: "worker"}},
					},
					{
						GUID: "destination-guid-2",
						App:  resources.RouteDestinationApp{GUID: "app-guid-2", Process: struct{ Type string }{Type: constant.ProcessTypeWeb}},
					},
					{
						GUID: "destination-guid-3",
						App:  resources.RouteDestinationApp{GUID: "app-guid-3", Process: struct{ Type string }{Type: constant.ProcessTypeWeb}},
					},
				}
			})

			It("returns an error and warnings", func() {
				Expect(destination).To(Equal(resources.RouteDestination{}))
				Expect(executeErr).To(MatchError(actionerror.RouteDestinationNotFoundError{
					AppGUID:     appGUID,
					ProcessType: constant.ProcessTypeWeb,
					RouteGUID:   routeGUID,
				}))
			})
		})
	})

	Describe("DeleteRoute", func() {
		When("deleting a route", func() {
			var (
				warnings   Warnings
				executeErr error
			)
			BeforeEach(func() {
				fakeCloudControllerClient.GetDomainsReturns(
					[]resources.Domain{
						{GUID: "domain-guid", Name: "domain.com"},
					},
					ccv3.Warnings{"get-domains-warning"},
					nil,
				)

				fakeCloudControllerClient.GetRoutesReturns(
					[]resources.Route{
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

			When("in the HTTP case", func() {
				BeforeEach(func() {
					warnings, executeErr = actor.DeleteRoute("domain.com", "hostname", "/path", 0)
				})

				It("delegates to the cloud controller client", func() {
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
						{Key: ccv3.DomainGUIDFilter, Values: []string{"domain-guid"}},
						{Key: ccv3.HostsFilter, Values: []string{"hostname"}},
						{Key: ccv3.PathsFilter, Values: []string{"/path"}},
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
			})

			When("in the TCP case", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetDomainsReturns(
						[]resources.Domain{
							{
								GUID:      "domain-guid",
								Name:      "domain.com",
								Protocols: []string{"tcp"},
							},
						},
						ccv3.Warnings{"get-domains-warning"},
						nil,
					)

					warnings, executeErr = actor.DeleteRoute("domain.com", "", "", 1026)
				})

				It("delegates to the cloud controller client", func() {
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
						{Key: ccv3.DomainGUIDFilter, Values: []string{"domain-guid"}},
						{Key: ccv3.HostsFilter, Values: []string{""}},
						{Key: ccv3.PathsFilter, Values: []string{""}},
						{Key: ccv3.PortsFilter, Values: []string{"1026"}},
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
					warnings, err := actor.DeleteRoute("domain.com", "hostname", "path", 0)
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
					warnings, err := actor.DeleteRoute("domain.com", "hostname", "path", 0)
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
					warnings, err := actor.DeleteRoute("domain.com", "hostname", "path", 0)
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
					warnings, err := actor.DeleteRoute("domain.com", "hostname", "path", 0)
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
						[]resources.Route{},
						ccv3.Warnings{"get-routes-warning"},
						nil,
					)
				})

				It("returns the error", func() {
					warnings, err := actor.DeleteRoute("domain.com", "hostname", "/path", 0)
					Expect(err).To(Equal(actionerror.RouteNotFoundError{
						DomainName: "domain.com",
						DomainGUID: "domain-guid",
						Host:       "hostname",
						Path:       "/path",
						Port:       0,
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
			hostname   string
			path       string
			port       int

			executeErr error
			warnings   Warnings
			route      resources.Route
			domain     resources.Domain
		)

		BeforeEach(func() {
			hostname = ""
			path = ""
			port = 0
		})

		JustBeforeEach(func() {
			route, warnings, executeErr = actor.GetRouteByAttributes(domain, hostname, path, port)
		})

		When("dealing with HTTP routes", func() {
			BeforeEach(func() {
				domain = resources.Domain{
					Name:      domainName,
					GUID:      domainGUID,
					Protocols: []string{"http"},
				}
				hostname = "hostname"
				path = "/path"
			})

			When("The cc client errors", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetRoutesReturns(nil, ccv3.Warnings{"get-routes-warning"}, errors.New("scooby"))
				})
				It("returns and empty route, warnings, and the error", func() {
					Expect(fakeCloudControllerClient.GetRoutesCallCount()).To(Equal(1))

					Expect(warnings).To(ConsistOf("get-routes-warning"))
					Expect(executeErr).To(MatchError(errors.New("scooby")))
				})
			})

			When("the cc client succeeds and a route is found", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetRoutesReturns([]resources.Route{{
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
					Expect(route).To(Equal(resources.Route{
						DomainGUID: domainGUID,
						Host:       hostname,
						Path:       path,
					}))
				})
			})

			When("the cc client succeeds and a route is not found", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetRoutesReturns([]resources.Route{}, ccv3.Warnings{"get-routes-warning"}, nil)
				})

				It("returns the route and the warnings", func() {
					Expect(fakeCloudControllerClient.GetRoutesCallCount()).To(Equal(1))

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

		When("dealing with TCP routes", func() {
			BeforeEach(func() {
				domain = resources.Domain{
					Name:      domainName,
					GUID:      domainGUID,
					Protocols: []string{"tcp"},
				}
				port = 1028
			})

			When("The cc client errors", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetRoutesReturns(nil, ccv3.Warnings{"get-routes-warning"}, errors.New("scooby"))
				})

				It("returns and empty route, warnings, and the error", func() {
					Expect(fakeCloudControllerClient.GetRoutesCallCount()).To(Equal(1))

					Expect(warnings).To(ConsistOf("get-routes-warning"))
					Expect(executeErr).To(MatchError(errors.New("scooby")))
				})
			})

			When("the cc client succeeds and a route is found", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetRoutesReturns([]resources.Route{{
						DomainGUID: domainGUID,
						Port:       port,
					}}, ccv3.Warnings{"get-routes-warning"}, nil)
				})

				It("returns the route and the warnings", func() {
					Expect(fakeCloudControllerClient.GetRoutesCallCount()).To(Equal(1))
					actualQueries := fakeCloudControllerClient.GetRoutesArgsForCall(0)
					Expect(actualQueries).To(ConsistOf(
						ccv3.Query{Key: ccv3.DomainGUIDFilter, Values: []string{domainGUID}},
						ccv3.Query{Key: ccv3.PortsFilter, Values: []string{fmt.Sprintf("%d", port)}},
						ccv3.Query{Key: ccv3.HostsFilter, Values: []string{""}},
						ccv3.Query{Key: ccv3.PathsFilter, Values: []string{""}},
					))

					Expect(warnings).To(ConsistOf("get-routes-warning"))
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(route).To(Equal(resources.Route{
						DomainGUID: domainGUID,
						Port:       port,
					}))
				})
			})

			When("the cc client succeeds and a route is not found", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetRoutesReturns([]resources.Route{}, ccv3.Warnings{"get-routes-warning"}, nil)
				})
				It("returns the route and the warnings", func() {
					Expect(fakeCloudControllerClient.GetRoutesCallCount()).To(Equal(1))

					Expect(warnings).To(ConsistOf("get-routes-warning"))
					Expect(executeErr).To(MatchError(actionerror.RouteNotFoundError{
						DomainName: domainName,
						DomainGUID: domainGUID,
						Port:       port,
					}))
				})
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

			routes     []resources.Route
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
					[]resources.Route{},
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
				fakeCloudControllerClient.GetApplicationRoutesReturns(
					[]resources.Route{
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
				Expect(warnings).To(ConsistOf("get-application-routes-warning"))

				Expect(fakeCloudControllerClient.GetApplicationRoutesCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetApplicationRoutesArgsForCall(0)).To(Equal(appGUID))

				Expect(routes).To(ConsistOf(
					resources.Route{
						GUID:       "some-route-guid",
						URL:        "some-url.sh",
						SpaceGUID:  "routes-space-guid",
						DomainGUID: "routes-domain-guid",
					},
				))
			})

			When("no routes are returned", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationRoutesReturns(
						[]resources.Route{},
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
		})
	})
})
