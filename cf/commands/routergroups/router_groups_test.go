package routergroups_test

import (
	"errors"

	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/flags"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	"github.com/cloudfoundry/cli/cf/commands/routergroups"
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

		Context("when arguments are provided", func() {
			var cmd command_registry.Command
			var flagContext flags.FlagContext

			BeforeEach(func() {
				cmd = &routergroups.RouterGroups{}
				cmd.SetDependency(deps, false)
				flagContext = flags.NewFlagContext(cmd.MetaData().Flags)
			})

			It("should fail with usage", func() {
				flagContext.Parse("blahblah")

				reqs := cmd.Requirements(requirementsFactory, flagContext)

				err := testcmd.RunRequirements(reqs)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Incorrect Usage"))
				Expect(err.Error()).To(ContainSubstring("No argument required"))
			})
		})
	})

	Context("when there are router groups", func() {
		BeforeEach(func() {
			routerGroups := models.RouterGroups{
				models.RouterGroup{
					Guid: "guid-0001",
					Name: "default-router-group",
					Type: "tcp",
				},
			}
			routingApiRepo.ListRouterGroupsStub = func(cb func(models.RouterGroup) bool) (apiErr error) {
				for _, r := range routerGroups {
					if !cb(r) {
						break
					}
				}
				return nil
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
			routingApiRepo.ListRouterGroupsReturns(errors.New("BOOM"))
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
