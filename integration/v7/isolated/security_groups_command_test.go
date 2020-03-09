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
		When("--help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("security-groups", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("security-groups - List all security groups"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf security-groups"))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("bind-running-security-group, bind-security-group, bind-staging-security-group, security-group"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(false, false, "", "security-groups")
		})
	})

	When("the environment is set up correctly", func() {
		var userName string

		BeforeEach(func() {
			userName = helpers.LoginCF()
		})

		When("security groups exists", func() {
			var (
				securityGroup1 helpers.SecurityGroup
				securityGroup2 helpers.SecurityGroup
				securityGroup3 helpers.SecurityGroup
				securityGroup4 helpers.SecurityGroup
				securityGroup5 helpers.SecurityGroup
				orgName        string
				spaceName      string
				ports          string
				description    string
			)

			BeforeEach(func() {
				orgName = helpers.NewOrgName()
				spaceName = helpers.NewSpaceName()
				helpers.CreateOrg(orgName)
				helpers.TargetOrg(orgName)
				helpers.CreateSpace(spaceName)

				ports = "3360"
				description = "Test security group"
				securityGroup1 = helpers.NewSecurityGroup(
					helpers.PrefixedRandomName("INTEGRATION-SECURITY-GROUP"),
					"tcp",
					"10.244.1.18",
					&ports,
					&description,
				)
				securityGroup2 = helpers.NewSecurityGroup(
					helpers.PrefixedRandomName("INTEGRATION-SECURITY-GROUP"),
					"udp",
					"127.0.0.1",
					&ports,
					&description,
				)
				securityGroup3 = helpers.NewSecurityGroup(
					helpers.PrefixedRandomName("INTEGRATION-SECURITY-GROUP"),
					"all",
					"0.0.0.0-5.6.7.8",
					nil,
					&description,
				)
				securityGroup4 = helpers.NewSecurityGroup(
					helpers.PrefixedRandomName("INTEGRATION-SECURITY-GROUP"),
					"all",
					"192.168.5.6",
					nil,
					&description,
				)
				securityGroup5 = helpers.NewSecurityGroup(
					helpers.PrefixedRandomName("INTEGRATION-SECURITY-GROUP"),
					"tcp",
					"172.16.0.1",
					&ports,
					&description,
				)
				securityGroup1.Create()
				securityGroup2.Create()
				securityGroup3.Create()
				securityGroup4.Create()
				securityGroup5.Create()

				session1 := helpers.CF(`bind-running-security-group`, securityGroup1.Name)
				session2 := helpers.CF("bind-security-group", securityGroup2.Name, orgName, spaceName)
				session3 := helpers.CF(`bind-staging-security-group`, securityGroup4.Name)
				session4 := helpers.CF("bind-security-group", securityGroup5.Name, orgName, spaceName, "--lifecycle", "staging")

				Eventually(session1).Should(Exit(0))
				Eventually(session2).Should(Exit(0))
				Eventually(session3).Should(Exit(0))
				Eventually(session4).Should(Exit(0))
			})

			AfterEach(func() {
				securityGroup1.Delete()
				securityGroup2.Delete()
				securityGroup3.Delete()
				securityGroup4.Delete()
				securityGroup5.Delete()
			})

			It("displays the security groups exits 0", func() {
				session := helpers.CF("security-groups")

				Eventually(session).Should(Say(`Getting security groups as %s\.\.\.`, userName))
				Eventually(session).Should(Say(`name\s+organization\s+space\s+lifecycle`))
				Eventually(session).Should(Say(`%s\s+<all>\s+<all>\s+running`, securityGroup1.Name))
				Eventually(session).Should(Say(`%s\s+%s\s+%s\s+running`, securityGroup2.Name, orgName, spaceName))
				Eventually(session).Should(Say(`%s\s+`, securityGroup3.Name))
				Eventually(session).Should(Say(`%s\s+<all>\s+<all>\s+staging`, securityGroup4.Name))
				Eventually(session).Should(Say(`%s\s+%s\s+%s\s+staging`, securityGroup5.Name, orgName, spaceName))

				Eventually(session).Should(Exit(0))
			})
		})
	})
})
