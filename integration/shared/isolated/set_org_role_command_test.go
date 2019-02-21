package isolated

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"

	"code.cloudfoundry.org/cli/integration/helpers"
)

var _ = Describe("set-org-role command", func() {
	When("the org and user both exist", func() {
		var (
			username string
			orgName  string
		)

		BeforeEach(func() {
			helpers.LoginCF()
			orgName = helpers.NewOrgName()
			helpers.CreateOrg(orgName)
			username, _ = helpers.CreateUser()
		})

		It("sets the org role for the user", func() {
			session := helpers.CF("set-org-role", username, orgName, "OrgAuditor")
			Eventually(session).Should(Say("Assigning role OrgAuditor to user %s in org %s as admin...", username, orgName))
			Eventually(session).Should(Say("OK"))
			Eventually(session).Should(Exit(0))
		})

		When("the user already has the desired role", func() {
			BeforeEach(func() {
				session := helpers.CF("set-org-role", username, orgName, "OrgManager")
				Eventually(session).Should(Say("Assigning role OrgManager to user %s in org %s as admin...", username, orgName))
				Eventually(session).Should(Exit(0))
			})

			It("is idempotent", func() {
				session := helpers.CF("set-org-role", username, orgName, "OrgManager")
				Eventually(session).Should(Say("Assigning role OrgManager to user %s in org %s as admin...", username, orgName))
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
