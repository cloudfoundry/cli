package isolated

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	"regexp"

	"code.cloudfoundry.org/cli/integration/helpers"
)

var _ = Describe("org-users command", func() {
	Describe("help", func() {
		When("--help flag is set", func() {
			It("displays command usage to output", func() {
				session := helpers.CF("org-users", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("org-users - Show org users by role"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say(regexp.QuoteMeta("cf org-users ORG")))
				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say(`--all-users, -a \s+List all users in the org including Org Users`))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("orgs, set-org-role"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the user is logged in", func() {
		var (
			orgName       string
			adminUsername string
		)

		BeforeEach(func() {
			adminUsername = helpers.LoginCF()
			orgName = helpers.NewOrgName()
			helpers.CreateOrg(orgName)
		})

		When("the target org has multiple users with different roles", func() {
			var (
				orgManagerUser1    string
				orgManagerUser2    string
				billingManagerUser string
				orgAuditorUser     string
			)

			BeforeEach(func() {
				orgManagerUser1, _ = helpers.CreateUserInOrgRole(orgName, "OrgManager")
				orgManagerUser2, _ = helpers.CreateUserInOrgRole(orgName, "OrgManager")
				billingManagerUser, _ = helpers.CreateUserInOrgRole(orgName, "BillingManager")
				orgAuditorUser, _ = helpers.CreateUserInOrgRole(orgName, "OrgAuditor")
			})

			It("prints the users in the target org under their roles", func() {
				session := helpers.CF("org-users", orgName)
				Eventually(session).Should(Say("Getting users in org %s as %s", orgName, adminUsername))
				Eventually(session).Should(Say("ORG MANAGER"))
				Eventually(session).Should(Say(`\s+%s (UAA)`, orgManagerUser1))
				Eventually(session).Should(Say(`\s+%s (UAA)`, orgManagerUser2))
				Eventually(session).Should(Say("BILLING MANAGER"))
				Eventually(session).Should(Say(`\s+%s (UAA)`, billingManagerUser))
				Eventually(session).Should(Say("ORG AUDITOR"))
				Eventually(session).Should(Say(`\s+%s (UAA)`, orgAuditorUser))
				Eventually(session).Should(Exit(0))
			})
		})

		When("the target org has a client-credentials user", func() {
			var clientID string

			BeforeEach(func() {
				clientID, _ = helpers.SkipIfClientCredentialsNotSet()
				Eventually(helpers.CF("set-org-role", clientID, orgName, "OrgManager", "--client")).Should(Exit(0))
			})

			It("prints the client-credentials user", func() {
				session := helpers.CF("org-users", orgName)
				Eventually(session).Should(Say("Getting users in org %s as %s", orgName, adminUsername))
				Eventually(session).Should(Say("ORG MANAGER"))
				Eventually(session).Should(Say(`\s+%s \(client\)`, clientID))
				Eventually(session).Should(Say("BILLING MANAGER"))
				Eventually(session).Should(Say(`\s+No BILLING MANAGER found`))
				Eventually(session).Should(Say("ORG AUDITOR"))
				Eventually(session).Should(Say(`\s+No ORG AUDITOR found`))
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
