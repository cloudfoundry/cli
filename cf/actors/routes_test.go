package actors_test

import (
	"errors"

	. "github.com/cloudfoundry/cli/cf/actors"
	"github.com/cloudfoundry/cli/cf/api/apifakes"
	cferrors "github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/errors/errorsfakes"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/cf/terminal/terminalfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Routes", func() {
	var (
		fakeUI               *terminalfakes.FakeUI
		fakeRouteRepository  *apifakes.FakeRouteRepository
		fakeDomainRepository *apifakes.FakeDomainRepository
		routeActor           RouteActor

		expectedRoute  models.Route
		expectedDomain models.DomainFields
	)

	BeforeEach(func() {
		fakeUI = &terminalfakes.FakeUI{}
		fakeRouteRepository = new(apifakes.FakeRouteRepository)
		fakeDomainRepository = new(apifakes.FakeDomainRepository)
		routeActor = NewRouteActor(fakeUI, fakeRouteRepository, fakeDomainRepository)
	})

	Describe("CreateRandomTCPRoute", func() {
		BeforeEach(func() {
			expectedDomain = models.DomainFields{
				Name: "dies-tcp.com",
			}

			expectedRoute = models.Route{
				GUID: "some-guid",
			}

			fakeRouteRepository.CreateReturns(expectedRoute, nil)
		})

		It("calls Create on the route repo", func() {
			routeActor.CreateRandomTCPRoute(expectedDomain)

			host, d, path, port, randomPort := fakeRouteRepository.CreateArgsForCall(0)
			Expect(host).To(BeEmpty())
			Expect(d).To(Equal(expectedDomain))
			Expect(path).To(BeEmpty())
			Expect(port).To(Equal(0))
			Expect(randomPort).To(BeTrue())
		})

		It("states that a route is being created", func() {
			routeActor.CreateRandomTCPRoute(expectedDomain)

			Expect(fakeUI.SayCallCount()).To(Equal(1))
			Expect(fakeUI.SayArgsForCall(0)).To(ContainSubstring("Creating random route for"))
		})

		It("returns the route retrieved from the repository", func() {
			actualRoute, err := routeActor.CreateRandomTCPRoute(expectedDomain)
			Expect(err).NotTo(HaveOccurred())

			Expect(actualRoute).To(Equal(expectedRoute))
		})

		It("prints an error when creating the route fails", func() {
			expectedError := errors.New("big bad error message")
			fakeRouteRepository.CreateReturns(models.Route{}, expectedError)

			actualRoute, err := routeActor.CreateRandomTCPRoute(expectedDomain)
			Expect(err).To(Equal(expectedError))
			Expect(actualRoute).To(Equal(models.Route{}))
		})
	})

	Describe("FindOrCreateRoute", func() {
		var (
			expectedHostname string
			expectedPath     string
		)

		BeforeEach(func() {
			expectedHostname = "hostname"
			expectedPath = "path"

			expectedDomain = models.DomainFields{
				Name: "foo.com",
			}

			expectedRoute = models.Route{
				Domain: expectedDomain,
				Host:   expectedHostname,
				Path:   expectedPath,
			}
		})

		Context("the route exists", func() {
			BeforeEach(func() {
				fakeRouteRepository.FindReturns(expectedRoute, nil)
			})

			It("does not create a route", func() {
				route, err := routeActor.FindOrCreateRoute(expectedHostname, expectedDomain, expectedPath, 0, false)
				Expect(route).To(Equal(expectedRoute))
				Expect(err).ToNot(HaveOccurred())

				Expect(fakeRouteRepository.CreateCallCount()).To(Equal(0))

				Expect(fakeUI.SayCallCount()).To(Equal(1))
				output, _ := fakeUI.SayArgsForCall(0)
				Expect(output).To(MatchRegexp("Using route.*hostname.foo.com/path"))
			})
		})

		Context("the route does not exist", func() {
			BeforeEach(func() {
				fakeRouteRepository.FindReturns(models.Route{}, cferrors.NewModelNotFoundError("foo", "bar"))
			})

			Context("with a random port", func() {
				var tcpRoute models.Route

				BeforeEach(func() {
					tcpRoute = models.Route{Port: 4}
					fakeRouteRepository.CreateReturns(tcpRoute, nil)
				})

				It("creates a route with a TCP Route", func() {
					route, err := routeActor.FindOrCreateRoute("", expectedDomain, "", 0, true)
					Expect(route).To(Equal(tcpRoute))
					Expect(err).ToNot(HaveOccurred())

					Expect(fakeRouteRepository.CreateCallCount()).To(Equal(1))
					hostname, domain, path, port, randomPort := fakeRouteRepository.CreateArgsForCall(0)
					Expect(hostname).To(BeEmpty())
					Expect(domain).To(Equal(expectedDomain))
					Expect(path).To(BeEmpty())
					Expect(port).To(Equal(0))
					Expect(randomPort).To(BeTrue())

					Expect(fakeUI.SayCallCount()).To(Equal(2))
					output, _ := fakeUI.SayArgsForCall(0)
					Expect(output).To(MatchRegexp("Creating random route for"))
				})
			})

			Context("without a specific port", func() {
				BeforeEach(func() {
					fakeRouteRepository.CreateReturns(expectedRoute, nil)
				})

				It("creates a route ", func() {
					route, err := routeActor.FindOrCreateRoute(expectedHostname, expectedDomain, "", 1337, false)
					Expect(route).To(Equal(expectedRoute))
					Expect(err).ToNot(HaveOccurred())

					Expect(fakeRouteRepository.CreateCallCount()).To(Equal(1))
					hostname, domain, path, port, randomPort := fakeRouteRepository.CreateArgsForCall(0)
					Expect(hostname).To(Equal(expectedHostname))
					Expect(domain).To(Equal(expectedDomain))
					Expect(path).To(Equal(""))
					Expect(port).To(Equal(1337))
					Expect(randomPort).To(BeFalse())

					Expect(fakeUI.SayCallCount()).To(Equal(2))
					output, _ := fakeUI.SayArgsForCall(0)
					Expect(output).To(MatchRegexp("Creating route.*hostname.foo.com:1337"))
				})
			})

			Context("with a path", func() {
				BeforeEach(func() {
					fakeRouteRepository.CreateReturns(expectedRoute, nil)
				})

				It("creates a route ", func() {
					route, err := routeActor.FindOrCreateRoute(expectedHostname, expectedDomain, expectedPath, 0, false)
					Expect(route).To(Equal(expectedRoute))
					Expect(err).ToNot(HaveOccurred())

					Expect(fakeRouteRepository.CreateCallCount()).To(Equal(1))
					hostname, domain, path, port, randomPort := fakeRouteRepository.CreateArgsForCall(0)
					Expect(hostname).To(Equal(expectedHostname))
					Expect(domain).To(Equal(expectedDomain))
					Expect(path).To(Equal(expectedPath))
					Expect(port).To(Equal(0))
					Expect(randomPort).To(BeFalse())

					Expect(fakeUI.SayCallCount()).To(Equal(2))
					output, _ := fakeUI.SayArgsForCall(0)
					Expect(output).To(MatchRegexp("Creating route.*hostname.foo.com/path"))
				})
			})
		})
	})

	Describe("BindRoute", func() {
		var (
			expectedApp models.Application
		)

		BeforeEach(func() {
			expectedRoute = models.Route{
				GUID: "route-guid",
			}
			expectedApp = models.Application{
				ApplicationFields: models.ApplicationFields{
					Name: "app-name",
					GUID: "app-guid",
				},
			}
		})

		Context("when the app has the route", func() {
			BeforeEach(func() {
				routeSummary := models.RouteSummary{
					GUID: expectedRoute.GUID,
				}
				expectedApp.Routes = append(expectedApp.Routes, routeSummary)
			})

			It("does nothing", func() {
				err := routeActor.BindRoute(expectedApp, expectedRoute)
				Expect(err).To(BeNil())

				Expect(fakeRouteRepository.BindCallCount()).To(Equal(0))
			})
		})

		Context("when the app does not have a route", func() {
			It("binds the route", func() {
				err := routeActor.BindRoute(expectedApp, expectedRoute)
				Expect(err).To(BeNil())

				Expect(fakeRouteRepository.BindCallCount()).To(Equal(1))
				routeGUID, appGUID := fakeRouteRepository.BindArgsForCall(0)
				Expect(routeGUID).To(Equal(expectedRoute.GUID))
				Expect(appGUID).To(Equal(expectedApp.ApplicationFields.GUID))

				Expect(fakeUI.SayArgsForCall(0)).To(MatchRegexp("Binding .* to .*app-name"))
				Expect(fakeUI.OkCallCount()).To(Equal(1))
			})

			Context("when the route is already in use", func() {
				var expectedErr *errorsfakes.FakeHTTPError
				BeforeEach(func() {
					expectedErr = new(errorsfakes.FakeHTTPError)
					expectedErr.ErrorCodeReturns(cferrors.InvalidRelation)
					fakeRouteRepository.BindReturns(expectedErr)
				})

				It("outputs the error", func() {
					err := routeActor.BindRoute(expectedApp, expectedRoute)
					Expect(err.Error()).To(MatchRegexp("The route *. is already in use"))
				})
			})
		})
	})

	Describe("UnbindAll", func() {
		var app models.Application

		BeforeEach(func() {
			app = models.Application{
				ApplicationFields: models.ApplicationFields{
					GUID: "my-app-guid",
				},
				Routes: []models.RouteSummary{
					{
						GUID:   "my-route-guid-1",
						Domain: models.DomainFields{Name: "mydomain1.com"},
					},
					{
						GUID:   "my-route-guid-2",
						Domain: models.DomainFields{Name: "mydomain2.com"},
					},
				},
			}
		})

		Context("when unbinding does not work", func() {
			var expectedError error

			BeforeEach(func() {
				expectedError = errors.New("ZOHMYGOD DUN BROKE")
				fakeRouteRepository.UnbindReturns(expectedError)
			})

			It("returns the error immediately", func() {
				err := routeActor.UnbindAll(app)
				Expect(err).To(Equal(expectedError))

				Expect(fakeRouteRepository.UnbindCallCount()).To(Equal(1))
			})
		})

		Context("when unbinding works", func() {
			It("unbinds the route for the app", func() {
				err := routeActor.UnbindAll(app)
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeRouteRepository.UnbindCallCount()).To(Equal(2))

				routeGUID, appGUID := fakeRouteRepository.UnbindArgsForCall(0)
				Expect(routeGUID).To(Equal("my-route-guid-1"))
				Expect(appGUID).To(Equal("my-app-guid"))

				routeGUID, appGUID = fakeRouteRepository.UnbindArgsForCall(1)
				Expect(routeGUID).To(Equal("my-route-guid-2"))
				Expect(appGUID).To(Equal("my-app-guid"))

				Expect(fakeUI.SayCallCount()).To(Equal(2))

				message, _ := fakeUI.SayArgsForCall(0)
				Expect(message).To(ContainSubstring("Removing route"))

				message, _ = fakeUI.SayArgsForCall(1)
				Expect(message).To(ContainSubstring("Removing route"))
			})
		})
	})

	Describe("FindDomain", func() {
		var (
			routeName           string
			hostname            string
			domain              models.DomainFields
			findDomainErr       error
			domainNotFoundError error
		)

		BeforeEach(func() {
			routeName = "my-hostname.my-domain.com"
			domainNotFoundError = cferrors.NewModelNotFoundError("Domain", routeName)
		})

		JustBeforeEach(func() {
			hostname, domain, findDomainErr = routeActor.FindDomain(routeName)
		})

		Context("when the route belongs to a private domain", func() {
			var privateDomain models.DomainFields

			BeforeEach(func() {
				privateDomain = models.DomainFields{
					GUID: "private-domain-guid",
				}
				fakeDomainRepository.FindPrivateByNameReturns(privateDomain, nil)
			})

			It("returns the private domain", func() {
				Expect(findDomainErr).NotTo(HaveOccurred())
				Expect(fakeDomainRepository.FindPrivateByNameCallCount()).To(Equal(1))
				Expect(fakeDomainRepository.FindPrivateByNameArgsForCall(0)).To(Equal("my-hostname.my-domain.com"))
				Expect(hostname).To(Equal(""))
				Expect(domain).To(Equal(privateDomain))
			})
		})

		Context("when the route belongs to a shared domain", func() {
			var (
				sharedDomain models.DomainFields
			)

			BeforeEach(func() {
				sharedDomain = models.DomainFields{
					GUID: "shared-domain-guid",
				}
				fakeDomainRepository.FindPrivateByNameStub = func(name string) (models.DomainFields, error) {
					return models.DomainFields{}, domainNotFoundError
				}
			})

			Context("when the route has no hostname", func() {
				BeforeEach(func() {
					fakeDomainRepository.FindSharedByNameStub = func(name string) (models.DomainFields, error) {
						if name == "my-hostname.my-domain.com" {
							return sharedDomain, nil
						}
						return models.DomainFields{}, domainNotFoundError
					}
				})

				It("returns the shared domain", func() {
					Expect(findDomainErr).NotTo(HaveOccurred())
					Expect(fakeDomainRepository.FindPrivateByNameCallCount()).To(Equal(1))
					Expect(fakeDomainRepository.FindSharedByNameCallCount()).To(Equal(1))
					Expect(fakeDomainRepository.FindSharedByNameArgsForCall(0)).To(Equal("my-hostname.my-domain.com"))
					Expect(hostname).To(Equal(""))
					Expect(domain).To(Equal(sharedDomain))
				})
			})

			Context("when the route has a hostname", func() {
				BeforeEach(func() {
					fakeDomainRepository.FindSharedByNameStub = func(name string) (models.DomainFields, error) {
						if name == "my-domain.com" {
							return sharedDomain, nil
						}
						return models.DomainFields{}, domainNotFoundError
					}
				})

				It("returns the shared domain and hostname", func() {
					Expect(findDomainErr).NotTo(HaveOccurred())
					Expect(fakeDomainRepository.FindPrivateByNameCallCount()).To(Equal(1))
					Expect(fakeDomainRepository.FindSharedByNameCallCount()).To(Equal(2))
					Expect(fakeDomainRepository.FindSharedByNameArgsForCall(0)).To(Equal("my-hostname.my-domain.com"))
					Expect(fakeDomainRepository.FindSharedByNameArgsForCall(1)).To(Equal("my-domain.com"))
					Expect(hostname).To(Equal("my-hostname"))
					Expect(domain).To(Equal(sharedDomain))
				})
			})
		})

		Context("when the route does not belong to any existing domains", func() {
			BeforeEach(func() {
				routeName = "non-existant-domain.com"
				fakeDomainRepository.FindPrivateByNameReturns(models.DomainFields{}, domainNotFoundError)
				fakeDomainRepository.FindSharedByNameReturns(models.DomainFields{}, domainNotFoundError)
			})

			It("returns an error", func() {
				Expect(findDomainErr).To(HaveOccurred())
				Expect(findDomainErr.Error()).To(Equal("The route non-existant-domain.com did not match any existing domains."))
			})
		})
	})

	Describe("FindPath", func() {
		Context("when there is a path", func() {
			It("returns the route without path and the path", func() {
				routeName := "host.domain/long/path"
				route, path := routeActor.FindPath(routeName)
				Expect(route).To(Equal("host.domain"))
				Expect(path).To(Equal("long/path"))
			})
		})

		Context("when there is no path", func() {
			It("returns the route path and the empty string", func() {
				routeName := "host.domain"
				route, path := routeActor.FindPath(routeName)
				Expect(route).To(Equal("host.domain"))
				Expect(path).To(Equal(""))
			})
		})
	})

	Describe("FindPort", func() {
		Context("when there is a port", func() {
			It("returns the route without port and the port", func() {
				routeName := "host.domain:12345"
				route, port, err := routeActor.FindPort(routeName)
				Expect(route).To(Equal("host.domain"))
				Expect(port).To(Equal(12345))
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when there is no port", func() {
			It("returns the route port and invalid port", func() {
				routeName := "host.domain"
				route, port, err := routeActor.FindPort(routeName)
				Expect(route).To(Equal("host.domain"))
				Expect(port).To(Equal(0))
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when there is an invalid port", func() {
			It("returns an error", func() {
				routeName := "host.domain:thisisnotaport"
				_, _, err := routeActor.FindPort(routeName)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("FindAndBindRoute", func() {
		var (
			routeName            string
			findAndBindRouteErr  error
			appParamsFromContext models.AppParams
		)

		BeforeEach(func() {
			appParamsFromContext = models.AppParams{}
		})

		JustBeforeEach(func() {
			appName := "app-name"
			findAndBindRouteErr = routeActor.FindAndBindRoute(
				routeName,
				models.Application{
					ApplicationFields: models.ApplicationFields{
						Name: appName,
						GUID: "app-guid",
					},
				},
				appParamsFromContext,
			)
		})

		Context("when the route is a HTTP route", func() {
			var httpDomain models.DomainFields

			BeforeEach(func() {
				httpDomain = models.DomainFields{
					Name: "domain.com",
					GUID: "domain-guid",
				}
				domainNotFoundError := cferrors.NewModelNotFoundError("Domain", "some-domain.com")

				fakeDomainRepository.FindPrivateByNameReturns(models.DomainFields{}, domainNotFoundError)
				fakeDomainRepository.FindSharedByNameStub = func(name string) (models.DomainFields, error) {
					if name == "domain.com" {
						return httpDomain, nil
					}
					return models.DomainFields{}, domainNotFoundError
				}
			})

			Context("and contains a port", func() {
				BeforeEach(func() {
					routeName = "domain.com:3333"
				})

				It("should return an error", func() {
					Expect(findAndBindRouteErr).To(HaveOccurred())
					Expect(findAndBindRouteErr.Error()).To(Equal("Port not allowed in HTTP route domain.com:3333"))
				})
			})

			Context("and does not contain a port", func() {
				BeforeEach(func() {
					routeName = "host.domain.com"

					fakeRouteRepository.FindReturns(models.Route{}, cferrors.NewModelNotFoundError("Route", "some-route"))
					fakeRouteRepository.CreateReturns(
						models.Route{
							GUID:   "route-guid",
							Domain: httpDomain,
							Path:   "path",
						},
						nil,
					)
					fakeRouteRepository.BindReturns(nil)
				})

				It("creates and binds the route", func() {
					Expect(findAndBindRouteErr).NotTo(HaveOccurred())

					actualDomainName := fakeDomainRepository.FindSharedByNameArgsForCall(1)
					Expect(actualDomainName).To(Equal("domain.com"))

					actualHost, actualDomain, actualPath, actualPort := fakeRouteRepository.FindArgsForCall(0)
					Expect(actualHost).To(Equal("host"))
					Expect(actualDomain).To(Equal(httpDomain))
					Expect(actualPath).To(Equal(""))
					Expect(actualPort).To(Equal(0))

					actualHost, actualDomain, actualPath, actualPort, actualUseRandomPort := fakeRouteRepository.CreateArgsForCall(0)
					Expect(actualHost).To(Equal("host"))
					Expect(actualDomain).To(Equal(httpDomain))
					Expect(actualPath).To(Equal(""))
					Expect(actualPort).To(Equal(0))
					Expect(actualUseRandomPort).To(BeFalse())

					routeGUID, appGUID := fakeRouteRepository.BindArgsForCall(0)
					Expect(routeGUID).To(Equal("route-guid"))
					Expect(appGUID).To(Equal("app-guid"))
				})

				Context("and contains a path", func() {
					BeforeEach(func() {
						routeName = "host.domain.com/path"
					})

					It("creates and binds the route", func() {
						Expect(findAndBindRouteErr).NotTo(HaveOccurred())

						actualDomainName := fakeDomainRepository.FindSharedByNameArgsForCall(1)
						Expect(actualDomainName).To(Equal("domain.com"))

						actualHost, actualDomain, actualPath, actualPort := fakeRouteRepository.FindArgsForCall(0)
						Expect(actualHost).To(Equal("host"))
						Expect(actualDomain).To(Equal(httpDomain))
						Expect(actualPath).To(Equal("path"))
						Expect(actualPort).To(Equal(0))

						actualHost, actualDomain, actualPath, actualPort, actualUseRandomPort := fakeRouteRepository.CreateArgsForCall(0)
						Expect(actualHost).To(Equal("host"))
						Expect(actualDomain).To(Equal(httpDomain))
						Expect(actualPath).To(Equal("path"))
						Expect(actualPort).To(Equal(0))
						Expect(actualUseRandomPort).To(BeFalse())

						routeGUID, appGUID := fakeRouteRepository.BindArgsForCall(0)
						Expect(routeGUID).To(Equal("route-guid"))
						Expect(appGUID).To(Equal("app-guid"))
					})
				})
			})

			Context("and the --hostname flag is provided", func() {
				BeforeEach(func() {
					appParamsFromContext = models.AppParams{
						Hosts: []string{"flag-hostname"},
					}
				})

				Context("and the route contains a hostname", func() {
					BeforeEach(func() {
						routeName = "host.domain.com/path"
					})

					It("should replace only the hostname", func() {
						Expect(findAndBindRouteErr).NotTo(HaveOccurred())

						actualHost, actualDomain, actualPath, actualPort := fakeRouteRepository.FindArgsForCall(0)
						Expect(actualHost).To(Equal("flag-hostname"))
						Expect(actualDomain).To(Equal(httpDomain))
						Expect(actualPath).To(Equal("path"))
						Expect(actualPort).To(Equal(0))
					})
				})

				Context("and the route does not contain a hostname", func() {
					BeforeEach(func() {
						routeName = "domain.com"
					})

					It("should set only the hostname", func() {
						Expect(findAndBindRouteErr).NotTo(HaveOccurred())

						actualHost, actualDomain, actualPath, actualPort := fakeRouteRepository.FindArgsForCall(0)
						Expect(actualHost).To(Equal("flag-hostname"))
						Expect(actualDomain).To(Equal(httpDomain))
						Expect(actualPath).To(Equal(""))
						Expect(actualPort).To(Equal(0))
					})
				})
			})
		})

		Context("when the route is a TCP route", func() {
			var tcpDomain models.DomainFields

			BeforeEach(func() {
				tcpDomain = models.DomainFields{
					Name:            "tcp-domain.com",
					GUID:            "tcp-domain-guid",
					RouterGroupGUID: "tcp-guid",
					RouterGroupType: "tcp",
				}
				domainNotFoundError := cferrors.NewModelNotFoundError("Domain", "some-domain.com")

				fakeDomainRepository.FindPrivateByNameReturns(models.DomainFields{}, domainNotFoundError)
				fakeDomainRepository.FindSharedByNameStub = func(name string) (models.DomainFields, error) {
					if name == "tcp-domain.com" {
						return tcpDomain, nil
					}
					return models.DomainFields{}, domainNotFoundError
				}
			})

			Context("and contains a path", func() {
				BeforeEach(func() {
					routeName = "tcp-domain.com:3333/path"
				})

				It("returns an error", func() {
					Expect(findAndBindRouteErr).To(HaveOccurred())
					Expect(findAndBindRouteErr.Error()).To(Equal("Path not allowed in TCP route tcp-domain.com:3333/path"))
				})
			})

			Context("and does not contain a path", func() {
				BeforeEach(func() {
					routeName = "tcp-domain.com:3333"

					fakeRouteRepository.FindReturns(models.Route{}, cferrors.NewModelNotFoundError("Route", "some-route"))
					fakeRouteRepository.CreateReturns(
						models.Route{
							GUID:   "route-guid",
							Domain: tcpDomain,
							Path:   "path",
						},
						nil,
					)
					fakeRouteRepository.BindReturns(nil)
				})

				It("creates and binds the route", func() {
					Expect(findAndBindRouteErr).NotTo(HaveOccurred())

					actualDomainName := fakeDomainRepository.FindSharedByNameArgsForCall(0)
					Expect(actualDomainName).To(Equal("tcp-domain.com"))

					actualHost, actualDomain, actualPath, actualPort := fakeRouteRepository.FindArgsForCall(0)
					Expect(actualHost).To(Equal(""))
					Expect(actualDomain).To(Equal(tcpDomain))
					Expect(actualPath).To(Equal(""))
					Expect(actualPort).To(Equal(3333))

					actualHost, actualDomain, actualPath, actualPort, actualUseRandomPort := fakeRouteRepository.CreateArgsForCall(0)
					Expect(actualHost).To(Equal(""))
					Expect(actualDomain).To(Equal(tcpDomain))
					Expect(actualPath).To(Equal(""))
					Expect(actualPort).To(Equal(3333))
					Expect(actualUseRandomPort).To(BeFalse())

					routeGUID, appGUID := fakeRouteRepository.BindArgsForCall(0)
					Expect(routeGUID).To(Equal("route-guid"))
					Expect(appGUID).To(Equal("app-guid"))
				})
			})

			Context("and the --hostname flag is provided", func() {
				BeforeEach(func() {
					routeName = "tcp-domain.com:3333"
					appParamsFromContext = models.AppParams{
						Hosts: []string{"flag-hostname"},
					}
				})

				It("should not change the route", func() {
					Expect(findAndBindRouteErr).NotTo(HaveOccurred())

					actualHost, actualDomain, actualPath, actualPort := fakeRouteRepository.FindArgsForCall(0)
					Expect(actualHost).To(Equal(""))
					Expect(actualDomain).To(Equal(tcpDomain))
					Expect(actualPath).To(Equal(""))
					Expect(actualPort).To(Equal(3333))
				})
			})
		})
	})
})
