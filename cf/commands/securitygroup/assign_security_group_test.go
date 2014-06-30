package securitygroup_test

import (
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"

	. "github.com/cloudfoundry/cli/cf/commands/securitygroup"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("assign-security-group command", func() {
	var (
		cmd                 AssignSecurityGroup
		requirementsFactory *testreq.FakeReqFactory
	)

	BeforeEach(func() {
		cmd = NewAssignSecurityGroup()
		requirementsFactory = &testreq.FakeReqFactory{}
	})

	runCommand := func(args ...string) {
		testcmd.RunCommand(cmd, args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("fails when the user is not logged in", func() {
			runCommand()

			Expect(testcmd.CommandDidPassRequirements).To(BeFalse())
		})

		It("succeeds when the user is logged in", func() {
			requirementsFactory.LoginSuccess = true
			runCommand()

			Expect(testcmd.CommandDidPassRequirements).To(BeTrue())
		})
	})
})
