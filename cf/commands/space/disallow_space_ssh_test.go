package space_test

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

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("disallow-space-ssh command", func() {
	var (
		ui                  *testterm.FakeUI
		requirementsFactory *testreq.FakeReqFactory
		spaceRepo           *testapi.FakeSpaceRepository
		configRepo          core_config.Repository
		deps                command_registry.Dependency
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{}
		spaceRepo = &testapi.FakeSpaceRepository{}
	})

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.Config = configRepo
		deps.RepoLocator = deps.RepoLocator.SetSpaceRepository(spaceRepo)
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("disallow-space-ssh").SetDependency(deps, pluginCall))
	}

	runCommand := func(args ...string) bool {
		return testcmd.RunCliCommand("disallow-space-ssh", args, requirementsFactory, updateCommandDependency, false)
	}

	Describe("requirements", func() {
		It("fails with usage when called without enough arguments", func() {
			requirementsFactory.LoginSuccess = true

			runCommand()
			Ω(ui.Outputs).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires", "argument"},
			))
		})

		It("fails requirements when not logged in", func() {
			Ω(runCommand("my-space")).To(BeFalse())
		})

		It("does not pass requirements if org is not targeted", func() {
			requirementsFactory.TargetedOrgSuccess = false

			Expect(runCommand("my-space")).To(BeFalse())
		})

		It("does not pass requirements if space does not exist", func() {
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedOrgSuccess = true
			requirementsFactory.SpaceRequirementFails = true

			Expect(runCommand("my-space")).To(BeFalse())
		})
	})

	Describe("disallow-space-ssh", func() {
		var space models.Space

		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedOrgSuccess = true

			space = models.Space{}
			space.Name = "the-space-name"
			space.Guid = "the-space-guid"
		})

		Context("when allow_ssh is already set to the false", func() {
			BeforeEach(func() {
				space.AllowSSH = false
				requirementsFactory.Space = space
			})

			It("notifies the user", func() {
				runCommand("the-space-name")

				Ω(ui.Outputs).To(ContainSubstrings([]string{"ssh support is already disabled in space 'the-space-name'"}))
			})
		})

		Context("Updating allow_ssh when not already set to false", func() {
			Context("Update successfully", func() {
				BeforeEach(func() {
					space.AllowSSH = true
					requirementsFactory.Space = space
				})

				It("updates the space's allow_ssh", func() {
					runCommand("the-space-name")

					Expect(spaceRepo.SetAllowSSHCallCount()).To(Equal(1))
					spaceGUID, allow := spaceRepo.SetAllowSSHArgsForCall(0)
					Expect(spaceGUID).To(Equal("the-space-guid"))
					Expect(allow).To(Equal(false))
					Ω(ui.Outputs).To(ContainSubstrings([]string{"Disabling ssh support for space 'the-space-name'"}))
					Ω(ui.Outputs).To(ContainSubstrings([]string{"OK"}))
				})
			})

			Context("Update fails", func() {
				BeforeEach(func() {
					space.AllowSSH = true
					requirementsFactory.Space = space
				})

				It("notifies user of any api error", func() {
					spaceRepo.SetAllowSSHReturns(errors.New("api error"))
					runCommand("the-space-name")

					Ω(ui.Outputs).To(ContainSubstrings(
						[]string{"FAILED"},
						[]string{"Error", "api error"},
					))
				})
			})

		})
	})

})
