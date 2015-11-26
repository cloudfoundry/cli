package route_test

import (
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	routeCmdFakes "github.com/cloudfoundry/cli/cf/commands/route/fakes"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("map-route command", func() {
	var (
		ui                  *testterm.FakeUI
		configRepo          core_config.Repository
		routeRepo           *testapi.FakeRouteRepository
		requirementsFactory *testreq.FakeReqFactory
		routeCreator        *routeCmdFakes.FakeRouteCreator
		OriginalCreateRoute command_registry.Command
		deps                command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.RepoLocator = deps.RepoLocator.SetRouteRepository(routeRepo)
		deps.Config = configRepo

		//save original create-route and restore later
		OriginalCreateRoute = command_registry.Commands.FindCommand("create-route")
		//inject fake 'CreateRoute' into registry
		command_registry.Register(routeCreator)

		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("map-route").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = new(testterm.FakeUI)
		configRepo = testconfig.NewRepositoryWithDefaults()
		routeRepo = new(testapi.FakeRouteRepository)
		routeCreator = &routeCmdFakes.FakeRouteCreator{}
		requirementsFactory = new(testreq.FakeReqFactory)
	})

	AfterEach(func() {
		command_registry.Register(OriginalCreateRoute)
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCliCommand("map-route", args, requirementsFactory, updateCommandDependency, false)
	}

	Describe("requirements", func() {
		It("fails when not invoked with exactly two args", func() {
			runCommand("whoops-all-crunchberries")
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires", "arguments"},
			))
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
