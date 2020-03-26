package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/resources"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("bind-running-security-group command", func() {
	var (
		userName     string
		secGroupName string
		helpText     func(session *Session)
	)

	BeforeEach(func() {
		secGroupName = helpers.NewSecurityGroupName()
		userName = helpers.LoginCF()
		helpText = func(session *Session) {
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Say(`\s+bind-running-security-group - Bind a security group to the list of security groups to be used for running applications`))
			Eventually(session).Should(Say("USAGE:"))
			Eventually(session).Should(Say(`\s+cf bind-running-security-group SECURITY_GROUP`))
			Eventually(session).Should(Say("SEE ALSO:"))
			Eventually(session).Should(Say(`\s+apps, bind-security-group, bind-staging-security-group, restart, running-security-groups, security-groups`))
		}
	})

	Describe("help", func() {
		When("--help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("bind-running-security-group", "--help")
				helpText(session)
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(false, false, ReadOnlyOrg, "bind-running-security-group", secGroupName)
		})
	})

	When("the security group is not provided", func() {
		It("fails with an incorrect usage message and displays help", func() {
			session := helpers.CF("bind-running-security-group")
			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `SECURITY_GROUP` was not provided"))
			helpText(session)
			Eventually(session).Should(Exit(1))
		})
	})

	When("the security group doesn't exist", func() {
		It("fails with a security group not found message", func() {
			session := helpers.CF("bind-running-security-group", "some-security-group-that-doesn't-exist")
			Eventually(session.Err).Should(Say("Security group 'some-security-group-that-doesn't-exist' not found"))
			Eventually(session).Should(Say("FAILED"))
			Eventually(session).Should(Exit(1))
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
		})

		It("globally binds the security group to running lifecycle", func() {
			session := helpers.CF("bind-running-security-group", secGroupName)
			Eventually(session).Should(Say("Binding security group %s to running as %s", secGroupName, userName))
			Eventually(session).Should(Say("OK"))
			Eventually(session).Should(Exit(0))
		})
	})
})
