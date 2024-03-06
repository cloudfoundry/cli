package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/integration/helpers/servicebrokerstub"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("update-service-broker command", func() {
	When("logged in", func() {
		var (
			org        string
			cfUsername string
		)

		BeforeEach(func() {
			org = helpers.SetupCFWithGeneratedOrgAndSpaceNames()
			cfUsername, _ = helpers.GetCredentials()
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(org)
		})

		It("updates the service broker", func() {
			broker1 := servicebrokerstub.Register()
			broker2 := servicebrokerstub.Create()
			defer broker1.Forget()
			defer broker2.Forget()

			session := helpers.CF("update-service-broker", broker1.Name, broker2.Username, broker2.Password, broker2.URL)

			Eventually(session.Wait().Out).Should(SatisfyAll(
				Say("Updating service broker %s as %s...", broker1.Name, cfUsername),
				Say("OK"),
			))
			Eventually(session).Should(Exit(0))
			session = helpers.CF("service-brokers")
			Eventually(session.Out).Should(Say("%s[[:space:]]+%s", broker1.Name, broker2.URL))
		})

		When("the service broker was updated but warnings happened", func() {
			var (
				serviceInstance string
				broker          *servicebrokerstub.ServiceBrokerStub
			)

			BeforeEach(func() {
				broker = servicebrokerstub.EnableServiceAccess()

				serviceInstance = helpers.NewServiceInstanceName()
				session := helpers.CF("create-service", broker.FirstServiceOfferingName(), broker.FirstServicePlanName(), serviceInstance, "-b", broker.Name)
				Eventually(session).Should(Exit(0))

				broker.Services[0].Plans[0].Name = "different-plan-name"
				broker.Services[0].Plans[0].ID = "different-plan-id"
				broker.Configure()
			})

			AfterEach(func() {
				helpers.CF("delete-service", "-f", serviceInstance)
				broker.Forget()
			})

			It("should yield a warning", func() {
				session := helpers.CF("update-service-broker", broker.Name, broker.Username, broker.Password, broker.URL)

				Eventually(session.Wait().Out).Should(SatisfyAll(
					Say("Updating service broker %s as %s...", broker.Name, cfUsername),
					Say("OK"),
				))
				Eventually(session.Err).Should(Say("Warning: Service plans are missing from the broker's catalog"))
			})
		})

		When("the service broker doesn't exist", func() {
			It("prints an error message", func() {
				session := helpers.CF("update-service-broker", "does-not-exist", "test-user", "test-password", "http://test.com")

				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(SatisfyAll(
					Say("Service broker 'does-not-exist' not found"),
				))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the update fails before starting a synchronization job", func() {
			It("prints an error message", func() {
				broker := servicebrokerstub.Register()

				session := helpers.CF("update-service-broker", broker.Name, broker.Username, broker.Password, "not-a-valid-url")

				Eventually(session.Wait().Out).Should(SatisfyAll(
					Say("Updating service broker %s as %s...", broker.Name, cfUsername),
					Say("FAILED"),
				))

				Eventually(session.Err).Should(
					Say("must be a valid url"),
				)

				Eventually(session).Should(Exit(1))
			})
		})

		When("the update fails after starting a synchronization job", func() {
			var broker *servicebrokerstub.ServiceBrokerStub

			BeforeEach(func() {
				broker = servicebrokerstub.Register()
				broker.WithCatalogResponse(500).Configure()
			})

			AfterEach(func() {
				broker.Forget()
			})

			It("prints an error message and the job guid", func() {
				session := helpers.CF("update-service-broker", broker.Name, broker.Username, broker.Password, broker.URL)

				Eventually(session.Wait().Out).Should(SatisfyAll(
					Say("Updating service broker %s as %s...", broker.Name, cfUsername),
					Say("FAILED"),
				))

				Eventually(session.Err).Should(SatisfyAll(
					Say("Job (.*) failed"),
					Say("The service broker returned an invalid response"),
					Say("Status Code: 500 Internal Server Error"),
				))

				Eventually(session).Should(Exit(1))
			})
		})
	})

	When("passing incorrect parameters", func() {
		It("prints an error message", func() {
			session := helpers.CF("update-service-broker", "b1")

			Eventually(session.Err).Should(Say("Incorrect Usage: the required arguments `USERNAME`, `PASSWORD` and `URL` were not provided"))
			eventuallyRendersUpdateServiceBrokerHelp(session)
			Eventually(session).Should(Exit(1))
		})
	})

	When("the environment is not targeted correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(false, false, ReadOnlyOrg, "update-service-broker", "broker-name", "username", "password", "https://test.com")
		})
	})

	When("passing --help", func() {
		It("displays command usage to output", func() {
			session := helpers.CF("update-service-broker", "--help")

			eventuallyRendersUpdateServiceBrokerHelp(session)
			Eventually(session).Should(Exit(0))
		})
	})
})

func eventuallyRendersUpdateServiceBrokerHelp(s *Session) {
	Eventually(s).Should(Say("NAME:"))
	Eventually(s).Should(Say("update-service-broker - Update a service broker"))
	Eventually(s).Should(Say("USAGE:"))
	Eventually(s).Should(Say("cf update-service-broker SERVICE_BROKER USERNAME PASSWORD URL"))
	Eventually(s).Should(Say("SEE ALSO:"))
	Eventually(s).Should(Say("rename-service-broker, service-brokers"))
}
