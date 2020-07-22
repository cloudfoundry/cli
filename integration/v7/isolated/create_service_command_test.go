package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/integration/helpers/servicebrokerstub"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("create-service command", func() {
	Describe("help", func() {

		matchHelpMessage := SatisfyAll(
			Say(`NAME:\n`),
			Say(`\s+create-service - Create a service instance\n`),
			Say(`\n`),
			Say(`USAGE:\n`),
			Say(`\s+cf create-service SERVICE PLAN SERVICE_INSTANCE \[-b SERVICE_BROKER\] \[-c JSON_PARAMS\] \[-t TAGS\]\n`),
			Say(`\s+Optionally provide service-specific configuration parameters in a valid JSON object in-line:\n`),
			Say(`\s+cf create-service SERVICE PLAN SERVICE_INSTANCE -c '{\"name\":\"value\",\"name\":\"value\"}'\n`),
			Say(`\s+Optionally provide a file containing service-specific configuration parameters in a valid JSON object\.\n`),
			Say(`\s+The path to the parameters file can be an absolute or relative path to a file:\n`),
			Say(`\s+cf create-service SERVICE PLAN SERVICE_INSTANCE -c PATH_TO_FILE\n`),
			Say(`\s+Example of valid JSON object:`),
			Say(`\s+{`),
			Say(`\s+\"cluster_nodes\": {`),
			Say(`\s+\"count\": 5,`),
			Say(`\s+\"memory_mb\": 1024`),
			Say(`\s+}`),
			Say(`\s+}`),
			Say(`TIP:`),
			Say(`\s+Use 'cf create-user-provided-service' to make user-provided service instances available to CF apps`),
			Say(`EXAMPLES:`),
			Say(`\s+Linux/Mac:\n`),
			Say(`\s+cf create-service db-service silver mydb -c '{\"ram_gb\":4}`),
			Say(`\s+Windows Command Line:`),
			Say(`\s+cf create-service db-service silver mydb -c \"{\\\"ram_gb\\\":4}\"`),
			Say(`\s+Windows PowerShell:`),
			Say(`\s+cf create-service db-service silver mydb -c '{\\\"ram_gb\\\":4}'`),
			Say(`\s+cf create-service db-service silver mydb -c ~/workspace/tmp/instance_config.json`),
			Say(`\s+cf create-service db-service silver mydb -t \"list, of, tags\"`),
			Say(`ALIAS:`),
			Say(`\s+cs`),
			Say(`OPTIONS:`),
			Say(`\s+-b      Create a service instance from a particular broker\. Required when service offering name is ambiguous`),
			Say(`\s+-c      Valid JSON object containing service-specific configuration parameters, provided either in-line or in a file\. For a list of supported configuration parameters, see documentation for the particular service offering\.`),
			Say(`\s+-t      User provided tags`),
			Say(`SEE ALSO:`),
			Say(`\s+bind-service, create-user-provided-service, marketplace, services`),
		)

		When("the -h flag is specified", func() {
			It("succeeds and prints help", func() {
				session := helpers.CF("create-service", "-h")
				Eventually(session).Should(Exit(0))
				Expect(session.Out).To(matchHelpMessage)
			})
		})

		When("the --help flag is specified", func() {
			It("succeeds and prints help", func() {
				session := helpers.CF("create-service", "--help")
				Eventually(session).Should(Exit(0))
				Expect(session.Out).To(matchHelpMessage)
			})
		})

		When("no arguments are provided", func() {
			It("displays a warning, the help text, and exits 1", func() {
				session := helpers.CF("create-service")
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Incorrect Usage: the required arguments `SERVICE`, `SERVICE_PLAN` and `SERVICE_INSTANCE` were not provided"))
				Expect(session.Out).To(matchHelpMessage)
			})
		})

		When("unknown flag is passed", func() {
			It("displays a warning, the help text, and exits 1", func() {
				session := helpers.CF("create-service", "-u")
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Incorrect Usage: unknown flag `u"))
				Expect(session.Out).To(matchHelpMessage)
			})
		})

		When("a flag is passed with no argument", func() {
			It("displays a warning, the help text, and exits 1", func() {
				session := helpers.CF("create-service", "-c")
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Incorrect Usage: expected argument for flag `-c'"))
				Expect(session.Out).To(matchHelpMessage)
			})
		})

	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "create-service", "foo", "foo", "foo")
		})
	})

	Context("targeting a space", func() {
		var (
			userName  string
			orgName   string
			spaceName string
		)
		assertCreateMessage := func(session *Session) {
			Eventually(session).Should(Say("Creating service instance %s in org %s / space %s as %s...",
				"my-service", orgName, spaceName, userName))
		}

		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()
			helpers.SetupCF(orgName, spaceName)
			userName, _ = helpers.GetCredentials()
		})

		When("there are two offerings with the same name from different brokers", func() {
			var (
				serviceOffering string
				servicePlan     string
				broker1         *servicebrokerstub.ServiceBrokerStub
				broker2         *servicebrokerstub.ServiceBrokerStub
			)

			BeforeEach(func() {
				broker1 = servicebrokerstub.EnableServiceAccess()
				serviceOffering = broker1.FirstServiceOfferingName()
				servicePlan = broker1.FirstServicePlanName()
				broker2 = servicebrokerstub.New()
				broker2.Services[0].Name = serviceOffering
				broker2.Services[0].Plans[0].Name = servicePlan
				broker2.EnableServiceAccess()
			})

			AfterEach(func() {
				helpers.QuickDeleteOrg(orgName)
				broker1.Forget()
				broker2.Forget()
			})

			It("displays an error message prompting to disambiguate", func() {
				session := helpers.CF("create-service", serviceOffering, servicePlan, "my-service")
				assertCreateMessage(session)
				Eventually(session.Err).Should(Say("Service plan '%s' is provided by multiple service offerings. Service offering '%s' is provided by multiple service brokers. Specify a broker name by using the '-b' flag.", servicePlan, serviceOffering))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})
		When("there are no plans matching", func() {
			var (
				serviceOffering string
				broker1         *servicebrokerstub.ServiceBrokerStub
			)

			BeforeEach(func() {
				broker1 = servicebrokerstub.EnableServiceAccess()
				serviceOffering = broker1.FirstServiceOfferingName()
			})

			AfterEach(func() {
				helpers.QuickDeleteOrg(orgName)
				broker1.Forget()
			})
			It("displays an error message", func() {
				session := helpers.CF("create-service", serviceOffering, "another-service-plan", "my-service", "-b", broker1.Name)
				assertCreateMessage(session)
				Eventually(session.Err).Should(Say("The plan '%s' could not be found for service offering '%s' and broker '%s'.", "another-service-plan", serviceOffering, broker1.Name))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})

	})
})
