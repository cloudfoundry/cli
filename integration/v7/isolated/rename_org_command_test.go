package isolated

import (
	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("rename-org command", func() {
	var (
		orgName    string
		orgNameNew string
	)

	BeforeEach(func() {
		orgName = helpers.NewOrgName()
		orgNameNew = helpers.NewOrgName()
	})

	Describe("help", func() {
		When("--help flag is set", func() {
			It("appears in cf help -a", func() {
				session := helpers.CF("help", "-a")
				Eventually(session).Should(Exit(0))
				Expect(session).To(HaveCommandInCategoryWithDescription("rename-org", "ORGS", "Rename an org"))
			})

			It("Displays command usage to output", func() {
				session := helpers.CF("rename-org", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("rename-org - Rename an org"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say(`cf rename-org ORG NEW_ORG_NAME`))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("orgs, quotas, set-org-role"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the org name is not provided", func() {
		It("tells the user that the org name is required, prints help text, and exits 1", func() {
			session := helpers.CF("rename-org")

			Eventually(session.Err).Should(Say("Incorrect Usage: the required arguments `ORG` and `NEW_ORG_NAME` were not provided"))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	When("the new org name is not provided", func() {
		It("tells the user that the org name is required, prints help text, and exits 1", func() {
			session := helpers.CF("rename-org", "org")

			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `NEW_ORG_NAME` was not provided"))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(false, false, "", "rename-org", orgName, orgNameNew)
		})
	})

	When("the environment is set up correctly", func() {
		BeforeEach(func() {
			helpers.LoginCF()
		})

		When("org does not exist", func() {
			It("tells the user that the org does not exist, prints help text, and exits 1", func() {
				session := helpers.CF("rename-org", "not-an-org", orgNameNew)

				Eventually(session.Err).Should(Say("Organization 'not-an-org' not found."))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the org does exist", func() {
			BeforeEach(func() {
				helpers.CreateOrg(orgName)
			})

			AfterEach(func() {
				helpers.QuickDeleteOrg(orgNameNew)
			})

			It("renames the org", func() {
				session := helpers.CF("rename-org", orgName, orgNameNew)
				userName, _ := helpers.GetCredentials()
				Eventually(session).Should(Say("Renaming org %s to %s as %s...", orgName, orgNameNew, userName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))

				session = helpers.CF("org", orgNameNew)
				Eventually(session).Should(Say(`name:\s+%s`, orgNameNew))
				Eventually(session).Should(Exit(0))
			})

			When("the new name is already taken", func() {
				BeforeEach(func() {
					helpers.CreateOrg(orgNameNew)
				})

				It("fails to rename the org", func() {
					session := helpers.CF("rename-org", orgName, orgNameNew)
					userName, _ := helpers.GetCredentials()
					Eventually(session).Should(Say("Renaming org %s to %s as %s...", orgName, orgNameNew, userName))
					Eventually(session.Err).Should(Say(`Organization name '%s' is already taken\.`, orgNameNew))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				})
			})
		})
	})
})
