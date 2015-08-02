package application_test

import (
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

var _ = Describe("unset-env command", func() {
	var (
		ui                  *testterm.FakeUI
		app                 models.Application
		appRepo             *testApplication.FakeApplicationRepository
		configRepo          core_config.Repository
		requirementsFactory *testreq.FakeReqFactory
		deps                command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.Config = configRepo
		deps.RepoLocator = deps.RepoLocator.SetApplicationRepository(appRepo)
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("unset-env").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		app = models.Application{}
		app.Name = "my-app"
		app.Guid = "my-app-guid"
		appRepo = &testApplication.FakeApplicationRepository{}
		requirementsFactory = &testreq.FakeReqFactory{}
		configRepo = testconfig.NewRepositoryWithDefaults()
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCliCommand("unset-env", args, requirementsFactory, updateCommandDependency, false)
	}

	Describe("requirements", func() {
		It("fails when not logged in", func() {
			requirementsFactory.TargetedSpaceSuccess = true
			requirementsFactory.Application = app

			Expect(runCommand("foo", "bar")).To(BeFalse())
		})

		It("fails when a space is not targeted", func() {
			requirementsFactory.LoginSuccess = true
			requirementsFactory.Application = app

			Expect(runCommand("foo", "bar")).To(BeFalse())
		})

		It("fails with usage when not provided with exactly 2 args", func() {
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedSpaceSuccess = true
			requirementsFactory.Application = app

			Expect(runCommand("too", "many", "args")).To(BeFalse())
		})
	})

	Context("when logged in, a space is targeted and provided enough args", func() {
		BeforeEach(func() {
			app.EnvironmentVars = map[string]interface{}{"foo": "bar", "DATABASE_URL": "mysql://example.com/my-db"}

			requirementsFactory.Application = app
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedSpaceSuccess = true
		})

		It("updates the app and tells the user what happened", func() {
			runCommand("my-app", "DATABASE_URL")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Removing env variable", "DATABASE_URL", "my-app", "my-org", "my-space", "my-user"},
				[]string{"OK"},
			))

			Expect(requirementsFactory.ApplicationName).To(Equal("my-app"))
			Expect(appRepo.UpdateAppGuid).To(Equal("my-app-guid"))
			Expect(*appRepo.UpdateParams.EnvironmentVars).To(Equal(map[string]interface{}{
				"foo": "bar",
			}))
		})

		Context("when updating the app fails", func() {
			BeforeEach(func() {
				appRepo.UpdateErr = true
				appRepo.ReadReturns.App = app
			})

			It("fails and alerts the user", func() {
				runCommand("does-not-exist", "DATABASE_URL")

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Removing env variable"},
					[]string{"FAILED"},
					[]string{"Error updating app."},
				))
			})
		})

		It("tells the user if the specified env var was not set", func() {
			runCommand("my-app", "CANT_STOP_WONT_STOP_UNSETTIN_THIS_ENV")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Removing env variable"},
				[]string{"OK"},
				[]string{"CANT_STOP_WONT_STOP_UNSETTIN_THIS_ENV", "was not set."},
			))
		})
	})
})
