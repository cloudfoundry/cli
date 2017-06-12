package isolated

import (
	"fmt"
	"os"
	"regexp"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("v3-app command", func() {
	var (
		orgName   string
		spaceName string
		appName   string
	)

	BeforeEach(func() {
		// TMP: this command also depends on https://www.pivotaltracker.com/story/show/146469509
		Skip("don't run in the pipeline until cf-deployment master supports it")

		orgName = helpers.NewOrgName()
		spaceName = helpers.NewSpaceName()
		appName = helpers.PrefixedRandomName("app")
	})

	Describe("help", func() {
		Context("when --help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("v3-app", "--help")

				Eventually(session.Out).Should(Say("NAME:"))
				Eventually(session.Out).Should(Say("v3-app - Display an app"))
				Eventually(session.Out).Should(Say("USAGE:"))
				Eventually(session.Out).Should(Say("cf v3-app -n APP_NAME"))
				Eventually(session.Out).Should(Say("OPTIONS:"))
				Eventually(session.Out).Should(Say("--name, -n\\s+The application name to display"))

				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("when the name flag is missing", func() {
		It("displays incorrect usage", func() {
			session := helpers.CF("v3-app")

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
				session := helpers.CF("v3-app", "--name", appName)
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
				session := helpers.CF("v3-app", "--name", appName)
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
				session := helpers.CF("v3-app", "--name", appName)
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
				session := helpers.CF("v3-app", "--name", appName)
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
				var packageGUID string
				Eventually(helpers.CF("v3-create-app", "--name", appName)).Should(Exit(0))

				prevDir, err := os.Getwd()
				Expect(err).ToNot(HaveOccurred())

				helpers.WithHelloWorldApp(func(appDir string) {
					err := os.Chdir(appDir)
					Expect(err).ToNot(HaveOccurred())

					pkgSession := helpers.CF("v3-create-package", "--name", appName)
					Eventually(pkgSession).Should(Exit(0))
					regex, err := regexp.Compile(`package guid: (.+)`)
					Expect(err).ToNot(HaveOccurred())
					matches := regex.FindStringSubmatch(string(pkgSession.Out.Contents()))
					Expect(matches).To(HaveLen(2))

					packageGUID = matches[1]
				})

				err = os.Chdir(prevDir)
				Expect(err).ToNot(HaveOccurred())

				stageSession := helpers.CF("v3-stage", "--name", appName, "--package-guid", packageGUID)
				Eventually(stageSession).Should(Exit(0))

				regex, err := regexp.Compile(`droplet: (.+)`)
				Expect(err).ToNot(HaveOccurred())
				matches := regex.FindStringSubmatch(string(stageSession.Out.Contents()))
				Expect(matches).To(HaveLen(2))

				dropletGUID := matches[1]
				setDropletSession := helpers.CF("v3-set-droplet", "--name", appName, "--droplet-guid", dropletGUID)
				Eventually(setDropletSession).Should(Exit(0))

				Eventually(helpers.CF("v3-start", "-n", appName)).Should(Exit(0))
			})

			It("displays the app summary", func() {
				userName, _ := helpers.GetCredentials()

				session := helpers.CF("v3-app", "-n", appName)
				Eventually(session.Out).Should(Say("Showing health and status for app %s in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, userName))

				Eventually(session.Out).Should(Say(fmt.Sprintf("name:\\s+%s", appName)))
				Eventually(session.Out).Should(Say("requested state:\\s+started"))
				Eventually(session.Out).Should(Say("processes:\\s+web:[01]/1"))
				Eventually(session.Out).Should(Say("buildpacks:\\s+staticfile 1.4.6"))
				Eventually(session.Out).Should(Say("#0\\s+(starting|running)\\s+\\d{4}-\\d{2}-\\d{2} \\d{2}:\\d{2}:\\d{2} [AP]M"))

				Eventually(session).Should(Exit(0))
			})

			Context("when the app is stopped", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("stop", appName)).Should(Exit(0))
				})

				It("displays that there are no running instances of the app", func() {
					userName, _ := helpers.GetCredentials()

					session := helpers.CF("v3-app", "-n", appName)

					Eventually(session.Out).Should(Say(`Showing health and status for app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
					Consistently(session.Out).ShouldNot(Say(`state\s+since\s+cpu\s+memory\s+disk`))
					Eventually(session.Out).Should(Say("There are no running instances of this app"))
				})
			})
		})

		Context("when the app does not exist", func() {
			It("displays app not found and exits 1", func() {
				invalidAppName := "invalid-app-name"
				session := helpers.CF("v3-app", "-n", invalidAppName)
				userName, _ := helpers.GetCredentials()

				Eventually(session.Out).Should(Say("Showing health and status for app %s in org %s / space %s as %s\\.\\.\\.", invalidAppName, orgName, spaceName, userName))
				Eventually(session.Err).Should(Say("App %s not found", invalidAppName))
				Eventually(session.Out).Should(Say("FAILED"))

				Eventually(session).Should(Exit(1))
			})
		})
	})
})
