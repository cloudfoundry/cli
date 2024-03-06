package isolated

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/integration/helpers/servicebrokerstub"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("bind-service command", func() {
	Describe("help", func() {
		When("--help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("bind-service", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("bind-service - Bind a service instance to an app"))

				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf bind-service APP_NAME SERVICE_INSTANCE \\[-c PARAMETERS_AS_JSON\\] \\[--binding-name BINDING_NAME\\]"))
				Eventually(session).Should(Say("Optionally provide service-specific configuration parameters in a valid JSON object in-line:"))
				Eventually(session).Should(Say("cf bind-service APP_NAME SERVICE_INSTANCE -c '{\"name\":\"value\",\"name\":\"value\"}'"))
				Eventually(session).Should(Say("Optionally provide a file containing service-specific configuration parameters in a valid JSON object."))
				Eventually(session).Should(Say("The path to the parameters file can be an absolute or relative path to a file."))
				Eventually(session).Should(Say("cf bind-service APP_NAME SERVICE_INSTANCE -c PATH_TO_FILE"))
				Eventually(session).Should(Say("Example of valid JSON object:"))
				Eventually(session).Should(Say("{"))
				Eventually(session).Should(Say("\"permissions\": \"read-only\""))
				Eventually(session).Should(Say("}"))
				Eventually(session).Should(Say("Optionally provide a binding name for the association between an app and a service instance:"))
				Eventually(session).Should(Say("cf bind-service APP_NAME SERVICE_INSTANCE --binding-name BINDING_NAME"))
				Eventually(session).Should(Say("EXAMPLES:"))
				Eventually(session).Should(Say("Linux/Mac:"))
				Eventually(session).Should(Say("cf bind-service myapp mydb -c '{\"permissions\":\"read-only\"}'"))
				Eventually(session).Should(Say("Windows Command Line:"))
				Eventually(session).Should(Say("cf bind-service myapp mydb -c \"{\\\\\"permissions\\\\\":\\\\\"read-only\\\\\"}\""))
				Eventually(session).Should(Say("Windows PowerShell:"))
				Eventually(session).Should(Say("cf bind-service myapp mydb -c '{\\\\\"permissions\\\\\":\\\\\"read-only\\\\\"}'"))
				Eventually(session).Should(Say("cf bind-service myapp mydb -c ~/workspace/tmp/instance_config.json --binding-name BINDING_NAME"))
				Eventually(session).Should(Say("ALIAS:"))
				Eventually(session).Should(Say("bs"))
				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say("--binding-name\\s+Name to expose service instance to app process with \\(Default: service instance name\\)"))
				Eventually(session).Should(Say("-c\\s+Valid JSON object containing service-specific configuration parameters, provided either in-line or in a file. For a list of supported configuration parameters, see documentation for the particular service offering."))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("services"))
				Eventually(session).Should(Exit(0))
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

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "bind-service", "app-name", "service-name")
		})
	})

	When("provided invalid flag values", func() {
		When("the --binding-name flag is provided and its value is the empty string", func() {
			It("returns an invalid usage error and the help text", func() {
				session := helpers.CF("bind-service", appName, serviceInstance, "--binding-name", "")
				Eventually(session.Err).Should(Say("--binding-name must be at least 1 character in length"))

				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Exit(1))
			})
		})
	})

	When("the environment is setup correctly", func() {
		var (
			org      string
			space    string
			username string
		)

		BeforeEach(func() {
			org = helpers.NewOrgName()
			space = helpers.NewSpaceName()
			username, _ = helpers.GetCredentials()

			helpers.SetupCF(org, space)
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(org)
		})

		When("the app does not exist", func() {
			It("displays FAILED and app not found", func() {
				session := helpers.CF("bind-service", "does-not-exist", serviceInstance)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("App '%s' not found", "does-not-exist"))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the app exists", func() {
			BeforeEach(func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CF("push", appName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
				})
			})

			When("the service does not exist", func() {
				It("displays FAILED and service not found", func() {
					session := helpers.CF("bind-service", appName, "does-not-exist")
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("Service instance %s not found", "does-not-exist"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("the service exists", func() {
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
					Eventually(session).Should(Say("Binding service %s to app %s in org %s / space %s as %s...", serviceInstance, appName, org, space, username))

					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Say("TIP: Use 'cf restage %s' to ensure your env variable changes take effect", appName))
					Eventually(session).Should(Exit(0))
				})

				When("the service is already bound to an app", func() {
					BeforeEach(func() {
						Eventually(helpers.CF("bind-service", appName, serviceInstance)).Should(Exit(0))
					})

					It("displays OK and that the app is already bound to the service", func() {
						session := helpers.CF("bind-service", appName, serviceInstance)

						Eventually(session).Should(Say("Binding service %s to app %s in org %s / space %s as %s...", serviceInstance, appName, org, space, username))
						Eventually(session).Should(Say("App %s is already bound to %s.", appName, serviceInstance))
						Eventually(session).Should(Say("OK"))

						Eventually(session).Should(Exit(0))
					})
				})

				When("the --binding-name flag is provided and the value is a non-empty string", func() {
					It("binds the service to the app, displays OK and TIP", func() {
						session := helpers.CF("bind-service", appName, serviceInstance, "--binding-name", "i-am-a-binding")
						Eventually(session.Out).Should(Say("Binding service %s to app %s with binding name %s in org %s / space %s as %s...", serviceInstance, appName, "i-am-a-binding", org, space, username))

						Eventually(session.Out).Should(Say("OK"))
						Eventually(session.Out).Should(Say("TIP: Use 'cf restage %s' to ensure your env variable changes take effect", appName))
						Eventually(session).Should(Exit(0))
					})
				})

				When("configuration parameters are provided in a file", func() {
					var configurationFile *os.File

					When("the file-path does not exist", func() {
						It("displays FAILED and the invalid configuration error", func() {
							session := helpers.CF("bind-service", appName, serviceInstance, "-c", "i-do-not-exist")
							Eventually(session.Err).Should(Say("Invalid configuration provided for -c flag. Please provide a valid JSON object or path to a file containing a valid JSON object."))

							Eventually(session).Should(Exit(1))
						})
					})

					When("the file contians invalid json", func() {
						BeforeEach(func() {
							var err error
							content := []byte("{i-am-very-bad-json")
							configurationFile, err = ioutil.TempFile("", "CF_CLI")
							Expect(err).ToNot(HaveOccurred())

							_, err = configurationFile.Write(content)
							Expect(err).ToNot(HaveOccurred())

							err = configurationFile.Close()
							Expect(err).ToNot(HaveOccurred())
						})

						AfterEach(func() {
							Expect(os.RemoveAll(configurationFile.Name())).ToNot(HaveOccurred())
						})

						It("displays FAILED and the invalid configuration error", func() {
							session := helpers.CF("bind-service", appName, serviceInstance, "-c", configurationFile.Name())
							Eventually(session.Err).Should(Say("Invalid configuration provided for -c flag. Please provide a valid JSON object or path to a file containing a valid JSON object."))

							Eventually(session).Should(Exit(1))
						})
					})

					When("the file-path is relative", func() {
						BeforeEach(func() {
							var err error
							content := []byte("{\"i-am-good-json\":\"good-boy\"}")
							configurationFile, err = ioutil.TempFile("", "CF_CLI")
							Expect(err).ToNot(HaveOccurred())

							_, err = configurationFile.Write(content)
							Expect(err).ToNot(HaveOccurred())

							err = configurationFile.Close()
							Expect(err).ToNot(HaveOccurred())
						})

						AfterEach(func() {
							Expect(os.RemoveAll(configurationFile.Name())).ToNot(HaveOccurred())
						})

						It("binds the service to the app, displays OK and TIP", func() {
							session := helpers.CF("bind-service", appName, serviceInstance, "-c", configurationFile.Name())
							Eventually(session).Should(Say("Binding service %s to app %s in org %s / space %s as %s...", serviceInstance, appName, org, space, username))

							Eventually(session).Should(Say("OK"))
							Eventually(session).Should(Say("TIP: Use 'cf restage %s' to ensure your env variable changes take effect", appName))
							Eventually(session).Should(Exit(0))
						})
					})

					When("the file-path is absolute", func() {
						BeforeEach(func() {
							var err error
							content := []byte("{\"i-am-good-json\":\"good-boy\"}")
							configurationFile, err = ioutil.TempFile("", "CF_CLI")
							Expect(err).ToNot(HaveOccurred())

							_, err = configurationFile.Write(content)
							Expect(err).ToNot(HaveOccurred())

							err = configurationFile.Close()
							Expect(err).ToNot(HaveOccurred())
						})

						AfterEach(func() {
							Expect(os.RemoveAll(configurationFile.Name())).ToNot(HaveOccurred())
						})

						It("binds the service to the app, displays OK and TIP", func() {
							absolutePath, err := filepath.Abs(configurationFile.Name())
							Expect(err).ToNot(HaveOccurred())
							session := helpers.CF("bind-service", appName, serviceInstance, "-c", absolutePath)
							Eventually(session).Should(Say("Binding service %s to app %s in org %s / space %s as %s...", serviceInstance, appName, org, space, username))

							Eventually(session).Should(Say("OK"))
							Eventually(session).Should(Say("TIP: Use 'cf restage %s' to ensure your env variable changes take effect", appName))
							Eventually(session).Should(Exit(0))
						})
					})
				})

				When("configuration paramters are provided as in-line JSON", func() {
					When("the JSON is invalid", func() {
						It("displays FAILED and the invalid configuration error", func() {
							session := helpers.CF("bind-service", appName, serviceInstance, "-c", "i-am-invalid-json")
							Eventually(session.Err).Should(Say("Invalid configuration provided for -c flag. Please provide a valid JSON object or path to a file containing a valid JSON object."))

							Eventually(session).Should(Exit(1))
						})
					})

					When("the JSON is valid", func() {
						It("binds the service to the app, displays OK and TIP", func() {
							session := helpers.CF("bind-service", appName, serviceInstance, "-c", "{\"i-am-valid-json\":\"dope dude\"}")
							Eventually(session).Should(Say("Binding service %s to app %s in org %s / space %s as %s...", serviceInstance, appName, org, space, username))

							Eventually(session).Should(Say("OK"))
							Eventually(session).Should(Say("TIP: Use 'cf restage %s' to ensure your env variable changes take effect", appName))
							Eventually(session).Should(Exit(0))
						})
					})
				})
			})

			When("the service is provided by a broker", func() {
				var broker *servicebrokerstub.ServiceBrokerStub

				AfterEach(func() {
					broker.Forget()
				})

				When("the service binding is blocking", func() {
					BeforeEach(func() {
						broker = servicebrokerstub.EnableServiceAccess()

						Eventually(helpers.CF("create-service", broker.FirstServiceOfferingName(), broker.FirstServicePlanName(), serviceInstance)).Should(Exit(0))
					})

					It("binds the service to the app, displays OK and TIP", func() {
						session := helpers.CF("bind-service", appName, serviceInstance, "-c", `{"wheres":"waldo"}`)
						Eventually(session).Should(Say("Binding service %s to app %s in org %s / space %s as %s...", serviceInstance, appName, org, space, username))

						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Say("TIP: Use 'cf restage %s' to ensure your env variable changes take effect", appName))
						Eventually(session).Should(Exit(0))

						session = helpers.CF("service", serviceInstance)
						Eventually(session).Should(Say(appName))
						Eventually(session).Should(Exit(0))
					})
				})

				When("the service binding is asynchronous", func() {
					BeforeEach(func() {
						broker = servicebrokerstub.New().WithAsyncDelay(time.Millisecond).EnableServiceAccess()

						Eventually(helpers.CF("create-service", broker.FirstServiceOfferingName(), broker.FirstServicePlanName(), serviceInstance)).Should(Exit(0))

						Eventually(func() *Session {
							session := helpers.CF("service", serviceInstance)
							return session.Wait()
						}, time.Minute*5, time.Second*5).Should(Say("create succeeded"))
					})

					It("binds the service to the app, displays OK and TIP", func() {
						session := helpers.CF("bind-service", appName, serviceInstance, "-c", `{"wheres":"waldo"}`)
						Eventually(session).Should(Say("Binding service %s to app %s in org %s / space %s as %s...", serviceInstance, appName, org, space, username))

						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Say("Binding in progress. Use 'cf service %s' to check operation status.", serviceInstance))
						Eventually(session).Should(Say("TIP: Once this operation succeeds, use 'cf restage %s' to ensure your env variable changes take effect", appName))
						Eventually(session).Should(Exit(0))

						session = helpers.CF("service", serviceInstance)
						Eventually(session).Should(Say(appName))
						Eventually(session).Should(Exit(0))
					})
				})
			})
		})
	})
})
