package appsecuritygroup_test

import (
	"errors"

	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/models"
	testapi "github.com/cloudfoundry/cli/testhelpers/api"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/commands/appsecuritygroup"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("add-default-staging-application-security-group command", func() {
	var (
		ui                           *testterm.FakeUI
		configRepo                   configuration.ReadWriter
		requirementsFactory          *testreq.FakeReqFactory
		fakeSecurityGroupRepo        *testapi.FakeAppSecurityGroup
		fakeStagingSecurityGroupRepo *testapi.FakeStagingSecurityGroupsRepo
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{}
		fakeSecurityGroupRepo = &testapi.FakeAppSecurityGroup{}
		fakeStagingSecurityGroupRepo = &testapi.FakeStagingSecurityGroupsRepo{}
	})

	runCommand := func(args ...string) {
		cmd := NewAddToDefaultStagingGroup(ui, configRepo, fakeSecurityGroupRepo, fakeStagingSecurityGroupRepo)
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
			group := models.ApplicationSecurityGroup{}
			group.Guid = "just-pretend-this-is-a-guid"
			group.Name = "a-security-group-name"
			fakeSecurityGroupRepo.ReadReturns.ApplicationSecurityGroup = group
		})

		JustBeforeEach(func() {
			runCommand("a-security-group-name")
		})

		It("adds the group to the default staging group set", func() {
			Expect(fakeSecurityGroupRepo.ReadCalledWith.Name).To(Equal("a-security-group-name"))
			Expect(fakeStagingSecurityGroupRepo.AddToDefaultStagingSetArgsForCall(0)).To(Equal("just-pretend-this-is-a-guid"))
		})

		It("describes what it's doing to the user", func() {
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Adding", "a-security-group-name", "as", "my-user"},
				[]string{"OK"},
			))
		})

		Context("when adding the security group to the default set fails", func() {
			BeforeEach(func() {
				fakeStagingSecurityGroupRepo.AddToDefaultStagingSetReturns(errors.New("WOAH. I know kung fu"))
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
				Expect(fakeStagingSecurityGroupRepo.AddToDefaultStagingSetCallCount()).To(Equal(0))
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"FAILED"},
				))
			})
		})
	})
})
