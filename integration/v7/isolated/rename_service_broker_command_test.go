package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/integration/helpers/servicebrokerstub"
	. "github.com/onsi/ginkgo/v2"
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

				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("Service broker 'does-not-exist' not found."))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the rename fails", func() {
			It("prints an error message", func() {
				originalName := helpers.NewServiceBrokerName()

				broker1 := servicebrokerstub.New().Create().Register()
				broker2 := servicebrokerstub.New().WithName(originalName).Create().Register()
				defer broker1.Forget()
				defer broker2.Forget()

				session := helpers.CF("rename-service-broker", broker2.Name, broker1.Name)

				Eventually(session.Wait().Out).Should(SatisfyAll(
					Say("Renaming service broker %s to %s as %s", broker2.Name, broker1.Name, cfUsername),
					Say("FAILED"),
				))
				Eventually(session.Err).Should(Say("Name must be unique"))
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
			helpers.CheckEnvironmentTargetedCorrectly(false, false, ReadOnlyOrg, "rename-service-broker", "foo", "bar")
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
