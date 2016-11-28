package application_test

import (
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

var _ = Describe("disable-ssh command", func() {
	var (
		ui                  *testterm.FakeUI
		requirementsFactory *requirementsfakes.FakeFactory
		configRepo          coreconfig.Repository
		deps                commandregistry.Dependency
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = new(requirementsfakes.FakeFactory)
	})

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.Config = configRepo
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("ssh-enabled").SetDependency(deps, pluginCall))
	}

	runCommand := func(args ...string) bool {
		return testcmd.RunCLICommand("ssh-enabled", args, requirementsFactory, updateCommandDependency, false, ui)
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
			Expect(runCommand("my-app", "none")).To(BeFalse())
		})

		It("fails if a space is not targeted", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
			requirementsFactory.NewTargetedSpaceRequirementReturns(requirements.Failing{Message: "not targeting space"})
			Expect(runCommand("my-app", "none")).To(BeFalse())
		})
	})

	Describe("ssh-enabled", func() {
		var (
			app models.Application
		)

		BeforeEach(func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
			requirementsFactory.NewTargetedSpaceRequirementReturns(requirements.Passing{})

			app = models.Application{}
			app.Name = "my-app"
			app.GUID = "my-app-guid"
		})

		Context("when enable_ssh is set to the true", func() {
			BeforeEach(func() {
				app.EnableSSH = true
				applicationReq := new(requirementsfakes.FakeApplicationRequirement)
				applicationReq.GetApplicationReturns(app)
				requirementsFactory.NewApplicationRequirementReturns(applicationReq)
			})

			It("notifies the user", func() {
				runCommand("my-app")

				Expect(ui.Outputs()).To(ContainSubstrings([]string{"ssh support is enabled for 'my-app'"}))
			})
		})

		Context("when enable_ssh is set to the false", func() {
			BeforeEach(func() {
				app.EnableSSH = false
				applicationReq := new(requirementsfakes.FakeApplicationRequirement)
				applicationReq.GetApplicationReturns(app)
				requirementsFactory.NewApplicationRequirementReturns(applicationReq)
			})

			It("notifies the user", func() {
				runCommand("my-app")

				Expect(ui.Outputs()).To(ContainSubstrings([]string{"ssh support is disabled for 'my-app'"}))
			})
		})

	})

})
