package application_test

import (
	appCmdFakes "github.com/cloudfoundry/cli/cf/commands/application/fakes"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("restart command", func() {
	var (
		ui                  *testterm.FakeUI
		requirementsFactory *testreq.FakeReqFactory
		starter             *appCmdFakes.FakeApplicationStarter
		stopper             *appCmdFakes.FakeApplicationStopper
		config              core_config.Repository
		app                 models.Application
		originalStop        command_registry.Command
		originalStart       command_registry.Command
		deps                command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.Config = config

		//inject fake 'stopper and starter' into registry
		command_registry.Register(starter)
		command_registry.Register(stopper)

		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("restart").SetDependency(deps, pluginCall))
	}

	runCommand := func(args ...string) bool {
		return testcmd.RunCliCommand("restart", args, requirementsFactory, updateCommandDependency, false)
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		deps = command_registry.NewDependency()
		requirementsFactory = &testreq.FakeReqFactory{}
		starter = &appCmdFakes.FakeApplicationStarter{}
		stopper = &appCmdFakes.FakeApplicationStopper{}
		config = testconfig.NewRepositoryWithDefaults()

		app = models.Application{}
		app.Name = "my-app"
		app.Guid = "my-app-guid"

		//save original command and restore later
		originalStart = command_registry.Commands.FindCommand("start")
		originalStop = command_registry.Commands.FindCommand("stop")

		//setup fakes to correctly interact with command_registry
		starter.SetDependencyStub = func(_ command_registry.Dependency, _ bool) command_registry.Command {
			return starter
		}
		starter.MetaDataReturns(command_registry.CommandMetadata{Name: "start"})

		stopper.SetDependencyStub = func(_ command_registry.Dependency, _ bool) command_registry.Command {
			return stopper
		}
		stopper.MetaDataReturns(command_registry.CommandMetadata{Name: "stop"})
	})

	AfterEach(func() {
		command_registry.Register(originalStart)
		command_registry.Register(originalStop)
	})

	Describe("requirements", func() {
		It("fails with usage when not provided exactly one arg", func() {
			requirementsFactory.LoginSuccess = true
			runCommand()
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires an argument"},
			))
		})

		It("fails when not logged in", func() {
			requirementsFactory.Application = app
			requirementsFactory.TargetedSpaceSuccess = true

			Expect(runCommand()).To(BeFalse())
		})

		It("fails when a space is not targeted", func() {
			requirementsFactory.Application = app
			requirementsFactory.LoginSuccess = true

			Expect(runCommand()).To(BeFalse())
		})
	})

	Context("when logged in, targeting a space, and an app name is provided", func() {
		BeforeEach(func() {
			requirementsFactory.Application = app
			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedSpaceSuccess = true

			stopper.ApplicationStopReturns(app, nil)
		})

		It("restarts the given app", func() {
			runCommand("my-app")

			application, orgName, spaceName := stopper.ApplicationStopArgsForCall(0)
			Expect(application).To(Equal(app))
			Expect(orgName).To(Equal(config.OrganizationFields().Name))
			Expect(spaceName).To(Equal(config.SpaceFields().Name))

			application, orgName, spaceName = starter.ApplicationStartArgsForCall(0)
			Expect(application).To(Equal(app))
			Expect(orgName).To(Equal(config.OrganizationFields().Name))
			Expect(spaceName).To(Equal(config.SpaceFields().Name))

			Expect(requirementsFactory.ApplicationName).To(Equal("my-app"))
		})
	})
})
