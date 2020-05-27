package global

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/resources"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("security-group command", func() {
	Describe("help", func() {
		When("--help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("security-group", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("security-group - Show a single security group"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf security-group SECURITY_GROUP"))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("bind-running-security-group, bind-security-group, bind-staging-security-group"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(false, false, "", "security-group", "bogus")
		})
	})

	When("the environment is set up correctly", func() {
		var userName string

		BeforeEach(func() {
			userName = helpers.LoginCF()
		})

		When("the security group does not exist", func() {
			It("displays security group not found and exits 1", func() {
				session := helpers.CF("security-group", "bogus")
				userName, _ := helpers.GetCredentials()
				Eventually(session).Should(Say(`Getting info for security group %s as %s\.\.\.`, "bogus", userName))
				Eventually(session.Err).Should(Say("Security group '%s' not found.", "bogus"))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the security group exists", func() {
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
			})

			AfterEach(func() {
				helpers.DeleteSecurityGroup(securityGroup)
				helpers.QuickDeleteOrg(orgName)
			})

			When("the security group does not have assigned spaces", func() {
				It("displays the security group without assigned spaces and exits 0", func() {
					session := helpers.CF("security-group", securityGroup.Name)

					Eventually(session).Should(Say(`Getting info for security group %s as %s\.\.\.`, securityGroup.Name, userName))
					Eventually(session).Should(Say(`name:\s+%s`, securityGroup.Name))
					Eventually(session).Should(Say(`rules:`))
					Eventually(session).Should(Say(`\[`))
					Eventually(session).Should(Say(`{`))
					Eventually(session).Should(Say(`"protocol": "%s"`, securityGroup.Rules[0].Protocol))
					Eventually(session).Should(Say(`"destination": "%s"`, securityGroup.Rules[0].Destination))
					Eventually(session).Should(Say(`"ports": "%s"`, ports))
					Eventually(session).Should(Say(`"description": "%s"`, description))
					Eventually(session).Should(Say(`}`))
					Eventually(session).Should(Say(`\]`))
					Eventually(session).Should(Say(`No spaces assigned`))

					Eventually(session).Should(Exit(0))
				})
			})

			When("the security group has assigned spaces", func() {
				BeforeEach(func() {
					session := helpers.CF("bind-security-group", securityGroup.Name, orgName, "--space", spaceName)
					Eventually(session).Should(Exit(0))
				})

				It("displays the security group with assigned spaces and exits 0", func() {
					session := helpers.CF("security-group", securityGroup.Name)

					Eventually(session).Should(Say(`Getting info for security group %s as %s\.\.\.`, securityGroup.Name, userName))
					Eventually(session).Should(Say(`name:\s+%s`, securityGroup.Name))
					Eventually(session).Should(Say(`rules:`))
					Eventually(session).Should(Say(`\[`))
					Eventually(session).Should(Say(`{`))
					Eventually(session).Should(Say(`"protocol": "%s"`, securityGroup.Rules[0].Protocol))
					Eventually(session).Should(Say(`"destination": "%s"`, securityGroup.Rules[0].Destination))
					Eventually(session).Should(Say(`"ports": "%s"`, ports))
					Eventually(session).Should(Say(`"description": "%s"`, description))
					Eventually(session).Should(Say(`}`))
					Eventually(session).Should(Say(`\]`))
					Eventually(session).Should(Say(`organization\s+space`))
					Eventually(session).Should(Say(`%s\s+%s`, orgName, spaceName))

					Eventually(session).Should(Exit(0))
				})
			})
		})
	})
})
