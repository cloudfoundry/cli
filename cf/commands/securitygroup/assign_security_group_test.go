package securitygroup_test

import (
	"github.com/cloudfoundry/cli/cf/errors"

	testapi "github.com/cloudfoundry/cli/cf/api/security_groups/fakes"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/commands/securitygroup"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("assign-security-group command", func() {
	var (
		ui                  *testterm.FakeUI
		cmd                 AssignSecurityGroup
		securityGroupRepo   *testapi.FakeSecurityGroup
		requirementsFactory *testreq.FakeReqFactory
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		securityGroupRepo = &testapi.FakeSecurityGroup{}
		cmd = NewAssignSecurityGroup(ui, securityGroupRepo)
		requirementsFactory = &testreq.FakeReqFactory{}
	})

	runCommand := func(args ...string) {
		testcmd.RunCommand(cmd, args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("fails when the user is not logged in", func() {
			runCommand("--space", "my-space", "--org", "my-org", "my-craaaaaazy-security-group")

			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})

		It("succeeds when the user is logged in", func() {
			requirementsFactory.LoginSuccess = true
			runCommand("--space", "my-space", "--org", "my-org", "my-craaaaaazy-security-group")

			Expect(testcmd.CommandDidPassRequirements).To(BeTrue())
		})

		It("fails with usage when not provided the name of a security group", func() {
			requirementsFactory.LoginSuccess = true
			runCommand()

			Expect(ui.FailedWithUsage).To(BeTrue())
		})
	})

	Context("when the user is logged in and provides the name of a security group", func() {
		BeforeEach(func() {
			requirementsFactory.LoginSuccess = true
		})

		It("fails when the user does not provide an org or space", func() {
			runCommand("my-under-specified-security-group")

			Expect(ui.FailedWithUsage).To(BeTrue())
		})

		Context("when a security group with that name does not exist", func() {
			BeforeEach(func() {
				securityGroupRepo.ReadReturns.Error = errors.NewModelNotFoundError("security group", "my-nonexistent-security-group")
			})

			It("fails and tells the user", func() {
				runCommand("--space", "my-space", "--org", "my-org", "my-nonexistent-security-group")

				Expect(securityGroupRepo.ReadCalledWith.Name).To(Equal("my-nonexistent-security-group"))
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"FAILED"},
					[]string{"security group", "my-nonexistent-security-group", "not found"},
				))
			})
		})
	})
})
