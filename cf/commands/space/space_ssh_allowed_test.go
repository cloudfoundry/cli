package space_test

import (
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("space-ssh-allowed command", func() {
	var (
		ui                  *testterm.FakeUI
		requirementsFactory *testreq.FakeReqFactory
		// configRepo          core_config.Repository
		deps command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		// deps.Config = configRepo
		// deps.RepoLocator = deps.RepoLocator.SetSpaceRepository(spaceRepo)
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("space-ssh-allowed").SetDependency(deps, pluginCall))
	}

	runCommand := func(args ...string) bool {
		return testcmd.RunCliCommand("space-ssh-allowed", args, requirementsFactory, updateCommandDependency, false)
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		// configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{}
		// spaceRepo = &testapi.FakeSpaceRepository{}
	})

	Describe("requirements", func() {
		It("fails with usage when called without enough arguments", func() {
			requirementsFactory.LoginSuccess = true

			runCommand()
			Ω(ui.Outputs).Should(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires", "argument"},
			))
		})

		It("fails requirements when not logged in", func() {
			Ω(runCommand("my-space")).To(BeFalse())
		})

		It("does not pass requirements if org is not targeted", func() {
			requirementsFactory.LoginSuccess = true
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

	Describe("space-ssh-allowed", func() {
		var space models.Space

		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedOrgSuccess = true

			space = models.Space{}
			space.Name = "the-space-name"
			space.Guid = "the-space-guid"
		})

		Context("when SSH is enabled for the space", func() {
			It("notifies the user", func() {
				space.AllowSSH = true
				requirementsFactory.Space = space

				runCommand("the-space-name")

				Ω(ui.Outputs).To(ContainSubstrings([]string{"ssh support is enabled in space 'the-space-name'"}))
			})
		})

		Context("when SSH is disabled for the space", func() {
			It("notifies the user", func() {
				requirementsFactory.Space = space

				runCommand("the-space-name")

				Ω(ui.Outputs).To(ContainSubstrings([]string{"ssh support is disabled in space 'the-space-name'"}))
			})
		})
	})
})
