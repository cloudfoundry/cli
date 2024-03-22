package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/integration/helpers/servicebrokerstub"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("enable service access command", func() {
	Describe("help", func() {
		matchHelpMessage := SatisfyAll(
			Say("NAME:"),
			Say("\\s+enable-service-access - Enable access to a service offering or service plan for one or all orgs"),
			Say("USAGE:"),
			Say("\\s+cf enable-service-access SERVICE_OFFERING \\[-b BROKER\\] \\[-p PLAN\\] \\[-o ORG\\]"),
			Say("OPTIONS:"),
			Say("\\s+\\-b\\s+Enable access to a service offering from a particular service broker. Required when service offering name is ambiguous"),
			Say("\\s+\\-o\\s+Enable access for a specified organization"),
			Say("\\s+\\-p\\s+Enable access to a specified service plan"),
			Say("SEE ALSO:"),
			Say("\\s+disable-service-access, marketplace, service-access, service-brokers"),
		)

		When("--help flag is set", func() {
			It("displays command usage to output", func() {
				session := helpers.CF("enable-service-access", "--help")
				Eventually(session).Should(Exit(0))
				Expect(session.Out).To(matchHelpMessage)
			})
		})

		When("no service argument was provided", func() {
			It("displays a warning, the help text, and exits 1", func() {
				session := helpers.CF("enable-service-access")
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Incorrect Usage: the required argument `SERVICE_OFFERING` was not provided"))
				Expect(session.Out).To(matchHelpMessage)
			})
		})

		When("two services arguments are provided", func() {
			It("displays an error, and exits 1", func() {
				session := helpers.CF("enable-service-access", "a-service", "another-service")
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say(`Incorrect Usage: unexpected argument "another-service"`))
				Expect(session.Out).To(Say("FAILED"))
				Expect(session.Out).To(matchHelpMessage)
			})
		})
	})

	When("logged in", func() {
		var username string
		BeforeEach(func() {
			username, _ = helpers.GetCredentials()
			helpers.LoginCF()
		})

		Context("the service does not exist", func() {
			It("displays FAILED, an informative error message, and exits 1", func() {
				session := helpers.CF("enable-service-access", "some-service")
				Eventually(session).Should(Exit(1))
				Expect(session).To(Say("Enabling access to all plans of service offering some-service for all orgs as %s\\.\\.\\.", username))
				Expect(session).To(Say("FAILED"))
				Expect(session.Err).To(Say("Service offering 'some-service' not found"))
			})
		})

		Context("service offerings exist", func() {
			var (
				orgName         string
				spaceName       string
				serviceOffering string
				servicePlan     string
				broker          *servicebrokerstub.ServiceBrokerStub
				secondBroker    *servicebrokerstub.ServiceBrokerStub
			)

			BeforeEach(func() {
				orgName = helpers.NewOrgName()
				spaceName = helpers.NewSpaceName()
				helpers.SetupCF(orgName, spaceName)

				broker = servicebrokerstub.New().WithPlans(2).Create().Register()
				serviceOffering = broker.FirstServiceOfferingName()
				servicePlan = broker.FirstServicePlanName()

				session := helpers.CF("service-access", "-e", serviceOffering, "-b", broker.Name)
				Eventually(session).Should(Exit(0))
				Expect(session).To(Say("%s\\s+%s\\s+none",
					serviceOffering,
					servicePlan,
				))
			})

			AfterEach(func() {
				helpers.QuickDeleteOrg(orgName)
				broker.Forget()
			})

			When("service offering name provided", func() {
				It("makes all the plans public", func() {
					session := helpers.CF("enable-service-access", serviceOffering)
					Eventually(session).Should(Exit(0))
					Expect(session).To(Say("Enabling access to all plans of service offering %s for all orgs as %s...", serviceOffering, username))
					Expect(session).To(Say("OK"))

					session = helpers.CF("service-access", "-e", serviceOffering)
					Eventually(session).Should(Exit(0))
					Expect(session).To(Say("broker:\\s+%s", broker.Name))
					Expect(session).To(Say("%s\\s+%s\\s+all",
						serviceOffering,
						servicePlan,
					))
				})
			})

			When("service offering name, plan name and org provided", func() {
				It("makes the plan public", func() {
					session := helpers.CF("enable-service-access", serviceOffering, "-p", servicePlan, "-o", orgName, "-v")
					Eventually(session).Should(Exit(0))
					Expect(session).To(Say("Enabling access to plan %s of service offering %s for org %s as %s...", servicePlan, serviceOffering, orgName, username))
					Expect(session).To(Say("OK"))

					session = helpers.CF("service-access", "-e", serviceOffering)
					Eventually(session).Should(Exit(0))
					Expect(session).To(Say("broker:\\s+%s", broker.Name))
					Expect(session).To(Say("%s\\s+%s\\s+%s\\s+%s",
						serviceOffering,
						servicePlan,
						"limited",
						orgName,
					))
				})
			})

			When("two services with the same name are registered", func() {
				BeforeEach(func() {
					secondBroker = servicebrokerstub.New()
					secondBroker.Services[0].Name = serviceOffering
					secondBroker.Services[0].Plans[0].Name = servicePlan
					secondBroker.Create().Register()
				})

				AfterEach(func() {
					secondBroker.Forget()
				})

				When("a serviceOffering name and broker name are provided", func() {
					It("displays an informative message, exits 0, and enables access to the serviceOffering", func() {
						session := helpers.CF("enable-service-access", serviceOffering, "-b", secondBroker.Name)
						Eventually(session).Should(Exit(0))
						Expect(session).To(Say("Enabling access to all plans of service offering %s from broker %s for all orgs as %s...", serviceOffering, secondBroker.Name, username))
						Expect(session).To(Say("OK"))

						session = helpers.CF("service-access", "-b", secondBroker.Name)
						Eventually(session).Should(Exit(0))
						Expect(session).To(Say("broker:\\s+%s", secondBroker.Name))
						Expect(session).To(Say("%s\\s+%s\\s+all",
							serviceOffering,
							servicePlan,
						))
					})
				})
			})

			Context("when access is already globally enabled", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("enable-service-access", serviceOffering)).Should(Exit(0))
				})

				When("when we try to enable access for an org", func() {
					It("should still be enabled only globally", func() {
						session := helpers.CF("enable-service-access", serviceOffering, "-o", orgName)
						Eventually(session).Should(Exit(0))
						Expect(session).To(Say("Did not update plan %s as it already has visibility all\\.", broker.Services[0].Plans[0].Name))
						Expect(session).To(Say("Did not update plan %s as it already has visibility all\\.", broker.Services[0].Plans[1].Name))
						Expect(session).To(Say("OK"))

						session = helpers.CF("service-access", "-e", serviceOffering)
						Eventually(session).Should(Exit(0))
						Expect(session).To(Say("broker:\\s+%s", broker.Name))
						Expect(session).To(Say("%s\\s+%s\\s+all",
							serviceOffering,
							servicePlan,
						))
						Expect(string(session.Out.Contents())).NotTo(ContainSubstring(orgName))
					})
				})

				When("when we try to enable access for an org for a plan", func() {
					It("should still be enabled only globally", func() {
						session := helpers.CF("enable-service-access", serviceOffering, "-o", orgName, "-p", servicePlan)
						Eventually(session).Should(Exit(0))
						Expect(session).To(Say("Did not update plan %s as it already has visibility all\\.", servicePlan))
						Expect(session).To(Say("OK"))

						session = helpers.CF("service-access", "-e", serviceOffering)
						Eventually(session).Should(Exit(0))
						Expect(session).To(Say("broker:\\s+%s", broker.Name))
						Expect(session).To(Say("%s\\s+%s\\s+all",
							serviceOffering,
							servicePlan,
						))
						Expect(string(session.Out.Contents())).NotTo(ContainSubstring(orgName))
					})
				})
			})
		})
	})

	Context("not logged in", func() {
		BeforeEach(func() {
			helpers.LogoutCF()
		})

		It("displays FAILED, an informative error message, and exits 1", func() {
			session := helpers.CF("enable-service-access", "does-not-matter")
			Eventually(session).Should(Exit(1))
			Expect(session).To(Say("FAILED"))
			Expect(session.Err).To(Say("Not logged in. Use 'cf login' or 'cf login --sso' to log in."))
		})
	})

})
