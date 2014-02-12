package route_test

import (
	. "cf/commands/route"
	"cf/models"
	"github.com/codegangsta/cli"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

var _ = Describe("Unmap Route Command", func() {
	It("TestUnmapRouteFailsWithUsage", func() {
		reqFactory := &testreq.FakeReqFactory{}
		routeRepo := &testapi.FakeRouteRepository{}

		ui := callUnmapRoute([]string{}, reqFactory, routeRepo)
		assert.True(mr.T(), ui.FailedWithUsage)

		ui = callUnmapRoute([]string{"foo"}, reqFactory, routeRepo)
		assert.True(mr.T(), ui.FailedWithUsage)

		ui = callUnmapRoute([]string{"foo", "bar"}, reqFactory, routeRepo)
		assert.False(mr.T(), ui.FailedWithUsage)
	})

	It("TestUnmapRouteRequirements", func() {
		routeRepo := &testapi.FakeRouteRepository{}
		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}

		callUnmapRoute([]string{"-n", "my-host", "my-app", "my-domain.com"}, reqFactory, routeRepo)
		assert.True(mr.T(), testcmd.CommandDidPassRequirements)
		assert.Equal(mr.T(), reqFactory.ApplicationName, "my-app")
		assert.Equal(mr.T(), reqFactory.DomainName, "my-domain.com")
	})

	It("TestUnmapRouteWhenUnbinding", func() {
		domain := models.DomainFields{
			Guid: "my-domain-guid",
			Name: "example.com",
		}
		route := models.Route{RouteSummary: models.RouteSummary{
			Domain: domain,
			RouteFields: models.RouteFields{
				Guid: "my-route-guid",
				Host: "foo",
			},
		}}
		app := models.Application{ApplicationFields: models.ApplicationFields{
			Guid: "my-app-guid",
			Name: "my-app",
		}}

		routeRepo := &testapi.FakeRouteRepository{FindByHostAndDomainRoute: route}
		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, Application: app, Domain: domain}

		ui := callUnmapRoute([]string{"-n", "my-host", "my-app", "my-domain.com"}, reqFactory, routeRepo)

		testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
			{"Removing route", "foo.example.com", "my-app", "my-org", "my-space", "my-user"},
			{"OK"},
		})

		assert.Equal(mr.T(), routeRepo.UnboundRouteGuid, "my-route-guid")
		assert.Equal(mr.T(), routeRepo.UnboundAppGuid, "my-app-guid")
	})
})

func callUnmapRoute(args []string, reqFactory *testreq.FakeReqFactory, routeRepo *testapi.FakeRouteRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	var ctxt *cli.Context = testcmd.NewContext("unmap-route", args)

	configRepo := testconfig.NewRepositoryWithDefaults()
	cmd := NewUnmapRoute(ui, configRepo, routeRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
