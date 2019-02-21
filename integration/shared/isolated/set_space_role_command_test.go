package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("set-space-role command", func() {
	Describe("help", func() {
		When("--help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("set-space-role", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("set-space-role - Assign a space role to a user"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf set-space-role USERNAME ORG SPACE ROLE"))
				Eventually(session).Should(Say("ROLES:"))
				Eventually(session).Should(Say("'SpaceManager' - Invite and manage users, and enable features for a given space"))
				Eventually(session).Should(Say("'SpaceDeveloper' - Create and manage apps and services, and see logs and reports"))
				Eventually(session).Should(Say("'SpaceAuditor' - View logs, reports, and settings on this space"))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("space-users"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the org, space, and user all exist", func() {
		var (
			username  string
			orgName   string
			spaceName string
		)

		BeforeEach(func() {
			helpers.LoginCF()
			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()
			helpers.CreateOrgAndSpace(orgName, spaceName)
			username, _ = helpers.CreateUser()
		})

		It("sets the space role for the user", func() {
			session := helpers.CF("set-space-role", username, orgName, spaceName, "SpaceAuditor")
			Eventually(session).Should(Say("Assigning role RoleSpaceAuditor to user %s in org %s / space %s as admin...", username, orgName, spaceName))
			Eventually(session).Should(Say("OK"))
			Eventually(session).Should(Exit(0))
		})

		When("the logged in user has insufficient permissions", func() {
			BeforeEach(func() {
				helpers.SwitchToSpaceRole(orgName, spaceName, "SpaceAuditor")
			})

			It("prints out the error message from CC API and exits 1", func() {
				session := helpers.CF("set-space-role", username, orgName, spaceName, "SpaceAuditor")
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Say("Server error, status code: 403, error code: 10003, message: You are not authorized to perform the requested action"))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the user already has the desired role", func() {
			BeforeEach(func() {
				session := helpers.CF("set-space-role", username, orgName, spaceName, "SpaceDeveloper")
				Eventually(session).Should(Say("Assigning role RoleSpaceDeveloper to user %s in org %s / space %s as admin...", username, orgName, spaceName))
				Eventually(session).Should(Exit(0))
			})

			It("is idempotent", func() {
				session := helpers.CF("set-space-role", username, orgName, spaceName, "SpaceDeveloper")
				Eventually(session).Should(Say("Assigning role RoleSpaceDeveloper to user %s in org %s / space %s as admin...", username, orgName, spaceName))
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
