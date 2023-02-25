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

var _ = Describe("delete-service-key command", func() {
	const command = "delete-service-key"

	Describe("help", func() {
		matchHelpMessage := SatisfyAll(
			Say(`NAME:\n`),
			Say(`\s+delete-service-key - Delete a service key\n`),
			Say(`\n`),
			Say(`USAGE:\n`),
			Say(`\s+cf delete-service-key SERVICE_INSTANCE SERVICE_KEY \[-f\] \[--wait\]\n`),
			Say(`\n`),
			Say(`EXAMPLES:\n`),
			Say(`\s+cf delete-service-key mydb mykey\n`),
			Say(`\n`),
			Say(`ALIAS:\n`),
			Say(`\s+dsk\n`),
			Say(`\n`),
			Say(`OPTIONS:\n`),
			Say(`\s+-f\s+Force deletion without confirmation\n`),
			Say(`\s+--wait, -w\s+Wait for the operation to complete\n`),
			Say(`\n`),
			Say(`SEE ALSO:\n`),
			Say(`\s+service-keys\n`),
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

		When("insufficient arguments are provided", func() {
			It("displays a warning, the help text, and exits 1", func() {
				session := helpers.CF(command, "instance")
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Incorrect Usage: the required argument `SERVICE_KEY` was not provided"))
				Expect(session.Out).To(matchHelpMessage)
			})
		})

		When("superfluous arguments are provided", func() {
			It("displays a warning, the help text, and exits 1", func() {
				session := helpers.CF(command, "instance", "key", "superfluous")
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say(`Incorrect Usage: unexpected argument "superfluous"`))
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

	When("the environment is correct", func() {
		var (
			orgName             string
			spaceName           string
			username            string
			broker              *servicebrokerstub.ServiceBrokerStub
			serviceInstanceName string
			serviceKeyName      string
		)

		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()
			helpers.SetupCF(orgName, spaceName)

			username, _ = helpers.GetCredentials()

			broker = servicebrokerstub.New().WithRouteService().EnableServiceAccess()
			serviceInstanceName = helpers.NewServiceInstanceName()
			serviceKeyName = helpers.PrefixedRandomName("KEY")
			helpers.CreateManagedServiceInstance(broker.FirstServiceOfferingName(), broker.FirstServicePlanName(), serviceInstanceName)

			Eventually(helpers.CF("create-service-key", serviceInstanceName, serviceKeyName)).Should(Exit(0))
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
			broker.Forget()
		})

		It("deletes the service key", func() {
			session := helpers.CF(command, "-f", serviceInstanceName, serviceKeyName)
			Eventually(session).Should(Exit(0))

			Expect(session.Out).To(SatisfyAll(
				Say(`Deleting key %s for service instance %s as %s\.\.\.\n`, serviceKeyName, serviceInstanceName, username),
				Say(`OK\n`),
			))
			Expect(string(session.Err.Contents())).To(BeEmpty())

			Expect(helpers.CF("service-keys", serviceInstanceName).Wait().Out).NotTo(Say(serviceKeyName))
		})

		When("broker response is asynchronous", func() {
			BeforeEach(func() {
				broker.WithAsyncDelay(time.Second).Configure()
			})

			It("starts to delete the service key", func() {
				session := helpers.CF(command, "-f", serviceInstanceName, serviceKeyName)
				Eventually(session).Should(Exit(0))

				Expect(session.Out).To(SatisfyAll(
					Say(`Deleting key %s for service instance %s as %s\.\.\.\n`, serviceKeyName, serviceInstanceName, username),
					Say(`OK\n`),
					Say(`\n`),
					Say(`Delete in progress\.\n`),
				))
				Expect(string(session.Err.Contents())).To(BeEmpty())

				Expect(helpers.CF("service-keys", serviceInstanceName).Wait().Out).To(Say(serviceKeyName))
			})

			It("can wait for the service key to be deleted", func() {
				session := helpers.CF(command, "-f", "--wait", serviceInstanceName, serviceKeyName)
				Eventually(session).Should(Exit(0))

				Expect(session.Out).To(SatisfyAll(
					Say(`Deleting key %s for service instance %s as %s\.\.\.\n`, serviceKeyName, serviceInstanceName, username),
					Say(`Waiting for the operation to complete\.+\n`),
					Say(`\n`),
					Say(`OK\n`),
				))
				Expect(string(session.Err.Contents())).To(BeEmpty())

				Expect(helpers.CF("service-keys", serviceInstanceName).Wait().Out).NotTo(Say(serviceKeyName))
			})
		})

		When("key does not exist", func() {
			It("succeeds with a message", func() {
				nonExistentKey := helpers.PrefixedRandomName("NO-SUCH-KEY")
				session := helpers.CF(command, "-f", serviceInstanceName, nonExistentKey)
				Eventually(session).Should(Exit(0))

				Expect(session.Out).To(SatisfyAll(
					Say(`Deleting key %s for service instance %s as %s\.\.\.\n`, nonExistentKey, serviceInstanceName, username),
					Say(`\n`),
					Say(`Service key %s does not exist for service instance %s\.\n`, nonExistentKey, serviceInstanceName),
					Say(`OK\n`),
				))
				Expect(string(session.Err.Contents())).To(BeEmpty())
			})
		})
	})
})
