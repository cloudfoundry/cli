package isolated

import (
	"regexp"

	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("restart command", func() {

	const (
		instanceStatsTitles = `\s+state\s+since\s+cpu\s+memory\s+disk\s+logging\s+cpu entitlement\s+details`
		instanceStatsValues = `#0\s+(starting|running)\s+\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z`
	)

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
				Expect(session).To(HaveCommandInCategoryWithDescription("restart", "APPS", "Stop all instances of the app, then start them again."))
			})

			It("displays command usage to output", func() {
				session := helpers.CF("restart", "--help")

				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say(`restart - Stop all instances of the app, then start them again\.`))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf restart APP_NAME"))
				Eventually(session).Should(Say("This command will cause downtime unless you use '--strategy"))
				Eventually(session).Should(Say("If the app's most recent package is unstaged, restarting the app will stage and run that package."))
				Eventually(session).Should(Say("Otherwise, the app's current droplet will be run."))
				Eventually(session).Should(Say("ALIAS:"))
				Eventually(session).Should(Say("rs"))
				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say("--max-in-flight"))
				Eventually(session).Should(Say(`--strategy\s+Deployment strategy can be canary, rolling or null.`))
				Eventually(session).Should(Say(`--no-wait\s+Exit when the first instance of the web process is healthy`))
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
					session := helpers.CF("restart", appName, "--strategy=rolling", "--max-in-flight=3")
					Eventually(session).Should(Say(`Restarting app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
					Eventually(session).Should(Say(`Creating deployment for app %s\.\.\.`, appName))
					Eventually(session).Should(Say(`Waiting for app to deploy\.\.\.`))
					Eventually(session).Should(Say(`name:\s+%s`, appName))
					Eventually(session).Should(Say(`requested state:\s+started`))
					Eventually(session).Should(Say(`routes:\s+%s.%s`, appName, helpers.DefaultSharedDomain()))
					Eventually(session).Should(Say(`type:\s+web`))
					Eventually(session).Should(Say(`instances:\s+1/1`))
					Eventually(session).Should(Say(`memory usage:\s+\d+(M|G)`))
					Eventually(session).Should(Say(instanceStatsTitles))
					Eventually(session).Should(Say(instanceStatsValues))
				})
			})

			When("strategy canary is given without the max-in-flight flag", func() {
				BeforeEach(func() {
					helpers.WithHelloWorldApp(func(appDir string) {
						Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "push", appName)).Should(Exit(0))
					})
				})
				It("creates a deploy without noting max-in-flight", func() {
					session := helpers.CF("restart", appName, "--strategy=canary")
					Eventually(session).Should(Say(`Restarting app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
					Eventually(session).Should(Say(`Creating deployment for app %s\.\.\.`, appName))
					Eventually(session).Should(Say(`Waiting for app to deploy\.\.\.`))
					Eventually(session).Should(Say(`name:\s+%s`, appName))
					Eventually(session).Should(Say(`requested state:\s+started`))
					Eventually(session).Should(Say(`routes:\s+%s.%s`, appName, helpers.DefaultSharedDomain()))
					Eventually(session).Should(Say(`type:\s+web`))
					Eventually(session).Should(Say(`instances:\s+1/1`))
					Eventually(session).Should(Say(`memory usage:\s+\d+(M|G)`))
					Eventually(session).Should(Say(instanceStatsTitles))
					Eventually(session).Should(Say(instanceStatsValues))
					Eventually(session).Should(Say("Canary deployment currently PAUSED"))
					Eventually(session).ShouldNot(Say("max-in-flight"))
					Eventually(session).Should(Say("Please run `cf continue-deployment %s` to promote the canary deployment, or `cf cancel-deployment %s` to rollback to the previous version.", appName, appName))
					Eventually(session).Should(Exit(0))
				})
			})

			When("strategy canary is given with a non-default max-in-flight value", func() {
				BeforeEach(func() {
					helpers.WithHelloWorldApp(func(appDir string) {
						Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "push", appName)).Should(Exit(0))
					})
				})
				It("creates a deploy and notes the max-in-flight value", func() {
					session := helpers.CF("restart", appName, "--strategy=canary", "--max-in-flight", "3")
					Eventually(session).Should(Say(`Restarting app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
					Eventually(session).Should(Say(`Creating deployment for app %s\.\.\.`, appName))
					Eventually(session).Should(Say(`Waiting for app to deploy\.\.\.`))
					Eventually(session).Should(Say(`name:\s+%s`, appName))
					Eventually(session).Should(Say(`requested state:\s+started`))
					Eventually(session).Should(Say(`routes:\s+%s.%s`, appName, helpers.DefaultSharedDomain()))
					Eventually(session).Should(Say(`type:\s+web`))
					Eventually(session).Should(Say(`instances:\s+1/1`))
					Eventually(session).Should(Say(`memory usage:\s+\d+(M|G)`))
					Eventually(session).Should(Say(instanceStatsTitles))
					Eventually(session).Should(Say(instanceStatsValues))
					Eventually(session).Should(Say("Canary deployment currently PAUSED"))
					Eventually(session).Should(Say("max-in-flight: 3"))
					Eventually(session).Should(Say("Please run `cf continue-deployment %s` to promote the canary deployment, or `cf cancel-deployment %s` to rollback to the previous version.", appName, appName))
					Eventually(session).Should(Exit(0))
				})
			})

			When("the app is running with no new packages", func() {
				BeforeEach(func() {
					helpers.WithHelloWorldApp(func(appDir string) {
						Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "push", appName)).Should(Exit(0))
					})
				})

				It("stops then restarts the app, without staging a package", func() {
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
					Eventually(session).Should(Say(instanceStatsTitles))
					Eventually(session).Should(Say(instanceStatsValues))

					Expect(session.Out.Contents()).NotTo(ContainSubstring("Staging app and tracing logs..."))

					Eventually(session).Should(Exit(0))
					Expect(session.Err).ToNot(Say(`timeout connecting to log server, no log will be shown`))
				})
			})

			When("the app is running with a new packages", func() {
				BeforeEach(func() {
					helpers.WithHelloWorldApp(func(appDir string) {
						Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "push", appName)).Should(Exit(0))
					})
					helpers.WithHelloWorldApp(func(appDir string) {
						pkgSession := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "create-package", appName)
						Eventually(pkgSession).Should(Exit(0))
						regex := regexp.MustCompile(`Package with guid '(.+)' has been created\.`)
						matches := regex.FindStringSubmatch(string(pkgSession.Out.Contents()))
						Expect(matches).To(HaveLen(2))
					})
				})

				It("stages the new package, stops then restarts the app", func() {
					session := helpers.CF("restart", appName)
					Eventually(session).Should(Say(`Restarting app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))

					Eventually(session).Should(Say("Staging app and tracing logs..."))
					Eventually(session).Should(Say(`Stopping app\.\.\.`))
					Eventually(session).Should(Say(`Waiting for app to start\.\.\.`))
					Eventually(session).Should(Say(`name:\s+%s`, appName))
					Eventually(session).Should(Say(`requested state:\s+started`))
					Eventually(session).Should(Say(`routes:\s+%s.%s`, appName, helpers.DefaultSharedDomain()))
					Eventually(session).Should(Say(`type:\s+web`))
					Eventually(session).Should(Say(`instances:\s+1/1`))
					Eventually(session).Should(Say(`memory usage:\s+\d+(M|G)`))
					Eventually(session).Should(Say(instanceStatsTitles))
					Eventually(session).Should(Say(instanceStatsValues))

					Eventually(session).Should(Exit(0))
					Expect(session.Err).ToNot(Say(`timeout connecting to log server, no log will be shown`))
				})
			})

			When("the app is stopped", func() {
				When("the app does not have a new package, and has a current droplet", func() {
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
						Eventually(session).Should(Say(instanceStatsTitles))
						Eventually(session).Should(Say(instanceStatsValues))

						Expect(session.Out.Contents()).NotTo(ContainSubstring("Staging app and tracing logs..."))
						Expect(session.Out.Contents()).NotTo(ContainSubstring("Stopping app..."))

						Eventually(session).Should(Exit(0))
					})
				})

				When("the app has a new package", func() {
					BeforeEach(func() {
						helpers.WithHelloWorldApp(func(appDir string) {
							Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "push", appName, "--no-start")).Should(Exit(0))
						})
					})

					It("stages the new package and starts the app", func() {

						session := helpers.CF("restart", appName)
						Eventually(session).Should(Say(`Restarting app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
						Eventually(session).Should(Say("Staging app and tracing logs..."))
						Eventually(session).Should(Say(`Waiting for app to start\.\.\.`))
						Eventually(session).Should(Say(`name:\s+%s`, appName))
						Eventually(session).Should(Say(`requested state:\s+started`))
						Eventually(session).Should(Say(`routes:\s+%s.%s`, appName, helpers.DefaultSharedDomain()))
						Eventually(session).Should(Say(`type:\s+web`))
						Eventually(session).Should(Say(`instances:\s+1/1`))
						Eventually(session).Should(Say(`memory usage:\s+\d+(M|G)`))
						Eventually(session).Should(Say(instanceStatsTitles))
						Eventually(session).Should(Say(instanceStatsValues))

						Expect(session.Out.Contents()).NotTo(ContainSubstring("Stopping app..."))

						Eventually(session).Should(Exit(0))
					})
				})

				When("the app does *not* have a ready package or current droplet", func() {
					BeforeEach(func() {
						Eventually(helpers.CF("create-app", appName)).Should(Exit(0))
					})

					It("complains about not having a droplet", func() {

						session := helpers.CF("restart", appName)
						Eventually(session).Should(Say(`Restarting app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
						Eventually(session.Err).Should(Say(`App cannot start without a package to stage or a droplet to run\.`))
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
