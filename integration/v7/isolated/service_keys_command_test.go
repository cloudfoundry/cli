package isolated

import (
	"time"

	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/integration/helpers/servicebrokerstub"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("service-keys command", func() {
	const command = "service-keys"

	Describe("help", func() {
		matchHelpMessage := SatisfyAll(
			Say(`NAME:\n`),
			Say(`\s+%s - List keys for a service instance\n`, command),
			Say(`\n`),
			Say(`USAGE:\n`),
			Say(`\s+cf %s SERVICE_INSTANCE\n`, command),
			Say(`\n`),
			Say(`EXAMPLES:\n`),
			Say(`\s+cf %s mydb\n`, command),
			Say(`\n`),
			Say(`ALIAS:\n`),
			Say(`\s+sk\n`),
			Say(`\n`),
			Say(`SEE ALSO:\n`),
			Say(`\s+delete-service-key\n`),
		)

		When("the -h flag is specified", func() {
			It("succeeds and prints help", func() {
				session := helpers.CF(command, "-h")
				Eventually(session).Should(Exit(0))
				Expect(session.Out).To(matchHelpMessage)
			})
		})

		When("the --help flag is specified", func() {
			It("succeeds and prints help", func() {
				session := helpers.CF(command, "--help")
				Eventually(session).Should(Exit(0))
				Expect(session.Out).To(matchHelpMessage)
			})
		})

		When("no arguments are provided", func() {
			It("displays a warning, the help text, and exits 1", func() {
				session := helpers.CF(command)
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Incorrect Usage: the required argument `SERVICE_INSTANCE` was not provided"))
				Expect(session.Out).To(matchHelpMessage)
			})
		})

		When("unknown flag is passed", func() {
			It("displays a warning, the help text, and exits 1", func() {
				session := helpers.CF(command, "-u")
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Incorrect Usage: unknown flag `u"))
				Expect(session.Out).To(matchHelpMessage)
			})
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, command, "foo")
		})
	})

	When("there is a service instance", func() {
		var (
			userName            string
			orgName             string
			spaceName           string
			serviceInstanceName string
			broker              *servicebrokerstub.ServiceBrokerStub
		)

		BeforeEach(func() {
			userName, _ = helpers.GetCredentials()

			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()
			helpers.SetupCF(orgName, spaceName)

			broker = servicebrokerstub.EnableServiceAccess()

			serviceInstanceName = helpers.NewServiceInstanceName()
			helpers.CreateManagedServiceInstance(broker.FirstServiceOfferingName(), broker.FirstServicePlanName(), serviceInstanceName)

			broker.WithAsyncDelay(time.Millisecond).Configure()
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
			broker.Forget()
		})

		It("reports that there are no keys", func() {
			session := helpers.CF(command, serviceInstanceName)
			Eventually(session).Should(Exit(0))
			Expect(string(session.Err.Contents())).To(BeEmpty())
			Expect(session.Out).To(SatisfyAll(
				Say(`Getting keys for service instance %s as %s\.\.\.\n`, serviceInstanceName, userName),
				Say(`\n`),
				Say(`No service keys for service instance %s\n`, serviceInstanceName),
			))
		})

		When("there are service keys", func() {
			var (
				keyName1 string
				keyName2 string
			)

			BeforeEach(func() {
				keyName1 = helpers.RandomName()
				keyName2 = helpers.RandomName()
				Eventually(helpers.CF("create-service-key", serviceInstanceName, keyName1, "--wait")).Should(Exit(0))
				Eventually(helpers.CF("create-service-key", serviceInstanceName, keyName2, "--wait")).Should(Exit(0))
			})

			It("prints the names of the keys in all spaces", func() {
				session := helpers.CF(command, serviceInstanceName)
				Eventually(session).Should(Exit(0))
				Expect(string(session.Err.Contents())).To(BeEmpty())
				Expect(session.Out).To(SatisfyAll(
					Say(`Getting keys for service instance %s as %s\.\.\.\n`, serviceInstanceName, userName),
					Say(`\n`),
					Say(`name\s+last operation\s+message\n`),
					Say(`%s\s+%s\s+%s\n`, keyName1, "create succeeded", "very happy service"),
					Say(`%s\s+%s\s+%s\n`, keyName2, "create succeeded", "very happy service"),
				))
			})
		})
	})
})
