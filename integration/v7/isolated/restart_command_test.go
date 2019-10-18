package isolated

import (
	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("restart command", func() {
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
				Expect(session).To(HaveCommandInCategoryWithDescription("restart", "APPS", "Stop all instances of the app, then start them again. This causes downtime."))
			})

			It("displays command usage to output", func() {
				session := helpers.CF("restart", "--help")

				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say(`restart - Stop all instances of the app, then start them again\. This causes downtime\.`))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf restart APP_NAME"))
				Eventually(session).Should(Say("ALIAS:"))
				Eventually(session).Should(Say("rs"))
				Eventually(session).Should(Say("ENVIRONMENT:"))
				Eventually(session).Should(Say(`CF_STAGING_TIMEOUT=15\s+Max wait time for staging, in minutes`))
				Eventually(session).Should(Say(`CF_STARTUP_TIMEOUT=5\s+Max wait time for app instance startup, in minutes`))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("restage, restart-app-instance"))

				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the app name is not provided", func() {
		It("tells the user that the app name is required, prints help text, and exits 1", func() {
			session := helpers.CF("restart")

			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `APP_NAME` was not provided"))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "restart", appName)
		})
	})

	When("the environment is set up correctly", func() {
		var (
			userName string
		)
		BeforeEach(func() {
			helpers.SetupCF(orgName, spaceName)
			userName, _ = helpers.GetCredentials()
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		When("the app exists", func() {
			When("strategy rolling is given", func() {
				BeforeEach(func() {
					helpers.WithHelloWorldApp(func(appDir string) {
						Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "push", appName)).Should(Exit(0))
					})
				})
				It("creates a deploy", func() {
					session := helpers.CF("restart", appName, "--strategy=rolling")
					Eventually(session).Should(Say(`Restarting app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
					Eventually(session).Should(Say(`Creating deployment for app %s\.\.\.`, appName))
					Eventually(session).Should(Say(`Waiting for app to deploy\.\.\.`))
					Eventually(session).Should(Say(`name:\s+%s`, appName))
					Eventually(session).Should(Say(`requested state:\s+started`))
					Eventually(session).Should(Say(`routes:\s+%s.%s`, appName, helpers.DefaultSharedDomain()))
					Eventually(session).Should(Say(`type:\s+web`))
					Eventually(session).Should(Say(`instances:\s+1/1`))
					Eventually(session).Should(Say(`memory usage:\s+\d+(M|G)`))
					Eventually(session).Should(Say(`\s+state\s+since\s+cpu\s+memory\s+disk\s+details`))
					Eventually(session).Should(Say(`#0\s+(starting|running)\s+\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z`))

				})
			})

			When("the app is running", func() {
				BeforeEach(func() {
					helpers.WithHelloWorldApp(func(appDir string) {
						Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "push", appName)).Should(Exit(0))
					})
				})

				It("stops then restarts the app", func() {

					session := helpers.CF("restart", appName)
					Eventually(session).Should(Say(`Restarting app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
					Eventually(session).Should(Say(`Stopping app\.\.\.`))
					Eventually(session).Should(Say(`Waiting for app to start\.\.\.`))
					Eventually(session).Should(Say(`name:\s+%s`, appName))
					Eventually(session).Should(Say(`requested state:\s+started`))
					Eventually(session).Should(Say(`routes:\s+%s.%s`, appName, helpers.DefaultSharedDomain()))
					Eventually(session).Should(Say(`type:\s+web`))
					Eventually(session).Should(Say(`instances:\s+1/1`))
					Eventually(session).Should(Say(`memory usage:\s+\d+(M|G)`))
					Eventually(session).Should(Say(`\s+state\s+since\s+cpu\s+memory\s+disk\s+details`))
					Eventually(session).Should(Say(`#0\s+(starting|running)\s+\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z`))

					Eventually(session).Should(Exit(0))
				})
			})

			When("the app is stopped", func() {
				When("the app is already staged", func() {
					BeforeEach(func() {
						helpers.WithHelloWorldApp(func(appDir string) {
							Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "push", appName)).Should(Exit(0))
						})
						Eventually(helpers.CF("stop", appName)).Should(Exit(0))
					})

					It("starts the app", func() {

						session := helpers.CF("restart", appName)
						Eventually(session).Should(Say(`Restarting app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
						Eventually(session).Should(Say(`Waiting for app to start\.\.\.`))
						Eventually(session).Should(Say(`name:\s+%s`, appName))
						Eventually(session).Should(Say(`requested state:\s+started`))
						Eventually(session).Should(Say(`routes:\s+%s.%s`, appName, helpers.DefaultSharedDomain()))
						Eventually(session).Should(Say(`type:\s+web`))
						Eventually(session).Should(Say(`instances:\s+1/1`))
						Eventually(session).Should(Say(`memory usage:\s+\d+(M|G)`))
						Eventually(session).Should(Say(`\s+state\s+since\s+cpu\s+memory\s+disk\s+details`))
						Eventually(session).Should(Say(`#0\s+(starting|running)\s+\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z`))

						Expect(session.Out.Contents()).NotTo(ContainSubstring("Stopping app..."))

						Eventually(session).Should(Exit(0))
					})
				})

				When("the app is *not* staged", func() {
					BeforeEach(func() {
						helpers.WithHelloWorldApp(func(appDir string) {
							Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "push", appName, "--no-start")).Should(Exit(0))
						})
					})

					It("complains about not having a droplet", func() {

						session := helpers.CF("restart", appName)
						Eventually(session).Should(Say(`Restarting app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
						Eventually(session.Err).Should(Say(`App can not start with out a package to stage or a droplet to run\.`))
						Eventually(session).Should(Say("FAILED"))
						Eventually(session).Should(Exit(1))
					})
				})
			})
		})

		When("the app does not exist", func() {
			It("displays app not found and exits 1", func() {
				invalidAppName := helpers.PrefixedRandomName("invalid-app")
				session := helpers.CF("restart", invalidAppName)

				Eventually(session.Err).Should(Say(`App '%s' not found\.`, invalidAppName))
				Eventually(session).Should(Say("FAILED"))

				Eventually(session).Should(Exit(1))
			})
		})
	})
})
