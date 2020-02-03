package isolated

import (
	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("set-quota command", func() {
	var (
		orgName   string
		quotaName string
	)
	BeforeEach(func() {
		orgName = helpers.NewOrgName()

		helpers.SetupCFWithOrgOnly(orgName)
		quotaName = helpers.QuotaName()
		session := helpers.CF("create-quota", quotaName)
		Eventually(session).Should(Exit(0))
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(false, false, "test-org", "set-org-quota", orgName, quotaName)
		})
	})

	When("the environment is setup correctly", func() {
		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		Describe("help", func() {
			When("--help flag is set", func() {
				It("appears in cf help -a", func() {
					session := helpers.CF("help", "-a")
					Eventually(session).Should(Exit(0))
					Expect(session).To(HaveCommandInCategoryWithDescription("set-org-quota", "ORG ADMIN", "Assign a quota to an org"))
				})

				It("Displays command usage to output", func() {
					session := helpers.CF("set-org-quota", "--help")
					Eventually(session).Should(Say("NAME:"))
					Eventually(session).Should(Say("set-org-quota - Assign a quota to an org"))
					Eventually(session).Should(Say("USAGE:"))
					Eventually(session).Should(Say("cf set-org-quota ORG QUOTA"))
					Eventually(session).Should(Say("TIP:"))
					Eventually(session).Should(Say("View allowable quotas with 'cf quotas'"))
					Eventually(session).Should(Say("SEE ALSO:"))
					Eventually(session).Should(Say("orgs, quotas"))
					Eventually(session).Should(Exit(0))
				})
			})
		})

		When("valid arguments are provided", func() {
			It("sets the quota on an org", func() {
				session := helpers.CF("set-org-quota", orgName, quotaName)
				Eventually(session).Should(Say("Setting quota %s to org %s", quotaName, orgName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))

				session = helpers.CF("org", orgName)
				Eventually(session).Should(Say(`(?i)quota:\s+%s`, quotaName))
				Eventually(session).Should(Exit(0))
			})

			When("the relationship already exists", func() {
				BeforeEach(func() {
					session := helpers.CF("set-org-quota", orgName, quotaName)
					Eventually(session).Should(Exit(0))
				})

				It("sets the quota on a org", func() {
					session := helpers.CF("set-org-quota", orgName, quotaName)
					Eventually(session).Should(Say("Setting quota %s to org %s", quotaName, orgName))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))
				})
			})
		})

		When("invalid arguments are provided", func() {
			It("fails and informs the user an invalid quota was provided", func() {
				session := helpers.CF("set-org-quota", orgName, "fake-name")
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Say("Quota fake-name not found"))
				Eventually(session).Should(Exit(1))
			})

			It("fails and informs the user an invalid org was provided", func() {
				session := helpers.CF("set-org-quota", "fake-name", quotaName)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Say("Organization fake-name not found"))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the incorrect number of positional arguments are provided", func() {
			It("fails and informs the user a positional argument is missing", func() {
				session := helpers.CF("set-org-quota", orgName)
				Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `QUOTA` was not provided"))
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Exit(1))
			})

			It("fails and reminds the user only two positional args are needed", func() {
				session := helpers.CF("set-org-quota", orgName, quotaName, "extra")
				Eventually(session).Should(Say("Incorrect Usage. Requires ORG_NAME, QUOTA as arguments"))
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Exit(1))
			})
		})
	})
})
