package securitygroup_test

import (
	fakeRunningDefaults "github.com/cloudfoundry/cli/cf/api/security_groups/defaults/running/fakes"
	fakeSecurityGroup "github.com/cloudfoundry/cli/cf/api/security_groups/fakes"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
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

var _ = Describe("unbind-running-security-group command", func() {
	var (
		ui                            *testterm.FakeUI
		configRepo                    core_config.ReadWriter
		requirementsFactory           *testreq.FakeReqFactory
		fakeSecurityGroupRepo         *fakeSecurityGroup.FakeSecurityGroupRepo
		fakeRunningSecurityGroupsRepo *fakeRunningDefaults.FakeRunningSecurityGroupsRepo
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		configRepo = testconfig.NewRepositoryWithDefaults()
		requirementsFactory = &testreq.FakeReqFactory{}
		fakeSecurityGroupRepo = &fakeSecurityGroup.FakeSecurityGroupRepo{}
		fakeRunningSecurityGroupsRepo = &fakeRunningDefaults.FakeRunningSecurityGroupsRepo{}
	})

	runCommand := func(args ...string) {
		cmd := NewUnbindFromRunningGroup(ui, configRepo, fakeSecurityGroupRepo, fakeRunningSecurityGroupsRepo)
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
		})

		Context("security group exists", func() {
			BeforeEach(func() {
				group := models.SecurityGroup{}
				group.Guid = "just-pretend-this-is-a-guid"
				group.Name = "a-security-group-name"
				fakeSecurityGroupRepo.ReadReturns(group, nil)
			})

			JustBeforeEach(func() {
				runCommand("a-security-group-name")
			})

			It("unbinds the group from the running group set", func() {
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Unbinding", "security group", "a-security-group-name", "my-user"},
					[]string{"OK"},
				))

				Expect(fakeSecurityGroupRepo.ReadArgsForCall(0)).To(Equal("a-security-group-name"))
				Expect(fakeRunningSecurityGroupsRepo.UnbindFromRunningSetArgsForCall(0)).To(Equal("just-pretend-this-is-a-guid"))
			})
		})

		Context("when the security group does not exist", func() {
			BeforeEach(func() {
				fakeSecurityGroupRepo.ReadReturns(models.SecurityGroup{}, errors.NewModelNotFoundError("security group", "anana-qui-parle"))
			})

			It("warns the user", func() {
				runCommand("anana-qui-parle")
				Expect(ui.WarnOutputs).To(ContainSubstrings(
					[]string{"Security group", "anana-qui-parle", "does not exist"},
				))

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"OK"},
				))
			})
		})
	})
})
