package v2actions_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actors/v2actions"
	"code.cloudfoundry.org/cli/actors/v2actions/v2actionsfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Route Actions", func() {
	var (
		actor                     Actor
		fakeCloudControllerClient *v2actionsfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v2actionsfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient)
	})

	Describe("GetOrphanedRoutesBySpace", func() {
		BeforeEach(func() {
			fakeCloudControllerClient.GetRouteApplicationsStub = func(routeGUID string) ([]ccv2.Application, ccv2.Warnings, error) {
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
						GUID:         "orphaned-route-guid-1",
						DomainFields: ccv2.Domain{GUID: "some-domain-guid", Name: " some-domain.com"},
					},
					{
						GUID:         "orphaned-route-guid-2",
						DomainFields: ccv2.Domain{GUID: "some-other-domain-guid", Name: "some-other-domain.com"},
					},
					{
						GUID: "not-orphaned-route-guid-3",
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
				Expect(fakeCloudControllerClient.GetSpaceRoutesArgsForCall(0)).To(Equal("space-guid"))
				Expect(fakeCloudControllerClient.GetRouteApplicationsCallCount()).To(Equal(3))
				Expect(fakeCloudControllerClient.GetRouteApplicationsArgsForCall(0)).To(Equal("orphaned-route-guid-1"))
				Expect(fakeCloudControllerClient.GetRouteApplicationsArgsForCall(1)).To(Equal("orphaned-route-guid-2"))
				Expect(fakeCloudControllerClient.GetRouteApplicationsArgsForCall(2)).To(Equal("not-orphaned-route-guid-3"))
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
				Expect(fakeCloudControllerClient.GetSpaceRoutesArgsForCall(0)).To(Equal("space-guid"))
				Expect(fakeCloudControllerClient.GetRouteApplicationsCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetRouteApplicationsArgsForCall(0)).To(Equal("not-orphaned-route-guid-3"))
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
				_, _, err := actor.GetOrphanedRoutesBySpace("space-guid")
				Expect(err).To(Equal(expectedErr))
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
				_, _, err := actor.GetOrphanedRoutesBySpace("space-guid")
				Expect(err).To(Equal(expectedErr))
			})
		})
	})

	Describe("DeleteRouteByGUID", func() {
		Context("when the route exists", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.DeleteRouteReturns(nil, nil)
			})

			It("deletes the route", func() {
				_, err := actor.DeleteRouteByGUID("some-route-guid")
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
				warnings, err := actor.DeleteRouteByGUID("some-route-guid")
				Expect(err).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("foo", "bar"))
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
			actualValue := route.String()

			Expect(actualValue).To(Equal(expectedValue))
		},

			Entry("has domain", "", "domain.com", "", 0, "domain.com"),
			Entry("has host, domain", "host", "domain.com", "", 0, "host.domain.com"),
			Entry("has domain, path", "", "domain.com", "path", 0, "domain.com/path"),
			Entry("has host, domain, path", "host", "domain.com", "path", 0, "host.domain.com/path"),
			Entry("has domain, port", "", "domain.com", "", 3333, "domain.com:3333"),
		)
	})
})
