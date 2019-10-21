package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/integration/helpers/fakeservicebroker"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	"github.com/onsi/gomega/types"
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
			broker1 := fakeservicebroker.New().EnsureBrokerIsAvailable()
			broker2 := fakeservicebroker.New().WithName("single-use").EnsureAppIsDeployed()
			defer broker2.Destroy()

			session := helpers.CF("update-service-broker", broker1.Name(), broker2.Username(), broker2.Password(), broker2.URL())

			Eventually(session.Wait().Out).Should(SatisfyAll(
				Say("Updating service broker %s as %s...", broker1.Name(), cfUsername),
				Say("OK"),
			))
			Eventually(session).Should(Exit(0))
			session = helpers.CF("service-brokers")
			Eventually(session.Out).Should(Say("%s[[:space:]]+%s", broker1.Name(), broker2.URL()))
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

		When("the update fails", func() {
			It("prints an error message", func() {
				broker := fakeservicebroker.New().EnsureBrokerIsAvailable()

				session := helpers.CF("update-service-broker", broker.Name(), broker.Username(), broker.Password(), "not-a-valid-url")

				Eventually(session.Wait().Out).Should(SatisfyAll(
					Say("Updating service broker %s as %s...", broker.Name(), cfUsername),
					Say("FAILED"),
				))

				Eventually(session.Err).Should(
					Say("Url must be a valid url"),
				)

				Eventually(session).Should(Exit(1))
			})
		})
	})

	When("passing incorrect parameters", func() {
		It("prints an error message", func() {
			session := helpers.CF("update-service-broker", "b1")

			Eventually(session.Err).Should(Say("Incorrect Usage: the required arguments `USERNAME`, `PASSWORD` and `URL` were not provided"))
			Eventually(session).Should(matchUpdateServiceBrokersHelpMessage())
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

			Eventually(session).Should(matchUpdateServiceBrokersHelpMessage())
			Eventually(session).Should(Exit(0))
		})
	})
})

func matchUpdateServiceBrokersHelpMessage() types.GomegaMatcher {
	return SatisfyAll(
		Say("NAME:"),
		Say("update-service-broker - Update a service broker"),
		Say("USAGE:"),
		Say("cf update-service-broker SERVICE_BROKER USERNAME PASSWORD URL"),
		Say("SEE ALSO:"),
		Say("rename-service-broker, service-brokers"),
	)
}
