package v2action_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Route Actions", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v2actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v2actionfakes.FakeCloudControllerClient)
		actor = NewActor(fakeCloudControllerClient, nil, nil)
	})

	Describe("Route", func() {
		DescribeTable("String", func(host string, domain string, path string, port int, expectedValue string) {
			route := Route{
				Host: host,
				Domain: Domain{
					Name: domain,
				},
				Path: path,
				Port: port,
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

	Describe("BindRouteToApplication", func() {
		Context("when no errors are encountered", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.BindRouteToApplicationReturns(
					ccv2.Route{},
					ccv2.Warnings{"bind warning"},
					nil)
			})

			It("binds the route to the application and returns all warnings", func() {
				warnings, err := actor.BindRouteToApplication("some-route-guid", "some-app-guid")
				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("bind warning"))

				Expect(fakeCloudControllerClient.BindRouteToApplicationCallCount()).To(Equal(1))
				routeGUID, appGUID := fakeCloudControllerClient.BindRouteToApplicationArgsForCall(0)
				Expect(routeGUID).To(Equal("some-route-guid"))
				Expect(appGUID).To(Equal("some-app-guid"))
			})
		})

		Context("when an error is encountered", func() {
			Context("InvalidRelationError", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.BindRouteToApplicationReturns(
						ccv2.Route{},
						ccv2.Warnings{"bind warning"},
						ccerror.InvalidRelationError{})
				})

				It("returns the error", func() {
					warnings, err := actor.BindRouteToApplication("some-route-guid", "some-app-guid")
					Expect(err).To(MatchError(RouteInDifferentSpaceError{}))
					Expect(warnings).To(ConsistOf("bind warning"))
				})
			})

			Context("generic error", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("bind route failed")
					fakeCloudControllerClient.BindRouteToApplicationReturns(
						ccv2.Route{},
						ccv2.Warnings{"bind warning"},
						expectedErr)
				})

				It("returns the error", func() {
					warnings, err := actor.BindRouteToApplication("some-route-guid", "some-app-guid")
					Expect(err).To(MatchError(expectedErr))
					Expect(warnings).To(ConsistOf("bind warning"))
				})
			})
		})
	})

	Describe("CreateRoute", func() {
		Context("when no errors are encountered", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CreateRouteReturns(
					ccv2.Route{
						GUID:       "some-route-guid",
						Host:       "some-host",
						Path:       "some-path",
						Port:       3333,
						DomainGUID: "some-domain-guid",
						SpaceGUID:  "some-space-guid",
					},
					ccv2.Warnings{"create route warning"},
					nil)
			})

			It("creates the route and returns all warnings", func() {
				route, warnings, err := actor.CreateRoute(
					Route{
						Domain: Domain{
							Name: "some-domain",
							GUID: "some-domain-guid",
						},
						Host:      "some-host",
						Path:      "some-path",
						Port:      3333,
						SpaceGUID: "some-space-guid",
					},
					true)
				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("create route warning"))
				Expect(route).To(Equal(Route{
					Domain: Domain{
						Name: "some-domain",
						GUID: "some-domain-guid",
					},
					GUID:      "some-route-guid",
					Host:      "some-host",
					Path:      "some-path",
					Port:      3333,
					SpaceGUID: "some-space-guid",
				}))

				Expect(fakeCloudControllerClient.CreateRouteCallCount()).To(Equal(1))
				passedRoute, generatePort := fakeCloudControllerClient.CreateRouteArgsForCall(0)
				Expect(passedRoute).To(Equal(ccv2.Route{
					DomainGUID: "some-domain-guid",
					Host:       "some-host",
					Path:       "some-path",
					Port:       3333,
					SpaceGUID:  "some-space-guid",
				}))
				Expect(generatePort).To(BeTrue())
			})
		})

		Context("when an error is encountered", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("bind route failed")
				fakeCloudControllerClient.CreateRouteReturns(
					ccv2.Route{},
					ccv2.Warnings{"create route warning"},
					expectedErr)
			})

			It("returns the error", func() {
				_, warnings, err := actor.CreateRoute(Route{}, true)
				Expect(err).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("create route warning"))
			})
		})
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
						GUID: "orphaned-route-guid-1",
						Domain: Domain{
							Name: "some-domain.com",
							GUID: "some-domain-guid",
						},
					},
					{
						GUID: "orphaned-route-guid-2",
						Domain: Domain{
							Name: "some-other-domain.com",
							GUID: "some-other-domain-guid",
						},
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

	Describe("GetApplicationRoutes", func() {
		Context("when the CC API client does not return any errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationRoutesReturns([]ccv2.Route{
					ccv2.Route{
						GUID:       "route-guid-1",
						SpaceGUID:  "some-space-guid",
						Host:       "host",
						Path:       "/path",
						Port:       1234,
						DomainGUID: "domain-1-guid",
					},
					ccv2.Route{
						GUID:       "route-guid-2",
						SpaceGUID:  "some-space-guid",
						Host:       "host",
						Path:       "/path",
						Port:       1234,
						DomainGUID: "domain-2-guid",
					},
				}, ccv2.Warnings{"get-application-routes-warning"}, nil)

				fakeCloudControllerClient.GetSharedDomainReturnsOnCall(0, ccv2.Domain{Name: "domain.com"}, nil, nil)
				fakeCloudControllerClient.GetSharedDomainReturnsOnCall(1, ccv2.Domain{Name: "other-domain.com"}, nil, nil)
			})

			It("returns the application routes and any warnings", func() {
				routes, warnings, err := actor.GetApplicationRoutes("application-guid")
				Expect(fakeCloudControllerClient.GetApplicationRoutesCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetApplicationRoutesArgsForCall(0)).To(Equal("application-guid"))
				Expect(fakeCloudControllerClient.GetSharedDomainCallCount()).To(Equal(2))
				Expect(fakeCloudControllerClient.GetSharedDomainArgsForCall(0)).To(Equal("domain-1-guid"))
				Expect(fakeCloudControllerClient.GetSharedDomainArgsForCall(1)).To(Equal("domain-2-guid"))

				Expect(warnings).To(ConsistOf("get-application-routes-warning"))
				Expect(err).NotTo(HaveOccurred())
				Expect(routes).To(ConsistOf([]Route{
					{
						Domain: Domain{
							Name: "domain.com",
						},
						GUID:      "route-guid-1",
						Host:      "host",
						Path:      "/path",
						Port:      1234,
						SpaceGUID: "some-space-guid",
					},
					{
						Domain: Domain{
							Name: "other-domain.com",
						},
						GUID:      "route-guid-2",
						Host:      "host",
						Path:      "/path",
						Port:      1234,
						SpaceGUID: "some-space-guid",
					},
				}))
			})
		})

		Context("when the CC API client returns an error", func() {
			Context("when getting application routes returns an error and warnings", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationRoutesReturns(
						[]ccv2.Route{}, ccv2.Warnings{"application-routes-warning"}, errors.New("get-application-routes-error"))
				})

				It("returns the error and warnings", func() {
					routes, warnings, err := actor.GetApplicationRoutes("application-guid")
					Expect(warnings).To(ConsistOf("application-routes-warning"))
					Expect(err).To(MatchError("get-application-routes-error"))
					Expect(routes).To(BeNil())
				})
			})

			Context("when getting the domain returns an error and warnings", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetApplicationRoutesReturns([]ccv2.Route{
						ccv2.Route{
							GUID:       "route-guid-1",
							SpaceGUID:  "some-space-guid",
							Host:       "host",
							Path:       "/path",
							Port:       1234,
							DomainGUID: "domain-1-guid",
						},
					}, nil, nil)
					fakeCloudControllerClient.GetSharedDomainReturns(ccv2.Domain{}, ccv2.Warnings{"domain-warning"}, errors.New("get-domain-error"))
				})

				It("returns the error and warnings", func() {
					routes, warnings, err := actor.GetApplicationRoutes("application-guid")
					Expect(warnings).To(ConsistOf("domain-warning"))
					Expect(err).To(MatchError("get-domain-error"))
					Expect(routes).To(BeNil())

					Expect(fakeCloudControllerClient.GetSharedDomainCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetSharedDomainArgsForCall(0)).To(Equal("domain-1-guid"))
				})
			})
		})

		Context("when the CC API client returns warnings and no errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetApplicationRoutesReturns([]ccv2.Route{
					ccv2.Route{
						GUID:       "route-guid-1",
						SpaceGUID:  "some-space-guid",
						Host:       "host",
						Path:       "/path",
						Port:       1234,
						DomainGUID: "domain-1-guid",
					},
				}, ccv2.Warnings{"application-routes-warning"}, nil)
				fakeCloudControllerClient.GetSharedDomainReturns(ccv2.Domain{}, ccv2.Warnings{"domain-warning"}, nil)
			})

			It("returns the warnings", func() {
				_, warnings, _ := actor.GetApplicationRoutes("application-guid")
				Expect(warnings).To(ConsistOf("application-routes-warning", "domain-warning"))
			})
		})
	})

	Describe("GetSpaceRoutes", func() {
		Context("when the CC API client does not return any errors", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetSpaceRoutesReturns([]ccv2.Route{
					ccv2.Route{
						GUID:       "route-guid-1",
						SpaceGUID:  "some-space-guid",
						Host:       "host",
						Path:       "/path",
						Port:       1234,
						DomainGUID: "domain-1-guid",
					},
					ccv2.Route{
						GUID:       "route-guid-2",
						SpaceGUID:  "some-space-guid",
						Host:       "host",
						Path:       "/path",
						Port:       1234,
						DomainGUID: "domain-2-guid",
					},
				}, ccv2.Warnings{"get-space-routes-warning"}, nil)
				fakeCloudControllerClient.GetSharedDomainReturnsOnCall(0, ccv2.Domain{Name: "domain.com"}, nil, nil)
				fakeCloudControllerClient.GetSharedDomainReturnsOnCall(1, ccv2.Domain{Name: "other-domain.com"}, nil, nil)
			})

			It("returns the space routes and any warnings", func() {
				routes, warnings, err := actor.GetSpaceRoutes("space-guid")
				Expect(fakeCloudControllerClient.GetSpaceRoutesCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetSpaceRoutesArgsForCall(0)).To(Equal("space-guid"))
				Expect(fakeCloudControllerClient.GetSharedDomainCallCount()).To(Equal(2))
				Expect(fakeCloudControllerClient.GetSharedDomainArgsForCall(0)).To(Equal("domain-1-guid"))
				Expect(fakeCloudControllerClient.GetSharedDomainArgsForCall(1)).To(Equal("domain-2-guid"))

				Expect(warnings).To(ConsistOf("get-space-routes-warning"))
				Expect(err).NotTo(HaveOccurred())
				Expect(routes).To(ConsistOf([]Route{
					{
						Domain: Domain{
							Name: "domain.com",
						},
						GUID:      "route-guid-1",
						Host:      "host",
						Path:      "/path",
						Port:      1234,
						SpaceGUID: "some-space-guid",
					},
					{
						Domain: Domain{
							Name: "other-domain.com",
						},
						GUID:      "route-guid-2",
						Host:      "host",
						Path:      "/path",
						Port:      1234,
						SpaceGUID: "some-space-guid",
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
					routes, warnings, err := actor.GetSpaceRoutes("space-guid")
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
							SpaceGUID:  "some-space-guid",
							Host:       "host",
							Path:       "/path",
							Port:       1234,
							DomainGUID: "domain-1-guid",
						},
					}, nil, nil)
					fakeCloudControllerClient.GetSharedDomainReturns(ccv2.Domain{}, ccv2.Warnings{"domain-warning"}, errors.New("get-domain-error"))
				})

				It("returns the error and warnings", func() {
					routes, warnings, err := actor.GetSpaceRoutes("space-guid")
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
						SpaceGUID:  "some-space-guid",
						Host:       "host",
						Path:       "/path",
						Port:       1234,
						DomainGUID: "domain-1-guid",
					},
				}, ccv2.Warnings{"space-routes-warning"}, nil)
				fakeCloudControllerClient.GetSharedDomainReturns(ccv2.Domain{}, ccv2.Warnings{"domain-warning"}, nil)
			})

			It("returns the warnings", func() {
				_, warnings, _ := actor.GetSpaceRoutes("space-guid")
				Expect(warnings).To(ConsistOf("space-routes-warning", "domain-warning"))
			})
		})
	})

	Describe("GetRouteByHostAndDomain", func() {
		var (
			host       string
			domainGUID string

			route      Route
			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			host = "some-host"
			domainGUID = "some-domain-guid"
		})

		JustBeforeEach(func() {
			route, warnings, executeErr = actor.GetRouteByHostAndDomain(host, domainGUID)
		})

		Context("when finding the route is successful", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetRoutesReturns([]ccv2.Route{
					{
						GUID:       "route-guid-1",
						SpaceGUID:  "some-space-guid",
						Host:       "host",
						Path:       "/path",
						Port:       1234,
						DomainGUID: "domain-1-guid",
					},
				}, ccv2.Warnings{"get-routes-warning"}, nil)
			})

			Context("when finding the domain is successful", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetSharedDomainReturns(
						ccv2.Domain{
							Name: "domain.com",
						}, ccv2.Warnings{"get-domain-warning"}, nil)
				})

				It("returns the routes and any warnings", func() {
					Expect(warnings).To(ConsistOf("get-routes-warning", "get-domain-warning"))
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(route).To(Equal(Route{
						Domain: Domain{
							Name: "domain.com",
						},
						GUID:      "route-guid-1",
						Host:      "host",
						Path:      "/path",
						Port:      1234,
						SpaceGUID: "some-space-guid",
					}))

					Expect(fakeCloudControllerClient.GetRoutesCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetRoutesArgsForCall(0)).To(Equal([]ccv2.Query{
						{Filter: ccv2.HostFilter, Operator: ccv2.EqualOperator, Value: host},
						{Filter: ccv2.DomainGUIDFilter, Operator: ccv2.EqualOperator, Value: domainGUID},
					}))

					Expect(fakeCloudControllerClient.GetSharedDomainCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetSharedDomainArgsForCall(0)).To(Equal("domain-1-guid"))
				})
			})

			Context("when getting the domain returns an error and warnings", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("get-domain-error")
					fakeCloudControllerClient.GetSharedDomainReturns(ccv2.Domain{}, ccv2.Warnings{"get-domain-warning"}, expectedErr)
				})

				It("returns the error and warnings", func() {
					Expect(executeErr).To(MatchError(expectedErr))
					Expect(warnings).To(ConsistOf("get-routes-warning", "get-domain-warning"))
				})
			})
		})

		Context("when getting routes returns an error and warnings", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("get-routes-err")
				fakeCloudControllerClient.GetRoutesReturns([]ccv2.Route{}, ccv2.Warnings{"get-routes-warning"}, expectedErr)
			})

			It("returns the error and warnings", func() {
				Expect(executeErr).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("get-routes-warning"))
			})
		})

		Context("when no route is found", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetRoutesReturns([]ccv2.Route{}, ccv2.Warnings{"get-routes-warning"}, nil)
			})

			It("returns a RouteNotFoundError and warnings", func() {
				Expect(executeErr).To(MatchError(RouteNotFoundError{Host: host, DomainGUID: domainGUID}))
				Expect(warnings).To(ConsistOf("get-routes-warning"))
			})
		})
	})

	Describe("CheckRoute", func() {
		Context("when the API calls succeed", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.CheckRouteReturns(true, ccv2.Warnings{"some-check-route-warnings"}, nil)
			})

			It("returns the bool and warnings", func() {
				exists, warnings, err := actor.CheckRoute(Route{
					Host: "some-host",
					Domain: Domain{
						GUID: "some-domain-guid",
					},
					Path: "some-path",
					Port: 42,
				})

				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("some-check-route-warnings"))
				Expect(exists).To(BeTrue())

				Expect(fakeCloudControllerClient.CheckRouteCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.CheckRouteArgsForCall(0)).To(Equal(ccv2.Route{
					Host:       "some-host",
					DomainGUID: "some-domain-guid",
					Path:       "some-path",
					Port:       42,
				}))
			})
		})

		Context("when the cc returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("booo")
				fakeCloudControllerClient.CheckRouteReturns(false, ccv2.Warnings{"some-check-route-warnings"}, expectedErr)
			})

			It("returns the bool and warnings", func() {
				exists, warnings, err := actor.CheckRoute(Route{
					Host: "some-host",
					Domain: Domain{
						GUID: "some-domain-guid",
					},
				})

				Expect(err).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("some-check-route-warnings"))
				Expect(exists).To(BeFalse())
			})
		})
	})

	Describe("FindRouteBoundToSpaceWithSettings", func() {
		var (
			route Route

			returnedRoute Route
			warnings      Warnings
			executeErr    error
		)

		BeforeEach(func() {
			route = Route{
				Domain: Domain{
					Name: "some-domain.com",
					GUID: "some-domain-guid",
				},
				Host:      "some-host",
				SpaceGUID: "some-space-guid",
			}

			fakeCloudControllerClient.GetSharedDomainReturns(
				ccv2.Domain{
					GUID: "some-domain-guid",
					Name: "some-domain.com",
				},
				ccv2.Warnings{"get domain warning"},
				nil)
		})

		JustBeforeEach(func() {
			returnedRoute, warnings, executeErr = actor.FindRouteBoundToSpaceWithSettings(route)
		})

		Context("when the route exists in the current space", func() {
			var existingRoute Route

			BeforeEach(func() {
				existingRoute = route
				existingRoute.GUID = "some-route-guid"
				fakeCloudControllerClient.GetRoutesReturns([]ccv2.Route{ActorToCCRoute(existingRoute)}, ccv2.Warnings{"get route warning"}, nil)
			})

			It("returns the route", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(returnedRoute).To(Equal(existingRoute))
				Expect(warnings).To(ConsistOf("get route warning", "get domain warning"))
			})
		})

		Context("when the route exists in a different space", func() {
			Context("when the user has access to the route", func() {
				BeforeEach(func() {
					existingRoute := route
					existingRoute.GUID = "some-route-guid"
					existingRoute.SpaceGUID = "some-other-space-guid"
					fakeCloudControllerClient.GetRoutesReturns([]ccv2.Route{ActorToCCRoute(existingRoute)}, ccv2.Warnings{"get route warning"}, nil)
				})

				It("returns a RouteInDifferentSpaceError", func() {
					Expect(executeErr).To(MatchError(RouteInDifferentSpaceError{Route: route.String()}))
					Expect(warnings).To(ConsistOf("get route warning", "get domain warning"))
				})
			})

			Context("when the user does not have access to the route", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetRoutesReturns([]ccv2.Route{}, ccv2.Warnings{"get route warning"}, nil)
					fakeCloudControllerClient.CheckRouteReturns(true, ccv2.Warnings{"check route warning"}, nil)
				})

				It("returns a RouteInDifferentSpaceError", func() {
					Expect(executeErr).To(MatchError(RouteInDifferentSpaceError{Route: route.String()}))
					Expect(warnings).To(ConsistOf("get route warning", "check route warning"))
				})
			})
		})

		Context("when the route does not exist", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = RouteNotFoundError{Host: route.Host, DomainGUID: route.Domain.GUID}
				fakeCloudControllerClient.GetRoutesReturns([]ccv2.Route{}, ccv2.Warnings{"get route warning"}, nil)
				fakeCloudControllerClient.CheckRouteReturns(false, ccv2.Warnings{"check route warning"}, nil)
			})

			It("returns the route", func() {
				Expect(executeErr).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("get route warning", "check route warning"))
			})
		})

		Context("when finding the route errors", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("booo")
				fakeCloudControllerClient.GetRoutesReturns(nil, ccv2.Warnings{"get route warning"}, expectedErr)
			})

			It("the error and warnings", func() {
				Expect(executeErr).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("get route warning"))
			})
		})
	})
})
