package service_test

import (
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
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

var _ = Describe("rename-service command", func() {
	var (
		ui                  *testterm.FakeUI
		config              core_config.Repository
		serviceRepo         *testapi.FakeServiceRepository
		requirementsFactory *testreq.FakeReqFactory
		deps                command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.RepoLocator = deps.RepoLocator.SetServiceRepository(serviceRepo)
		deps.Config = config
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("rename-service").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		config = testconfig.NewRepositoryWithDefaults()
		serviceRepo = &testapi.FakeServiceRepository{}
		requirementsFactory = &testreq.FakeReqFactory{}
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCliCommand("rename-service", args, requirementsFactory, updateCommandDependency, false)
	}

	Describe("requirements", func() {
		It("Fails with usage when exactly two parameters not passed", func() {
			runCommand("whatever")
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires", "arguments"},
			))
		})

		It("fails when not logged in", func() {
			requirementsFactory.TargetedSpaceSuccess = true

			Expect(runCommand("banana", "fppants")).To(BeFalse())
		})

		It("fails when a space is not targeted", func() {
			requirementsFactory.LoginSuccess = true

			Expect(runCommand("banana", "faaaaasdf")).To(BeFalse())
		})
	})

	Context("when logged in and a space is targeted", func() {
		var serviceInstance models.ServiceInstance

		BeforeEach(func() {
			serviceInstance = models.ServiceInstance{}
			serviceInstance.Name = "different-name"
			serviceInstance.Guid = "different-name-guid"

			requirementsFactory.LoginSuccess = true
			requirementsFactory.TargetedSpaceSuccess = true
			requirementsFactory.ServiceInstance = serviceInstance
		})

		It("renames the service, obviously", func() {
			runCommand("my-service", "new-name")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Renaming service", "different-name", "new-name", "my-org", "my-space", "my-user"},
				[]string{"OK"},
			))

			actualServiceInstance, actualServiceName := serviceRepo.RenameServiceArgsForCall(0)
			Expect(actualServiceInstance).To(Equal(serviceInstance))
			Expect(actualServiceName).To(Equal("new-name"))
		})
	})
})
