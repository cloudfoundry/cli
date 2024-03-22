package isolated

import (
	"os"
	"time"

	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/integration/helpers/servicebrokerstub"
	"code.cloudfoundry.org/cli/resources"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("bind-service command", func() {
	const command = "bind-service"

	Describe("help", func() {
		matchHelpMessage := SatisfyAll(
			Say(`NAME:\n`),
			Say(`\s+bind-service - Bind a service instance to an app\n`),
			Say(`\n`),
			Say(`USAGE:\n`),
			Say(`\s+cf bind-service APP_NAME SERVICE_INSTANCE \[-c PARAMETERS_AS_JSON\] \[--binding-name BINDING_NAME\]\n`),
			Say(`\n`),
			Say(`\s+Optionally provide service-specific configuration parameters in a valid JSON object in-line:\n`),
			Say(`\n`),
			Say(`\s+cf bind-service APP_NAME SERVICE_INSTANCE -c '\{"name":"value","name":"value"\}'\n`),
			Say(`\n`),
			Say(`\s+Optionally provide a file containing service-specific configuration parameters in a valid JSON object.\n`),
			Say(`\s+The path to the parameters file can be an absolute or relative path to a file.\n`),
			Say(`\s+cf bind-service APP_NAME SERVICE_INSTANCE -c PATH_TO_FILE\n`),
			Say(`\n`),
			Say(`\s+Example of valid JSON object:\n`),
			Say(`\s+{\n`),
			Say(`\s+"permissions": "read-only"\n`),
			Say(`\s+}\n`),
			Say(`\n`),
			Say(`\s+Optionally provide a binding name for the association between an app and a service instance:\n`),
			Say(`\n`),
			Say(`\s+cf bind-service APP_NAME SERVICE_INSTANCE --binding-name BINDING_NAME\n`),
			Say(`\n`),
			Say(`EXAMPLES:\n`),
			Say(`\s+Linux/Mac:\n`),
			Say(`\s+cf bind-service myapp mydb -c '\{"permissions":"read-only"\}'\n`),
			Say(`\n`),
			Say(`\s+Windows Command Line:\n`),
			Say(`\s+cf bind-service myapp mydb -c "\{\\"permissions\\":\\"read-only\\"\}"\n`),
			Say(`\n`),
			Say(`\s+Windows PowerShell:\n`),
			Say(`\s+cf bind-service myapp mydb -c '\{\\"permissions\\":\\"read-only\\"\}'\n`),
			Say(`\n`),
			Say(`\s+cf bind-service myapp mydb -c ~/workspace/tmp/instance_config.json --binding-name BINDING_NAME\n`),
			Say(`\n`),
			Say(`ALIAS:\n`),
			Say(`\s+bs\n`),
			Say(`\n`),
			Say(`OPTIONS:\n`),
			Say(`\s+--binding-name\s+Name to expose service instance to app process with \(Default: service instance name\)\n`),
			Say(`\s+-c\s+Valid JSON object containing service-specific configuration parameters, provided either in-line or in a file. For a list of supported configuration parameters, see documentation for the particular service offering.\n`),
			Say(`\s+--wait, -w\s+Wait for the operation to complete\n`),
			Say(`\n`),
			Say(`SEE ALSO:\n`),
			Say(`\s+services\n`),
		)

		When("the -h flag is specified", func() {
			It("succeeds and prints help", func() {
				session := helpers.CF(command, "-h")
				Eventually(session).Should(Exit(0))
				Expect(session.Out).To(matchHelpMessage)
			})
		})

		When("the --help flag is specified", func() {
			It("succeeds and prints help", func() {
				session := helpers.CF(command, "--help")
				Eventually(session).Should(Exit(0))
				Expect(session.Out).To(matchHelpMessage)
			})
		})

		When("no arguments are provided", func() {
			It("displays a warning, the help text, and exits 1", func() {
				session := helpers.CF(command)
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Incorrect Usage: the required arguments `APP_NAME` and `SERVICE_INSTANCE` were not provided"))
				Expect(session.Out).To(matchHelpMessage)
			})
		})

		When("unknown flag is passed", func() {
			It("displays a warning, the help text, and exits 1", func() {
				session := helpers.CF(command, "-u")
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Incorrect Usage: unknown flag `u"))
				Expect(session.Out).To(matchHelpMessage)
			})
		})

		When("-c is provided with invalid JSON", func() {
			It("displays a warning, the help text, and exits 1", func() {
				session := helpers.CF(command, "-c", `{"not":json"}`)
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Incorrect Usage: Invalid configuration provided for -c flag. Please provide a valid JSON object or path to a file containing a valid JSON object."))
				Expect(session.Out).To(matchHelpMessage)
			})
		})

		When("-c is provided with invalid JSON file", func() {
			It("displays a warning, the help text, and exits 1", func() {
				filename := helpers.TempFileWithContent(`{"not":json"}`)
				defer os.Remove(filename)

				session := helpers.CF(command, "-c", filename)
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Incorrect Usage: Invalid configuration provided for -c flag. Please provide a valid JSON object or path to a file containing a valid JSON object."))
				Expect(session.Out).To(matchHelpMessage)
			})
		})

		When("--binding-name is provided with empty value", func() {
			It("displays a warning, the help text, and exits 1", func() {
				session := helpers.CF(command, "appName", "serviceInstanceName", "--binding-name", "")
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Incorrect Usage: --binding-name must be at least 1 character in length"))
				Expect(session.Out).To(matchHelpMessage)
			})
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "bind-service", "app-name", "service-name")
		})
	})

	When("targeting a space", func() {
		var (
			orgName   string
			spaceName string
			username  string
		)

		getBinding := func(serviceInstanceName string) resources.ServiceCredentialBinding {
			var receiver struct {
				Resources []resources.ServiceCredentialBinding `json:"resources"`
			}
			helpers.Curl(&receiver, "/v3/service_credential_bindings?service_instance_names=%s", serviceInstanceName)
			Expect(receiver.Resources).To(HaveLen(1))
			return receiver.Resources[0]
		}

		getParameters := func(serviceInstanceName string) (receiver map[string]interface{}) {
			binding := getBinding(serviceInstanceName)
			helpers.Curl(&receiver, "/v3/service_credential_bindings/%s/parameters", binding.GUID)
			return
		}

		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()
			helpers.SetupCF(orgName, spaceName)

			username, _ = helpers.GetCredentials()
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		Context("user-provided route service", func() {
			var (
				serviceInstanceName string
				appName             string
				bindingName         string
			)

			BeforeEach(func() {
				serviceInstanceName = helpers.NewServiceInstanceName()
				Eventually(helpers.CF("cups", serviceInstanceName)).Should(Exit(0))

				appName = helpers.NewAppName()
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CF("push", appName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
				})

				bindingName = helpers.RandomName()
			})

			It("creates a binding", func() {
				session := helpers.CF(command, appName, serviceInstanceName, "--binding-name", bindingName)
				Eventually(session).Should(Exit(0))

				Expect(session.Out).To(SatisfyAll(
					Say(`Binding service instance %s to app %s in org %s / space %s as %s\.\.\.\n`, serviceInstanceName, appName, orgName, spaceName, username),
					Say(`OK\n`),
					Say(`\n`),
					Say(`TIP: Use 'cf restage %s' to ensure your env variable changes take effect`, appName),
				))

				Expect(string(session.Err.Contents())).To(BeEmpty())

				binding := getBinding(serviceInstanceName)
				Expect(binding.Name).To(Equal(bindingName))
				Expect(binding.LastOperation.State).To(BeEquivalentTo("succeeded"))
			})

			When("parameters are specified", func() {
				It("fails with an error returned by the CC", func() {
					session := helpers.CF(command, appName, serviceInstanceName, "-c", `{"foo":"bar"}`)
					Eventually(session).Should(Exit(1))

					Expect(session.Out).To(SatisfyAll(
						Say(`Binding service instance %s to app %s in org %s / space %s as %s\.\.\.\n`, serviceInstanceName, appName, orgName, spaceName, username),
						Say(`FAILED\n`),
					))

					Expect(session.Err).To(Say(`Binding parameters are not supported for user-provided service instances\n`))
				})
			})
		})

		Context("managed service instance with synchronous broker response", func() {
			var (
				broker              *servicebrokerstub.ServiceBrokerStub
				serviceInstanceName string
				appName             string
				bindingName         string
			)

			BeforeEach(func() {
				broker = servicebrokerstub.EnableServiceAccess()
				serviceInstanceName = helpers.NewServiceInstanceName()
				helpers.CreateManagedServiceInstance(broker.FirstServiceOfferingName(), broker.FirstServicePlanName(), serviceInstanceName)

				appName = helpers.NewAppName()
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CF("push", appName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
				})

				bindingName = helpers.RandomName()
			})

			AfterEach(func() {
				broker.Forget()
			})

			It("creates a binding", func() {
				session := helpers.CF(command, appName, serviceInstanceName, "--binding-name", bindingName)
				Eventually(session).Should(Exit(0))

				Expect(session.Out).To(SatisfyAll(
					Say(`Binding service instance %s to app %s in org %s / space %s as %s\.\.\.\n`, serviceInstanceName, appName, orgName, spaceName, username),
					Say(`OK\n`),
					Say(`\n`),
					Say(`TIP: Use 'cf restage %s' to ensure your env variable changes take effect`, appName),
				))

				Expect(string(session.Err.Contents())).To(BeEmpty())

				binding := getBinding(serviceInstanceName)
				Expect(binding.Name).To(Equal(bindingName))
				Expect(binding.LastOperation.State).To(BeEquivalentTo("succeeded"))
			})

			When("parameters are specified", func() {
				It("sends the parameters to the broker", func() {
					session := helpers.CF(command, appName, serviceInstanceName, "-c", `{"foo":"bar"}`)
					Eventually(session).Should(Exit(0))

					Expect(getParameters(serviceInstanceName)).To(Equal(map[string]interface{}{"foo": "bar"}))
				})
			})
		})

		Context("managed service instance with asynchronous broker response", func() {
			var (
				broker              *servicebrokerstub.ServiceBrokerStub
				serviceInstanceName string
				appName             string
				bindingName         string
			)

			BeforeEach(func() {
				broker = servicebrokerstub.New().WithAsyncDelay(time.Second).EnableServiceAccess()
				serviceInstanceName = helpers.NewServiceInstanceName()
				helpers.CreateManagedServiceInstance(broker.FirstServiceOfferingName(), broker.FirstServicePlanName(), serviceInstanceName)

				appName = helpers.NewAppName()
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CF("push", appName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
				})

				bindingName = helpers.RandomName()
			})

			AfterEach(func() {
				broker.Forget()
			})

			It("start to create a binding", func() {
				session := helpers.CF(command, appName, serviceInstanceName, "--binding-name", bindingName)
				Eventually(session).Should(Exit(0))

				Expect(session.Out).To(SatisfyAll(
					Say(`Binding service instance %s to app %s in org %s / space %s as %s\.\.\.\n`, serviceInstanceName, appName, orgName, spaceName, username),
					Say(`OK\n`),
					Say(`\n`),
					Say(`Binding in progress. Use 'cf service %s' to check operation status\.\n`, serviceInstanceName),
					Say(`\n`),
					Say(`TIP: Once this operation succeeds, use 'cf restage %s' to ensure your env variable changes take effect`, appName),
				))

				Expect(string(session.Err.Contents())).To(BeEmpty())

				binding := getBinding(serviceInstanceName)
				Expect(binding.Name).To(Equal(bindingName))
				Expect(binding.LastOperation.State).To(BeEquivalentTo("in progress"))
			})

			When("--wait flag specified", func() {
				It("waits for completion", func() {
					session := helpers.CF(command, appName, serviceInstanceName, "--binding-name", bindingName, "--wait")
					Eventually(session).Should(Exit(0))

					Expect(session.Out).To(SatisfyAll(
						Say(`Binding service instance %s to app %s in org %s / space %s as %s\.\.\.\n`, serviceInstanceName, appName, orgName, spaceName, username),
						Say(`Waiting for the operation to complete\.+\n`),
						Say(`\n`),
						Say(`OK\n`),
					))

					Expect(string(session.Err.Contents())).To(BeEmpty())

					Expect(getBinding(serviceInstanceName).LastOperation.State).To(BeEquivalentTo("succeeded"))
				})
			})
		})

		Context("binding already exists", func() {
			var (
				serviceInstanceName string
				appName             string
			)

			BeforeEach(func() {
				serviceInstanceName = helpers.NewServiceInstanceName()
				Eventually(helpers.CF("cups", serviceInstanceName)).Should(Exit(0))

				appName = helpers.NewAppName()
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CF("push", appName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
				})

				Eventually(helpers.CF(command, appName, serviceInstanceName)).Should(Exit(0))
			})

			It("says OK", func() {
				session := helpers.CF(command, appName, serviceInstanceName)
				Eventually(session).Should(Exit(0))

				Expect(session.Out).To(SatisfyAll(
					Say(`Binding service instance %s to app %s in org %s / space %s as %s\.\.\.\n`, serviceInstanceName, appName, orgName, spaceName, username),
					Say(`App %s is already bound to service instance %s.\n`, appName, serviceInstanceName),
					Say(`OK\n`),
				))

				Expect(string(session.Err.Contents())).To(BeEmpty())
			})
		})

		Context("app does not exist", func() {
			var serviceInstanceName string

			BeforeEach(func() {
				serviceInstanceName = helpers.NewServiceInstanceName()
				Eventually(helpers.CF("cups", serviceInstanceName)).Should(Exit(0))
			})

			It("displays FAILED and app not found", func() {
				session := helpers.CF(command, "does-not-exist", serviceInstanceName)
				Eventually(session).Should(Exit(1))
				Expect(session.Out).To(Say("FAILED"))
				Expect(session.Err).To(Say("App 'does-not-exist' not found"))
			})
		})

		Context("service instance does not exist", func() {
			var appName string

			BeforeEach(func() {
				appName = helpers.NewAppName()
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CF("push", appName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
				})
			})

			It("displays FAILED and service not found", func() {
				session := helpers.CF(command, appName, "does-not-exist")
				Eventually(session).Should(Exit(1))
				Expect(session.Out).To(Say("FAILED"))
				Expect(session.Err).To(Say("Service instance 'does-not-exist' not found"))
			})
		})
	})
})
