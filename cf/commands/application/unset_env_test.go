package application_test

import (
	"errors"

	"code.cloudfoundry.org/cli/cf/api/applications/applicationsfakes"
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

var _ = Describe("unset-env command", func() {
	var (
		ui                  *testterm.FakeUI
		app                 models.Application
		appRepo             *applicationsfakes.FakeRepository
		configRepo          coreconfig.Repository
		requirementsFactory *requirementsfakes.FakeFactory
		deps                commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.Config = configRepo
		deps.RepoLocator = deps.RepoLocator.SetApplicationRepository(appRepo)
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("unset-env").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		app = models.Application{}
		app.Name = "my-app"
		app.GUID = "my-app-guid"
		appRepo = new(applicationsfakes.FakeRepository)
		requirementsFactory = new(requirementsfakes.FakeFactory)
		configRepo = testconfig.NewRepositoryWithDefaults()
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCLICommand("unset-env", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	Describe("requirements", func() {
		It("fails when not logged in", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Failing{Message: "not logged in"})

			Expect(runCommand("foo", "bar")).To(BeFalse())
		})

		It("fails when a space is not targeted", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
			requirementsFactory.NewTargetedSpaceRequirementReturns(requirements.Failing{Message: "not targeting space"})

			Expect(runCommand("foo", "bar")).To(BeFalse())
		})

		It("fails with usage when not provided with exactly 2 args", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
			requirementsFactory.NewTargetedSpaceRequirementReturns(requirements.Passing{})
			applicationReq := new(requirementsfakes.FakeApplicationRequirement)
			applicationReq.GetApplicationReturns(app)
			requirementsFactory.NewApplicationRequirementReturns(applicationReq)

			Expect(runCommand("too", "many", "args")).To(BeFalse())
		})
	})

	Context("when logged in, a space is targeted and provided enough args", func() {
		BeforeEach(func() {
			app.EnvironmentVars = map[string]interface{}{"foo": "bar", "DATABASE_URL": "mysql://example.com/my-db"}

			applicationReq := new(requirementsfakes.FakeApplicationRequirement)
			applicationReq.GetApplicationReturns(app)
			requirementsFactory.NewApplicationRequirementReturns(applicationReq)
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
			requirementsFactory.NewTargetedSpaceRequirementReturns(requirements.Passing{})
		})

		It("updates the app and tells the user what happened", func() {
			runCommand("my-app", "DATABASE_URL")

			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Removing env variable", "DATABASE_URL", "my-app", "my-org", "my-space", "my-user"},
				[]string{"OK"},
			))

			appGUID, params := appRepo.UpdateArgsForCall(0)
			Expect(appGUID).To(Equal("my-app-guid"))
			Expect(*params.EnvironmentVars).To(Equal(map[string]interface{}{
				"foo": "bar",
			}))
		})

		Context("when updating the app fails", func() {
			BeforeEach(func() {
				appRepo.UpdateReturns(models.Application{}, errors.New("Error updating app."))
				appRepo.ReadReturns(app, nil)
			})

			It("fails and alerts the user", func() {
				runCommand("does-not-exist", "DATABASE_URL")

				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Removing env variable"},
					[]string{"FAILED"},
					[]string{"Error updating app."},
				))
			})
		})

		It("tells the user if the specified env var was not set", func() {
			runCommand("my-app", "CANT_STOP_WONT_STOP_UNSETTIN_THIS_ENV")

			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Removing env variable"},
				[]string{"OK"},
				[]string{"CANT_STOP_WONT_STOP_UNSETTIN_THIS_ENV", "was not set."},
			))
		})
	})
})
