package route_test

import (
	"code.cloudfoundry.org/cli/cf/api/spaces/spacesfakes"
	"code.cloudfoundry.org/cli/cf/flags"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"
	testcmd "code.cloudfoundry.org/cli/cf/util/testhelpers/commands"
	testconfig "code.cloudfoundry.org/cli/cf/util/testhelpers/configuration"
	testterm "code.cloudfoundry.org/cli/cf/util/testhelpers/terminal"

	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/commands/route"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("delete-orphaned-routes command", func() {
	var (
		ui                  *testterm.FakeUI
		spaceRepo           *spacesfakes.FakeSpaceRepository
		configRepo          coreconfig.Repository
		requirementsFactory *requirementsfakes.FakeFactory
		deps                commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.RepoLocator = deps.RepoLocator.SetSpaceRepository(spaceRepo)
		deps.Config = configRepo
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("delete-orphaned-routes").SetDependency(deps, pluginCall))
	}

	callDeleteOrphanedRoutes := func(confirmation string, args []string, requirementsFactory *requirementsfakes.FakeFactory, spaceRepo *spacesfakes.FakeSpaceRepository) (*testterm.FakeUI, bool) {
		ui = &testterm.FakeUI{Inputs: []string{confirmation}}
		configRepo = testconfig.NewRepositoryWithDefaults()
		passed := testcmd.RunCLICommand("delete-orphaned-routes", args, requirementsFactory, updateCommandDependency, false, ui)

		return ui, passed
	}

	BeforeEach(func() {
		spaceRepo = new(spacesfakes.FakeSpaceRepository)
		requirementsFactory = new(requirementsfakes.FakeFactory)
	})

	It("fails requirements when not logged in", func() {
		requirementsFactory.NewLoginRequirementReturns(requirements.Failing{Message: "not logged in"})
		_, passed := callDeleteOrphanedRoutes("y", []string{}, requirementsFactory, spaceRepo)
		Expect(passed).To(BeFalse())
	})

	Context("when arguments are provided", func() {
		var cmd commandregistry.Command
		var flagContext flags.FlagContext

		BeforeEach(func() {
			cmd = &route.DeleteOrphanedRoutes{}
			cmd.SetDependency(deps, false)
			flagContext = flags.NewFlagContext(cmd.MetaData().Flags)
		})

		It("should fail with usage", func() {
			flagContext.Parse("blahblah")

			reqs, err := cmd.Requirements(requirementsFactory, flagContext)
			Expect(err).NotTo(HaveOccurred())

			err = testcmd.RunRequirements(reqs)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Incorrect Usage"))
			Expect(err.Error()).To(ContainSubstring("No argument required"))
		})
	})

	Context("when logged in successfully", func() {

		BeforeEach(func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
		})

		It("passes requirements when logged in", func() {
			_, passed := callDeleteOrphanedRoutes("y", []string{}, requirementsFactory, spaceRepo)
			Expect(passed).To(BeTrue())
		})

		It("passes when confirmation is provided", func() {
			var ui *testterm.FakeUI

			ui, _ = callDeleteOrphanedRoutes("y", []string{}, requirementsFactory, spaceRepo)

			Expect(ui.Prompts).To(ContainSubstrings(
				[]string{"Really delete orphaned routes"},
			))

			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"OK"},
			))

			Expect(spaceRepo.DeleteUnmappedRoutesCallCount()).To(Equal(1))
		})

		It("passes when the force flag is used", func() {
			var ui *testterm.FakeUI

			ui, _ = callDeleteOrphanedRoutes("", []string{"-f"}, requirementsFactory, spaceRepo)

			Expect(len(ui.Prompts)).To(Equal(0))

			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"OK"},
			))
			Expect(spaceRepo.DeleteUnmappedRoutesCallCount()).To(Equal(1))
		})
	})
})
