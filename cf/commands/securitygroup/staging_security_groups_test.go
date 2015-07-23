package securitygroup_test

import (
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"

	testapi "github.com/cloudfoundry/cli/cf/api/security_groups/defaults/staging/fakes"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("staging-security-groups command", func() {
	var (
		ui                           *testterm.FakeUI
		configRepo                   core_config.Repository
		fakeStagingSecurityGroupRepo *testapi.FakeStagingSecurityGroupsRepo
		requirementsFactory          *testreq.FakeReqFactory
		deps                         command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.RepoLocator = deps.RepoLocator.SetStagingSecurityGroupRepository(fakeStagingSecurityGroupRepo)
		deps.Config = configRepo
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("staging-security-groups").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		fakeStagingSecurityGroupRepo = &testapi.FakeStagingSecurityGroupsRepo{}
		requirementsFactory = &testreq.FakeReqFactory{}
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCliCommand("staging-security-groups", args, requirementsFactory, updateCommandDependency, false)
	}

	Describe("requirements", func() {
		It("should fail when not logged in", func() {
			Expect(runCommand()).ToNot(HavePassedRequirements())
		})
	})

	Context("when the user is logged in", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
		})

		Context("when there are some security groups set for staging", func() {
			BeforeEach(func() {
				fakeStagingSecurityGroupRepo.ListReturns([]models.SecurityGroupFields{
					{Name: "hiphopopotamus"},
					{Name: "my lyrics are bottomless"},
					{Name: "steve"},
				}, nil)
			})

			It("shows the user the name of the security groups for staging", func() {
				Expect(runCommand()).To(BeTrue())
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Acquiring", "staging security group", "my-user"},
					[]string{"hiphopopotamus"},
					[]string{"my lyrics are bottomless"},
					[]string{"steve"},
				))
			})
		})

		Context("when the API returns an error", func() {
			BeforeEach(func() {
				fakeStagingSecurityGroupRepo.ListReturns(nil, errors.New("uh oh"))
			})

			It("fails loudly", func() {
				runCommand()
				Expect(ui.Outputs).To(ContainSubstrings([]string{"FAILED"}))
			})
		})

		Context("when there are no security groups set for staging", func() {
			It("tells the user that there are none", func() {
				runCommand()
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"No", "staging security group", "set"},
				))
			})
		})
	})
})
