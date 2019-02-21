package isolated

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"

	"code.cloudfoundry.org/cli/integration/helpers"
)

var _ = Describe("set-org-role command", func() {
	var orgName string

	BeforeEach(func() {
		helpers.LoginCF()
		orgName = helpers.NewOrgName()
		helpers.CreateOrg(orgName)
	})

	When("the org and user both exist", func() {
		var username string

		BeforeEach(func() {
			username, _ = helpers.CreateUser()
		})

		It("sets the org role for the user", func() {
			session := helpers.CF("set-org-role", username, orgName, "OrgAuditor")
			Eventually(session).Should(Say("Assigning role OrgAuditor to user %s in org %s as admin...", username, orgName))
			Eventually(session).Should(Say("OK"))
			Eventually(session).Should(Exit(0))
		})

		When("the logged in user has insufficient permissions", func() {
			BeforeEach(func() {
				helpers.SwitchToOrgRole(orgName, "OrgAuditor")
			})

			It("prints out the error message from CC API and exits 1", func() {
				session := helpers.CF("set-org-role", username, orgName, "OrgAuditor")
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Say("Server error, status code: 403, error code: 10003, message: You are not authorized to perform the requested action"))
				Eventually(session).Should(Exit(1))
			})
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

	When("the user does not exist", func() {
		It("prints an appropriate error and exits 1", func() {
			session := helpers.CF("set-org-role", "not-exists", orgName, "OrgAuditor")
			Eventually(session).Should(Say("Assigning role OrgAuditor to user not-exists in org %s as admin...", orgName))
			Eventually(session).Should(Say("FAILED"))
			Eventually(session).Should(Say("Server error, status code: 404, error code: 20003, message: The user could not be found: not-exists"))
			Eventually(session).Should(Exit(1))
		})
	})
})
