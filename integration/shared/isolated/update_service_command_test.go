package isolated

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/integration/assets/hydrabroker/config"
	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/integration/helpers/servicebrokerstub"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("update-service command", func() {
	Describe("help", func() {
		When("--help flag is set", func() {
			It("displays command usage to output", func() {
				session := helpers.CF("update-service", "--help")

				Eventually(session).Should(Exit(0))

				Expect(session).Should(Say("NAME:"))
				Expect(session).Should(Say(`\s+update-service - Update a service instance`))
				Expect(session).Should(Say(`USAGE:`))
				Expect(session).Should(Say(`\s+cf update-service SERVICE_INSTANCE \[-p NEW_PLAN\] \[-c PARAMETERS_AS_JSON\] \[-t TAGS\] \[--upgrade\]`))
				Expect(session).Should(Say(`\s+Optionally provide service-specific configuration parameters in a valid JSON object in-line:`))
				Expect(session).Should(Say(`\s+cf update-service SERVICE_INSTANCE -c '{\"name\":\"value\",\"name\":\"value\"}'`))
				Expect(session).Should(Say(`\s+Optionally provide a file containing service-specific configuration parameters in a valid JSON object\.`))
				Expect(session).Should(Say(`\s+The path to the parameters file can be an absolute or relative path to a file:`))
				Expect(session).Should(Say(`\s+cf update-service SERVICE_INSTANCE -c PATH_TO_FILE`))
				Expect(session).Should(Say(`\s+Example of valid JSON object:`))
				Expect(session).Should(Say(`\s+{`))
				Expect(session).Should(Say(`\s+\"cluster_nodes\": {`))
				Expect(session).Should(Say(`\s+\"count\": 5,`))
				Expect(session).Should(Say(`\s+\"memory_mb\": 1024`))
				Expect(session).Should(Say(`\s+}`))
				Expect(session).Should(Say(`\s+}`))
				Expect(session).Should(Say(`\s+ Optionally provide a list of comma-delimited tags that will be written to the VCAP_SERVICES environment variable for any bound applications.`))
				Expect(session).Should(Say(`EXAMPLES:`))
				Expect(session).Should(Say(`\s+cf update-service mydb -p gold`))
				Expect(session).Should(Say(`\s+cf update-service mydb -c '{\"ram_gb\":4}'`))
				Expect(session).Should(Say(`\s+cf update-service mydb -c ~/workspace/tmp/instance_config.json`))
				Expect(session).Should(Say(`\s+cf update-service mydb -t "list, of, tags"`))
				Expect(session).Should(Say(`\s+cf update-service mydb --upgrade`))
				Expect(session).Should(Say(`\s+cf update-service mydb --upgrade --force`))
				Expect(session).Should(Say(`OPTIONS:`))
				Expect(session).Should(Say(`\s+-c\s+Valid JSON object containing service-specific configuration parameters, provided either in-line or in a file\. For a list of supported configuration parameters, see documentation for the particular service offering\.`))
				Expect(session).Should(Say(`\s+-p\s+Change service plan for a service instance`))
				Expect(session).Should(Say(`\s+-t\s+User provided tags`))
				Expect(session).Should(Say(`\s+--upgrade, -u\s+Upgrade the service instance to the latest version of the service plan available. It cannot be combined with flags: -c, -p, -t.`))
				Expect(session).Should(Say(`\s+--force, -f\s+Force the upgrade to the latest available version of the service plan. It can only be used with: -u, --upgrade.`))
				Expect(session).Should(Say(`SEE ALSO:`))
				Expect(session).Should(Say(`\s+rename-service, services, update-user-provided-service`))
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

					Eventually(session).Should(Exit(1))

					Expect(session.Err).Should(Say("Service instance non-existent-service not found"))
				})
			})
		})

		When("providing other arguments while upgrading", func() {
			It("displays an informative error message and exits 1", func() {
				session := helpers.CF("update-service", "irrelevant", "--upgrade", "-c", "{\"hello\": \"world\"}")

				Eventually(session).Should(Exit(1))
				Expect(session.Err).Should(Say("Incorrect Usage: The following arguments cannot be used together: --upgrade, -t, -c, -p"))
				Expect(session).Should(Say("FAILED"))
				Expect(session).Should(Say("USAGE:"))
				Expect(session).Should(Exit(1))
			})
		})

		When("there is a service instance", func() {
			var (
				broker              *servicebrokerstub.ServiceBrokerStub
				serviceInstanceName string
				tags                string
				username            string
			)

			BeforeEach(func() {
				broker = servicebrokerstub.EnableServiceAccess()

				serviceInstanceName = helpers.PrefixedRandomName("SI")
				tags = "a tag"
				Eventually(helpers.CF("create-service", broker.FirstServiceOfferingName(), broker.FirstServicePlanName(), serviceInstanceName, "-t", tags)).Should(Exit(0))

				username, _ = helpers.GetCredentials()
			})

			AfterEach(func() {
				Eventually(helpers.CF("delete-service", serviceInstanceName, "-f")).Should(Exit(0))
				broker.Forget()
			})

			When("updating to a service plan that does not exist", func() {
				It("displays an informative error message, exits 1", func() {
					session := helpers.CF("update-service", serviceInstanceName, "-p", "non-existing-service-plan")
					Eventually(session).Should(Exit(1))

					Expect(session).Should(Say("Plan does not exist for the %s service", broker.FirstServiceOfferingName()))
				})
			})

			When("updating to the same service plan (no-op)", func() {
				It("displays an informative success message, exits 0", func() {
					session := helpers.CF("update-service", serviceInstanceName, "-p", broker.FirstServicePlanName())

					Eventually(session).Should(Exit(0))
					Expect(session).Should(Say("Updating service instance %s as %s...", serviceInstanceName, username))

					Expect(session).Should(Say("OK"))
					Expect(session).Should(Say("No changes were made"))
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

							Eventually(session).Should(Exit(0))

							Expect(session).Should(Say("You are about to update %s", serviceInstanceName))
							Expect(session).Should(Say("Warning: This operation may be long running and will block further operations on the service until complete."))
							Expect(session).Should(Say("Really update service %s\\? \\[yN\\]:", serviceInstanceName))
							Expect(session).Should(Say("Update cancelled"))
							Expect(session).Should(Exit(0))
						})
					})

					When("proceeding with the update", func() {
						BeforeEach(func() {
							_, err := buffer.Write([]byte("y\n"))
							Expect(err).ToNot(HaveOccurred())
						})

						When("updating a parameter", func() {
							It("updates the service without removing the tags", func() {
								session := helpers.CFWithStdin(buffer, "update-service", serviceInstanceName, "-c", "{\"tls\": true}")

								Eventually(session).Should(Exit(0))

								Expect(session).Should(Say("Updating service instance %s as %s...", serviceInstanceName, username))

								session = helpers.CF("service", serviceInstanceName)

								Eventually(session).Should(Exit(0))
								Expect(session).Should(Say("a tag"))
							})
						})

						When("upgrade is available", func() {
							BeforeEach(func() {
								broker.Services[0].Plans[0].MaintenanceInfo = &config.MaintenanceInfo{Version: "9.1.2"}
								broker.Configure().Register()
							})

							It("updates the service", func() {

								session := helpers.CFWithStdin(buffer, "update-service", serviceInstanceName, "--upgrade")
								By("displaying an informative message")
								Eventually(session).Should(Exit(0))

								Expect(session).Should(Say("You are about to update %s", serviceInstanceName))
								Expect(session).Should(Say("Warning: This operation may be long running and will block further operations on the service until complete."))
								Expect(session).Should(Say("Really update service %s\\? \\[yN\\]:", serviceInstanceName))
								Expect(session).Should(Say("Updating service instance %s as %s...", serviceInstanceName, username))

								By("requesting an upgrade from the platform")
								session = helpers.CF("service", serviceInstanceName)
								Eventually(session).Should(Say("status:\\s+update succeeded"))
							})
						})

						When("no upgrade is available", func() {
							It("does not update the service and outputs informative message", func() {
								session := helpers.CFWithStdin(buffer, "update-service", serviceInstanceName, "--upgrade")

								Eventually(session).Should(Exit(1))

								Expect(session).Should(Say("You are about to update %s", serviceInstanceName))
								Expect(session).Should(Say("Warning: This operation may be long running and will block further operations on the service until complete."))
								Expect(session).Should(Say("Really update service %s\\? \\[yN\\]:", serviceInstanceName))
								Expect(session).Should(Say("Updating service instance %s as %s...", serviceInstanceName, username))
								Expect(session.Err).Should(Say("No upgrade is available."))
								Expect(session.Err).Should(Say("TIP: To find out if upgrade is available run `cf service %s`.", serviceInstanceName))
							})
						})
					})

					When("providing --force argument and upgrade is available", func() {
						BeforeEach(func() {
							broker.Services[0].Plans[0].MaintenanceInfo = &config.MaintenanceInfo{Version: "9.1.2"}
							broker.Configure().Register()
						})

						It("updates the service without prompting", func() {
							session := helpers.CFWithStdin(buffer, "update-service", serviceInstanceName, "--upgrade", "--force")

							By("displaying an informative message")
							Eventually(session).Should(Exit(0))
							Expect(session).Should(Say("Updating service instance %s as %s...", serviceInstanceName, username))

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
