package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("set-health-check command", func() {
	Describe("help", func() {
		Context("when --help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("set-health-check", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("set-health-check - Change type of health check performed on an app"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf set-health-check APP_NAME \\(process \\| port \\| http \\[--endpoint PATH\\]\\)"))
				Eventually(session).Should(Say("TIP: 'none' has been deprecated but is accepted for 'process'."))
				Eventually(session).Should(Say("EXAMPLES:"))
				Eventually(session).Should(Say("   cf set-health-check worker-app process"))
				Eventually(session).Should(Say("   cf set-health-check my-web-app http --endpoint /foo"))
				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say("   --endpoint      Path on the app \\(Default: /\\)"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("when the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "set-health-check", "app-name", "port")
		})
	})

	Context("when the input is invalid", func() {
		DescribeTable("fails with incorrect usage method",
			func(args ...string) {
				cmd := append([]string{"set-health-check"}, args...)
				session := helpers.CF(cmd...)
				Eventually(session.Err).Should(Say("Incorrect Usage:"))
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Exit(1))
			},
			Entry("when app-name and health-check-type are not passed in"),
			Entry("when health-check-type is not passed in", "some-app"),
			Entry("when health-check-type is invalid", "some-app", "wut"),
		)
	})

	Context("when the environment is set up correctly", func() {
		var (
			orgName   string
			spaceName string
		)

		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()

			setupCF(orgName, spaceName)
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		Context("when the app does not exist", func() {
			It("tells the user that the app is not found and exits 1", func() {
				appName := helpers.PrefixedRandomName("app")
				session := helpers.CF("set-health-check", appName, "port")

				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("App %s not found", appName))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the app exists", func() {
			var appName string

			BeforeEach(func() {
				appName = helpers.PrefixedRandomName("app")
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CF("push", appName, "-p", appDir, "-b", "staticfile_buildpack", "--no-start")).Should(Exit(0))
				})
			})

			DescribeTable("Updates health-check-type and exits 0",
				func(settingType string) {
					session := helpers.CF("set-health-check", appName, settingType)

					username, _ := helpers.GetCredentials()
					Eventually(session).Should(Say("Updating health check type for app %s in org %s / space %s as %s", appName, orgName, spaceName, username))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))

					getSession := helpers.CF("get-health-check", appName)
					Eventually(getSession).Should(Say("health check type:\\s+%s", settingType))
					Eventually(getSession).Should(Exit(0))
				},
				Entry("when setting the health-check-type to 'none'", "none"),
				Entry("when setting the health-check-type to 'process'", "process"),
				Entry("when setting the health-check-type to 'port'", "port"),
				Entry("when setting the health-check-type to 'http'", "http"),
			)

			Context("when no http health check endpoint is given", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("set-health-check", appName, "http")).Should(Exit(0))
				})

				It("sets the http health check endpoint to /", func() {
					session := helpers.CF("get-health-check", appName)
					Eventually(session.Out).Should(Say("endpoint \\(for http type\\):\\s+/"))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when a valid http health check endpoint is given", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("set-health-check", appName, "http", "--endpoint", "/foo")).Should(Exit(0))
				})

				It("sets the http health check endpoint to the given endpoint", func() {
					session := helpers.CF("get-health-check", appName)
					Eventually(session.Out).Should(Say("endpoint \\(for http type\\):\\s+/foo"))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when an invalid http health check endpoint is given", func() {
				It("outputs an error and exits 1", func() {
					session := helpers.CF("set-health-check", appName, "http", "--endpoint", "invalid")
					Eventually(session.Err).Should(Say("The app is invalid: health_check_http_endpoint HTTP health check endpoint is not a valid URI path: invalid"))
					Eventually(session).Should(Exit(1))
				})
			})

			Context("when an endpoint is given with a non-http health check type", func() {
				It("outputs an error and exits 1", func() {
					session := helpers.CF("set-health-check", appName, "port", "--endpoint", "/foo")
					Eventually(session.Err).Should(Say("Health check type must be 'http' to set a health check HTTP endpoint\\."))
					Eventually(session).Should(Exit(1))
				})
			})

			Context("when the app is started", func() {
				BeforeEach(func() {
					appName = helpers.PrefixedRandomName("app")
					helpers.WithHelloWorldApp(func(appDir string) {
						Eventually(helpers.CF("push", appName, "-p", appDir, "-b", "staticfile_buildpack")).Should(Exit(0))
					})
				})

				It("displays tip to restart the app", func() {
					session := helpers.CF("set-health-check", appName, "port")
					Eventually(session).Should(Say("TIP: An app restart is required for the change to take affect\\."))
					Eventually(session).Should(Exit(0))
				})
			})
		})
	})
})
