package v2action_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Route Actions", func() {
	var (
		actor                     Actor
		fakeCloudControllerClient *v2actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v2actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil)
	})

	Describe("GetOrphanedRoutesBySpace", func() {
		BeforeEach(func() {
			fakeCloudControllerClient.GetRouteApplicationsStub = func(routeGUID string, queries []ccv2.Query) ([]ccv2.Application, ccv2.Warnings, error) {
				switch routeGUID {
				case "orphaned-route-guid-1":
					return []ccv2.Application{}, nil, nil
				case "orphaned-route-guid-2":
					return []ccv2.Application{}, nil, nil
				case "not-orphaned-route-guid-3":
					return []ccv2.Application{
						{GUID: "app-guid"},
					}, nil, nil
				}
				Fail("Unexpected route-guid")
				return []ccv2.Application{}, nil, nil
			}
		})

		Context("when there are orphaned routes", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpaceRoutesReturns([]ccv2.Route{
					{
						GUID:       "orphaned-route-guid-1",
						DomainGUID: "some-domain-guid",
					},
					{
						GUID:       "orphaned-route-guid-2",
						DomainGUID: "some-other-domain-guid",
					},
					{
						GUID:       "not-orphaned-route-guid-3",
						DomainGUID: "not-orphaned-route-domain-guid",
					},
				}, nil, nil)
				fakeCloudControllerClient.GetSharedDomainStub = func(domainGUID string) (ccv2.Domain, ccv2.Warnings, error) {
					switch domainGUID {
					case "some-domain-guid":
						return ccv2.Domain{
							GUID: "some-domain-guid",
							Name: "some-domain.com",
						}, nil, nil
					case "some-other-domain-guid":
						return ccv2.Domain{
							GUID: "some-other-domain-guid",
							Name: "some-other-domain.com",
						}, nil, nil
					case "not-orphaned-route-domain-guid":
						return ccv2.Domain{
							GUID: "not-orphaned-route-domain-guid",
							Name: "not-orphaned-route-domain.com",
						}, nil, nil
					}
					return ccv2.Domain{}, nil, errors.New("Unexpected domain GUID")
				}
			})

			It("returns the orphaned routes with the domain names", func() {
				orphanedRoutes, _, err := actor.GetOrphanedRoutesBySpace("space-guid")
				Expect(err).NotTo(HaveOccurred())
				Expect(orphanedRoutes).To(ConsistOf([]Route{
					{
						GUID:   "orphaned-route-guid-1",
						Domain: "some-domain.com",
					},
					{
						GUID:   "orphaned-route-guid-2",
						Domain: "some-other-domain.com",
					},
				}))

				Expect(fakeCloudControllerClient.GetSpaceRoutesCallCount()).To(Equal(1))

				spaceGUID, queries := fakeCloudControllerClient.GetSpaceRoutesArgsForCall(0)
				Expect(spaceGUID).To(Equal("space-guid"))
				Expect(queries).To(BeNil())

				Expect(fakeCloudControllerClient.GetRouteApplicationsCallCount()).To(Equal(3))

				routeGUID, queries := fakeCloudControllerClient.GetRouteApplicationsArgsForCall(0)
				Expect(routeGUID).To(Equal("orphaned-route-guid-1"))
				Expect(queries).To(BeNil())

				routeGUID, queries = fakeCloudControllerClient.GetRouteApplicationsArgsForCall(1)
				Expect(routeGUID).To(Equal("orphaned-route-guid-2"))
				Expect(queries).To(BeNil())

				routeGUID, queries = fakeCloudControllerClient.GetRouteApplicationsArgsForCall(2)
				Expect(routeGUID).To(Equal("not-orphaned-route-guid-3"))
				Expect(queries).To(BeNil())
			})
		})

		Context("when there are no orphaned routes", func() {
			var expectedErr OrphanedRoutesNotFoundError

			BeforeEach(func() {
				fakeCloudControllerClient.GetSpaceRoutesReturns([]ccv2.Route{
					ccv2.Route{GUID: "not-orphaned-route-guid-3"},
				}, nil, nil)
			})

			It("returns an OrphanedRoutesNotFoundError", func() {
				orphanedRoutes, _, err := actor.GetOrphanedRoutesBySpace("space-guid")
				Expect(err).To(MatchError(expectedErr))
				Expect(orphanedRoutes).To(BeNil())

				Expect(fakeCloudControllerClient.GetSpaceRoutesCallCount()).To(Equal(1))

				spaceGUID, queries := fakeCloudControllerClient.GetSpaceRoutesArgsForCall(0)
				Expect(spaceGUID).To(Equal("space-guid"))
				Expect(queries).To(BeNil())

				Expect(fakeCloudControllerClient.GetRouteApplicationsCallCount()).To(Equal(1))

				routeGUID, queries := fakeCloudControllerClient.GetRouteApplicationsArgsForCall(0)
				Expect(routeGUID).To(Equal("not-orphaned-route-guid-3"))
				Expect(queries).To(BeNil())
			})
		})

		Context("when there are warnings", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpaceRoutesReturns([]ccv2.Route{
					ccv2.Route{GUID: "route-guid-1"},
					ccv2.Route{GUID: "route-guid-2"},
				}, ccv2.Warnings{"get-routes-warning"}, nil)
				fakeCloudControllerClient.GetRouteApplicationsReturns(nil, ccv2.Warnings{"get-applications-warning"}, nil)
				fakeCloudControllerClient.GetSharedDomainReturns(ccv2.Domain{GUID: "some-guid"}, ccv2.Warnings{"get-shared-domain-warning"}, nil)
			})

			It("returns all the warnings", func() {
				_, warnings, err := actor.GetOrphanedRoutesBySpace("space-guid")
				Expect(err).NotTo(HaveOccurred())
				Expect(warnings).To(ConsistOf("get-routes-warning", "get-applications-warning", "get-shared-domain-warning", "get-applications-warning", "get-shared-domain-warning"))
			})
		})

		Context("when the spaces routes API request returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("spaces routes error")
				fakeCloudControllerClient.GetSpaceRoutesReturns(nil, nil, expectedErr)
			})

			It("returns the error", func() {
				routes, _, err := actor.GetOrphanedRoutesBySpace("space-guid")
				Expect(err).To(Equal(expectedErr))
				Expect(routes).To(BeNil())
			})
		})

		Context("when a route's applications API request returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("application error")
				fakeCloudControllerClient.GetSpaceRoutesReturns([]ccv2.Route{
					ccv2.Route{GUID: "route-guid"},
				}, nil, nil)
				fakeCloudControllerClient.GetRouteApplicationsReturns(nil, nil, expectedErr)
			})

			It("returns the error", func() {
				routes, _, err := actor.GetOrphanedRoutesBySpace("space-guid")
				Expect(err).To(Equal(expectedErr))
				Expect(routes).To(BeNil())
			})
		})
	})

	Describe("DeleteRoute", func() {
		Context("when the route exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.DeleteRouteReturns(nil, nil)
			})

			It("deletes the route", func() {
				_, err := actor.DeleteRoute("some-route-guid")
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeCloudControllerClient.DeleteRouteCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.DeleteRouteArgsForCall(0)).To(Equal("some-route-guid"))
			})
		})

		Context("when the API returns both warnings and an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("bananahammock")
				fakeCloudControllerClient.DeleteRouteReturns(ccv2.Warnings{"foo", "bar"}, expectedErr)
			})

			It("returns both the warnings and the error", func() {
				warnings, err := actor.DeleteRoute("some-route-guid")
				Expect(err).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("foo", "bar"))
			})
		})
	})

	Describe("GetSpaceRoutes", func() {
		Context("when the CC API client does not return any errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpaceRoutesReturns([]ccv2.Route{
					ccv2.Route{
						GUID:       "route-guid-1",
						Host:       "host",
						Path:       "/path",
						Port:       1234,
						DomainGUID: "domain-1-guid",
					},
					ccv2.Route{
						GUID:       "route-guid-2",
						Host:       "host",
						Path:       "/path",
						Port:       1234,
						DomainGUID: "domain-2-guid",
					},
				}, ccv2.Warnings{"get-space-routes-warning"}, nil)
				fakeCloudControllerClient.GetSharedDomainReturns(
					ccv2.Domain{
						Name: "domain",
					}, nil, nil)
			})

			It("returns the space routes and any warnings", func() {
				routes, warnings, err := actor.GetSpaceRoutes("space-guid", nil)
				Expect(fakeCloudControllerClient.GetSpaceRoutesCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetSpaceRoutesArgsForCall(0)).To(Equal("space-guid"))
				Expect(fakeCloudControllerClient.GetSharedDomainCallCount()).To(Equal(2))
				Expect(fakeCloudControllerClient.GetSharedDomainArgsForCall(0)).To(Equal("domain-1-guid"))
				Expect(fakeCloudControllerClient.GetSharedDomainArgsForCall(1)).To(Equal("domain-2-guid"))

				Expect(warnings).To(ConsistOf("get-space-routes-warning"))
				Expect(err).NotTo(HaveOccurred())
				Expect(routes).To(ConsistOf([]Route{
					{
						GUID:   "route-guid-1",
						Host:   "host",
						Domain: "domain",
						Path:   "/path",
						Port:   1234,
					},
					{
						GUID:   "route-guid-2",
						Host:   "host",
						Domain: "domain",
						Path:   "/path",
						Port:   1234,
					},
				}))
			})
		})

		Context("when the CC API client returns an error", func() {
			Context("when getting space routes returns an error and warnings", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetSpaceRoutesReturns(
						[]ccv2.Route{}, ccv2.Warnings{"space-routes-warning"}, errors.New("get-space-routes-error"))
				})

				It("returns the error and warnings", func() {
					routes, warnings, err := actor.GetSpaceRoutes("space-guid", nil)
					Expect(warnings).To(ConsistOf("space-routes-warning"))
					Expect(err).To(MatchError("get-space-routes-error"))
					Expect(routes).To(BeNil())
				})
			})

			Context("when getting the domain returns an error and warnings", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetSpaceRoutesReturns([]ccv2.Route{
						ccv2.Route{
							GUID:       "route-guid-1",
							Host:       "host",
							Path:       "/path",
							Port:       1234,
							DomainGUID: "domain-1-guid",
						},
					}, nil, nil)
					fakeCloudControllerClient.GetSharedDomainReturns(ccv2.Domain{}, ccv2.Warnings{"domain-warning"}, errors.New("get-domain-error"))
				})

				It("returns the error and warnings", func() {
					routes, warnings, err := actor.GetSpaceRoutes("space-guid", nil)
					Expect(fakeCloudControllerClient.GetSharedDomainCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetSharedDomainArgsForCall(0)).To(Equal("domain-1-guid"))

					Expect(warnings).To(ConsistOf("domain-warning"))
					Expect(err).To(MatchError("get-domain-error"))
					Expect(routes).To(BeNil())
				})
			})
		})

		Context("when the CC API client returns warnings and no errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpaceRoutesReturns([]ccv2.Route{
					ccv2.Route{
						GUID:       "route-guid-1",
						Host:       "host",
						Path:       "/path",
						Port:       1234,
						DomainGUID: "domain-1-guid",
					},
				}, ccv2.Warnings{"space-routes-warning"}, nil)
				fakeCloudControllerClient.GetSharedDomainReturns(ccv2.Domain{}, ccv2.Warnings{"domain-warning"}, nil)
			})

			It("returns the warnings", func() {
				_, warnings, _ := actor.GetSpaceRoutes("space-guid", nil)
				Expect(warnings).To(ConsistOf("space-routes-warning", "domain-warning"))
			})
		})

		Context("when a query parameter exists", func() {
			It("passes the query to the client", func() {
				expectedQuery := []ccv2.Query{
					{
						Filter:   "space_guid",
						Operator: ":",
						Value:    "space-guid",
					}}

				_, _, err := actor.GetSpaceRoutes("space-guid", expectedQuery)
				Expect(err).ToNot(HaveOccurred())
				_, query := fakeCloudControllerClient.GetSpaceRoutesArgsForCall(0)
				Expect(query).To(Equal(expectedQuery))
			})
		})
	})

	Describe("Route", func() {
		DescribeTable("String", func(host string, domain string, path string, port int, expectedValue string) {
			route := Route{
				Host:   host,
				Domain: domain,
				Path:   path,
				Port:   port,
			}
			Expect(route.String()).To(Equal(expectedValue))
		},

			Entry("has domain", "", "domain.com", "", 0, "domain.com"),
			Entry("has host, domain", "host", "domain.com", "", 0, "host.domain.com"),
			Entry("has domain, path", "", "domain.com", "/path", 0, "domain.com/path"),
			Entry("has host, domain, path", "host", "domain.com", "/path", 0, "host.domain.com/path"),
			Entry("has domain, port", "", "domain.com", "", 3333, "domain.com:3333"),
			Entry("has host, domain, path, port", "host", "domain.com", "/path", 3333, "domain.com:3333"),
		)
	})
})
