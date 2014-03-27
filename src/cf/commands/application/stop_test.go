package application_test

import (
	"cf/api"
	. "cf/commands/application"
	"cf/models"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

var _ = Describe("stop command", func() {
	var (
		requirementsFactory *testreq.FakeReqFactory
	)

	BeforeEach(func() {
		requirementsFactory = &testreq.FakeReqFactory{}
	})

	It("fails requirements when not logged in", func() {
		requirementsFactory.LoginSuccess = false
		appRepo := &testapi.FakeApplicationRepository{}
		cmd := NewStop(new(testterm.FakeUI), testconfig.NewRepository(), appRepo)
		testcmd.RunCommand(cmd, testcmd.NewContext("stop", []string{"some-app-name"}), requirementsFactory)

		Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
	})

	Context("when logged in", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
		})

		It("fails with usage when the app name is not given", func() {
			app := models.Application{}
			app.Name = "my-app"
			app.Guid = "my-app-guid"
			appRepo := &testapi.FakeApplicationRepository{ReadApp: app}

			ui := callStop([]string{}, requirementsFactory, appRepo)
			Expect(ui.FailedWithUsage).To(BeTrue())
		})

		It("stops the app with the given name", func() {
			app := models.Application{}
			app.Name = "my-app"
			app.Guid = "my-app-guid"
			appRepo := &testapi.FakeApplicationRepository{ReadApp: app}
			args := []string{"my-app"}
			requirementsFactory.Application = app
			ui := callStop(args, requirementsFactory, appRepo)

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"Stopping app", "my-app", "my-org", "my-space", "my-user"},
				{"OK"},
			})

			Expect(requirementsFactory.ApplicationName).To(Equal("my-app"))
			Expect(appRepo.UpdateAppGuid).To(Equal("my-app-guid"))
		})

		It("warns the user when stopping the app fails", func() {
			app := models.Application{}
			app.Name = "my-app"
			app.Guid = "my-app-guid"
			appRepo := &testapi.FakeApplicationRepository{ReadApp: app, UpdateErr: true}
			args := []string{"my-app"}
			requirementsFactory.Application = app
			ui := callStop(args, requirementsFactory, appRepo)

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"Stopping", "my-app"},
				{"FAILED"},
				{"Error updating app."},
			})
			Expect(appRepo.UpdateAppGuid).To(Equal("my-app-guid"))
		})

		It("warns the user when the app is already stopped", func() {
			app := models.Application{}
			app.Name = "my-app"
			app.Guid = "my-app-guid"
			app.State = "stopped"
			appRepo := &testapi.FakeApplicationRepository{ReadApp: app}
			args := []string{"my-app"}
			requirementsFactory.Application = app
			ui := callStop(args, requirementsFactory, appRepo)

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"my-app", "is already stopped"},
			})
			Expect(appRepo.UpdateAppGuid).To(Equal(""))
		})

		It("returns the updated app model from ApplicationStop()", func() {
			appToStop := models.Application{}
			appToStop.Name = "my-app"
			appToStop.Guid = "my-app-guid"
			appToStop.State = "started"
			expectedStoppedApp := models.Application{}
			expectedStoppedApp.Name = "my-stopped-app"
			expectedStoppedApp.Guid = "my-stopped-app-guid"
			expectedStoppedApp.State = "stopped"

			appRepo := &testapi.FakeApplicationRepository{UpdateAppResult: expectedStoppedApp}
			config := testconfig.NewRepository()
			stopper := NewStop(new(testterm.FakeUI), config, appRepo)
			actualStoppedApp, err := stopper.ApplicationStop(appToStop)

			Expect(err).NotTo(HaveOccurred())
			Expect(expectedStoppedApp).To(Equal(actualStoppedApp))
		})

		It("TestApplicationStopReturnsUpdatedAppWhenAppIsAlreadyStopped", func() {
			appToStop := models.Application{}
			appToStop.Name = "my-app"
			appToStop.Guid = "my-app-guid"
			appToStop.State = "stopped"
			appRepo := &testapi.FakeApplicationRepository{}
			config := testconfig.NewRepository()
			stopper := NewStop(new(testterm.FakeUI), config, appRepo)
			updatedApp, err := stopper.ApplicationStop(appToStop)

			Expect(err).NotTo(HaveOccurred())
			Expect(appToStop).To(Equal(updatedApp))
		})
	})
})

func callStop(args []string, requirementsFactory *testreq.FakeReqFactory, appRepo api.ApplicationRepository) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("stop", args)

	configRepo := testconfig.NewRepositoryWithDefaults()
	cmd := NewStop(ui, configRepo, appRepo)
	testcmd.RunCommand(cmd, ctxt, requirementsFactory)
	return
}
