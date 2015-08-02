package environmentvariablegroup_test

import (
	test_environmentVariableGroups "github.com/cloudfoundry/cli/cf/api/environment_variable_groups/fakes"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("staging-environment-variable-group command", func() {
	var (
		ui                           *testterm.FakeUI
		requirementsFactory          *testreq.FakeReqFactory
		configRepo                   core_config.Repository
		environmentVariableGroupRepo *test_environmentVariableGroups.FakeEnvironmentVariableGroupsRepository
		deps                         command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.RepoLocator = deps.RepoLocator.SetEnvironmentVariableGroupsRepository(environmentVariableGroupRepo)
		deps.Config = configRepo
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("staging-environment-variable-group").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true}
		environmentVariableGroupRepo = &test_environmentVariableGroups.FakeEnvironmentVariableGroupsRepository{}
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCliCommand("staging-environment-variable-group", args, requirementsFactory, updateCommandDependency, false)
	}

	Describe("requirements", func() {
		It("requires the user to be logged in", func() {
			requirementsFactory.LoginSuccess = false
			Expect(runCommand()).ToNot(HavePassedRequirements())
		})
		It("should fail with usage when provided any arguments", func() {
			requirementsFactory.LoginSuccess = true
			Expect(runCommand("blahblah")).To(BeFalse())
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Incorrect Usage", "No argument"},
			))
		})
	})

	Describe("when logged in", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
			environmentVariableGroupRepo.ListStagingReturns(
				[]models.EnvironmentVariable{
					models.EnvironmentVariable{Name: "abc", Value: "123"},
					models.EnvironmentVariable{Name: "def", Value: "456"},
				}, nil)
		})

		It("Displays the staging environment variable group", func() {
			runCommand()

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Retrieving the contents of the staging environment variable group as my-user..."},
				[]string{"OK"},
				[]string{"Variable Name", "Assigned Value"},
				[]string{"abc", "123"},
				[]string{"def", "456"},
			))
		})
	})
})
