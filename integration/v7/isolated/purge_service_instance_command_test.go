package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/integration/helpers/servicebrokerstub"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("purge-service-instance command", func() {
	const command = "purge-service-instance"

	Describe("help", func() {
		matchHelpMessage := SatisfyAll(
			Say(`NAME:\n`),
			Say(`\s+%s - Recursively remove a service instance and child objects from Cloud Foundry database without making requests to a service broker\n`, command),
			Say(`\n`),
			Say(`USAGE:\n`),
			Say(`\s+cf purge-service-instance SERVICE_INSTANCE \[-f\]\n`),
			Say(`\n`),
			Say(`WARNING: This operation assumes that the service broker responsible for this service instance is no longer available or is not responding with a 200 or 410, and the service instance has been deleted, leaving orphan records in Cloud Foundry's database. All knowledge of the service instance will be removed from Cloud Foundry, including service bindings and service keys.\n`),
			Say(`\n`),
			Say(`OPTIONS:\n`),
			Say(`\s+--force, -f\s+Force deletion without confirmation\n`),
			Say(`\n`),
			Say(`SEE ALSO:\n`),
			Say(`\s+delete-service, service-brokers, services\n`),
		)

		When("--help is specified", func() {
			It("exits successfully and print the help message", func() {
				session := helpers.CF(command, "--help")
				Eventually(session).Should(Exit(0))

				Expect(session.Out).To(matchHelpMessage)
				Expect(string(session.Err.Contents())).To(BeEmpty())
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
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, command, "service-instance-name")
		})
	})

	When("targeting a space", func() {
		var (
			serviceInstanceName string
			orgName             string
			spaceName           string
			username            string
			appName             string
		)

		bindToApp := func() {
			appName = helpers.NewAppName()
			helpers.WithHelloWorldApp(func(appDir string) {
				Eventually(helpers.CF("push", appName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
			})

			Eventually(helpers.CF("bind-service", appName, serviceInstanceName)).Should(Exit(0))
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

		When("the service instance does not exist", func() {
			It("prints a message and exits successfully", func() {
				session := helpers.CF(command, "-f", serviceInstanceName)
				Eventually(session).Should(Exit(0))

				Expect(session.Out).To(SatisfyAll(
					Say("Purging service instance %s in org %s / space %s as %s...", serviceInstanceName, orgName, spaceName, username),
					Say("\n"),
					Say("Service instance %s did not exist.\n", serviceInstanceName),
				))

				Expect(string(session.Err.Contents())).To(BeEmpty())
			})
		})

		When("the service instance is user-provided", func() {
			BeforeEach(func() {
				session := helpers.CF("cups", serviceInstanceName)
				Eventually(session).Should(Exit(0))

				bindToApp()
			})

			It("prints a message and exits successfully", func() {
				session := helpers.CF(command, "-f", serviceInstanceName)
				Eventually(session).Should(Exit(0))

				Expect(session.Out).To(SatisfyAll(
					Say("Purging service instance %s in org %s / space %s as %s...", serviceInstanceName, orgName, spaceName, username),
					Say("\n"),
					Say("Service instance %s purged.\n", serviceInstanceName),
				))

				Expect(string(session.Err.Contents())).To(BeEmpty())

				session = helpers.CF("services").Wait()
				Expect(session.Out).NotTo(Say(serviceInstanceName))
			})
		})

		When("the service instance is managed", func() {
			var broker *servicebrokerstub.ServiceBrokerStub

			BeforeEach(func() {
				broker = servicebrokerstub.EnableServiceAccess()
				helpers.CreateManagedServiceInstance(broker.FirstServiceOfferingName(), broker.FirstServicePlanName(), serviceInstanceName)

				bindToApp()
			})

			AfterEach(func() {
				broker.Forget()
			})

			It("prints a message and exits successfully", func() {
				session := helpers.CF(command, "-f", serviceInstanceName)
				Eventually(session).Should(Exit(0))

				Expect(session.Out).To(SatisfyAll(
					Say("Purging service instance %s in org %s / space %s as %s...", serviceInstanceName, orgName, spaceName, username),
					Say("\n"),
					Say("Service instance %s purged.\n", serviceInstanceName),
				))

				Expect(string(session.Err.Contents())).To(BeEmpty())
			})
		})
	})
})
