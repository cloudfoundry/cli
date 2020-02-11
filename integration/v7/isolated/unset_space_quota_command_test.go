package isolated

import (
	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("unset-space-quota command", func() {
	var (
		orgName        string
		spaceQuotaName string
		spaceName      string
	)

	Describe("help", func() {
		When("--help flag is set", func() {
			It("appears in cf help -a", func() {
				session := helpers.CF("help", "-a")
				Eventually(session).Should(Exit(0))
				Expect(session).To(HaveCommandInCategoryWithDescription("unset-space-quota", "SPACE ADMIN", "Unassign a quota from a space"))
			})

			It("Displays command usage to output", func() {
				session := helpers.CF("unset-space-quota", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("unset-space-quota - Unassign a quota from a space"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say(`cf unset-space-quota SPACE SPACE_QUOTA`))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("space\n"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, false, orgName, "unset-space-quota", spaceName, spaceQuotaName)
		})
	})

	When("the environment is set up correctly", func() {
		var (
			userName string
		)

		BeforeEach(func() {
			userName = helpers.LoginCF()
			orgName = helpers.CreateAndTargetOrg()
			spaceQuotaName = helpers.QuotaName()
			session := helpers.CF("create-space-quota", spaceQuotaName)
			Eventually(session).Should(Exit(0))
			spaceName = helpers.NewSpaceName()
			helpers.CreateSpace(spaceName)
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		When("no arguments are provided", func() {
			It("tells the user that the space and quota are required, prints help text, and exits 1", func() {
				session := helpers.CF("unset-space-quota")

				Eventually(session.Err).Should(Say("Incorrect Usage: the required arguments `SPACE_NAME` and `SPACE_QUOTA` were not provided"))
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Exit(1))
			})
		})

		When("only one argument is provided", func() {
			It("tells the user that the quota name is required, prints help text, and exits 1", func() {
				session := helpers.CF("unset-space-quota", spaceName)

				Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `SPACE_QUOTA` was not provided"))
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Exit(1))
			})
		})

		When("a nonexistent flag is provided", func() {
			It("tells the user that the flag is invalid", func() {
				session := helpers.CF("unset-space-quota", "--nonexistent")

				Eventually(session.Err).Should(Say("Incorrect Usage: unknown flag `nonexistent'"))
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the space does not exist", func() {
			It("fails and reports that the space could not be found", func() {
				session := helpers.CF("unset-space-quota", "bad-space-name", spaceQuotaName)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say(`Space 'bad-space-name' not found\.`))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the quota does not exist", func() {
			It("fails and reports that the quota could not be found", func() {
				session := helpers.CF("unset-space-quota", spaceName, "bad-quota-name")
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say(`Space quota with name 'bad-quota-name' not found\.`))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the quota is not associated with the space", func() {
			It("succeeds", func() {
				session := helpers.CF("unset-space-quota", spaceName, spaceQuotaName)
				Eventually(session).Should(Say(`Unassigning space quota %s from space %s as %s\.\.\.`, spaceQuotaName, spaceName, userName))
				Eventually(session.Err).Should(Say("Unable to remove quota from space with guid '[a-z0-9-]+'. Ensure the space quota is applied to this space."))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the quota is associated with the space", func() {
			BeforeEach(func() {
				session := helpers.CF("set-space-quota", spaceName, spaceQuotaName)
				Eventually(session).Should(Exit(0))
			})

			It("unsets the quota from the space", func() {
				session := helpers.CF("unset-space-quota", spaceName, spaceQuotaName)
				Eventually(session).Should(Say(`Unassigning space quota %s from space %s as %s\.\.\.`, spaceQuotaName, spaceName, userName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
