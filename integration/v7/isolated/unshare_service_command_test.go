package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("unshare-service command", func() {
	const unshareServiceCommand = "v3-unshare-service"

	Describe("help", func() {
		const serviceInstanceName = "fake-service-instance-name"

		matchHelpMessage := SatisfyAll(
			Say("NAME:"),
			Say("unshare-service - Unshare a shared service instance from a space"),
			Say("USAGE:"),
			Say(`cf unshare-service SERVICE_INSTANCE OTHER_SPACE \[-o OTHER_ORG\] \[-f\]`),
			Say("OPTIONS:"),
			Say(`-o\s+Org of the other space \(Default: targeted org\)`),
			Say(`-f\s+Force unshare without confirmation`),
			Say("SEE ALSO:"),
			Say("delete-service, service, services, share-service, unbind-service"),
		)

		When("the -h flag is specified", func() {
			It("succeeds and prints help", func() {
				session := helpers.CF(unshareServiceCommand, "-h")
				Eventually(session).Should(Exit(0))
				Expect(session.Out).To(matchHelpMessage)

			})
		})

		When("the service instance name and space name is missing", func() {
			It("fails with an error and prints help", func() {
				session := helpers.CF(unshareServiceCommand)
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Incorrect Usage: the required arguments `SERVICE_INSTANCE` and `OTHER_SPACE` were not provided"))
				Expect(session.Out).To(matchHelpMessage)
			})
		})

		When("the space name is missing", func() {
			It("fails with an error and prints help", func() {
				session := helpers.CF(unshareServiceCommand, serviceInstanceName)
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Incorrect Usage: the required argument `OTHER_SPACE` was not provided"))
				Expect(session.Out).To(matchHelpMessage)
			})
		})

		When("the org name is missing for the flag", func() {
			It("fails with an error and prints help", func() {
				session := helpers.CF(unshareServiceCommand, serviceInstanceName, "space-name", "-o")
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Incorrect Usage: expected argument for flag `-o"))
				Expect(session.Out).To(matchHelpMessage)
			})
		})

		When("an extra parameter is specified", func() {
			It("fails with an error and prints help", func() {
				session := helpers.CF(unshareServiceCommand, serviceInstanceName, "space-name", "anotherRandomParameter")
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say(`Incorrect Usage: unexpected argument "anotherRandomParameter"`))
				Expect(session.Out).To(SatisfyAll(
					Say(`FAILED\n\n`),
					matchHelpMessage,
				))
			})
		})

		When("an extra flag is specified", func() {
			It("fails with an error and prints help", func() {
				session := helpers.CF(unshareServiceCommand, serviceInstanceName, "--anotherRandomFlag")
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Incorrect Usage: unknown flag `anotherRandomFlag'"))
				Expect(session.Out).To(matchHelpMessage)
			})
		})
	})
})
