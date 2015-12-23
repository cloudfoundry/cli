package route_test

import (
	"errors"

	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

var _ = Describe("routes command", func() {
	var (
		ui                  *testterm.FakeUI
		routeRepo           *testapi.FakeRouteRepository
		configRepo          core_config.Repository
		requirementsFactory *testreq.FakeReqFactory
		deps                command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.RepoLocator = deps.RepoLocator.SetRouteRepository(routeRepo)
		deps.Config = configRepo
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("routes").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{
			LoginSuccess:         true,
			TargetedSpaceSuccess: true,
		}
		routeRepo = &testapi.FakeRouteRepository{}
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCliCommand("routes", args, requirementsFactory, updateCommandDependency, false)
	}

	Describe("login requirements", func() {
		It("fails if the user is not logged in", func() {
			requirementsFactory.LoginSuccess = false
			Expect(runCommand()).To(BeFalse())
		})

		It("fails when an org and space is not targeted", func() {
			requirementsFactory.TargetedSpaceSuccess = false

			Expect(runCommand()).To(BeFalse())
		})
		It("should fail with usage when provided any arguments", func() {
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedSpaceSuccess = true
			Expect(runCommand("blahblah")).To(BeFalse())
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Incorrect Usage", "No argument required"},
			))
		})
	})

	Context("when there are routes", func() {
		BeforeEach(func() {
			routeRepo.ListRoutesStub = func(cb func(models.Route) bool) error {
				app1 := models.ApplicationFields{Name: "dora"}

				route := models.Route{}
				route.Host = "hostname-1"
				route.Domain = models.DomainFields{Name: "example.com"}
				route.Apps = []models.ApplicationFields{app1}

				cb(route)

				domain2 := models.DomainFields{Name: "cookieclicker.co"}
				app2 := models.ApplicationFields{Name: "bora"}

				route2 := models.Route{}
				route2.Host = "hostname-2"
				route2.Path = "/foo"
				route2.Domain = domain2
				route2.Apps = []models.ApplicationFields{app1, app2}

				cb(route2)

				return nil
			}
		})

		It("lists routes", func() {
			runCommand()

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Getting routes", "my-user"},
				[]string{"host", "domain", "apps"},
				[]string{"hostname-1", "example.com", "dora"},
				[]string{"hostname-2", "cookieclicker.co", "/foo", "dora", "bora"},
			))
		})
	})

	Context("when there are routes in different spaces", func() {
		BeforeEach(func() {
			routeRepo.ListAllRoutesStub = func(cb func(models.Route) bool) error {
				space1 := models.SpaceFields{Name: "space-1"}
				space2 := models.SpaceFields{Name: "space-2"}

				domain := models.DomainFields{Name: "example.com"}
				domain2 := models.DomainFields{Name: "cookieclicker.co"}

				app1 := models.ApplicationFields{Name: "dora"}
				app2 := models.ApplicationFields{Name: "bora"}

				route := models.Route{}
				route.Host = "hostname-1"
				route.Domain = domain
				route.Apps = []models.ApplicationFields{app1}
				route.Space = space1

				route2 := models.Route{}
				route2.Host = "hostname-2"
				route2.Path = "/foo"
				route2.Domain = domain2
				route2.Apps = []models.ApplicationFields{app1, app2}
				route2.Space = space2

				cb(route)
				cb(route2)

				return nil
			}
		})

		It("lists routes at orglevel", func() {
			runCommand("--orglevel")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Getting routes", "my-user"},
				[]string{"space", "host", "domain", "apps"},
				[]string{"space-1", "hostname-1", "example.com", "dora"},
				[]string{"space-2", "hostname-2", "cookieclicker.co", "/foo", "dora", "bora"},
			))
		})
	})

	Context("when there are not routes", func() {
		It("tells the user when no routes were found", func() {
			runCommand()

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Getting routes"},
				[]string{"No routes found"},
			))
		})
	})

	Context("when there is an error listing routes", func() {
		BeforeEach(func() {
			routeRepo.ListRoutesReturns(errors.New("an-error"))
		})

		It("returns an error to the user", func() {
			runCommand()

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Getting routes"},
				[]string{"FAILED"},
			))
		})
	})
})
