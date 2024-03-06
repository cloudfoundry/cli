package experimental

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("v3-restart command", func() {
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
			It("displays command usage to output", func() {
				session := helpers.CF("v3-restart", "--help")

				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say(`v3-restart - Stop all instances of the app, then start them again\. This causes downtime\.`))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf v3-restart APP_NAME"))
				Eventually(session).Should(Say("ENVIRONMENT:"))
				Eventually(session).Should(Say(`CF_STARTUP_TIMEOUT=5\s+Max wait time for app instance startup, in minutes`))

				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the app name is not provided", func() {
		It("tells the user that the app name is required, prints help text, and exits 1", func() {
			session := helpers.CF("v3-restart")

			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `APP_NAME` was not provided"))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	It("displays the experimental warning", func() {
		session := helpers.CF("v3-restart", appName)
		Eventually(session.Err).Should(Say("This command is in EXPERIMENTAL stage and may change without notice"))
		Eventually(session).Should(Exit())
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "v3-restart", appName)
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

			When("the app is running", func() {
				It("stops then restarts the app", func() {
					userName, _ := helpers.GetCredentials()

					session := helpers.CF("v3-restart", appName)
					Eventually(session).Should(Say(`Stopping app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Say(`Starting app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
					Eventually(session).Should(Say("OK"))

					Eventually(session).Should(Exit(0))
				})
			})

			When("the app is stopped", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("v3-stop", appName)).Should(Exit(0))
				})

				It("starts the app", func() {
					userName, _ := helpers.GetCredentials()

					session := helpers.CF("v3-restart", appName)
					Eventually(session).Should(Say(`Starting app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
					Eventually(session).Should(Say("OK"))

					Eventually(session).Should(Exit(0))
				})
			})

		})

		When("the app does not exist", func() {
			It("displays app not found and exits 1", func() {
				invalidAppName := helpers.PrefixedRandomName("invalid-app")
				session := helpers.CF("v3-restart", invalidAppName)

				Eventually(session.Err).Should(Say("App '%s' not found", invalidAppName))
				Eventually(session).Should(Say("FAILED"))

				Eventually(session).Should(Exit(1))
			})
		})
	})
})
