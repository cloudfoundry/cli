package global

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/resources"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("running-security-groups command", func() {
	Describe("help", func() {
		When("--help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("running-security-groups", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("running-security-groups - List security groups globally configured for running applications"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf running-security-groups"))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("bind-running-security-group, security-group, unbind-running-security-group"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(false, false, "", "running-security-groups")
		})
	})

	When("the environment is set up correctly", func() {
		var userName string

		BeforeEach(func() {
			userName = helpers.LoginCF()
		})

		When("running security groups exists", func() {
			var (
				securityGroup resources.SecurityGroup
				orgName       string
				spaceName     string
				ports         string
				description   string
			)

			BeforeEach(func() {
				orgName = helpers.NewOrgName()
				spaceName = helpers.NewSpaceName()
				helpers.CreateOrg(orgName)
				helpers.TargetOrg(orgName)
				helpers.CreateSpace(spaceName)

				ports = "3360"
				description = "Test security group"
				securityGroup = helpers.NewSecurityGroup(
					helpers.PrefixedRandomName("INTEGRATION-SECURITY-GROUP"),
					"tcp",
					"10.244.1.18",
					&ports,
					&description,
				)
				helpers.CreateSecurityGroup(securityGroup)
				session := helpers.CF(`bind-running-security-group`, securityGroup.Name)
				Eventually(session).Should(Exit(0))
			})

			AfterEach(func() {
				helpers.DeleteSecurityGroup(securityGroup)
				helpers.QuickDeleteOrg(orgName)
			})

			It("displays the globally enabled running security groups exits 0", func() {
				session := helpers.CF("running-security-groups")

				Eventually(session).Should(Say(`Getting global running security groups as %s\.\.\.`, userName))
				Eventually(session).Should(Say(`name`))
				Eventually(session).Should(Say(`public_networks`))
				Eventually(session).Should(Say(`dns`))
				Eventually(session).Should(Say(securityGroup.Name))

				Eventually(session).Should(Exit(0))
			})
		})

	})
})
