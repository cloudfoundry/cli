package isolated

import (
	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("ssh-enabled command", func() {
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
			It("appears in cf help -a", func() {
				session := helpers.CF("help", "-a")
				Eventually(session).Should(Exit(0))
				Expect(session).To(HaveCommandInCategoryWithDescription("ssh-enabled", "APPS", "Reports whether SSH is enabled on an application container instance"))
			})

			It("displays command usage to output", func() {
				session := helpers.CF("ssh-enabled", "--help")

				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("ssh-enabled - Reports whether SSH is enabled on an application container instance"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf ssh-enabled APP_NAME"))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("enable-ssh, space-ssh-allowed, ssh"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the app name is not provided", func() {
		It("tells the user that the app name is required, prints help text, and exits 1", func() {
			session := helpers.CF("ssh-enabled")

			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `APP_NAME` was not provided"))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	When("the environment is set up correctly", func() {

		BeforeEach(func() {
			helpers.SetupCF(orgName, spaceName)
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		When("the app does not exist", func() {
			It("displays app not found and exits 1", func() {
				invalidAppName := "invalid-app-name"
				session := helpers.CF("ssh-enabled", invalidAppName)

				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Say("App %s not found", invalidAppName))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the app exists", func() {
			// Consider using cf curl /v3/apps
			BeforeEach(func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CF("push", appName, "-p", appDir, "--no-start")).Should(Exit(0))
				})
			})

			When("when ssh is enabled", func() {
				It("informs the user ssh is disabled", func() {
					session := helpers.CF("ssh-enabled", appName)

					Eventually(session).Should(Say("ssh support is enabled for '%s'", appName))
					Eventually(session).Should(Exit(0))
				})
			})

			When("ssh is disabled", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("disable-ssh", appName)).Should(Exit(0))
				})

				It("informs the user and exits 0", func() {
					session := helpers.CF("ssh-enabled", appName)

					Eventually(session).Should(Say("ssh support is disabled for '%s'", appName))
				})
			})
		})
	})
})
