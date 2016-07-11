package actors_test

import (
	"errors"

	. "github.com/cloudfoundry/cli/cf/actors"
	"github.com/cloudfoundry/cli/cf/errors/errorsfakes"

	"github.com/cloudfoundry/cli/cf/api/apifakes"
	cferrors "github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/cli/cf/terminal/terminalfakes"
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

			host, d, path, randomPort := fakeRouteRepository.CreateArgsForCall(0)
			Expect(host).To(BeEmpty())
			Expect(d).To(Equal(expectedDomain))
			Expect(path).To(BeEmpty())
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
				route, err := routeActor.FindOrCreateRoute(expectedHostname, expectedDomain, expectedPath, false)
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
					route, err := routeActor.FindOrCreateRoute("", expectedDomain, "", true)
					Expect(route).To(Equal(tcpRoute))
					Expect(err).ToNot(HaveOccurred())

					Expect(fakeRouteRepository.CreateCallCount()).To(Equal(1))
					hostname, domain, path, randomPort := fakeRouteRepository.CreateArgsForCall(0)
					Expect(hostname).To(BeEmpty())
					Expect(domain).To(Equal(expectedDomain))
					Expect(path).To(BeEmpty())
					Expect(randomPort).To(BeTrue())

					Expect(fakeUI.SayCallCount()).To(Equal(2))
					output, _ := fakeUI.SayArgsForCall(0)
					Expect(output).To(MatchRegexp("Creating random route for"))
				})
			})

			Context("with out a random port", func() {
				BeforeEach(func() {
					fakeRouteRepository.CreateReturns(expectedRoute, nil)
				})

				It("creates a route ", func() {
					route, err := routeActor.FindOrCreateRoute(expectedHostname, expectedDomain, expectedPath, false)
					Expect(route).To(Equal(expectedRoute))
					Expect(err).ToNot(HaveOccurred())

					Expect(fakeRouteRepository.CreateCallCount()).To(Equal(1))
					hostname, domain, path, randomPort := fakeRouteRepository.CreateArgsForCall(0)
					Expect(hostname).To(Equal(expectedHostname))
					Expect(domain).To(Equal(expectedDomain))
					Expect(path).To(Equal(expectedPath))
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
				Context("and the domain is HTTP", func() {
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

				Context("and the domain is not HTTP", func() {
					BeforeEach(func() {
						fakeDomainRepository.FindSharedByNameStub = func(name string) (models.DomainFields, error) {
							return models.DomainFields{
									GUID:            "non-http-shared-domain-guid",
									Name:            "non-http-shared-domain",
									RouterGroupType: "not-null",
								},
								nil
						}
					})

					It("returns an error", func() {
						Expect(findDomainErr).To(HaveOccurred())
						Expect(findDomainErr.Error()).To(Equal("The domain non-http-shared-domain is not an HTTP domain, unable to map route."))
					})
				})
			})

			Context("when the route has a hostname", func() {
				Context("and the domain is HTTP", func() {
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

				Context("and the domain is not HTTP", func() {
					BeforeEach(func() {
						fakeDomainRepository.FindSharedByNameStub = func(name string) (models.DomainFields, error) {
							if name == "my-domain.com" {
								return models.DomainFields{
										GUID:            "non-http-shared-domain-guid",
										Name:            "non-http-shared-domain",
										RouterGroupType: "not-null",
									},
									nil
							}
							return models.DomainFields{}, domainNotFoundError
						}
					})

					It("returns an error", func() {
						Expect(findDomainErr).To(HaveOccurred())
						Expect(fakeDomainRepository.FindSharedByNameCallCount()).To(Equal(2))
						Expect(findDomainErr.Error()).To(Equal("The domain non-http-shared-domain is not an HTTP domain, unable to map route."))
					})
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
})
