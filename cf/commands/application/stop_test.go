package application_test

import (
	testApplication "github.com/cloudfoundry/cli/cf/api/applications/fakes"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/commands/application"
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
		config              core_config.ReadWriter
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		config = testconfig.NewRepositoryWithDefaults()
		appRepo = &testApplication.FakeApplicationRepository{}
		requirementsFactory = &testreq.FakeReqFactory{}
	})

	runCommand := func(args ...string) {
		testcmd.RunCommand(NewStop(ui, config, appRepo), args, requirementsFactory)
	}

	It("fails requirements when not logged in", func() {
		runCommand("some-app-name")
		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})

	Context("when logged in and an app exists", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true

			app = models.Application{}
			app.Name = "my-app"
			app.Guid = "my-app-guid"
			app.State = "started"
		})

		JustBeforeEach(func() {
			appRepo.ReadReturns.App = app
			requirementsFactory.Application = app
		})

		It("fails with usage when the app name is not given", func() {
			runCommand()
			Expect(ui.FailedWithUsage).To(BeTrue())
		})

		It("stops the app with the given name", func() {
			runCommand("my-app")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Stopping app", "my-app", "my-org", "my-space", "my-user"},
				[]string{"OK"},
			))

			Expect(requirementsFactory.ApplicationName).To(Equal("my-app"))
			Expect(appRepo.UpdateAppGuid).To(Equal("my-app-guid"))
		})

		It("warns the user when stopping the app fails", func() {
			appRepo.UpdateErr = true
			runCommand("my-app")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Stopping", "my-app"},
				[]string{"FAILED"},
				[]string{"Error updating app."},
			))
			Expect(appRepo.UpdateAppGuid).To(Equal("my-app-guid"))
		})

		Context("when the app is stopped", func() {
			BeforeEach(func() {
				app.State = "stopped"
			})

			It("warns the user when the app is already stopped", func() {
				runCommand("my-app")

				Expect(ui.Outputs).To(ContainSubstrings([]string{"my-app", "is already stopped"}))
				Expect(appRepo.UpdateAppGuid).To(Equal(""))
			})
		})

		Describe(".ApplicationStop()", func() {
			It("returns the updated app model from ApplicationStop()", func() {
				expectedStoppedApp := app
				expectedStoppedApp.State = "stopped"

				appRepo.UpdateAppResult = expectedStoppedApp
				stopper := NewStop(ui, config, appRepo)
				actualStoppedApp, err := stopper.ApplicationStop(app, config.OrganizationFields().Name, config.SpaceFields().Name)

				Expect(err).NotTo(HaveOccurred())
				Expect(expectedStoppedApp).To(Equal(actualStoppedApp))
			})

			Context("When the app is already stopped", func() {
				BeforeEach(func() {
					app.State = "stopped"
				})

				It("returns the app without updating it", func() {
					stopper := NewStop(ui, config, appRepo)
					updatedApp, err := stopper.ApplicationStop(app, config.OrganizationFields().Name, config.SpaceFields().Name)

					Expect(err).NotTo(HaveOccurred())
					Expect(app).To(Equal(updatedApp))
				})
			})
		})
	})
})
