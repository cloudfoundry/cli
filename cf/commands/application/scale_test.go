package application_test

import (
	testApplication "github.com/cloudfoundry/cli/cf/api/applications/fakes"
	"github.com/cloudfoundry/cli/cf/command_registry"
	appCmdFakes "github.com/cloudfoundry/cli/cf/commands/application/fakes"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	"github.com/cloudfoundry/cli/testhelpers/maker"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

var _ = Describe("scale command", func() {
	var (
		requirementsFactory *testreq.FakeReqFactory
		restarter           *appCmdFakes.FakeApplicationRestarter
		appRepo             *testApplication.FakeApplicationRepository
		ui                  *testterm.FakeUI
		config              core_config.Repository
		app                 models.Application
		OriginalCommand     command_registry.Command
		deps                command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.RepoLocator = deps.RepoLocator.SetApplicationRepository(appRepo)
		deps.Config = config

		//inject fake 'command dependency' into registry
		command_registry.Register(restarter)

		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("scale").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true}

		//save original command and restore later
		OriginalCommand = command_registry.Commands.FindCommand("restart")

		restarter = &appCmdFakes.FakeApplicationRestarter{}
		//setup fakes to correctly interact with command_registry
		restarter.SetDependencyStub = func(_ command_registry.Dependency, _ bool) command_registry.Command {
			return restarter
		}
		restarter.MetaDataReturns(command_registry.CommandMetadata{Name: "restart"})

		appRepo = &testApplication.FakeApplicationRepository{}
		ui = new(testterm.FakeUI)
		config = testconfig.NewRepositoryWithDefaults()
	})

	AfterEach(func() {
		command_registry.Register(OriginalCommand)
	})

	Describe("requirements", func() {
		It("requires the user to be logged in with a targed space", func() {
			args := []string{"-m", "1G", "my-app"}

			requirementsFactory.LoginSuccess = false
			requirementsFactory.TargetedSpaceSuccess = true

			Expect(testcmd.RunCliCommand("scale", args, requirementsFactory, updateCommandDependency, false)).To(BeFalse())

			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedSpaceSuccess = false

			Expect(testcmd.RunCliCommand("scale", args, requirementsFactory, updateCommandDependency, false)).To(BeFalse())
		})

		It("requires an app to be specified", func() {
			passed := testcmd.RunCliCommand("scale", []string{"-m", "1G"}, requirementsFactory, updateCommandDependency, false)

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires", "argument"},
			))
			Expect(passed).To(BeFalse())
		})

		It("does not require any flags", func() {
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedSpaceSuccess = true

			Expect(testcmd.RunCliCommand("scale", []string{"my-app"}, requirementsFactory, updateCommandDependency, false)).To(BeTrue())
		})
	})

	Describe("scaling an app", func() {
		BeforeEach(func() {
			app = maker.NewApp(maker.Overrides{"name": "my-app", "guid": "my-app-guid"})
			app.InstanceCount = 42
			app.DiskQuota = 1024
			app.Memory = 256

			requirementsFactory.Application = app
			appRepo.UpdateReturns(app, nil)
		})

		Context("when no flags are specified", func() {
			It("prints a description of the app's limits", func() {
				testcmd.RunCliCommand("scale", []string{"my-app"}, requirementsFactory, updateCommandDependency, false)

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Showing", "my-app", "my-org", "my-space", "my-user"},
					[]string{"OK"},
					[]string{"memory", "256M"},
					[]string{"disk", "1G"},
					[]string{"instances", "42"},
				))

				Expect(ui.Outputs).ToNot(ContainSubstrings([]string{"Scaling", "my-app", "my-org", "my-space", "my-user"}))
			})
		})

		Context("when the user does not confirm 'yes'", func() {
			It("does not restart the app", func() {
				ui.Inputs = []string{"whatever"}
				testcmd.RunCliCommand("scale", []string{"-i", "5", "-m", "512M", "-k", "2G", "my-app"}, requirementsFactory, updateCommandDependency, false)

				Expect(restarter.ApplicationRestartCallCount()).To(Equal(0))
			})
		})

		Context("when the user provides the -f flag", func() {
			It("does not prompt the user", func() {
				testcmd.RunCliCommand("scale", []string{"-f", "-i", "5", "-m", "512M", "-k", "2G", "my-app"}, requirementsFactory, updateCommandDependency, false)

				application, orgName, spaceName := restarter.ApplicationRestartArgsForCall(0)
				Expect(application).To(Equal(app))
				Expect(orgName).To(Equal(config.OrganizationFields().Name))
				Expect(spaceName).To(Equal(config.SpaceFields().Name))
			})
		})

		Context("when the user confirms they want to restart", func() {
			BeforeEach(func() {
				ui.Inputs = []string{"yes"}
			})

			It("can set an app's instance count, memory limit and disk limit", func() {
				testcmd.RunCliCommand("scale", []string{"-i", "5", "-m", "512M", "-k", "2G", "my-app"}, requirementsFactory, updateCommandDependency, false)

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Scaling", "my-app", "my-org", "my-space", "my-user"},
					[]string{"OK"},
				))

				Expect(ui.Prompts).To(ContainSubstrings([]string{"This will cause the app to restart", "Are you sure", "my-app"}))

				application, orgName, spaceName := restarter.ApplicationRestartArgsForCall(0)
				Expect(application).To(Equal(app))
				Expect(orgName).To(Equal(config.OrganizationFields().Name))
				Expect(spaceName).To(Equal(config.SpaceFields().Name))

				appGUID, params := appRepo.UpdateArgsForCall(0)
				Expect(appGUID).To(Equal("my-app-guid"))
				Expect(*params.Memory).To(Equal(int64(512)))
				Expect(*params.InstanceCount).To(Equal(5))
				Expect(*params.DiskQuota).To(Equal(int64(2048)))
			})

			It("does not scale the memory and disk limits if they are not specified", func() {
				testcmd.RunCliCommand("scale", []string{"-i", "5", "my-app"}, requirementsFactory, updateCommandDependency, false)

				Expect(restarter.ApplicationRestartCallCount()).To(Equal(0))

				appGUID, params := appRepo.UpdateArgsForCall(0)
				Expect(appGUID).To(Equal("my-app-guid"))
				Expect(*params.InstanceCount).To(Equal(5))
				Expect(params.DiskQuota).To(BeNil())
				Expect(params.Memory).To(BeNil())
			})

			It("does not scale the app's instance count if it is not specified", func() {
				testcmd.RunCliCommand("scale", []string{"-m", "512M", "my-app"}, requirementsFactory, updateCommandDependency, false)

				application, orgName, spaceName := restarter.ApplicationRestartArgsForCall(0)
				Expect(application).To(Equal(app))
				Expect(orgName).To(Equal(config.OrganizationFields().Name))
				Expect(spaceName).To(Equal(config.SpaceFields().Name))

				appGUID, params := appRepo.UpdateArgsForCall(0)
				Expect(appGUID).To(Equal("my-app-guid"))
				Expect(*params.Memory).To(Equal(int64(512)))
				Expect(params.DiskQuota).To(BeNil())
				Expect(params.InstanceCount).To(BeNil())
			})
		})
	})
})
