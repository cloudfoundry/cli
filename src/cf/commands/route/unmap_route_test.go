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

var _ = Describe("Unmap Route Command", func() {
	It("TestUnmapRouteFailsWithUsage", func() {
		reqFactory := &testreq.FakeReqFactory{}
		routeRepo := &testapi.FakeRouteRepository{}

		ui := callUnmapRoute([]string{}, reqFactory, routeRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callUnmapRoute([]string{"foo"}, reqFactory, routeRepo)
		Expect(ui.FailedWithUsage).To(BeTrue())

		ui = callUnmapRoute([]string{"foo", "bar"}, reqFactory, routeRepo)
		Expect(ui.FailedWithUsage).To(BeFalse())
	})

	It("TestUnmapRouteRequirements", func() {
		routeRepo := &testapi.FakeRouteRepository{}
		reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}

		callUnmapRoute([]string{"-n", "my-host", "my-app", "my-domain.com"}, reqFactory, routeRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeTrue())
		Expect(reqFactory.ApplicationName).To(Equal("my-app"))
		Expect(reqFactory.DomainName).To(Equal("my-domain.com"))
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

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Removing route", "foo.example.com", "my-app", "my-org", "my-space", "my-user"},
			{"OK"},
		})

		Expect(routeRepo.UnboundRouteGuid).To(Equal("my-route-guid"))
		Expect(routeRepo.UnboundAppGuid).To(Equal("my-app-guid"))
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
