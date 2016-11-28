package space_test

import (
	"errors"

	"code.cloudfoundry.org/cli/cf/api/spaces/spacesfakes"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"
	testcmd "code.cloudfoundry.org/cli/util/testhelpers/commands"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"

	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("allow-space-ssh command", func() {
	var (
		ui                  *testterm.FakeUI
		requirementsFactory *requirementsfakes.FakeFactory
		spaceRepo           *spacesfakes.FakeSpaceRepository
		configRepo          coreconfig.Repository
		deps                commandregistry.Dependency
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = new(requirementsfakes.FakeFactory)
		spaceRepo = new(spacesfakes.FakeSpaceRepository)
	})

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.Config = configRepo
		deps.RepoLocator = deps.RepoLocator.SetSpaceRepository(spaceRepo)
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("allow-space-ssh").SetDependency(deps, pluginCall))
	}

	runCommand := func(args ...string) bool {
		return testcmd.RunCLICommand("allow-space-ssh", args, requirementsFactory, updateCommandDependency, false, ui)
	}

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

	Describe("allow-space-ssh", func() {
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

		Context("when allow_ssh is already set to the true", func() {
			BeforeEach(func() {
				space.AllowSSH = true
				spaceReq := new(requirementsfakes.FakeSpaceRequirement)
				spaceReq.GetSpaceReturns(space)
				requirementsFactory.NewSpaceRequirementReturns(spaceReq)
			})

			It("notifies the user", func() {
				runCommand("the-space-name")

				Expect(ui.Outputs()).To(ContainSubstrings([]string{"ssh support is already enabled in space 'the-space-name'"}))
			})
		})

		Context("Updating allow_ssh when not already set to true", func() {
			Context("Update successfully", func() {
				BeforeEach(func() {
					space.AllowSSH = false
					spaceReq := new(requirementsfakes.FakeSpaceRequirement)
					spaceReq.GetSpaceReturns(space)
					requirementsFactory.NewSpaceRequirementReturns(spaceReq)
				})

				It("updates the space's allow_ssh", func() {
					runCommand("the-space-name")

					Expect(spaceRepo.SetAllowSSHCallCount()).To(Equal(1))
					spaceGUID, allow := spaceRepo.SetAllowSSHArgsForCall(0)
					Expect(spaceGUID).To(Equal("the-space-guid"))
					Expect(allow).To(Equal(true))
					Expect(ui.Outputs()).To(ContainSubstrings([]string{"Enabling ssh support for space 'the-space-name'"}))
					Expect(ui.Outputs()).To(ContainSubstrings([]string{"OK"}))
				})
			})

			Context("Update fails", func() {
				It("notifies user of any api error", func() {
					spaceRepo.SetAllowSSHReturns(errors.New("api error"))
					runCommand("the-space-name")

					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"FAILED"},
						[]string{"Error", "api error"},
					))

				})
			})

		})
	})

})
