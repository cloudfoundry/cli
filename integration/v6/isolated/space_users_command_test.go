package isolated

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"

	"code.cloudfoundry.org/cli/integration/helpers"
)

var _ = Describe("space-users command", func() {
	When("the user is logged in", func() {
		var (
			orgName       string
			spaceName     string
			adminUsername string
		)

		BeforeEach(func() {
			adminUsername = helpers.LoginCF()
			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()
			helpers.CreateOrgAndSpace(orgName, spaceName)
		})

		When("the target space has multiple users with different roles", func() {
			var (
				spaceManagerUser   string
				spaceDeveloperUser string
				spaceAuditorUser1  string
				spaceAuditorUser2  string
			)

			BeforeEach(func() {
				spaceManagerUser, _ = helpers.CreateUserInSpaceRole(orgName, spaceName, "SpaceManager")
				spaceDeveloperUser, _ = helpers.CreateUserInSpaceRole(orgName, spaceName, "SpaceDeveloper")
				spaceAuditorUser1, _ = helpers.CreateUserInSpaceRole(orgName, spaceName, "SpaceAuditor")
				spaceAuditorUser2, _ = helpers.CreateUserInSpaceRole(orgName, spaceName, "SpaceAuditor")
			})

			It("prints the users in the target space under their roles", func() {
				session := helpers.CF("space-users", orgName, spaceName)
				Eventually(session).Should(Say("Getting users in org %s / space %s as %s", orgName, spaceName, adminUsername))
				Eventually(session).Should(Say("SPACE MANAGER"))
				Eventually(session).Should(Say(`\s+%s`, spaceManagerUser))
				Eventually(session).Should(Say("SPACE DEVELOPER"))
				Eventually(session).Should(Say(`\s+%s`, spaceDeveloperUser))
				Eventually(session).Should(Say("SPACE AUDITOR"))
				Eventually(session).Should(Say(`\s+%s`, spaceAuditorUser1))
				Eventually(session).Should(Say(`\s+%s`, spaceAuditorUser2))
				Eventually(session).Should(Exit(0))
			})
		})

		When("the target space has a client-credentials user", func() {
			var (
				clientID         string
				spaceManagerUser string
			)

			BeforeEach(func() {
				spaceManagerUser, _ = helpers.CreateUserInSpaceRole(orgName, spaceName, "SpaceManager")
				clientID, _ = helpers.SkipIfClientCredentialsNotSet()
				Eventually(helpers.CF("set-space-role", clientID, orgName, spaceName, "SpaceDeveloper", "--client")).Should(Exit(0))
			})

			It("prints the client-credentials user", func() {
				session := helpers.CF("space-users", orgName, spaceName)
				Eventually(session).Should(Say("Getting users in org %s / space %s as %s", orgName, spaceName, adminUsername))
				Eventually(session).Should(Say("SPACE MANAGER"))
				Eventually(session).Should(Say(`\s+%s`, spaceManagerUser))
				Eventually(session).Should(Say("SPACE DEVELOPER"))
				Eventually(session).Should(Say(`\s+%s \(client\)`, clientID))
				Eventually(session).Should(Say("SPACE AUDITOR"))
				Eventually(session).Should(Say(`\s+No SPACE AUDITOR found`))
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
