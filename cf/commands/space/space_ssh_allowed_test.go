package space_test

import (
	"errors"

	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"
	testcmd "code.cloudfoundry.org/cli/util/testhelpers/commands"
	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("space-ssh-allowed command", func() {
	var (
		ui                  *testterm.FakeUI
		requirementsFactory *requirementsfakes.FakeFactory
		deps                commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("space-ssh-allowed").SetDependency(deps, pluginCall))
	}

	runCommand := func(args ...string) bool {
		return testcmd.RunCLICommand("space-ssh-allowed", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		requirementsFactory = new(requirementsfakes.FakeFactory)
	})

	Describe("requirements", func() {
		It("fails with usage when called without enough arguments", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})

			runCommand()
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires", "argument"},
			))

		})

		It("fails requirements when not logged in", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Failing{Message: "not logged in"})
			Expect(runCommand("my-space")).To(BeFalse())
		})

		It("does not pass requirements if org is not targeted", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
			targetedOrgReq := new(requirementsfakes.FakeTargetedOrgRequirement)
			targetedOrgReq.ExecuteReturns(errors.New("no org targeted"))
			requirementsFactory.NewTargetedOrgRequirementReturns(targetedOrgReq)

			Expect(runCommand("my-space")).To(BeFalse())
		})

		It("does not pass requirements if space does not exist", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
			requirementsFactory.NewTargetedOrgRequirementReturns(new(requirementsfakes.FakeTargetedOrgRequirement))
			spaceReq := new(requirementsfakes.FakeSpaceRequirement)
			spaceReq.ExecuteReturns(errors.New("no space"))
			requirementsFactory.NewSpaceRequirementReturns(spaceReq)

			Expect(runCommand("my-space")).To(BeFalse())
		})
	})

	Describe("space-ssh-allowed", func() {
		var space models.Space

		BeforeEach(func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
			requirementsFactory.NewTargetedOrgRequirementReturns(new(requirementsfakes.FakeTargetedOrgRequirement))

			space = models.Space{}
			space.Name = "the-space-name"
			space.GUID = "the-space-guid"
			spaceReq := new(requirementsfakes.FakeSpaceRequirement)
			spaceReq.GetSpaceReturns(space)
			requirementsFactory.NewSpaceRequirementReturns(spaceReq)
		})

		Context("when SSH is enabled for the space", func() {
			It("notifies the user", func() {
				space.AllowSSH = true
				spaceReq := new(requirementsfakes.FakeSpaceRequirement)
				spaceReq.GetSpaceReturns(space)
				requirementsFactory.NewSpaceRequirementReturns(spaceReq)

				runCommand("the-space-name")

				Expect(ui.Outputs()).To(ContainSubstrings([]string{"ssh support is enabled in space 'the-space-name'"}))
			})
		})

		Context("when SSH is disabled for the space", func() {
			It("notifies the user", func() {
				runCommand("the-space-name")

				Expect(ui.Outputs()).To(ContainSubstrings([]string{"ssh support is disabled in space 'the-space-name'"}))
			})
		})
	})
})
