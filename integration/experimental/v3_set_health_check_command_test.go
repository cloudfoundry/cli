package experimental

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("v3-set-health-check command", func() {
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
				session := helpers.CF("v3-set-health-check", "--help")

				Eventually(session.Out).Should(Say("NAME:"))
				Eventually(session.Out).Should(Say("v3-set-health-check - Change type of health check performed on an app's process"))
				Eventually(session.Out).Should(Say("USAGE:"))
				Eventually(session.Out).Should(Say(`cf v3-set-health-check APP_NAME \(process \| port \| http \[--endpoint PATH\]\) \[--process PROCESS\]`))

				Eventually(session.Out).Should(Say("EXAMPLES:"))
				Eventually(session.Out).Should(Say("cf v3-set-health-check worker-app process --process worker"))
				Eventually(session.Out).Should(Say("cf v3-set-health-check my-web-app http --endpoint /foo"))

				Eventually(session.Out).Should(Say("OPTIONS:"))
				Eventually(session.Out).Should(Say(`--endpoint\s+Path on the app \(Default: /\)`))
				Eventually(session.Out).Should(Say(`--process\s+App process to update \(Default: web\)`))

				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("when the app name is not provided", func() {
		It("tells the user that the app name is required, prints help text, and exits 1", func() {
			session := helpers.CF("v3-set-health-check")

			Eventually(session.Err).Should(Say("Incorrect Usage: the required arguments `APP_NAME` and `HEALTH_CHECK_TYPE` were not provided"))
			Eventually(session.Out).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	Context("when the health check type is not provided", func() {
		It("tells the user that health check type is required, prints help text, and exits 1", func() {
			session := helpers.CF("v3-set-health-check", appName)

			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `HEALTH_CHECK_TYPE` was not provided"))
			Eventually(session.Out).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	It("displays the experimental warning", func() {
		session := helpers.CF("v3-set-health-check", appName, "port")
		Eventually(session.Out).Should(Say("This command is in EXPERIMENTAL stage and may change without notice"))
		Eventually(session).Should(Exit())
	})

	Context("when the environment is not setup correctly", func() {
		Context("when no API endpoint is set", func() {
			BeforeEach(func() {
				helpers.UnsetAPI()
			})

			It("fails with no API endpoint set message", func() {
				session := helpers.CF("v3-set-health-check", appName, "port")
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
				session := helpers.CF("v3-set-health-check", appName, "port")
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
				session := helpers.CF("v3-set-health-check", appName, "port")
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
				session := helpers.CF("v3-set-health-check", appName, "port")
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
				session := helpers.CF("v3-set-health-check", appName, "port")
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
				session := helpers.CF("v3-set-health-check", appName, "port")
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("No space targeted, use 'cf target -s SPACE' to target a space\\."))
				Eventually(session).Should(Exit(1))
			})
		})
	})

	Context("when the environment is set up correctly", func() {
		var userName string

		BeforeEach(func() {
			setupCF(orgName, spaceName)
			userName, _ = helpers.GetCredentials()
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		Context("when the app exists", func() {
			BeforeEach(func() {
				helpers.WithProcfileApp(func(appDir string) {
					Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "v3-push", appName)).Should(Exit(0))
				})
			})

			It("displays the health check types for each process", func() {
				session := helpers.CF("v3-set-health-check", appName, "http", "--endpoint", "/healthcheck", "--process", "console")
				Eventually(session.Out).Should(Say("Updating health check type for app %s process console in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, userName))
				Eventually(session.Out).Should(Say("TIP: An app restart is required for the change to take effect\\."))
				Eventually(session).Should(Exit(0))

				session = helpers.CF("v3-get-health-check", appName)
				Eventually(session.Out).Should(Say("Getting process health check types for app %s in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, userName))
				Eventually(session.Out).Should(Say(`process\s+health check\s+endpoint \(for http\)\n`))
				Eventually(session.Out).Should(Say(`web\s+port\s+\n`))
				Eventually(session.Out).Should(Say(`console\s+http\s+/healthcheck`))

				Eventually(session).Should(Exit(0))
			})

			Context("when the process type does not exist", func() {
				BeforeEach(func() {
					helpers.WithProcfileApp(func(appDir string) {
						Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "v3-push", appName)).Should(Exit(0))
					})
				})

				It("returns a process not found error", func() {
					session := helpers.CF("v3-set-health-check", appName, "http", "--endpoint", "/healthcheck", "--process", "nonexistant-type")
					Eventually(session.Out).Should(Say("Updating health check type for app %s process nonexistant-type in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, userName))
					Eventually(session.Err).Should(Say("Process nonexistant-type not found"))
					Eventually(session.Out).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				})
			})
		})

		Context("when the app does not exist", func() {
			It("displays app not found and exits 1", func() {
				invalidAppName := "invalid-app-name"
				session := helpers.CF("v3-set-health-check", invalidAppName, "port")

				Eventually(session.Out).Should(Say("Updating health check type for app %s process web in org %s / space %s as %s\\.\\.\\.", invalidAppName, orgName, spaceName, userName))
				Eventually(session.Err).Should(Say("App %s not found", invalidAppName))
				Eventually(session.Out).Should(Say("FAILED"))

				Eventually(session).Should(Exit(1))
			})
		})
	})
})
