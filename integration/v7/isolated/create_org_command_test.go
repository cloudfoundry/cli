package isolated

import (
	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("create-org command", func() {
	var (
		orgNameNew string
	)

	BeforeEach(func() {
		orgNameNew = helpers.NewOrgName()
	})

	Describe("help", func() {
		When("--help flag is set", func() {
			It("appears in cf help -a", func() {
				session := helpers.CF("help", "-a")
				Eventually(session).Should(Exit(0))
				Expect(session).To(HaveCommandInCategoryWithDescription("create-org", "ORGS", "Create an org"))
			})

			It("Displays command usage to output", func() {
				session := helpers.CF("create-org", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("create-org - Create an org"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say(`cf create-org ORG \[-q ORG_QUOTA\]`))
				Eventually(session).Should(Say("ALIAS:"))
				Eventually(session).Should(Say(`co`))
				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say(`--quota, -q\s+Quota to assign to the newly created org`))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("create-space, org-quotas, orgs, set-org-role"))
				Eventually(session).Should(Exit(0))
			})
		})

		When("the org name is not provided", func() {
			It("tells the user that the org name is required, prints help text, and exits 1", func() {
				session := helpers.CF("create-org")

				Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `ORG` was not provided"))
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Exit(1))
			})
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(false, false, ReadOnlyOrg, "create-org", orgNameNew)
		})
	})

	When("the environment is set up correctly", func() {
		BeforeEach(func() {
			helpers.LoginCF()
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgNameNew)
		})

		When("the org does not exist", func() {
			It("creates the org", func() {
				session := helpers.CF("create-org", orgNameNew)
				userName, _ := helpers.GetCredentials()
				Eventually(session).Should(Say("Creating org %s as %s...", orgNameNew, userName))
				Eventually(session).Should(Say("OK"))

				Eventually(session).Should(Say("Assigning role OrgManager to user %s in org %s as %s...", userName, orgNameNew, userName))
				Eventually(session).Should(Say("OK"))

				Eventually(session).Should(Say(`TIP: Use 'cf target -o "%s"' to target new org`, orgNameNew))
				Eventually(session).Should(Exit(0))

				session = helpers.CF("org", orgNameNew)
				Eventually(session).Should(Say(`name:\s+%s`, orgNameNew))
				Eventually(session).Should(Exit(0))
			})
		})

		When("the org already exists", func() {
			BeforeEach(func() {
				Eventually(helpers.CF("create-org", orgNameNew)).Should(Exit(0))
			})

			It("fails to create the org", func() {
				session := helpers.CF("create-org", orgNameNew)
				userName, _ := helpers.GetCredentials()
				Eventually(session).Should(Say("Creating org %s as %s...", orgNameNew, userName))
				Eventually(session).Should(Say(`Organization '%s' already exists\.`, orgNameNew))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))
			})
		})

		When("the -quota flag is passed", func() {
			When("the quota does not exist", func() {
				var (
					quotaName = "garb-quota"
				)

				It("makes an org with the default quota and informs the user setting the quota has failed", func() {
					session := helpers.CF("create-org", orgNameNew, "-q", quotaName)
					userName, _ := helpers.GetCredentials()
					Eventually(session).Should(Say("Creating org %s as %s...", orgNameNew, userName))
					Eventually(session).Should(Say("OK"))

					Eventually(session).Should(Say("Setting org quota %s to org %s as %s...", quotaName, orgNameNew, userName))
					Eventually(session.Err).Should(Say("Organization quota with name '%s' not found.", quotaName))
					Eventually(session).Should(Say("FAILED"))

					session = helpers.CF("org", orgNameNew)
					Eventually(session).Should(Say(`name:\s+%s`, orgNameNew))
					Eventually(session).Should(Say(`quota:\s+%s`, `default`))
					Eventually(session).Should(Exit(0))
				})
			})

			When("the quota exists", func() {
				var (
					quotaName = helpers.QuotaName()
				)

				BeforeEach(func() {
					Eventually(helpers.CF("create-org-quota", quotaName)).Should(Exit(0))
				})

				It("makes an org with the given quota", func() {
					session := helpers.CF("create-org", orgNameNew, "-q", quotaName)
					userName, _ := helpers.GetCredentials()
					Eventually(session).Should(Say("Creating org %s as %s...", orgNameNew, userName))
					Eventually(session).Should(Say("OK"))

					Eventually(session).Should(Say("Setting org quota %s to org %s as %s...", quotaName, orgNameNew, userName))
					Eventually(session).Should(Say("OK"))

					Eventually(session).Should(Say("Assigning role OrgManager to user %s in org %s as %s...", userName, orgNameNew, userName))
					Eventually(session).Should(Say("OK"))

					Eventually(session).Should(Say(`TIP: Use 'cf target -o "%s"' to target new org`, orgNameNew))
					Eventually(session).Should(Exit(0))

					session = helpers.CF("org", orgNameNew)
					Eventually(session).Should(Say(`name:\s+%s`, orgNameNew))
					Eventually(session).Should(Say(`quota:\s+%s`, quotaName))
					Eventually(session).Should(Exit(0))
				})
			})
		})
	})
})
