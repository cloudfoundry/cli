package actors_test

import (
	"github.com/cloudfoundry/cli/cf/actors"
	"github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	testhelpers "github.com/cloudfoundry/cli/testhelpers/models"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Route Actor Integration Tests", func() {
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

	Describe("Complete Route Workflow", func() {
		It("creates route, binds to app, and unbinds successfully", func() {
			// Setup: Create test domain and app
			domain := testhelpers.MakeDomain("example.com", true)
			app := testhelpers.MakeApplication("my-app")

			// Step 1: Route doesn't exist, so create it
			routeRepo.FindByHostAndDomainReturns.Error = errors.NewModelNotFoundError("Route", "my-app")
			createdRoute := testhelpers.MakeRoute("my-app", "example.com")
			routeRepo.CreatedRoute = createdRoute

			route := actor.FindOrCreateRoute("my-app", domain)

			// Verify route was created
			Expect(route.Guid).To(Equal(createdRoute.Guid))
			Expect(ui.Outputs).To(ContainElement(ContainSubstring("Creating route")))
			Expect(ui.Outputs).To(ContainElement(ContainSubstring("OK")))

			// Step 2: Bind route to app
			routeRepo.BindErr = nil
			actor.BindRoute(app, route)

			// Verify binding happened
			Expect(routeRepo.BoundRouteGuid).To(Equal(route.Guid))
			Expect(routeRepo.BoundAppGuid).To(Equal(app.Guid))
			Expect(ui.Outputs).To(ContainElement(ContainSubstring("Binding")))

			// Step 3: App now has the route
			app.Routes = []models.RouteSummary{
				{
					Guid: route.Guid,
					Host: route.Host,
					Domain: route.Domain,
				},
			}

			// Step 4: Unbind all routes
			actor.UnbindAll(app)

			// Verify unbinding happened
			Expect(routeRepo.UnboundRouteGuid).To(Equal(route.Guid))
			Expect(ui.Outputs).To(ContainElement(ContainSubstring("Removing route")))
		})

		It("reuses existing route and binds to app", func() {
			// Setup
			domain := testhelpers.MakeDomain("example.com", true)
			app := testhelpers.MakeApplication("my-app")

			// Route already exists
			existingRoute := testhelpers.MakeRoute("my-app", "example.com")
			routeRepo.FindByHostAndDomainReturns.Route = existingRoute
			routeRepo.FindByHostAndDomainReturns.Error = nil

			// Find existing route
			route := actor.FindOrCreateRoute("my-app", domain)

			// Verify route was found, not created
			Expect(route.Guid).To(Equal(existingRoute.Guid))
			Expect(ui.Outputs).To(ContainElement(ContainSubstring("Using route")))
			Expect(routeRepo.CreatedHost).To(BeEmpty())

			// Bind to app
			actor.BindRoute(app, route)

			// Verify binding
			Expect(routeRepo.BoundRouteGuid).To(Equal(route.Guid))
		})

		It("handles multiple routes on single app", func() {
			// Setup app with multiple routes
			routes := []models.RouteSummary{
				{Guid: "route-1", Host: "host1", Domain: models.DomainFields{Name: "example.com"}},
				{Guid: "route-2", Host: "host2", Domain: models.DomainFields{Name: "example.com"}},
				{Guid: "route-3", Host: "host3", Domain: models.DomainFields{Name: "test.com"}},
			}

			app := testhelpers.MakeApplication("my-app",
				testhelpers.WithRoutes(routes...),
			)

			// Unbind all routes
			actor.UnbindAll(app)

			// Verify all routes were unbound
			Expect(ui.Outputs).To(ContainElement(ContainSubstring("host1.example.com")))
			Expect(ui.Outputs).To(ContainElement(ContainSubstring("host2.example.com")))
			Expect(ui.Outputs).To(ContainElement(ContainSubstring("host3.test.com")))
		})

		It("handles route already bound scenario", func() {
			// Setup
			domain := testhelpers.MakeDomain("example.com", true)
			route := testhelpers.MakeRoute("my-app", "example.com")

			// App already has this route
			app := testhelpers.MakeApplication("my-app",
				testhelpers.WithRoutes(models.RouteSummary{
					Guid:   route.Guid,
					Host:   route.Host,
					Domain: route.Domain,
				}),
			)

			// Try to bind again
			actor.BindRoute(app, route)

			// Verify no binding happened (route already bound)
			Expect(routeRepo.BoundRouteGuid).To(BeEmpty())
			Expect(len(ui.Outputs)).To(Equal(0))
		})

		It("handles error when route is in use by another app", func() {
			// Setup
			domain := testhelpers.MakeDomain("example.com", true)
			app := testhelpers.MakeApplication("my-app")

			// Route creation fails because another app is using it
			routeRepo.FindByHostAndDomainReturns.Error = errors.NewModelNotFoundError("Route", "my-app")
			route := testhelpers.MakeRoute("my-app", "example.com")
			routeRepo.CreatedRoute = route

			// Create route
			createdRoute := actor.FindOrCreateRoute("my-app", domain)

			// Binding fails with INVALID_RELATION
			routeRepo.BindErr = errors.NewHttpError(400, errors.INVALID_RELATION, "Route already in use")

			// Try to bind
			actor.BindRoute(app, createdRoute)

			// Verify error message displayed
			Expect(ui.Outputs).To(ContainElement(ContainSubstring("FAILED")))
			Expect(ui.Outputs).To(ContainElement(ContainSubstring("already in use")))
		})
	})

	Describe("Complex Routing Scenarios", func() {
		It("handles app with no routes", func() {
			app := testhelpers.MakeApplication("my-app")

			// Unbind all (there are none)
			actor.UnbindAll(app)

			// Verify no unbinding happened
			Expect(routeRepo.UnboundRouteGuid).To(BeEmpty())
		})

		It("handles creating multiple routes for same app", func() {
			app := testhelpers.MakeApplication("my-app")
			domain1 := testhelpers.MakeDomain("example.com", true)
			domain2 := testhelpers.MakeDomain("test.com", true)

			// Create first route
			routeRepo.FindByHostAndDomainReturns.Error = errors.NewModelNotFoundError("Route", "host1")
			route1 := testhelpers.MakeRoute("host1", "example.com")
			routeRepo.CreatedRoute = route1

			createdRoute1 := actor.FindOrCreateRoute("host1", domain1)
			actor.BindRoute(app, createdRoute1)

			firstBoundRoute := routeRepo.BoundRouteGuid

			// Create second route
			routeRepo.FindByHostAndDomainReturns.Error = errors.NewModelNotFoundError("Route", "host2")
			route2 := testhelpers.MakeRoute("host2", "test.com")
			routeRepo.CreatedRoute = route2

			createdRoute2 := actor.FindOrCreateRoute("host2", domain2)
			actor.BindRoute(app, createdRoute2)

			secondBoundRoute := routeRepo.BoundRouteGuid

			// Verify both routes were bound
			Expect(firstBoundRoute).To(Equal(route1.Guid))
			Expect(secondBoundRoute).To(Equal(route2.Guid))
		})
	})
})
