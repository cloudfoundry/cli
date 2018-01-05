package isolated

import (
	"fmt"
	"net/http"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Logs Command", func() {
	Describe("help", func() {
		It("displays command usage to output", func() {
			session := helpers.CF("logs", "--help")
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Say("logs - Tail or show recent logs for an app"))
			Eventually(session).Should(Say("USAGE:"))
			Eventually(session).Should(Say("cf logs APP_NAME"))
			Eventually(session).Should(Say("OPTIONS:"))
			Eventually(session).Should(Say("--recent\\s+Dump recent logs instead of tailing"))
			Eventually(session).Should(Say("SEE ALSO:"))
			Eventually(session).Should(Say("app, apps, ssh"))
			Eventually(session).Should(Exit(0))
		})
	})

	Context("when the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "logs", "app-name")
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

		Context("when input is invalid", func() {
			Context("because no app name is provided", func() {
				It("gives an incorrect usage message", func() {
					session := helpers.CF("logs")
					Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `APP_NAME` was not provided"))
					Eventually(session).Should(Say("NAME:"))
					Eventually(session).Should(Say("logs - Tail or show recent logs for an app"))
					Eventually(session).Should(Say("USAGE:"))
					Eventually(session).Should(Say("cf logs APP_NAME"))
					Eventually(session).Should(Say("OPTIONS:"))
					Eventually(session).Should(Say("--recent\\s+Dump recent logs instead of tailing"))
					Eventually(session).Should(Say("SEE ALSO:"))
					Eventually(session).Should(Say("app, apps, ssh"))
					Eventually(session).Should(Exit(1))
				})
			})

			Context("because the app does not exist", func() {
				It("fails with an app not found message", func() {
					session := helpers.CF("logs", "dora")
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("App dora not found"))
					Eventually(session).Should(Exit(1))
				})
			})
		})

		Context("when the specified app exists", func() {
			var appName string

			BeforeEach(func() {
				appName = helpers.PrefixedRandomName("app")
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CF("push", appName, "-p", appDir, "-b", "staticfile_buildpack", "-u", "http")).Should(Exit(0))
				})
			})

			Context("without the --recent flag", func() {
				It("streams logs out to the screen", func() {
					session := helpers.CF("logs", appName)
					defer session.Terminate()

					userName, _ := helpers.GetCredentials()
					Eventually(session).Should(Say("Retrieving logs for app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName))

					response, err := http.Get(fmt.Sprintf("http://%s.%s", appName, defaultSharedDomain()))
					Expect(err).NotTo(HaveOccurred())
					Expect(response.StatusCode).To(Equal(http.StatusOK))
					Eventually(session).Should(Say("%s \\[APP/PROC/WEB/0\\]\\s+OUT .*? \"GET / HTTP/1.1\" 200 \\d+", helpers.ISO8601Regex))
				})
			})

			Context("with the --recent flag", func() {
				It("displays the most recent logs and closes the stream", func() {
					session := helpers.CF("logs", appName, "--recent")
					userName, _ := helpers.GetCredentials()
					Eventually(session).Should(Say("Retrieving logs for app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName))
					Eventually(session).Should(Say("%s \\[API/\\d+\\]\\s+OUT Created app with guid %s", helpers.ISO8601Regex, helpers.GUIDRegex))
					Eventually(session).Should(Exit(0))
				})
			})
		})
	})
})
