package global

import (
	"strings"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("security group-related command", func() {
	When("there are no security groups", func() {
		var (
			unprivilegedUsername string
			password             string
			orgName              string
			spaceName            string
			runningGroups        []string
			stagingGroups        []string
		)

		BeforeEach(func() {
			helpers.LoginCF()
			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()
			helpers.SetupCF(orgName, spaceName)

			session := helpers.CF("running-security-groups")
			Eventually(session).Should(Exit(0))
			runningGroups = strings.Split(string(session.Out.Contents()), "\n")[2:]
			for _, runningGroup := range runningGroups {
				if runningGroup != "" {
					helpers.CF(`unbind-running-security-group`, runningGroup)
				}
			}
			session = helpers.CF("staging-security-groups")
			Eventually(session).Should(Exit(0))
			stagingGroups = strings.Split(string(session.Out.Contents()), "\n")[2:]
			for _, stagingGroup := range stagingGroups {
				if stagingGroup != "" {
					helpers.CF(`unbind-staging-security-group`, stagingGroup)
				}
			}
			unprivilegedUsername, password = helpers.CreateUserInSpaceRole(orgName, spaceName, "SpaceDeveloper")
			helpers.LogoutCF()
			helpers.LoginAs(unprivilegedUsername, password)
			helpers.TargetOrg(orgName)
		})

		AfterEach(func() {
			helpers.LoginCF()
			for _, runningGroup := range runningGroups {
				if runningGroup != "" {
					helpers.CF(`bind-running-security-group`, runningGroup)
				}
			}
			for _, stagingGroup := range stagingGroups {
				if stagingGroup != "" {
					helpers.CF(`bind-staging-security-group`, stagingGroup)
				}
			}
			helpers.DeleteUser(unprivilegedUsername)
		})

		It("displays no security groups found and exits 0", func() {
			session := helpers.CF("security-groups")
			Eventually(session).Should(Say(`Getting security groups as %s\.\.\.`, unprivilegedUsername))
			Eventually(session).Should(Say("No security groups found."))
			Eventually(session).Should(Exit(0))
		})

		It("displays no global staging security groups found and exits 0", func() {
			session := helpers.CF("staging-security-groups")
			Eventually(session).Should(Say(`Getting global staging security groups as %s\.\.\.`, unprivilegedUsername))
			Eventually(session).Should(Say("No global staging security groups found."))
			Eventually(session).Should(Exit(0))
		})

		It("displays no security groups found on the space and exits 0", func() {
			session := helpers.CF("space", spaceName, "--security-group-rules")
			Eventually(session).Should(Say(`name:\s+%s`, spaceName))
			Eventually(session).Should(Say(`running security groups:\s*\n`))
			Eventually(session).Should(Say(`staging security groups:\s*\n`))
			Eventually(session).Should(Say("No security group rules found."))
			Eventually(session).Should(Exit(0))
		})
	})
})
