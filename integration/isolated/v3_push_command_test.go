package isolated

import (
	"os"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("v3-push command", func() {
	var (
		orgName   string
		spaceName string
		appName   string
		userName  string
	)

	BeforeEach(func() {
		// TMP: this command also depends on https://www.pivotaltracker.com/story/show/146469509
		Skip("don't run in the pipeline until cf-deployment master supports it")

		orgName = helpers.NewOrgName()
		spaceName = helpers.NewSpaceName()
		appName = helpers.PrefixedRandomName("app")
		userName, _ = helpers.GetCredentials()
	})

	Describe("help", func() {
		Context("when --help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("v3-push", "--help")
				Eventually(session.Out).Should(Say("NAME:"))
				Eventually(session.Out).Should(Say("v3-push - Push a new app or sync changes to an existing app"))
				Eventually(session.Out).Should(Say("USAGE:"))
				Eventually(session.Out).Should(Say("cf v3-push -n APP_NAME"))
				Eventually(session.Out).Should(Say("OPTIONS:"))
				Eventually(session.Out).Should(Say("--name, -n\\s+The application name to push"))

				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("when the name flag is missing", func() {
		It("displays incorrect usage", func() {
			session := helpers.CF("v3-push")

			Eventually(session.Err).Should(Say("Incorrect Usage: the required flag `-n, --name' was not specified"))
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
				session := helpers.CF("v3-push", "--name", appName)
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
				session := helpers.CF("v3-push", "--name", appName)
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
				session := helpers.CF("v3-push", "--name", appName)
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
				session := helpers.CF("v3-push", "--name", appName)
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
			var session *Session
			BeforeEach(func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					err := os.Chdir(appDir)
					Expect(err).ToNot(HaveOccurred())

					Eventually(helpers.CF("v3-push", "--name", appName)).Should(Exit(0))
				})

				helpers.WithHelloWorldApp(func(appDir string) {
					err := os.Chdir(appDir)
					Expect(err).ToNot(HaveOccurred())

					session = helpers.CF("v3-push", "--name", appName)
					Eventually(session).Should(Exit(0))
				})
			})

			It("pushes the app", func() {
				Eventually(session.Out).Should(Say("Updating app %s in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, userName))
				Eventually(session.Out).Should(Say("OK"))
				Eventually(session.Out).Should(Say(""))
				Eventually(session.Out).Should(Say("Uploading app %s in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, userName))
				Eventually(session.Out).Should(Say("OK"))
				Eventually(session.Out).Should(Say(""))
				Eventually(session.Out).Should(Say("Staging package for app %s in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, userName))
				Eventually(session.Out).Should(Say(" Downloading staticfile_buildpack\\.\\.\\."))
				Eventually(session.Out).Should(Say(" Creating container"))
				Eventually(session.Out).Should(Say(" Successfully created container"))
				Eventually(session.Out).Should(Say(" Downloading app package"))
				Eventually(session.Out).Should(Say(" Downloaded app package"))
				Eventually(session.Out).Should(Say(" Staging\\.\\.\\."))
				Eventually(session.Out).Should(Say(" Exit status 0"))
				Eventually(session.Out).Should(Say(" Staging complete"))
				Eventually(session.Out).Should(Say(" Uploading droplet, build artifacts cache\\.\\.\\."))
				Eventually(session.Out).Should(Say(" Creating droplet for app with guid"))
				Eventually(session.Out).Should(Say(" Uploading complete"))
				Eventually(session.Out).Should(Say(" Destroying container"))
				Eventually(session.Out).Should(Say("droplet:"))
				Eventually(session.Out).Should(Say("OK"))
				Eventually(session.Out).Should(Say("Stopping app %s in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, userName))
				Eventually(session.Out).Should(Say("OK"))
				Eventually(session.Out).Should(Say(""))
				Eventually(session.Out).Should(Say("Setting app %s to droplet .+ in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, userName))
				Eventually(session.Out).Should(Say("OK"))
				Eventually(session.Out).Should(Say(""))
				Eventually(session.Out).Should(Say("Starting app %s in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, userName))
				Eventually(session.Out).Should(Say("OK"))
				Eventually(session.Out).Should(Say(""))
				Eventually(session.Out).Should(Say("Waiting for app to start\\.\\.\\."))
				Eventually(session.Out).Should(Say("Showing health and status for app %s in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, userName))
				Eventually(session.Out).Should(Say(""))
				Eventually(session.Out).Should(Say("name:\\s+%s", appName))
				Eventually(session.Out).Should(Say("requested state:\\s+started"))
				Eventually(session.Out).Should(Say("processes:\\s+web:1/1"))
				Eventually(session.Out).Should(Say("memory usage:\\s+32M x 1"))
				Eventually(session.Out).Should(Say("stack:\\s+cflinuxfs2"))
				Eventually(session.Out).Should(Say("buildpacks:\\s+staticfile"))
				Eventually(session.Out).Should(Say(""))
				Eventually(session.Out).Should(Say("web"))
				Eventually(session.Out).Should(Say(`state\s+since\s+cpu\s+memory\s+disk`))
				Eventually(session.Out).Should(Say("#0\\s+running\\s+\\d{4}-\\d{2}-\\d{2} \\d{2}:\\d{2}:\\d{2} [AP]M"))
			})
		})

		Context("when the app does not already exist", func() {
			var session *Session

			BeforeEach(func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					err := os.Chdir(appDir)
					Expect(err).ToNot(HaveOccurred())

					session = helpers.CF("v3-push", "--name", appName)
					Eventually(session).Should(Exit(0))
				})
			})

			It("pushes the app", func() {
				Eventually(session.Out).Should(Say("Creating app %s in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, userName))
				Eventually(session.Out).Should(Say("OK"))
				Eventually(session.Out).Should(Say(""))
				Eventually(session.Out).Should(Say("Uploading app %s in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, userName))
				Eventually(session.Out).Should(Say("OK"))
				Eventually(session.Out).Should(Say(""))
				Eventually(session.Out).Should(Say("Staging package for app %s in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, userName))
				Eventually(session.Out).Should(Say(" Downloading staticfile_buildpack\\.\\.\\."))
				Eventually(session.Out).Should(Say(" Creating container"))
				Eventually(session.Out).Should(Say(" Successfully created container"))
				Eventually(session.Out).Should(Say(" Downloading app package"))
				Eventually(session.Out).Should(Say(" Downloaded app package"))
				Eventually(session.Out).Should(Say(" Staging\\.\\.\\."))
				Eventually(session.Out).Should(Say(" Exit status 0"))
				Eventually(session.Out).Should(Say(" Staging complete"))
				Eventually(session.Out).Should(Say(" Uploading droplet, build artifacts cache\\.\\.\\."))
				Eventually(session.Out).Should(Say(" Creating droplet for app with guid"))
				Eventually(session.Out).Should(Say(" Uploading complete"))
				Eventually(session.Out).Should(Say(" Destroying container"))
				Eventually(session.Out).Should(Say("droplet:"))
				Eventually(session.Out).Should(Say("OK"))
				Consistently(session.Out).ShouldNot(Say("Stopping"))
				Eventually(session.Out).Should(Say("Setting app %s to droplet .+ in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, userName))
				Eventually(session.Out).Should(Say("OK"))
				Eventually(session.Out).Should(Say(""))
				Eventually(session.Out).Should(Say("Starting app %s in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, userName))
				Eventually(session.Out).Should(Say("OK"))
				Eventually(session.Out).Should(Say(""))
				Eventually(session.Out).Should(Say("Waiting for app to start\\.\\.\\."))
				Eventually(session.Out).Should(Say("Showing health and status for app %s in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, userName))
				Eventually(session.Out).Should(Say(""))
				Eventually(session.Out).Should(Say("name:\\s+%s", appName))
				Eventually(session.Out).Should(Say("requested state:\\s+started"))
				Eventually(session.Out).Should(Say("processes:\\s+web:1/1"))
				Eventually(session.Out).Should(Say("memory usage:\\s+32M x 1"))
				Eventually(session.Out).Should(Say("stack:\\s+cflinuxfs2"))
				Eventually(session.Out).Should(Say("buildpacks:\\s+staticfile"))
				Eventually(session.Out).Should(Say(""))
				Eventually(session.Out).Should(Say("web"))
				Eventually(session.Out).Should(Say(`state\s+since\s+cpu\s+memory\s+disk`))
				Eventually(session.Out).Should(Say("#0\\s+running\\s+\\d{4}-\\d{2}-\\d{2} \\d{2}:\\d{2}:\\d{2} [AP]M"))
			})
		})
	})
})
