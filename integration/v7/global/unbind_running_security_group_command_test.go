package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/resources"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("unbind-running-security-group command", func() {
	var (
		secGroupName string
	)

	BeforeEach(func() {
		secGroupName = helpers.NewSecurityGroupName()
		helpers.LoginCF()
	})

	Describe("help", func() {
		When("--help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("unbind-running-security-group", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say(`\s+unbind-running-security-group - Unbind a security group from the set of security groups for running applications globally`))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say(`\s+cf unbind-running-security-group SECURITY_GROUP`))
				Eventually(session).Should(Say(`TIP: Changes require an app restart \(for running\) or restage \(for staging\) to apply to existing applications\.`))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say(`\s+apps, restart, running-security-groups, security-groups`))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(false, false, ReadOnlyOrg, "unbind-running-security-group", "security-group-name")
		})
	})

	When("the input is invalid", func() {
		When("the security group is not provided", func() {
			It("fails with an incorrect usage message and displays help", func() {
				session := helpers.CF("unbind-running-security-group")
				Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `SECURITY_GROUP` was not provided"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Exit(1))
			})
		})
	})

	When("the security group doesn't exist", func() {
		It("succeeds with a security group not found message", func() {
			session := helpers.CF("unbind-running-security-group", "some-security-group-that-doesn't-exist")
			userName, _ := helpers.GetCredentials()
			Eventually(session).Should(Say("Unbinding security group %s from defaults for running as %s...", "some-security-group-that-doesn't-exist", userName))
			Eventually(session.Err).Should(Say("Security group 'some-security-group-that-doesn't-exist' not found."))
			Eventually(session).Should(Say("OK"))
			Eventually(session).Should(Exit(0))
		})
	})

	When("the security group exists", func() {
		var (
			someSecurityGroup resources.SecurityGroup
			ports             string
			description       string
		)

		BeforeEach(func() {
			ports = "53"
			description = "SG"
			someSecurityGroup = helpers.NewSecurityGroup(secGroupName, "tcp", "0.0.0.0/0", &ports, &description)
			helpers.CreateSecurityGroup(someSecurityGroup)
			session := helpers.CF("bind-running-security-group", secGroupName)
			Eventually(session).Should(Exit(0))
		})

		It("it unbinds the running security group globally", func() {
			session := helpers.CF("unbind-running-security-group", secGroupName)
			userName, _ := helpers.GetCredentials()
			Eventually(session).Should(Say("Unbinding security group %s from defaults for running as %s...", secGroupName, userName))
			Eventually(session).Should(Say("OK"))
			Eventually(session).Should(Say(`TIP: Changes require an app restart \(for running\) or restage \(for staging\) to apply to existing applications\.`))
			Eventually(session).Should(Exit(0))

			session = helpers.CF("unbind-running-security-group", secGroupName)
			Eventually(session).Should(Say("Unbinding security group %s from defaults for running as %s...", secGroupName, userName))
			Eventually(session).Should(Say("OK"))
			Eventually(session).Should(Say(`TIP: Changes require an app restart \(for running\) or restage \(for staging\) to apply to existing applications\.`))
			Eventually(session).Should(Exit(0))
		})
	})
})
