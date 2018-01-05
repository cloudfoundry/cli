package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

// skipping until refactor
var _ = XDescribe("apps command", func() {
	var (
		orgName   string
		spaceName string
		appName1  string
		appName2  string
	)

	BeforeEach(func() {
		orgName = helpers.NewOrgName()
		spaceName = helpers.NewSpaceName()
		appName1 = helpers.NewAppName()
		appName2 = helpers.NewAppName()
	})

	Describe("help", func() {
		Context("when --help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("apps", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("apps - List all apps in the target space"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf apps"))
				Eventually(session).Should(Say("ALIAS:"))
				Eventually(session).Should(Say("a"))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("events, logs, map-route, push, restart, scale, start, stop"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("when the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "apps")
		})
	})

	Context("when the environment is set up correctly", func() {
		var userName string

		BeforeEach(func() {
			setupCF(orgName, spaceName)
			userName, _ = helpers.GetCredentials()
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		Context("with no apps", func() {
			It("displays empty list", func() {
				session := helpers.CF("apps")
				Eventually(session).Should(Say("Getting apps in org %s / space %s as %s\\.\\.\\.", orgName, spaceName, userName))
				Eventually(session).Should(Say("No apps found"))
				Eventually(session).Should(Exit(0))
			})
		})

		Context("with existing apps", func() {
			var domainName string

			BeforeEach(func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "v2-push", appName1)).Should(Exit(0))
					Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "v2-push", appName2)).Should(Exit(0))
				})

				domainName = defaultSharedDomain()
			})

			It("displays apps in the list", func() {
				session := helpers.CF("apps")
				Eventually(session).Should(Say("Getting apps in org %s / space %s as %s\\.\\.\\.", orgName, spaceName, userName))
				Eventually(session).Should(Say("name\\s+requested state\\s+instances\\s+memory\\s+disk\\s+urls"))
				Eventually(session).Should(Say("%s\\s+started\\s+1/1\\s+8M\\s+8M\\s+%s\\.%s", appName1, appName1, domainName))
				Eventually(session).Should(Say("%s\\s+started\\s+1/1\\s+8M\\s+8M\\s+%s\\.%s", appName2, appName2, domainName))

				Eventually(session).Should(Exit(0))
			})

			Context("when one app is stopped", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("stop", appName1)).Should(Exit(0))
				})

				It("displays app as stopped", func() {
					session := helpers.CF("apps")
					Eventually(session).Should(Say("Getting apps in org %s / space %s as %s\\.\\.\\.", orgName, spaceName, userName))
					Eventually(session).Should(Say("name\\s+requested state\\s+instances\\s+memory\\s+disk\\s+urls"))
					Eventually(session).Should(Say("%s\\s+stopped\\s+1/1\\s+8M\\s+8M\\s+%s\\.%s", appName1, appName1, domainName))
					Eventually(session).Should(Say("%s\\s+started\\s+1/1\\s+8M\\s+8M\\s+%s\\.%s", appName2, appName2, domainName))

					Eventually(session).Should(Exit(0))
				})
			})
		})
	})
})
