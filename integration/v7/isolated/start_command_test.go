package isolated

import (
	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	"regexp"
)

const (
	PushCommandName = "push"
)

var _ = Describe("start command", func() {
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
				Expect(session).To(HaveCommandInCategoryWithDescription("start", "APPS", "Start an app"))
			})

			It("Displays command usage to output", func() {
				session := helpers.CF("start", "--help")

				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("start - Start an app"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf start APP_NAME"))
				Eventually(session).Should(Say("ALIAS:"))
				Eventually(session).Should(Say("st"))
				Eventually(session).Should(Say("ENVIRONMENT:"))
				Eventually(session).Should(Say(`CF_STAGING_TIMEOUT=15\s+Max wait time for staging, in minutes`))
				Eventually(session).Should(Say(`CF_STARTUP_TIMEOUT=5\s+Max wait time for app instance startup, in minutes`))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("apps, logs, restart, run-task, scale, ssh, stop"))

				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the app name is not provided", func() {
		It("tells the user that the app name is required, prints help text, and exits 1", func() {
			session := helpers.CF("start")

			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `APP_NAME` was not provided"))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "start", appName)
		})
	})

	When("the environment is set up correctly", func() {
		BeforeEach(func() {
			helpers.SetupCF(orgName, spaceName)
			Eventually(helpers.CF("create-app", appName)).Should(Exit(0))
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		When("the app exists", func() {
			When("the app does not need to be staged", func() {
				BeforeEach(func() {
					var packageGUID string

					// TODO: uncomment when map-route works in V7
					// mapRouteSession := helpers.CF("map-route", appName, helpers.DefaultSharedDomain(), "-n", appName)
					// Eventually(mapRouteSession).Should(Exit(0))

					helpers.WithHelloWorldApp(func(dir string) {
						pkgSession := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, "create-package", appName)
						Eventually(pkgSession).Should(Exit(0))
						regex := regexp.MustCompile(`Package with guid '(.+)' has been created.`)
						matches := regex.FindStringSubmatch(string(pkgSession.Out.Contents()))
						Expect(matches).To(HaveLen(2))

						packageGUID = matches[1]
					})

					stageSession := helpers.CF("stage", appName, "--package-guid", packageGUID)
					Eventually(stageSession).Should(Exit(0))

					regex := regexp.MustCompile(`droplet guid:\s+(.+)`)
					matches := regex.FindStringSubmatch(string(stageSession.Out.Contents()))
					Expect(matches).To(HaveLen(2))

					dropletGUID := matches[1]
					setDropletSession := helpers.CF("set-droplet", appName, "--droplet-guid", dropletGUID)
					Eventually(setDropletSession).Should(Exit(0))
				})

				It("starts the app", func() {
					userName, _ := helpers.GetCredentials()

					session := helpers.CF("start", appName)
					Eventually(session).Should(Say(`Starting app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
					Eventually(session).Should(Say(`Waiting for app to start\.\.\.`))
					Eventually(session).Should(Say(`name:\s+%s`, appName))
					Eventually(session).Should(Say(`requested state:\s+started`))
					// TODO: uncomment when map-route works in V7
					// Eventually(session).Should(Say(`routes:\s+%s.%s`, appName, helpers.DefaultSharedDomain()))
					Eventually(session).Should(Say(`type:\s+web`))
					Eventually(session).Should(Say(`instances:\s+1/1`))
					Eventually(session).Should(Say(`memory usage:\s+32M`))
					Eventually(session).Should(Say(`\s+state\s+since\s+cpu\s+memory\s+disk\s+details`))
					Eventually(session).Should(Say(`#0\s+(starting|running)\s+\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z`))

					Eventually(session).Should(Exit(0))
				})

				When("the app is already started", func() {
					BeforeEach(func() {
						Eventually(helpers.CF("start", appName)).Should(Exit(0))
					})

					It("displays app already started and exits 0", func() {
						session := helpers.CF("start", appName)

						Eventually(session).Should(Say(`App '%s' is already started\.`, appName))
						Eventually(session).Should(Say("OK"))

						Eventually(session).Should(Exit(0))
					})
				})
			})

			When("the app needs to be staged", func() {
				BeforeEach(func() {
					helpers.WithHelloWorldApp(func(dir string) {
						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName)
						Eventually(session).Should(Say(`\s+name:\s+%s`, appName))
						Eventually(session).Should(Say(`requested state:\s+started`))
						Eventually(session).Should(Exit(0))
					})

					helpers.WithBananaPantsApp(func(dir string) {
						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--no-start")
						Eventually(session).Should(Say(`Uploading files\.\.\.`))
						Eventually(session).Should(Say("100.00%"))
						Eventually(session).Should(Say(`Waiting for API to complete processing files\.\.\.`))
						Eventually(session).Should(Say(`\s+name:\s+%s`, appName))
						Eventually(session).Should(Say(`requested state:\s+stopped`))
						Eventually(session).Should(Exit(0))
					})
				})

				It("stages and starts the app", func() {
					session := helpers.CF("start", appName)
					userName, _ := helpers.GetCredentials()

					Eventually(session).Should(Say(`Staging package for %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
					Eventually(session).Should(Say("Package staged"))
					Eventually(session).Should(Say(`droplet guid:\s+%s`, helpers.GUIDRegex))
					Eventually(session).Should(Say(`state:\s+staged`))
					Eventually(session).Should(Say(`created:\s+%s`, helpers.UserFriendlyDateRegex))

					Eventually(session).Should(Say(`Starting app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
					Eventually(session).Should(Say(`Waiting for app to start\.\.\.`))
					Eventually(session).Should(Say(`name:\s+%s`, appName))
					Eventually(session).Should(Say(`requested state:\s+started`))
					Eventually(session).Should(Say(`type:\s+web`))
					Eventually(session).Should(Say(`instances:\s+1/1`))
					Eventually(session).Should(Say(`memory usage:\s+32M`))
					Eventually(session).Should(Say(`\s+state\s+since\s+cpu\s+memory\s+disk\s+details`))
					Eventually(session).Should(Say(`#0\s+(starting|running)\s+\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z`))

					Eventually(session).Should(Exit(0))
				})
			})

			When("the app cannot be started or staged", func() {
				It("gives an error", func() {
					session := helpers.CF("start", appName)

					Eventually(session.Err).Should(Say(`App can not start with out a package to stage or a droplet to run.`))
					Eventually(session).Should(Say("FAILED"))

					Eventually(session).Should(Exit(1))
				})
			})
		})

		When("the app does not exist", func() {
			It("displays app not found and exits 1", func() {
				invalidAppName := "invalid-app-name"
				session := helpers.CF("start", invalidAppName)

				Eventually(session.Err).Should(Say(`App '%s' not found\.`, invalidAppName))
				Eventually(session).Should(Say("FAILED"))

				Eventually(session).Should(Exit(1))
			})
		})
	})
})
