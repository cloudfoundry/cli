package experimental

import (
	"regexp"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("v3-set-droplet command", func() {
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
			It("Displays command usage to output", func() {
				session := helpers.CF("v3-set-droplet", "--help")

				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("v3-set-droplet - Set the droplet used to run an app"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf v3-set-droplet APP_NAME -d DROPLET_GUID"))
				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say(`--droplet-guid, -d\s+The guid of the droplet to use`))

				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the app name is not provided", func() {
		It("tells the user that the app name is required, prints help text, and exits 1", func() {
			session := helpers.CF("v3-set-droplet", "-d", "some-droplet-guid")

			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `APP_NAME` was not provided"))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	It("displays the experimental warning", func() {
		session := helpers.CF("v3-set-droplet", appName, "--droplet-guid", "some-droplet-guid")
		Eventually(session.Err).Should(Say("This command is in EXPERIMENTAL stage and may change without notice"))
		Eventually(session).Should(Exit())
	})

	When("the package GUID flag is missing", func() {
		It("displays incorrect usage", func() {
			session := helpers.CF("v3-set-droplet", "some-app")

			Eventually(session.Err).Should(Say("Incorrect Usage: the required flag `-d, --droplet-guid' was not specified"))
			Eventually(session).Should(Say("NAME:"))

			Eventually(session).Should(Exit(1))
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "v3-set-droplet", appName, "--droplet-guid", "some-droplet-guid")
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
			var dropletGUID string

			BeforeEach(func() {
				var packageGUID string
				Eventually(helpers.CF("v3-create-app", appName)).Should(Exit(0))

				helpers.WithHelloWorldApp(func(appDir string) {
					pkgSession := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "v3-create-package", appName)
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

				dropletGUID = matches[1]
			})

			It("sets the droplet for the app", func() {
				userName, _ := helpers.GetCredentials()

				session := helpers.CF("v3-set-droplet", appName, "-d", dropletGUID)
				Eventually(session).Should(Say(`Setting app %s to droplet %s in org %s / space %s as %s\.\.\.`, appName, dropletGUID, orgName, spaceName, userName))
				Eventually(session).Should(Say("OK"))

				Eventually(session).Should(Exit(0))
			})

			When("the app does not exist", func() {
				It("displays app not found and exits 1", func() {
					invalidAppName := "invalid-app-name"
					session := helpers.CF("v3-set-droplet", invalidAppName, "-d", dropletGUID)
					userName, _ := helpers.GetCredentials()

					Eventually(session).Should(Say(`Setting app %s to droplet %s in org %s / space %s as %s\.\.\.`, invalidAppName, dropletGUID, orgName, spaceName, userName))
					Eventually(session.Err).Should(Say("App '%s' not found", invalidAppName))
					Eventually(session).Should(Say("FAILED"))

					Eventually(session).Should(Exit(1))
				})
			})

			When("the droplet does not exist", func() {
				It("displays droplet not found and exits 1", func() {
					invalidDropletGUID := "some-droplet-guid"
					session := helpers.CF("v3-set-droplet", appName, "-d", invalidDropletGUID)
					userName, _ := helpers.GetCredentials()

					Eventually(session).Should(Say(`Setting app %s to droplet %s in org %s / space %s as %s\.\.\.`, appName, invalidDropletGUID, orgName, spaceName, userName))
					Eventually(session.Err).Should(Say(`Unable to assign droplet: Unable to assign current droplet\. Ensure the droplet exists and belongs to this app\.`))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				})
			})
		})
	})
})
