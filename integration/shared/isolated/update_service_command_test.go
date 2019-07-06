package isolated

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/integration/helpers/fakeservicebroker"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("update-service command", func() {
	BeforeEach(func() {
		helpers.SkipIfClientCredentialsTestMode()
	})

	Describe("help", func() {
		When("--help flag is set", func() {
			It("displays command usage to output", func() {
				session := helpers.CF("update-service", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say(`\s+update-service - Update a service instance`))
				Eventually(session).Should(Say(`USAGE:`))
				Eventually(session).Should(Say(`\s+cf update-service SERVICE_INSTANCE \[-p NEW_PLAN\] \[-c PARAMETERS_AS_JSON\] \[-t TAGS\] \[--upgrade\]`))
				Eventually(session).Should(Say(`\s+Optionally provide service-specific configuration parameters in a valid JSON object in-line:`))
				Eventually(session).Should(Say(`\s+cf update-service SERVICE_INSTANCE -c '{\"name\":\"value\",\"name\":\"value\"}'`))
				Eventually(session).Should(Say(`\s+Optionally provide a file containing service-specific configuration parameters in a valid JSON object\.`))
				Eventually(session).Should(Say(`\s+The path to the parameters file can be an absolute or relative path to a file:`))
				Eventually(session).Should(Say(`\s+cf update-service SERVICE_INSTANCE -c PATH_TO_FILE`))
				Eventually(session).Should(Say(`\s+Example of valid JSON object:`))
				Eventually(session).Should(Say(`\s+{`))
				Eventually(session).Should(Say(`\s+\"cluster_nodes\": {`))
				Eventually(session).Should(Say(`\s+\"count\": 5,`))
				Eventually(session).Should(Say(`\s+\"memory_mb\": 1024`))
				Eventually(session).Should(Say(`\s+}`))
				Eventually(session).Should(Say(`\s+}`))
				Eventually(session).Should(Say(`\s+ Optionally provide a list of comma-delimited tags that will be written to the VCAP_SERVICES environment variable for any bound applications.`))
				Eventually(session).Should(Say(`EXAMPLES:`))
				Eventually(session).Should(Say(`\s+cf update-service mydb -p gold`))
				Eventually(session).Should(Say(`\s+cf update-service mydb -c '{\"ram_gb\":4}'`))
				Eventually(session).Should(Say(`\s+cf update-service mydb -c ~/workspace/tmp/instance_config.json`))
				Eventually(session).Should(Say(`\s+cf update-service mydb -t "list, of, tags"`))
				Eventually(session).Should(Say(`\s+cf update-service mydb --upgrade`))
				Eventually(session).Should(Say(`OPTIONS:`))
				Eventually(session).Should(Say(`\s+-c\s+Valid JSON object containing service-specific configuration parameters, provided either in-line or in a file\. For a list of supported configuration parameters, see documentation for the particular service offering\.`))
				Eventually(session).Should(Say(`\s+-p\s+Change service plan for a service instance`))
				Eventually(session).Should(Say(`\s+-t\s+User provided tags`))
				Eventually(session).Should(Say(`\s+-u\s+Upgrade the service instance to the latest version of the service plan available. This flag is in EXPERIMENTAL stage and may change without notice. It cannot be combined with other flags.`))
				Eventually(session).Should(Say(`SEE ALSO:`))
				Eventually(session).Should(Say(`\s+rename-service, services, update-user-provided-service`))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the environment is not setup correctly", func() {
		BeforeEach(func() {
			helpers.SkipIfVersionLessThan(ccversion.MinVersionUpdateServiceInstanceMaintenanceInfoV2)
		})

		It("fails with the appropriate errors", func() {
			// the upgrade flag is passed here to exercise a particular code path before refactoring
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "update-service", "foo", "--upgrade")
		})
	})

	When("an api is targeted, the user is logged in, and an org and space are targeted", func() {
		var (
			orgName string
		)

		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			var spaceName = helpers.NewSpaceName()
			helpers.SetupCF(orgName, spaceName)
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		When("there are no service instances", func() {
			When("upgrading", func() {
				BeforeEach(func() {
					helpers.SkipIfVersionLessThan(ccversion.MinVersionUpdateServiceInstanceMaintenanceInfoV2)
				})

				It("displays an informative error before prompting and exits 1", func() {
					session := helpers.CF("update-service", "non-existent-service", "--upgrade")
					Eventually(session.Err).Should(Say("Service instance non-existent-service not found"))
					Eventually(session).Should(Exit(1))
				})
			})
		})

		When("providing other arguments while upgrading", func() {
			It("displays an informative error message and exits 1", func() {
				session := helpers.CF("update-service", "irrelevant", "--upgrade", "-c", "{\"hello\": \"world\"}")
				Eventually(session.Err).Should(Say("Incorrect Usage: The following arguments cannot be used together: --upgrade, -t, -c, -p"))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Exit(1))
			})
		})

		When("there is a service instance", func() {
			var (
				broker              *fakeservicebroker.FakeServiceBroker
				serviceInstanceName string
				username            string
			)

			BeforeEach(func() {
				broker = fakeservicebroker.New().Register()
				Eventually(helpers.CF("enable-service-access", broker.ServiceName())).Should(Exit(0))

				serviceInstanceName = helpers.PrefixedRandomName("SI")
				Eventually(helpers.CF("create-service", broker.ServiceName(), broker.ServicePlanName(), serviceInstanceName)).Should(Exit(0))

				username, _ = helpers.GetCredentials()
			})

			AfterEach(func() {
				Eventually(helpers.CF("delete-service", serviceInstanceName, "-f")).Should(Exit(0))
				broker.Destroy()
			})

			When("updating to a service plan that does not exist", func() {
				It("displays an informative error message, exits 1", func() {
					session := helpers.CF("update-service", serviceInstanceName, "-p", "non-existing-service-plan")
					Eventually(session).Should(Say("Plan does not exist for the %s service", broker.ServiceName()))
					Eventually(session).Should(Exit(1))
				})
			})

			When("updating to the same service plan (no-op)", func() {
				It("displays an informative success message, exits 0", func() {
					session := helpers.CF("update-service", serviceInstanceName, "-p", broker.ServicePlanName())
					Eventually(session).Should(Say("Updating service instance %s as %s...", serviceInstanceName, username))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Say("No changes were made"))
					Eventually(session).Should(Exit(0))
				})
			})

			When("upgrading", func() {
				var buffer *Buffer

				BeforeEach(func() {
					buffer = NewBuffer()
				})

				When("the user provides --upgrade in an unsupported CAPI version", func() {
					BeforeEach(func() {
						helpers.SkipIfVersionAtLeast(ccversion.MinVersionUpdateServiceInstanceMaintenanceInfoV2)
					})

					It("should report that the version of CAPI is too low", func() {
						session := helpers.CF("update-service", serviceInstanceName, "--upgrade")
						Eventually(session.Err).Should(Say(`Option '--upgrade' requires CF API version %s or higher. Your target is 2\.\d+\.\d+`, ccversion.MinVersionUpdateServiceInstanceMaintenanceInfoV2))
						Eventually(session).Should(Exit(1))
					})
				})

				When("when CAPI supports service instance maintenance_info updates", func() {
					BeforeEach(func() {
						helpers.SkipIfVersionLessThan(ccversion.MinVersionUpdateServiceInstanceMaintenanceInfoV2)
					})

					When("cancelling the update", func() {
						BeforeEach(func() {
							_, err := buffer.Write([]byte("n\n"))
							Expect(err).ToNot(HaveOccurred())
						})

						It("does not proceed", func() {
							session := helpers.CFWithStdin(buffer, "update-service", serviceInstanceName, "--upgrade")
							Eventually(session).Should(Say("You are about to update %s", serviceInstanceName))
							Eventually(session).Should(Say("Warning: This operation may be long running and will block further operations on the service until complete."))
							Eventually(session).Should(Say("Really update service %s\\? \\[yN\\]:", serviceInstanceName))
							Eventually(session).Should(Say("Update cancelled"))
							Eventually(session).Should(Exit(0))
						})
					})

					When("proceeding with the update", func() {
						BeforeEach(func() {
							_, err := buffer.Write([]byte("y\n"))
							Expect(err).ToNot(HaveOccurred())
						})

						It("updates the service", func() {
							session := helpers.CFWithStdin(buffer, "update-service", serviceInstanceName, "--upgrade")

							By("displaying an informative message")
							Eventually(session).Should(Say("You are about to update %s", serviceInstanceName))
							Eventually(session).Should(Say("Warning: This operation may be long running and will block further operations on the service until complete."))
							Eventually(session).Should(Say("Really update service %s\\? \\[yN\\]:", serviceInstanceName))
							Eventually(session).Should(Exit(0))

							By("requesting an upgrade from the platform")
							session = helpers.CF("service", serviceInstanceName)
							Eventually(session).Should(Say("status:\\s+update succeeded"))
						})
					})
				})
			})
		})
	})
})
