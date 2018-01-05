package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("security-groups command", func() {
	var (
		session *Session
	)

	Describe("help", func() {
		Context("when --help flag is provided", func() {
			It("displays command usage to output", func() {
				session = helpers.CF("security-groups", "--help")
				Eventually(session.Out).Should(Say("NAME:"))
				Eventually(session.Out).Should(Say("security-groups - List all security groups"))
				Eventually(session.Out).Should(Say("USAGE:"))
				Eventually(session.Out).Should(Say("cf security-groups"))
				Eventually(session.Out).Should(Say("SEE ALSO:"))
				Eventually(session.Out).Should(Say("bind-running-security-group, bind-security-group, bind-staging-security-group, security-group"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("when the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(false, false, ReadOnlyOrg, "security-groups")
		})
	})

	Context("when the environment is set up correctly", func() {
		BeforeEach(func() {
			helpers.LoginCF()
		})

		JustBeforeEach(func() {
			session = helpers.CF("security-groups")
		})

		Context("when there are security groups", func() {
			var (
				securityGroup1 helpers.SecurityGroup
				securityGroup2 helpers.SecurityGroup
				securityGroup3 helpers.SecurityGroup
				securityGroup4 helpers.SecurityGroup
				securityGroup5 helpers.SecurityGroup
				securityGroup6 helpers.SecurityGroup
				securityGroup7 helpers.SecurityGroup

				org11 string
				org12 string
				org13 string
				org21 string
				org23 string
				org33 string

				space11 string
				space12 string
				space13 string
				space21 string
				space22 string
				space23 string
				space31 string
				space32 string
				space33 string
			)

			BeforeEach(func() {
				helpers.ClearTarget()

				// Create Security Groups, Organizations, and Spaces with predictable and unique names for testing sorting
				securityGroup1 = helpers.NewSecurityGroup(helpers.PrefixedRandomName("INTEGRATION-SEC-GROUP-1"), "tcp", "11.1.1.0/24", "80,443", "SG1")
				securityGroup1.Create()
				securityGroup2 = helpers.NewSecurityGroup(helpers.PrefixedRandomName("INTEGRATION-SEC-GROUP-2"), "tcp", "11.1.1.0/24", "80,443", "SG1")
				securityGroup2.Create()
				securityGroup3 = helpers.NewSecurityGroup(helpers.PrefixedRandomName("INTEGRATION-SEC-GROUP-3"), "tcp", "11.1.1.0/24", "80,443", "SG1")
				securityGroup3.Create()
				securityGroup4 = helpers.NewSecurityGroup(helpers.PrefixedRandomName("INTEGRATION-SEC-GROUP-4"), "tcp", "11.1.1.0/24", "80,443", "SG1")
				securityGroup4.Create()
				securityGroup5 = helpers.NewSecurityGroup(helpers.PrefixedRandomName("INTEGRATION-SEC-GROUP-5"), "tcp", "11.1.1.0/24", "80,443", "SG1")
				securityGroup5.Create()
				securityGroup6 = helpers.NewSecurityGroup(helpers.PrefixedRandomName("INTEGRATION-SEC-GROUP-6"), "tcp", "11.1.1.0/24", "80,443", "SG1")
				securityGroup6.Create()
				securityGroup7 = helpers.NewSecurityGroup(helpers.PrefixedRandomName("INTEGRATION-SEC-GROUP-7"), "tcp", "11.1.1.0/24", "80,443", "SG1")
				securityGroup7.Create()

				org11 = helpers.PrefixedRandomName("INTEGRATION-ORG-11")
				org12 = helpers.PrefixedRandomName("INTEGRATION-ORG-12")
				org13 = helpers.PrefixedRandomName("INTEGRATION-ORG-13")
				org21 = helpers.PrefixedRandomName("INTEGRATION-ORG-21")
				org23 = helpers.PrefixedRandomName("INTEGRATION-ORG-23")
				org33 = helpers.PrefixedRandomName("INTEGRATION-ORG-33")

				space11 = helpers.PrefixedRandomName("INTEGRATION-SPACE-11")
				space12 = helpers.PrefixedRandomName("INTEGRATION-SPACE-12")
				space13 = helpers.PrefixedRandomName("INTEGRATION-SPACE-13")
				space21 = helpers.PrefixedRandomName("INTEGRATION-SPACE-21")
				space22 = helpers.PrefixedRandomName("INTEGRATION-SPACE-22")
				space23 = helpers.PrefixedRandomName("INTEGRATION-SPACE-23")
				space31 = helpers.PrefixedRandomName("INTEGRATION-SPACE-31")
				space32 = helpers.PrefixedRandomName("INTEGRATION-SPACE-32")
				space33 = helpers.PrefixedRandomName("INTEGRATION-SPACE-33")

				helpers.CreateOrgAndSpace(org11, space11)
				Eventually(helpers.CF("bind-running-security-group", securityGroup1.Name)).Should(Exit(0))
				Eventually(helpers.CF("bind-security-group", securityGroup1.Name, org11, space11)).Should(Exit(0))
				helpers.CreateSpace(space22)
				Eventually(helpers.CF("bind-security-group", securityGroup2.Name, org11, space22, "--lifecycle", "staging")).Should(Exit(0))
				helpers.CreateSpace(space32)
				Eventually(helpers.CF("bind-security-group", securityGroup4.Name, org11, space32)).Should(Exit(0))
				helpers.CreateOrgAndSpace(org12, space12)
				Eventually(helpers.CF("bind-security-group", securityGroup1.Name, org12, space12, "--lifecycle", "staging")).Should(Exit(0))
				helpers.CreateOrgAndSpace(org13, space13)
				Eventually(helpers.CF("bind-staging-security-group", securityGroup2.Name)).Should(Exit(0))
				Eventually(helpers.CF("bind-security-group", securityGroup1.Name, org13, space13)).Should(Exit(0))
				helpers.CreateOrgAndSpace(org21, space21)
				Eventually(helpers.CF("bind-security-group", securityGroup2.Name, org21, space21)).Should(Exit(0))
				helpers.CreateOrgAndSpace(org23, space23)
				Eventually(helpers.CF("bind-security-group", securityGroup2.Name, org23, space23)).Should(Exit(0))
				helpers.CreateSpace(space31)
				Eventually(helpers.CF("bind-security-group", securityGroup4.Name, org23, space31)).Should(Exit(0))
				helpers.CreateOrgAndSpace(org33, space33)
				Eventually(helpers.CF("bind-security-group", securityGroup4.Name, org33, space33)).Should(Exit(0))
				Eventually(helpers.CF("bind-running-security-group", securityGroup5.Name)).Should(Exit(0))
				Eventually(helpers.CF("bind-staging-security-group", securityGroup6.Name)).Should(Exit(0))
				Eventually(helpers.CF("bind-running-security-group", securityGroup7.Name)).Should(Exit(0))
				Eventually(helpers.CF("bind-staging-security-group", securityGroup7.Name)).Should(Exit(0))
			})

			AfterEach(func() {
				helpers.QuickDeleteOrg(org11)
				helpers.QuickDeleteOrg(org12)
				helpers.QuickDeleteOrg(org13)
				helpers.QuickDeleteOrg(org21)
				helpers.QuickDeleteOrg(org23)
				helpers.QuickDeleteOrg(org33)
			})

			It("lists the security groups", func() {
				Eventually(session.Out).Should(Say("Getting security groups as admin"))
				Eventually(session.Out).Should(Say("OK\\n\\n"))
				Eventually(session.Out).Should(Say("\\s+name\\s+organization\\s+space\\s+lifecycle"))
				// How to test alphabetization with auto-generated names?  Here's how.
				Eventually(session.Out).Should(Say("#\\d+\\s+%s\\s+<all>\\s+<all>\\s+running", securityGroup1.Name))
				Eventually(session.Out).Should(Say("\\s+%s\\s+%s\\s+%s\\s+running", securityGroup1.Name, org11, space11))
				Eventually(session.Out).Should(Say("\\s+%s\\s+%s\\s+%s\\s+staging", securityGroup1.Name, org12, space12))
				Eventually(session.Out).Should(Say("\\s+%s\\s+%s\\s+%s\\s+running", securityGroup1.Name, org13, space13))
				Eventually(session.Out).Should(Say("#\\d+\\s+%s\\s+<all>\\s+<all>\\s+staging", securityGroup2.Name))
				Eventually(session.Out).Should(Say("\\s+%s\\s+%s\\s+%s\\s+staging", securityGroup2.Name, org11, space22))
				Eventually(session.Out).Should(Say("\\s+%s\\s+%s\\s+%s\\s+running", securityGroup2.Name, org21, space21))
				Eventually(session.Out).Should(Say("\\s+%s\\s+%s\\s+%s\\s+running", securityGroup2.Name, org23, space23))
				Eventually(session.Out).Should(Say("#\\d+\\s+%s", securityGroup3.Name))
				Eventually(session.Out).Should(Say("#\\d+\\s+%s\\s+%s\\s+%s\\s+running", securityGroup4.Name, org11, space32))
				Eventually(session.Out).Should(Say("\\s+%s\\s+%s\\s+%s\\s+running", securityGroup4.Name, org23, space31))
				Eventually(session.Out).Should(Say("\\s+%s\\s+%s\\s+%s\\s+running", securityGroup4.Name, org33, space33))
				Eventually(session.Out).Should(Say("#\\d+\\s+%s\\s+<all>\\s+<all>\\s+running", securityGroup5.Name))
				Eventually(session.Out).Should(Say("#\\d+\\s+%s\\s+<all>\\s+<all>\\s+staging", securityGroup6.Name))
				Eventually(session.Out).Should(Say("#\\d+\\s+%s\\s+<all>\\s+<all>\\s+running", securityGroup7.Name))
				Eventually(session.Out).Should(Say("\\s+%s\\s+<all>\\s+<all>\\s+staging", securityGroup7.Name))
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
