package isolated

import (
	"fmt"

	"code.cloudfoundry.org/cli/v8/integration/helpers"

	. "code.cloudfoundry.org/cli/v8/cf/util/testhelpers/matchers"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Continue Deployment", func() {
	Context("Help", func() {
		It("appears in cf help -a", func() {
			session := helpers.CF("help", "-a")
			Eventually(session).Should(Exit(0))
			Expect(session).To(HaveCommandInCategoryWithDescription("continue-deployment", "APPS", "Continue the most recent deployment for an app."))
		})

		It("displays the help information", func() {
			session := helpers.CF("continue-deployment", "--help")
			Eventually(session).Should(Say(`NAME:`))
			Eventually(session).Should(Say(`continue-deployment - Continue the most recent deployment for an app.\n`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`USAGE:`))
			Eventually(session).Should(Say(`cf continue-deployment APP_NAME \[--no-wait\]\n`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`EXAMPLES:`))
			Eventually(session).Should(Say(`cf continue-deployment my-app\n`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`OPTIONS:`))
			Eventually(session).Should(Say(`--no-wait\s+Exit when the first instance of the web process is healthy`))
			Eventually(session).Should(Say(`\n`))

			Eventually(session).Should(Say(`SEE ALSO:`))
			Eventually(session).Should(Say(`app, push`))

			Eventually(session).Should(Exit(0))
		})
	})

	Context("when the environment is not set up correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "continue-deployment", "appName")
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
				Eventually(helpers.CF("push", appName, "-p", appDir, "-b", "staticfile_buildpack", "-i", "3", "--strategy", "canary")).Should(Exit(0))
			})
		})

		AfterEach(func() {
			Eventually(helpers.CF("delete", appName, "-f")).Should(Exit(0))
		})

		Context("when there are no deployments", func() {
			It("errors with a no deployments found error", func() {
				session := helpers.CF("continue-deployment", appName)
				Eventually(session).Should(Say(fmt.Sprintf("Continuing deployment for app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName)))
				Eventually(session.Err).Should(Say(`No active deployment found for app\.`))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the continue is successful", func() {
			When("There is a canary deployment", func() {
				When("instance steps are provided", func() {
					It("displays the number of steps", func() {
						helpers.WithHelloWorldApp(func(appDir string) {
							helpers.CF("push", appName, "-p", appDir, "--strategy=canary", "--instance-steps", "10,20,30,70", "-i", "5").Wait()
						})

						session := helpers.CF("app", appName)
						Eventually(session).Should(Say("canary-steps:    1/4"))
						session = helpers.CF("continue-deployment", appName)
						Eventually(session).Should(Say("canary-steps:    2/4"))
						Eventually(session).Should(Exit(0))
					})
				})

				When("instance steps are NOT provided", func() {
					It("succeeds", func() {
						helpers.WithHelloWorldApp(func(appDir string) {
							helpers.CF("push", appName, "-p", appDir, "--strategy=canary").Wait()
						})

						session := helpers.CF("continue-deployment", appName)
						Eventually(session).Should(Say(fmt.Sprintf("Continuing deployment for app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName)))
						Eventually(session).Should(Say(fmt.Sprintf(`TIP: Run 'cf app %s' to view app status.`, appName)))
						Eventually(session).Should(Exit(0))
					})
				})
			})
		})
	})
})
