package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
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

		When("the environment is not setup correctly", func() {
			It("fails with the appropriate errors", func() {
				helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, command, "service-instance-name")
			})
		})

		When("logged in and targeting a space", func() {
			var (
				orgName, spaceName, serviceInstanceName string
			)

			BeforeEach(func() {
				orgName = helpers.NewOrgName()
				spaceName = helpers.NewSpaceName()
				helpers.SetupCF(orgName, spaceName)
				serviceInstanceName = helpers.NewServiceInstanceName()
			})

			AfterEach(func() {
				helpers.QuickDeleteOrg(orgName)
			})

			It("fails with not implemented", func() {
				session := helpers.CF(command, serviceInstanceName)

				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("WIP: Not yet implemented"))
			})
		})
	})
})
