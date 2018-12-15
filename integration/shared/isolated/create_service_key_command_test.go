// +build !partialPush

package isolated

import (
	"io/ioutil"
	"os"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("create-service-key command", func() {
	Describe("help", func() {
		When("--help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("create-service-key", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say(`\s+create-service-key - Create key for a service instance`))

				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say(`\s+cf create-service-key SERVICE_INSTANCE SERVICE_KEY \[-c PARAMETERS_AS_JSON\]`))
				Eventually(session).Should(Say(`\s+Optionally provide service-specific configuration parameters in a valid JSON object in-line.`))
				Eventually(session).Should(Say(`\s+cf create-service-key SERVICE_INSTANCE SERVICE_KEY -c '{\"name\":\"value\",\"name\":\"value\"}'`))
				Eventually(session).Should(Say(`\s+Optionally provide a file containing service-specific configuration parameters in a valid JSON object. The path to the parameters file can be an absolute or relative path to a file.`))
				Eventually(session).Should(Say(`\s+cf create-service-key SERVICE_INSTANCE SERVICE_KEY -c PATH_TO_FILE`))
				Eventually(session).Should(Say(`\s+Example of valid JSON object:`))
				Eventually(session).Should(Say(`\s+{`))
				Eventually(session).Should(Say(`\s+\"permissions\": \"read-only\"`))
				Eventually(session).Should(Say(`\s+}`))

				Eventually(session).Should(Say("EXAMPLES:"))
				Eventually(session).Should(Say(`\s+cf create-service-key mydb mykey -c '{\"permissions\":\"read-only\"}'`))
				Eventually(session).Should(Say(`\s+cf create-service-key mydb mykey -c ~/workspace/tmp/instance_config.json`))
				Eventually(session).Should(Say("ALIAS:"))
				Eventually(session).Should(Say(`\s+csk`))

				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say(`\s+-c      Valid JSON object containing service-specific configuration parameters, provided either in-line or in a file. For a list of supported configuration parameters, see documentation for the particular service offering.`))

				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say(`\s+service-key`))

				Eventually(session).Should(Exit(0))
			})
		})
	})

	var (
		serviceInstance string
		serviceKeyName  string
		service         string
		servicePlan     string
	)

	BeforeEach(func() {
		serviceInstance = helpers.PrefixedRandomName("si")
		serviceKeyName = "service-key"
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "create-service-key", "service-name", "key-name")
		})
	})

	When("the environment is setup correctly", func() {
		var (
			org      string
			space    string
			domain   string
			username string
		)

		BeforeEach(func() {
			org = helpers.NewOrgName()
			space = helpers.NewSpaceName()
			username, _ = helpers.GetCredentials()

			helpers.SetupCF(org, space)
			domain = helpers.DefaultSharedDomain()
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(org)
		})

		When("not providing any arguments", func() {
			It("displays an invalid usage error and the help text, and exits 1", func() {
				session := helpers.CF("create-service-key")
				Eventually(session.Err).Should(Say("Incorrect Usage: the required arguments `SERVICE_INSTANCE` and `SERVICE_KEY` were not provided"))

				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say(`\s+create-service-key - Create key for a service instance`))
				Eventually(session).Should(Exit(1))
			})
		})

		When("provided with service instance name but no service key name", func() {
			It("displays an invalid usage error and the help text, and exits 1", func() {
				session := helpers.CF("create-service-key", serviceInstance)
				Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `SERVICE_KEY` was not provided"))
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Exit(1))
			})
		})

		When("provided with a service instance name that doesn't exist", func() {
			It("displays FAILED and an informative error, and exits 1", func() {
				session := helpers.CF("create-service-key", "missing-service-instance", serviceKeyName)
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("Service instance missing-service-instance not found"))
				Eventually(session).Should(Exit(1))
			})
		})

		When("provided with a brokered service instance", func() {
			var broker helpers.ServiceBroker

			BeforeEach(func() {
				service = helpers.PrefixedRandomName("SERVICE")
				servicePlan = helpers.PrefixedRandomName("SERVICE-PLAN")
				broker = helpers.NewServiceBroker(helpers.NewServiceBrokerName(), helpers.NewAssets().ServiceBroker, domain, service, servicePlan)
				broker.Push()
				broker.Configure(true)
				broker.Create()

				Eventually(helpers.CF("enable-service-access", service)).Should(Exit(0))
				Eventually(helpers.CF("create-service", service, servicePlan, serviceInstance)).Should(Exit(0))
			})

			AfterEach(func() {
				Eventually(helpers.CF("delete-service-key", serviceInstance, serviceKeyName)).Should(Exit(0))
				broker.Destroy()
			})

			It("creates the service key and displays OK", func() {
				session := helpers.CF("create-service-key", serviceInstance, serviceKeyName)
				Eventually(session).Should(Say("Creating service key %s for service instance %s as %s...", serviceKeyName, serviceInstance, username))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))
			})

			When("provided valid configuration parameters", func() {
				It("creates the service key and displays OK", func() {
					session := helpers.CF("create-service-key", serviceInstance, serviceKeyName, "-c", `{"wheres":"waldo"}`)
					Eventually(session).Should(Say("Creating service key %s for service instance %s as %s...", serviceKeyName, serviceInstance, username))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))
				})
			})

			When("provided invalid configuration parameters", func() {
				It("displays an invalid configuration error", func() {
					session := helpers.CF("create-service-key", serviceInstance, serviceKeyName, "-c", `{"bad json"}`)
					Eventually(session.Err).Should(Say("Invalid configuration provided for -c flag. Please provide a valid JSON object or path to a file containing a valid JSON object."))
					Eventually(session).Should(Exit(1))
				})
			})

			When("configuration parameters are provided in a file", func() {

				When("the file-path does not exist", func() {
					It("displays an invalid configuration error, and exits 1", func() {
						session := helpers.CF("create-service-key", serviceInstance, serviceKeyName, "-c", `/this/is/not/valid`)
						Eventually(session.Err).Should(Say("Invalid configuration provided for -c flag. Please provide a valid JSON object or path to a file containing a valid JSON object."))
						Eventually(session).Should(Exit(1))
					})
				})

				When("the file contains invalid json", func() {
					var configurationFile *os.File

					BeforeEach(func() {
						var err error
						content := []byte("{i-am-very-bad-json")
						configurationFile, err = ioutil.TempFile("", "CF_CLI")
						Expect(err).NotTo(HaveOccurred())

						_, err = configurationFile.Write(content)
						Expect(err).NotTo(HaveOccurred())

						Expect(configurationFile.Close()).To(Succeed())
					})

					AfterEach(func() {
						Expect(os.RemoveAll(configurationFile.Name())).To(Succeed())
					})

					It("displays an invalid configuration error, and exits 1", func() {
						session := helpers.CF("create-service-key", serviceInstance, serviceKeyName, "-c", configurationFile.Name())
						Eventually(session.Err).Should(Say("Invalid configuration provided for -c flag. Please provide a valid JSON object or path to a file containing a valid JSON object."))
						Eventually(session).Should(Exit(1))
					})
				})
			})

			When("a service key with the provided name already exists", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("create-service-key", serviceInstance, serviceKeyName)).Should(Exit(0))
				})

				It("displays OK and an informative error message", func() {
					session := helpers.CF("create-service-key", serviceInstance, serviceKeyName)
					Eventually(session).Should(Say("OK"))
					Eventually(session.Err).Should(Say("Service key %s already exists", serviceKeyName))
					Eventually(session).Should(Exit(0))
				})
			})

			When("the service is not bindable", func() {
				BeforeEach(func() {
					broker.Service.Bindable = false
					broker.Configure(true)
					broker.Update()
				})

				It("displays FAILED and an informative error, and exits 1", func() {
					session := helpers.CF("create-service-key", serviceInstance, serviceKeyName)
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("This service doesn't support creation of keys."))
					Eventually(session).Should(Exit(1))
				})
			})
		})

		When("provided with a user-provided service instance", func() {
			BeforeEach(func() {
				Eventually(helpers.CF("create-user-provided-service", serviceInstance)).Should(Exit(0))
			})

			It("Displays an informative error message and FAILED, and exits 1", func() {
				session := helpers.CF("create-service-key", serviceInstance, serviceKeyName)
				Eventually(session.Err).Should(Say("Service keys are not supported for user-provided service instances."))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})
	})
})
