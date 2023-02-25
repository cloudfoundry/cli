package isolated

import (
	"fmt"

	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("enable-ssh command", func() {
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
				Expect(session).To(HaveCommandInCategoryWithDescription("enable-ssh", "APPS", "Enable ssh for the application"))
			})

			It("displays command usage to output", func() {
				session := helpers.CF("enable-ssh", "--help")

				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("enable-ssh - Enable ssh for the application"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf enable-ssh APP_NAME"))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("allow-space-ssh, space-ssh-allowed, ssh, ssh-enabled"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the app name is not provided", func() {
		It("tells the user that the app name is required, prints help text, and exits 1", func() {
			session := helpers.CF("enable-ssh")

			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `APP_NAME` was not provided"))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "enable-ssh", appName)
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

		When("the app does not exist", func() {
			It("displays app not found and exits 1", func() {
				invalidAppName := "invalid-app-name"
				session := helpers.CF("enable-ssh", invalidAppName)

				Eventually(session).Should(Say(`Enabling ssh support for app %s as %s\.\.\.`, invalidAppName, userName))
				Eventually(session.Err).Should(Say("App '%s' not found", invalidAppName))
				Eventually(session).Should(Say("FAILED"))
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

			When("when ssh has not been enabled yet", func() {
				It("enables ssh for the app", func() {
					session := helpers.CF("enable-ssh", appName)

					Eventually(session).Should(Say(`Enabling ssh support for app %s as %s\.\.\.`, appName, userName))

					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))

					session = helpers.CF("curl", fmt.Sprintf("v3/apps/%s/ssh_enabled", helpers.AppGUID(appName)))
					Eventually(session).Should(Exit(0))

					bytes := session.Out.Contents()

					actualEnablementValue := helpers.GetsEnablementValue(bytes)
					Expect(actualEnablementValue).To(Equal(true))
				})
			})

			When("ssh was previously enabled for the app", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("enable-ssh", appName)).Should(Exit(0))
				})

				It("informs the user and exits 0", func() {
					session := helpers.CF("enable-ssh", appName)

					Eventually(session).Should(Say(`Enabling ssh support for app %s as %s\.\.\.`, appName, userName))

					Eventually(session).Should(Say("ssh support for app '%s' is already enabled.", appName))
					Eventually(session).Should(Say("OK"))

					session = helpers.CF("curl", fmt.Sprintf("v3/apps/%s/ssh_enabled", helpers.AppGUID(appName)))
					Eventually(session).Should(Exit(0))

					bytes := session.Out.Contents()

					actualEnablementValue := helpers.GetsEnablementValue(bytes)
					Expect(actualEnablementValue).To(Equal(true))
				})
			})

			When("ssh is disabled space-wide", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("disallow-space-ssh", spaceName)).Should(Exit(0))
				})

				It("informs the user", func() {
					session := helpers.CF("enable-ssh", appName)

					Eventually(session).Should(Say(`Enabling ssh support for app %s as %s\.\.\.`, appName, userName))

					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Say("TIP: Ensure ssh is also enabled on the space and global level."))
					Eventually(session).Should(Exit(0))
				})
			})

			// We're not testing the failure case when ssh is globally disabled;
			// It would require us to redeploy CAPI with the cloud_controller_ng job property
			// cc.allow_app_ssh_access set to false, which is time-consuming & brittle
			// https://github.com/cloudfoundry/capi-release/blob/8d08628fecf116a6a1512d8ad6413414b092fba8/jobs/cloud_controller_ng/spec#L754-L756

		})
	})
})
