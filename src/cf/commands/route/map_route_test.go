package route_test

import (
	. "cf/commands/route"
	"cf/models"
	"github.com/codegangsta/cli"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

func callMapRoute(args []string, reqFactory *testreq.FakeReqFactory, routeRepo *testapi.FakeRouteRepository, createRoute *testcmd.FakeRouteCreator) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	var ctxt *cli.Context = testcmd.NewContext("map-route", args)

	configRepo := testconfig.NewRepositoryWithDefaults()
	cmd := NewMapRoute(ui, configRepo, routeRepo, createRoute)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}

var _ = Describe("Testing with ginkgo", func() {
	It("TestMapRouteFailsWithUsage", func() {
		reqFactory := &testreq.FakeReqFactory{}
		routeRepo := &testapi.FakeRouteRepository{}

		ui := callMapRoute([]string{}, reqFactory, routeRepo, &testcmd.FakeRouteCreator{})
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callMapRoute([]string{"foo"}, reqFactory, routeRepo, &testcmd.FakeRouteCreator{})
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callMapRoute([]string{"foo", "bar"}, reqFactory, routeRepo, &testcmd.FakeRouteCreator{})
		Expect(ui.FailedWithUsage).To(BeFalse())
	})
	It("TestMapRouteRequirements", func() {

		routeRepo := &testapi.FakeRouteRepository{}
		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}

		callMapRoute([]string{"-n", "my-host", "my-app", "my-domain.com"}, reqFactory, routeRepo, &testcmd.FakeRouteCreator{})
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())
		Expect(reqFactory.ApplicationName).To(Equal("my-app"))
		Expect(reqFactory.DomainName).To(Equal("my-domain.com"))
	})
	It("TestMapRouteWhenBinding", func() {

		domain := models.DomainFields{}
		domain.Guid = "my-domain-guid"
		domain.Name = "example.com"
		route := models.Route{}
		route.Guid = "my-route-guid"
		route.Host = "foo"
		route.Domain = domain

		app := models.Application{}
		app.Guid = "my-app-guid"
		app.Name = "my-app"

		routeRepo := &testapi.FakeRouteRepository{}
		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, Application: app, Domain: domain}
		routeCreator := &testcmd.FakeRouteCreator{ReservedRoute: route}

		ui := callMapRoute([]string{"-n", "my-host", "my-app", "my-domain.com"}, reqFactory, routeRepo, routeCreator)

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Adding route", "foo.example.com", "my-app", "my-org", "my-space", "my-user"},
			{"OK"},
		})

		Expect(routeRepo.BoundRouteGuid).To(Equal("my-route-guid"))
		Expect(routeRepo.BoundAppGuid).To(Equal("my-app-guid"))
	})
	It("TestMapRouteWhenRouteNotReserved", func() {

		domain := models.DomainFields{}
		domain.Name = "my-domain.com"
		route := models.Route{}
		route.Guid = "my-app-guid"
		route.Host = "my-host"
		route.Domain = domain
		app := models.Application{}
		app.Guid = "my-app-guid"
		app.Name = "my-app"

		routeRepo := &testapi.FakeRouteRepository{}
		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, Application: app}
		routeCreator := &testcmd.FakeRouteCreator{ReservedRoute: route}

		callMapRoute([]string{"-n", "my-host", "my-app", "my-domain.com"}, reqFactory, routeRepo, routeCreator)

		Expect(routeCreator.ReservedRoute).To(Equal(route))
	})
})
