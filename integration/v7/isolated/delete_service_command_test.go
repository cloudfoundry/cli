package isolated

import (
	"time"

	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/integration/helpers/servicebrokerstub"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("delete-service command", func() {
	const command = "delete-service"

	Describe("help", func() {
		matchHelpMessage := SatisfyAll(
			Say(`NAME:\n`),
			Say(`\s+%s - Delete a service instance\n`, command),
			Say(`\n`),
			Say(`USAGE:\n`),
			Say(`\s+cf delete-service SERVICE_INSTANCE \[-f\] \[-w\]\n`),
			Say(`\n`),
			Say(`ALIAS:\n`),
			Say(`\s+ds\n`),
			Say(`\n`),
			Say(`OPTIONS:\n`),
			Say(`\s+--force, -f\s+Force deletion without confirmation\n`),
			Say(`\s+--wait, -w\s+Wait for the operation to complete\n`),
			Say(`\n`),
			Say(`SEE ALSO:\n`),
			Say(`\s+services, unbind-service\n`),
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

		When("the service instance does not exist", func() {
			It("prints a message and exits successfully", func() {
				session := helpers.CF(command, "-f", serviceInstanceName)
				Eventually(session).Should(Exit(0))

				Expect(session.Out).To(SatisfyAll(
					Say("Deleting service instance %s in org %s / space %s as %s...", serviceInstanceName, orgName, spaceName, username),
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
			})

			It("prints a message and exits successfully", func() {
				session := helpers.CF(command, "-f", serviceInstanceName)
				Eventually(session).Should(Exit(0))

				Expect(session.Out).To(SatisfyAll(
					Say("Deleting service instance %s in org %s / space %s as %s...", serviceInstanceName, orgName, spaceName, username),
					Say("\n"),
					Say("Service instance %s deleted.\n", serviceInstanceName),
				))

				Expect(string(session.Err.Contents())).To(BeEmpty())

				session = helpers.CF("services").Wait()
				Expect(session.Out).NotTo(Say(serviceInstanceName))
			})
		})

		When("the service instance is managed by a synchronous broker", func() {
			var broker *servicebrokerstub.ServiceBrokerStub

			BeforeEach(func() {
				broker = servicebrokerstub.EnableServiceAccess()
				helpers.CreateManagedServiceInstance(broker.FirstServiceOfferingName(), broker.FirstServicePlanName(), serviceInstanceName)
			})

			AfterEach(func() {
				broker.Forget()
			})

			It("prints a message and exits successfully", func() {
				session := helpers.CF(command, "-f", serviceInstanceName)
				Eventually(session).Should(Exit(0))

				Expect(session.Out).To(SatisfyAll(
					Say("Deleting service instance %s in org %s / space %s as %s...", serviceInstanceName, orgName, spaceName, username),
					Say("\n"),
					Say(`Service instance %s deleted\.\n`, serviceInstanceName),
				))

				Expect(string(session.Err.Contents())).To(BeEmpty())
			})

			When("the wait flag is specified", func() {
				It("waits for the delete operation", func() {
					session := helpers.CF(command, "-f", "-w", serviceInstanceName)
					Eventually(session).Should(Exit(0))

					Expect(session.Out).To(SatisfyAll(
						Say("Deleting service instance %s in org %s / space %s as %s...", serviceInstanceName, orgName, spaceName, username),
						Say("\n"),
						Say("Waiting for the operation to complete."),
						Say("\n"),
						Say("Service instance %s deleted.\n", serviceInstanceName),
					))

					Expect(string(session.Err.Contents())).To(BeEmpty())

					session = helpers.CF("services").Wait()
					Expect(session.Out).NotTo(Say(serviceInstanceName))
				})
			})
		})

		When("the service instance is managed by an asynchronous broker", func() {
			var broker *servicebrokerstub.ServiceBrokerStub

			BeforeEach(func() {
				broker = servicebrokerstub.New().WithAsyncDelay(time.Second).EnableServiceAccess()
				helpers.CreateManagedServiceInstance(broker.FirstServiceOfferingName(), broker.FirstServicePlanName(), serviceInstanceName)
			})

			AfterEach(func() {
				broker.Forget()
			})

			It("prints a message and exits successfully", func() {
				session := helpers.CF(command, "-f", serviceInstanceName)
				Eventually(session).Should(Exit(0))

				Expect(session.Out).To(SatisfyAll(
					Say("Deleting service instance %s in org %s / space %s as %s...", serviceInstanceName, orgName, spaceName, username),
					Say("\n"),
					Say("Delete in progress. Use 'cf services' or 'cf service %s' to check operation status.\n", serviceInstanceName),
				))

				Expect(string(session.Err.Contents())).To(BeEmpty())
			})

			When("the wait flag is specified", func() {
				It("waits for the delete operation", func() {
					session := helpers.CF(command, "-f", "-w", serviceInstanceName)
					Eventually(session).Should(Exit(0))

					Expect(session.Out).To(SatisfyAll(
						Say("Deleting service instance %s in org %s / space %s as %s...", serviceInstanceName, orgName, spaceName, username),
						Say("\n"),
						Say("Waiting for the operation to complete."),
						Say("\n"),
						Say("Service instance %s deleted.\n", serviceInstanceName),
					))

					Expect(string(session.Err.Contents())).To(BeEmpty())

					session = helpers.CF("services").Wait()
					Expect(session.Out).NotTo(Say(serviceInstanceName))
				})
			})
		})
	})
})
