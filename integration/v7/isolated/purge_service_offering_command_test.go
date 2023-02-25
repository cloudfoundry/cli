package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/integration/helpers/servicebrokerstub"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("purge-service-offering command", func() {
	Describe("help", func() {
		matchHelpMessage := SatisfyAll(
			Say("NAME:"),
			Say("purge-service-offering - Recursively remove a service offering and child objects from Cloud Foundry database without making requests to a service broker"),
			Say("USAGE:"),
			Say(`cf purge-service-offering SERVICE_OFFERING \[-b BROKER\] \[-f\]`),
			Say("WARNING: This operation assumes that the service broker responsible for this service offering is no longer available, and all service instances have been deleted, leaving orphan records in Cloud Foundry's database\\. All knowledge of the service offering will be removed from Cloud Foundry, including service instances and service bindings\\. No attempt will be made to contact the service broker; running this command without destroying the service broker will cause orphan service instances\\. After running this command you may want to run either delete-service-auth-token or delete-service-broker to complete the cleanup\\."),
			Say("OPTIONS:"),
			Say("-b\\s+Purge a service offering from a particular service broker. Required when service offering name is ambiguous"),
			Say("-f\\s+Force deletion without confirmation"),
			Say("SEE ALSO:"),
			Say("marketplace, purge-service-instance, service-brokers"),
		)

		When("the --help flag is set", func() {
			It("displays command usage to output", func() {
				session := helpers.CF("purge-service-offering", "--help")

				Eventually(session).Should(Exit(0))
				Expect(session.Out).To(matchHelpMessage)
			})
		})

		When("no args are passed", func() {
			It("displays an error message with help text", func() {
				session := helpers.CF("purge-service-offering")

				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Incorrect Usage: the required argument `SERVICE_OFFERING` was not provided"))
				Expect(session.Out).To(matchHelpMessage)
			})
		})

		When("more than required number of args are passed", func() {
			It("displays an error message with help text and exits 1", func() {
				session := helpers.CF("purge-service-offering", "service-name", "extra")

				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say(`Incorrect Usage: unexpected argument "extra"`))
				Expect(session.Out).To(matchHelpMessage)
			})
		})
	})

	When("logged in", func() {
		var orgName, spaceName string

		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()
			helpers.SetupCF(orgName, spaceName)
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		When("the service exists", func() {
			var broker *servicebrokerstub.ServiceBrokerStub

			BeforeEach(func() {
				broker = servicebrokerstub.EnableServiceAccess()
			})

			AfterEach(func() {
				broker.Forget()
			})

			It("purges the service offering and plans", func() {
				session := helpers.CF("purge-service-offering", broker.FirstServiceOfferingName(), "-f")
				Eventually(session).Should(Exit(0))

				Expect(session).To(Say(`Purging service offering %s\.\.\.`, broker.FirstServiceOfferingName()))
				Expect(session).To(Say(`OK`))

				session = helpers.CF("marketplace")
				Eventually(session).Should(Exit(0))
				Expect(session).NotTo(Say(broker.FirstServiceOfferingName()))
				Expect(session).NotTo(Say(broker.FirstServicePlanName()))
			})

			When("the service name is ambiguous", func() {
				var secondBroker *servicebrokerstub.ServiceBrokerStub

				BeforeEach(func() {
					secondBroker = servicebrokerstub.New()
					secondBroker.Services[0].Name = broker.FirstServiceOfferingName()
					secondBroker.Create().Register().EnableServiceAccess()
				})

				AfterEach(func() {
					secondBroker.Forget()
				})

				It("fails, asking the user to disambiguate", func() {
					session := helpers.CF("purge-service-offering", broker.FirstServiceOfferingName(), "-f")
					Eventually(session).Should(Exit(1))
					Expect(session.Err).To(Say(`Service '%s' is provided by multiple service brokers: %s, %s`, broker.FirstServiceOfferingName(), broker.Name, secondBroker.Name))
					Expect(session.Err).To(Say(`Specify a broker by using the '-b' flag.`))
				})
			})
		})

		When("the service does not exist", func() {
			It("succeeds, printing a message", func() {
				session := helpers.CF("purge-service-offering", "no-such-service", "-f")

				Eventually(session).Should(Exit(0))
				Expect(session.Out).To(Say(`Service offering 'no-such-service' not found.`))
				Expect(session.Out).To(Say(`OK`))
			})
		})
	})

	When("the environment is not targeted correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(false, false, ReadOnlyOrg, "purge-service-offering", "-f", "foo")
		})
	})
})
