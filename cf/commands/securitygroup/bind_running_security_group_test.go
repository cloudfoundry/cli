package securitygroup_test

import (
	"errors"

	"code.cloudfoundry.org/cli/cf/api/securitygroups/defaults/running/runningfakes"
	"code.cloudfoundry.org/cli/cf/api/securitygroups/securitygroupsfakes"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"
	testcmd "code.cloudfoundry.org/cli/util/testhelpers/commands"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"

	"code.cloudfoundry.org/cli/cf/commandregistry"
	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("bind-running-security-group command", func() {
	var (
		ui                           *testterm.FakeUI
		configRepo                   coreconfig.Repository
		requirementsFactory          *requirementsfakes.FakeFactory
		fakeSecurityGroupRepo        *securitygroupsfakes.FakeSecurityGroupRepo
		fakeRunningSecurityGroupRepo *runningfakes.FakeSecurityGroupsRepo
		deps                         commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.RepoLocator = deps.RepoLocator.SetSecurityGroupRepository(fakeSecurityGroupRepo)
		deps.RepoLocator = deps.RepoLocator.SetRunningSecurityGroupRepository(fakeRunningSecurityGroupRepo)
		deps.Config = configRepo
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("bind-running-security-group").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = new(requirementsfakes.FakeFactory)
		fakeSecurityGroupRepo = new(securitygroupsfakes.FakeSecurityGroupRepo)
		fakeRunningSecurityGroupRepo = new(runningfakes.FakeSecurityGroupsRepo)
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCLICommand("bind-running-security-group", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	Describe("requirements", func() {
		It("fails when the user is not logged in", func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Failing{Message: "not logged in"})
			Expect(runCommand("name")).To(BeFalse())
		})

		It("fails with usage when a name is not provided", func() {
			runCommand()
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires", "argument"},
			))
		})
	})

	Context("when the user is logged in and provides the name of a group", func() {
		BeforeEach(func() {
			requirementsFactory.NewLoginRequirementReturns(requirements.Passing{})
			group := models.SecurityGroup{}
			group.GUID = "being-a-guid"
			group.Name = "security-group-name"
			fakeSecurityGroupRepo.ReadReturns(group, nil)
		})

		JustBeforeEach(func() {
			runCommand("security-group-name")
		})

		It("Describes what it is doing to the user", func() {
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Binding", "security-group-name", "as", "my-user"},
				[]string{"OK"},
				[]string{"TIP: Changes will not apply to existing running applications until they are restarted."},
			))
		})

		It("binds the group to the running group set", func() {
			Expect(fakeSecurityGroupRepo.ReadArgsForCall(0)).To(Equal("security-group-name"))
			Expect(fakeRunningSecurityGroupRepo.BindToRunningSetArgsForCall(0)).To(Equal("being-a-guid"))
		})

		Context("when binding the security group to the running set fails", func() {
			BeforeEach(func() {
				fakeRunningSecurityGroupRepo.BindToRunningSetReturns(errors.New("WOAH. I know kung fu"))
			})

			It("fails and describes the failure to the user", func() {
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"WOAH. I know kung fu"},
				))
			})
		})

		Context("when the security group with the given name cannot be found", func() {
			BeforeEach(func() {
				fakeSecurityGroupRepo.ReadReturns(models.SecurityGroup{}, errors.New("Crème insufficiently brûlée'd"))
			})

			It("fails and tells the user that the security group does not exist", func() {
				Expect(fakeRunningSecurityGroupRepo.BindToRunningSetCallCount()).To(Equal(0))
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"FAILED"},
				))
			})
		})
	})
})
