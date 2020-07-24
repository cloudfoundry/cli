package isolated

import (
	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("set-space-quota command", func() {
	var (
		orgName   string
		spaceName string
		quotaName string
	)
	BeforeEach(func() {
		orgName = helpers.NewOrgName()
		spaceName = helpers.NewSpaceName()

		helpers.SetupCF(orgName, spaceName)
		quotaName = helpers.QuotaName()
		session := helpers.CF("create-space-quota", quotaName)
		Eventually(session).Should(Exit(0))
	})

	AfterEach(func() {
		helpers.QuickDeleteOrg(orgName)
	})

	Describe("help", func() {
		When("--help flag is set", func() {
			It("appears in cf help -a", func() {
				session := helpers.CF("help", "-a")
				Eventually(session).Should(Exit(0))
				Expect(session).To(HaveCommandInCategoryWithDescription("set-space-quota", "SPACE ADMIN", "Assign a space quota definition to a space"))
			})

			It("Displays command usage to output", func() {
				session := helpers.CF("set-space-quota", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("set-space-quota - Assign a space quota definition to a space"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say(`cf set-space-quota SPACE_NAME SPACE_QUOTA_NAME`))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("space, space-quotas, spaces"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.LoginCF()
			session := helpers.CF("set-space-quota", spaceName, quotaName)
			Eventually(session).Should(Say("FAILED"))
			Eventually(session.Out).Should(Say("No org targeted, use 'cf target -o ORG' to target an org\\."))
			Eventually(session).Should(Exit(1))
			//helpers.CheckEnvironmentTargetedCorrectly(true, false, orgName, "set-space-quota", spaceName, quotaName)
		})
	})

	It("sets the space quota on a space", func() {
		session := helpers.CF("set-space-quota", spaceName, quotaName)
		Eventually(session).Should(Say("Assigning space quota %s to space %s", quotaName, spaceName))
		Eventually(session).Should(Exit(0))

		session = helpers.CF("space", spaceName)
		Eventually(session).Should(Say(`(?i)space quota:\s+%s`, quotaName))
		Eventually(session).Should(Exit(0))
	})
})
