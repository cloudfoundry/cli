package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/integration/helpers/servicebrokerstub"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("disable service access command", func() {
	Describe("help", func() {
		matchHelpMessage := SatisfyAll(
			Say("NAME:"),
			Say("\\s+disable-service-access - Disable access to a service offering or service plan for one or all orgs"),
			Say("USAGE:"),
			Say("\\s+cf disable-service-access SERVICE_OFFERING \\[-b BROKER\\] \\[-p PLAN\\] \\[-o ORG\\]"),
			Say("OPTIONS:"),
			Say("\\s+\\-b\\s+Disable access to a service offering from a particular service broker. Required when service offering name is ambiguous"),
			Say("\\s+\\-o\\s+Disable access for a specified organization"),
			Say("\\s+\\-p\\s+Disable access to a specified service plan"),
			Say("SEE ALSO:"),
			Say("\\s+enable-service-access, marketplace, service-access, service-brokers"),
		)

		When("--help flag is set", func() {
			It("displays command usage to output", func() {
				session := helpers.CF("disable-service-access", "--help")
				Eventually(session).Should(Exit(0))
				Expect(session.Out).To(matchHelpMessage)
			})
		})

		When("no service argument was provided", func() {
			It("displays a warning, the help text, and exits 1", func() {
				session := helpers.CF("disable-service-access")
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Incorrect Usage: the required argument `SERVICE_OFFERING` was not provided"))
				Expect(session.Out).To(matchHelpMessage)
			})
		})

		When("an extra argument is provided", func() {
			It("displays an error, and exits 1", func() {
				session := helpers.CF("disable-service-access", "a-service", "extra-arg")
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say(`Incorrect Usage: unexpected argument "extra-arg"`))
				Expect(session.Out).To(Say("FAILED"))
				Expect(session.Out).To(matchHelpMessage)
			})
		})
	})

	Context("not logged in", func() {
		BeforeEach(func() {
			helpers.LogoutCF()
		})

		It("displays FAILED, an informative error message, and exits 1", func() {
			session := helpers.CF("disable-service-access", "does-not-matter")
			Eventually(session).Should(Exit(1))
			Expect(session).To(Say("FAILED"))
			Expect(session.Err).To(Say("Not logged in. Use 'cf login' or 'cf login --sso' to log in."))
		})
	})

	Context("logged in", func() {
		var username string

		BeforeEach(func() {
			username, _ = helpers.GetCredentials()
			helpers.LoginCF()
		})

		When("service does not exist", func() {
			It("displays FAILED, an informative error message, and exits 1", func() {
				session := helpers.CF("disable-service-access", "some-service")
				Eventually(session).Should(Exit(1))
				Expect(session).To(Say("Disabling access to all plans of service offering some-service for all orgs as %s\\.\\.\\.", username))
				Expect(session.Err).To(Say("Service offering 'some-service' not found"))
				Expect(session).To(Say("FAILED"))
			})
		})

		Context("a service broker is registered", func() {
			var (
				orgName   string
				spaceName string
				broker    *servicebrokerstub.ServiceBrokerStub
			)

			BeforeEach(func() {
				orgName = helpers.NewOrgName()
				spaceName = helpers.NewSpaceName()
				helpers.SetupCF(orgName, spaceName)

				broker = servicebrokerstub.New().WithPlans(2).Create().Register()
			})

			AfterEach(func() {
				helpers.QuickDeleteOrg(orgName)
				broker.Forget()
			})

			When("plans are public", func() {
				BeforeEach(func() {
					session := helpers.CF("enable-service-access", broker.FirstServiceOfferingName())
					Eventually(session).Should(Exit(0))
				})

				When("a service name is provided", func() {
					It("displays an informative message, exits 0, and disables the service for all orgs", func() {
						session := helpers.CF("disable-service-access", broker.FirstServiceOfferingName())
						Eventually(session).Should(Exit(0))
						Expect(session).To(Say("Disabling access to all plans of service offering %s for all orgs as %s...", broker.FirstServiceOfferingName(), username))
						Expect(session).To(Say("OK"))

						session = helpers.CF("service-access", "-e", broker.FirstServiceOfferingName())
						Eventually(session).Should(Exit(0))
						Expect(session).To(Say("broker:\\s+%s", broker.Name))
						Expect(session).To(Say("%s\\s+%s\\s+none",
							broker.FirstServiceOfferingName(),
							broker.FirstServicePlanName(),
						))
					})
				})

				When("a service name and plan name are provided", func() {
					It("displays an informative message, exits 0, and disables the plan for all orgs", func() {
						session := helpers.CF("disable-service-access", broker.FirstServiceOfferingName(), "-p", broker.FirstServicePlanName())
						Eventually(session).Should(Exit(0))
						Expect(session).To(Say("Disabling access to plan %s of service offering %s for all orgs as %s...", broker.FirstServicePlanName(), broker.FirstServiceOfferingName(), username))
						Expect(session).To(Say("OK"))

						session = helpers.CF("service-access", "-e", broker.FirstServiceOfferingName())
						Eventually(session).Should(Exit(0))

						Expect(string(session.Out.Contents())).To(MatchRegexp(
							`%s\s+%s\s+none`,
							broker.FirstServiceOfferingName(),
							broker.FirstServicePlanName(),
						))

						Expect(string(session.Out.Contents())).To(MatchRegexp(
							`%s\s+%s\s+all`,
							broker.FirstServiceOfferingName(),
							broker.Services[0].Plans[1].Name,
						))
					})
				})

				When("an org name is provided", func() {
					It("fails", func() {
						session := helpers.CF("disable-service-access", broker.FirstServiceOfferingName(), "-o", orgName)
						Eventually(session).Should(Exit(1))
						Expect(session).To(Say("Disabling access to all plans of service offering %s for org %s as %s...", broker.FirstServiceOfferingName(), orgName, username))
						Expect(session).To(Say("FAILED"))
						Expect(session.Err).To(Say("Cannot remove organization level access for public plans\\."))
					})
				})
			})

			When("a plan is enabled for multiple orgs", func() {
				var orgName2 string

				BeforeEach(func() {
					orgName2 = helpers.NewOrgName()
					spaceName2 := helpers.NewSpaceName()
					helpers.CreateOrgAndSpace(orgName2, spaceName2)
					Eventually(helpers.CF("enable-service-access", broker.FirstServiceOfferingName(), "-o", orgName, "-p", broker.FirstServicePlanName())).Should(Exit(0))
					Eventually(helpers.CF("enable-service-access", broker.FirstServiceOfferingName(), "-o", orgName2, "-p", broker.FirstServicePlanName())).Should(Exit(0))
				})

				AfterEach(func() {
					helpers.QuickDeleteOrg(orgName2)
				})

				When("a service name and org is provided", func() {
					It("displays an informative message, and exits 0, and disables the service for the given org", func() {
						session := helpers.CF("disable-service-access", broker.FirstServiceOfferingName(), "-o", orgName)
						Eventually(session).Should(Exit(0))
						Expect(session).To(Say("Disabling access to all plans of service offering %s for org %s as %s...", broker.FirstServiceOfferingName(), orgName, username))
						Expect(session).To(Say("Did not update plan %s as it already has visibility none\\.", broker.Services[0].Plans[1].Name))
						Expect(session).To(Say("OK"))

						session = helpers.CF("service-access", "-e", broker.FirstServiceOfferingName())
						Eventually(session).Should(Exit(0))
						Expect(session).To(Say("broker:\\s+%s", broker.Name))
						Expect(session).To(Say("%s\\s+%s\\s+limited\\s+%s",
							broker.FirstServiceOfferingName(),
							broker.FirstServicePlanName(),
							orgName2,
						))
					})
				})

				When("a service name, plan name and org is provided", func() {
					It("displays an informative message, and exits 0, disables the service for the given org and plan", func() {
						session := helpers.CF("disable-service-access", broker.FirstServiceOfferingName(), "-p", broker.FirstServicePlanName(), "-o", orgName)
						Eventually(session).Should(Exit(0))
						Expect(session).To(Say("Disabling access to plan %s of service offering %s for org %s as %s...", broker.FirstServicePlanName(), broker.FirstServiceOfferingName(), orgName, username))
						Expect(session).To(Say("OK"))

						session = helpers.CF("service-access", "-e", broker.FirstServiceOfferingName())
						Eventually(session).Should(Exit(0))
						Expect(session).To(Say("broker:\\s+%s", broker.Name))
						Expect(session).To(Say("%s\\s+%s\\s+limited\\s+%s",
							broker.FirstServiceOfferingName(),
							broker.FirstServicePlanName(),
							orgName2,
						))
					})
				})
			})

			When("the org does not exist", func() {
				It("displays FAILED, an informative error message, and exits 1", func() {
					session := helpers.CF("disable-service-access", broker.FirstServiceOfferingName(), "-o", "not-a-real-org")
					Eventually(session).Should(Exit(1))
					Expect(session).To(Say("Disabling access to all plans of service offering %s for org not-a-real-org as %s...", broker.FirstServiceOfferingName(), username))
					Expect(session).To(Say("FAILED"))
					Expect(session.Err).To(Say("Organization 'not-a-real-org' not found"))
				})
			})

			When("the plan does not exist", func() {
				It("displays FAILED, an informative error message, and exits 1", func() {
					session := helpers.CF("disable-service-access", broker.FirstServiceOfferingName(), "-p", "plan-does-not-exist")
					Eventually(session).Should(Exit(1))
					Expect(session).To(Say("Disabling access to plan plan-does-not-exist of service offering %s for all orgs as %s...", broker.FirstServiceOfferingName(), username))
					Expect(session).To(Say("FAILED"))
					Expect(session.Err).To(Say("The plan 'plan-does-not-exist' could not be found for service offering '%s'", broker.FirstServiceOfferingName()))
				})
			})

			When("two services with the same name are enabled", func() {
				var secondBroker *servicebrokerstub.ServiceBrokerStub

				BeforeEach(func() {
					secondBroker = servicebrokerstub.New()
					secondBroker.Services[0].Name = broker.FirstServiceOfferingName()
					secondBroker.Create().Register()
					Eventually(helpers.CF("enable-service-access", broker.FirstServiceOfferingName(), "-b", broker.Name)).Should(Exit(0))
					Eventually(helpers.CF("enable-service-access", secondBroker.FirstServiceOfferingName(), "-b", secondBroker.Name)).Should(Exit(0))
				})

				AfterEach(func() {
					secondBroker.Forget()
				})

				When("no broker name is provided", func() {
					It("fails and asks for disambiguation", func() {
						session := helpers.CF("disable-service-access", broker.FirstServiceOfferingName())
						Eventually(session).Should(Exit(1))
						Expect(session).To(Say("Disabling access to all plans of service offering %s for all orgs as %s...", broker.FirstServiceOfferingName(), username))
						Expect(session.Err).To(Say(
							"Service '%s' is provided by multiple service brokers: %s, %s\nSpecify a broker by using the '-b' flag.",
							broker.FirstServiceOfferingName(),
							broker.Name,
							secondBroker.Name,
						))
					})
				})

				When("a broker name is provided", func() {
					It("displays an informative message, exits 0, and disables access to the service", func() {
						session := helpers.CF("disable-service-access", broker.FirstServiceOfferingName(), "-b", secondBroker.Name)
						Eventually(session).Should(Exit(0))
						Expect(session).To(Say("Disabling access to all plans of service offering %s from broker %s for all orgs as %s...", broker.FirstServiceOfferingName(), secondBroker.Name, username))
						Expect(session).To(Say("OK"))

						session = helpers.CF("service-access", "-b", secondBroker.Name)
						Eventually(session).Should(Exit(0))
						Expect(session).To(Say("broker:\\s+%s", secondBroker.Name))
						Expect(session).To(Say("%s\\s+%s\\s+none",
							secondBroker.FirstServiceOfferingName(),
							secondBroker.FirstServicePlanName(),
						))
					})
				})
			})
		})
	})
})
