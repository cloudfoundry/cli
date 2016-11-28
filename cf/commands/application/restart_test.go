package application_test

import (
	"os"

	"code.cloudfoundry.org/cli/cf/commands/application/applicationfakes"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"
	"code.cloudfoundry.org/cli/cf/trace/tracefakes"
	testcmd "code.cloudfoundry.org/cli/util/testhelpers/commands"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"

	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("restart command", func() {
	var (
		ui                  *testterm.FakeUI
		requirementsFactory *requirementsfakes.FakeFactory
		starter             *applicationfakes.FakeStarter
		stopper             *applicationfakes.FakeStopper
		config              coreconfig.Repository
		app                 models.Application
		originalStop        commandregistry.Command
		originalStart       commandregistry.Command
		deps                commandregistry.Dependency
		applicationReq      *requirementsfakes.FakeApplicationRequirement
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.Config = config

		//inject fake 'stopper and starter' into registry
		commandregistry.Register(starter)
		commandregistry.Register(stopper)

		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("restart").SetDependency(deps, pluginCall))
	}

	runCommand := func(args ...string) bool {
		return testcmd.RunCLICommand("restart", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		deps = commandregistry.NewDependency(os.Stdout, new(tracefakes.FakePrinter), "")
		requirementsFactory = new(requirementsfakes.FakeFactory)
		starter = new(applicationfakes.FakeStarter)
		stopper = new(applicationfakes.FakeStopper)
		config = testconfig.NewRepositoryWithDefaults()

		app = models.Application{}
		app.Name = "my-app"
		app.GUID = "my-app-guid"

		applicationReq = new(requirementsfakes.FakeApplicationRequirement)
		applicationReq.GetApplicationReturns(app)

		//save original command and restore later
		originalStart = commandregistry.Commands.FindCommand("start")
		originalStop = commandregistry.Commands.FindCommand("stop")

		//setup fakes to correctly interact with commandregistry
		starter.SetDependencyStub = func(_ commandregistry.Dependency, _ bool) commandregistry.Command {
			return starter
		}
		starter.MetaDataReturns(commandregistry.CommandMetadata{Name: "start"})

		stopper.SetDependencyStub = func(_ commandregistry.Dependency, _ bool) commandregistry.Command {
			return stopper
		}
		stopper.MetaDataReturns(commandregistry.CommandMetadata{Name: "stop"})
	})

	AfterEach(func() {
		commandregistry.Register(originalStart)
		commandregistry.Register(originalStop)
	})

	Describe("requirements", func() {
		It("fails with usage when not provided exactly one arg", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
			runCommand()
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires an argument"},
			))
		})

		It("fails when not logged in", func() {
			requirementsFactory.NewApplicationRequirementReturns(applicationReq)
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
			Expect(runCommand()).To(BeFalse())
		})

		It("fails when a space is not targeted", func() {
			requirementsFactory.NewApplicationRequirementReturns(applicationReq)
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})

			Expect(runCommand()).To(BeFalse())
		})
	})

	Context("when logged in, targeting a space, and an app name is provided", func() {
		BeforeEach(func() {
			requirementsFactory.NewApplicationRequirementReturns(applicationReq)
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
			requirementsFactory.NewTargetedSpaceRequirementReturns(requirements.Passing{})

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
		})
	})
})
