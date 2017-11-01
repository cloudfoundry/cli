package experimental

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	. "github.com/onsi/gomega/ghttp"
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
		Context("when --help flag is set", func() {
			It("displays command usage to output", func() {
				session := helpers.CF("v3-env", "--help")

				Eventually(session.Out).Should(Say("NAME:"))
				Eventually(session.Out).Should(Say("v3-env - Show all env variables for an app"))
				Eventually(session.Out).Should(Say("USAGE:"))
				Eventually(session.Out).Should(Say("cf v3-env APP_NAME"))
				Eventually(session.Out).Should(Say("SEE ALSO:"))
				Eventually(session.Out).Should(Say("running-environment-variable-group, staging-environment-variable-group, v3-app, v3-apps, v3-set-env, v3-unset-env"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("when the app name is not provided", func() {
		It("tells the user that the app name is required, prints help text, and exits 1", func() {
			session := helpers.CF("v3-env")

			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `APP_NAME` was not provided"))
			Eventually(session.Out).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	It("displays the experimental warning", func() {
		session := helpers.CF("v3-env", appName)
		Eventually(session.Out).Should(Say("This command is in EXPERIMENTAL stage and may change without notice"))
		Eventually(session).Should(Exit())
	})

	Context("when the environment is not setup correctly", func() {
		Context("when the v3 api does not exist", func() {
			var server *Server

			BeforeEach(func() {
				server = helpers.StartAndTargetServerWithoutV3API()
			})

			AfterEach(func() {
				server.Close()
			})

			It("fails with error message that the minimum version is not met", func() {
				session := helpers.CF("v3-env", appName, envVarName, envVarValue)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("This command requires CF API version 3\\.27\\.0 or higher\\."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the v3 api version is lower than the minimum version", func() {
			var server *Server

			BeforeEach(func() {
				server = helpers.StartAndTargetServerWithV3Version("3.0.0")
			})

			AfterEach(func() {
				server.Close()
			})

			It("fails with error message that the minimum version is not met", func() {
				session := helpers.CF("v3-env", appName, envVarName, envVarValue)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("This command requires CF API version 3\\.27\\.0 or higher\\."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when no API endpoint is set", func() {
			BeforeEach(func() {
				helpers.UnsetAPI()
			})

			It("fails with no API endpoint set message", func() {
				session := helpers.CF("v3-env", appName, envVarName, envVarValue)
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
				session := helpers.CF("v3-env", appName, envVarName, envVarValue)
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
				session := helpers.CF("v3-env", appName, envVarName, envVarValue)
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
				session := helpers.CF("v3-env", appName, envVarName, envVarValue)
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("No space targeted, use 'cf target -s SPACE' to target a space\\."))
				Eventually(session).Should(Exit(1))
			})
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

		Context("when the app does not exist", func() {
			It("displays app not found and exits 1", func() {
				invalidAppName := "invalid-app-name"
				session := helpers.CF("v3-env", invalidAppName)

				Eventually(session.Out).Should(Say("Getting env variables for app %s in org %s / space %s as %s\\.\\.\\.", invalidAppName, orgName, spaceName, userName))
				Eventually(session.Err).Should(Say("App %s not found", invalidAppName))
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the app exists", func() {
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
				Eventually(session.Out).Should(Say("Getting env variables for app %s in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, userName))
				Eventually(session.Out).Should(Say("System-Provided:"))
				Eventually(session.Out).Should(Say("VCAP_SERVICES"))
				Eventually(session.Out).Should(Say("VCAP_APPLICATION"))

				Eventually(session.Out).Should(Say("User-Provided:"))
				Eventually(session.Out).Should(Say(`user-provided-env-name: "user-provided-env-value"`))

				Eventually(session.Out).Should(Say("Running Environment Variable Groups:"))
				Eventually(session.Out).Should(Say(`running-env-name: "running-env-value"`))

				Eventually(session.Out).Should(Say("Staging Environment Variable Groups:"))
				Eventually(session.Out).Should(Say(`staging-env-name: "staging-env-value"`))
				Eventually(session).Should(Exit(0))

				By("displaying help messages when they are not set")
				Eventually(helpers.CF("v3-unset-env", appName, "user-provided-env-name")).Should(Exit(0))
				Eventually(helpers.CF("unbind-service", appName, userProvidedServiceName)).Should(Exit(0))
				Eventually(helpers.CF("set-running-environment-variable-group", `{}`)).Should(Exit(0))
				Eventually(helpers.CF("set-staging-environment-variable-group", `{}`)).Should(Exit(0))

				session = helpers.CF("v3-env", appName)
				Eventually(session.Out).Should(Say("Getting env variables for app %s in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, userName))
				Eventually(session.Out).Should(Say("System-Provided:"))
				Eventually(session.Out).ShouldNot(Say("VCAP_SERVICES"))
				Eventually(session.Out).Should(Say("VCAP_APPLICATION"))

				Eventually(session.Out).Should(Say("No user-provided env variables have been set"))

				Eventually(session.Out).Should(Say("No running env variables have been set"))

				Eventually(session.Out).Should(Say("No staging env variables have been set"))
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
