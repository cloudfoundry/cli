package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = XDescribe("restart command", func() {

	BeforeEach(func() {
		helpers.SkipIfClientCredentialsTestMode()
	})

	Describe("help", func() {
		When("--help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("restart-app-instance", "--help")

				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("restart-app-instance - Terminate, then restart an app instance"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf restart-app-instance APP_NAME INDEX"))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("restart"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "restart-app-instance", "app-name", "2")
		})
	})

	When("the environment is set up correctly", func() {
		var (
			orgName   string
			spaceName string
			userName  string
			appName   string
		)

		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()
			appName = helpers.NewAppName()

			helpers.SetupCF(orgName, spaceName)
			userName, _ = helpers.GetCredentials()
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		When("the app does not exist", func() {
			It("tells the user that the start is not found and exits 1", func() {
				session := helpers.CF("restart", appName, "0")

				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("App '%s' not found", appName))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the app does exist", func() {
			BeforeEach(func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CF("push", appName, "-p", appDir, "-b", "staticfile_buildpack")).Should(Exit(0))
				})

				It("restarts app instance", func() {
					session := helpers.CF("restart", appName)
					Eventually(session).Should(Say(`Restarting instance %d of the app %s in org %s / space %s as %s\.\.\.`, 10, appName, orgName, spaceName, userName))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))
				})
			})
		})
	})
})
