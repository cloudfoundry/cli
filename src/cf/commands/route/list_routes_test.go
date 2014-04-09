package route_test

import (
	. "cf/commands/route"
	"cf/configuration"
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
		ui                  *testterm.FakeUI
		routeRepo           *testapi.FakeRouteRepository
		configRepo          configuration.ReadWriter
		requirementsFactory *testreq.FakeReqFactory
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{
			LoginSuccess: true,
			TargetedSpaceSuccess: true,
		}
		routeRepo = &testapi.FakeRouteRepository{}
	})

	runCommand := func(args ...string) {
		cmd := NewListRoutes(ui, configRepo, routeRepo)
		testcmd.RunCommand(cmd, testcmd.NewContext("routes", args), requirementsFactory)
	}

	Describe("login requirements", func() {
		It("fails if the user is not logged in", func() {
			requirementsFactory.LoginSuccess = false
			runCommand()
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})

		It("fails when an org and space is not targeted", func() {
			requirementsFactory.TargetedSpaceSuccess = false
			runCommand()
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})
	})

	Context("when there are routes", func() {
		BeforeEach(func() {
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
			routeRepo.Routes = []models.Route{route, route2}
		})

		It("lists routes", func() {
			runCommand()

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"Getting routes", "my-user"},
				{"host", "domain", "apps"},
				{"hostname-1", "example.com", "dora"},
				{"hostname-2", "cookieclicker.co", "dora", "bora"},
			})
		})
	})

	Context("when there are not routes", func() {
		It("tells the user when no routes were found", func() {
			runCommand()

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"Getting routes"},
				{"No routes found"},
			})
		})
	})

	Context("when there is an error listing routes", func() {
		BeforeEach(func() {
			routeRepo.ListErr = true
		})

		It("returns an error to the user", func() {
			runCommand()

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"Getting routes"},
				{"FAILED"},
			})
		})
	})
})
