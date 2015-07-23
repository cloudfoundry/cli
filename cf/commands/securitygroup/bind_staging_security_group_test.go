package securitygroup_test

import (
	"errors"

	fakeStaging "github.com/cloudfoundry/cli/cf/api/security_groups/defaults/staging/fakes"
	fakeSecurityGroup "github.com/cloudfoundry/cli/cf/api/security_groups/fakes"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	"github.com/cloudfoundry/cli/cf/command_registry"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("bind-staging-security-group command", func() {
	var (
		ui                           *testterm.FakeUI
		configRepo                   core_config.Repository
		requirementsFactory          *testreq.FakeReqFactory
		fakeSecurityGroupRepo        *fakeSecurityGroup.FakeSecurityGroupRepo
		fakeStagingSecurityGroupRepo *fakeStaging.FakeStagingSecurityGroupsRepo
		deps                         command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.RepoLocator = deps.RepoLocator.SetSecurityGroupRepository(fakeSecurityGroupRepo)
		deps.RepoLocator = deps.RepoLocator.SetStagingSecurityGroupRepository(fakeStagingSecurityGroupRepo)
		deps.Config = configRepo
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("bind-staging-security-group").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{}
		fakeSecurityGroupRepo = &fakeSecurityGroup.FakeSecurityGroupRepo{}
		fakeStagingSecurityGroupRepo = &fakeStaging.FakeStagingSecurityGroupsRepo{}
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCliCommand("bind-staging-security-group", args, requirementsFactory, updateCommandDependency, false)
	}

	Describe("requirements", func() {
		It("fails when the user is not logged in", func() {
			Expect(runCommand("name")).To(BeFalse())
		})

		It("fails with usage when a name is not provided", func() {
			runCommand()
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Incorrect Usage", "Requires", "argument"},
			))
		})
	})

	Context("when the user is logged in and provides the name of a group", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
			group := models.SecurityGroup{}
			group.Guid = "just-pretend-this-is-a-guid"
			group.Name = "a-security-group-name"
			fakeSecurityGroupRepo.ReadReturns(group, nil)
		})

		JustBeforeEach(func() {
			runCommand("a-security-group-name")
		})

		It("binds the group to the default staging group set", func() {
			Expect(fakeSecurityGroupRepo.ReadArgsForCall(0)).To(Equal("a-security-group-name"))
			Expect(fakeStagingSecurityGroupRepo.BindToStagingSetArgsForCall(0)).To(Equal("just-pretend-this-is-a-guid"))
		})

		It("describes what it's doing to the user", func() {
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Binding", "a-security-group-name", "as", "my-user"},
				[]string{"OK"},
			))
		})

		Context("when binding the security group to the default set fails", func() {
			BeforeEach(func() {
				fakeStagingSecurityGroupRepo.BindToStagingSetReturns(errors.New("WOAH. I know kung fu"))
			})

			It("fails and describes the failure to the user", func() {
				Expect(ui.Outputs).To(ContainSubstrings(
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
				Expect(fakeStagingSecurityGroupRepo.BindToStagingSetCallCount()).To(Equal(0))
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"FAILED"},
				))
			})
		})
	})
})
