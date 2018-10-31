package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("get-health-check command", func() {
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
				session := helpers.CF("get-health-check", "--help")

				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("get-health-check - Show the type of health check performed on an app"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf get-health-check APP_NAME"))

				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the app name is not provided", func() {
		It("tells the user that the app name is required, prints help text, and exits 1", func() {
			session := helpers.CF("get-health-check")

			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `APP_NAME` was not provided"))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "get-health-check", appName)
		})
	})

	When("the environment is set up correctly", func() {
		var username string

		BeforeEach(func() {
			helpers.SetupCF(orgName, spaceName)
			username, _ = helpers.GetCredentials()
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
					session := helpers.CF("get-health-check", appName, "extra")

					Eventually(session).Should(Say(`Getting health check type for app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, username))
					Eventually(session.Err).Should(Say("App %s not found", appName))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				})
			})
		})

		When("the app exists", func() {
			BeforeEach(func() {
				helpers.WithProcfileApp(func(appDir string) {
					Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "push", appName)).Should(Exit(0))
				})
			})

			It("displays the health check types for each process", func() {
				session := helpers.CF("get-health-check", appName)

				Eventually(session).Should(Say(`Getting health check type for app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, username))
				Eventually(session).Should(Say(`process\s+health check\s+endpoint \(for http\)\s+invocation timeout\n`))
				Eventually(session).Should(Say(`web\s+port\s+1\n`))
				Eventually(session).Should(Say(`console\s+process\s+1\n`))

				Eventually(session).Should(Exit(0))
			})

			When("the health check type is http", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("set-health-check", appName, "http")).Should(Exit(0))
				})

				It("shows the health check type is http with an endpoint of `/`", func() {
					session := helpers.CF("get-health-check", appName)

					Eventually(session).Should(Say(`Getting health check type for app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, username))
					Eventually(session).Should(Say(`process\s+health check\s+endpoint \(for http\)\s+invocation timeout\n`))
					Eventually(session).Should(Say(`web\s+http\s+/\s+1\n`))
					Eventually(session).Should(Say(`console\s+process\s+1\n`))

					Eventually(session).Should(Exit(0))
				})
			})

			When("the health check type is http with a custom endpoint", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("set-health-check", appName, "http", "--endpoint", "/some-endpoint")).Should(Exit(0))
				})

				It("shows the health check type is http with the custom endpoint", func() {
					session := helpers.CF("get-health-check", appName)

					Eventually(session).Should(Say(`Getting health check type for app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, username))
					Eventually(session).Should(Say(`process\s+health check\s+endpoint \(for http\)\s+invocation timeout\n`))
					Eventually(session).Should(Say(`web\s+http\s+/some-endpoint\s+1\n`))
					Eventually(session).Should(Say(`console\s+process\s+1\n`))

					Eventually(session).Should(Exit(0))
				})
			})

			When("the health check type is port", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("set-health-check", appName, "port")).Should(Exit(0))
				})

				It("shows that the health check type is port", func() {
					session := helpers.CF("get-health-check", appName)

					Eventually(session).Should(Say(`Getting health check type for app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, username))
					Eventually(session).Should(Say(`web\s+port\s+\d+`))

					Eventually(session).Should(Exit(0))
				})
			})

			When("the health check type is process", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("set-health-check", appName, "process")).Should(Exit(0))
				})

				It("shows that the health check type is process", func() {
					session := helpers.CF("get-health-check", appName)

					Eventually(session).Should(Say(`Getting health check type for app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, username))
					Eventually(session).Should(Say(`web\s+process\s+\d+`))

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

					Consistently(session).ShouldNot(Say("/some-endpoint"))
					Eventually(session).Should(Say(`Getting health check type for app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, username))
					Eventually(session).Should(Say("\n\n"))
					Eventually(session).Should(Say(`web\s+process\s+\d+`))
					Eventually(session).Should(Say(`console\s+process\s+\d+`))
					Eventually(session).Should(Say(`rake\s+process\s+\d+`))

					Eventually(session).Should(Exit(0))
				})
			})
		})

		When("the app does not exist", func() {
			It("displays app not found and exits 1", func() {
				session := helpers.CF("get-health-check", appName)

				Eventually(session).Should(Say(`Getting health check type for app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, username))
				Eventually(session.Err).Should(Say("App %s not found", appName))
				Eventually(session).Should(Say("FAILED"))

				Eventually(session).Should(Exit(1))
			})
		})
	})
})
