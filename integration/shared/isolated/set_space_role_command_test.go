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
})
