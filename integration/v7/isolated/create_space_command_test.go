package isolated

import (
	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("create-space command", func() {
	var (
		orgName      string
		otherOrgName string
		spaceName    string
		spaceNameNew string
	)

	BeforeEach(func() {
		orgName = helpers.NewOrgName()
		otherOrgName = helpers.NewOrgName()
		spaceName = helpers.NewSpaceName()
		spaceNameNew = helpers.PrefixedRandomName("space")
	})

	Describe("help", func() {
		When("--help flag is set", func() {
			It("appears in cf help -a", func() {
				session := helpers.CF("help", "-a")
				Eventually(session).Should(Exit(0))
				Expect(session).To(HaveCommandInCategoryWithDescription("create-space", "SPACES", "Create a space"))
			})

			It("Displays command usage to output", func() {
				session := helpers.CF("create-space", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("create-space - Create a space"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say(`cf create-space SPACE \[-o ORG\]`))
				Eventually(session).Should(Say("ALIAS:"))
				Eventually(session).Should(Say(`csp`))
				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say(`-o\s+Organization`))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("set-space-isolation-segment, space-quotas, spaces, target"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the space name is not provided", func() {
		It("tells the user that the space name is required, prints help text, and exits 1", func() {
			session := helpers.CF("create-space")

			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `SPACE` was not provided"))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(false, false, ReadOnlyOrg, "create-space", spaceNameNew)
		})
	})

	When("the environment is set up correctly", func() {
		BeforeEach(func() {
			helpers.SetupCF(orgName, spaceName)
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		When("org does not exist", func() {
			It("tells the user that the org does not exist, prints help text, and exits 1", func() {
				session := helpers.CF("create-space", spaceNameNew, "-o", "not-an-org")

				Eventually(session.Err).Should(Say("Organization 'not-an-org' not found."))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the space does not exist", func() {

			When("the actor is a user", func() {

				It("creates the space in the targeted org", func() {
					session := helpers.CF("create-space", spaceNameNew)
					userName, _ := helpers.GetCredentials()
					Eventually(session).Should(Say("Creating space %s in org %s as %s...", spaceNameNew, orgName, userName))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Say(`TIP: Use 'cf target -o "%s" -s "%s"' to target new space`, orgName, spaceNameNew))
					Eventually(session).Should(Exit(0))

					session = helpers.CF("space", spaceNameNew)
					Eventually(session).Should(Say(`name:\s+%s`, spaceNameNew))
					Eventually(session).Should(Exit(0))
				})
			})

			When("the actor is a client", func() {
				var clientID string

				BeforeEach(func() {
					clientID = helpers.LoginCFWithClientCredentials()
					helpers.TargetOrg(orgName)
				})

				It("creates the space in the targeted org", func() {
					session := helpers.CF("create-space", spaceNameNew)
					Eventually(session).Should(Say("Creating space %s in org %s as %s...", spaceNameNew, orgName, clientID))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Say(`TIP: Use 'cf target -o "%s" -s "%s"' to target new space`, orgName, spaceNameNew))
					Eventually(session).Should(Exit(0))

					session = helpers.CF("space", spaceNameNew)
					Eventually(session).Should(Say(`name:\s+%s`, spaceNameNew))
					Eventually(session).Should(Exit(0))

					session = helpers.CF("space-users", orgName, spaceNameNew)
					Eventually(session).Should(Say("SPACE MANAGER"))
					Eventually(session).Should(Say(`\s+%s \(client\)`, clientID))
					Eventually(session).Should(Say("SPACE DEVELOPER"))
					Eventually(session).Should(Say(`\s+%s \(client\)`, clientID))
				})
			})

			When("org is specified", func() {
				BeforeEach(func() {
					helpers.CreateOrg(otherOrgName)
				})

				AfterEach(func() {
					helpers.QuickDeleteOrg(otherOrgName)
				})

				It("creates the space in the specified org and assigns roles to the user", func() {
					session := helpers.CF("create-space", spaceNameNew, "-o", otherOrgName)

					userName, _ := helpers.GetCredentials()
					Eventually(session).Should(Say("Creating space %s in org %s as %s...", spaceNameNew, otherOrgName, userName))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Say(`Assigning role SpaceManager to user %s in org %s / space %s as %s\.\.\.`, userName, otherOrgName, spaceNameNew, userName))
					Eventually(session).Should(Say(`OK\n`))
					Eventually(session).Should(Say(`Assigning role SpaceDeveloper to user %s in org %s / space %s as %s\.\.\.`, userName, otherOrgName, spaceNameNew, userName))
					Eventually(session).Should(Say(`OK\n\n`))
					Eventually(session).Should(Say(`TIP: Use 'cf target -o "%s" -s "%s"' to target new space`, otherOrgName, spaceNameNew))
					Eventually(session).Should(Exit(0))

					helpers.TargetOrg(otherOrgName)
					session = helpers.CF("space", spaceNameNew)
					Eventually(session).Should(Say(`name:\s+%s`, spaceNameNew))
					Eventually(session).Should(Exit(0))

					session = helpers.CF("space-users", otherOrgName, spaceNameNew)
					Eventually(session).Should(Say("SPACE MANAGER"))
					Eventually(session).Should(Say(`\s+%s`, userName))
					Eventually(session).Should(Say("SPACE DEVELOPER"))
					Eventually(session).Should(Say(`\s+%s`, userName))
				})
			})
		})

		When("the space already exists", func() {
			BeforeEach(func() {
				Eventually(helpers.CF("create-space", spaceNameNew)).Should(Exit(0))
			})

			It("fails to create the space", func() {
				session := helpers.CF("create-space", spaceNameNew)
				userName, _ := helpers.GetCredentials()
				Eventually(session).Should(Say("Creating space %s in org %s as %s...", spaceNameNew, orgName, userName))
				Eventually(session).Should(Say(`Space '%s' already exists\.`, spaceNameNew))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
