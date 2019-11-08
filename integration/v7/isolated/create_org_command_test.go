package isolated

import (
	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = FDescribe("create-org command", func() {
	var (
		orgName    string
		orgNameNew string
		spaceName  string
	)

	BeforeEach(func() {
		orgName = helpers.NewOrgName()
		orgNameNew = helpers.NewOrgName()
		spaceName = helpers.NewSpaceName()
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
				Eventually(session).Should(Say(`cf create-org ORG`))
				Eventually(session).Should(Say("ALIAS:"))
				Eventually(session).Should(Say(`co`))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("create-space, orgs, quotas, set-org-role"))
				Eventually(session).Should(Exit(0))
			})
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

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(false, false, ReadOnlyOrg, "create-org", orgNameNew)
		})
	})

	When("the environment is set up correctly", func() {
		BeforeEach(func() {
			helpers.SetupCF(orgName, spaceName)
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
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
	})
})
