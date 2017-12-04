package experimental

import (
	"fmt"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("v3-set-env command", func() {
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
				session := helpers.CF("v3-set-env", "--help")

				Eventually(session.Out).Should(Say("NAME:"))
				Eventually(session.Out).Should(Say("v3-set-env - Set an env variable for an app"))
				Eventually(session.Out).Should(Say("USAGE:"))
				Eventually(session.Out).Should(Say("cf v3-set-env APP_NAME ENV_VAR_NAME ENV_VAR_VALUE"))
				Eventually(session.Out).Should(Say("SEE ALSO:"))
				Eventually(session.Out).Should(Say("set-running-environment-variable-group, set-staging-environment-variable-group, v3-apps, v3-env, v3-restart, v3-stage, v3-unset-env"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("when the app name is not provided", func() {
		It("tells the user that the app name is required, prints help text, and exits 1", func() {
			session := helpers.CF("v3-set-env")

			Eventually(session.Err).Should(Say("Incorrect Usage: the required arguments `APP_NAME`, `ENV_VAR_NAME` and `ENV_VAR_VALUE` were not provided"))
			Eventually(session.Out).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	Context("when ENV_VAR_NAME is not provided", func() {
		It("tells the user that ENV_VAR_NAME is required, prints help text, and exits 1", func() {
			session := helpers.CF("v3-set-env", appName)

			Eventually(session.Err).Should(Say("Incorrect Usage: the required arguments `ENV_VAR_NAME` and `ENV_VAR_VALUE` were not provided"))
			Eventually(session.Out).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	Context("when the ENV_VAR_VALUE is not provided", func() {
		It("tells the user that ENV_VAR_VALUE is required, prints help text, and exits 1", func() {
			session := helpers.CF("v3-set-env", appName, envVarName)

			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `ENV_VAR_VALUE` was not provided"))
			Eventually(session.Out).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	It("displays the experimental warning", func() {
		session := helpers.CF("v3-set-env", appName, envVarName, envVarValue)
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
				session := helpers.CF("v3-set-env", appName, envVarName, envVarValue)
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
				session := helpers.CF("v3-set-env", appName, envVarName, envVarValue)
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
				session := helpers.CF("v3-set-env", appName, envVarName, envVarValue)
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
				session := helpers.CF("v3-set-env", appName, envVarName, envVarValue)
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
				session := helpers.CF("v3-set-env", appName, envVarName, envVarValue)
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
				session := helpers.CF("v3-set-env", appName, envVarName, envVarValue)
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
				session := helpers.CF("v3-set-env", invalidAppName, envVarName, envVarValue)

				Eventually(session.Out).Should(Say("Setting env variable %s for app %s in org %s / space %s as %s\\.\\.\\.", envVarName, invalidAppName, orgName, spaceName, userName))
				Eventually(session.Err).Should(Say("App %s not found", invalidAppName))
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the app exists", func() {
			BeforeEach(func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CF("v3-push", appName, "-p", appDir)).Should(Exit(0))
				})
			})

			Context("when the environment variable has not been previously set", func() {
				It("sets the environment variable value pair", func() {
					session := helpers.CF("v3-set-env", appName, envVarName, envVarValue)

					Eventually(session.Out).Should(Say("Setting env variable %s for app %s in org %s / space %s as %s\\.\\.\\.", envVarName, appName, orgName, spaceName, userName))
					Eventually(session.Out).Should(Say("OK"))
					Eventually(session.Out).Should(Say("TIP: Use 'cf v3-stage %s' to ensure your env variable changes take effect\\.", appName))
					Eventually(session).Should(Exit(0))

					session = helpers.CF("curl", fmt.Sprintf("v3/apps/%s/environment_variables", helpers.AppGUID(appName)))
					Eventually(session.Out).Should(Say(`"%s": "%s"`, envVarName, envVarValue))
					Eventually(session).Should(Exit(0))
				})

				// This is to prevent the '-' being read in as another flag
				Context("when the environment variable value starts with a '-' character", func() {
					BeforeEach(func() {
						envVarValue = "-" + envVarValue
					})

					It("sets the environment variable value pair", func() {
						session := helpers.CF("v3-set-env", appName, envVarName, envVarValue)

						Eventually(session.Out).Should(Say("Setting env variable %s for app %s in org %s / space %s as %s\\.\\.\\.", envVarName, appName, orgName, spaceName, userName))
						Eventually(session.Out).Should(Say("OK"))
						Eventually(session.Out).Should(Say("TIP: Use 'cf v3-stage %s' to ensure your env variable changes take effect\\.", appName))
						Eventually(session).Should(Exit(0))

						session = helpers.CF("curl", fmt.Sprintf("v3/apps/%s/environment_variables", helpers.AppGUID(appName)))
						Eventually(session.Out).Should(Say(`"%s": "%s"`, envVarName, envVarValue))
						Eventually(session).Should(Exit(0))
					})
				})
			})

			Context("when the environment variable has been previously set", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("v3-set-env", appName, envVarName, envVarValue)).Should(Exit(0))
				})

				It("overrides the value of the existing environment variable", func() {
					someOtherValue := "some-other-value"
					session := helpers.CF("v3-set-env", appName, envVarName, someOtherValue)

					Eventually(session.Out).Should(Say("Setting env variable %s for app %s in org %s / space %s as %s\\.\\.\\.", envVarName, appName, orgName, spaceName, userName))
					Eventually(session.Out).Should(Say("OK"))
					Eventually(session.Out).Should(Say("TIP: Use 'cf v3-stage %s' to ensure your env variable changes take effect\\.", appName))
					Eventually(session).Should(Exit(0))

					session = helpers.CF("curl", fmt.Sprintf("v3/apps/%s/environment_variables", helpers.AppGUID(appName)))
					Eventually(session.Out).Should(Say(`"%s": "%s"`, envVarName, someOtherValue))
					Eventually(session).Should(Exit(0))
				})
			})

		})
	})
})
