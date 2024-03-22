package global

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"

	"code.cloudfoundry.org/cli/integration/helpers"
)

var _ = Describe("set-org-role command", func() {
	var orgName string

	When("the set_roles_by_username flag is disabled", func() {
		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			helpers.LoginCF()
			helpers.CreateOrg(orgName)
			helpers.DisableFeatureFlag("set_roles_by_username")
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
			helpers.EnableFeatureFlag("set_roles_by_username")
		})

		When("the user does not exist", func() {
			It("prints the error from UAA and exits 1", func() {
				session := helpers.CF("set-org-role", "not-exists", orgName, "OrgAuditor")
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("No user exists with the username 'not-exists'."))
				Eventually(session).Should(Exit(1))
			})
		})
	})
})
