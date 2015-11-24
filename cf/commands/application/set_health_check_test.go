package application_test

import (
	"errors"

	testApplication "github.com/cloudfoundry/cli/cf/api/applications/fakes"
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

var _ = Describe("set-health-check command", func() {
	var (
		ui                  *testterm.FakeUI
		requirementsFactory *testreq.FakeReqFactory
		appRepo             *testApplication.FakeApplicationRepository
		configRepo          core_config.Repository
		deps                command_registry.Dependency
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{}
		appRepo = &testApplication.FakeApplicationRepository{}
	})

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.Config = configRepo
		deps.RepoLocator = deps.RepoLocator.SetApplicationRepository(appRepo)
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("set-health-check").SetDependency(deps, pluginCall))
	}

	runCommand := func(args ...string) bool {
		return testcmd.RunCliCommand("set-health-check", args, requirementsFactory, updateCommandDependency, false)
	}

	Describe("requirements", func() {
		It("fails with usage when called without enough arguments", func() {
			requirementsFactory.LoginSuccess = true

			runCommand("FAKE_APP")
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires", "argument"},
			))
		})

		It("fails with usage when health_check_type is not provided with 'none' or 'port'", func() {
			requirementsFactory.LoginSuccess = true

			runCommand("FAKE_APP", "bad_type")
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Incorrect Usage", "HEALTH_CHECK_TYPE", "port", "none"},
			))
		})

		It("fails requirements when not logged in", func() {
			Expect(runCommand("my-app", "none")).To(BeFalse())
		})

		It("fails if a space is not targeted", func() {
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedSpaceSuccess = false
			Expect(runCommand("my-app", "none")).To(BeFalse())
		})
	})

	Describe("setting health_check_type", func() {
		var (
			app models.Application
		)

		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedSpaceSuccess = true

			app = models.Application{}
			app.Name = "my-app"
			app.Guid = "my-app-guid"
			app.HealthCheckType = "none"

			requirementsFactory.Application = app
		})

		Context("when health_check_type is already set to the desired state", func() {
			It("notifies the user", func() {
				runCommand("my-app", "none")

				Ω(ui.Outputs).To(ContainSubstrings([]string{"my-app", "already set to 'none'"}))
			})
		})

		Context("Updating health_check_type when not already set to the desired state", func() {
			Context("Update successfully", func() {
				BeforeEach(func() {
					app = models.Application{}
					app.Name = "my-app"
					app.Guid = "my-app-guid"
					app.HealthCheckType = "port"

					appRepo.UpdateReturns(app, nil)
				})

				It("updates the app's health_check_type", func() {
					runCommand("my-app", "port")

					Expect(appRepo.UpdateCallCount()).To(Equal(1))
					appGUID, params := appRepo.UpdateArgsForCall(0)
					Ω(appGUID).To(Equal("my-app-guid"))
					Ω(*params.HealthCheckType).To(Equal("port"))
					Ω(ui.Outputs).To(ContainSubstrings([]string{"Updating", "my-app", "port"}))
					Ω(ui.Outputs).To(ContainSubstrings([]string{"OK"}))
				})
			})

			Context("Update fails", func() {
				It("notifies user of any api error", func() {
					appRepo.UpdateReturns(models.Application{}, errors.New("Error updating app."))
					runCommand("my-app", "port")

					Ω(appRepo.UpdateCallCount()).To(Equal(1))
					Ω(ui.Outputs).To(ContainSubstrings(
						[]string{"FAILED"},
						[]string{"Error updating app"},
					))
				})

				It("notifies user when updated result is not in the desired state", func() {
					app = models.Application{}
					app.Name = "my-app"
					app.Guid = "my-app-guid"
					app.HealthCheckType = "none"
					appRepo.UpdateReturns(app, nil)

					runCommand("my-app", "port")

					Ω(appRepo.UpdateCallCount()).To(Equal(1))
					Ω(ui.Outputs).To(ContainSubstrings(
						[]string{"FAILED"},
						[]string{"health_check_type", "not set"},
					))
				})
			})
		})
	})

})
