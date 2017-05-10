package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

// TODO: Change error outputs to session.Err after refactor
// as well as any other suggested changes in tracker stories:
// https://www.pivotaltracker.com/story/show/144626837
// https://www.pivotaltracker.com/story/show/126256765
var _ = Describe("bind-service command", func() {
	BeforeEach(func() {
		helpers.RunIfExperimental("command is currently experimental")
	})

	Describe("help", func() {
		Context("when --help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("bind-service", "--help")
				Eventually(session.Out).Should(Say("NAME:"))
				Eventually(session.Out).Should(Say("bind-service - Bind a service instance to an app"))

				Eventually(session.Out).Should(Say("USAGE:"))
				Eventually(session.Out).Should(Say("cf bind-service APP_NAME SERVICE_INSTANCE \\[-c PARAMETERS_AS_JSON\\]"))
				Eventually(session.Out).Should(Say("Optionally provide service-specific configuration parameters in a valid JSON object in-line:"))
				Eventually(session.Out).Should(Say("cf bind-service APP_NAME SERVICE_INSTANCE -c '{\"name\":\"value\",\"name\":\"value\"}'"))
				Eventually(session.Out).Should(Say("Optionally provide a file containing service-specific configuration parameters in a valid JSON object."))
				Eventually(session.Out).Should(Say("The path to the parameters file can be an absolute or relative path to a file."))
				Eventually(session.Out).Should(Say("cf bind-service APP_NAME SERVICE_INSTANCE -c PATH_TO_FILE"))
				Eventually(session.Out).Should(Say("Example of valid JSON object:"))
				Eventually(session.Out).Should(Say("{"))
				Eventually(session.Out).Should(Say("\"permissions\": \"read-only\""))
				Eventually(session.Out).Should(Say("}"))
				Eventually(session.Out).Should(Say("EXAMPLES:"))
				Eventually(session.Out).Should(Say("Linux/Mac:"))
				Eventually(session.Out).Should(Say("cf bind-service myapp mydb -c '{\"permissions\":\"read-only\"}'"))
				Eventually(session.Out).Should(Say("Windows Command Line:"))
				Eventually(session.Out).Should(Say("cf bind-service myapp mydb -c \"{\\\\\"permissions\\\\\":\\\\\"read-only\\\\\"}\""))
				Eventually(session.Out).Should(Say("Windows PowerShell:"))
				Eventually(session.Out).Should(Say("cf bind-service myapp mydb -c '{\\\\\"permissions\\\\\":\\\\\"read-only\\\\\"}'"))
				Eventually(session.Out).Should(Say("cf bind-service myapp mydb -c ~/workspace/tmp/instance_config.json"))
				Eventually(session.Out).Should(Say("ALIAS:"))
				Eventually(session.Out).Should(Say("bs"))
				Eventually(session.Out).Should(Say("OPTIONS:"))
				Eventually(session.Out).Should(Say("-c      Valid JSON object containing service-specific configuration parameters, provided either in-line or in a file. For a list of supported configuration parameters, see documentation for the particular service offering."))
				Eventually(session.Out).Should(Say("SEE ALSO:"))
				Eventually(session.Out).Should(Say("services"))
			})
		})
	})

	var (
		serviceInstance string
		appName         string
	)

	BeforeEach(func() {
		serviceInstance = helpers.PrefixedRandomName("si")
		appName = helpers.PrefixedRandomName("app")
	})

	Context("when the environment is not setup correctly", func() {
		Context("when no API endpoint is set", func() {
			BeforeEach(func() {
				helpers.UnsetAPI()
			})

			It("fails with no API endpoint set message", func() {
				session := helpers.CF("bind-service", appName, serviceInstance)
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Out).Should(Say("No API endpoint set. Use 'cf login' or 'cf api' to target an endpoint."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when not logged in", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
			})

			It("fails with not logged in message", func() {
				session := helpers.CF("bind-service", appName, serviceInstance)
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Out).Should(Say("Not logged in. Use 'cf login' to log in."))
				Eventually(session).Should(Exit(1))
			})
		})

		// TODO: Figure out this wack shi
		XContext("when there no org set", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
				helpers.LoginCF()
			})

			It("fails with no targeted org error message", func() {
				session := helpers.CF("bind-service", appName, serviceInstance)
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Out).Should(Say("No org targeted, use 'cf target -o ORG' to target an org."))
				Eventually(session).Should(Exit(1))
			})
		})

		// TODO: Figure out this wack shi
		XContext("when there no space set", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
				helpers.LoginCF()
				helpers.TargetOrg(ReadOnlyOrg)
			})

			It("fails with no targeted space error message", func() {
				session := helpers.CF("bind-service", appName, serviceInstance)
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Out).Should(Say("No space targeted, use 'cf target -s SPACE' to target a space."))
				Eventually(session).Should(Exit(1))
			})
		})
	})

	Context("when the environment is setup correctly", func() {
		var (
			org         string
			space       string
			service     string
			servicePlan string
			domain      string
			username    string
		)

		BeforeEach(func() {
			org = helpers.NewOrgName()
			space = helpers.NewSpaceName()
			service = helpers.PrefixedRandomName("SERVICE")
			servicePlan = helpers.PrefixedRandomName("SERVICE-PLAN")
			username, _ = helpers.GetCredentials()

			setupCF(org, space)
			domain = defaultSharedDomain()
		})

		Context("when the app does not exist", func() {
			It("displays FAILED and app not found", func() {
				session := helpers.CF("bind-service", "does-not-exist", serviceInstance)
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Out).Should(Say("App %s not found", "does-not-exist"))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the app exists", func() {
			BeforeEach(func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CF("push", appName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
				})
			})

			Context("when the service does not exist", func() {
				It("displays FAILED and service not found", func() {
					session := helpers.CF("bind-service", appName, "does-not-exist")
					Eventually(session.Out).Should(Say("FAILED"))
					Eventually(session.Out).Should(Say("Service instance %s not found", "does-not-exist"))
					Eventually(session).Should(Exit(1))
				})
			})

			Context("when the service exists", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("create-user-provided-service", serviceInstance, "-p", "{}")).Should(Exit(0))
					helpers.WithHelloWorldApp(func(appDir string) {
						Eventually(helpers.CF("push", appName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
					})
				})

				AfterEach(func() {
					Eventually(helpers.CF("unbind-service", appName, serviceInstance)).Should(Exit(0))
					Eventually(helpers.CF("delete-service", serviceInstance, "-f")).Should(Exit(0))
				})

				It("binds the service to the app, displays OK and TIP", func() {
					session := helpers.CF("bind-service", appName, serviceInstance)
					Eventually(session.Out).Should(Say("Binding service %s to app %s in org %s / space %s as %s...", serviceInstance, appName, org, space, username))

					Eventually(session.Out).Should(Say("OK"))
					Eventually(session.Out).Should(Say("TIP: Use 'cf restage %s' to ensure your env variable changes take effect", appName))
					Eventually(session).Should(Exit(0))
				})

				FContext("when the service is already bound to an app", func() {
					BeforeEach(func() {
						Eventually(helpers.CF("bind-service", appName, serviceInstance)).Should(Exit(0))
					})

					It("displays OK and that the app is already bound to the service", func() {
						session := helpers.CF("bind-service", appName, serviceInstance)

						Eventually(session.Out).Should(Say("Binding service %s to app %s in org %s / space %s as %s...", serviceInstance, appName, org, space, username))
						Eventually(session.Out).Should(Say("OK"))
						Eventually(session.Out).Should(Say("App %s is already bound to %s.", appName, serviceInstance))

						Eventually(session).Should(Exit(0))
					})
				})

				Context("when configuration parameters are provided in a file", func() {
					Context("when the file-path does not exist", func() {
						It("displays FAILED and the invalid configuration error", func() {
						})
					})

					Context("when the file-path is relative", func() {
						It("binds the service to the app, displays OK and TIP", func() {
						})
					})

					Context("when the file-path is absolute", func() {
						It("binds the service to the app, displays OK and TIP", func() {
						})
					})
				})

				Context("when configuration paramters are provided as in-line JSON", func() {
					Context("when the JSON is invalid", func() {
						It("displays FAILED and the  invalid configuration error", func() {
						})
					})

					Context("when the JSON is valid", func() {
						It("binds the service to the app, displays OK and TIP", func() {
						})
					})
				})
			})

			Context("when the service doesn't allow apps to bind", func() {
				It("displays FAILED and service server error", func() {
				})
			})
		})
	})
})
