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
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("get-health-check").SetDependency(deps, pluginCall))
	}

	runCommand := func(args ...string) bool {
		return testcmd.RunCliCommand("get-health-check", args, requirementsFactory, updateCommandDependency, false)
	}

	Describe("requirements", func() {
		It("fails with usage when called without enough arguments", func() {
			requirementsFactory.LoginSuccess = true

			runCommand()
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"get-health-check"},
				[]string{"Incorrect Usage", "Requires", "argument"},
			))
		})

		It("fails requirements when not logged in", func() {
			Expect(runCommand("my-app")).To(BeFalse())
		})

		It("fails if a space is not targeted", func() {
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedSpaceSuccess = false
			Expect(runCommand("my-app")).To(BeFalse())
		})
	})

	Describe("getting health_check_type", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedSpaceSuccess = true
		})

		Context("when application is not found", func() {
			It("Fails", func() {
				requirementsFactory.ApplicationFails = true
				Ω(runCommand("non-exist-app")).To(BeFalse())
			})
		})

		Context("when application exists", func() {
			BeforeEach(func() {
				app := models.Application{}
				app.Name = "my-app"
				app.Guid = "my-app-guid"
				app.HealthCheckType = "port"

				requirementsFactory.Application = app
			})

			It("shows the health_check_type", func() {
				runCommand("my-app")

				Ω(ui.Outputs).To(ContainSubstrings([]string{"Getting", "my-app", "health_check_type"}))
				Ω(ui.Outputs).To(ContainSubstrings([]string{"port"}))
			})
		})
	})

})
