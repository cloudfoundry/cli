package isolated

import (
	"os"
	"time"

	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/integration/helpers/servicebrokerstub"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("update-service command", func() {
	const command = "update-service"

	Describe("help", func() {
		matchHelpMessage := SatisfyAll(
			Say(`NAME:\n`),
			Say(`\s+%s - Update a service instance\n`, command),
			Say(`\n`),
			Say(`USAGE:\n`),
			Say(`\s+cf update-service SERVICE_INSTANCE \[-p NEW_PLAN\] \[-c PARAMETERS_AS_JSON\] \[-t TAGS\]\n`),
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
			Say(`OPTIONS:\n`),
			Say(`\s+-c\s+Valid JSON object containing service-specific configuration parameters, provided either in-line or in a file\. For a list of supported configuration parameters, see documentation for the particular service offering\.\n`),
			Say(`\s+-p\s+Change service plan for a service instance\n`),
			Say(`\s+-t\s+User provided tags\n`),
			Say(`\s+--wait, -w\s+Wait for the operation to complete\n`),
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
		const serviceCommand = "service"

		var (
			orgName             string
			spaceName           string
			username            string
			serviceInstanceName string
		)

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
				Expect(session.Err).To(Say("Service instance '%s' not found\n", serviceInstanceName))
			})
		})

		When("the service instance exists", func() {
			const (
				originalTags = "foo, bar"
				newTags      = "bax, quz"
			)
			var broker *servicebrokerstub.ServiceBrokerStub

			BeforeEach(func() {
				broker = servicebrokerstub.New().WithPlans(2).EnableServiceAccess()
				helpers.CreateManagedServiceInstance(
					broker.FirstServiceOfferingName(),
					broker.FirstServicePlanName(),
					serviceInstanceName,
					"-t", originalTags,
				)
				session := helpers.CF("m")
				Eventually(session).Should(Exit(0))

			})

			AfterEach(func() {
				broker.Forget()
			})

			It("can update tags, parameters and plan", func() {
				By("performing the update")
				newPlanName := broker.Services[0].Plans[1].Name
				session := helpers.CF(command, serviceInstanceName, "-t", newTags, "-p", newPlanName, "-c", `{"foo":"bar"}`)
				Eventually(session).Should(Exit(0))

				Expect(session.Out).To(SatisfyAll(
					Say(`Updating service instance %s in org %s / space %s as %s\.\.\.\n`, serviceInstanceName, orgName, spaceName, username),
					Say(`\n`),
					Say(`Update of service instance %s complete\.\n`, serviceInstanceName),
					Say(`OK\n`),
				))

				Expect(string(session.Err.Contents())).To(BeEmpty())

				By("checking that the update changed things")
				session = helpers.CF(serviceCommand, serviceInstanceName)
				Eventually(session).Should(Exit(0))

				Expect(session.Out).To(SatisfyAll(
					Say(`plan:\s+%s`, newPlanName),
					Say(`tags:\s+%s`, newTags),
				))
			})

			Describe("updating tags", func() {
				It("can update tags alone", func() {
					session := helpers.CF(command, serviceInstanceName, "-t", newTags)
					Eventually(session).Should(Exit(0))

					Expect(session.Out).To(SatisfyAll(
						Say(`Updating service instance %s in org %s / space %s as %s\.\.\.\n`, serviceInstanceName, orgName, spaceName, username),
						Say(`\n`),
						Say(`Update of service instance %s complete\.\n`, serviceInstanceName),
						Say(`OK\n`),
					))

					Expect(string(session.Err.Contents())).To(BeEmpty())

					session = helpers.CF(serviceCommand, serviceInstanceName)
					Eventually(session).Should(Exit(0))
					Expect(session.Out).To(Say(`tags:\s+%s`, newTags))
				})
			})

			Describe("updating parameters", func() {
				const (
					validParams   = `{"funky":"chicken"}`
					invalidParams = `{"funky":chicken"}`
				)

				checkParams := func() {
					session := helpers.CF(serviceCommand, serviceInstanceName, "--params")
					Eventually(session).Should(Exit(0))
					Expect(session.Out).To(SatisfyAll(
						Say(`\{\n`),
						Say(`  "funky": "chicken"\n`),
						Say(`\}\n`),
					))
				}

				It("accepts JSON on the command line", func() {
					session := helpers.CF(command, serviceInstanceName, "-c", validParams)
					Eventually(session).Should(Exit(0))

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

			Describe("updating plan", func() {
				It("updates the plan", func() {
					newPlanName := broker.Services[0].Plans[1].Name
					session := helpers.CF(command, serviceInstanceName, "-p", newPlanName)
					Eventually(session).Should(Exit(0))

					Eventually(helpers.CF("service", serviceInstanceName)).Should(
						SatisfyAll(
							Say(`plan:\s+%s`, newPlanName),
							Say(`status:\s+update succeeded`),
						),
					)
				})

				When("plan does not exist", func() {
					const invalidPlan = "invalid-plan"

					It("displays an error and exits 1", func() {
						session := helpers.CF(command, serviceInstanceName, "-p", invalidPlan)
						Eventually(session).Should(Exit(1))
						Expect(session.Err).To(Say("The plan '%s' could not be found for service offering '%s' and broker '%s'.", invalidPlan, broker.Services[0].Name, broker.Name))
					})
				})

				When("plan is the same as current", func() {
					It("displays a message and exits 0", func() {
						session := helpers.CF(command, serviceInstanceName, "-p", broker.Services[0].Plans[0].Name)
						Eventually(session).Should(Exit(0))
						Expect(session.Out).To(SatisfyAll(
							Say(`No changes were made\.\n`),
							Say(`OK\n`),
						))
					})
				})
			})

			When("the broker responds asynchronously", func() {
				BeforeEach(func() {
					broker.WithAsyncDelay(time.Second).Configure()
				})

				It("says that the operation is in progress", func() {
					newPlanName := broker.Services[0].Plans[1].Name
					session := helpers.CF(command, serviceInstanceName, "-p", newPlanName)
					Eventually(session).Should(Exit(0))

					Expect(session.Out).To(SatisfyAll(
						Say(`Updating service instance %s in org %s / space %s as %s\.\.\.\n`, serviceInstanceName, orgName, spaceName, username),
						Say(`\n`),
						Say(`Update in progress. Use 'cf services' or 'cf service %s' to check operation status\.\n`, serviceInstanceName),
						Say(`OK\n`),
					))
				})

				It("accepts the --wait flag to wait for completion", func() {
					session := helpers.CF(command, serviceInstanceName, "-c", `{"funky":"chicken"}`, "--wait")
					Eventually(session).Should(Exit(0))

					Expect(session.Out).To(SatisfyAll(
						Say(`Updating service instance %s in org %s / space %s as %s\.\.\.\n`, serviceInstanceName, orgName, spaceName, username),
						Say(`\n`),
						Say(`Waiting for the operation to complete\.+\n`),
						Say(`\n`),
						Say(`Update of service instance %s complete\.\n`, serviceInstanceName),
						Say(`OK\n`),
					))

					Expect(string(session.Err.Contents())).To(BeEmpty())
					session = helpers.CF(serviceCommand, serviceInstanceName)
					Eventually(session).Should(Exit(0))
					Expect(session.Out).To(Say(`status:\s+update succeeded`))
				})
			})
		})
	})
})
