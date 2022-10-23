package isolated

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/integration/helpers/servicebrokerstub"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

const serviceCommand = "service"

var _ = Describe("create-service command", func() {
	Describe("help", func() {

		matchHelpMessage := SatisfyAll(
			Say(`NAME:\n`),
			Say(`\s+create-service - Create a service instance\n`),
			Say(`\n`),
			Say(`USAGE:\n`),
			Say(`\s+cf create-service SERVICE_OFFERING PLAN SERVICE_INSTANCE \[-b SERVICE_BROKER\] \[-c PARAMETERS_AS_JSON\] \[-t TAGS\]\n`),
			Say(`\s+Optionally provide service-specific configuration parameters in a valid JSON object in-line:\n`),
			Say(`\s+cf create-service SERVICE_OFFERING PLAN SERVICE_INSTANCE -c '{\"name\":\"value\",\"name\":\"value\"}'\n`),
			Say(`\s+Optionally provide a file containing service-specific configuration parameters in a valid JSON object\.\n`),
			Say(`\s+The path to the parameters file can be an absolute or relative path to a file:\n`),
			Say(`\s+cf create-service SERVICE_OFFERING PLAN SERVICE_INSTANCE -c PATH_TO_FILE\n`),
			Say(`\s+Example of valid JSON object:`),
			Say(`\s+{`),
			Say(`\s+\"cluster_nodes\": {`),
			Say(`\s+\"count\": 5,`),
			Say(`\s+\"memory_mb\": 1024`),
			Say(`\s+}`),
			Say(`\s+}`),
			Say(`TIP:`),
			Say(`\s+Use 'cf create-user-provided-service' to make user-provided service instances available to CF apps`),
			Say(`EXAMPLES:`),
			Say(`\s+Linux/Mac:\n`),
			Say(`\s+cf create-service db-service silver mydb -c '{\"ram_gb\":4}`),
			Say(`\s+Windows Command Line:`),
			Say(`\s+cf create-service db-service silver mydb -c \"{\\\"ram_gb\\\":4}\"`),
			Say(`\s+Windows PowerShell:`),
			Say(`\s+cf create-service db-service silver mydb -c '{\\\"ram_gb\\\":4}'`),
			Say(`\s+cf create-service db-service silver mydb -c ~/workspace/tmp/instance_config.json`),
			Say(`\s+cf create-service db-service silver mydb -t \"list, of, tags\"`),
			Say(`ALIAS:`),
			Say(`\s+cs`),
			Say(`OPTIONS:`),
			Say(`\s+-b\s+Create a service instance from a particular broker\. Required when service offering name is ambiguous`),
			Say(`\s+-c\s+Valid JSON object containing service-specific configuration parameters, provided either in-line or in a file\. For a list of supported configuration parameters, see documentation for the particular service offering\.`),
			Say(`\s+-t\s+User provided tags`),
			Say(`\s+--wait, -w\s+Wait for the operation to complete`),
			Say(`SEE ALSO:`),
			Say(`\s+bind-service, create-user-provided-service, marketplace, services`),
		)

		When("the -h flag is specified", func() {
			It("succeeds and prints help", func() {
				session := helpers.CF("create-service", "-h")
				Eventually(session).Should(Exit(0))
				Expect(session.Out).To(matchHelpMessage)
			})
		})

		When("the --help flag is specified", func() {
			It("succeeds and prints help", func() {
				session := helpers.CF("create-service", "--help")
				Eventually(session).Should(Exit(0))
				Expect(session.Out).To(matchHelpMessage)
			})
		})

		When("no arguments are provided", func() {
			It("displays a warning, the help text, and exits 1", func() {
				session := helpers.CF("create-service")
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Incorrect Usage: the required arguments `SERVICE_OFFERING`, `SERVICE_PLAN` and `SERVICE_INSTANCE` were not provided"))
				Expect(session.Out).To(matchHelpMessage)
			})
		})

		When("unknown flag is passed", func() {
			It("displays a warning, the help text, and exits 1", func() {
				session := helpers.CF("create-service", "-u")
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Incorrect Usage: unknown flag `u"))
				Expect(session.Out).To(matchHelpMessage)
			})
		})

		When("a flag is passed with no argument", func() {
			It("displays a warning, the help text, and exits 1", func() {
				session := helpers.CF("create-service", "-c")
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Incorrect Usage: expected argument for flag `-c'"))
				Expect(session.Out).To(matchHelpMessage)
			})
		})

	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "create-service", "foo", "foo", "foo")
		})
	})

	Context("targeting a space", func() {
		var (
			userName  string
			orgName   string
			spaceName string
		)
		const serviceInstanceName = "my-service"

		assertCreateMessage := func(session *Session) {
			Eventually(session).Should(Say("Creating service instance %s in org %s / space %s as %s...",
				serviceInstanceName, orgName, spaceName, userName))
		}

		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()
			helpers.SetupCF(orgName, spaceName)
			userName, _ = helpers.GetCredentials()
		})

		When("succeeds", func() {
			var (
				serviceOffering string
				servicePlan     string
				broker          *servicebrokerstub.ServiceBrokerStub
			)

			BeforeEach(func() {
				broker = servicebrokerstub.EnableServiceAccess()
				serviceOffering = broker.FirstServiceOfferingName()
				servicePlan = broker.FirstServicePlanName()
			})

			AfterEach(func() {
				helpers.QuickDeleteOrg(orgName)
				broker.Forget()
			})

			It("displays a message, OK and creates the instance", func() {
				session := helpers.CF("create-service", serviceOffering, servicePlan, serviceInstanceName, "-t", "a-tag,another-tag")
				assertCreateMessage(session)
				Eventually(session).Should(Say(`Service instance %s created\.\n`, serviceInstanceName))
				Eventually(session).Should(Exit(0))

				session = helpers.CF(serviceCommand, serviceInstanceName)
				Eventually(session).Should(Exit(0))
				Expect(session).To(SatisfyAll(
					Say(`name:\s+%s`, serviceInstanceName),
					Say(`type:\s+%s`, "managed"),
					Say(`tags:\s+a-tag,\s*another-tag`),
				))
			})

			When("creating with valid params json", func() {
				const parametersJSON = `{"valid":"json"}`

				It("displays an informative success message, and creates the instance with parameters", func() {
					session := helpers.CF("create-service", serviceOffering, servicePlan, serviceInstanceName, "-c", parametersJSON)
					Eventually(session).Should(Say("Creating service instance %s in org %s / space %s as %s\\.\\.\\.",
						serviceInstanceName, orgName, spaceName, userName))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))

					session = helpers.CF(serviceCommand, serviceInstanceName, "--params")
					Eventually(session).Should(Exit(0))
					Eventually(session).Should(SatisfyAll(
						Say(`\{\n`),
						Say(`  "valid": "json"\n`),
						Say(`\}\n`),
					))
				})
			})

			When("creating with valid params json in a file", func() {
				const parametersJSON = `{"valid":"json"}`
				var tempFilePath string

				BeforeEach(func() {
					tempFilePath = helpers.TempFileWithContent(parametersJSON)
				})

				AfterEach(func() {
					Expect(os.Remove(tempFilePath)).To(Succeed())
				})

				It("displays an informative success message, exits 0", func() {
					session := helpers.CF("create-service", serviceOffering, servicePlan, serviceInstanceName, "-c", tempFilePath)
					Eventually(session).Should(Say("Creating service instance %s in org %s / space %s as %s...",
						serviceInstanceName, orgName, spaceName, userName))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))

					session = helpers.CF(serviceCommand, serviceInstanceName, "--params")
					Eventually(session).Should(Exit(0))
					Eventually(session).Should(SatisfyAll(
						Say(`\{\n`),
						Say(`  "valid": "json"\n`),
						Say(`\}\n`),
					))
				})
			})

			When("the service broker responds asynchronously", func() {
				BeforeEach(func() {
					broker.WithAsyncDelay(time.Second).Configure()
				})

				It("displays a message, OK and creates the instance", func() {
					session := helpers.CF("create-service", serviceOffering, servicePlan, serviceInstanceName, "-t", "a-tag,another-tag")
					assertCreateMessage(session)
					Eventually(session).Should(Say("Create in progress. Use 'cf services' or 'cf service my-service' to check operation status."))
					Eventually(session).Should(Exit(0))

					session = helpers.CF(serviceCommand, serviceInstanceName)
					Eventually(session).Should(Exit(0))
					Expect(session).To(SatisfyAll(
						Say(`name:\s+%s`, serviceInstanceName),
						Say(`type:\s+%s`, "managed"),
						Say(`tags:\s+a-tag,\s*another-tag`),
					))

				})
			})

			When("creating with --wait flag", func() {
				BeforeEach(func() {
					broker.WithAsyncDelay(5 * time.Second).Configure()
				})

				It("displays a message, OK and creates the instance", func() {
					session := helpers.CF("create-service", serviceOffering, servicePlan, serviceInstanceName, "--wait")
					Eventually(session).Should(Exit(0))
					assertCreateMessage(session)
					Expect(session.Out).To(SatisfyAll(
						Say(`Service instance %s created\.\n`, serviceInstanceName),
						Say(`OK\n`),
					))

					session = helpers.CF(serviceCommand, serviceInstanceName)
					Eventually(session).Should(Exit(0))
					Expect(session).To(SatisfyAll(
						Say(`name:\s+%s`, serviceInstanceName),
						Say(`status:\s+create succeeded`),
					))
				})
			})
		})

		When("there are two offerings with the same name from different brokers", func() {
			var (
				serviceOffering string
				servicePlan     string
				broker1         *servicebrokerstub.ServiceBrokerStub
				broker2         *servicebrokerstub.ServiceBrokerStub
			)

			BeforeEach(func() {
				broker1 = servicebrokerstub.EnableServiceAccess()
				serviceOffering = broker1.FirstServiceOfferingName()
				servicePlan = broker1.FirstServicePlanName()
				broker2 = servicebrokerstub.New()
				broker2.Services[0].Name = serviceOffering
				broker2.Services[0].Plans[0].Name = servicePlan
				broker2.EnableServiceAccess()
			})

			AfterEach(func() {
				helpers.QuickDeleteOrg(orgName)
				broker1.Forget()
				broker2.Forget()
			})

			It("displays an error message prompting to disambiguate", func() {
				session := helpers.CF("create-service", serviceOffering, servicePlan, serviceInstanceName)
				assertCreateMessage(session)
				Eventually(session.Err).Should(Say("Service offering '%s' is provided by multiple service brokers. Specify a broker name by using the '-b' flag.", serviceOffering))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})

		When("there are no plans matching", func() {
			var (
				serviceOffering string
				broker1         *servicebrokerstub.ServiceBrokerStub
			)

			BeforeEach(func() {
				broker1 = servicebrokerstub.EnableServiceAccess()
				serviceOffering = broker1.FirstServiceOfferingName()
			})

			AfterEach(func() {
				helpers.QuickDeleteOrg(orgName)
				broker1.Forget()
			})
			It("displays an error message", func() {
				session := helpers.CF("create-service", serviceOffering, "another-service-plan", serviceInstanceName, "-b", broker1.Name)
				assertCreateMessage(session)
				Eventually(session.Err).Should(Say("The plan '%s' could not be found for service offering '%s' and broker '%s'.", "another-service-plan", serviceOffering, broker1.Name))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})

		When("invalid arguments are passed", func() {
			When("with an invalid json for -c", func() {
				It("displays an informative error message, exits 1", func() {
					session := helpers.CF("create-service", "foo", "bar", serviceInstanceName, "-c", "{")
					Eventually(session.Err).Should(Say("Invalid configuration provided for -c flag. Please provide a valid JSON object or path to a file containing a valid JSON object."))
					Eventually(session).Should(Exit(1))
				})
			})

			When("the provided file contains invalid json", func() {
				var tempFilePath string

				BeforeEach(func() {
					tempFilePath = helpers.TempFileWithContent(`{"invalid"}`)
				})

				AfterEach(func() {
					Expect(os.Remove(tempFilePath)).To(Succeed())
				})

				It("displays an informative message and exits 1", func() {
					session := helpers.CF("create-service", "foo", "bar", serviceInstanceName, "-c", tempFilePath)
					Eventually(session.Err).Should(Say("Invalid configuration provided for -c flag. Please provide a valid JSON object or path to a file containing a valid JSON object."))
					Eventually(session).Should(Exit(1))
				})
			})

			When("the provided file cannot be read", func() {
				var emptyDir string

				BeforeEach(func() {
					var err error
					emptyDir, err = ioutil.TempDir("", "")
					Expect(err).NotTo(HaveOccurred())
				})

				AfterEach(func() {
					Expect(os.RemoveAll(emptyDir)).To(Succeed())
				})

				It("displays an informative message and exits 1", func() {
					session := helpers.CF("create-service", "foo", "bar", serviceInstanceName, "-c", filepath.Join(emptyDir, "nonexistent-file"))
					Eventually(session.Err).Should(Say("Invalid configuration provided for -c flag. Please provide a valid JSON object or path to a file containing a valid JSON object."))
					Eventually(session).Should(Exit(1))
				})
			})
		})
	})
})
