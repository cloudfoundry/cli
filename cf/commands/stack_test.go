package commands_test

import (
	"errors"

	"code.cloudfoundry.org/cli/cf/api/stacks/stacksfakes"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"
	testcmd "code.cloudfoundry.org/cli/util/testhelpers/commands"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
)

var _ = Describe("stack command", func() {
	var (
		ui                  *testterm.FakeUI
		config              coreconfig.Repository
		repo                *stacksfakes.FakeStackRepository
		requirementsFactory *requirementsfakes.FakeFactory
		deps                commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.Config = config
		deps.RepoLocator = deps.RepoLocator.SetStackRepository(repo)
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("stack").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		config = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = new(requirementsfakes.FakeFactory)
		requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
		repo = new(stacksfakes.FakeStackRepository)
	})

	Describe("login requirements", func() {
		It("fails if the user is not logged in", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Failing{Message: "not logged in"})

			Expect(testcmd.RunCLICommand("stack", []string{}, requirementsFactory, updateCommandDependency, false, ui)).To(BeFalse())
		})

		It("fails with usage when not provided exactly one arg", func() {
			Expect(testcmd.RunCLICommand("stack", []string{}, requirementsFactory, updateCommandDependency, false, ui)).To(BeFalse())
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"FAILED"},
				[]string{"Incorrect Usage.", "Requires stack name as argument"},
			))
		})
	})

	It("returns the stack guid when '--guid' flag is provided", func() {
		stack1 := models.Stack{
			Name:        "Stack-1",
			Description: "Stack 1 Description",
			GUID:        "Stack-1-GUID",
		}

		repo.FindByNameReturns(stack1, nil)

		testcmd.RunCLICommand("stack", []string{"Stack-1", "--guid"}, requirementsFactory, updateCommandDependency, false, ui)

		Expect(len(ui.Outputs())).To(Equal(1))
		Expect(ui.Outputs()[0]).To(Equal("Stack-1-GUID"))
	})

	It("returns the empty string as guid when '--guid' flag is provided and stack doesn't exist", func() {
		stack1 := models.Stack{
			Name:        "Stack-1",
			Description: "Stack 1 Description",
			GUID:        "Stack-1-GUID",
		}

		repo.FindByNameReturns(stack1, nil)

		testcmd.RunCLICommand("stack", []string{"Stack-1", "--guid"}, requirementsFactory, updateCommandDependency, false, ui)

		Expect(len(ui.Outputs())).To(Equal(1))
		Expect(ui.Outputs()[0]).To(Equal("Stack-1-GUID"))
	})

	It("lists the stack requested", func() {
		repo.FindByNameReturns(models.Stack{}, errors.New("Stack Stack-1 not found"))

		testcmd.RunCLICommand("stack", []string{"Stack-1", "--guid"}, requirementsFactory, updateCommandDependency, false, ui)

		Expect(len(ui.Outputs())).To(Equal(1))
		Expect(ui.Outputs()[0]).To(Equal(""))
	})

	It("informs user if stack is not found", func() {
		repo.FindByNameReturns(models.Stack{}, errors.New("Stack Stack-1 not found"))

		testcmd.RunCLICommand("stack", []string{"Stack-1"}, requirementsFactory, updateCommandDependency, false, ui)

		Expect(ui.Outputs()).To(BeInDisplayOrder(
			[]string{"FAILED"},
			[]string{"Stack Stack-1 not found"},
		))
	})
})
