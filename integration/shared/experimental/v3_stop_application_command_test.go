// +build !partialPush

package experimental

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("v3-stop-application command", func() {
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
				session := helpers.CF("v3-stop", "--help")

				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("v3-stop - Stop an app"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf v3-stop APP_NAME"))

				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the app name is not provided", func() {
		It("tells the user that the app name is required, prints help text, and exits 1", func() {
			session := helpers.CF("v3-stop")

			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `APP_NAME` was not provided"))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	It("displays the experimental warning", func() {
		session := helpers.CF("v3-stop", appName)
		Eventually(session.Err).Should(Say("This command is in EXPERIMENTAL stage and may change without notice"))
		Eventually(session).Should(Exit())
	})

	When("the environment is not setup correctly", func() {
		When("the v3 api version is lower than the minimum version", func() {
			var server *Server

			BeforeEach(func() {
				server = helpers.StartAndTargetServerWithAPIVersions(helpers.DefaultV2Version, ccversion.MinV3ClientVersion)
			})

			AfterEach(func() {
				server.Close()
			})

			It("fails with error message that the minimum version is not met", func() {
				session := helpers.CF("v3-stop", appName)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say(`This command requires CF API version 3\.27\.0 or higher\.`))
				Eventually(session).Should(Exit(1))
			})
		})

		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "v3-stop", appName)
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
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "v3-push", appName)).Should(Exit(0))
				})
			})

			It("stops the app", func() {
				userName, _ := helpers.GetCredentials()

				session := helpers.CF("v3-stop", appName)
				Eventually(session).Should(Say(`Stopping app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
				Eventually(session).Should(Say("OK"))

				Eventually(session).Should(Exit(0))
			})

			When("the app is already stopped", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("v3-stop", appName)).Should(Exit(0))
				})

				It("displays that the app is already stopped", func() {
					session := helpers.CF("v3-stop", appName)

					Eventually(session.Err).Should(Say("App %s is already stopped", appName))
					Eventually(session).Should(Say("OK"))

					Eventually(session).Should(Exit(0))
				})
			})

			When("the app does not exist", func() {
				It("displays app not found and exits 1", func() {
					invalidAppName := "invalid-app-name"
					session := helpers.CF("v3-stop", invalidAppName)

					Eventually(session.Err).Should(Say("App %s not found", invalidAppName))
					Eventually(session).Should(Say("FAILED"))

					Eventually(session).Should(Exit(1))
				})
			})
		})
	})
})
