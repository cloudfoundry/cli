/*
                       WARNING WARNING WARNING

                Attention all potential contributors

   This testfile is not in the best state. We've been slowly transitioning
   from the built in "testing" package to using Ginkgo. As you can see, we've
   changed the format, but a lot of the setup, test body, descriptions, etc
   are either hardcoded, completely lacking, or misleading.

   For example:

   Describe("Testing with ginkgo"...)      // This is not a great description
   It("TestDoesSoemthing"...)              // This is a horrible description

   Describe("create-user command"...       // Describe the actual object under test
   It("creates a user when provided ..."   // this is more descriptive

   For good examples of writing Ginkgo tests for the cli, refer to

   src/github.com/cloudfoundry/cli/cf/commands/application/delete_app_test.go
   src/github.com/cloudfoundry/cli/cf/terminal/ui_test.go
   src/github.com/cloudfoundry/loggregator_consumer/consumer_test.go
*/

package application_test

import (
	"github.com/cloudfoundry/cli/cf/api"
	. "github.com/cloudfoundry/cli/cf/commands/application"
	"github.com/cloudfoundry/cli/cf/models"
	testapi "github.com/cloudfoundry/cli/testhelpers/api"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
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
			appRepo := &testapi.FakeApplicationRepository{}
			appRepo.ReadReturns.App = app

			ui := callStop([]string{}, requirementsFactory, appRepo)
			Expect(ui.FailedWithUsage).To(BeTrue())
		})

		It("stops the app with the given name", func() {
			app := models.Application{}
			app.Name = "my-app"
			app.Guid = "my-app-guid"
			appRepo := &testapi.FakeApplicationRepository{}
			appRepo.ReadReturns.App = app
			args := []string{"my-app"}
			requirementsFactory.Application = app
			ui := callStop(args, requirementsFactory, appRepo)

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Stopping app", "my-app", "my-org", "my-space", "my-user"},
				[]string{"OK"},
			))

			Expect(requirementsFactory.ApplicationName).To(Equal("my-app"))
			Expect(appRepo.UpdateAppGuid).To(Equal("my-app-guid"))
		})

		It("warns the user when stopping the app fails", func() {
			app := models.Application{}
			app.Name = "my-app"
			app.Guid = "my-app-guid"
			appRepo := &testapi.FakeApplicationRepository{UpdateErr: true}
			appRepo.ReadReturns.App = app
			args := []string{"my-app"}
			requirementsFactory.Application = app
			ui := callStop(args, requirementsFactory, appRepo)

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Stopping", "my-app"},
				[]string{"FAILED"},
				[]string{"Error updating app."},
			))
			Expect(appRepo.UpdateAppGuid).To(Equal("my-app-guid"))
		})

		It("warns the user when the app is already stopped", func() {
			app := models.Application{}
			app.Name = "my-app"
			app.Guid = "my-app-guid"
			app.State = "stopped"
			appRepo := &testapi.FakeApplicationRepository{}
			appRepo.ReadReturns.App = app

			args := []string{"my-app"}
			requirementsFactory.Application = app
			ui := callStop(args, requirementsFactory, appRepo)

			Expect(ui.Outputs).To(ContainSubstrings([]string{"my-app", "is already stopped"}))
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
