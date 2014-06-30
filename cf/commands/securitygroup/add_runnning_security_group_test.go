package securitygroup_test

import (
	"errors"

	fakeRunning "github.com/cloudfoundry/cli/cf/api/security_groups/defaults/running/fakes"
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

var _ = Describe("add-running-security-group command", func() {
	var (
		ui                           *testterm.FakeUI
		configRepo                   configuration.ReadWriter
		requirementsFactory          *testreq.FakeReqFactory
		fakeSecurityGroupRepo        *fakeSecurityGroup.FakeSecurityGroup
		fakeRunningSecurityGroupRepo *fakeRunning.FakeRunningSecurityGroupsRepo
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{}
		fakeSecurityGroupRepo = &fakeSecurityGroup.FakeSecurityGroup{}
		fakeRunningSecurityGroupRepo = &fakeRunning.FakeRunningSecurityGroupsRepo{}
	})

	runCommand := func(args ...string) {
		cmd := NewAddToRunningGroup(ui, configRepo, fakeSecurityGroupRepo, fakeRunningSecurityGroupRepo)
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

		It("adds the group to the running group set", func() {
			Expect(fakeSecurityGroupRepo.ReadCalledWith.Name).To(Equal("security-group-name"))
			Expect(fakeRunningSecurityGroupRepo.AddToRunningSetArgsForCall(0)).To(Equal("being-a-guid"))
		})

		Context("when adding the security group to the running set fails", func() {
			BeforeEach(func() {
				fakeRunningSecurityGroupRepo.AddToRunningSetReturns(errors.New("WOAH. I know kung fu"))
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
				Expect(fakeRunningSecurityGroupRepo.AddToRunningSetCallCount()).To(Equal(0))
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"FAILED"},
				))
			})
		})
	})
})
