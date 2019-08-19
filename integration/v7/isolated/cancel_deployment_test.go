package isolated

import (
	"fmt"

	"code.cloudfoundry.org/cli/integration/helpers"

	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Cancel Deployment", func() {
	Context("Help", func() {
		It("appears in cf help -a", func() {
			session := helpers.CF("help", "-a")
			Eventually(session).Should(Exit(0))
			Expect(session).To(HaveCommandInCategoryWithDescription("cancel-deployment", "APPS", "Cancel the most recent deployment for an app. Resets the current droplet to the previous deployment's droplet."))
		})

		It("displays the help information", func() {
			session := helpers.CF("cancel-deployment", "--help")
			Eventually(session).Should(Say(`NAME:`))
			Eventually(session).Should(Say(`cancel-deployment - Cancel the most recent deployment for an app. Resets the current droplet to the previous deployment's droplet.\n`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`USAGE:`))
			Eventually(session).Should(Say(`cf cancel-deployment APP_NAME\n`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`EXAMPLES:`))
			Eventually(session).Should(Say(`cf cancel-deployment my-app\n`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`SEE ALSO:`))
			Eventually(session).Should(Say(`app, push`))

			Eventually(session).Should(Exit(0))
		})
	})

	Context("when the environment is not set up correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "cancel-deployment", "appName")
		})
	})

	Context("When the environment is set up correctly", func() {
		var (
			orgName   string
			spaceName string
			appName   string
			userName  string
		)

		BeforeEach(func() {
			appName = helpers.NewAppName()
			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()

			helpers.SetupCF(orgName, spaceName)
			userName, _ = helpers.GetCredentials()

			helpers.WithHelloWorldApp(func(appDir string) {
				Eventually(helpers.CF("push", appName, "-p", appDir, "-b", "staticfile_buildpack", "-i", "3", "--no-start")).Should(Exit(0))
			})
		})

		AfterEach(func() {
			Eventually(helpers.CF("delete", appName, "-f")).Should(Exit(0))
		})

		Context("when there are no deployments", func() {
			It("errors with a no deployments found error", func() {
				session := helpers.CF("cancel-deployment", appName)
				Eventually(session).Should(Say(fmt.Sprintf("Canceling deployment for app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName)))
				Eventually(session.Err).Should(Say(`No active deployment found for app\.`))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the cancel is successful", func() {
			BeforeEach(func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CF("push", appName, "-p", appDir)).Should(Exit(0))
				})
			})

			It("succeeds", func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CF("push", appName, "-p", appDir, "--strategy=rolling", "--no-wait")).Should(Exit(0))
				})

				session := helpers.CF("cancel-deployment", appName)
				Eventually(session).Should(Say(fmt.Sprintf("Canceling deployment for app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName)))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Say(fmt.Sprintf(`TIP: Run 'cf app %s' to view app status.`, appName)))
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
