package commands_test

import (
	testapi "github.com/cloudfoundry/cli/cf/api/stacks/fakes"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/flags"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/cli/cf/commands"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

var _ = Describe("stacks command", func() {
	var (
		ui                  *testterm.FakeUI
		repo                *testapi.FakeStackRepository
		config              core_config.Repository
		requirementsFactory *testreq.FakeReqFactory
		deps                command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.Config = config
		deps.RepoLocator = deps.RepoLocator.SetStackRepository(repo)
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("stacks").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		config = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true}
		repo = &testapi.FakeStackRepository{}
	})

	Describe("login requirements", func() {
		It("fails if the user is not logged in", func() {
			requirementsFactory.LoginSuccess = false

			Expect(testcmd.RunCliCommand("stacks", []string{}, requirementsFactory, updateCommandDependency, false)).To(BeFalse())
		})

		Context("when arguments are provided", func() {
			var cmd command_registry.Command
			var flagContext flags.FlagContext

			BeforeEach(func() {
				cmd = &commands.ListStacks{}
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

	It("lists the stacks", func() {
		stack1 := models.Stack{
			Name:        "Stack-1",
			Description: "Stack 1 Description",
		}
		stack2 := models.Stack{
			Name:        "Stack-2",
			Description: "Stack 2 Description",
		}

		repo.FindAllReturns([]models.Stack{stack1, stack2}, nil)
		testcmd.RunCliCommand("stacks", []string{}, requirementsFactory, updateCommandDependency, false)

		Expect(ui.Outputs).To(ContainSubstrings(
			[]string{"Getting stacks in org", "my-org", "my-space", "my-user"},
			[]string{"OK"},
			[]string{"Stack-1", "Stack 1 Description"},
			[]string{"Stack-2", "Stack 2 Description"},
		))
	})
})
