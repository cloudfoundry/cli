package experimental

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("v3-env command", func() {
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
				session := helpers.CF("v3-env", "--help")

				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("v3-env - Show all env variables for an app"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf v3-env APP_NAME"))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("running-environment-variable-group, staging-environment-variable-group, v3-app, v3-apps, v3-set-env, v3-unset-env"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the app name is not provided", func() {
		It("tells the user that the app name is required, prints help text, and exits 1", func() {
			session := helpers.CF("v3-env")

			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `APP_NAME` was not provided"))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	It("displays the experimental warning", func() {
		session := helpers.CF("v3-env", appName)
		Eventually(session.Err).Should(Say("This command is in EXPERIMENTAL stage and may change without notice"))
		Eventually(session).Should(Exit())
	})

	When("the environment is not setup correctly", func() {
		When("no API endpoint is set", func() {
			BeforeEach(func() {
				helpers.UnsetAPI()
			})

			It("fails with no API endpoint set message", func() {
				session := helpers.CF("v3-env", appName, envVarName, envVarValue)
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
				session := helpers.CF("v3-env", appName, envVarName, envVarValue)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say(`Not logged in\. Use 'cf login' or 'cf login --sso' to log in\.`))
				Eventually(session).Should(Exit(1))
			})
		})

		When("there is no org set", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
				helpers.LoginCF()
			})

			It("fails with no org targeted error message", func() {
				session := helpers.CF("v3-env", appName, envVarName, envVarValue)
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
				session := helpers.CF("v3-env", appName, envVarName, envVarValue)
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

		When("the app does not exist", func() {
			It("displays app not found and exits 1", func() {
				invalidAppName := "invalid-app-name"
				session := helpers.CF("v3-env", invalidAppName)

				Eventually(session).Should(Say(`Getting env variables for app %s in org %s / space %s as %s\.\.\.`, invalidAppName, orgName, spaceName, userName))
				Eventually(session.Err).Should(Say("App '%s' not found", invalidAppName))
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
					Eventually(helpers.CF("v3-push", appName, "-p", appDir)).Should(Exit(0))
				})

				Eventually(helpers.CF("v3-set-env", appName, "user-provided-env-name", "user-provided-env-value")).Should(Exit(0))
				Eventually(helpers.CF("create-user-provided-service", userProvidedServiceName)).Should(Exit(0))
				Eventually(helpers.CF("bind-service", appName, userProvidedServiceName)).Should(Exit(0))
				Eventually(helpers.CF("set-running-environment-variable-group", `{"running-env-name": "running-env-value"}`)).Should(Exit(0))
				Eventually(helpers.CF("set-staging-environment-variable-group", `{"staging-env-name": "staging-env-value"}`)).Should(Exit(0))
			})

			AfterEach(func() {
				Eventually(helpers.CF("unbind-service", appName, userProvidedServiceName)).Should(Exit(0))
				Eventually(helpers.CF("delete-service", userProvidedServiceName)).Should(Exit(0))
				Eventually(helpers.CF("set-running-environment-variable-group", `{}`)).Should(Exit(0))
				Eventually(helpers.CF("set-staging-environment-variable-group", `{}`)).Should(Exit(0))
			})

			It("displays the environment variables", func() {
				By("displaying env variables when they are set")
				session := helpers.CF("v3-env", appName)
				Eventually(session).Should(Say(`Getting env variables for app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
				Eventually(session).Should(Say("System-Provided:"))
				Eventually(session).Should(Say("VCAP_SERVICES"))
				Eventually(session).Should(Say("VCAP_APPLICATION"))

				Eventually(session).Should(Say("User-Provided:"))
				Eventually(session).Should(Say(`user-provided-env-name: "user-provided-env-value"`))

				Eventually(session).Should(Say("Running Environment Variable Groups:"))
				Eventually(session).Should(Say(`running-env-name: "running-env-value"`))

				Eventually(session).Should(Say("Staging Environment Variable Groups:"))
				Eventually(session).Should(Say(`staging-env-name: "staging-env-value"`))
				Eventually(session).Should(Exit(0))

				By("displaying help messages when they are not set")
				Eventually(helpers.CF("v3-unset-env", appName, "user-provided-env-name")).Should(Exit(0))
				Eventually(helpers.CF("unbind-service", appName, userProvidedServiceName)).Should(Exit(0))
				Eventually(helpers.CF("set-running-environment-variable-group", `{}`)).Should(Exit(0))
				Eventually(helpers.CF("set-staging-environment-variable-group", `{}`)).Should(Exit(0))

				session = helpers.CF("v3-env", appName)
				Eventually(session).Should(Exit(0))
				Expect(session).Should(Say(`Getting env variables for app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
				Expect(session).Should(Say("System-Provided:"))
				Expect(session).Should(Say(`VCAP_SERVICES: \{\}`))
				Expect(session).Should(Say("VCAP_APPLICATION"))

				Expect(session).Should(Say("No user-provided env variables have been set"))

				Expect(session).Should(Say("No running env variables have been set"))

				Expect(session).Should(Say("No staging env variables have been set"))
			})
		})
	})
})
