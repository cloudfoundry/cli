package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("delete-service command", func() {
	const command = "v3-delete-service"

	Describe("help", func() {
		matchHelpMessage := SatisfyAll(
			Say(`NAME:\n`),
			Say(`\s+%s - Delete a service instance\n`, command),
			Say(`\n`),
			Say(`USAGE:\n`),
			Say(`\s+cf delete-service SERVICE_INSTANCE \[-f\]\n`),
			Say(`\n`),
			Say(`ALIAS:\n`),
			Say(`\s+ds\n`),
			Say(`\n`),
			Say(`OPTIONS:\n`),
			Say(`\s+-f\s+Force deletion without confirmation\n`),
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
})
