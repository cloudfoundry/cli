package isolated

import (
	"time"

	"code.cloudfoundry.org/cli/integration/assets/hydrabroker/config"
	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/integration/helpers/servicebrokerstub"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("upgrade-service command", func() {
	const command = "upgrade-service"

	Describe("help", func() {
		When("--help flag is set", func() {
			helpMessage := SatisfyAll(
				Say("NAME:"),
				Say(`\s+upgrade-service - Upgrade a service instance to the latest available version of its current service plan`),
				Say(`USAGE:`),
				Say(`\s+cf upgrade-service SERVICE_INSTANCE`),
				Say(`OPTIONS:`),
				Say(`\s+--force, -f\s+Force upgrade without asking for confirmation`),
				Say(`\s+--wait, -w\s+Wait for the operation to complete\n`),
				Say(`SEE ALSO:`),
				Say(`\s+services, update-service, update-user-provided-service`),
			)

			It("exits successfully and prints the help message", func() {
				session := helpers.CF(command, "--help")

				Eventually(session).Should(Exit(0))
				Expect(session.Out).To(helpMessage)
				Expect(session.Err.Contents()).To(BeEmpty())
			})

			When("the service instance name is omitted", func() {
				It("fails and prints the help message", func() {
					session := helpers.CF(command)

					Eventually(session).Should(Exit(1))
					Expect(session.Out).To(helpMessage)
					Expect(session.Err).To(Say("Incorrect Usage: the required argument `SERVICE_INSTANCE` was not provided\n"))
				})
			})

			When("an extra parameter is provided", func() {
				It("fails and prints the help message", func() {
					session := helpers.CF(command, "service-instance-name", "invalid-extra-parameter")

					Eventually(session).Should(Exit(1))
					Expect(session.Out).To(helpMessage)
					Expect(session.Err).To(Say(`Incorrect Usage: unexpected argument "invalid-extra-parameter"`))
				})
			})

			When("an extra flag is provided", func() {
				It("fails and prints the help message", func() {
					session := helpers.CF(command, "service-instance-name", "--invalid")

					Eventually(session).Should(Exit(1))
					Expect(session.Out).To(helpMessage)
					Expect(session.Err).To(Say("Incorrect Usage: unknown flag `invalid'"))
				})
			})
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, command, "service-instance-name")
		})
	})

	When("logged in and targeting a space", func() {
		var orgName, spaceName, serviceInstanceName, username string

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
			It("prints a message and exits with error", func() {
				session := helpers.CF(command, "-f", serviceInstanceName)
				Eventually(session).Should(Exit(1))

				Expect(session.Out).To(SatisfyAll(
					Say("Upgrading service instance %s in org %s / space %s as %s...", serviceInstanceName, orgName, spaceName, username),
					Say("\n"),
					Say("FAILED"),
				))

				Expect(session.Err).To(
					Say("Service instance '%s' not found\n", serviceInstanceName),
				)
			})
		})

		When("the service instance exists", func() {
			var broker *servicebrokerstub.ServiceBrokerStub

			BeforeEach(func() {
				broker = servicebrokerstub.New().WithPlans(2).WithAsyncDelay(time.Microsecond).EnableServiceAccess()
				helpers.CreateManagedServiceInstance(
					broker.FirstServiceOfferingName(),
					broker.FirstServicePlanName(),
					serviceInstanceName,
				)
			})

			AfterEach(func() {
				broker.Forget()
			})

			Context("but there is no upgrade available", func() {
				It("prints a message and exits successfully", func() {
					session := helpers.CF(command, "-f", serviceInstanceName)
					Eventually(session).Should(Exit(0))

					Expect(session.Out).To(SatisfyAll(
						Say("Upgrading service instance %s in org %s / space %s as %s...", serviceInstanceName, orgName, spaceName, username),
						Say("\n"),
						Say("No upgrade is available."),
						Say("\n"),
						Say("OK"),
					))
				})
			})

			Context("and there's an upgrade available", func() {
				BeforeEach(func() {
					broker.Services[0].Plans[0].MaintenanceInfo = &config.MaintenanceInfo{Version: "9.1.2"}
					broker.Configure().Register()
				})

				It("upgrades the service instance", func() {
					session := helpers.CF(command, "-f", serviceInstanceName, "--wait")

					Eventually(session).Should(Exit(0))
					Expect(session.Out).To(SatisfyAll(
						Say(`Upgrading service instance %s in org %s / space %s as %s\.\.\.\n`, serviceInstanceName, orgName, spaceName, username),
						Say(`\n`),
						Say(`Waiting for the operation to complete\.+\n`),
						Say(`\n`),
						Say(`Upgrade of service instance %s complete\.\n`, serviceInstanceName),
						Say(`OK\n`),
					))
				})
			})
		})
	})
})
