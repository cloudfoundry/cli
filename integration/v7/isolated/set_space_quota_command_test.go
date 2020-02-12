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
		userName  string
	)
	BeforeEach(func() {
		orgName = helpers.NewOrgName()
		spaceName = helpers.NewSpaceName()

		helpers.SetupCF(orgName, spaceName)
		quotaName = helpers.QuotaName()
		session := helpers.CF("create-space-quota", quotaName)
		Eventually(session).Should(Exit(0))
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, false, orgName, "set-space-quota", spaceName, quotaName)
		})
	})

	When("the environment is setup correctly", func() {
		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		BeforeEach(func() {
			userName, _ = helpers.GetCredentials()
		})

		Describe("help", func() {
			When("--help flag is set", func() {
				It("appears in cf help -a", func() {
					session := helpers.CF("help", "-a")
					Eventually(session).Should(Exit(0))
					Expect(session).To(HaveCommandInCategoryWithDescription("set-space-quota", "SPACE ADMIN", "Assign a quota to a space"))
				})

				It("Displays command usage to output", func() {
					session := helpers.CF("set-space-quota", "--help")
					Eventually(session).Should(Say("NAME:"))
					Eventually(session).Should(Say("set-space-quota - Assign a quota to a space"))
					Eventually(session).Should(Say("USAGE:"))
					Eventually(session).Should(Say("cf set-space-quota SPACE QUOTA"))
					Eventually(session).Should(Say("SEE ALSO:"))
					Eventually(session).Should(Say("space, space-quotas, spaces"))
					Eventually(session).Should(Exit(0))
				})
			})
		})

		When("valid arguments are provided", func() {
			It("sets the quota on a space", func() {
				session := helpers.CF("set-space-quota", spaceName, quotaName)
				Eventually(session).Should(Say("Setting space quota %s to space %s as %s...", quotaName, spaceName, userName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))

				session = helpers.CF("space", spaceName)
				Eventually(session).Should(Say(`(?i)quota:\s+%s`, quotaName))
				Eventually(session).Should(Exit(0))
			})

			When("the quota is already applied to the space", func() {
				BeforeEach(func() {
					session := helpers.CF("set-space-quota", spaceName, quotaName)
					Eventually(session).Should(Exit(0))
				})

				It("sets the quota on the space", func() {
					session := helpers.CF("set-space-quota", spaceName, quotaName)
					Eventually(session).Should(Say("Setting space quota %s to space %s as %s...", quotaName, spaceName, userName))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))
				})
			})
		})

		When("invalid arguments are provided", func() {
			It("fails and informs the user an invalid quota was provided", func() {
				session := helpers.CF("set-space-quota", spaceName, "fake-name")
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("Space quota with name '%s' not found.", "fake-name"))
				Eventually(session).Should(Exit(1))
			})

			It("fails and informs the user an invalid space was provided", func() {
				session := helpers.CF("set-space-quota", "fake-name", quotaName)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("Space 'fake-name' not found."))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the incorrect number of positional arguments are provided", func() {
			It("fails and informs the user a positional argument is missing", func() {
				session := helpers.CF("set-space-quota", spaceName)
				Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `QUOTA` was not provided"))
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Exit(1))
			})

			It("fails and reminds the user only two positional args are needed", func() {
				session := helpers.CF("set-space-quota", spaceName, quotaName, "extra")
				Eventually(session.Err).Should(Say(`Incorrect Usage: unexpected argument "extra"`))
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Exit(1))
			})
		})
	})
})
