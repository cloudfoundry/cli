package isolated

import (
	"strings"

	. "github.com/onsi/gomega/gbytes"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("service command", func() {
	Describe("help", func() {
		const serviceInstanceName = "fake-service-instance-name"

		matchHelpMessage := SatisfyAll(
			Say(`NAME:\n`),
			Say(`\s+service - Show service instance info\n`),
			Say(`\n`),
			Say(`USAGE:\n`),
			Say(`\s+cf service SERVICE_INSTANCE\n`),
			Say(`\n`),
			Say(`OPTIONS:\n`),
			Say(`\s+--guid\s+Retrieve and display the given service's guid. All other output for the service is suppressed.\n`),
			Say(`\n`),
			Say(`SEE ALSO:\n`),
			Say(`\s+bind-service, rename-service, update-service\n`),
			Say(`$`),
		)

		When("the -h flag is specified", func() {
			It("succeeds and prints help", func() {
				session := helpers.CF("service", serviceInstanceName, "-h")
				Eventually(session).Should(Exit(0))
				Expect(session.Out).To(matchHelpMessage)
			})
		})

		When("the service instance name is missing", func() {
			It("fails with an error and prints help", func() {
				session := helpers.CF("service")
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Incorrect Usage: the required argument `SERVICE_INSTANCE` was not provided"))
				Expect(session.Out).To(matchHelpMessage)
			})
		})

		When("an extra parameter is specified", func() {
			It("fails with an error and prints help", func() {
				session := helpers.CF("service", serviceInstanceName, "anotherRandomParameter")
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
				session := helpers.CF("service", serviceInstanceName, "--anotherRandomFlag")
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Incorrect Usage: unknown flag `anotherRandomFlag'"))
				Expect(session.Out).To(matchHelpMessage)
			})
		})
	})

	When("environment is not set up", func() {
		It("displays an error and exits 1", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "service", "serviceInstanceName")
		})
	})

	When("user is logged in and targeting a space", func() {
		var (
			serviceInstanceName string
			orgName             string
			spaceName           string
		)

		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()
			helpers.SetupCF(orgName, spaceName)

			serviceInstanceName = helpers.NewServiceInstanceName()
			Eventually(helpers.CF("create-user-provided-service", serviceInstanceName)).Should(Exit(0))
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		It("can show the GUID", func() {
			session := helpers.CF("service", serviceInstanceName, "--guid")
			Eventually(session).Should(Exit(0))
			Expect(strings.TrimSpace(string(session.Out.Contents()))).To(HaveLen(36), "GUID wrong length")
		})
	})
})
