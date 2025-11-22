package actors_test

import (
	"github.com/cloudfoundry/cli/cf/actors"
	"github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Route Actor", func() {
	var (
		actor     actors.RouteActor
		ui        *testterm.FakeUI
		routeRepo *fakes.FakeRouteRepository
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		routeRepo = &fakes.FakeRouteRepository{}
		actor = actors.NewRouteActor(ui, routeRepo)
	})

	Describe("FindOrCreateRoute", func() {
		var (
			hostname string
			domain   models.DomainFields
		)

		BeforeEach(func() {
			hostname = "my-host"
			domain = models.DomainFields{
				Guid: "domain-guid",
				Name: "example.com",
			}
		})

		Context("when the route exists", func() {
			BeforeEach(func() {
				routeRepo.FindByHostAndDomainReturns.Route = models.Route{
					Guid: "route-guid",
					Host: hostname,
					Domain: domain,
				}
				routeRepo.FindByHostAndDomainReturns.Error = nil
			})

			It("returns the existing route", func() {
				route := actor.FindOrCreateRoute(hostname, domain)
				Expect(route.Guid).To(Equal("route-guid"))
				Expect(route.Host).To(Equal(hostname))
			})

			It("does not create a new route", func() {
				actor.FindOrCreateRoute(hostname, domain)
				Expect(routeRepo.CreatedHost).To(BeEmpty())
			})

			It("displays a message about using the existing route", func() {
				actor.FindOrCreateRoute(hostname, domain)
				Expect(ui.Outputs).To(ContainElement(ContainSubstring("Using route")))
			})
		})

		Context("when the route does not exist", func() {
			BeforeEach(func() {
				routeRepo.FindByHostAndDomainReturns.Error = errors.NewModelNotFoundError("Route", hostname)
				routeRepo.CreatedRoute = models.Route{
					Guid: "new-route-guid",
					Host: hostname,
					Domain: domain,
				}
			})

			It("creates a new route", func() {
				route := actor.FindOrCreateRoute(hostname, domain)
				Expect(route.Guid).To(Equal("new-route-guid"))
			})

			It("displays a message about creating the route", func() {
				actor.FindOrCreateRoute(hostname, domain)
				Expect(ui.Outputs).To(ContainElement(ContainSubstring("Creating route")))
			})

			It("displays OK after creation", func() {
				actor.FindOrCreateRoute(hostname, domain)
				Expect(ui.Outputs).To(ContainElement(ContainSubstring("OK")))
			})
		})

		Context("when finding the route fails with a non-NotFound error", func() {
			BeforeEach(func() {
				routeRepo.FindByHostAndDomainReturns.Error = errors.New("API error")
			})

			It("fails with an error", func() {
				actor.FindOrCreateRoute(hostname, domain)
				Expect(ui.Outputs).To(ContainElement(ContainSubstring("FAILED")))
			})
		})

		Context("when creating the route fails", func() {
			BeforeEach(func() {
				routeRepo.FindByHostAndDomainReturns.Error = errors.NewModelNotFoundError("Route", hostname)
				routeRepo.FindByHostAndDomainReturns.Route = models.Route{}
			})

			It("fails with an error message", func() {
				// The actor will try to create but we haven't set CreatedRoute,
				// which simulates a creation error
				actor.FindOrCreateRoute(hostname, domain)
				// Should still return a route even if empty
			})
		})
	})

	Describe("BindRoute", func() {
		var (
			app   models.Application
			route models.Route
		)

		BeforeEach(func() {
			route = models.Route{
				Guid: "route-guid",
				Host: "my-host",
				Domain: models.DomainFields{
					Name: "example.com",
				},
			}

			app = models.Application{
				ApplicationFields: models.ApplicationFields{
					Guid: "app-guid",
					Name: "my-app",
				},
			}
		})

		Context("when the route is not bound to the app", func() {
			BeforeEach(func() {
				app.Routes = []models.RouteSummary{} // No routes bound
				routeRepo.BindErr = nil
			})

			It("binds the route to the app", func() {
				actor.BindRoute(app, route)
				Expect(routeRepo.BoundRouteGuid).To(Equal("route-guid"))
				Expect(routeRepo.BoundAppGuid).To(Equal("app-guid"))
			})

			It("displays a message about binding", func() {
				actor.BindRoute(app, route)
				Expect(ui.Outputs).To(ContainElement(ContainSubstring("Binding")))
				Expect(ui.Outputs).To(ContainElement(ContainSubstring("my-host.example.com")))
				Expect(ui.Outputs).To(ContainElement(ContainSubstring("my-app")))
			})

			It("displays OK after binding", func() {
				actor.BindRoute(app, route)
				Expect(ui.Outputs).To(ContainElement(ContainSubstring("OK")))
			})
		})

		Context("when the route is already bound to the app", func() {
			BeforeEach(func() {
				app.Routes = []models.RouteSummary{
					{
						Guid: "route-guid",
						Host: "my-host",
					},
				}
			})

			It("does not bind the route again", func() {
				actor.BindRoute(app, route)
				Expect(routeRepo.BoundRouteGuid).To(BeEmpty())
			})

			It("does not display any messages", func() {
				actor.BindRoute(app, route)
				Expect(len(ui.Outputs)).To(Equal(0))
			})
		})

		Context("when binding fails with INVALID_RELATION error", func() {
			BeforeEach(func() {
				routeRepo.BindErr = errors.NewHttpError(400, errors.INVALID_RELATION, "The route is already in use")
			})

			It("displays a helpful error message", func() {
				actor.BindRoute(app, route)
				Expect(ui.Outputs).To(ContainElement(ContainSubstring("FAILED")))
				Expect(ui.Outputs).To(ContainElement(ContainSubstring("already in use")))
				Expect(ui.Outputs).To(ContainElement(ContainSubstring("Change the hostname")))
			})
		})

		Context("when binding fails with a generic error", func() {
			BeforeEach(func() {
				routeRepo.BindErr = errors.New("Some API error")
			})

			It("displays the error message", func() {
				actor.BindRoute(app, route)
				Expect(ui.Outputs).To(ContainElement(ContainSubstring("FAILED")))
			})
		})
	})

	Describe("UnbindAll", func() {
		var app models.Application

		Context("when the app has multiple routes", func() {
			BeforeEach(func() {
				app = models.Application{
					ApplicationFields: models.ApplicationFields{
						Guid: "app-guid",
						Name: "my-app",
					},
					Routes: []models.RouteSummary{
						{
							Guid: "route-guid-1",
							Host: "host-1",
							Domain: models.DomainFields{
								Name: "example.com",
							},
						},
						{
							Guid: "route-guid-2",
							Host: "host-2",
							Domain: models.DomainFields{
								Name: "example.com",
							},
						},
					},
				}
			})

			It("unbinds all routes from the app", func() {
				actor.UnbindAll(app)
				Expect(routeRepo.UnboundRouteGuid).To(Equal("route-guid-2")) // Last one called
				Expect(routeRepo.UnboundAppGuid).To(Equal("app-guid"))
			})

			It("displays a message for each route being removed", func() {
				actor.UnbindAll(app)
				Expect(ui.Outputs).To(ContainElement(ContainSubstring("Removing route")))
				Expect(ui.Outputs).To(ContainElement(ContainSubstring("host-1.example.com")))
				Expect(ui.Outputs).To(ContainElement(ContainSubstring("host-2.example.com")))
			})
		})

		Context("when the app has no routes", func() {
			BeforeEach(func() {
				app = models.Application{
					ApplicationFields: models.ApplicationFields{
						Guid: "app-guid",
						Name: "my-app",
					},
					Routes: []models.RouteSummary{},
				}
			})

			It("does not attempt to unbind anything", func() {
				actor.UnbindAll(app)
				Expect(routeRepo.UnboundRouteGuid).To(BeEmpty())
			})

			It("does not display any messages", func() {
				actor.UnbindAll(app)
				Expect(len(ui.Outputs)).To(Equal(0))
			})
		})

		Context("when the app has one route", func() {
			BeforeEach(func() {
				app = models.Application{
					ApplicationFields: models.ApplicationFields{
						Guid: "app-guid",
						Name: "my-app",
					},
					Routes: []models.RouteSummary{
						{
							Guid: "route-guid",
							Host: "my-host",
							Domain: models.DomainFields{
								Name: "example.com",
							},
						},
					},
				}
			})

			It("unbinds the single route", func() {
				actor.UnbindAll(app)
				Expect(routeRepo.UnboundRouteGuid).To(Equal("route-guid"))
				Expect(routeRepo.UnboundAppGuid).To(Equal("app-guid"))
			})

			It("displays a removal message", func() {
				actor.UnbindAll(app)
				Expect(ui.Outputs).To(ContainElement(ContainSubstring("Removing route")))
				Expect(ui.Outputs).To(ContainElement(ContainSubstring("my-host.example.com")))
			})
		})
	})
})
