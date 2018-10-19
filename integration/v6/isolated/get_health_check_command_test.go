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
		When("--help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("get-health-check", "--help")

				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("   get-health-check - Show the type of health check performed on an app"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("   cf get-health-check APP_NAME"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "get-health-check", "app-name")
		})
	})

	When("the environment is set up correctly", func() {
		var (
			orgName   string
			spaceName string
		)

		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()

			helpers.SetupCF(orgName, spaceName)
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		When("the input is invalid", func() {
			When("there are not enough arguments", func() {
				It("outputs the usage and exits 1", func() {
					session := helpers.CF("get-health-check")

					Eventually(session.Err).Should(Say("Incorrect Usage:"))
					Eventually(session).Should(Say("NAME:"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("there too many arguments", func() {
				It("ignores the extra arguments", func() {
					appName := helpers.PrefixedRandomName("app")
					session := helpers.CF("get-health-check", appName, "extra")
					username, _ := helpers.GetCredentials()

					Eventually(session).Should(Say("Getting health check type for app %s in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, username))
					Eventually(session.Err).Should(Say("App %s not found", appName))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				})
			})
		})

		When("the app does not exist", func() {
			It("tells the user that the app is not found and exits 1", func() {
				appName := helpers.PrefixedRandomName("app")
				session := helpers.CF("get-health-check", appName)
				username, _ := helpers.GetCredentials()

				Eventually(session).Should(Say("Getting health check type for app %s in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, username))
				Eventually(session.Err).Should(Say("App %s not found", appName))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the app exists", func() {
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

			When("the health check type is http", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("set-health-check", appName, "http")).Should(Exit(0))
				})

				It("shows an endpoint", func() {
					session := helpers.CF("get-health-check", appName)

					Eventually(session).Should(Say("Getting health check type for app %s in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, username))
					Eventually(session).Should(Say("\n\n"))
					Eventually(session).Should(Say("health check type:          http"))
					Eventually(session).Should(Say("endpoint \\(for http type\\):   /"))
					Eventually(session).Should(Exit(0))
				})
			})

			When("the health check type is http with a custom endpoint", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("set-health-check", appName, "http", "--endpoint", "/some-endpoint")).Should(Exit(0))
				})

				It("show the custom endpoint", func() {
					session := helpers.CF("get-health-check", appName)

					Eventually(session).Should(Say("Getting health check type for app %s in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, username))
					Eventually(session).Should(Say("\n\n"))
					Eventually(session).Should(Say("health check type:          http"))
					Eventually(session).Should(Say("endpoint \\(for http type\\):   /some-endpoint"))
					Eventually(session).Should(Exit(0))
				})
			})

			When("the health check type is none", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("set-health-check", appName, "none")).Should(Exit(0))
				})

				It("does not show an endpoint", func() {
					session := helpers.CF("get-health-check", appName)

					Eventually(session).Should(Say("Getting health check type for app %s in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, username))
					Eventually(session).Should(Say("\n\n"))
					Eventually(session).Should(Say("health check type:          none"))
					Eventually(session).Should(Say("(?m)endpoint \\(for http type\\):   $"))
					Eventually(session).Should(Exit(0))
				})
			})

			When("the health check type is port", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("set-health-check", appName, "port")).Should(Exit(0))
				})

				It("does not show an endpoint", func() {
					session := helpers.CF("get-health-check", appName)

					Eventually(session).Should(Say("Getting health check type for app %s in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, username))
					Eventually(session).Should(Say("\n\n"))
					Eventually(session).Should(Say("health check type:          port"))
					Eventually(session).Should(Say("(?m)endpoint \\(for http type\\):   $"))
					Eventually(session).Should(Exit(0))
				})
			})

			When("the health check type is process", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("set-health-check", appName, "process")).Should(Exit(0))
				})

				It("does not show an endpoint", func() {
					session := helpers.CF("get-health-check", appName)

					Eventually(session).Should(Say("Getting health check type for app %s in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, username))
					Eventually(session).Should(Say("\n\n"))
					Eventually(session).Should(Say("health check type:          process"))
					Eventually(session).Should(Say("(?m)endpoint \\(for http type\\):   $"))
					Eventually(session).Should(Exit(0))
				})
			})

			When("the health check type changes from http to another type", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("set-health-check", appName, "http", "--endpoint", "/some-endpoint")).Should(Exit(0))
					Eventually(helpers.CF("set-health-check", appName, "process")).Should(Exit(0))
				})

				It("does not show an endpoint", func() {
					session := helpers.CF("get-health-check", appName)

					Eventually(session).Should(Say("Getting health check type for app %s in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, username))
					Eventually(session).Should(Say("\n\n"))
					Eventually(session).Should(Say("health check type:          process"))
					Eventually(session).Should(Say("(?m)endpoint \\(for http type\\):   $"))
					Eventually(session).Should(Exit(0))
				})
			})
		})
	})
})
