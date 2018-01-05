package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("get-health-check command", func() {
	Describe("help", func() {
		Context("when --help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("get-health-check", "--help")

				Eventually(session.Out).Should(Say("NAME:"))
				Eventually(session.Out).Should(Say("   get-health-check - Show the type of health check performed on an app"))
				Eventually(session.Out).Should(Say("USAGE:"))
				Eventually(session.Out).Should(Say("   cf get-health-check APP_NAME"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("when the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "get-health-check", "app-name")
		})
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

		Context("when the input is invalid", func() {
			Context("when there are not enough arguments", func() {
				It("outputs the usage and exits 1", func() {
					session := helpers.CF("get-health-check")

					Eventually(session.Err).Should(Say("Incorrect Usage:"))
					Eventually(session.Out).Should(Say("NAME:"))
					Eventually(session).Should(Exit(1))
				})
			})

			Context("when there too many arguments", func() {
				It("ignores the extra arguments", func() {
					appName := helpers.PrefixedRandomName("app")
					session := helpers.CF("get-health-check", appName, "extra")
					username, _ := helpers.GetCredentials()

					Eventually(session.Out).Should(Say("Getting health check type for app %s in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, username))
					Eventually(session.Err).Should(Say("App %s not found", appName))
					Eventually(session.Out).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				})
			})
		})

		Context("when the app does not exist", func() {
			It("tells the user that the app is not found and exits 1", func() {
				appName := helpers.PrefixedRandomName("app")
				session := helpers.CF("get-health-check", appName)
				username, _ := helpers.GetCredentials()

				Eventually(session.Out).Should(Say("Getting health check type for app %s in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, username))
				Eventually(session.Err).Should(Say("App %s not found", appName))
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the app exists", func() {
			var (
				appName  string
				username string
			)

			BeforeEach(func() {
				appName = helpers.PrefixedRandomName("app")
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CF("push", appName, "-p", appDir, "-b", "staticfile_buildpack", "--no-start")).Should(Exit(0))
				})
				username, _ = helpers.GetCredentials()
			})

			Context("when the health check type is http", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("set-health-check", appName, "http")).Should(Exit(0))
				})

				It("shows an endpoint", func() {
					session := helpers.CF("get-health-check", appName)

					Eventually(session.Out).Should(Say("Getting health check type for app %s in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, username))
					Eventually(session.Out).Should(Say("\n\n"))
					Eventually(session.Out).Should(Say("health check type:          http"))
					Eventually(session.Out).Should(Say("endpoint \\(for http type\\):   /"))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when the health check type is http with a custom endpoint", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("set-health-check", appName, "http", "--endpoint", "/some-endpoint")).Should(Exit(0))
				})

				It("show the custom endpoint", func() {
					session := helpers.CF("get-health-check", appName)

					Eventually(session.Out).Should(Say("Getting health check type for app %s in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, username))
					Eventually(session.Out).Should(Say("\n\n"))
					Eventually(session.Out).Should(Say("health check type:          http"))
					Eventually(session.Out).Should(Say("endpoint \\(for http type\\):   /some-endpoint"))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when the health check type is none", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("set-health-check", appName, "none")).Should(Exit(0))
				})

				It("does not show an endpoint", func() {
					session := helpers.CF("get-health-check", appName)

					Eventually(session.Out).Should(Say("Getting health check type for app %s in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, username))
					Eventually(session.Out).Should(Say("\n\n"))
					Eventually(session.Out).Should(Say("health check type:          none"))
					Eventually(session.Out).Should(Say("(?m)endpoint \\(for http type\\):   $"))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when the health check type is port", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("set-health-check", appName, "port")).Should(Exit(0))
				})

				It("does not show an endpoint", func() {
					session := helpers.CF("get-health-check", appName)

					Eventually(session.Out).Should(Say("Getting health check type for app %s in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, username))
					Eventually(session.Out).Should(Say("\n\n"))
					Eventually(session.Out).Should(Say("health check type:          port"))
					Eventually(session.Out).Should(Say("(?m)endpoint \\(for http type\\):   $"))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when the health check type is process", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("set-health-check", appName, "process")).Should(Exit(0))
				})

				It("does not show an endpoint", func() {
					session := helpers.CF("get-health-check", appName)

					Eventually(session.Out).Should(Say("Getting health check type for app %s in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, username))
					Eventually(session.Out).Should(Say("\n\n"))
					Eventually(session.Out).Should(Say("health check type:          process"))
					Eventually(session.Out).Should(Say("(?m)endpoint \\(for http type\\):   $"))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when the health check type changes from http to another type", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("set-health-check", appName, "http", "--endpoint", "/some-endpoint")).Should(Exit(0))
					Eventually(helpers.CF("set-health-check", appName, "process")).Should(Exit(0))
				})

				It("does not show an endpoint", func() {
					session := helpers.CF("get-health-check", appName)

					Eventually(session.Out).Should(Say("Getting health check type for app %s in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, username))
					Eventually(session.Out).Should(Say("\n\n"))
					Eventually(session.Out).Should(Say("health check type:          process"))
					Eventually(session.Out).Should(Say("(?m)endpoint \\(for http type\\):   $"))
					Eventually(session).Should(Exit(0))
				})
			})
		})
	})
})
