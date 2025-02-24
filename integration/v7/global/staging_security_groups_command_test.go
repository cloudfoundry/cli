package global

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/resources"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("staging-security-groups command", func() {
	Describe("help", func() {
		When("--help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("staging-security-groups", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("staging-security-groups - List security groups globally configured for staging applications"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf staging-security-groups"))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("bind-staging-security-group, security-group, unbind-staging-security-group"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(false, false, "", "staging-security-groups")
		})
	})

	When("the environment is set up correctly", func() {
		var userName string

		BeforeEach(func() {
			userName = helpers.LoginCF()
		})

		When("staging security groups exists", func() {
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
				session := helpers.CF(`bind-staging-security-group`, securityGroup.Name)
				Eventually(session).Should(Exit(0))
			})

			AfterEach(func() {
				helpers.DeleteSecurityGroup(securityGroup)
				helpers.QuickDeleteOrg(orgName)
			})

			It("displays the globally enabled staging security groups exits 0", func() {
				session := helpers.CF("staging-security-groups")

				Eventually(session).Should(Say(`Getting global staging security groups as %s\.\.\.`, userName))
				Eventually(session).Should(Say(`name`))
				Eventually(session).Should(Say(`public_networks`))
				Eventually(session).Should(Say(`dns`))
				Eventually(session).Should(Say(securityGroup.Name))

				Eventually(session).Should(Exit(0))
			})
		})
	})
})
