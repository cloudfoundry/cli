package isolated

import (
	"fmt"

	"code.cloudfoundry.org/cli/integration/assets/hydrabroker/config"
	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/integration/helpers/servicebrokerstub"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("marketplace command", func() {
	Describe("help", func() {
		matchMarketplaceHelpMessage := SatisfyAll(
			Say(`NAME:`),
			Say(`marketplace - List available offerings in the marketplace`),
			Say(`USAGE:`),
			Say(`cf marketplace \[-e SERVICE_OFFERING\] \[-b SERVICE_BROKER\] \[--no-plans\]`),
			Say(`ALIAS:`),
			Say(`m`),
			Say(`OPTIONS:`),
			Say(`-e\s+Show plan details for a particular service offering`),
			Say(`--no-plans\s+Hide plan information for service offerings`),
			Say(`--show-unavailable\s+Show plans that are not available for use`),
			Say(`create-service, services`),
		)

		When("the --help flag is set", func() {
			It("displays command usage to output", func() {
				session := helpers.CF("marketplace", "--help")
				Eventually(session).Should(Exit(0))
				Expect(session.Out).To(matchMarketplaceHelpMessage)
			})
		})

		When("more than required number of args are passed", func() {
			It("displays an error message with help text and exits 1", func() {
				session := helpers.CF("marketplace", "lala")
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say(`Incorrect Usage: unexpected argument "lala"`))
				Expect(session.Out).To(matchMarketplaceHelpMessage)
			})
		})
	})

	When("an API target is not set", func() {
		BeforeEach(func() {
			helpers.UnsetAPI()
		})

		It("displays an error message that no API endpoint is set", func() {
			session := helpers.CF("marketplace")
			Eventually(session).Should(Exit(1))

			Expect(session.Out).To(Say("FAILED"))
			Expect(session.Err).To(Say(`No API endpoint set\. Use 'cf login' or 'cf api' to target an endpoint\.`))
		})
	})

	When("service offerings are registered", func() {
		const (
			mixedOfferingIndex = iota
			publicOfferingIndex
			privateOfferingIndex
		)

		const (
			privatePlanIndex = iota
			publicChargedPlanIndex
			org1PlanIndex
			org2PlanIndex
		)

		var (
			session             *Session
			args                []string
			username            string
			org1, org2          string
			space1, space2      string
			privatePlanName     string
			publicPlanName      string
			org1PlanName        string
			org2PlanName        string
			mixedOfferingName   string
			publicOfferingName  string
			privateOfferingName string
			broker              *servicebrokerstub.ServiceBrokerStub
		)

		deactivateOrg1Plan := func() {
			helpers.CreateManagedServiceInstance(
				broker.Services[mixedOfferingIndex].Name,
				broker.Services[mixedOfferingIndex].Plans[org1PlanIndex].Name,
				helpers.NewServiceInstanceName(),
			)

			broker.Services[mixedOfferingIndex].Plans[org1PlanIndex].Name = helpers.NewServiceOfferingName()
			broker.Services[mixedOfferingIndex].Plans[org1PlanIndex].ID = helpers.RandomName()

			broker.Configure().Register()
		}

		JustBeforeEach(func() {
			session = helpers.CF(append([]string{"marketplace"}, args...)...)
			Eventually(session).Should(Exit(0))
		})

		BeforeEach(func() {
			args = nil

			helpers.LoginCF()

			org1 = helpers.NewOrgName()
			org2 = helpers.NewOrgName()
			space1 = helpers.NewSpaceName()
			space2 = helpers.NewSpaceName()

			helpers.CreateOrgAndSpace(org1, space1)
			helpers.CreateOrgAndSpace(org2, space2)

			username, _ = helpers.GetCredentials()

			broker = servicebrokerstub.New().WithServiceOfferings(3).WithPlans(5)

			mixedOfferingName = helpers.PrefixedRandomName("INTEGRATION-OFFERING-MIXED")
			broker.Services[mixedOfferingIndex].Name = mixedOfferingName

			privatePlanName = helpers.PrefixedRandomName("INTEGRATION-PLAN-PRIVATE")
			broker.Services[mixedOfferingIndex].Plans[privatePlanIndex].Name = privatePlanName

			publicPlanName = helpers.PrefixedRandomName("INTEGRATION-PLAN-PUBLIC")
			broker.Services[mixedOfferingIndex].Plans[publicChargedPlanIndex].Name = publicPlanName
			broker.Services[mixedOfferingIndex].Plans[publicChargedPlanIndex].Free = false
			broker.Services[mixedOfferingIndex].Plans[publicChargedPlanIndex].Costs = []config.Cost{
				{
					Amount: map[string]float64{"gbp": 600.00, "usd": 649.00},
					Unit:   "MONTHLY",
				},
				{
					Amount: map[string]float64{"usd": 1.00},
					Unit:   "1GB of messages over 20GB",
				},
			}
			broker.ServiceAccessConfig = append(broker.ServiceAccessConfig, servicebrokerstub.ServiceAccessConfig{
				OfferingName: mixedOfferingName,
				PlanName:     publicPlanName,
			})

			org1PlanName = helpers.PrefixedRandomName("INTEGRATION-PLAN-ORG1")
			broker.Services[mixedOfferingIndex].Plans[org1PlanIndex].Name = org1PlanName
			broker.ServiceAccessConfig = append(broker.ServiceAccessConfig, servicebrokerstub.ServiceAccessConfig{
				OfferingName: mixedOfferingName,
				PlanName:     org1PlanName,
				OrgName:      org1,
			})

			org2PlanName = helpers.PrefixedRandomName("INTEGRATION-PLAN-ORG2")
			broker.Services[mixedOfferingIndex].Plans[org2PlanIndex].Name = org2PlanName
			broker.ServiceAccessConfig = append(broker.ServiceAccessConfig, servicebrokerstub.ServiceAccessConfig{
				OfferingName: mixedOfferingName,
				PlanName:     org2PlanName,
				OrgName:      org2,
			})

			publicOfferingName = helpers.PrefixedRandomName("INTEGRATION-OFFERING-PUBLIC")
			broker.Services[publicOfferingIndex].Name = publicOfferingName
			broker.ServiceAccessConfig = append(broker.ServiceAccessConfig, servicebrokerstub.ServiceAccessConfig{
				OfferingName: publicOfferingName,
			})

			privateOfferingName = helpers.PrefixedRandomName("INTEGRATION-OFFERING-PUBLIC")
			broker.Services[privateOfferingIndex].Name = privateOfferingName

			broker.EnableServiceAccess()

			helpers.TargetOrgAndSpace(org1, space1)
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(org1)
			helpers.QuickDeleteOrg(org2)
			broker.Forget()
		})

		Describe("service offerings table", func() {
			It("shows the available offerings and plans", func() {
				expectedMixedOfferingPlans := fmt.Sprintf("%s, %s", publicPlanName, org1PlanName)
				Expect(session.Out).To(SatisfyAll(
					Say(`Getting all service offerings from marketplace in org %s / space %s as %s\.\.\.\n`, org1, space1, username),
					Say(`\n`),
					Say(`offering\s+plans\s+description\s+broker\n`),
					Say(`%s\s+%s\s+%s\s+%s\n`, mixedOfferingName, expectedMixedOfferingPlans, broker.Services[mixedOfferingIndex].Description, broker.Name),
					Say(`%s\s+%s\s+%s\s+%s\n`, publicOfferingName, broker.Services[publicOfferingIndex].Plans[0].Name, broker.Services[publicOfferingIndex].Description, broker.Name),
					Say(`\n`),
					Say(`TIP: Use 'cf marketplace -e SERVICE_OFFERING' to view descriptions of individual plans of a given service offering\.\n`),
				))

				Expect(session.Out).NotTo(SatisfyAny(
					Say(privateOfferingName),
					Say(privatePlanName),
					Say(org2PlanName),
				))
			})

			When("logged out", func() {
				BeforeEach(func() {
					helpers.LogoutCF()
				})

				AfterEach(func() {
					helpers.LoginCF()
				})

				It("only shows the public plans", func() {
					Expect(session.Out).To(SatisfyAll(
						Say(`Getting all service offerings from marketplace\.\.\.\n`),
						Say(`\n`),
						Say(`offering\s+plans\s+description\s+broker\n`),
						Say(`%s\s+%s\s+%s\s+%s\n`, mixedOfferingName, publicPlanName, broker.Services[mixedOfferingIndex].Description, broker.Name),
						Say(`%s\s+%s\s+%s\s+%s\n`, publicOfferingName, broker.Services[publicOfferingIndex].Plans[0].Name, broker.Services[publicOfferingIndex].Description, broker.Name),
						Say(`\n`),
						Say(`TIP: Use 'cf marketplace -e SERVICE_OFFERING' to view descriptions of individual plans of a given service offering\.\n`),
					))

					Expect(session.Out).NotTo(SatisfyAny(
						Say(privateOfferingName),
						Say(privatePlanName),
						Say(org1PlanName),
						Say(org2PlanName),
					))
				})
			})

			When("filtering by service broker name", func() {
				var secondServiceBroker *servicebrokerstub.ServiceBrokerStub

				BeforeEach(func() {
					secondServiceBroker = servicebrokerstub.New().WithServiceOfferings(1).WithPlans(1).EnableServiceAccess()

					args = append(args, "-b", secondServiceBroker.Name)
				})

				AfterEach(func() {
					secondServiceBroker.Forget()
				})

				It("only shows plans for the selected broker", func() {
					Expect(session.Out).To(SatisfyAll(
						Say(`Getting all service offerings from marketplace for service broker %s in org %s / space %s as %s\.\.\.\n`, secondServiceBroker.Name, org1, space1, username),
						Say(`\n`),
						Say(`offering\s+plans\s+description\s+broker\n`),
						Say(`%s\s+%s\s+%s\s+%s\n`, secondServiceBroker.FirstServiceOfferingName(), secondServiceBroker.FirstServicePlanName(), secondServiceBroker.FirstServiceOfferingDescription(), secondServiceBroker.Name),
						Say(`\n`),
						Say(`TIP: Use 'cf marketplace -e SERVICE_OFFERING' to view descriptions of individual plans of a given service offering\.\n`),
					))

					Expect(session.Out).NotTo(Say(broker.Name))
				})
			})

			When("a service plan is unavailable", func() {
				BeforeEach(func() {
					deactivateOrg1Plan()
				})

				It("does not show the unavailable plan", func() {
					Expect(string(session.Out.Contents())).NotTo(ContainSubstring(org1PlanName))
				})

				When("the --show-unavailable flag is specified", func() {
					BeforeEach(func() {
						args = append(args, "--show-unavailable")
					})

					It("shows unavailable plans", func() {
						expectedPlans := fmt.Sprintf("%s, %s", publicPlanName, org1PlanName)
						Expect(session.Out).To(Say(`%s\s+%s\s+%s\s+%s\n`, mixedOfferingName, expectedPlans, broker.Services[mixedOfferingIndex].Description, broker.Name))
					})
				})
			})
		})

		Describe("service plans table", func() {
			BeforeEach(func() {
				args = append(args, "-e", mixedOfferingName)
			})

			It("shows the available plans", func() {
				Expect(session.Out).To(SatisfyAll(
					Say(`Getting service plan information for service offering %s in org %s / space %s as %s\.\.\.\n`, mixedOfferingName, org1, space1, username),
					Say(`\n`),
					Say(`broker: %s\n`, broker.Name),
					Say(`plan\s+description\s+free or paid\s+costs\n`),
					Say(`%s\s+%s\s+%s\s+%s\n`, publicPlanName, broker.Services[mixedOfferingIndex].Plans[publicChargedPlanIndex].Description, "paid", "GBP 600.00/MONTHLY, USD 649.00/MONTHLY, USD 1.00/1GB of messages over 20GB"),
					Say(`%s\s+%s\s+%s\s+\n`, org1PlanName, broker.Services[mixedOfferingIndex].Plans[org1PlanIndex].Description, "free"),
				))

				Expect(session.Out).NotTo(SatisfyAny(
					Say(publicOfferingName),
					Say(privatePlanName),
					Say(org2PlanName),
				))
			})

			When("logged out", func() {
				BeforeEach(func() {
					helpers.LogoutCF()
				})

				AfterEach(func() {
					helpers.LoginCF()
				})

				It("only shows the public plans", func() {
					Expect(session.Out).To(SatisfyAll(
						Say(`Getting service plan information for service offering %s\.\.\.\n`, mixedOfferingName),
						Say(`\n`),
						Say(`broker: %s\n`, broker.Name),
						Say(`plan\s+description\s+free or paid\s+costs\n`),
						Say(`%s\s+%s\s+%s\s+%s\n`, publicPlanName, broker.Services[mixedOfferingIndex].Plans[publicChargedPlanIndex].Description, "paid", "GBP 600.00/MONTHLY, USD 649.00/MONTHLY, USD 1.00/1GB of messages over 20GB"),
					))

					Expect(session.Out).NotTo(SatisfyAny(
						Say(publicOfferingName),
						Say(privatePlanName),
						Say(org1PlanName),
						Say(org2PlanName),
					))
				})
			})

			When("filtering by service broker name", func() {
				var secondServiceBroker *servicebrokerstub.ServiceBrokerStub

				BeforeEach(func() {
					secondServiceBroker = servicebrokerstub.New().WithServiceOfferings(1).WithPlans(1)
					secondServiceBroker.Services[0].Name = mixedOfferingName
					secondServiceBroker.EnableServiceAccess()

					args = append(args, "-b", secondServiceBroker.Name)
				})

				AfterEach(func() {
					secondServiceBroker.Forget()
				})

				It("only shows plans for the selected broker", func() {
					Expect(session.Out).To(SatisfyAll(
						Say(`Getting service plan information for service offering %s from service broker %s in org %s / space %s as %s\.\.\.\n`, mixedOfferingName, secondServiceBroker.Name, org1, space1, username),
						Say(`\n`),
						Say(`broker: %s\n`, secondServiceBroker.Name),
						Say(`plan\s+description\s+free or paid\s+costs\n`),
						Say(`%s\s+%s\s+%s\s+\n`, secondServiceBroker.FirstServicePlanName(), secondServiceBroker.FirstServicePlanDescription(), "free"),
					))

					Expect(session.Out).NotTo(SatisfyAny(
						Say(broker.Name),
						Say(publicPlanName),
						Say(org1PlanName),
					))
				})
			})

			When("a service plan is unavailable", func() {
				BeforeEach(func() {
					deactivateOrg1Plan()
				})

				It("does not show the unavailable plan", func() {
					Expect(string(session.Out.Contents())).NotTo(ContainSubstring(org1PlanName))
				})

				When("the --show-unavailable flag is specified", func() {
					BeforeEach(func() {
						args = append(args, "--show-unavailable")
					})

					It("shows unavailable plans and the plan availability column", func() {
						Expect(session.Out).To(SatisfyAll(
							Say(`Getting service plan information for service offering %s in org %s / space %s as %s\.\.\.\n`, mixedOfferingName, org1, space1, username),
							Say(`\n`),
							Say(`broker: %s\n`, broker.Name),
							Say(`plan\s+description\s+free or paid\s+costs\s+available\n`),
							Say(`%s\s+%s\s+%s\s+%s\s+%s\n`, publicPlanName, broker.Services[mixedOfferingIndex].Plans[publicChargedPlanIndex].Description, "paid", "GBP 600.00/MONTHLY, USD 649.00/MONTHLY, USD 1.00/1GB of messages over 20GB", "yes"),
							Say(`%s\s+%s\s+%s\s+%s\n`, org1PlanName, broker.Services[mixedOfferingIndex].Plans[org1PlanIndex].Description, "free", "no"),
						))
					})
				})
			})

			// Available

			//When("a plan has cost information", func() {
			//	var brokerWithCosts *servicebrokerstub.ServiceBrokerStub
			//
			//	BeforeEach(func() {
			//		brokerWithCosts = servicebrokerstub.New().WithServiceOfferings(1).WithPlans(4)
			//
			//		brokerWithCosts.Services[0].Plans[0].Free = false
			//		brokerWithCosts.Services[0].Plans[0].Costs = []config.Cost{
			//			{
			//				Amount: map[string]float64{"gbp": 600.00, "usd": 649.00},
			//				Unit:   "MONTHLY",
			//			},
			//			{
			//				Amount: map[string]float64{"usd": 0.999},
			//				Unit:   "1GB of messages over 20GB",
			//			},
			//		}
			//
			//		brokerWithCosts.Services[0].Plans[1].Free = false
			//		brokerWithCosts.Services[0].Plans[1].Costs = []config.Cost{{
			//			Amount: map[string]float64{"gbp": 600.00},
			//			Unit:   "MONTHLY",
			//		}}
			//
			//		brokerWithCosts.Services[0].Plans[2].Free = false
			//
			//		brokerWithCosts.EnableServiceAccess()
			//	})
			//
			//	AfterEach(func() {
			//		brokerWithCosts.Forget()
			//	})
			//
			//	It("shows the costs", func() {
			//		Expect(session.Out).To(SatisfyAll(
			//			Say(`Getting service plan information for service offering %s in org %s / space %s as %s\.\.\.\n`, mixedOfferingName, org1, space1, username),
			//			Say(`\n`),
			//			Say(`broker: %s\n`, brokerWithCosts.Name),
			//			Say(`plan\s+description\s+free or paid\s+costs\n`),
			//			Say(`%s\s+%s\s+%s\s+%s\n`, brokerWithCosts.Services[0].Plans[0].Name, brokerWithCosts.Services[0].Plans[0].Description, "paid", "GBP 600.00/MONTHLY, USD 649.00/MONTHLY, USD 1.00/1GB of messages over 20GB"),
			//			Say(`%s\s+%s\s+%s\s+%s\n`, brokerWithCosts.Services[0].Plans[1].Name, brokerWithCosts.Services[0].Plans[1].Description, "paid", "GBP 600.00/MONTHLY"),
			//			Say(`%s\s+%s\s+%s\s+\n`, brokerWithCosts.Services[0].Plans[2].Name, brokerWithCosts.Services[0].Plans[2].Description, "paid"),
			//			Say(`%s\s+%s\s+%s\s+\n`, brokerWithCosts.Services[0].Plans[3].Name, brokerWithCosts.Services[0].Plans[3].Description, "free"),
			//		))
			//	})
			//})
		})

	})
})
