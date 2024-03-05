package environmentvariablegroup_test

import (
	"code.cloudfoundry.org/cli/cf/api/environmentvariablegroups/environmentvariablegroupsfakes"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	cf_errors "code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"
	testcmd "code.cloudfoundry.org/cli/cf/util/testhelpers/commands"
	testconfig "code.cloudfoundry.org/cli/cf/util/testhelpers/configuration"
	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	testterm "code.cloudfoundry.org/cli/cf/util/testhelpers/terminal"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("set-staging-environment-variable-group command", func() {
	var (
		ui                           *testterm.FakeUI
		requirementsFactory          *requirementsfakes.FakeFactory
		environmentVariableGroupRepo *environmentvariablegroupsfakes.FakeRepository
		configRepo                   coreconfig.Repository
		deps                         commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.RepoLocator = deps.RepoLocator.SetEnvironmentVariableGroupsRepository(environmentVariableGroupRepo)
		deps.Config = configRepo
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("set-staging-environment-variable-group").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = new(requirementsfakes.FakeFactory)
		environmentVariableGroupRepo = new(environmentvariablegroupsfakes.FakeRepository)
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCLICommand("set-staging-environment-variable-group", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	Describe("requirements", func() {
		It("requires the user to be logged in", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Failing{Message: "not logged in"})
			Expect(runCommand()).ToNot(HavePassedRequirements())
		})

		It("fails with usage when it does not receive any arguments", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
			runCommand()
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires an argument"},
			))
		})
	})

	Describe("when logged in", func() {
		BeforeEach(func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
		})

		It("Sets the staging environment variable group", func() {
			runCommand(`{"abc":"123", "def": "456"}`)

			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Setting the contents of the staging environment variable group as my-user..."},
				[]string{"OK"},
			))
			Expect(environmentVariableGroupRepo.SetStagingArgsForCall(0)).To(Equal(`{"abc":"123", "def": "456"}`))
		})

		It("Fails with a reasonable message when invalid JSON is passed", func() {
			environmentVariableGroupRepo.SetStagingReturns(cf_errors.NewHTTPError(400, cf_errors.MessageParseError, "Request invalid due to parse error"))
			runCommand(`{"abc":"123", "invalid : "json"}`)
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Setting the contents of the staging environment variable group as my-user..."},
				[]string{"FAILED"},
				[]string{`Your JSON string syntax is invalid.  Proper syntax is this:  cf set-staging-environment-variable-group '{"name":"value","name":"value"}'`},
			))
		})
	})
})
