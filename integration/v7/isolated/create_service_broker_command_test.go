package isolated

import (
	"io"

	"code.cloudfoundry.org/cli/integration/helpers/servicebrokerstub"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("create-service-broker command", func() {
	var brokerName string

	BeforeEach(func() {
		helpers.SkipIfV7AndVersionLessThan(ccversion.MinVersionCreateServiceBrokerV3)

		brokerName = helpers.NewServiceBrokerName()

		helpers.LoginCF()
	})

	Describe("help", func() {
		When("--help flag is set", func() {
			It("displays command usage to output", func() {
				session := helpers.CF("create-service-broker", "--help")
				Eventually(session).Should(Exit(0))

				expectToRenderCreateServiceBrokerHelp(session)
			})
		})
	})

	When("not logged in", func() {
		BeforeEach(func() {
			helpers.LogoutCF()
		})

		It("displays an informative error that the user must be logged in", func() {
			session := helpers.CF("create-service-broker", brokerName, "user", "pass", "http://example.com")
			Eventually(session).Should(Say("FAILED"))
			Eventually(session.Err).Should(Say("Not logged in. Use 'cf login' or 'cf login --sso' to log in."))
			Eventually(session).Should(Exit(1))
		})
	})

	When("logged in", func() {
		When("all arguments are provided", func() {
			When("no org or space is targeted", func() {
				var (
					username  string
					orgName   string
					spaceName string
					broker    *servicebrokerstub.ServiceBrokerStub
				)

				BeforeEach(func() {
					username, _ = helpers.GetCredentials()
					orgName = helpers.NewOrgName()
					spaceName = helpers.NewSpaceName()
					helpers.SetupCF(orgName, spaceName)
					broker = servicebrokerstub.Create()
					helpers.ClearTarget()
				})

				AfterEach(func() {
					Eventually(helpers.CF("delete-service-broker", brokerName, "-f")).Should(Exit(0))
					helpers.QuickDeleteOrg(orgName)
					broker.Forget()
				})

				It("registers the broker and service offerings and plans are available", func() {
					session := helpers.CF("create-service-broker", brokerName, broker.Username, broker.Password, broker.URL)
					Eventually(session).Should(Say("Creating service broker %s as %s...", brokerName, username))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))

					session = helpers.CF("service-access", "-b", brokerName)
					Eventually(session).Should(Say(broker.FirstServiceOfferingName()))
					Eventually(session).Should(Say(broker.FirstServicePlanName()))

					session = helpers.CF("service-brokers")
					Eventually(session).Should(Say(brokerName))
				})
			})

			When("the --space-scoped flag is passed", func() {
				BeforeEach(func() {
					helpers.SkipIfV7AndVersionLessThan(ccversion.MinVersionCreateSpaceScopedServiceBrokerV3)
				})

				When("no org or space is targeted", func() {
					BeforeEach(func() {
						helpers.ClearTarget()
					})

					It("displays an informative error that a space must be targeted", func() {
						session := helpers.CF("create-service-broker", "space-scoped-broker", "username", "password", "http://example.com", "--space-scoped")
						Eventually(session).Should(Say("FAILED"))
						Eventually(session.Err).Should(Say("No org targeted, use 'cf target -o ORG' to target an org."))
						Eventually(session).Should(Exit(1))
					})
				})

				When("both org and space are targeted", func() {
					var (
						username  string
						orgName   string
						spaceName string
						broker    *servicebrokerstub.ServiceBrokerStub
					)

					BeforeEach(func() {
						username, _ = helpers.GetCredentials()
						orgName = helpers.NewOrgName()
						spaceName = helpers.NewSpaceName()
						helpers.SetupCF(orgName, spaceName)
						broker = servicebrokerstub.Create()
					})

					AfterEach(func() {
						Eventually(helpers.CF("delete-service-broker", brokerName, "-f")).Should(Exit(0))
						helpers.QuickDeleteOrg(orgName)
						broker.Forget()
					})

					It("registers the broker and exposes its services only to the targeted space", func() {
						session := helpers.CF("create-service-broker", brokerName, broker.Username, broker.Password, broker.URL, "--space-scoped")
						Eventually(session).Should(Say(
							"Creating service broker %s in org %s / space %s as %s...", brokerName, orgName, spaceName, username))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Exit(0))

						session = helpers.CF("service-brokers")
						Eventually(session).Should(Say(brokerName))

						Eventually(func() io.Reader {
							session := helpers.CF("marketplace")
							Eventually(session).Should(Exit(0))

							return session.Out
						}).Should(Say(broker.FirstServicePlanName()))

						helpers.TargetOrgAndSpace(ReadOnlyOrg, ReadOnlySpace)
						session = helpers.CF("marketplace")
						Eventually(session).ShouldNot(Say(broker.FirstServicePlanName()))
					})
				})
			})

			When("the broker already exists", func() {
				var (
					org        string
					cfUsername string
					broker     *servicebrokerstub.ServiceBrokerStub
					newBroker  *servicebrokerstub.ServiceBrokerStub
				)

				BeforeEach(func() {
					org = helpers.SetupCFWithGeneratedOrgAndSpaceNames()
					cfUsername, _ = helpers.GetCredentials()
					broker = servicebrokerstub.Register()
					newBroker = servicebrokerstub.Create()
				})

				AfterEach(func() {
					broker.Forget()
					newBroker.Forget()
					helpers.QuickDeleteOrg(org)
				})

				It("fails", func() {
					session := helpers.CF("create-service-broker", broker.Name, newBroker.Username, newBroker.Password, newBroker.URL)
					Eventually(session).Should(Exit(1), "expected duplicate create-service-broker to fail")

					Expect(session.Out).To(SatisfyAll(
						Say(`Creating service broker %s as %s...\n`, broker.Name, cfUsername),
						Say(`FAILED\n`),
					))
					Expect(session.Err).To(Say("Name must be unique"))
				})

				When("the --update-if-exists flag is passed", func() {
					It("updates the existing broker", func() {
						session := helpers.CF("create-service-broker", broker.Name, newBroker.Username, newBroker.Password, newBroker.URL, "--update-if-exists")
						Eventually(session).Should(Exit(0))

						Expect(session.Out).To(SatisfyAll(
							Say("Updating service broker %s as %s...", broker.Name, cfUsername),
							Say("OK"),
						))

						By("checking the URL has been updated")
						session = helpers.CF("service-brokers")
						Eventually(session.Out).Should(Say("%s[[:space:]]+%s", broker.Name, newBroker.URL))
					})
				})
			})
		})
	})

	When("the broker URL is invalid", func() {
		BeforeEach(func() {
			// TODO: replace skip with versioned skip when
			// https://www.pivotaltracker.com/story/show/166215494 is resolved.
			helpers.SkipIfV7()
		})

		It("displays a relevant error", func() {
			session := helpers.CF("create-service-broker", brokerName, "user", "pass", "not-a-valid-url")
			Eventually(session).Should(Say("FAILED"))
			Eventually(session.Err).Should(Say("not-a-valid-url is not a valid URL"))
			Eventually(session).Should(Exit(1))
		})
	})

	When("no arguments are provided", func() {
		It("displays an error, naming each of the missing args and the help text", func() {
			session := helpers.CF("create-service-broker")
			Eventually(session).Should(Exit(1))

			expectToRenderCreateServiceBrokerHelp(session)
		})
	})
})

func expectToRenderCreateServiceBrokerHelp(s *Session) {
	Expect(s).To(SatisfyAll(
		Say(`NAME:`),
		Say(`\s+create-service-broker - Create a service broker`),

		Say(`USAGE:`),
		Say(`\s+cf create-service-broker SERVICE_BROKER USERNAME PASSWORD URL \[--space-scoped\]`),
		Say(`\s+cf create-service-broker SERVICE_BROKER USERNAME URL \[--space-scoped\]`),

		Say(`WARNING:`),
		Say(`\s+Providing your password as a command line option is highly discouraged`),
		Say(`\s+Your password may be visible to others and may be recorded in your shell history`),

		Say(`ALIAS:`),
		Say(`\s+csb`),

		Say(`OPTIONS:`),
		Say(`\s+--space-scoped\s+Make the broker's service plans only visible within the targeted space`),
		Say(`\s+--update-if-exists\s+If the broker already exists, update it rather than failing. Ignores --space-scoped.`),

		Say(`ENVIRONMENT:`),
		Say(`\s+CF_BROKER_PASSWORD=password\s+Password associated with user. Overridden if PASSWORD argument is provided`),

		Say(`SEE ALSO:`),
		Say(`\s+enable-service-access, service-brokers, target`),
	))
}
