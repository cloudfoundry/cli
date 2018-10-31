package isolated

import (
	"fmt"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("unset-env command", func() {
	var (
		orgName    string
		spaceName  string
		appName    string
		envVarName string
	)

	BeforeEach(func() {
		orgName = helpers.NewOrgName()
		spaceName = helpers.NewSpaceName()
		appName = helpers.PrefixedRandomName("app")
		envVarName = "SOME_ENV_VAR"
	})

	Describe("help", func() {
		When("--help flag is set", func() {
			It("displays command usage to output", func() {
				session := helpers.CF("unset-env", "--help")

				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("unset-env - Remove an env variable from an app"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf unset-env APP_NAME ENV_VAR_NAME"))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("env, set-env, v3-apps, v3-restart, v3-stage"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the app name is not provided", func() {
		It("tells the user that the app name is required, prints help text, and exits 1", func() {
			session := helpers.CF("unset-env")

			Eventually(session.Err).Should(Say("Incorrect Usage: the required arguments `APP_NAME` and `ENV_VAR_NAME` were not provided"))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	When("ENV_VAR_NAME is not provided", func() {
		It("tells the user that ENV_VAR_NAME is required, prints help text, and exits 1", func() {
			session := helpers.CF("unset-env", appName)

			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `ENV_VAR_NAME` was not provided"))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "unset-env", appName, envVarName)
		})
	})

	When("the environment is set up correctly", func() {
		var userName string

		BeforeEach(func() {
			helpers.SetupCF(orgName, spaceName)
			userName, _ = helpers.GetCredentials()
		})

		When("the app does not exist", func() {
			It("displays app not found and exits 1", func() {
				invalidAppName := "invalid-app-name"
				session := helpers.CF("unset-env", invalidAppName, envVarName)

				Eventually(session).Should(Say(`Removing env variable %s from app %s in org %s / space %s as %s\.\.\.`, envVarName, invalidAppName, orgName, spaceName, userName))
				Eventually(session.Err).Should(Say("App %s not found", invalidAppName))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the app exists", func() {
			BeforeEach(func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CF("push", appName, "-p", appDir)).Should(Exit(0))
				})
			})

			When("the environment variable has not been previously set", func() {
				It("returns a warning indicating variable was not set", func() {
					session := helpers.CF("unset-env", appName, envVarName)

					Eventually(session).Should(Say(`Removing env variable %s from app %s in org %s / space %s as %s\.\.\.`, envVarName, appName, orgName, spaceName, userName))
					Eventually(session).Should(Say("Env variable %s was not set.", envVarName))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))
				})
			})

			When("the environment variable has been previously set", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("set-env", appName, envVarName, "some-value")).Should(Exit(0))
				})

				It("overrides the value of the existing environment variable", func() {
					session := helpers.CF("unset-env", appName, envVarName)

					Eventually(session).Should(Say(`Removing env variable %s from app %s in org %s / space %s as %s\.\.\.`, envVarName, appName, orgName, spaceName, userName))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Say(`TIP: Use 'cf v3-stage %s' to ensure your env variable changes take effect\.`, appName))

					session = helpers.CF("curl", fmt.Sprintf("v3/apps/%s/environment_variables", helpers.AppGUID(appName)))
					Eventually(session).ShouldNot(Say(`"%s"`, envVarName))
					Eventually(session).Should(Exit(0))
				})
			})
		})
	})
})
