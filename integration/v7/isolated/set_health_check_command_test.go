// +build !partialPush

package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("set-health-check command", func() {
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
				session := helpers.CF("set-health-check", "--help")

				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("set-health-check - Change type of health check performed on an app's process"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say(`cf set-health-check APP_NAME \(process \| port \| http \[--endpoint PATH\]\) \[--process PROCESS\] \[--invocation-timeout INVOCATION_TIMEOUT\]`))

				Eventually(session).Should(Say("EXAMPLES:"))
				Eventually(session).Should(Say("cf set-health-check worker-app process --process worker"))
				Eventually(session).Should(Say("cf set-health-check my-web-app http --endpoint /foo"))
				Eventually(session).Should(Say("cf set-health-check my-web-app http --invocation-timeout 10"))

				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say(`--endpoint\s+Path on the app \(Default: /\)`))
				Eventually(session).Should(Say(`--invocation-timeout\s+Time \(in seconds\) that controls individual health check invocations`))
				Eventually(session).Should(Say(`--process\s+App process to update \(Default: web\)`))

				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the app name is not provided", func() {
		It("tells the user that the app name is required, prints help text, and exits 1", func() {
			session := helpers.CF("set-health-check")

			Eventually(session.Err).Should(Say("Incorrect Usage: the required arguments `APP_NAME` and `HEALTH_CHECK_TYPE` were not provided"))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	When("the health check type is not provided", func() {
		It("tells the user that health check type is required, prints help text, and exits 1", func() {
			session := helpers.CF("set-health-check", appName)

			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `HEALTH_CHECK_TYPE` was not provided"))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "set-health-check", appName, "port")
		})
	})

	When("the environment is set up correctly", func() {
		var userName string

		BeforeEach(func() {
			helpers.SetupCF(orgName, spaceName)
			userName, _ = helpers.GetCredentials()
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		When("the app exists", func() {
			BeforeEach(func() {
				helpers.WithProcfileApp(func(appDir string) {
					Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "push", appName)).Should(Exit(0))
				})
			})

			When("the process type is set", func() {
				It("displays the health check types for each process", func() {
					session := helpers.CF("set-health-check", appName, "http", "--endpoint", "/healthcheck", "--process", "console")
					Eventually(session).Should(Say(`Updating health check type for app %s process console in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
					Eventually(session).Should(Say(`TIP: An app restart is required for the change to take effect\.`))
					Eventually(session).Should(Exit(0))

					session = helpers.CF("get-health-check", appName)
					Eventually(session).Should(Say(`process\s+health check\s+endpoint \(for http\)\s+invocation timeout`))
					Eventually(session).Should(Say(`web\s+port\s+1`))
					Eventually(session).Should(Say(`console\s+http\s+/healthcheck\s+1`))

					Eventually(session).Should(Exit(0))
				})
			})

			When("the invocation timeout is set", func() {
				It("displays the health check types for each process", func() {
					session := helpers.CF("set-health-check", appName, "http", "--endpoint", "/healthcheck", "--invocation-timeout", "2")
					Eventually(session).Should(Say(`Updating health check type for app %s process web in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
					Eventually(session).Should(Say(`TIP: An app restart is required for the change to take effect\.`))
					Eventually(session).Should(Exit(0))

					session = helpers.CF("get-health-check", appName)
					Eventually(session).Should(Say(`process\s+health check\s+endpoint \(for http\)\s+invocation timeout`))
					Eventually(session).Should(Say(`web\s+http\s+/healthcheck\s+2`))

					Eventually(session).Should(Exit(0))
				})

			})

			When("process type and invocation timeout are not set", func() {
				It("displays the health check types for each process", func() {
					session := helpers.CF("set-health-check", appName, "http", "--endpoint", "/healthcheck")
					Eventually(session).Should(Say(`Updating health check type for app %s process web in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
					Eventually(session).Should(Say(`TIP: An app restart is required for the change to take effect\.`))
					Eventually(session).Should(Exit(0))

					session = helpers.CF("get-health-check", appName)
					Eventually(session).Should(Say(`process\s+health check\s+endpoint \(for http\)\s+invocation timeout`))
					Eventually(session).Should(Say(`web\s+http\s+/healthcheck\s+1`))
					Eventually(session).Should(Say(`console\s+process\s+1`))

					Eventually(session).Should(Exit(0))
				})
			})

			When("no http health check endpoint is given", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("set-health-check", appName, "http")).Should(Exit(0))
				})

				It("sets the http health check endpoint to /", func() {
					session := helpers.CF("get-health-check", appName)
					Eventually(session).Should(Say(`web\s+http\s+/`))
					Eventually(session).Should(Exit(0))
				})
			})

			When("the process type does not exist", func() {
				BeforeEach(func() {
					helpers.WithProcfileApp(func(appDir string) {
						Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "push", appName)).Should(Exit(0))
					})
				})

				It("returns a process not found error", func() {
					session := helpers.CF("set-health-check", appName, "http", "--endpoint", "/healthcheck", "--process", "nonexistent-type")
					Eventually(session).Should(Say(`Updating health check type for app %s process nonexistent-type in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
					Eventually(session.Err).Should(Say("Process nonexistent-type not found"))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("health check type is not 'http' and endpoint is set", func() {
				It("returns an error", func() {
					session := helpers.CF("set-health-check", appName, "port", "--endpoint", "oh-no", "--process", "console")
					Eventually(session.Out).Should(Say(`Updating health check type for app %s process console in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
					Eventually(session.Err).Should(Say("Health check type must be 'http' to set a health check HTTP endpoint."))
					Eventually(session.Out).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("an invalid http health check endpoint is given", func() {
				It("outputs an error and exits 1", func() {
					session := helpers.CF("set-health-check", appName, "http", "--endpoint", "invalid")
					Eventually(session.Err).Should(Say("Health check endpoint must be a valid URI path"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("the invocation timeout is less than one", func() {
				It("returns an error", func() {
					session := helpers.CF("set-health-check", appName, "port", "--invocation-timeout", "0", "--process", "console")
					Eventually(session.Err).Should(Say("Value must be greater than or equal to 1."))
					Eventually(session).Should(Exit(1))
				})
			})

			When("None is passed in", func() {
				It("returns an error", func() {
					session := helpers.CF("set-health-check", appName, "none")
					Eventually(session.Err).Should(Say(`Health check type must be "port", "process", or "http"`))
					Eventually(session).Should(Exit(1))
				})
			})
		})

		When("the app does not exist", func() {
			It("displays app not found and exits 1", func() {
				invalidAppName := "invalid-app-name"
				session := helpers.CF("set-health-check", invalidAppName, "port")

				Eventually(session.Out).Should(Say(`Updating health check type for app %s process web in org %s / space %s as %s\.\.\.`, invalidAppName, orgName, spaceName, userName))
				Eventually(session.Err).Should(Say("App %s not found", invalidAppName))
				Eventually(session.Out).Should(Say("FAILED"))

				Eventually(session).Should(Exit(1))
			})
		})
	})
})
