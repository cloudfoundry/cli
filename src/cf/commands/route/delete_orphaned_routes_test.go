package route_test

import (
	. "cf/commands/route"
	"cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

var _ = Describe("delete-orphaned-routes command", func() {
	var routeRepo *testapi.FakeRouteRepository
	var reqFactory *testreq.FakeReqFactory

	BeforeEach(func() {
		routeRepo = &testapi.FakeRouteRepository{}
		reqFactory = &testreq.FakeReqFactory{}
	})

	It("fails requirements when not logged in", func() {
		callDeleteOrphanedRoutes("y", []string{}, reqFactory, routeRepo)
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})

	Context("when logged in successfully", func() {

		BeforeEach(func() {
			reqFactory.LoginSuccess = true
		})

		It("passes requirements when logged in", func() {
			callDeleteOrphanedRoutes("y", []string{}, reqFactory, routeRepo)
			Expect(testcmd.CommandDidPassRequirements).To(BeTrue())
		})

		It("passes when confirmation is provided", func() {
			var ui *testterm.FakeUI
			domain := models.DomainFields{Name: "example.com"}
			domain2 := models.DomainFields{Name: "cookieclicker.co"}

			app1 := models.ApplicationFields{Name: "dora"}

			route := models.Route{}
			route.Host = "hostname-1"
			route.Domain = domain
			route.Apps = []models.ApplicationFields{app1}

			route2 := models.Route{}
			route2.Guid = "route2-guid"
			route2.Host = "hostname-2"
			route2.Domain = domain2

			routeRepo.Routes = []models.Route{route, route2}

			ui = callDeleteOrphanedRoutes("y", []string{}, reqFactory, routeRepo)

			testassert.SliceContains(ui.Prompts, testassert.Lines{
				{"Really delete orphaned routes"},
			})

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"Deleting route", "hostname-2.cookieclicker.co"},
				{"OK"},
			})
			Expect(routeRepo.DeleteRouteGuid).To(Equal("route2-guid"))
		})

		It("passes when the force flag is used", func() {
			var ui *testterm.FakeUI
			domain := models.DomainFields{Name: "example.com"}
			domain2 := models.DomainFields{Name: "cookieclicker.co"}

			app1 := models.ApplicationFields{Name: "dora"}

			route := models.Route{}
			route.Host = "hostname-1"
			route.Domain = domain
			route.Apps = []models.ApplicationFields{app1}

			route2 := models.Route{}
			route2.Guid = "route2-guid"
			route2.Host = "hostname-2"
			route2.Domain = domain2

			routeRepo.Routes = []models.Route{route, route2}

			ui = callDeleteOrphanedRoutes("", []string{"-f"}, reqFactory, routeRepo)

			Expect(len(ui.Prompts)).To(Equal(0))

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"Deleting route", "hostname-2.cookieclicker.co"},
				{"OK"},
			})
			Expect(routeRepo.DeleteRouteGuid).To(Equal("route2-guid"))
		})
	})
})

func callDeleteOrphanedRoutes(confirmation string, args []string, reqFactory *testreq.FakeReqFactory, routeRepo *testapi.FakeRouteRepository) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{Inputs: []string{confirmation}}
	ctxt := testcmd.NewContext("delete-orphaned-routes", args)
	configRepo := testconfig.NewRepositoryWithDefaults()

	cmd := NewDeleteOrphanedRoutes(ui, configRepo, routeRepo)

	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
