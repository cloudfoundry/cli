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

var _ = Describe("set-health-check command", func() {
	var (
		ui                  *testterm.FakeUI
		requirementsFactory *requirementsfakes.FakeFactory
		appRepo             *applicationsfakes.FakeRepository
		configRepo          coreconfig.Repository
		deps                commandregistry.Dependency
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = new(requirementsfakes.FakeFactory)
		appRepo = new(applicationsfakes.FakeRepository)
	})

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.Config = configRepo
		deps.RepoLocator = deps.RepoLocator.SetApplicationRepository(appRepo)
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("set-health-check").SetDependency(deps, pluginCall))
	}

	runCommand := func(args ...string) bool {
		return testcmd.RunCLICommand("set-health-check", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	Describe("requirements", func() {
		It("fails with usage when called without enough arguments", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})

			runCommand("FAKE_APP")
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires", "argument"},
			))
		})

		It("fails with usage when health_check_type is not provided with 'none' or 'port'", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})

			runCommand("FAKE_APP", "bad_type")
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Incorrect Usage", "HEALTH_CHECK_TYPE", "port", "none"},
			))
		})

		It("fails requirements when not logged in", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Failing{Message: "not logged in"})
			Expect(runCommand("my-app", "none")).To(BeFalse())
		})

		It("fails if a space is not targeted", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
			requirementsFactory.NewTargetedSpaceRequirementReturns(requirements.Failing{Message: "not targeting space"})
			Expect(runCommand("my-app", "none")).To(BeFalse())
		})
	})

	Describe("setting health_check_type", func() {
		var (
			app models.Application
		)

		BeforeEach(func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
			requirementsFactory.NewTargetedSpaceRequirementReturns(requirements.Passing{})

			app = models.Application{}
			app.Name = "my-app"
			app.GUID = "my-app-guid"
			app.HealthCheckType = "none"

			applicationReq := new(requirementsfakes.FakeApplicationRequirement)
			applicationReq.GetApplicationReturns(app)
			requirementsFactory.NewApplicationRequirementReturns(applicationReq)
		})

		Context("when health_check_type is already set to the desired state", func() {
			It("notifies the user", func() {
				runCommand("my-app", "none")

				Expect(ui.Outputs()).To(ContainSubstrings([]string{"my-app", "already set to 'none'"}))
			})
		})

		Context("Updating health_check_type when not already set to the desired state", func() {
			Context("Update successfully", func() {
				BeforeEach(func() {
					app = models.Application{}
					app.Name = "my-app"
					app.GUID = "my-app-guid"
					app.HealthCheckType = "port"

					appRepo.UpdateReturns(app, nil)
				})

				It("updates the app's health_check_type", func() {
					runCommand("my-app", "port")

					Expect(appRepo.UpdateCallCount()).To(Equal(1))
					appGUID, params := appRepo.UpdateArgsForCall(0)
					Expect(appGUID).To(Equal("my-app-guid"))
					Expect(*params.HealthCheckType).To(Equal("port"))
					Expect(ui.Outputs()).To(ContainSubstrings([]string{"Updating", "my-app", "port"}))
					Expect(ui.Outputs()).To(ContainSubstrings([]string{"OK"}))
				})
			})

			Context("Update fails", func() {
				It("notifies user of any api error", func() {
					appRepo.UpdateReturns(models.Application{}, errors.New("Error updating app."))
					runCommand("my-app", "port")

					Expect(appRepo.UpdateCallCount()).To(Equal(1))
					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"FAILED"},
						[]string{"Error updating app"},
					))

				})

				It("notifies user when updated result is not in the desired state", func() {
					app = models.Application{}
					app.Name = "my-app"
					app.GUID = "my-app-guid"
					app.HealthCheckType = "none"
					appRepo.UpdateReturns(app, nil)

					runCommand("my-app", "port")

					Expect(appRepo.UpdateCallCount()).To(Equal(1))
					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"FAILED"},
						[]string{"health_check_type", "not set"},
					))

				})
			})
		})
	})

})
