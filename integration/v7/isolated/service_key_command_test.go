package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers/servicebrokerstub"
	. "github.com/onsi/gomega/gexec"

	. "github.com/onsi/gomega/gbytes"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("service-key command", func() {
	const command = "service-key"

	Describe("help", func() {
		matchHelpMessage := SatisfyAll(
			Say(`NAME:\n`),
			Say(`\s+service-key - Show service key info\n`),
			Say(`\n`),
			Say(`USAGE:\n`),
			Say(`\s+cf service-key SERVICE_INSTANCE SERVICE_KEY\n`),
			Say(`\n`),
			Say(`EXAMPLES:\n`),
			Say(`\s+cf service-key mydb mykey\n`),
			Say(`\n`),
			Say(`OPTIONS:\n`),
			Say(`\s+--guid\s+Retrieve and display the given service-key's guid. All other output is suppressed\.\n`),
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
				Expect(session.Err).To(Say("Incorrect Usage: the required arguments `SERVICE_INSTANCE` and `SERVICE_KEY` were not provided"))
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
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, command, "foo", "bar")
		})
	})

	When("targeting a space", func() {
		var (
			org      string
			space    string
			username string
		)

		BeforeEach(func() {
			org = helpers.NewOrgName()
			space = helpers.NewSpaceName()
			helpers.SetupCF(org, space)

			username, _ = helpers.GetCredentials()
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(org)
		})

		When("service instance does not exit", func() {
			It("reports a helpful error", func() {
				session := helpers.CF(command, "no-such-instance", "no-such-key")
				Eventually(session).Should(Exit(1))

				Expect(session.Out).To(SatisfyAll(
					Say(`Getting key no-such-key for service instance no-such-instance as %s\.\.\.\n`, username),
					Say(`FAILED\n`),
				))

				Expect(session.Err).To(Say(`Service instance 'no-such-instance' not found`))
			})
		})

		When("service instance exists", func() {
			var (
				broker          *servicebrokerstub.ServiceBrokerStub
				serviceInstance string
			)

			BeforeEach(func() {
				broker = servicebrokerstub.EnableServiceAccess()
				serviceInstance = helpers.NewServiceInstanceName()
				helpers.CreateManagedServiceInstance(broker.FirstServiceOfferingName(), broker.FirstServicePlanName(), serviceInstance)
			})

			AfterEach(func() {
				broker.Forget()
			})

			When("key does not exist", func() {
				It("reports a helpful error", func() {
					session := helpers.CF(command, serviceInstance, "no-such-key")
					Eventually(session).Should(Exit(1))

					Expect(session.Out).To(SatisfyAll(
						Say(`Getting key no-such-key for service instance %s as %s\.\.\.\n`, serviceInstance, username),
						Say(`FAILED\n`),
					))

					Expect(session.Err).To(Say(`No service key no-such-key found for service instance %s`, serviceInstance))
				})
			})

			When("keys exists", func() {
				var keyName string

				BeforeEach(func() {
					keyName = helpers.PrefixedRandomName("key")
					Eventually(helpers.CF("create-service-key", serviceInstance, keyName, "-c", `{"foo":"bar"}`)).Should(Exit(0))
				})

				It("prints the details", func() {
					session := helpers.CF(command, serviceInstance, keyName)
					Eventually(session).Should(Exit(0))

					Expect(session.Out).To(SatisfyAll(
						Say(`Getting key %s for service instance %s as %s\.\.\.\n`, keyName, serviceInstance, username),
						Say(`\n`),
						Say(`\{\n`),
						Say(`  "password": "%s",\n`, broker.Password),
						Say(`  "username": "%s"\n`, broker.Username),
						Say(`\}\n`),
					))

					Expect(string(session.Err.Contents())).To(BeEmpty())
				})

				When("the --guid option is given", func() {
					It("prints just the GUID", func() {
						session := helpers.CF(command, serviceInstance, keyName, "--guid")
						Eventually(session).Should(Exit(0))

						Expect(string(session.Err.Contents())).To(BeEmpty())
						Eventually(session).Should(Say(`[\da-f]{8}-[\da-f]{4}-[\da-f]{4}-[\da-f]{4}-[\da-f]{12}\n`))
					})
				})
			})
		})
	})
})
