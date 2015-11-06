package routergroups_test

import (
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("RouterGroups", func() {

	var (
		ui                  *testterm.FakeUI
		routingApiRepo      *testapi.FakeRoutingApiRepository
		configRepo          core_config.Repository
		requirementsFactory *testreq.FakeReqFactory
		deps                command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.RepoLocator = deps.RepoLocator.SetRoutingApiRepository(routingApiRepo)
		deps.Config = configRepo
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("router-groups").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{
			LoginSuccess:              true,
			RoutingAPIEndpointSuccess: true,
		}
		routingApiRepo = &testapi.FakeRoutingApiRepository{}
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCliCommand("router-groups", args, requirementsFactory, updateCommandDependency, false)
	}

	Describe("login requirements", func() {
		It("fails if the user is not logged in", func() {
			requirementsFactory.LoginSuccess = false
			Expect(runCommand()).To(BeFalse())
		})

		It("fails when the routing API endpoint is not set", func() {
			requirementsFactory.RoutingAPIEndpointSuccess = false
			Expect(runCommand()).To(BeFalse())
		})

		It("should fail with usage when provided any arguments", func() {
			Expect(runCommand("notrequired-option")).To(BeFalse())
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Incorrect Usage", "No argument required"},
			))
		})
	})

	Context("when there are router groups", func() {
		BeforeEach(func() {
			routingApiRepo.RouterGroups = models.RouterGroups{
				models.RouterGroup{
					Guid: "guid-0001",
					Name: "default-router-group",
					Type: "tcp",
				},
			}
		})

		It("lists router groups", func() {
			runCommand()

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Getting router groups", "my-user"},
				[]string{"name", "type"},
				[]string{"default-router-group", "tcp"},
			))
		})
	})

	Context("when there are no router groups", func() {
		It("tells the user when no router groups were found", func() {
			runCommand()

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Getting router groups"},
				[]string{"No router groups found"},
			))
		})
	})

	Context("when there is an error listing router groups", func() {
		BeforeEach(func() {
			routingApiRepo.ListError = true
		})

		It("returns an error to the user", func() {
			runCommand()

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Getting router groups"},
				[]string{"Failed fetching router groups"},
			))
		})
	})

})
