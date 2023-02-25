package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/integration/helpers/servicebrokerstub"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("rename-service command", func() {
	Describe("help", func() {
		expectHelpMessage := func(session *Session) {
			Expect(session).To(SatisfyAll(
				Say(`NAME:\n`),
				Say(`rename-service - Rename a service instance\n`),
				Say(`\n`),
				Say(`USAGE:\n`),
				Say(`cf rename-service SERVICE_INSTANCE NEW_SERVICE_INSTANCE\n`),
				Say(`\n`),
				Say(`SEE ALSO:\n`),
				Say(`services, update-service\n`),
			))
		}

		When("the --help flag is set", func() {
			It("displays command usage to output", func() {
				session := helpers.CF("rename-service", "--help")
				Eventually(session).Should(Exit(0))
				expectHelpMessage(session)
			})
		})

		When("no args are passed", func() {
			It("displays an error message with help text and exits 1", func() {
				session := helpers.CF("rename-service")
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Incorrect Usage: the required arguments `SERVICE_INSTANCE` and `NEW_SERVICE_INSTANCE` were not provided"))
				expectHelpMessage(session)
			})
		})

		When("one arg is passed", func() {
			It("displays an error message with help text and exits 1", func() {
				session := helpers.CF("rename-service", "lala")
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Incorrect Usage: the required argument `NEW_SERVICE_INSTANCE` was not provided"))
				expectHelpMessage(session)
			})
		})

		When("more than required number of args are passed", func() {
			It("displays an error message with help text and exits 1", func() {
				session := helpers.CF("rename-service", "lala", "papa", "mama")
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say(`Incorrect Usage: unexpected argument "mama"`))
				expectHelpMessage(session)
			})
		})

		When("an invalid flag is passed", func() {
			It("displays an error message with help text and exits 1", func() {
				session := helpers.CF("rename-service", "--unicorn-mode")
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Incorrect Usage: unknown flag `unicorn-mode'"))
				expectHelpMessage(session)
			})
		})
	})

	When("environment is not set up", func() {
		It("displays an error and exits 1", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "rename-service", "foo", "bar")
		})
	})

	When("logged in and targeting a space", func() {
		var (
			currentName string
			newName     string
			orgName     string
			spaceName   string
			username    string
		)

		BeforeEach(func() {
			currentName = helpers.NewServiceInstanceName()
			newName = helpers.NewServiceInstanceName()

			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()
			helpers.SetupCF(orgName, spaceName)

			username, _ = helpers.GetCredentials()
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		When("the service instance does not exist", func() {
			It("fails with an appropriate error", func() {
				session := helpers.CF("rename-service", currentName, newName)

				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Service instance '%s' not found", currentName))
			})
		})

		testRename := func() {
			It("renames the service instance", func() {
				originalGUID := helpers.ServiceInstanceGUID(currentName)

				session := helpers.CF("rename-service", currentName, newName)

				Eventually(session).Should(Exit(0))
				Expect(session.Out).To(SatisfyAll(
					Say(`Renaming service %s to %s in org %s / space %s as %s...\n`, currentName, newName, orgName, spaceName, username),
					Say(`OK\n`),
				))
				Expect(session.Err.Contents()).To(BeEmpty())

				Expect(helpers.ServiceInstanceGUID(newName)).To(Equal(originalGUID))
			})

			When("the service instance name is taken", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("create-user-provided-service", newName)).Should(Exit(0))
				})

				It("fails and explains why", func() {
					session := helpers.CF("rename-service", currentName, newName)

					Eventually(session).Should(Exit(1))
					Expect(session.Out).To(SatisfyAll(
						Say(`Renaming service %s to %s in org %s / space %s as %s...\n`, currentName, newName, orgName, spaceName, username),
						Say(`FAILED\n`),
					))
					Expect(session.Err).To(Say(`The service instance name is taken: %s\.?\n`, newName))
				})
			})
		}

		Context("user-provided service instance", func() {
			BeforeEach(func() {
				Eventually(helpers.CF("create-user-provided-service", currentName)).Should(Exit(0))
			})

			testRename()
		})

		Context("managed service instance", func() {
			var broker *servicebrokerstub.ServiceBrokerStub

			BeforeEach(func() {
				broker = servicebrokerstub.EnableServiceAccess()

				helpers.CreateManagedServiceInstance(broker.FirstServiceOfferingName(), broker.FirstServicePlanName(), currentName)
			})

			AfterEach(func() {
				broker.Forget()
			})

			testRename()
		})
	})
})
