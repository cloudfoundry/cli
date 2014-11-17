package securitygroup_test

import (
	"errors"

	fakeRunning "github.com/cloudfoundry/cli/cf/api/security_groups/defaults/running/fakes"
	fakeSecurityGroup "github.com/cloudfoundry/cli/cf/api/security_groups/fakes"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/models"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/commands/securitygroup"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("bind-running-security-group command", func() {
	var (
		ui                           *testterm.FakeUI
		configRepo                   core_config.ReadWriter
		requirementsFactory          *testreq.FakeReqFactory
		fakeSecurityGroupRepo        *fakeSecurityGroup.FakeSecurityGroupRepo
		fakeRunningSecurityGroupRepo *fakeRunning.FakeRunningSecurityGroupsRepo
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{}
		fakeSecurityGroupRepo = &fakeSecurityGroup.FakeSecurityGroupRepo{}
		fakeRunningSecurityGroupRepo = &fakeRunning.FakeRunningSecurityGroupsRepo{}
	})

	runCommand := func(args ...string) bool {
		cmd := NewBindToRunningGroup(ui, configRepo, fakeSecurityGroupRepo, fakeRunningSecurityGroupRepo)
		return testcmd.RunCommand(cmd, args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("fails when the user is not logged in", func() {
			Expect(runCommand("name")).To(BeFalse())
		})

		It("fails with usage when a name is not provided", func() {
			runCommand()
			Expect(ui.FailedWithUsage).To(BeTrue())
		})
	})

	Context("when the user is logged in and provides the name of a group", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
			group := models.SecurityGroup{}
			group.Guid = "being-a-guid"
			group.Name = "security-group-name"
			fakeSecurityGroupRepo.ReadReturns(group, nil)
		})

		JustBeforeEach(func() {
			runCommand("security-group-name")
		})

		It("Describes what it is doing to the user", func() {
			Expect(ui.Outputs).To(ContainSubstrings(
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
				Expect(fakeRunningSecurityGroupRepo.BindToRunningSetCallCount()).To(Equal(0))
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"FAILED"},
				))
			})
		})
	})
})
