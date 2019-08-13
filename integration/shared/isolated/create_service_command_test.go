package isolated

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/integration/helpers/fakeservicebroker"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("create-service command", func() {
	Describe("help", func() {
		When("--help flag is set", func() {
			It("displays command usage to output", func() {
				session := helpers.CF("create-service", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say(`\s+create-service - Create a service instance`))
				Eventually(session).Should(Say(`USAGE:`))
				Eventually(session).Should(Say(`\s+cf create-service SERVICE PLAN SERVICE_INSTANCE \[-b BROKER\] \[-c PARAMETERS_AS_JSON\] \[-t TAGS\]`))
				Eventually(session).Should(Say(`\s+Optionally provide service-specific configuration parameters in a valid JSON object in-line:`))
				Eventually(session).Should(Say(`\s+cf create-service SERVICE PLAN SERVICE_INSTANCE -c '{\"name\":\"value\",\"name\":\"value\"}'`))
				Eventually(session).Should(Say(`\s+Optionally provide a file containing service-specific configuration parameters in a valid JSON object\.`))
				Eventually(session).Should(Say(`\s+The path to the parameters file can be an absolute or relative path to a file:`))
				Eventually(session).Should(Say(`\s+cf create-service SERVICE PLAN SERVICE_INSTANCE -c PATH_TO_FILE`))
				Eventually(session).Should(Say(`\s+Example of valid JSON object:`))
				Eventually(session).Should(Say(`\s+{`))
				Eventually(session).Should(Say(`\s+\"cluster_nodes\": {`))
				Eventually(session).Should(Say(`\s+\"count\": 5,`))
				Eventually(session).Should(Say(`\s+\"memory_mb\": 1024`))
				Eventually(session).Should(Say(`\s+}`))
				Eventually(session).Should(Say(`\s+}`))
				Eventually(session).Should(Say(`TIP:`))
				Eventually(session).Should(Say(`\s+Use 'cf create-user-provided-service' to make user-provided services available to CF apps`))
				Eventually(session).Should(Say(`EXAMPLES:`))
				Eventually(session).Should(Say(`\s+Linux/Mac:`))
				Eventually(session).Should(Say(`\s+cf create-service db-service silver mydb -c '{\"ram_gb\":4}'`))
				Eventually(session).Should(Say(`\s+Windows Command Line:`))
				Eventually(session).Should(Say(`\s+cf create-service db-service silver mydb -c \"{\\\"ram_gb\\\":4}\"`))
				Eventually(session).Should(Say(`\s+Windows PowerShell:`))
				Eventually(session).Should(Say(`\s+cf create-service db-service silver mydb -c '{\\\"ram_gb\\\":4}'`))
				Eventually(session).Should(Say(`\s+cf create-service db-service silver mydb -c ~/workspace/tmp/instance_config.json`))
				Eventually(session).Should(Say(`\s+cf create-service db-service silver mydb -t \"list, of, tags\"`))
				Eventually(session).Should(Say(`ALIAS:`))
				Eventually(session).Should(Say(`\s+cs`))
				Eventually(session).Should(Say(`OPTIONS:`))
				Eventually(session).Should(Say(`\s+-b      Create a service instance from a particular broker\. Required when service name is ambiguous`))
				Eventually(session).Should(Say(`\s+-c      Valid JSON object containing service-specific configuration parameters, provided either in-line or in a file\. For a list of supported configuration parameters, see documentation for the particular service offering\.`))
				Eventually(session).Should(Say(`\s+-t      User provided tags`))
				Eventually(session).Should(Say(`SEE ALSO:`))
				Eventually(session).Should(Say(`\s+bind-service, create-user-provided-service, marketplace, services`))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("not logged in", func() {
		BeforeEach(func() {
			helpers.LogoutCF()
		})

		It("displays FAILED, an informative error message, and exits 1", func() {
			session := helpers.CF("create-service", "service", "plan", "my-service")
			Eventually(session).Should(Say("FAILED"))
			Eventually(session.Err).Should(Say("Not logged in. Use 'cf login' to log in\\."))
			Eventually(session).Should(Exit(1))
		})
	})

	When("logged in", func() {
		BeforeEach(func() {
			helpers.LoginCF()
		})

		When("the environment is not setup correctly", func() {
			It("fails with the appropriate errors", func() {
				helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "create-service", "service-name", "simple", "new-service")
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

				helpers.SetupCF(org, space)

				username, _ = helpers.GetCredentials()
			})

			AfterEach(func() {
				helpers.QuickDeleteOrg(org)
			})

			When("not providing any arguments", func() {
				It("displays an invalid usage error and the help text, and exits 1", func() {
					session := helpers.CF("create-service")
					Eventually(session.Err).Should(Say("Incorrect Usage: the required arguments `SERVICE`, `SERVICE_PLAN` and `SERVICE_INSTANCE` were not provided"))

					// checking partial help text, too long and it's tested earlier
					Eventually(session).Should(Say("NAME:"))
					Eventually(session).Should(Say(`\s+create-service - Create a service instance`))
					Eventually(session).Should(Exit(1))
				})
			})

			When("invalid arguments are passed", func() {
				When("with an invalid json for -c", func() {
					It("displays an informative error message, exits 1", func() {
						session := helpers.CF("create-service", "foo", "bar", "my-service", "-c", "{")
						Eventually(session.Err).Should(Say("Invalid configuration provided for -c flag. Please provide a valid JSON object or path to a file containing a valid JSON object\\."))
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
						session := helpers.CF("create-service", "foo", "bar", "my-service", "-c", tempFilePath)
						Eventually(session.Err).Should(Say("Invalid configuration provided for -c flag. Please provide a valid JSON object or path to a file containing a valid JSON object\\."))
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
						session := helpers.CF("create-service", "foo", "bar", "my-service", "-c", filepath.Join(emptyDir, "non-existent-file"))
						Eventually(session.Err).Should(Say("Invalid configuration provided for -c flag. Please provide a valid JSON object or path to a file containing a valid JSON object\\."))
						Eventually(session).Should(Exit(1))
					})
				})
			})

			When("the service provided is not accessible", func() {
				It("displays an informative message, exits 1", func() {
					session := helpers.CF("create-service", "some-service", "some-plan", "my-service")
					Eventually(session).Should(Say("Creating service instance %s in org %s / space %s as %s\\.\\.\\.",
						"my-service", org, space, username))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("Service offering 'some-service' not found"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("the service provided is accessible", func() {
				var (
					service     string
					servicePlan string
					broker      *fakeservicebroker.FakeServiceBroker
				)

				BeforeEach(func() {
					broker = fakeservicebroker.New().Register()
					service = broker.ServiceName()
					servicePlan = broker.ServicePlanName()

					Eventually(helpers.CF("enable-service-access", service)).Should(Exit(0))
				})

				AfterEach(func() {
					broker.Destroy()
				})

				It("displays an informative success message, exits 0", func() {
					By("creating the service")
					session := helpers.CF("create-service", service, servicePlan, "my-service")
					Eventually(session).Should(Say("Creating service instance %s in org %s / space %s as %s\\.\\.\\.",
						"my-service", org, space, username))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))

					session = helpers.CF("services")
					Eventually(session).Should(Exit(0))
					Eventually(session).Should(Say("%s\\s+%s\\s+%s\\s+create succeeded",
						"my-service",
						service,
						servicePlan,
					))

					By("displaying the service already exists when using a duplicate name")
					session = helpers.CF("create-service", service, servicePlan, "my-service")
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Say("Service my-service already exists"))
					Eventually(session).Should(Exit(0))
				})

				When("the provided plan does not exist", func() {
					It("displays an informative error message, exits 1", func() {
						session := helpers.CF("create-service", service, "some-plan", "service-instance")
						Eventually(session).Should(Say("Creating service instance %s in org %s / space %s as %s\\.\\.\\.",
							"service-instance", org, space, username))
						Eventually(session).Should(Say("FAILED"))
						Eventually(session.Err).Should(Say("The plan %s could not be found for service %s", "some-plan", service))
						Eventually(session).Should(Exit(1))
					})
				})

				When("creating with valid params json", func() {
					It("displays an informative success message, exits 0", func() {
						session := helpers.CF("create-service", service, servicePlan, "my-service", "-c", "{}")
						Eventually(session).Should(Say("Creating service instance %s in org %s / space %s as %s\\.\\.\\.",
							"my-service", org, space, username))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Exit(0))

						session = helpers.CF("services")
						Eventually(session).Should(Exit(0))
						Eventually(session).Should(Say("%s\\s+%s\\s+%s\\s+create succeeded",
							"my-service",
							service,
							servicePlan,
						))
					})
				})

				When("creating with valid params json in a file", func() {
					var tempFilePath string

					BeforeEach(func() {
						tempFilePath = helpers.TempFileWithContent(`{"valid":"json"}`)
					})

					AfterEach(func() {
						Expect(os.Remove(tempFilePath)).To(Succeed())
					})

					It("displays an informative success message, exits 0", func() {
						session := helpers.CF("create-service", service, servicePlan, "my-service", "-c", tempFilePath)
						Eventually(session).Should(Say("Creating service instance %s in org %s / space %s as %s\\.\\.\\.",
							"my-service", org, space, username))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Exit(0))

						session = helpers.CF("services")
						Eventually(session).Should(Exit(0))
						Eventually(session).Should(Say("%s\\s+%s\\s+%s\\s+create succeeded",
							"my-service",
							service,
							servicePlan,
						))
					})
				})

				When("creating with tags", func() {
					It("displays an informative message, exits 0, and creates the service with tags", func() {
						session := helpers.CF("create-service", service, servicePlan, "my-service", "-t", "sapi, rocks")
						Eventually(session).Should(Say("Creating service instance %s in org %s / space %s as %s\\.\\.\\.",
							"my-service", org, space, username))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Exit(0))

						session = helpers.CF("service", "my-service")
						Eventually(session).Should(Exit(0))
						Eventually(session).Should(Say("tags:\\s+sapi, rocks"))
					})
				})
			})

			When("the service provided is async and accessible", func() {
				var (
					service     string
					servicePlan string
					broker      *fakeservicebroker.FakeServiceBroker
				)

				BeforeEach(func() {
					broker = fakeservicebroker.New().Async().Register()
					service = broker.ServiceName()
					servicePlan = broker.ServicePlanName()
					Eventually(helpers.CF("enable-service-access", service)).Should(Exit(0))
				})

				AfterEach(func() {
					broker.Destroy()
				})

				It("creates the service and displays a message that creation is in progress", func() {
					session := helpers.CF("create-service", service, servicePlan, "my-service")
					Eventually(session).Should(Say("Creating service instance %s in org %s / space %s as %s\\.\\.\\.",
						"my-service", org, space, username))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Say("Create in progress. Use 'cf services' or 'cf service my-service' to check operation status."))
					Eventually(session).Should(Exit(0))
				})
			})

			When("there are two services with the same name from different brokers", func() {
				var (
					service     string
					servicePlan string
					broker1     *fakeservicebroker.FakeServiceBroker
					broker2     *fakeservicebroker.FakeServiceBroker
				)

				BeforeEach(func() {
					helpers.SkipIfVersionLessThan(ccversion.MinVersionMultiServiceRegistrationV2)

					broker1 = fakeservicebroker.New().Register()
					service = broker1.ServiceName()
					servicePlan = broker1.ServicePlanName()
					broker2 = fakeservicebroker.NewAlternate()
					broker2.Services[0].Name = service
					broker2.Services[0].Plans[0].Name = servicePlan
					broker2.Register()

					Eventually(helpers.CF("enable-service-access", service, "-b", broker1.Name())).Should(Exit(0))
					Eventually(helpers.CF("enable-service-access", service, "-b", broker2.Name())).Should(Exit(0))
				})

				AfterEach(func() {
					broker1.Destroy()
					broker2.Destroy()
				})

				When("the user does not specify which broker to use", func() {
					It("displays an informative error message, exits 1", func() {
						session := helpers.CF("create-service", service, servicePlan, "my-service")
						Eventually(session).Should(Say("Creating service instance %s in org %s / space %s as %s\\.\\.\\.",
							"my-service", org, space, username))
						Eventually(session.Err).Should(Say("Service '%s' is provided by multiple service brokers\\. Specify a broker by using the '-b' flag\\.", service))
						Eventually(session).Should(Say("FAILED"))
						Eventually(session).Should(Exit(1))
					})
				})

				When("the user specifies which broker to use", func() {
					When("the user is a space developer", func() {
						BeforeEach(func() {
							username = helpers.SwitchToSpaceRole(org, space, "SpaceDeveloper")
							helpers.TargetOrgAndSpace(org, space)
						})

						AfterEach(func() {
							helpers.SetupCF(org, space)
						})

						It("displays an informative success message, exits 0", func() {
							By("creating the service with -b flag")
							session := helpers.CF("create-service", service, servicePlan, "my-service", "-b", broker1.Name())
							Eventually(session).Should(Say("Creating service instance %s in org %s / space %s as %s\\.\\.\\.",
								"my-service", org, space, username))
							Eventually(session).Should(Say("OK"))
							Eventually(session).Should(Exit(0))

							session = helpers.CF("services")
							Eventually(session).Should(Exit(0))
							Eventually(session).Should(Say("%s\\s+%s\\s+%s\\s+create succeeded",
								"my-service",
								service,
								servicePlan,
							))
						})

						Context("the broker is not accessible by that user", func() {
							It("displays an informative error message, exits 1", func() {
								session := helpers.CF("create-service", service, servicePlan, "my-service", "-b", "non-existent-broker")
								Eventually(session).Should(Say("Creating service instance %s in org %s / space %s as %s\\.\\.\\.",
									"my-service", org, space, username))
								Eventually(session.Err).Should(Say("Service '%s' provided by service broker '%s' not found\\.", service, "non-existent-broker"))
								Eventually(session).Should(Say("FAILED"))
								Eventually(session).Should(Exit(1))
							})
						})
					})
				})
			})
		})
	})
})
