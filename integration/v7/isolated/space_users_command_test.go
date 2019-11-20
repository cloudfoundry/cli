package isolated

import (
	"regexp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"

	"code.cloudfoundry.org/cli/integration/helpers"
)

var _ = Describe("space-users command", func() {
	Describe("help", func() {
		When("--help flag is set", func() {
			It("displays command usage to output", func() {
				session := helpers.CF("space-users", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("space-users - Show space users by role"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say(regexp.QuoteMeta("cf space-users ORG SPACE")))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("org-users, orgs, set-space-role, spaces, unset-space-role"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

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
			)

			BeforeEach(func() {
				spaceManagerUser, _ = helpers.CreateUserInSpaceRole(orgName, spaceName, "SpaceManager")
				spaceDeveloperUser, _ = helpers.CreateUserInSpaceRole(orgName, spaceName, "SpaceDeveloper")
				spaceAuditorUser1, _ = helpers.CreateUserInSpaceRole(orgName, spaceName, "SpaceAuditor")
			})

			It("prints the users in the target space under their roles", func() {
				session := helpers.CF("space-users", orgName, spaceName)
				Eventually(session).Should(Say("Getting users in org %s / space %s as %s", orgName, spaceName, adminUsername))
				Eventually(session).Should(Say("SPACE MANAGER"))
				Eventually(session).Should(Say(`\s+%s \(uaa\)`, spaceManagerUser))
				Eventually(session).Should(Say("SPACE DEVELOPER"))
				Eventually(session).Should(Say(`\s+%s \(uaa\)`, spaceDeveloperUser))
				Eventually(session).Should(Say("SPACE AUDITOR"))
				Eventually(session).Should(Say(`\s+%s \(uaa\)`, spaceAuditorUser1))
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
