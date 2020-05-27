package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/integration/helpers/servicebrokerstub"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("rename-service-broker command", func() {
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

		It("renames the service broker", func() {
			originalName := helpers.NewServiceBrokerName()
			updatedName := helpers.NewServiceBrokerName()

			broker := servicebrokerstub.New().WithName(originalName).Create().Register()
			defer broker.Forget()

			session := helpers.CF("rename-service-broker", originalName, updatedName)

			Eventually(session.Wait().Out).Should(SatisfyAll(
				Say("Renaming service broker %s to %s as %s", originalName, updatedName, cfUsername),
				Say("OK"),
			))
			Eventually(session).Should(Exit(0))
			session = helpers.CF("service-brokers")
			Eventually(session.Out).Should(Say(updatedName))

			broker.WithName(updatedName) // Forget() needs to know the new name
		})

		When("the service broker doesn't exist", func() {
			It("prints an error message", func() {
				session := helpers.CF("rename-service-broker", "does-not-exist", "new-name")

				Eventually(session).Should(SatisfyAll(
					Say("FAILED"),
					Say("Service Broker does-not-exist not found"),
				))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the rename fails", func() {
			It("prints an error message", func() {
				broker1 := servicebrokerstub.Register()
				broker2 := servicebrokerstub.Register()
				defer broker1.Forget()
				defer broker2.Forget()

				session := helpers.CF("rename-service-broker", broker2.Name, broker1.Name)

				Eventually(session.Wait().Out).Should(SatisfyAll(
					Say("Renaming service broker %s to %s as %s", broker2.Name, broker1.Name, cfUsername),
					Say("FAILED"),
					Say("Server error, status code: 400, error code: 270002, message: The service broker name is taken"),
				))
				Eventually(session).Should(Exit(1))
			})
		})
	})

	When("passing incorrect parameters", func() {
		It("prints an error message", func() {
			session := helpers.CF("rename-service-broker")

			Eventually(session.Err).Should(Say("Incorrect Usage: the required arguments `SERVICE_BROKER` and `NEW_SERVICE_BROKER` were not provided"))
			eventuallyRendersRenameServiceBrokerHelp(session)
			Eventually(session).Should(Exit(1))
		})
	})

	When("the environment is not targeted correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.UnrefactoredCheckEnvironmentTargetedCorrectly(false, false, ReadOnlyOrg, "rename-service-broker", "foo", "bar")
		})
	})

	When("passing --help", func() {
		It("displays command usage to output", func() {
			session := helpers.CF("rename-service-broker", "--help")

			eventuallyRendersRenameServiceBrokerHelp(session)
			Eventually(session).Should(Exit(0))
		})
	})
})

func eventuallyRendersRenameServiceBrokerHelp(s *Session) {
	Eventually(s).Should(Say("NAME:"))
	Eventually(s).Should(Say("rename-service-broker - Rename a service broker"))
	Eventually(s).Should(Say("USAGE:"))
	Eventually(s).Should(Say("cf rename-service-broker SERVICE_BROKER NEW_SERVICE_BROKER"))
	Eventually(s).Should(Say("SEE ALSO:"))
	Eventually(s).Should(Say("service-brokers, update-service-broker"))
}
