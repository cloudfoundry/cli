package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo"
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
				Eventually(session).Should(Say("set-health-check - Set health_check_type flag to either 'port' or 'none'"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf set-health-check APP_NAME \\('port' \\| 'none'\\)"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("when the environment is not setup correctly", func() {
		Context("when no API endpoint is set", func() {
			BeforeEach(func() {
				helpers.UnsetAPI()
			})

			It("fails with no API endpoint set message", func() {
				session := helpers.CF("set-health-check", "some-app", "port")
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("No API endpoint set. Use 'cf login' or 'cf api' to target an endpoint."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when not logged in", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
			})

			It("fails with not logged in message", func() {
				session := helpers.CF("set-health-check", "some-app", "port")
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("Not logged in. Use 'cf login' to log in."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when there is no org and space set", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
				helpers.LoginCF()
			})

			It("fails with no targeted org error message", func() {
				session := helpers.CF("set-health-check", "some-app", "port")
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("No org targeted, use 'cf target -o ORG' to target an org."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when there is no space set", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
				helpers.LoginCF()
				helpers.TargetOrg(ReadOnlyOrg)
			})

			It("fails with no targeted space error message", func() {
				session := helpers.CF("set-health-check", "some-app", "port")
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("No space targeted, use 'cf target -s SPACE' to target a space."))
				Eventually(session).Should(Exit(1))
			})
		})
	})

	Context("when app-name and health-check-type are not passed in", func() {
		It("fails with incorrect ussage error message and displays help", func() {
			session := helpers.CF("set-health-check")
			Eventually(session.Err).Should(Say("Incorrect Usage: the required arguments `APP_NAME` and `HEALTH_CHECK_TYPE` were not provided"))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Say("set-health-check - Set health_check_type flag to either 'port' or 'none'"))
			Eventually(session).Should(Say("USAGE:"))
			Eventually(session).Should(Say("cf set-health-check APP_NAME \\('port' \\| 'none'\\)"))
			Eventually(session).Should(Exit(1))
		})
	})

	Context("when health-check-type is not passed in", func() {
		It("fails with incorrect usage error message and displays help", func() {
			session := helpers.CF("set-health-check", "some-app")
			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `HEALTH_CHECK_TYPE` was not provided"))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Say("set-health-check - Set health_check_type flag to either 'port' or 'none'"))
			Eventually(session).Should(Say("USAGE:"))
			Eventually(session).Should(Say("cf set-health-check APP_NAME \\('port' \\| 'none'\\)"))
			Eventually(session).Should(Exit(1))
		})
	})

	Context("when health-check-type is invalid", func() {
		It("fails with incorrect usage error message and displays help", func() {
			session := helpers.CF("set-health-check", "some-app", "wut")
			Eventually(session.Err).Should(Say(`Incorrect Usage: HEALTH_CHECK_TYPE must be "port" or "none"`))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Say("set-health-check - Set health_check_type flag to either 'port' or 'none'"))
			Eventually(session).Should(Say("USAGE:"))
			Eventually(session).Should(Say("cf set-health-check APP_NAME \\('port' \\| 'none'\\)"))
			Eventually(session).Should(Exit(1))
		})
	})

	Context("when the environment is set up correctly", func() {
		var (
			orgName   string
			spaceName string
		)

		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			spaceName = helpers.PrefixedRandomName("SPACE")

			setupCF(orgName, spaceName)
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
			var (
				appName string
			)

			BeforeEach(func() {
				appName = helpers.PrefixedRandomName("app")
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CF("push", appName, "-p", appDir, "-b", "staticfile_buildpack")).Should(Exit(0))
				})
			})

			Context("when setting the health-check-type to 'none'", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("set-health-check", appName, "port")).Should(Exit(0))
				})

				It("updates the new health-check-type and exits 0", func() {
					session := helpers.CF("set-health-check", appName, "none")

					username, _ := helpers.GetCredentials()
					Eventually(session).Should(Say("Updating app %s in org %s / space %s as %s", appName, orgName, spaceName, username))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when setting the health-check-type to 'port'", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("set-health-check", appName, "none")).Should(Exit(0))
				})

				It("updates the new health-check-type and exits 0", func() {
					session := helpers.CF("set-health-check", appName, "port")

					username, _ := helpers.GetCredentials()
					Eventually(session).Should(Say("Updating app %s in org %s / space %s as %s", appName, orgName, spaceName, username))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))
				})
			})
		})
	})
})
