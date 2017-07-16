package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("v3-restart command", func() {
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
			It("displays command usage to output", func() {
				session := helpers.CF("v3-restart", "--help")

				Eventually(session.Out).Should(Say("NAME:"))
				Eventually(session.Out).Should(Say("v3-restart - Stop all instances of the app, then start them again\\. This may cause downtime\\."))
				Eventually(session.Out).Should(Say("USAGE:"))
				Eventually(session.Out).Should(Say("cf v3-restart APP_NAME"))
				Eventually(session.Out).Should(Say("ENVIRONMENT:"))
				Eventually(session.Out).Should(Say("CF_STARTUP_TIMEOUT=5\\s+Max wait time for app instance startup, in minutes"))

				Eventually(session).Should(Exit(0))
			})
		})

		Context("when 'help -a' is called", func() {
			It("does not display v3-restart", func() {
				session := helpers.CF("help", "-a")
				Consistently(session.Out).ShouldNot(Say("v3-restart"))
			})
		})
	})

	Context("when the app name is not provided", func() {
		It("tells the user that the app name is required, prints help text, and exits 1", func() {
			session := helpers.CF("v3-restart")

			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `APP_NAME` was not provided"))
			Eventually(session.Out).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	Context("when the environment is not setup correctly", func() {
		Context("when no API endpoint is set", func() {
			BeforeEach(func() {
				helpers.UnsetAPI()
			})

			It("fails with no API endpoint set message", func() {
				session := helpers.CF("v3-restart", appName)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("No API endpoint set\\. Use 'cf login' or 'cf api' to target an endpoint\\."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when not logged in", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
			})

			It("fails with not logged in message", func() {
				session := helpers.CF("v3-restart", appName)
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
				session := helpers.CF("v3-restart", appName)
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
				session := helpers.CF("v3-restart", appName)
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

		Context("when the app exists", func() {
			BeforeEach(func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CF("v3-push", appName)).Should(Exit(0))
				})
			})

			Context("when the app is running", func() {
				It("stops then restarts the app", func() {
					userName, _ := helpers.GetCredentials()

					session := helpers.CF("v3-restart", appName)
					Eventually(session.Out).Should(Say("Stopping app %s in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, userName))
					Eventually(session.Out).Should(Say("OK"))
					Eventually(session.Out).Should(Say("Starting app %s in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, userName))
					Eventually(session.Out).Should(Say("OK"))

					Eventually(session).Should(Exit(0))
				})
			})

			Context("when the app is stopped", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("v3-stop", appName)).Should(Exit(0))
				})

				It("starts the app", func() {
					userName, _ := helpers.GetCredentials()

					session := helpers.CF("v3-restart", appName)
					Eventually(session.Out).Should(Say("Starting app %s in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, userName))
					Eventually(session.Out).Should(Say("OK"))

					Eventually(session).Should(Exit(0))
				})
			})

		})

		Context("when the app does not exist", func() {
			It("displays app not found and exits 1", func() {
				invalidAppName := helpers.PrefixedRandomName("invalid-app")
				session := helpers.CF("v3-restart", invalidAppName)

				Eventually(session.Err).Should(Say("App %s not found", invalidAppName))
				Eventually(session.Out).Should(Say("FAILED"))

				Eventually(session).Should(Exit(1))
			})
		})
	})
})
