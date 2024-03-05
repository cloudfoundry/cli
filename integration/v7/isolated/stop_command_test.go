package isolated

import (
	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("stop command", func() {
	var (
		orgName   string
		spaceName string
		appName   string
	)

	BeforeEach(func() {
		orgName = helpers.NewOrgName()
		spaceName = helpers.NewSpaceName()
		appName = helpers.PrefixedRandomName("app")
	})

	Describe("help", func() {
		When("--help flag is set", func() {
			It("appears in cf help -a", func() {
				session := helpers.CF("help", "-a")
				Eventually(session).Should(Exit(0))
				Expect(session).To(HaveCommandInCategoryWithDescription("stop", "APPS", "Stop an app"))
			})

			It("Displays command usage to output", func() {
				session := helpers.CF("stop", "--help")

				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("stop - Stop an app"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf stop APP_NAME"))
				Eventually(session).Should(Say("ALIAS:"))
				Eventually(session).Should(Say("sp"))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("restart, scale, start"))

				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the app name is not provided", func() {
		It("tells the user that the app name is required, prints help text, and exits 1", func() {
			session := helpers.CF("stop")

			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `APP_NAME` was not provided"))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "stop", appName)
		})
	})

	When("the environment is set up correctly", func() {
		BeforeEach(func() {
			helpers.SetupCF(orgName, spaceName)
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		When("the app exists", func() {
			BeforeEach(func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "v3-push", appName)).Should(Exit(0))
				})
			})

			It("stops the app", func() {
				userName, _ := helpers.GetCredentials()

				session := helpers.CF("stop", appName)
				Eventually(session).Should(Say(`Stopping app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
				Eventually(session).Should(Say("OK"))

				Eventually(session).Should(Exit(0))
			})

			When("the app is already stopped", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("stop", appName)).Should(Exit(0))
				})

				It("displays that the app is already stopped", func() {
					session := helpers.CF("stop", appName)

					Eventually(session).Should(Say(`App %s is already stopped\.`, appName))
					Eventually(session).Should(Say("OK"))

					Eventually(session).Should(Exit(0))
				})
			})

			When("the app does not exist", func() {
				It("displays app not found and exits 1", func() {
					invalidAppName := "invalid-app-name"
					session := helpers.CF("stop", invalidAppName)

					Eventually(session.Err).Should(Say("App '%s' not found", invalidAppName))
					Eventually(session).Should(Say("FAILED"))

					Eventually(session).Should(Exit(1))
				})
			})
		})
	})
})
