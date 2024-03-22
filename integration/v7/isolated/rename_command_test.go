package isolated

import (
	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("rename command", func() {
	var (
		appName    string
		appNameNew string
		orgName    string
		spaceName  string
	)

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "rename", "appName", "newAppName")
		})
	})

	Describe("help", func() {
		When("--help flag is set", func() {
			It("appears in cf help -a", func() {
				session := helpers.CF("help", "-a")
				Eventually(session).Should(Exit(0))
				Expect(session).To(HaveCommandInCategoryWithDescription("rename", "APPS", "Rename an app"))
			})

			It("Displays command usage to output", func() {
				session := helpers.CF("rename", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("rename - Rename an app"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say(`cf rename APP_NAME NEW_APP_NAME`))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("apps, delete"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the environment is set up correctly", func() {

		BeforeEach(func() {
			appName = helpers.NewAppName()
			appNameNew = helpers.NewAppName()
			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()
			helpers.SetupCF(orgName, spaceName)
		})

		AfterEach(func() {
			helpers.QuickDeleteSpace(spaceName)
			helpers.QuickDeleteOrg(orgName)
		})

		When("the app name is not provided", func() {
			It("tells the user that the app name is required, prints help text, and exits 1", func() {
				session := helpers.CF("rename")

				Eventually(session.Err).Should(Say("Incorrect Usage: the required arguments `APP_NAME` and `NEW_APP_NAME` were not provided"))
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the new app name is not provided", func() {
			It("tells the user that the app name is required, prints help text, and exits 1", func() {
				session := helpers.CF("rename", "app")

				Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `NEW_APP_NAME` was not provided"))
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Exit(1))
			})
		})

		When("app does not exist", func() {
			It("tells the user that the app does not exist, prints help text, and exits 1", func() {
				session := helpers.CF("rename", "not-an-app", appNameNew)

				Eventually(session.Err).Should(Say("App 'not-an-app' not found."))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the app does exist", func() {
			BeforeEach(func() {
				helpers.CreateApp(appName)
			})

			AfterEach(func() {
				helpers.QuickDeleteApp(appNameNew)
			})

			It("renames the app", func() {
				session := helpers.CF("rename", appName, appNameNew)
				userName, _ := helpers.GetCredentials()
				Eventually(session).Should(Say("Renaming app %s to %s in org %s / space %s as %s...", appName, appNameNew, orgName, spaceName, userName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))

				session = helpers.CF("app", appNameNew)
				Eventually(session).Should(Say(`name:\s+%s`, appNameNew))
				Eventually(session).Should(Exit(0))
			})

			When("the new name is already taken", func() {
				BeforeEach(func() {
					helpers.CreateApp(appNameNew)
				})

				It("fails to rename the app", func() {
					session := helpers.CF("rename", appName, appNameNew)
					userName, _ := helpers.GetCredentials()
					Eventually(session).Should(Say("Renaming app %s to %s in org %s / space %s as %s...", appName, appNameNew, orgName, spaceName, userName))
					Eventually(session.Err).Should(Say("App with the name '%s' already exists.", appNameNew))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				})
			})
		})
	})
})
