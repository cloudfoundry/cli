package application_test

import (
	"errors"

	testApplication "github.com/cloudfoundry/cli/cf/api/applications/fakes"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/commands/application"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("stop command", func() {
	var (
		ui                  *testterm.FakeUI
		app                 models.Application
		appRepo             *testApplication.FakeApplicationRepository
		requirementsFactory *testreq.FakeReqFactory
		config              core_config.Repository
		deps                command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.RepoLocator = deps.RepoLocator.SetApplicationRepository(appRepo)
		deps.Config = config
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("stop").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		config = testconfig.NewRepositoryWithDefaults()
		appRepo = &testApplication.FakeApplicationRepository{}
		requirementsFactory = &testreq.FakeReqFactory{}
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCliCommand("stop", args, requirementsFactory, updateCommandDependency, false)
	}

	It("fails requirements when not logged in", func() {
		Expect(runCommand("some-app-name")).To(BeFalse())
	})
	It("fails requirements when a space is not targeted", func() {
		requirementsFactory.LoginSuccess = true
		requirementsFactory.TargetedSpaceSuccess = false
		Expect(runCommand("some-app-name")).To(BeFalse())
	})

	Context("when logged in and an app exists", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedSpaceSuccess = true

			app = models.Application{}
			app.Name = "my-app"
			app.Guid = "my-app-guid"
			app.State = "started"
		})

		JustBeforeEach(func() {
			appRepo.ReadReturns(app, nil)
			requirementsFactory.Application = app
		})

		It("fails with usage when the app name is not given", func() {
			runCommand()
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires an argument"},
			))
		})

		It("stops the app with the given name", func() {
			runCommand("my-app")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Stopping app", "my-app", "my-org", "my-space", "my-user"},
				[]string{"OK"},
			))

			Expect(requirementsFactory.ApplicationName).To(Equal("my-app"))
			appGUID, _ := appRepo.UpdateArgsForCall(0)
			Expect(appGUID).To(Equal("my-app-guid"))
		})

		It("warns the user when stopping the app fails", func() {
			appRepo.UpdateReturns(models.Application{}, errors.New("Error updating app."))
			runCommand("my-app")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Stopping", "my-app"},
				[]string{"FAILED"},
				[]string{"Error updating app."},
			))
			appGUID, _ := appRepo.UpdateArgsForCall(0)
			Expect(appGUID).To(Equal("my-app-guid"))
		})

		Context("when the app is stopped", func() {
			BeforeEach(func() {
				app.State = "stopped"
			})

			It("warns the user when the app is already stopped", func() {
				runCommand("my-app")

				Expect(ui.Outputs).To(ContainSubstrings([]string{"my-app", "is already stopped"}))
				Expect(appRepo.UpdateCallCount()).To(BeZero())
			})
		})

		Describe(".ApplicationStop()", func() {
			It("returns the updated app model from ApplicationStop()", func() {
				expectedStoppedApp := app
				expectedStoppedApp.State = "stopped"

				appRepo.UpdateReturns(expectedStoppedApp, nil)
				updateCommandDependency(false)
				stopper := command_registry.Commands.FindCommand("stop").(*application.Stop)
				actualStoppedApp, err := stopper.ApplicationStop(app, config.OrganizationFields().Name, config.SpaceFields().Name)

				Expect(err).NotTo(HaveOccurred())
				Expect(expectedStoppedApp).To(Equal(actualStoppedApp))
			})

			Context("When the app is already stopped", func() {
				BeforeEach(func() {
					app.State = "stopped"
				})

				It("returns the app without updating it", func() {
					stopper := command_registry.Commands.FindCommand("stop").(*application.Stop)
					updatedApp, err := stopper.ApplicationStop(app, config.OrganizationFields().Name, config.SpaceFields().Name)

					Expect(err).NotTo(HaveOccurred())
					Expect(app).To(Equal(updatedApp))
				})
			})
		})
	})
})
