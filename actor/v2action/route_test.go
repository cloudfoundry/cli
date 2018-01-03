package v2action_test

import (
	"errors"
	"fmt"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/v2action"
	"code.cloudfoundry.org/cli/actor/v2action/v2actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccerror"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2/constant"
	"code.cloudfoundry.org/cli/types"
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
		DescribeTable("String",
			func(host string, domain string, path string, port types.NullInt, expectedValue string) {
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

			Entry("has domain", "", "domain.com", "", types.NullInt{IsSet: false}, "domain.com"),
			Entry("has host, domain", "host", "domain.com", "", types.NullInt{IsSet: false}, "host.domain.com"),
			Entry("has domain, path", "", "domain.com", "/path", types.NullInt{IsSet: false}, "domain.com/path"),
			Entry("has domain, path", "", "domain.com", "path", types.NullInt{IsSet: false}, "domain.com/path"),
			Entry("has host, domain, path", "host", "domain.com", "/path", types.NullInt{IsSet: false}, "host.domain.com/path"),
			Entry("has domain, port", "", "domain.com", "", types.NullInt{IsSet: true, Value: 3333}, "domain.com:3333"),
			Entry("has host, domain, path, port", "host", "domain.com", "/path", types.NullInt{IsSet: true, Value: 3333}, "host.domain.com:3333/path"),
		)

		Describe("RandomTCPPort", func() {
			var (
				route  Route
				domain Domain
				port   types.NullInt
			)

			JustBeforeEach(func() {
				route = Route{
					Domain: domain,
					Port:   port,
				}
			})

			Context("when the domain is a tcp domain and there is no port specified", func() {
				BeforeEach(func() {
					domain = Domain{
						RouterGroupType: constant.TCPRouterGroup,
					}
					port = types.NullInt{}
				})

				It("returns true", func() {
					Expect(route.RandomTCPPort()).To(BeTrue())
				})
			})

			Context("when the domain is a tcp domain and the port is specified", func() {
				BeforeEach(func() {
					domain = Domain{
						RouterGroupType: constant.TCPRouterGroup,
					}
					port = types.NullInt{IsSet: true}
				})

				It("returns false", func() {
					Expect(route.RandomTCPPort()).To(BeFalse())
				})
			})

			Context("when the domain is a not tcp domain", func() {
				BeforeEach(func() {
					domain = Domain{}
					port = types.NullInt{}
				})

				It("returns false", func() {
					Expect(route.RandomTCPPort()).To(BeFalse())
				})
			})
		})

		DescribeTable("Validate",
			func(route Route, expectedErr error) {
				err := route.Validate()
				if expectedErr == nil {
					Expect(err).To(BeNil())
				} else {
					Expect(err).To(Equal(expectedErr))
				}
			},

			Entry("valid - host and path on HTTP domain",
				Route{
					Host: "some-host",
					Path: "some-path",
					Domain: Domain{
						Name: "some-domain",
					},
				},
				nil,
			),

			Entry("valid - port on TCP domain",
				Route{
					Port: types.NullInt{IsSet: true},
					Domain: Domain{
						Name:            "some-domain",
						RouterGroupType: constant.TCPRouterGroup,
					},
				},
				nil,
			),

			Entry("error - no host on shared HTTP domain",
				Route{
					Path: "some-path",
					Domain: Domain{
						Name: "some-domain",
						Type: constant.SharedDomain,
					},
				},
				actionerror.NoHostnameAndSharedDomainError{},
			),

			Entry("error - port on HTTP domain",
				Route{
					Port: types.NullInt{IsSet: true},
					Domain: Domain{
						Name: "some-domain",
					},
				},
				actionerror.InvalidHTTPRouteSettings{Domain: "some-domain"},
			),

			Entry("error - hostname on TCP domain",
				Route{
					Host: "some-host",
					Domain: Domain{
						Name:            "some-domain",
						RouterGroupType: constant.TCPRouterGroup,
					},
				},
				actionerror.InvalidTCPRouteSettings{Domain: "some-domain"},
			),

			Entry("error - path on TCP domain",
				Route{
					Path: "some-path",
					Domain: Domain{
						Name:            "some-domain",
						RouterGroupType: constant.TCPRouterGroup,
					},
				},
				actionerror.InvalidTCPRouteSettings{Domain: "some-domain"},
			),
		)

		DescribeTable("ValidateWithRandomPort",
			func(route Route, randomPort bool, expectedErr error) {
				err := route.ValidateWithRandomPort(randomPort)
				if expectedErr == nil {
					Expect(err).To(BeNil())
				} else {
					Expect(err).To(Equal(expectedErr))
				}
			},

			Entry("valid - host and path on HTTP domain",
				Route{
					Host: "some-host",
					Path: "some-path",
					Domain: Domain{
						Name: "some-domain",
					},
				},
				false,
				nil,
			),

			Entry("valid - port on TCP domain",
				Route{
					Port: types.NullInt{IsSet: true},
					Domain: Domain{
						Name:            "some-domain",
						RouterGroupType: constant.TCPRouterGroup,
					},
				},
				false,
				nil,
			),

			Entry("error - port on HTTP domain",
				Route{
					Port: types.NullInt{IsSet: true},
					Domain: Domain{
						Name: "some-domain",
					},
				},
				false,
				actionerror.InvalidHTTPRouteSettings{Domain: "some-domain"},
			),

			Entry("error - randomport on HTTP domain",
				Route{
					Port: types.NullInt{IsSet: false},
					Domain: Domain{
						Name: "some-domain",
					},
				},
				true,
				actionerror.InvalidHTTPRouteSettings{Domain: "some-domain"},
			),

			Entry("error - hostname on TCP domain",
				Route{
					Host: "some-host",
					Domain: Domain{
						Name:            "some-domain",
						RouterGroupType: constant.TCPRouterGroup,
					},
				},
				true,
				actionerror.InvalidTCPRouteSettings{Domain: "some-domain"},
			),

			Entry("error - path on TCP domain",
				Route{
					Path: "some-path",
					Domain: Domain{
						Name:            "some-domain",
						RouterGroupType: constant.TCPRouterGroup,
					},
				},
				true,
				actionerror.InvalidTCPRouteSettings{Domain: "some-domain"},
			),
		)
	})

	Describe("MapRouteToApplication", func() {
		Context("when no errors are encountered", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.UpdateRouteApplicationReturns(
					ccv2.Route{},
					ccv2.Warnings{"map warning"},
					nil)
			})

			It("maps the route to the application and returns all warnings", func() {
				warnings, err := actor.MapRouteToApplication("some-route-guid", "some-app-guid")
				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("map warning"))

				Expect(fakeCloudControllerClient.UpdateRouteApplicationCallCount()).To(Equal(1))
				routeGUID, appGUID := fakeCloudControllerClient.UpdateRouteApplicationArgsForCall(0)
				Expect(routeGUID).To(Equal("some-route-guid"))
				Expect(appGUID).To(Equal("some-app-guid"))
			})
		})

		Context("when an error is encountered", func() {
			Context("InvalidRelationError", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.UpdateRouteApplicationReturns(
						ccv2.Route{},
						ccv2.Warnings{"map warning"},
						ccerror.InvalidRelationError{})
				})

				It("returns the error", func() {
					warnings, err := actor.MapRouteToApplication("some-route-guid", "some-app-guid")
					Expect(err).To(MatchError(actionerror.RouteInDifferentSpaceError{}))
					Expect(warnings).To(ConsistOf("map warning"))
				})
			})

			Context("generic error", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("map route failed")
					fakeCloudControllerClient.UpdateRouteApplicationReturns(
						ccv2.Route{},
						ccv2.Warnings{"map warning"},
						expectedErr)
				})

				It("returns the error", func() {
					warnings, err := actor.MapRouteToApplication("some-route-guid", "some-app-guid")
					Expect(err).To(MatchError(expectedErr))
					Expect(warnings).To(ConsistOf("map warning"))
				})
			})
		})
	})

	Describe("UnmapRouteFromApplication", func() {
		Context("when no errors are encountered", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.DeleteRouteApplicationReturns(
					ccv2.Warnings{"map warning"},
					nil)
			})

			It("unmaps the route from the application and returns all warnings", func() {
				warnings, err := actor.UnmapRouteFromApplication("some-route-guid", "some-app-guid")
				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("map warning"))

				Expect(fakeCloudControllerClient.DeleteRouteApplicationCallCount()).To(Equal(1))
				routeGUID, appGUID := fakeCloudControllerClient.DeleteRouteApplicationArgsForCall(0)
				Expect(routeGUID).To(Equal("some-route-guid"))
				Expect(appGUID).To(Equal("some-app-guid"))
			})
		})

		Context("when an error is encountered", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("map route failed")
				fakeCloudControllerClient.DeleteRouteApplicationReturns(
					ccv2.Warnings{"map warning"},
					expectedErr)
			})

			It("returns the error", func() {
				warnings, err := actor.UnmapRouteFromApplication("some-route-guid", "some-app-guid")
				Expect(err).To(MatchError(expectedErr))
				Expect(warnings).To(ConsistOf("map warning"))
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
						Port:       types.NullInt{IsSet: true, Value: 3333},
						DomainGUID: "some-domain-guid",
						SpaceGUID:  "some-space-guid",
					},
					ccv2.Warnings{"create route warning"},
					nil)
			})

			It("creates the route and returns all warnings", func() {
				route, warnings, err := actor.CreateRoute(Route{
					Domain: Domain{
						Name: "some-domain",
						GUID: "some-domain-guid",
					},
					Host:      "some-host",
					Path:      "/some-path",
					Port:      types.NullInt{IsSet: true, Value: 3333},
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
					Port:      types.NullInt{IsSet: true, Value: 3333},
					SpaceGUID: "some-space-guid",
				}))

				Expect(fakeCloudControllerClient.CreateRouteCallCount()).To(Equal(1))
				passedRoute, generatePort := fakeCloudControllerClient.CreateRouteArgsForCall(0)
				Expect(passedRoute).To(Equal(ccv2.Route{
					DomainGUID: "some-domain-guid",
					Host:       "some-host",
					Path:       "/some-path",
					Port:       types.NullInt{IsSet: true, Value: 3333},
					SpaceGUID:  "some-space-guid",
				}))
				Expect(generatePort).To(BeTrue())
			})
		})

		Context("when path does not start with /", func() {
			It("prepends / to path", func() {
				_, _, err := actor.CreateRoute(
					Route{
						Domain: Domain{
							Name: "some-domain",
							GUID: "some-domain-guid",
						},
						Host:      "some-host",
						Path:      "some-path",
						Port:      types.NullInt{IsSet: true, Value: 3333},
						SpaceGUID: "some-space-guid",
					},
					true,
				)
				Expect(err).ToNot(HaveOccurred())
				Expect(fakeCloudControllerClient.CreateRouteCallCount()).To(Equal(1))
				passedRoute, _ := fakeCloudControllerClient.CreateRouteArgsForCall(0)
				Expect(passedRoute).To(Equal(ccv2.Route{
					DomainGUID: "some-domain-guid",
					Host:       "some-host",
					Path:       "/some-path",
					Port:       types.NullInt{IsSet: true, Value: 3333},
					SpaceGUID:  "some-space-guid",
				}))
			})
		})

		Context("when is not provided", func() {
			It("passes empty path", func() {
				_, _, err := actor.CreateRoute(
					Route{
						Domain: Domain{
							Name: "some-domain",
							GUID: "some-domain-guid",
						},
						Host:      "some-host",
						Port:      types.NullInt{IsSet: true, Value: 3333},
						SpaceGUID: "some-space-guid",
					},
					true,
				)
				Expect(err).ToNot(HaveOccurred())
				Expect(fakeCloudControllerClient.CreateRouteCallCount()).To(Equal(1))
				passedRoute, _ := fakeCloudControllerClient.CreateRouteArgsForCall(0)
				Expect(passedRoute).To(Equal(ccv2.Route{
					DomainGUID: "some-domain-guid",
					Host:       "some-host",
					Port:       types.NullInt{IsSet: true, Value: 3333},
					SpaceGUID:  "some-space-guid",
				}))
			})
		})

		Context("when an error is encountered", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("map route failed")
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

	Describe("CreateRouteWithExistenceCheck", func() {
		var (
			route               Route
			generatePort        bool
			createdRoute        Route
			createRouteWarnings Warnings
			createRouteErr      error
		)

		BeforeEach(func() {
			fakeCloudControllerClient.GetSpacesReturns(
				[]ccv2.Space{
					{
						GUID:                     "some-space-guid",
						Name:                     "some-space",
						AllowSSH:                 true,
						SpaceQuotaDefinitionGUID: "some-space-quota-guid",
					},
				},
				ccv2.Warnings{"get-space-warning"},
				nil)

			fakeCloudControllerClient.CreateRouteReturns(
				ccv2.Route{
					GUID:       "some-route-guid",
					Host:       "some-host",
					Path:       "some-path",
					DomainGUID: "some-domain-guid",
					SpaceGUID:  "some-space-guid",
				},
				ccv2.Warnings{"create-route-warning"},
				nil)

			fakeCloudControllerClient.GetRoutesReturns([]ccv2.Route{}, ccv2.Warnings{"get-routes-warning"}, nil)

			route = Route{
				Domain: Domain{
					Name: "some-domain",
					GUID: "some-domain-guid",
				},
				Host: "some-host",
				Path: "some-path",
			}
			generatePort = false
		})

		JustBeforeEach(func() {
			createdRoute, createRouteWarnings, createRouteErr = actor.CreateRouteWithExistenceCheck(
				"some-org-guid",
				"some-space",
				route,
				generatePort)
		})

		Context("when route does not exist", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetRoutesReturns([]ccv2.Route{}, ccv2.Warnings{"get-routes-warning"}, nil)
			})

			It("creates the route and returns all warnings", func() {
				Expect(createRouteErr).ToNot(HaveOccurred())
				Expect(createRouteWarnings).To(ConsistOf(
					"get-space-warning",
					"get-routes-warning",
					"create-route-warning",
				))

				Expect(createdRoute).To(Equal(Route{
					Domain: Domain{
						Name: "some-domain",
						GUID: "some-domain-guid",
					},
					GUID:      "some-route-guid",
					Host:      "some-host",
					Path:      "some-path",
					SpaceGUID: "some-space-guid",
				}))

				Expect(fakeCloudControllerClient.GetSpacesCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetSpacesArgsForCall(0)).To(ConsistOf(
					ccv2.QQuery{
						Filter:   ccv2.OrganizationGUIDFilter,
						Operator: ccv2.EqualOperator,
						Values:   []string{"some-org-guid"},
					},
					ccv2.QQuery{
						Filter:   ccv2.NameFilter,
						Operator: ccv2.EqualOperator,
						Values:   []string{"some-space"},
					}))

				Expect(fakeCloudControllerClient.CreateRouteCallCount()).To(Equal(1))
				passedRoute, passedGeneratePort := fakeCloudControllerClient.CreateRouteArgsForCall(0)
				Expect(passedRoute).To(Equal(ccv2.Route{
					DomainGUID: "some-domain-guid",
					Host:       "some-host",
					Path:       "/some-path",
					SpaceGUID:  "some-space-guid",
				}))
				Expect(passedGeneratePort).To(BeFalse())
			})

			Context("when creating route errors", func() {
				var expectedErr error

				BeforeEach(func() {
					expectedErr = errors.New("map route failed")
					fakeCloudControllerClient.CreateRouteReturns(
						ccv2.Route{},
						ccv2.Warnings{"create-route-warning"},
						expectedErr)
				})

				It("returns the error and warnings", func() {
					Expect(createRouteErr).To(MatchError(expectedErr))
					Expect(createRouteWarnings).To(ConsistOf("get-space-warning", "get-routes-warning", "create-route-warning"))
				})
			})
		})

		Context("when route already exists", func() {
			var foundRoute ccv2.Route

			BeforeEach(func() {
				foundRoute = ccv2.Route{
					DomainGUID: "some-domain-guid",
					Host:       "some-host",
					Path:       "some-path",
					Port:       types.NullInt{IsSet: true, Value: 3333},
					SpaceGUID:  "some-space-guid",
				}
				fakeCloudControllerClient.GetRoutesReturns(
					[]ccv2.Route{foundRoute},
					ccv2.Warnings{"get-routes-warning"},
					nil)
				fakeCloudControllerClient.GetSharedDomainReturns(
					ccv2.Domain{Name: "some-domain", GUID: "some-domain-guid"},
					ccv2.Warnings{"get-domain-warning"},
					nil)
			})

			It("returns the error and warnings", func() {
				routeString := CCToActorRoute(foundRoute, Domain{Name: "some-domain", GUID: "some-domain-guid"}).String()
				Expect(createRouteErr).To(MatchError(actionerror.RouteAlreadyExistsError{Route: routeString}))
				Expect(createRouteWarnings).To(ConsistOf("get-space-warning", "get-routes-warning"))
			})
		})

		Context("when looking up the domain GUID", func() {
			BeforeEach(func() {
				route = Route{
					Domain: Domain{
						Name: "some-domain",
					},
				}
			})

			Context("when the domain exists", func() {
				Context("when the domain is an HTTP domain", func() {
					BeforeEach(func() {
						route.Host = "some-host"
						route.Path = "some-path"

						fakeCloudControllerClient.GetSharedDomainsReturns(
							[]ccv2.Domain{},
							ccv2.Warnings{"get-shared-domains-warning"},
							nil,
						)
						fakeCloudControllerClient.GetOrganizationPrivateDomainsReturns(
							[]ccv2.Domain{{Name: "some-domain", GUID: "some-requested-domain-guid"}},
							ccv2.Warnings{"get-private-domains-warning"},
							nil,
						)
					})

					It("gets domain and finds route with fully instantiated domain", func() {
						Expect(createRouteErr).ToNot(HaveOccurred())
						Expect(createRouteWarnings).To(ConsistOf(
							"get-space-warning",
							"get-shared-domains-warning",
							"get-private-domains-warning",
							"get-routes-warning",
							"create-route-warning",
						))
						Expect(createdRoute).To(Equal(Route{
							Domain: Domain{
								Name: "some-domain",
								GUID: "some-requested-domain-guid",
							},
							GUID:      "some-route-guid",
							Host:      "some-host",
							Path:      "some-path",
							SpaceGUID: "some-space-guid",
						}))
						Expect(fakeCloudControllerClient.GetSharedDomainsCallCount()).To(Equal(1))
						Expect(fakeCloudControllerClient.GetOrganizationPrivateDomainsCallCount()).To(Equal(1))
						orgGUID, queries := fakeCloudControllerClient.GetOrganizationPrivateDomainsArgsForCall(0)
						Expect(orgGUID).To(Equal("some-org-guid"))
						Expect(queries).To(HaveLen(1))
						Expect(queries[0]).To(Equal(ccv2.QQuery{
							Filter:   ccv2.NameFilter,
							Operator: ccv2.InOperator,
							Values:   []string{"some-domain"},
						}))
					})
				})

				Context("when the domain is a TCP domain", func() {
					BeforeEach(func() {
						fakeCloudControllerClient.GetSharedDomainsReturns(
							[]ccv2.Domain{{
								Name:            "some-tcp-domain",
								GUID:            "some-requested-domain-guid",
								RouterGroupType: constant.TCPRouterGroup,
							}},
							ccv2.Warnings{"get-shared-domains-warning"},
							nil,
						)
					})

					Context("when specifying a port", func() {
						BeforeEach(func() {
							route.Port = types.NullInt{IsSet: true, Value: 1234}
							fakeCloudControllerClient.CreateRouteReturns(
								ccv2.Route{
									GUID:       "some-route-guid",
									DomainGUID: "some-domain-guid",
									Port:       types.NullInt{IsSet: true, Value: 1234},
									SpaceGUID:  "some-space-guid",
								},
								ccv2.Warnings{"create-route-warning"},
								nil)
						})

						It("gets domain and finds route with fully instantiated domain", func() {
							Expect(createRouteErr).ToNot(HaveOccurred())
							Expect(createRouteWarnings).To(ConsistOf(
								"get-space-warning",
								"get-shared-domains-warning",
								"get-routes-warning",
								"create-route-warning",
							))
							Expect(createdRoute).To(Equal(Route{
								Domain: Domain{
									Name:            "some-domain",
									GUID:            "some-requested-domain-guid",
									RouterGroupType: constant.TCPRouterGroup,
								},
								GUID:      "some-route-guid",
								Port:      types.NullInt{IsSet: true, Value: 1234},
								SpaceGUID: "some-space-guid",
							}))
							Expect(fakeCloudControllerClient.GetSharedDomainsCallCount()).To(Equal(1))
							Expect(fakeCloudControllerClient.GetOrganizationPrivateDomainsCallCount()).To(Equal(1))
							orgGUID, queries := fakeCloudControllerClient.GetOrganizationPrivateDomainsArgsForCall(0)
							Expect(orgGUID).To(Equal("some-org-guid"))
							Expect(queries).To(HaveLen(1))
							Expect(queries[0]).To(Equal(ccv2.QQuery{
								Filter:   ccv2.NameFilter,
								Operator: ccv2.InOperator,
								Values:   []string{"some-domain"},
							}))

							Expect(fakeCloudControllerClient.GetRoutesCallCount()).To(Equal(1))
						})
					})

					Context("when generating a random port", func() {
						BeforeEach(func() {
							generatePort = true
							fakeCloudControllerClient.CreateRouteReturns(
								ccv2.Route{
									GUID:       "some-route-guid",
									DomainGUID: "some-domain-guid",
									Port:       types.NullInt{IsSet: true, Value: 1234},
									SpaceGUID:  "some-space-guid",
								},
								ccv2.Warnings{"create-route-warning"},
								nil)
						})

						It("creates a route with a generated port, and doesn't check for existence", func() {
							Expect(createRouteErr).ToNot(HaveOccurred())
							Expect(createRouteWarnings).To(ConsistOf(
								"get-space-warning",
								"get-shared-domains-warning",
								"create-route-warning",
							))
							Expect(createdRoute).To(Equal(Route{
								Domain: Domain{
									Name:            "some-domain",
									GUID:            "some-requested-domain-guid",
									RouterGroupType: constant.TCPRouterGroup,
								},
								GUID:      "some-route-guid",
								Port:      types.NullInt{IsSet: true, Value: 1234},
								SpaceGUID: "some-space-guid",
							}))
							Expect(fakeCloudControllerClient.GetSharedDomainsCallCount()).To(Equal(1))
							Expect(fakeCloudControllerClient.GetOrganizationPrivateDomainsCallCount()).To(Equal(1))
							orgGUID, queries := fakeCloudControllerClient.GetOrganizationPrivateDomainsArgsForCall(0)
							Expect(orgGUID).To(Equal("some-org-guid"))
							Expect(queries).To(HaveLen(1))
							Expect(queries[0]).To(Equal(ccv2.QQuery{
								Filter:   ccv2.NameFilter,
								Operator: ccv2.InOperator,
								Values:   []string{"some-domain"},
							}))

							Expect(fakeCloudControllerClient.GetRoutesCallCount()).To(Equal(0))
						})
					})

					Context("when no port options are provided", func() {
						BeforeEach(func() {
							generatePort = false
							route.Port.IsSet = false
						})

						It("returns a TCPRouteOptionsNotProvidedError", func() {
							Expect(createRouteErr).To(MatchError(actionerror.TCPRouteOptionsNotProvidedError{}))
						})
					})
				})
			})

			Context("when the domain doesn't exist", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetSharedDomainsReturns(
						[]ccv2.Domain{},
						ccv2.Warnings{"get-shared-domains-warning"},
						nil,
					)
					fakeCloudControllerClient.GetOrganizationPrivateDomainsReturns(
						[]ccv2.Domain{},
						ccv2.Warnings{"get-private-domains-warning"},
						nil,
					)
				})

				It("returns all warnings and domain not found err", func() {
					Expect(createRouteErr).To(Equal(actionerror.DomainNotFoundError{Name: "some-domain"}))
					Expect(createRouteWarnings).To(ConsistOf(
						"get-space-warning",
						"get-shared-domains-warning",
						"get-private-domains-warning",
					))
				})
			})
		})

		Context("when the requested route is invalid", func() {
			BeforeEach(func() {
				generatePort = true
			})

			It("returns a validation error", func() {
				Expect(createRouteErr).To(MatchError(actionerror.InvalidHTTPRouteSettings{Domain: route.Domain.Name}))
				Expect(createRouteWarnings).To(ConsistOf("get-space-warning"))
			})
		})

		Context("when getting space errors", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("map route failed")
				fakeCloudControllerClient.GetSpacesReturns(
					[]ccv2.Space{},
					ccv2.Warnings{"get-space-warning"},
					expectedErr)
			})

			It("returns the error and warnings", func() {
				Expect(createRouteErr).To(MatchError(expectedErr))
				Expect(createRouteWarnings).To(ConsistOf("get-space-warning"))
			})
		})

		Context("when getting routes errors", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("map route failed")
				fakeCloudControllerClient.GetRoutesReturns([]ccv2.Route{}, ccv2.Warnings{"get-routes-warning"}, expectedErr)
			})

			It("returns the error and warnings", func() {
				Expect(createRouteErr).To(MatchError(expectedErr))
				Expect(createRouteWarnings).To(ConsistOf("get-space-warning", "get-routes-warning"))
			})
		})
	})

	Describe("GetOrphanedRoutesBySpace", func() {
		BeforeEach(func() {
			fakeCloudControllerClient.GetRouteApplicationsStub = func(routeGUID string, queries ...ccv2.QQuery) ([]ccv2.Application, ccv2.Warnings, error) {
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
			var expectedErr actionerror.OrphanedRoutesNotFoundError

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
						Port:       types.NullInt{IsSet: true, Value: 1234},
						DomainGUID: "domain-1-guid",
					},
					ccv2.Route{
						GUID:       "route-guid-2",
						SpaceGUID:  "some-space-guid",
						Host:       "host",
						Path:       "/path",
						Port:       types.NullInt{IsSet: true, Value: 1234},
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
						Port:      types.NullInt{IsSet: true, Value: 1234},
						SpaceGUID: "some-space-guid",
					},
					{
						Domain: Domain{
							Name: "other-domain.com",
						},
						GUID:      "route-guid-2",
						Host:      "host",
						Path:      "/path",
						Port:      types.NullInt{IsSet: true, Value: 1234},
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
							Port:       types.NullInt{IsSet: true, Value: 1234},
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
						Port:       types.NullInt{IsSet: true, Value: 1234},
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
						Port:       types.NullInt{IsSet: true, Value: 1234},
						DomainGUID: "domain-1-guid",
					},
					ccv2.Route{
						GUID:       "route-guid-2",
						SpaceGUID:  "some-space-guid",
						Host:       "host",
						Path:       "/path",
						Port:       types.NullInt{IsSet: true, Value: 1234},
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
						Port:      types.NullInt{IsSet: true, Value: 1234},
						SpaceGUID: "some-space-guid",
					},
					{
						Domain: Domain{
							Name: "other-domain.com",
						},
						GUID:      "route-guid-2",
						Host:      "host",
						Path:      "/path",
						Port:      types.NullInt{IsSet: true, Value: 1234},
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
							Port:       types.NullInt{IsSet: true, Value: 1234},
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
						Port:       types.NullInt{IsSet: true, Value: 1234},
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

	Describe("GetRouteByComponents", func() {
		var (
			domain     Domain
			inputRoute Route

			route      Route
			warnings   Warnings
			executeErr error
		)

		BeforeEach(func() {
			domain = Domain{
				GUID: "some-domain-guid",
				Name: "domain.com",
			}

			inputRoute = Route{
				Domain: domain,
			}
		})

		JustBeforeEach(func() {
			route, warnings, executeErr = actor.GetRouteByComponents(inputRoute)
		})

		Context("validation", func() {
			Context("when the route's domain is a TCP domain", func() {
				BeforeEach(func() {
					inputRoute.Domain.RouterGroupType = constant.TCPRouterGroup
				})

				Context("when a port isn't provided for the query", func() {
					BeforeEach(func() {
						inputRoute.Port.IsSet = false
					})

					It("returns a PortNotProvidedForQueryError", func() {
						Expect(executeErr).To(MatchError(actionerror.PortNotProvidedForQueryError{}))
					})
				})
			})

			Context("when the route's domain is an HTTP shared domain", func() {
				BeforeEach(func() {
					inputRoute.Domain.RouterGroupType = constant.HTTPRouterGroup
					inputRoute.Domain.Type = constant.SharedDomain
				})

				Context("when a host is not provided", func() {
					BeforeEach(func() {
						inputRoute.Host = ""
					})

					It("returns a NoHostnameAndSharedDomainError", func() {
						Expect(executeErr).To(MatchError(actionerror.NoHostnameAndSharedDomainError{}))
					})
				})
			})
		})

		Context("when finding the route is successful and returns one route", func() {
			Context("when hostname and path aren't provided", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetRoutesReturns([]ccv2.Route{
						{
							GUID:       "route-guid-1",
							SpaceGUID:  "some-space-guid",
							DomainGUID: domain.GUID,
						},
					}, ccv2.Warnings{"get-routes-warning"}, nil)
				})

				It("explicitly queries for empty hostname and path", func() {
					Expect(warnings).To(ConsistOf("get-routes-warning"))
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(route).To(Equal(Route{
						Domain:    domain,
						GUID:      "route-guid-1",
						Host:      inputRoute.Host,
						SpaceGUID: "some-space-guid",
					}))

					Expect(fakeCloudControllerClient.GetRoutesCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetRoutesArgsForCall(0)).To(Equal([]ccv2.QQuery{
						{
							Filter:   ccv2.DomainGUIDFilter,
							Operator: ccv2.EqualOperator,
							Values:   []string{domain.GUID},
						},
						{
							Filter:   ccv2.HostFilter,
							Operator: ccv2.EqualOperator,
							Values:   []string{""},
						},
						{
							Filter:   ccv2.PathFilter,
							Operator: ccv2.EqualOperator,
							Values:   []string{""},
						},
					}))
				})
			})

			Context("when the hostname is provided", func() {
				BeforeEach(func() {
					inputRoute.Host = "some-host"

					fakeCloudControllerClient.GetRoutesReturns([]ccv2.Route{
						{
							GUID:       "route-guid-1",
							SpaceGUID:  "some-space-guid",
							Host:       inputRoute.Host,
							DomainGUID: domain.GUID,
						},
					}, ccv2.Warnings{"get-routes-warning"}, nil)
				})

				It("returns the route and any warnings", func() {
					Expect(warnings).To(ConsistOf("get-routes-warning"))
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(route).To(Equal(Route{
						Domain:    domain,
						GUID:      "route-guid-1",
						Host:      inputRoute.Host,
						SpaceGUID: "some-space-guid",
					}))

					Expect(fakeCloudControllerClient.GetRoutesCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetRoutesArgsForCall(0)).To(Equal([]ccv2.QQuery{
						{
							Filter:   ccv2.DomainGUIDFilter,
							Operator: ccv2.EqualOperator,
							Values:   []string{domain.GUID},
						},
						{
							Filter:   ccv2.HostFilter,
							Operator: ccv2.EqualOperator,
							Values:   []string{inputRoute.Host},
						},
						{
							Filter:   ccv2.PathFilter,
							Operator: ccv2.EqualOperator,
							Values:   []string{""},
						},
					}))
				})
			})

			Context("when the path is provided", func() {
				BeforeEach(func() {
					inputRoute.Path = "/some-path"

					fakeCloudControllerClient.GetRoutesReturns([]ccv2.Route{
						{
							GUID:       "route-guid-1",
							SpaceGUID:  "some-space-guid",
							Path:       inputRoute.Path,
							DomainGUID: domain.GUID,
						},
					}, ccv2.Warnings{"get-routes-warning"}, nil)
				})

				It("returns the routes and any warnings", func() {
					Expect(warnings).To(ConsistOf("get-routes-warning"))
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(route).To(Equal(Route{
						Domain:    domain,
						GUID:      "route-guid-1",
						Path:      inputRoute.Path,
						SpaceGUID: "some-space-guid",
					}))

					Expect(fakeCloudControllerClient.GetRoutesCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetRoutesArgsForCall(0)).To(Equal([]ccv2.QQuery{
						{
							Filter:   ccv2.DomainGUIDFilter,
							Operator: ccv2.EqualOperator,
							Values:   []string{domain.GUID},
						},
						{
							Filter:   ccv2.HostFilter,
							Operator: ccv2.EqualOperator,
							Values:   []string{inputRoute.Host},
						},
						{
							Filter:   ccv2.PathFilter,
							Operator: ccv2.EqualOperator,
							Values:   []string{inputRoute.Path},
						},
					}))
				})
			})

			Context("when the port is provided", func() {
				BeforeEach(func() {
					inputRoute.Port = types.NullInt{Value: 1234, IsSet: true}

					fakeCloudControllerClient.GetRoutesReturns([]ccv2.Route{
						{
							GUID:       "route-guid-1",
							SpaceGUID:  "some-space-guid",
							Port:       inputRoute.Port,
							DomainGUID: domain.GUID,
						},
					}, ccv2.Warnings{"get-routes-warning"}, nil)
				})

				It("returns the routes and any warnings", func() {
					Expect(warnings).To(ConsistOf("get-routes-warning"))
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(route).To(Equal(Route{
						Domain:    domain,
						GUID:      "route-guid-1",
						Port:      inputRoute.Port,
						SpaceGUID: "some-space-guid",
					}))

					Expect(fakeCloudControllerClient.GetRoutesCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetRoutesArgsForCall(0)).To(Equal([]ccv2.QQuery{
						{
							Filter:   ccv2.DomainGUIDFilter,
							Operator: ccv2.EqualOperator,
							Values:   []string{domain.GUID},
						},
						{
							Filter:   ccv2.HostFilter,
							Operator: ccv2.EqualOperator,
							Values:   []string{inputRoute.Host},
						},
						{
							Filter:   ccv2.PathFilter,
							Operator: ccv2.EqualOperator,
							Values:   []string{inputRoute.Path},
						},
						{
							Filter:   ccv2.PortFilter,
							Operator: ccv2.EqualOperator,
							Values:   []string{fmt.Sprint(inputRoute.Port.Value)},
						},
					}))
				})
			})

			Context("when all parts of the route are provided", func() {
				BeforeEach(func() {
					inputRoute.Host = "some-host"
					inputRoute.Path = "/some-path"
					inputRoute.Port = types.NullInt{Value: 1234, IsSet: true}

					fakeCloudControllerClient.GetRoutesReturns([]ccv2.Route{{
						DomainGUID: domain.GUID,
						GUID:       "route-guid-1",
						Host:       inputRoute.Host,
						Path:       inputRoute.Path,
						Port:       inputRoute.Port,
						SpaceGUID:  "some-space-guid",
					}}, ccv2.Warnings{"get-routes-warning"}, nil)
				})

				It("returns the routes and any warnings", func() {
					Expect(warnings).To(ConsistOf("get-routes-warning"))
					Expect(executeErr).NotTo(HaveOccurred())
					Expect(route).To(Equal(Route{
						Domain:    domain,
						GUID:      "route-guid-1",
						Host:      inputRoute.Host,
						Path:      inputRoute.Path,
						Port:      inputRoute.Port,
						SpaceGUID: "some-space-guid",
					}))

					Expect(fakeCloudControllerClient.GetRoutesCallCount()).To(Equal(1))
					Expect(fakeCloudControllerClient.GetRoutesArgsForCall(0)).To(Equal([]ccv2.QQuery{
						{
							Filter:   ccv2.DomainGUIDFilter,
							Operator: ccv2.EqualOperator,
							Values:   []string{domain.GUID},
						},
						{
							Filter:   ccv2.HostFilter,
							Operator: ccv2.EqualOperator,
							Values:   []string{inputRoute.Host},
						},
						{
							Filter:   ccv2.PathFilter,
							Operator: ccv2.EqualOperator,
							Values:   []string{inputRoute.Path},
						},
						{
							Filter:   ccv2.PortFilter,
							Operator: ccv2.EqualOperator,
							Values:   []string{fmt.Sprint(inputRoute.Port.Value)},
						},
					}))
				})
			})
		})

		Context("when finding the route is successful and returns no routes", func() {
			BeforeEach(func() {
				inputRoute.Host = "some-host"
				inputRoute.Path = "/some-path"
				inputRoute.Port = types.NullInt{Value: 1234, IsSet: true}
				fakeCloudControllerClient.GetRoutesReturns([]ccv2.Route{}, ccv2.Warnings{"get-routes-warning"}, nil)
			})

			It("returns a RouteNotFoundError and warnings", func() {
				Expect(executeErr).To(MatchError(actionerror.RouteNotFoundError{
					Host:       inputRoute.Host,
					Path:       inputRoute.Path,
					Port:       inputRoute.Port.Value,
					DomainGUID: domain.GUID,
				}))
				Expect(warnings).To(ConsistOf("get-routes-warning"))
			})
		})

		Context("when finding the route returns an error", func() {
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
					Port: types.NullInt{IsSet: true, Value: 42},
				})

				Expect(err).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("some-check-route-warnings"))
				Expect(exists).To(BeTrue())

				Expect(fakeCloudControllerClient.CheckRouteCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.CheckRouteArgsForCall(0)).To(Equal(ccv2.Route{
					Host:       "some-host",
					DomainGUID: "some-domain-guid",
					Path:       "some-path",
					Port:       types.NullInt{IsSet: true, Value: 42},
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
				Path:      "some-path",
				SpaceGUID: "some-space-guid",
			}
		})

		JustBeforeEach(func() {
			returnedRoute, warnings, executeErr = actor.FindRouteBoundToSpaceWithSettings(route)
		})

		Context("when the route exists in the current space", func() {
			var existingRoute Route

			Context("when the route uses an HTTP domain", func() {
				BeforeEach(func() {
					existingRoute = route
					existingRoute.GUID = "some-route-guid"
					fakeCloudControllerClient.GetRoutesReturns([]ccv2.Route{ActorToCCRoute(existingRoute)}, ccv2.Warnings{"get route warning"}, nil)
				})

				It("returns the route", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(returnedRoute).To(Equal(existingRoute))
					Expect(warnings).To(ConsistOf("get route warning"))
				})
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
					Expect(executeErr).To(MatchError(actionerror.RouteInDifferentSpaceError{Route: route.String()}))
					Expect(warnings).To(ConsistOf("get route warning"))
				})
			})

			Context("when the user does not have access to the route", func() {
				BeforeEach(func() {
					fakeCloudControllerClient.GetRoutesReturns([]ccv2.Route{}, ccv2.Warnings{"get route warning"}, nil)
					fakeCloudControllerClient.CheckRouteReturns(true, ccv2.Warnings{"check route warning"}, nil)
				})

				It("returns a RouteInDifferentSpaceError", func() {
					Expect(executeErr).To(MatchError(actionerror.RouteInDifferentSpaceError{Route: route.String()}))
					Expect(warnings).To(ConsistOf("get route warning", "check route warning"))
				})
			})
		})

		Context("when the route does not exist", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = actionerror.RouteNotFoundError{Host: route.Host, DomainGUID: route.Domain.GUID, Path: route.Path}
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
