package securitygroup_test

import (
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/commands/securitygroup"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("assign-security-group command", func() {
	var (
		ui                  *testterm.FakeUI
		cmd                 AssignSecurityGroup
		requirementsFactory *testreq.FakeReqFactory
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		cmd = NewAssignSecurityGroup(ui)
		requirementsFactory = &testreq.FakeReqFactory{}
	})

	runCommand := func(args ...string) {
		testcmd.RunCommand(cmd, args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("fails when the user is not logged in", func() {
			runCommand("my-craaaaaazy-security-group")

			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})

		It("succeeds when the user is logged in", func() {
			requirementsFactory.LoginSuccess = true
			runCommand("my-craaaaaazy-security-group")

			Expect(testcmd.CommandDidPassRequirements).To(BeTrue())
		})

		It("fails with usage when not provided the name of a security group", func() {
			requirementsFactory.LoginSuccess = true
			runCommand()

			Expect(ui.FailedWithUsage).To(BeTrue())
		})
	})
})
