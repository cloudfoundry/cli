package environmentvariablegroup_test

import (
	test_environmentVariableGroups "github.com/cloudfoundry/cli/cf/api/environment_variable_groups/fakes"
	. "github.com/cloudfoundry/cli/cf/commands/environmentvariablegroup"
	cf_errors "github.com/cloudfoundry/cli/cf/errors"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("set-running-environment-variable-group command", func() {
	var (
		ui                           *testterm.FakeUI
		requirementsFactory          *testreq.FakeReqFactory
		environmentVariableGroupRepo *test_environmentVariableGroups.FakeEnvironmentVariableGroupsRepository
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true}
		environmentVariableGroupRepo = &test_environmentVariableGroups.FakeEnvironmentVariableGroupsRepository{}
	})

	runCommand := func(args ...string) bool {
		cmd := NewSetRunningEnvironmentVariableGroup(ui, testconfig.NewRepositoryWithDefaults(), environmentVariableGroupRepo)
		return testcmd.RunCommand(cmd, args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("requires the user to be logged in", func() {
			requirementsFactory.LoginSuccess = false
			Expect(runCommand()).ToNot(HavePassedRequirements())
		})

		It("fails with usage when it does not receive any arguments", func() {
			requirementsFactory.LoginSuccess = true
			runCommand()
			Expect(ui.FailedWithUsage).To(BeTrue())
		})
	})

	Describe("when logged in", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
		})

		It("Sets the running environment variable group", func() {
			runCommand(`{"abc":"123", "def": "456"}`)

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Setting the contents of the running environment variable group as my-user..."},
				[]string{"OK"},
			))
			Expect(environmentVariableGroupRepo.SetRunningArgsForCall(0)).To(Equal(`{"abc":"123", "def": "456"}`))
		})

		It("Fails with a reasonable message when invalid JSON is passed", func() {
			environmentVariableGroupRepo.SetRunningReturns(cf_errors.NewHttpError(400, cf_errors.PARSE_ERROR, "Request invalid due to parse error"))
			runCommand(`{"abc":"123", "invalid : "json"}`)
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Setting the contents of the running environment variable group as my-user..."},
				[]string{"FAILED"},
				[]string{`Your JSON string syntax is invalid.  Proper syntax is this:  cf set-running-environment-variable-group '{"name":"value","name":"value"}'`},
			))
		})
	})
})
