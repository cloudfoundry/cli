package isolated

import (
	"os"
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/integration/assets/hydrabroker/config"
	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/integration/helpers/servicebrokerstub"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("update-service command (original)", func() {
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

var _ = Describe("update-service command", func() {
	const command = "v3-update-service"

	Describe("help", func() {
		matchHelpMessage := SatisfyAll(
			Say(`NAME:\n`),
			Say(`\s+%s - Update a service instance\n`, command),
			Say(`\n`),
			Say(`USAGE:\n`),
			Say(`\s+cf update-service SERVICE_INSTANCE \[-p NEW_PLAN\] \[-c PARAMETERS_AS_JSON\] \[-t TAGS\] \[--upgrade\]\n`),
			Say(`\n`),
			Say(`\s+Optionally provide service-specific configuration parameters in a valid JSON object in-line:\n`),
			Say(`\s+cf update-service SERVICE_INSTANCE -c '{\"name\":\"value\",\"name\":\"value\"}'\n`),
			Say(`\s+Optionally provide a file containing service-specific configuration parameters in a valid JSON object\.\n`),
			Say(`\s+The path to the parameters file can be an absolute or relative path to a file:\n`),
			Say(`\s+cf update-service SERVICE_INSTANCE -c PATH_TO_FILE\n`),
			Say(`\s+Example of valid JSON object:\n`),
			Say(`\s+{\n`),
			Say(`\s+\"cluster_nodes\": {\n`),
			Say(`\s+\"count\": 5,\n`),
			Say(`\s+\"memory_mb\": 1024\n`),
			Say(`\s+}\n`),
			Say(`\s+}\n`),
			Say(`\s+ Optionally provide a list of comma-delimited tags that will be written to the VCAP_SERVICES environment variable for any bound applications.\n`),
			Say(`EXAMPLES:\n`),
			Say(`\s+cf update-service mydb -p gold\n`),
			Say(`\s+cf update-service mydb -c '{\"ram_gb\":4}'\n`),
			Say(`\s+cf update-service mydb -c ~/workspace/tmp/instance_config.json\n`),
			Say(`\s+cf update-service mydb -t "list, of, tags"\n`),
			Say(`\s+cf update-service mydb --upgrade\n`),
			Say(`\s+cf update-service mydb --upgrade --force\n`),
			Say(`OPTIONS:\n`),
			Say(`\s+-c\s+Valid JSON object containing service-specific configuration parameters, provided either in-line or in a file\. For a list of supported configuration parameters, see documentation for the particular service offering\.\n`),
			Say(`\s+-p\s+Change service plan for a service instance\n`),
			Say(`\s+-t\s+User provided tags\n`),
			Say(`\s+--upgrade, -u\s+Upgrade the service instance to the latest version of the service plan available. It cannot be combined with flags: -c, -p, -t.\n`),
			Say(`\s+--force, -f\s+Force the upgrade to the latest available version of the service plan. It can only be used with: -u, --upgrade.\n`),
			Say(`SEE ALSO:\n`),
			Say(`\s+rename-service, services, update-user-provided-service\n`),
		)

		When("the --help flag is specified", func() {
			It("exits successfully and prints the help message", func() {
				session := helpers.CF(command, "--help")

				Eventually(session).Should(Exit(0))
				Expect(session.Out).To(matchHelpMessage)
				Expect(session.Err.Contents()).To(BeEmpty())
			})
		})

		When("the service instance name is omitted", func() {
			It("fails and prints the help message", func() {
				session := helpers.CF(command)

				Eventually(session).Should(Exit(1))
				Expect(session.Out).To(matchHelpMessage)
				Expect(session.Err).To(Say("Incorrect Usage: the required argument `SERVICE_INSTANCE` was not provided\n"))
			})
		})

		When("an extra parameter is provided", func() {
			It("fails and prints the help message", func() {
				session := helpers.CF(command, "service-instance-name", "invalid-extra-parameter")

				Eventually(session).Should(Exit(1))
				Expect(session.Out).To(matchHelpMessage)
				Expect(session.Err).To(Say(`Incorrect Usage: unexpected argument "invalid-extra-parameter"`))
			})
		})

		When("an extra flag is provided", func() {
			It("fails and prints the help message", func() {
				session := helpers.CF(command, "service-instance-name", "--invalid")

				Eventually(session).Should(Exit(1))
				Expect(session.Out).To(matchHelpMessage)
				Expect(session.Err).To(Say("Incorrect Usage: unknown flag `invalid'"))
			})
		})

		When("a flag is provided without a value", func() {
			DescribeTable(
				"it fails and prints a help message",
				func(flag string) {
					session := helpers.CF(command, "service-instance-name", flag)

					Eventually(session).Should(Exit(1))
					Expect(session.Out).To(matchHelpMessage)
					Expect(session.Err).To(Say("Incorrect Usage: expected argument for flag `%s'", flag))
				},
				Entry("configuration", "-c"),
				Entry("plan", "-p"),
				Entry("tags", "-t"),
			)
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, command, "service-instance-name")
		})
	})

	When("logged in and targeting a space", func() {
		var (
			orgName             string
			spaceName           string
			username            string
			serviceInstanceName string
		)

		waitUntilUpdateComplete := func(serviceInstanceName string) {
			Eventually(helpers.CF("service", serviceInstanceName)).Should(Say(`status:\s+update succeeded`))
		}

		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()
			helpers.SetupCF(orgName, spaceName)

			username, _ = helpers.GetCredentials()

			serviceInstanceName = helpers.NewServiceInstanceName()
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		When("the service instance doesn't exist", func() {
			It("fails with an appropriate error", func() {
				session := helpers.CF(command, serviceInstanceName, "-t", "important")
				Eventually(session).Should(Exit(1))
				Expect(session.Out).To(SatisfyAll(
					Say("Updating service instance %s in org %s / space %s as %s...", serviceInstanceName, orgName, spaceName, username),
					Say("FAILED"),
				))
				Expect(session.Err).To(Say("Service instance %s not found\n", serviceInstanceName))
			})
		})

		When("the service instance exists", func() {
			const (
				originalTags = "foo, bar"
				newTags      = "bax, quz"
			)
			var broker *servicebrokerstub.ServiceBrokerStub

			BeforeEach(func() {
				broker = servicebrokerstub.New().WithAsyncDelay(time.Microsecond).EnableServiceAccess()
				helpers.CreateManagedServiceInstance(
					broker.FirstServiceOfferingName(),
					broker.FirstServicePlanName(),
					serviceInstanceName,
					"-t", originalTags,
				)
			})

			AfterEach(func() {
				broker.Forget()
			})

			It("can update tags", func() {
				session := helpers.CF(command, serviceInstanceName, "-t", newTags)
				Eventually(session).Should(Exit(0))

				Expect(session.Out).To(SatisfyAll(
					Say("Updating service instance %s in org %s / space %s as %s...", serviceInstanceName, orgName, spaceName, username),
					Say(`\n`),
					Say(`OK`),
				))

				Expect(string(session.Err.Contents())).To(BeEmpty())

				waitUntilUpdateComplete(serviceInstanceName)
				session = helpers.CF("service", serviceInstanceName)
				Eventually(session).Should(Exit(0))
				Expect(session.Out).To(Say(`tags:\s+%s`, newTags))
			})

			Describe("updating parameters", func() {
				const (
					validParams   = `{"funky":"chicken"}`
					invalidParams = `{"funky":chicken"}`
				)

				checkParams := func() {
					session := helpers.CF("service", serviceInstanceName)
					Eventually(session).Should(Exit(0))
					Expect(session.Out).To(SatisfyAll(
						Say(`Showing parameters for service instance %s...\n`, serviceInstanceName),
						Say(`\n`),
						Say(`%s\n`, validParams),
					))
				}

				It("accepts JSON on the command line", func() {
					session := helpers.CF(command, serviceInstanceName, "-c", validParams)
					Eventually(session).Should(Exit(0))

					waitUntilUpdateComplete(serviceInstanceName)
					checkParams()
				})

				It("rejects invalid JSON on the command line", func() {
					session := helpers.CF(command, serviceInstanceName, "-c", invalidParams)
					Eventually(session).Should(Exit(1))

					Expect(session.Err).To(Say("Incorrect Usage: Invalid configuration provided for -c flag. Please provide a valid JSON object or path to a file containing a valid JSON object.\n"))
				})

				It("accepts JSON from a file", func() {
					path := helpers.TempFileWithContent(validParams)
					defer os.Remove(path)

					session := helpers.CF(command, serviceInstanceName, "-c", path, "-v")
					Eventually(session).Should(Exit(0))

					waitUntilUpdateComplete(serviceInstanceName)
					checkParams()
				})

				It("rejects invalid JSON from a file", func() {
					path := helpers.TempFileWithContent(invalidParams)
					defer os.Remove(path)

					session := helpers.CF(command, serviceInstanceName, "-c", path)
					Eventually(session).Should(Exit(1))

					Expect(session.Err).To(Say("Incorrect Usage: Invalid configuration provided for -c flag. Please provide a valid JSON object or path to a file containing a valid JSON object.\n"))
				})
			})
		})
	})
})
