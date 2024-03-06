package experimental

import (
	"regexp"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("v3-stage command", func() {
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
				session := helpers.CF("v3-stage", "--help")

				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("   v3-stage - Create a new droplet for an app"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("   cf v3-stage APP_NAME --package-guid PACKAGE_GUID"))
				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say("   --package-guid      The guid of the package to stage"))

				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the app name is not provided", func() {
		It("tells the user that the app name is required, prints help text, and exits 1", func() {
			session := helpers.CF("v3-stage", "--package-guid", "some-package-guid")

			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `APP_NAME` was not provided"))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	When("the package GUID flag is missing", func() {
		It("displays incorrect usage", func() {
			session := helpers.CF("v3-stage", "some-app")

			Eventually(session.Err).Should(Say("Incorrect Usage: the required flag `--package-guid' was not specified"))
			Eventually(session).Should(Say("NAME:"))

			Eventually(session).Should(Exit(1))
		})
	})

	It("displays the experimental warning", func() {
		session := helpers.CF("v3-stage", appName, "--package-guid", "some-package-guid")
		Eventually(session.Err).Should(Say("This command is in EXPERIMENTAL stage and may change without notice"))
		Eventually(session).Should(Exit())
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "v3-stage", appName, "--package-guid", "some-package-guid")
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
				Eventually(helpers.CF("v3-create-app", appName)).Should(Exit(0))

				helpers.WithHelloWorldApp(func(appDir string) {
					pkgSession := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "v3-create-package", appName)
					Eventually(pkgSession).Should(Exit(0))
					regex := regexp.MustCompile(`package guid: (.+)`)
					matches := regex.FindStringSubmatch(string(pkgSession.Out.Contents()))
					Expect(matches).To(HaveLen(2))

					packageGUID = matches[1]
				})
			})

			It("stages the package", func() {
				session := helpers.CF("v3-stage", appName, "--package-guid", packageGUID)
				userName, _ := helpers.GetCredentials()

				Eventually(session).Should(Say(`Staging package for %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
				Eventually(session).Should(Say("Package staged"))
				Eventually(session).Should(Say(`droplet guid:\s+%s`, helpers.GUIDRegex))
				Eventually(session).Should(Say(`state:\s+staged`))
				Eventually(session).Should(Say(`created:\s+%s`, helpers.UserFriendlyDateRegex))

				Eventually(session).Should(Exit(0))
			})
		})

		When("the app does not exist", func() {
			It("displays app not found and exits 1", func() {
				session := helpers.CF("v3-stage", appName, "--package-guid", "some-package-guid")
				userName, _ := helpers.GetCredentials()

				Eventually(session).Should(Say(`Staging package for %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
				Eventually(session.Err).Should(Say("App '%s' not found", appName))
				Eventually(session).Should(Say("FAILED"))

				Eventually(session).Should(Exit(1))
			})
		})

		When("the package does not exist", func() {
			BeforeEach(func() {
				Eventually(helpers.CF("v3-create-app", appName)).Should(Exit(0))
			})

			It("displays package not found and exits 1", func() {
				session := helpers.CF("v3-stage", appName, "--package-guid", "some-package-guid")
				userName, _ := helpers.GetCredentials()

				Eventually(session).Should(Say(`Staging package for %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
				Eventually(session.Err).Should(Say(`Unable to use package\. Ensure that the package exists and you have access to it\.`))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})
	})
})
