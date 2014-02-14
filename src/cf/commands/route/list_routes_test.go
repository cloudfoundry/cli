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

var _ = Describe("routes command", func() {
	var (
		ui         *testterm.FakeUI
		cmd        ListRoutes
		repo       *testapi.FakeRouteRepository
		reqFactory *testreq.FakeReqFactory
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		config := testconfig.NewRepositoryWithDefaults()
		reqFactory = &testreq.FakeReqFactory{LoginSuccess: true}
		repo = &testapi.FakeRouteRepository{}
		cmd = NewListRoutes(ui, config, repo)
	})

	Describe("login requirements", func() {
		It("fails if the user is not logged in", func() {
			reqFactory.LoginSuccess = false
			context := testcmd.NewContext("routes", []string{""})
			testcmd.RunCommand(cmd, context, reqFactory)
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})
	})

	It("lists routes", func() {
		domain := models.DomainFields{Name: "example.com"}
		domain2 := models.DomainFields{Name: "cookieclicker.co"}

		app1 := models.ApplicationFields{Name: "dora"}
		app2 := models.ApplicationFields{Name: "bora"}

		route := models.Route{}
		route.Host = "hostname-1"
		route.Domain = domain
		route.Apps = []models.ApplicationFields{app1}

		route2 := models.Route{}
		route2.Host = "hostname-2"
		route2.Domain = domain2
		route2.Apps = []models.ApplicationFields{app1, app2}

		repo.Routes = []models.Route{route, route2}
		context := testcmd.NewContext("routes", []string{})
		testcmd.RunCommand(cmd, context, reqFactory)

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Getting routes", "my-user"},
			{"host", "domain", "apps"},
			{"hostname-1", "example.com", "dora"},
			{"hostname-2", "cookieclicker.co", "dora", "bora"},
		})
	})

	It("tells the user when no routes were found", func() {
		context := testcmd.NewContext("routes", []string{})
		testcmd.RunCommand(cmd, context, reqFactory)

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Getting routes"},
			{"No routes found"},
		})
	})

	It("reports an error when finding routes fails", func() {
		repo.ListErr = true
		context := testcmd.NewContext("routes", []string{})
		testcmd.RunCommand(cmd, context, reqFactory)

		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Getting routes"},
			{"FAILED"},
		})
	})
})
