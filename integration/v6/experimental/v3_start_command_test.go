package experimental

import (
	"regexp"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("v3-start-application command", func() {
	var (
		orgName   string
		spaceName string
		appName   string
	)

	BeforeEach(func() {
		helpers.SkipIfClientCredentialsTestMode()
		orgName = helpers.NewOrgName()
		spaceName = helpers.NewSpaceName()
		appName = helpers.PrefixedRandomName("app")
	})

	Describe("help", func() {
		When("--help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("v3-start", "--help")

				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("v3-start - Start an app"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf v3-start APP_NAME"))

				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the app name is not provided", func() {
		It("tells the user that the app name is required, prints help text, and exits 1", func() {
			session := helpers.CF("v3-start")

			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `APP_NAME` was not provided"))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	It("displays the experimental warning", func() {
		session := helpers.CF("v3-start", appName)
		Eventually(session.Err).Should(Say("This command is in EXPERIMENTAL stage and may change without notice"))
		Eventually(session).Should(Exit())
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "v3-start", appName)
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
				var packageGUID string
				Eventually(helpers.CF("v3-create-app", appName)).Should(Exit(0))

				helpers.WithHelloWorldApp(func(dir string) {
					pkgSession := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, "v3-create-package", appName)
					Eventually(pkgSession).Should(Exit(0))
					regex := regexp.MustCompile(`package guid: (.+)`)
					matches := regex.FindStringSubmatch(string(pkgSession.Out.Contents()))
					Expect(matches).To(HaveLen(2))

					packageGUID = matches[1]
				})

				stageSession := helpers.CF("v3-stage", appName, "--package-guid", packageGUID)
				Eventually(stageSession).Should(Exit(0))

				regex := regexp.MustCompile(`droplet guid:\s+(.+)`)
				matches := regex.FindStringSubmatch(string(stageSession.Out.Contents()))
				Expect(matches).To(HaveLen(2))

				dropletGUID := matches[1]
				setDropletSession := helpers.CF("v3-set-droplet", appName, "--droplet-guid", dropletGUID)
				Eventually(setDropletSession).Should(Exit(0))
			})

			It("starts the app", func() {
				userName, _ := helpers.GetCredentials()

				session := helpers.CF("v3-start", appName)
				Eventually(session).Should(Say(`Starting app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
				Eventually(session).Should(Say("OK"))

				Eventually(session).Should(Exit(0))
			})

			When("the app is already started", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("v3-start", appName)).Should(Exit(0))
				})

				It("displays app already started and exits 0", func() {
					session := helpers.CF("v3-start", appName)

					Eventually(session.Err).Should(Say("App %s is already started", appName))
					Eventually(session).Should(Say("OK"))

					Eventually(session).Should(Exit(0))
				})
			})
		})

		When("the app does not exist", func() {
			It("displays app not found and exits 1", func() {
				invalidAppName := "invalid-app-name"
				session := helpers.CF("v3-start", invalidAppName)

				Eventually(session.Err).Should(Say("App '%s' not found", invalidAppName))
				Eventually(session).Should(Say("FAILED"))

				Eventually(session).Should(Exit(1))
			})
		})
	})
})
