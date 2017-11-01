package experimental

import (
	"regexp"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	. "github.com/onsi/gomega/ghttp"
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
		Context("when --help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("v3-stage", "--help")

				Eventually(session.Out).Should(Say("NAME:"))
				Eventually(session.Out).Should(Say("   v3-stage - Create a new droplet for an app"))
				Eventually(session.Out).Should(Say("USAGE:"))
				Eventually(session.Out).Should(Say("   cf v3-stage APP_NAME --package-guid PACKAGE_GUID"))
				Eventually(session.Out).Should(Say("OPTIONS:"))
				Eventually(session.Out).Should(Say("   --package-guid      The guid of the package to stage"))

				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("when the app name is not provided", func() {
		It("tells the user that the app name is required, prints help text, and exits 1", func() {
			session := helpers.CF("v3-stage", "--package-guid", "some-package-guid")

			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `APP_NAME` was not provided"))
			Eventually(session.Out).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	Context("when the package GUID flag is missing", func() {
		It("displays incorrect usage", func() {
			session := helpers.CF("v3-stage", "some-app")

			Eventually(session.Err).Should(Say("Incorrect Usage: the required flag `--package-guid' was not specified"))
			Eventually(session.Out).Should(Say("NAME:"))

			Eventually(session).Should(Exit(1))
		})
	})

	It("displays the experimental warning", func() {
		session := helpers.CF("v3-stage", appName, "--package-guid", "some-package-guid")
		Eventually(session.Out).Should(Say("This command is in EXPERIMENTAL stage and may change without notice"))
		Eventually(session).Should(Exit())
	})

	Context("when the environment is not setup correctly", func() {
		Context("when no API endpoint is set", func() {
			BeforeEach(func() {
				helpers.UnsetAPI()
			})

			It("fails with no API endpoint set message", func() {
				session := helpers.CF("v3-stage", appName, "--package-guid", "some-package-guid")
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("No API endpoint set\\. Use 'cf login' or 'cf api' to target an endpoint\\."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the v3 api does not exist", func() {
			var server *Server

			BeforeEach(func() {
				server = helpers.StartAndTargetServerWithoutV3API()
			})

			AfterEach(func() {
				server.Close()
			})

			It("fails with error message that the minimum version is not met", func() {
				session := helpers.CF("v3-stage", appName, "--package-guid", "some-package-guid")
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("This command requires CF API version 3\\.27\\.0 or higher\\."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the v3 api version is lower than the minimum version", func() {
			var server *Server

			BeforeEach(func() {
				server = helpers.StartAndTargetServerWithV3Version("3.0.0")
			})

			AfterEach(func() {
				server.Close()
			})

			It("fails with error message that the minimum version is not met", func() {
				session := helpers.CF("v3-stage", appName, "--package-guid", "some-package-guid")
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("This command requires CF API version 3\\.27\\.0 or higher\\."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when not logged in", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
			})

			It("fails with not logged in message", func() {
				session := helpers.CF("v3-stage", appName, "--package-guid", "some-package-guid")
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("Not logged in\\. Use 'cf login' to log in\\."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when there is no org set", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
				helpers.LoginCF()
			})

			It("fails with no org targeted error message", func() {
				session := helpers.CF("v3-stage", appName, "--package-guid", "some-package-guid")
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("No org targeted, use 'cf target -o ORG' to target an org\\."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when there is no space set", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
				helpers.LoginCF()
				helpers.TargetOrg(ReadOnlyOrg)
			})

			It("fails with no space targeted error message", func() {
				session := helpers.CF("v3-stage", appName, "--package-guid", "some-package-guid")
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("No space targeted, use 'cf target -s SPACE' to target a space\\."))
				Eventually(session).Should(Exit(1))
			})
		})
	})

	Context("when the environment is set up correctly", func() {
		BeforeEach(func() {
			setupCF(orgName, spaceName)
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		Context("when the app exists", func() {
			var packageGUID string

			BeforeEach(func() {
				Eventually(helpers.CF("v3-create-app", appName)).Should(Exit(0))

				helpers.WithHelloWorldApp(func(appDir string) {
					pkgSession := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "v3-create-package", appName)
					Eventually(pkgSession).Should(Exit(0))
					regex, err := regexp.Compile(`package guid: (.+)`)
					Expect(err).ToNot(HaveOccurred())
					matches := regex.FindStringSubmatch(string(pkgSession.Out.Contents()))
					Expect(matches).To(HaveLen(2))

					packageGUID = matches[1]
				})
			})

			It("stages the package", func() {
				session := helpers.CF("v3-stage", appName, "--package-guid", packageGUID)
				userName, _ := helpers.GetCredentials()

				Eventually(session.Out).Should(Say("Staging package for %s in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, userName))
				Eventually(session.Out).Should(Say("Package staged"))
				Eventually(session.Out).Should(Say("droplet guid:\\s+%s", helpers.GUIDRegex))
				Eventually(session.Out).Should(Say("state:\\s+staged"))
				Eventually(session.Out).Should(Say("created:\\s+%s", helpers.UserFriendlyDateRegex))

				Eventually(session).Should(Exit(0))
			})
		})

		Context("when the app does not exist", func() {
			It("displays app not found and exits 1", func() {
				session := helpers.CF("v3-stage", appName, "--package-guid", "some-package-guid")
				userName, _ := helpers.GetCredentials()

				Eventually(session.Out).Should(Say("Staging package for %s in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, userName))
				Eventually(session.Err).Should(Say("App %s not found", appName))
				Eventually(session.Out).Should(Say("FAILED"))

				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the package does not exist", func() {
			BeforeEach(func() {
				Eventually(helpers.CF("v3-create-app", appName)).Should(Exit(0))
			})

			It("displays package not found and exits 1", func() {
				session := helpers.CF("v3-stage", appName, "--package-guid", "some-package-guid")
				userName, _ := helpers.GetCredentials()

				Eventually(session.Out).Should(Say("Staging package for %s in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, userName))
				Eventually(session.Err).Should(Say("Unable to use package\\. Ensure that the package exists and you have access to it\\."))
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})
	})
})
