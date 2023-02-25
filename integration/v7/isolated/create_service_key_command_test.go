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

var _ = Describe("create-service-key command", func() {
	Describe("help", func() {
		matchHelpMessage := SatisfyAll(
			Say(`NAME:\n`),
			Say(`\s+create-service-key - Create key for a service instance\n`),
			Say(`\n`),
			Say(`USAGE:\n`),
			Say(`\s+cf create-service-key SERVICE_INSTANCE SERVICE_KEY \[-c PARAMETERS_AS_JSON\] \[--wait\]\n`),
			Say(`\s+Optionally provide service-specific configuration parameters in a valid JSON object in-line.\n`),
			Say(`\s+cf create-service-key SERVICE_INSTANCE SERVICE_KEY -c '{\"name\":\"value\",\"name\":\"value\"}'\n`),
			Say(`\s+Optionally provide a file containing service-specific configuration parameters in a valid JSON object. The path to the parameters file can be an absolute or relative path to a file.\n`),
			Say(`\s+cf create-service-key SERVICE_INSTANCE SERVICE_KEY -c PATH_TO_FILE\n`),
			Say(`\s+Example of valid JSON object:\n`),
			Say(`\s+{\n`),
			Say(`\s+\"permissions\": \"read-only\"\n`),
			Say(`\s+}\n`),
			Say(`\n`),
			Say(`EXAMPLES:\n`),
			Say(`\s+cf create-service-key mydb mykey -c '{\"permissions\":\"read-only\"}'\n`),
			Say(`\s+cf create-service-key mydb mykey -c ~/workspace/tmp/instance_config.json\n`),
			Say(`ALIAS:\n`),
			Say(`\s+csk\n`),
			Say(`\n`),
			Say(`OPTIONS:\n`),
			Say(`\s+-c\s+Valid JSON object containing service-specific configuration parameters, provided either in-line or in a file. For a list of supported configuration parameters, see documentation for the particular service offering.\n`),
			Say(`\s+--wait, -w\s+Wait for the operation to complete\n`),
			Say(`\n`),
			Say(`SEE ALSO:\n`),
			Say(`\s+service-key\n`),
		)

		When("--help flag is set", func() {
			It("displays command usage to output", func() {
				session := helpers.CF("create-service-key", "--help")
				Eventually(session).Should(Exit(0))
				Expect(string(session.Err.Contents())).To(HaveLen(0))
				Expect(session.Out).To(matchHelpMessage)
			})
		})

		When("there are no arguments", func() {
			It("fails with a help message", func() {
				session := helpers.CF("create-service-key")
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Incorrect Usage: the required arguments `SERVICE_INSTANCE` and `SERVICE_KEY` were not provided"))
				Expect(session.Out).To(matchHelpMessage)
			})
		})
		When("there are insufficient arguments", func() {
			It("fails with a help message", func() {
				session := helpers.CF("create-service-key", "service-instance")
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Incorrect Usage: the required argument `SERVICE_KEY` was not provided"))
				Expect(session.Out).To(matchHelpMessage)
			})
		})

		When("there are superfluous arguments", func() {
			It("fails with a help message", func() {
				session := helpers.CF("create-service-key", "service-instance", "service-key", "superfluous")
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say(`Incorrect Usage: unexpected argument "superfluous"`))
				Expect(session.Out).To(matchHelpMessage)
			})
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "create-service-key", "service-name", "key-name")
		})
	})

	When("the environment is setup correctly", func() {
		var (
			org                 string
			space               string
			username            string
			serviceInstanceName string
			serviceKeyName      string
		)

		BeforeEach(func() {
			org = helpers.NewOrgName()
			space = helpers.NewSpaceName()
			username, _ = helpers.GetCredentials()

			helpers.SetupCF(org, space)

			serviceInstanceName = helpers.NewServiceInstanceName()
			serviceKeyName = helpers.PrefixedRandomName("KEY")
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(org)
		})

		When("provided with a service instance name that doesn't exist", func() {
			It("displays FAILED and an informative error, and exits 1", func() {
				session := helpers.CF("create-service-key", serviceInstanceName, serviceKeyName)
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("Service instance '%s' not found", serviceInstanceName))
				Eventually(session).Should(Exit(1))
			})
		})

		When("user-provided service instance exists", func() {
			BeforeEach(func() {
				Eventually(helpers.CF("cups", serviceInstanceName)).Should(Exit(0))
			})

			It("fails to create the service key", func() {
				session := helpers.CF("create-service-key", serviceInstanceName, serviceKeyName)
				Eventually(session).Should(Exit(1))

				Expect(session.Out).To(Say(`FAILED\n`))
				Expect(session.Err).To(Say(`Service credential bindings of type 'key' are not supported for user-provided service instances.\n`))
			})
		})

		When("managed service instance exists", func() {
			var broker *servicebrokerstub.ServiceBrokerStub

			BeforeEach(func() {
				broker = servicebrokerstub.EnableServiceAccess()
				helpers.CreateManagedServiceInstance(broker.FirstServiceOfferingName(), broker.FirstServicePlanName(), serviceInstanceName)
			})

			AfterEach(func() {
				broker.Forget()
			})

			It("creates the service key and displays OK", func() {
				session := helpers.CF("create-service-key", serviceInstanceName, serviceKeyName)
				Eventually(session).Should(Exit(0))
				Expect(session.Out).To(SatisfyAll(
					Say(`Creating service key %s for service instance %s as %s\.\.\.\n`, serviceKeyName, serviceInstanceName, username),
					Say(`OK\n`),
				))
				Expect(string(session.Err.Contents())).To(HaveLen(0))
			})

			When("a service key with the provided name already exists", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("create-service-key", serviceInstanceName, serviceKeyName)).Should(Exit(0))
				})

				It("displays OK and an informative error message", func() {
					session := helpers.CF("create-service-key", serviceInstanceName, serviceKeyName)
					Eventually(session).Should(Exit(0))

					Expect(session.Out).To(SatisfyAll(
						Say(`\n`),
						Say(`Service key %s already exists\n`, serviceKeyName),
						Say(`OK\n`),
					))
					Expect(string(session.Err.Contents())).To(HaveLen(0))
				})
			})

			When("the service is not bindable", func() {
				BeforeEach(func() {
					broker.Services[0].Bindable = false
					broker.Configure().Register()
				})

				It("displays FAILED and an informative error, and exits 1", func() {
					session := helpers.CF("create-service-key", serviceInstanceName, serviceKeyName)
					Eventually(session).Should(Exit(1))
					Expect(session.Out).To(Say(`FAILED\n`))
					Expect(session.Err).To(Say(`Service plan does not allow bindings\.\n`))
				})
			})

			Describe("parameters", func() {
				checkParameters := func(params map[string]interface{}) {
					var receiver struct {
						Resources []resources.ServiceCredentialBinding `json:"resources"`
					}
					helpers.Curl(&receiver, "/v3/service_credential_bindings?service_instance_names=%s", serviceInstanceName)
					Expect(receiver.Resources).To(HaveLen(1))

					var parametersReceiver map[string]interface{}
					helpers.Curl(&parametersReceiver, `/v3/service_credential_bindings/%s/parameters`, receiver.Resources[0].GUID)
					Expect(parametersReceiver).To(Equal(params))
				}

				When("provided valid configuration parameters", func() {
					It("sends the parameters to the service broker", func() {
						session := helpers.CF("create-service-key", serviceInstanceName, serviceKeyName, "-c", `{"wheres":"waldo"}`)
						Eventually(session).Should(Exit(0))

						checkParameters(map[string]interface{}{"wheres": "waldo"})
					})
				})

				When("provided invalid configuration parameters", func() {
					It("displays an invalid configuration error", func() {
						session := helpers.CF("create-service-key", serviceInstanceName, serviceKeyName, "-c", `{"bad json"}`)
						Eventually(session).Should(Exit(1))
						Expect(session.Err).To(Say("Invalid configuration provided for -c flag. Please provide a valid JSON object or path to a file containing a valid JSON object."))
					})
				})

				When("configuration parameters are provided in a file", func() {
					When("the file contains valid JSON", func() {
						const parametersJSON = `{"valid":"json"}`
						var tempFilePath string

						BeforeEach(func() {
							tempFilePath = helpers.TempFileWithContent(parametersJSON)
						})

						AfterEach(func() {
							Expect(os.Remove(tempFilePath)).To(Succeed())
						})

						It("sends the parameters to the service broker", func() {
							session := helpers.CF("create-service-key", serviceInstanceName, serviceKeyName, "-c", tempFilePath)
							Eventually(session).Should(Exit(0))

							checkParameters(map[string]interface{}{"valid": "json"})
						})
					})

					When("the file-path does not exist", func() {
						It("displays an invalid configuration error, and exits 1", func() {
							session := helpers.CF("create-service-key", serviceInstanceName, serviceKeyName, "-c", `/this/is/not/valid`)
							Eventually(session).Should(Exit(1))
							Expect(session.Err).To(Say("Invalid configuration provided for -c flag. Please provide a valid JSON object or path to a file containing a valid JSON object."))
						})
					})

					When("the file contains invalid json", func() {
						var tempFilePath string

						BeforeEach(func() {
							tempFilePath = helpers.TempFileWithContent("{i-am-very-bad-json")
						})

						AfterEach(func() {
							Expect(os.Remove(tempFilePath)).To(Succeed())
						})

						It("displays an invalid configuration error, and exits 1", func() {
							session := helpers.CF("create-service-key", serviceInstanceName, serviceKeyName, "-c", tempFilePath)
							Eventually(session).Should(Exit(1))
							Expect(session.Err).To(Say("Invalid configuration provided for -c flag. Please provide a valid JSON object or path to a file containing a valid JSON object."))
						})
					})
				})
			})

			Describe("asynchronous broker response", func() {
				BeforeEach(func() {
					broker.WithAsyncDelay(time.Second).Configure()
				})

				It("starts to create the service key", func() {
					session := helpers.CF("create-service-key", serviceInstanceName, serviceKeyName)
					Eventually(session).Should(Exit(0))
					Expect(session.Out).To(SatisfyAll(
						Say(`Creating service key %s for service instance %s as %s\.\.\.\n`, serviceKeyName, serviceInstanceName, username),
						Say(`OK\n`),
						Say(`Create in progress\.\n`),
					))

					Expect(string(session.Err.Contents())).To(BeEmpty())
				})

				When("--wait flag specified", func() {
					It("waits for completion", func() {
						session := helpers.CF("create-service-key", serviceInstanceName, serviceKeyName, "--wait")
						Eventually(session).Should(Exit(0))
						Expect(session.Out).To(SatisfyAll(
							Say(`Creating service key %s for service instance %s as %s\.\.\.\n`, serviceKeyName, serviceInstanceName, username),
							Say(`Waiting for the operation to complete\.+\n`),
							Say(`\n`),
							Say(`OK\n`),
						))

						Expect(string(session.Err.Contents())).To(BeEmpty())
					})
				})
			})
		})
	})
})
