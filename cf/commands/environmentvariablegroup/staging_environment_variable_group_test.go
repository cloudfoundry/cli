package environmentvariablegroup_test

import (
	test_environmentVariableGroups "github.com/cloudfoundry/cli/cf/api/environment_variable_groups/fakes"
	. "github.com/cloudfoundry/cli/cf/commands/environmentvariablegroup"
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
		environmentVariableGroupRepo *test_environmentVariableGroups.FakeEnvironmentVariableGroupsRepository
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		requirementsFactory = &testreq.FakeReqFactory{LoginSuccess: true}
		environmentVariableGroupRepo = &test_environmentVariableGroups.FakeEnvironmentVariableGroupsRepository{}
	})

	runCommand := func(args ...string) bool {
		cmd := NewStagingEnvironmentVariableGroup(ui, testconfig.NewRepositoryWithDefaults(), environmentVariableGroupRepo)
		return testcmd.RunCommand(cmd, args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("requires the user to be logged in", func() {
			requirementsFactory.LoginSuccess = false
			Expect(runCommand()).ToNot(HavePassedRequirements())
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
