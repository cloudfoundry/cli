package route_test

import (
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/commands/route"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("map-route command", func() {
	var (
		ui                  *testterm.FakeUI
		configRepo          core_config.ReadWriter
		routeRepo           *testapi.FakeRouteRepository
		requirementsFactory *testreq.FakeReqFactory
		routeCreator        *testcmd.FakeRouteCreator
	)

	BeforeEach(func() {
		ui = new(testterm.FakeUI)
		configRepo = testconfig.NewRepositoryWithDefaults()
		routeRepo = new(testapi.FakeRouteRepository)
		routeCreator = &testcmd.FakeRouteCreator{}
		requirementsFactory = new(testreq.FakeReqFactory)
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCommand(NewMapRoute(ui, configRepo, routeRepo, routeCreator), args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("fails when not invoked with exactly two args", func() {
			runCommand("whoops-all-crunchberries")
			Expect(ui.FailedWithUsage).To(BeTrue())
		})

		It("fails when not logged in", func() {
			Expect(runCommand("whatever", "shuttup")).To(BeFalse())
		})
	})

	Context("when the user is logged in", func() {
		BeforeEach(func() {
			domain := models.DomainFields{Guid: "my-domain-guid", Name: "example.com"}
			route := models.Route{Guid: "my-route-guid", Host: "foo", Domain: domain}

			app := models.Application{}
			app.Guid = "my-app-guid"
			app.Name = "my-app"

			requirementsFactory.LoginSuccess = true
			requirementsFactory.Application = app
			requirementsFactory.Domain = domain
			routeCreator.ReservedRoute = route
		})

		It("maps a route, obviously", func() {
			passed := runCommand("-n", "my-host", "my-app", "my-domain.com")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Adding route", "foo.example.com", "my-app", "my-org", "my-space", "my-user"},
				[]string{"OK"},
			))

			Expect(routeRepo.BoundRouteGuid).To(Equal("my-route-guid"))
			Expect(routeRepo.BoundAppGuid).To(Equal("my-app-guid"))
			Expect(passed).To(BeTrue())
			Expect(requirementsFactory.ApplicationName).To(Equal("my-app"))
			Expect(requirementsFactory.DomainName).To(Equal("my-domain.com"))
		})
	})
})
