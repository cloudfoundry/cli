package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("security-groups command", func() {
	Describe("help", func() {
		Context("when --help flag is provided", func() {
			It("displays command usage to output", func() {
				session := helpers.CF("security-groups", "--help")
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

	Describe("Unrefactored command", func() {
		var (
			username string
		)

		BeforeEach(func() {
			helpers.SkipIfExperimental("skipping tests around the old code behavior")
			username = helpers.LoginCF()
		})

		Context("when no API endpoint is set", func() {
			BeforeEach(func() {
				helpers.UnsetAPI()
			})

			It("fails with no API endpoint set message", func() {
				session := helpers.CF("security-groups")
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("No API endpoint set\\. Use 'cf login' or 'cf api' to target an endpoint\\."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when not logged in", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
			})

			It("fails with not logged in message", func() {
				session := helpers.CF("security-groups")
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Out).Should(Say("Not logged in\\. Use 'cf login' to log in\\."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when too many arguments are provided", func() {
			It("fails with too many arguments message", func() {
				session := helpers.CF("security-groups", "foooo")
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Out).Should(Say("Incorrect Usage\\. No argument required"))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when there are security groups", func() {
			var (
				orgName        string
				spaceName1     string
				spaceName2     string
				securityGroup1 helpers.SecurityGroup
				securityGroup2 helpers.SecurityGroup
			)

			BeforeEach(func() {
				helpers.ClearTarget()
				orgName = helpers.NewOrgName()
				spaceName1 = helpers.NewSpaceName()
				spaceName2 = helpers.NewSpaceName()
				helpers.CreateOrgAndSpace(orgName, spaceName1)
				helpers.CreateSpace(spaceName2)

				securityGroup1 = helpers.NewSecurityGroup(helpers.NewSecGroupName(), "tcp", "11.1.1.0/24", "80,443", "SG1")
				securityGroup1.Create()
				Eventually(helpers.CF("bind-security-group", securityGroup1.Name, orgName)).Should(Exit(0))

				securityGroup2 = helpers.NewSecurityGroup(helpers.NewSecGroupName(), "tcp", "125.5.1.0/24", "25555", "SG2")
				securityGroup2.Create()
			})

			AfterEach(func() {
				Eventually(helpers.CF("unbind-security-group", securityGroup1.Name, orgName, spaceName1)).Should(Exit(0))
				Eventually(helpers.CF("unbind-security-group", securityGroup1.Name, orgName, spaceName2)).Should(Exit(0))
				securityGroup1.Delete()
				securityGroup2.Delete()
			})

			It("lists the security groups", func() {
				session := helpers.CF("security-groups")
				Eventually(session.Out).Should(Say("Getting security groups as admin"))
				Eventually(session.Out).Should(Say("OK\\n\\n"))
				Eventually(session.Out).Should(Say("\\s+Name\\s+Organization\\s+Space"))
				Eventually(session.Out).Should(Say("#\\d+\\s+%s\\s+%s\\s+%s", securityGroup1.Name, orgName, spaceName1))
				Eventually(session.Out).Should(Say("\\s+%s\\s+%s\\s+%s", securityGroup1.Name, orgName, spaceName2))
				Eventually(session.Out).Should(Say("#\\d+\\s+%s", securityGroup2.Name))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Describe("Refactored command", func() {
		var (
			username string
			session  *Session
		)

		BeforeEach(func() {
			helpers.RunIfExperimental("skipping until approved")
			username = helpers.LoginCF()
		})

		JustBeforeEach(func() {
			session = helpers.CF("security-groups")
		})

		Context("when no API endpoint is set", func() {
			BeforeEach(func() {
				helpers.UnsetAPI()
			})

			It("fails with no API endpoint set message", func() {
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("No API endpoint set. Use 'cf login' or 'cf api' to target an endpoint."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when not logged in", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
			})

			It("fails with not logged in message", func() {
				session := helpers.CF("security-groups")
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("Not logged in\\. Use 'cf login' to log in\\."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when too many arguments are provided", func() {
			It("succeeds and ignores the additional arguments", func() {
				session := helpers.CF("security-groups", "foooo")
				Eventually(session.Out).Should(Say("OK"))
				Eventually(session).Should(Exit(0))
			})
		})

		Context("when there are security groups", func() {
			var (
				securityGroup1 helpers.SecurityGroup
				securityGroup2 helpers.SecurityGroup
				securityGroup3 helpers.SecurityGroup
				securityGroup4 helpers.SecurityGroup

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
				Eventually(helpers.CF("bind-security-group", securityGroup1.Name, org11, space11)).Should(Exit(0))
				helpers.CreateSpace(space22)
				Eventually(helpers.CF("bind-security-group", securityGroup2.Name, org11, space22)).Should(Exit(0))
				helpers.CreateSpace(space32)
				Eventually(helpers.CF("bind-security-group", securityGroup4.Name, org11, space32)).Should(Exit(0))
				helpers.CreateOrgAndSpace(org12, space12)
				Eventually(helpers.CF("bind-security-group", securityGroup1.Name, org12, space12)).Should(Exit(0))
				helpers.CreateOrgAndSpace(org13, space13)
				Eventually(helpers.CF("bind-security-group", securityGroup1.Name, org13, space13)).Should(Exit(0))
				helpers.CreateOrgAndSpace(org21, space21)
				Eventually(helpers.CF("bind-security-group", securityGroup2.Name, org21, space21)).Should(Exit(0))
				helpers.CreateOrgAndSpace(org23, space23)
				Eventually(helpers.CF("bind-security-group", securityGroup2.Name, org23, space23)).Should(Exit(0))
				helpers.CreateSpace(space31)
				Eventually(helpers.CF("bind-security-group", securityGroup4.Name, org23, space31)).Should(Exit(0))
				helpers.CreateOrgAndSpace(org33, space33)
				Eventually(helpers.CF("bind-security-group", securityGroup4.Name, org33, space33)).Should(Exit(0))
			})

			AfterEach(func() {
				// Delete Security Groups, Organizations, and Spaces with predictable and unique names for testing sorting
				Eventually(helpers.CF("unbind-security-group", securityGroup1.Name, org11, space11)).Should(Exit(0))
				Eventually(helpers.CF("unbind-security-group", securityGroup2.Name, org11, space22)).Should(Exit(0))
				Eventually(helpers.CF("unbind-security-group", securityGroup4.Name, org11, space32)).Should(Exit(0))
				Eventually(helpers.CF("unbind-security-group", securityGroup1.Name, org12, space12)).Should(Exit(0))
				Eventually(helpers.CF("unbind-security-group", securityGroup1.Name, org13, space13)).Should(Exit(0))
				Eventually(helpers.CF("unbind-security-group", securityGroup2.Name, org21, space21)).Should(Exit(0))
				Eventually(helpers.CF("unbind-security-group", securityGroup2.Name, org23, space23)).Should(Exit(0))
				Eventually(helpers.CF("unbind-security-group", securityGroup4.Name, org23, space31)).Should(Exit(0))
				Eventually(helpers.CF("unbind-security-group", securityGroup4.Name, org33, space33)).Should(Exit(0))

				securityGroup1.Delete()
				securityGroup2.Delete()
				securityGroup3.Delete()
				securityGroup4.Delete()
			})

			It("lists the security groups", func() {
				Eventually(session.Out).Should(Say("Getting security groups as admin"))
				Eventually(session.Out).Should(Say("OK\\n\\n"))
				Eventually(session.Out).Should(Say("\\s+name\\s+organization\\s+space"))
				// How to test alphabetization with auto-generated names?  Here's how.
				Eventually(session.Out).Should(Say("#\\d+\\s+%s\\s+%s\\s+%s", securityGroup1.Name, org11, space11))
				Eventually(session.Out).Should(Say("\\s+%s\\s+%s\\s+%s", securityGroup1.Name, org12, space12))
				Eventually(session.Out).Should(Say("\\s+%s\\s+%s\\s+%s", securityGroup1.Name, org13, space13))
				Eventually(session.Out).Should(Say("#\\d+\\s+%s\\s+%s\\s+%s", securityGroup2.Name, org11, space22))
				Eventually(session.Out).Should(Say("\\s+%s\\s+%s\\s+%s", securityGroup2.Name, org21, space21))
				Eventually(session.Out).Should(Say("\\s+%s\\s+%s\\s+%s", securityGroup2.Name, org23, space23))
				Eventually(session.Out).Should(Say("#\\d+\\s+%s", securityGroup3.Name))
				Eventually(session.Out).Should(Say("#\\d+\\s+%s\\s+%s\\s+%s", securityGroup4.Name, org11, space32))
				Eventually(session.Out).Should(Say("\\s+%s\\s+%s\\s+%s", securityGroup4.Name, org23, space31))
				Eventually(session.Out).Should(Say("\\s+%s\\s+%s\\s+%s", securityGroup4.Name, org33, space33))
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
