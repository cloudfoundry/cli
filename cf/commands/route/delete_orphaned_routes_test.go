package route_test

import (
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
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

var _ = Describe("delete-orphaned-routes command", func() {
	var (
		ui         *testterm.FakeUI
		routeRepo  *testapi.FakeRouteRepository
		configRepo core_config.Repository
		reqFactory *testreq.FakeReqFactory
		deps       command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.RepoLocator = deps.RepoLocator.SetRouteRepository(routeRepo)
		deps.Config = configRepo
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("delete-orphaned-routes").SetDependency(deps, pluginCall))
	}

	callDeleteOrphanedRoutes := func(confirmation string, args []string, reqFactory *testreq.FakeReqFactory, routeRepo *testapi.FakeRouteRepository) (*testterm.FakeUI, bool) {
		ui = &testterm.FakeUI{Inputs: []string{confirmation}}
		configRepo = testconfig.NewRepositoryWithDefaults()
		passed := testcmd.RunCliCommand("delete-orphaned-routes", args, reqFactory, updateCommandDependency, false)

		return ui, passed
	}

	BeforeEach(func() {
		routeRepo = &testapi.FakeRouteRepository{}
		reqFactory = &testreq.FakeReqFactory{}
	})

	It("fails requirements when not logged in", func() {
		_, passed := callDeleteOrphanedRoutes("y", []string{}, reqFactory, routeRepo)
		Expect(passed).To(BeFalse())
	})
	It("should fail with usage when provided any arguments", func() {
		reqFactory.LoginSuccess = true
		ui, passed := callDeleteOrphanedRoutes("y", []string{"blahblah"}, reqFactory, routeRepo)
		Expect(passed).To(BeFalse())
		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Incorrect Usage", "No argument required"},
		))
	})

	Context("when logged in successfully", func() {

		BeforeEach(func() {
			reqFactory.LoginSuccess = true
		})

		It("passes requirements when logged in", func() {
			_, passed := callDeleteOrphanedRoutes("y", []string{}, reqFactory, routeRepo)
			Expect(passed).To(BeTrue())
		})

		It("passes when confirmation is provided", func() {
			var ui *testterm.FakeUI
			domain := models.DomainFields{Name: "example.com"}
			domain2 := models.DomainFields{Name: "cookieclicker.co"}

			app1 := models.ApplicationFields{Name: "dora"}

			routeRepo.ListRoutesStub = func(cb func(models.Route) bool) error {
				route := models.Route{}
				route.Guid = "route1-guid"
				route.Host = "hostname-1"
				route.Domain = domain
				route.Apps = []models.ApplicationFields{app1}

				route2 := models.Route{}
				route2.Guid = "route2-guid"
				route2.Host = "hostname-2"
				route2.Domain = domain2

				cb(route)
				cb(route2)

				return nil
			}

			ui, _ = callDeleteOrphanedRoutes("y", []string{}, reqFactory, routeRepo)

			Expect(ui.Prompts).To(ContainSubstrings(
				[]string{"Really delete orphaned routes"},
			))

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Deleting route", "hostname-2.cookieclicker.co"},
				[]string{"OK"},
			))

			Expect(routeRepo.DeleteCallCount()).To(Equal(1))
			Expect(routeRepo.DeleteArgsForCall(0)).To(Equal("route2-guid"))
		})

		It("passes when the force flag is used", func() {
			var ui *testterm.FakeUI

			routeRepo.ListRoutesStub = func(cb func(models.Route) bool) error {
				route := models.Route{}
				route.Host = "hostname-1"
				route.Domain = models.DomainFields{Name: "example.com"}
				route.Apps = []models.ApplicationFields{
					{
						Name: "dora",
					},
				}

				route2 := models.Route{}
				route2.Guid = "route2-guid"
				route2.Host = "hostname-2"
				route2.Domain = models.DomainFields{Name: "cookieclicker.co"}

				cb(route)
				cb(route2)

				return nil
			}

			ui, _ = callDeleteOrphanedRoutes("", []string{"-f"}, reqFactory, routeRepo)

			Expect(len(ui.Prompts)).To(Equal(0))

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Deleting route", "hostname-2.cookieclicker.co"},
				[]string{"OK"},
			))
			Expect(routeRepo.DeleteCallCount()).To(Equal(1))
			Expect(routeRepo.DeleteArgsForCall(0)).To(Equal("route2-guid"))
		})
	})
})
