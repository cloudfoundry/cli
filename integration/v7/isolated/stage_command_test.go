package isolated

import (
	"regexp"

	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("stage command", func() {
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
				Expect(session).To(HaveCommandInCategoryWithDescription("stage", "APPS", "Create a new droplet for an app"))
			})

			It("Displays command usage to output", func() {
				session := helpers.CF("stage", "--help")

				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("   stage - Create a new droplet for an app"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("   cf stage APP_NAME --package-guid PACKAGE_GUID"))
				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say("   --package-guid      The guid of the package to stage"))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("app, create-package, droplets, packages, push, set-droplet, stage"))

				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the app name is not provided", func() {
		It("tells the user that the app name is required, prints help text, and exits 1", func() {
			session := helpers.CF("stage", "--package-guid", "some-package-guid")

			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `APP_NAME` was not provided"))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	When("the package GUID flag is missing", func() {
		It("displays incorrect usage", func() {
			session := helpers.CF("stage", "some-app")

			Eventually(session.Err).Should(Say("Incorrect Usage: the required flag `--package-guid' was not specified"))
			Eventually(session).Should(Say("NAME:"))

			Eventually(session).Should(Exit(1))
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "stage", appName, "--package-guid", "some-package-guid")
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
			var packageGUID string

			BeforeEach(func() {
				Eventually(helpers.CF("create-app", appName)).Should(Exit(0))

				helpers.WithHelloWorldApp(func(appDir string) {
					pkgSession := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "create-package", appName)
					Eventually(pkgSession).Should(Exit(0))
					regex := regexp.MustCompile(`Package with guid '(.+)' has been created.`)
					matches := regex.FindStringSubmatch(string(pkgSession.Out.Contents()))
					Expect(matches).To(HaveLen(2))

					packageGUID = matches[1]
				})
			})

			It("stages the package", func() {
				session := helpers.CF("stage", appName, "--package-guid", packageGUID)
				userName, _ := helpers.GetCredentials()

				Eventually(session).Should(Say(`Staging package for %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
				Eventually(session).Should(Say(`Downloading staticfile_buildpack\.\.\.`), "Error streaming logs")
				Eventually(session).Should(Say(`Uploading droplet\.\.\.`), "Error streaming logs")
				Eventually(session).Should(Say("Package staged"))
				Eventually(session).Should(Say(`droplet guid:\s+%s`, helpers.GUIDRegex))
				Eventually(session).Should(Say(`state:\s+staged`))
				Eventually(session).Should(Say(`created:\s+%s`, helpers.UserFriendlyDateRegex))

				Eventually(session).Should(Exit(0))
			})

			When("the package belongs to a different app", func() {
				var otherAppName string

				BeforeEach(func() {
					otherAppName = helpers.PrefixedRandomName("app")
					Eventually(helpers.CF("create-app", otherAppName)).Should(Exit(0))
				})

				It("errors saying the package does *not* exist", func() {
					session := helpers.CF("stage", otherAppName, "--package-guid", packageGUID)
					userName, _ := helpers.GetCredentials()

					Eventually(session).Should(Say(`Staging package for %s in org %s / space %s as %s\.\.\.`, otherAppName, orgName, spaceName, userName))
					Eventually(session.Err).Should(Say(`Package with guid '%s' not found in app '%s'.`, packageGUID, otherAppName))

					Eventually(session).Should(Exit(1))
				})
			})
		})

		When("the app does not exist", func() {
			It("displays app not found and exits 1", func() {
				session := helpers.CF("stage", appName, "--package-guid", "some-package-guid")
				userName, _ := helpers.GetCredentials()

				Eventually(session).Should(Say(`Staging package for %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
				Eventually(session.Err).Should(Say("App '%s' not found", appName))
				Eventually(session).Should(Say("FAILED"))

				Eventually(session).Should(Exit(1))
			})
		})

		When("the package does not exist", func() {
			BeforeEach(func() {
				Eventually(helpers.CF("create-app", appName)).Should(Exit(0))
			})

			It("displays package not found and exits 1", func() {
				session := helpers.CF("stage", appName, "--package-guid", "some-package-guid")
				userName, _ := helpers.GetCredentials()

				Eventually(session).Should(Say(`Staging package for %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
				Eventually(session.Err).Should(Say(`Package with guid '%s' not found in app '%s'.`, "some-package-guid", appName))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})
	})
})
