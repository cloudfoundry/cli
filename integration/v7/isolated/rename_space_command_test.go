package isolated

import (
	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("rename-space command", func() {
	var (
		orgName      string
		spaceName    string
		spaceNameNew string
	)

	BeforeEach(func() {
		orgName = helpers.NewOrgName()
		spaceName = helpers.NewSpaceName()
		spaceNameNew = helpers.NewSpaceName()
	})

	Describe("help", func() {
		When("--help flag is set", func() {
			It("appears in cf help -a", func() {
				session := helpers.CF("help", "-a")
				Eventually(session).Should(Exit(0))
				Expect(session).To(HaveCommandInCategoryWithDescription("rename-space", "SPACES", "Rename a space"))
			})

			It("Displays command usage to output", func() {
				session := helpers.CF("rename-space", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("rename-space - Rename a space"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say(`cf rename-space SPACE NEW_SPACE_NAME`))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("space, space-quotas, space-users, spaces, target"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the space name is not provided", func() {
		It("tells the user that the space name is required, prints help text, and exits 1", func() {
			session := helpers.CF("rename-space")

			Eventually(session.Err).Should(Say("Incorrect Usage: the required arguments `SPACE` and `NEW_SPACE_NAME` were not provided"))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	When("the new space name is not provided", func() {
		It("tells the user that the space name is required, prints help text, and exits 1", func() {
			session := helpers.CF("rename-space", "space")

			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `NEW_SPACE_NAME` was not provided"))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, false, ReadOnlyOrg, "rename-space", spaceName, spaceNameNew)
		})
	})

	When("the environment is set up correctly", func() {
		BeforeEach(func() {
			helpers.SetupCF(orgName, spaceName)
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		When("space does not exist", func() {
			It("tells the user that the space does not exist, prints help text, and exits 1", func() {
				session := helpers.CF("rename-space", "not-a-space", spaceNameNew)

				Eventually(session.Err).Should(Say("Space 'not-a-space' not found."))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the space does exist", func() {
			It("renames the space in the targeted org", func() {
				session := helpers.CF("rename-space", spaceName, spaceNameNew)
				userName, _ := helpers.GetCredentials()
				Eventually(session).Should(Say("Renaming space %s to %s as %s...", spaceName, spaceNameNew, userName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))

				session = helpers.CF("space", spaceNameNew)
				Eventually(session).Should(Say(`name:\s+%s`, spaceNameNew))
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
