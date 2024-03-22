package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/resources"
	. "github.com/onsi/ginkgo/v2"
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
				securityGroup1 resources.SecurityGroup
				securityGroup2 resources.SecurityGroup
				securityGroup3 resources.SecurityGroup
				securityGroup4 resources.SecurityGroup
				securityGroup5 resources.SecurityGroup
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
				helpers.CreateSecurityGroup(securityGroup1)
				helpers.CreateSecurityGroup(securityGroup2)
				helpers.CreateSecurityGroup(securityGroup3)
				helpers.CreateSecurityGroup(securityGroup4)
				helpers.CreateSecurityGroup(securityGroup5)

				session1 := helpers.CF(`bind-running-security-group`, securityGroup1.Name)
				session2 := helpers.CF("bind-security-group", securityGroup2.Name, orgName, "--space", spaceName)
				session3 := helpers.CF(`bind-staging-security-group`, securityGroup4.Name)
				session4 := helpers.CF("bind-security-group", securityGroup5.Name, orgName, "--space", spaceName, "--lifecycle", "staging")

				Eventually(session1).Should(Exit(0))
				Eventually(session2).Should(Exit(0))
				Eventually(session3).Should(Exit(0))
				Eventually(session4).Should(Exit(0))
			})

			AfterEach(func() {
				helpers.DeleteSecurityGroup(securityGroup1)
				helpers.DeleteSecurityGroup(securityGroup2)
				helpers.DeleteSecurityGroup(securityGroup3)
				helpers.DeleteSecurityGroup(securityGroup4)
				helpers.DeleteSecurityGroup(securityGroup5)
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
