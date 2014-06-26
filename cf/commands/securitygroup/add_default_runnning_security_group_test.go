package securitygroup_test

import (
	"errors"

	fakeRunningDefaults "github.com/cloudfoundry/cli/cf/api/security_groups/defaults/running/fakes"
	fakeSecurityGroup "github.com/cloudfoundry/cli/cf/api/security_groups/fakes"
	"github.com/cloudfoundry/cli/cf/configuration"
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

var _ = Describe("add-default-running-security-group command", func() {
	var (
		ui                           *testterm.FakeUI
		configRepo                   configuration.ReadWriter
		requirementsFactory          *testreq.FakeReqFactory
		fakeSecurityGroupRepo        *fakeSecurityGroup.FakeSecurityGroup
		fakeRunningSecurityGroupRepo *fakeRunningDefaults.FakeRunningSecurityGroupsRepo
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{}
		fakeSecurityGroupRepo = &fakeSecurityGroup.FakeSecurityGroup{}
		fakeRunningSecurityGroupRepo = &fakeRunningDefaults.FakeRunningSecurityGroupsRepo{}
	})

	runCommand := func(args ...string) {
		cmd := NewAddToDefaultRunningGroup(ui, configRepo, fakeSecurityGroupRepo, fakeRunningSecurityGroupRepo)
		testcmd.RunCommand(cmd, args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("fails when the user is not logged in", func() {
			runCommand("name")
			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
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
			fakeSecurityGroupRepo.ReadReturns.SecurityGroup = group
		})

		JustBeforeEach(func() {
			runCommand("security-group-name")
		})

		It("Describes what it is doing to the user", func() {
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Adding", "security-group-name", "as", "my-user"},
				[]string{"OK"},
			))
		})

		It("adds the group to the default running group set", func() {
			Expect(fakeSecurityGroupRepo.ReadCalledWith.Name).To(Equal("security-group-name"))
			Expect(fakeRunningSecurityGroupRepo.AddToDefaultRunningSetArgsForCall(0)).To(Equal("being-a-guid"))
		})

		Context("when adding the security group to the default set fails", func() {
			BeforeEach(func() {
				fakeRunningSecurityGroupRepo.AddToDefaultRunningSetReturns(errors.New("WOAH. I know kung fu"))
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
				fakeSecurityGroupRepo.ReadReturns.Error = errors.New("Crème insufficiently brûlée'd")
			})

			It("fails and tells the user that the security group does not exist", func() {
				Expect(fakeRunningSecurityGroupRepo.AddToDefaultRunningSetCallCount()).To(Equal(0))
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"FAILED"},
				))
			})
		})
	})
})
