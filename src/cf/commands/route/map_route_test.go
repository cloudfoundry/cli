/*
                       WARNING WARNING WARNING

                Attention all potential contributors

   This testfile is not in the best state. We've been slowly transitioning
   from the built in "testing" package to using Ginkgo. As you can see, we've
   changed the format, but a lot of the setup, test body, descriptions, etc
   are either hardcoded, completely lacking, or misleading.

   For example:

   Describe("Testing with ginkgo"...)      // This is not a great description
   It("TestDoesSoemthing"...)              // This is a horrible description

   Describe("create-user command"...       // Describe the actual object under test
   It("creates a user when provided ..."   // this is more descriptive

   For good examples of writing Ginkgo tests for the cli, refer to

   src/cf/commands/application/delete_app_test.go
   src/cf/terminal/ui_test.go
   src/github.com/cloudfoundry/loggregator_consumer/consumer_test.go
*/

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

func callMapRoute(args []string, requirementsFactory *testreq.FakeReqFactory, routeRepo *testapi.FakeRouteRepository, createRoute *testcmd.FakeRouteCreator) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	var ctxt *cli.Context = testcmd.NewContext("map-route", args)

	configRepo := testconfig.NewRepositoryWithDefaults()
	cmd := NewMapRoute(ui, configRepo, routeRepo, createRoute)
	testcmd.RunCommand(cmd, ctxt, requirementsFactory)
	return
}

var _ = Describe("Testing with ginkgo", func() {
	It("TestMapRouteFailsWithUsage", func() {
		requirementsFactory := &testreq.FakeReqFactory{}
		routeRepo := &testapi.FakeRouteRepository{}

		ui := callMapRoute([]string{}, requirementsFactory, routeRepo, &testcmd.FakeRouteCreator{})
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callMapRoute([]string{"foo"}, requirementsFactory, routeRepo, &testcmd.FakeRouteCreator{})
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callMapRoute([]string{"foo", "bar"}, requirementsFactory, routeRepo, &testcmd.FakeRouteCreator{})
		Expect(ui.FailedWithUsage).To(BeFalse())
	})
	It("TestMapRouteRequirements", func() {

		routeRepo := &testapi.FakeRouteRepository{}
		requirementsFactory := &testreq.FakeReqFactory{LoginSuccess: true}

		callMapRoute([]string{"-n", "my-host", "my-app", "my-domain.com"}, requirementsFactory, routeRepo, &testcmd.FakeRouteCreator{})
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())
		Expect(requirementsFactory.ApplicationName).To(Equal("my-app"))
		Expect(requirementsFactory.DomainName).To(Equal("my-domain.com"))
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
		requirementsFactory := &testreq.FakeReqFactory{LoginSuccess: true, Application: app, Domain: domain}
		routeCreator := &testcmd.FakeRouteCreator{ReservedRoute: route}

		ui := callMapRoute([]string{"-n", "my-host", "my-app", "my-domain.com"}, requirementsFactory, routeRepo, routeCreator)

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
		requirementsFactory := &testreq.FakeReqFactory{LoginSuccess: true, Application: app}
		routeCreator := &testcmd.FakeRouteCreator{ReservedRoute: route}

		callMapRoute([]string{"-n", "my-host", "my-app", "my-domain.com"}, requirementsFactory, routeRepo, routeCreator)

		Expect(routeCreator.ReservedRoute).To(Equal(route))
	})
})
