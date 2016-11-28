package application_test

import (
	"errors"

	"code.cloudfoundry.org/cli/cf/api/applications/applicationsfakes"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"
	testcmd "code.cloudfoundry.org/cli/util/testhelpers/commands"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"

	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/commands/application"
	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("stop command", func() {
	var (
		ui                  *testterm.FakeUI
		app                 models.Application
		appRepo             *applicationsfakes.FakeRepository
		requirementsFactory *requirementsfakes.FakeFactory
		config              coreconfig.Repository
		deps                commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.RepoLocator = deps.RepoLocator.SetApplicationRepository(appRepo)
		deps.Config = config
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("stop").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		config = testconfig.NewRepositoryWithDefaults()
		appRepo = new(applicationsfakes.FakeRepository)
		requirementsFactory = new(requirementsfakes.FakeFactory)
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCLICommand("stop", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	It("fails requirements when not logged in", func() {
		requirementsFactory.NewLoginRequirementReturns(requirements.Failing{Message: "not logged in"})
		Expect(runCommand("some-app-name")).To(BeFalse())
	})

	It("fails requirements when a space is not targeted", func() {
		requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
		requirementsFactory.NewTargetedSpaceRequirementReturns(requirements.Failing{Message: "not targeting space"})
		Expect(runCommand("some-app-name")).To(BeFalse())
	})

	Context("when logged in and an app exists", func() {
		BeforeEach(func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
			requirementsFactory.NewTargetedSpaceRequirementReturns(requirements.Passing{})

			app = models.Application{}
			app.Name = "my-app"
			app.GUID = "my-app-guid"
			app.State = "started"
		})

		JustBeforeEach(func() {
			appRepo.ReadReturns(app, nil)
			applicationReq := new(requirementsfakes.FakeApplicationRequirement)
			applicationReq.GetApplicationReturns(app)
			requirementsFactory.NewApplicationRequirementReturns(applicationReq)
		})

		It("fails with usage when the app name is not given", func() {
			runCommand()
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires an argument"},
			))
		})

		It("stops the app with the given name", func() {
			runCommand("my-app")

			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Stopping app", "my-app", "my-org", "my-space", "my-user"},
				[]string{"OK"},
			))

			appGUID, _ := appRepo.UpdateArgsForCall(0)
			Expect(appGUID).To(Equal("my-app-guid"))
		})

		It("warns the user when stopping the app fails", func() {
			appRepo.UpdateReturns(models.Application{}, errors.New("Error updating app."))
			runCommand("my-app")

			Expect(ui.Outputs()).To(ContainSubstrings(
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

				Expect(ui.Outputs()).To(ContainSubstrings([]string{"my-app", "is already stopped"}))
				Expect(appRepo.UpdateCallCount()).To(BeZero())
			})
		})

		Describe(".ApplicationStop()", func() {
			It("returns the updated app model from ApplicationStop()", func() {
				expectedStoppedApp := app
				expectedStoppedApp.State = "stopped"

				appRepo.UpdateReturns(expectedStoppedApp, nil)
				updateCommandDependency(false)
				stopper := commandregistry.Commands.FindCommand("stop").(*application.Stop)
				actualStoppedApp, err := stopper.ApplicationStop(app, config.OrganizationFields().Name, config.SpaceFields().Name)

				Expect(err).NotTo(HaveOccurred())
				Expect(expectedStoppedApp).To(Equal(actualStoppedApp))
			})

			Context("When the app is already stopped", func() {
				BeforeEach(func() {
					app.State = "stopped"
				})

				It("returns the app without updating it", func() {
					stopper := commandregistry.Commands.FindCommand("stop").(*application.Stop)
					updatedApp, err := stopper.ApplicationStop(app, config.OrganizationFields().Name, config.SpaceFields().Name)

					Expect(err).NotTo(HaveOccurred())
					Expect(app).To(Equal(updatedApp))
				})
			})
		})
	})
})
