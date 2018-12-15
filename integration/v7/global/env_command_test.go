package global

import (
	"fmt"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("env command", func() {
	var (
		orgName     string
		spaceName   string
		appName     string
		envVarName  string
		envVarValue string
	)

	BeforeEach(func() {
		orgName = helpers.NewOrgName()
		spaceName = helpers.NewSpaceName()
		appName = helpers.PrefixedRandomName("app")
		envVarName = "SOME_ENV_VAR"
		envVarValue = "SOME_ENV_VAR_VALUE"
	})

	Describe("help", func() {
		When("--help flag is set", func() {
			It("displays command usage to output", func() {
				session := helpers.CF("env", "--help")

				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("env - Show all env variables for an app"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf env APP_NAME"))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("app, running-environment-variable-group, set-env, staging-environment-variable-group, unset-env, v3-apps"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the app name is not provided", func() {
		It("tells the user that the app name is required, prints help text, and exits 1", func() {
			session := helpers.CF("env")

			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `APP_NAME` was not provided"))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "env", appName, envVarName, envVarValue)
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
				session := helpers.CF("env", invalidAppName)

				Eventually(session).Should(Say(`Getting env variables for app %s in org %s / space %s as %s\.\.\.`, invalidAppName, orgName, spaceName, userName))
				Eventually(session).Should(Say("OK"))

				Eventually(session.Err).Should(Say("App %s not found", invalidAppName))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the app exists", func() {
			var (
				userProvidedServiceName string
			)
			BeforeEach(func() {
				userProvidedServiceName = helpers.PrefixedRandomName("service")
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CF("push", appName, "-p", appDir)).Should(Exit(0))
				})

				Eventually(helpers.CF("set-env", appName, "user-provided-env-name", "user-provided-env-value")).Should(Exit(0))
				Eventually(helpers.CF("set-env", appName, "an-out-of-order-name", "some-env-value")).Should(Exit(0))
				Eventually(helpers.CF("set-env", appName, "Capital-env-var", "some-other-env-value")).Should(Exit(0))
				Eventually(helpers.CF("create-user-provided-service", userProvidedServiceName)).Should(Exit(0))
				Eventually(helpers.CF("bind-service", appName, userProvidedServiceName)).Should(Exit(0))
				Eventually(helpers.CF("set-running-environment-variable-group", `{"running-env-name": "running-env-value", "number": 1, "Xstring": "X"}`)).Should(Exit(0))
				Eventually(helpers.CF("set-staging-environment-variable-group", `{"staging-env-name": "staging-env-value", "number": 2, "Ystring": "Y"}`)).Should(Exit(0))
			})

			AfterEach(func() {
				Eventually(helpers.CF("unbind-service", appName, userProvidedServiceName)).Should(Exit(0))
				Eventually(helpers.CF("delete-service", userProvidedServiceName)).Should(Exit(0))
				Eventually(helpers.CF("set-running-environment-variable-group", `{}`)).Should(Exit(0))
				Eventually(helpers.CF("set-staging-environment-variable-group", `{}`)).Should(Exit(0))
			})

			It("displays the environment variables", func() {
				By("displaying env variables when they are set")
				session := helpers.CF("env", appName)

				Eventually(session).Should(Say(fmt.Sprintf(`Getting env variables for app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName)))
				Eventually(session).Should(Say("OK"))

				Eventually(session).Should(Say("System-Provided:"))
				Eventually(session).Should(Say(`VCAP_SERVICES: {\s*\n`))
				Eventually(session).Should(Say("VCAP_APPLICATION"))

				Eventually(session).Should(Say("User-Provided:"))
				Eventually(session).Should(Say(`Capital-env-var:\s+some-other-env-value`))
				Eventually(session).Should(Say(`an-out-of-order-name:\s+some-env-value`))
				Eventually(session).Should(Say(`user-provided-env-name:\s+user-provided-env-value`))

				Eventually(session).Should(Say("Running Environment Variable Groups:"))
				Eventually(session).Should(Say(`Xstring:\s+X`))
				Eventually(session).Should(Say(`number:\s+1`))
				Eventually(session).Should(Say(`running-env-name:\s+running-env-value`))

				Eventually(session).Should(Say("Staging Environment Variable Groups:"))
				Eventually(session).Should(Say(`Ystring:\s+Y`))
				Eventually(session).Should(Say(`number:\s+2`))
				Eventually(session).Should(Say(`staging-env-name:\s+staging-env-value`))

				Eventually(session).Should(Exit(0))

				By("displaying help messages when they are not set")
				Eventually(helpers.CF("unset-env", appName, "user-provided-env-name")).Should(Exit(0))
				Eventually(helpers.CF("unset-env", appName, "an-out-of-order-name")).Should(Exit(0))
				Eventually(helpers.CF("unset-env", appName, "Capital-env-var")).Should(Exit(0))
				Eventually(helpers.CF("unbind-service", appName, userProvidedServiceName)).Should(Exit(0))
				Eventually(helpers.CF("set-running-environment-variable-group", `{}`)).Should(Exit(0))
				Eventually(helpers.CF("set-staging-environment-variable-group", `{}`)).Should(Exit(0))

				session = helpers.CF("env", appName)
				Eventually(session).Should(Say(`Getting env variables for app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
				Eventually(session).Should(Say("OK"))

				Eventually(session).Should(Say("System-Provided:"))
				Eventually(session).Should(Say("VCAP_SERVICES: {}"))
				Eventually(session).Should(Say("VCAP_APPLICATION"))

				Eventually(session).Should(Say("No user-provided env variables have been set"))

				Eventually(session).Should(Say("No running env variables have been set"))

				Eventually(session).Should(Say("No staging env variables have been set"))
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
