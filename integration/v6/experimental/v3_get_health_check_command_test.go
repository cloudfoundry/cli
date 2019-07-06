package experimental

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("v3-get-health-check command", func() {
	var (
		orgName   string
		spaceName string
		appName   string
	)

	BeforeEach(func() {
		helpers.SkipIfClientCredentialsTestMode()
		orgName = helpers.NewOrgName()
		spaceName = helpers.NewSpaceName()
		appName = helpers.PrefixedRandomName("app")
	})

	Describe("help", func() {
		When("--help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("v3-get-health-check", "--help")

				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("v3-get-health-check - Show the type of health check performed on an app"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf v3-get-health-check APP_NAME"))

				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the app name is not provided", func() {
		It("tells the user that the app name is required, prints help text, and exits 1", func() {
			session := helpers.CF("v3-get-health-check")

			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `APP_NAME` was not provided"))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	It("displays the experimental warning", func() {
		session := helpers.CF("v3-get-health-check", appName)
		Eventually(session.Err).Should(Say("This command is in EXPERIMENTAL stage and may change without notice"))
		Eventually(session).Should(Exit())
	})

	When("the environment is not setup correctly", func() {
		When("no API endpoint is set", func() {
			BeforeEach(func() {
				helpers.UnsetAPI()
			})

			It("fails with no API endpoint set message", func() {
				session := helpers.CF("v3-get-health-check", appName)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say(`No API endpoint set\. Use 'cf login' or 'cf api' to target an endpoint\.`))
				Eventually(session).Should(Exit(1))
			})
		})

		When("not logged in", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
			})

			It("fails with not logged in message", func() {
				session := helpers.CF("v3-get-health-check", appName)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say(`Not logged in\. Use 'cf login' to log in\.`))
				Eventually(session).Should(Exit(1))
			})
		})

		When("there is no org set", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
				helpers.LoginCF()
			})

			It("fails with no org targeted error message", func() {
				session := helpers.CF("v3-get-health-check", appName)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say(`No org targeted, use 'cf target -o ORG' to target an org\.`))
				Eventually(session).Should(Exit(1))
			})
		})

		When("there is no space set", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
				helpers.LoginCF()
				helpers.TargetOrg(ReadOnlyOrg)
			})

			It("fails with no space targeted error message", func() {
				session := helpers.CF("v3-get-health-check", appName)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say(`No space targeted, use 'cf target -s SPACE' to target a space\.`))
				Eventually(session).Should(Exit(1))
			})
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
					Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "v3-push", appName)).Should(Exit(0))
				})
			})

			It("displays the health check types for each process", func() {
				userName, _ = helpers.GetCredentials()

				session := helpers.CF("v3-get-health-check", appName)
				Eventually(session).Should(Say(`Getting process health check types for app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
				Eventually(session).Should(Say(`process\s+health check\s+endpoint \(for http\)\s+invocation timeout\n`))
				Eventually(session).Should(Say(`web\s+port\s+1\n`))
				Eventually(session).Should(Say(`console\s+process\s+1\n`))

				Eventually(session).Should(Exit(0))
			})
		})

		When("the app does not exist", func() {
			It("displays app not found and exits 1", func() {
				invalidAppName := "invalid-app-name"
				session := helpers.CF("v3-get-health-check", invalidAppName)

				Eventually(session).Should(Say(`Getting process health check types for app %s in org %s / space %s as %s\.\.\.`, invalidAppName, orgName, spaceName, userName))
				Eventually(session.Err).Should(Say("App '%s' not found", invalidAppName))
				Eventually(session).Should(Say("FAILED"))

				Eventually(session).Should(Exit(1))
			})
		})
	})
})
