package isolated

import (
	"strings"

	"code.cloudfoundry.org/cli/integration/assets/hydrabroker/config"
	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/integration/helpers/servicebrokerstub"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("marketplace command", func() {
	Describe("help", func() {
		expectMarketplaceHelpMessage := func(session *Session) {
			Expect(session).To(SatisfyAll(
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
			))
		}

		When("the --help flag is set", func() {
			It("displays command usage to output", func() {
				session := helpers.CF("marketplace", "--help")
				Eventually(session).Should(Exit(0))
				expectMarketplaceHelpMessage(session)
			})
		})

		When("more than required number of args are passed", func() {
			It("displays an error message with help text and exits 1", func() {
				session := helpers.CF("marketplace", "lala")
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say(`Incorrect Usage: unexpected argument "lala"`))
				expectMarketplaceHelpMessage(session)
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

			Expect(session).To(Say("FAILED"))
			Expect(session.Err).To(Say(`No API endpoint set\. Use 'cf login' or 'cf api' to target an endpoint\.`))
		})
	})

	When("services and plans are registered", func() {
		var (
			session                            *Session
			org1, org2, space1, space2         string
			serviceInstanceName                string
			unavailablePlanName                string
			brokerWithPublicPlans              *servicebrokerstub.ServiceBrokerStub
			brokerWithSomePrivatePlans         *servicebrokerstub.ServiceBrokerStub
			brokerWithSameOfferingNameAndCosts *servicebrokerstub.ServiceBrokerStub
			brokerWithUnavailablePlan          *servicebrokerstub.ServiceBrokerStub
		)

		BeforeEach(func() {
			helpers.LoginCF()

			org1 = helpers.NewOrgName()
			org2 = helpers.NewOrgName()
			space1 = helpers.NewSpaceName()
			space2 = helpers.NewSpaceName()

			helpers.CreateOrgAndSpace(org1, space1)
			helpers.CreateOrgAndSpace(org2, space2)

			brokerWithPublicPlans = servicebrokerstub.New()
			brokerWithSomePrivatePlans = servicebrokerstub.New().WithServiceOfferings(2).WithPlans(4)
			brokerWithSomePrivatePlans.Services[0].Plans[0].Name = helpers.PrefixedRandomName("INTEGRATION-PLAN-ORG1")
			brokerWithSomePrivatePlans.Services[0].Plans[1].Name = helpers.PrefixedRandomName("INTEGRATION-PLAN-ORG2")
			brokerWithSomePrivatePlans.Services[0].Plans[3].Name = helpers.PrefixedRandomName("INTEGRATION-PLAN-PUBLIC")
			brokerWithSomePrivatePlans.Services[0].Plans[0].Name = helpers.PrefixedRandomName("INTEGRATION-PLAN-NO-ACCESS")

			brokerWithSomePrivatePlans.ServiceAccessConfig = []servicebrokerstub.ServiceAccessConfig{
				{
					OfferingName: brokerWithSomePrivatePlans.FirstServiceOfferingName(),
					PlanName:     brokerWithSomePrivatePlans.Services[0].Plans[0].Name,
					OrgName:      org1,
				},
				{
					OfferingName: brokerWithSomePrivatePlans.FirstServiceOfferingName(),
					PlanName:     brokerWithSomePrivatePlans.Services[0].Plans[1].Name,
					OrgName:      org2,
				},
				{
					OfferingName: brokerWithSomePrivatePlans.FirstServiceOfferingName(),
					PlanName:     brokerWithSomePrivatePlans.Services[0].Plans[2].Name,
				},
				{
					OfferingName: brokerWithSomePrivatePlans.Services[1].Name,
				},
			}

			brokerWithSameOfferingNameAndCosts = servicebrokerstub.New().WithPlans(2)
			brokerWithSameOfferingNameAndCosts.Services[0].Name = brokerWithSomePrivatePlans.Services[0].Name
			brokerWithSameOfferingNameAndCosts.Services[0].Plans[0].Free = false
			brokerWithSameOfferingNameAndCosts.Services[0].Plans[0].Costs = []config.Cost{
				{
					Amount: map[string]float64{"gbp": 600.00, "usd": 649.00},
					Unit:   "MONTHLY",
				},
				{
					Amount: map[string]float64{"usd": 0.999},
					Unit:   "1GB of messages over 20GB",
				},
			}
			brokerWithSameOfferingNameAndCosts.Services[0].Plans[1].Free = false

			unavailablePlanName = helpers.PrefixedRandomName("UNAVAILABLE-PLAN")
			brokerWithUnavailablePlan = servicebrokerstub.New().WithServiceOfferings(1).WithPlans(1)
			brokerWithUnavailablePlan.Services[0].Plans[0].Name = unavailablePlanName

			brokerWithPublicPlans.EnableServiceAccess()
			brokerWithSomePrivatePlans.EnableServiceAccess()
			brokerWithSameOfferingNameAndCosts.EnableServiceAccess()
			brokerWithUnavailablePlan.EnableServiceAccess()

			serviceInstanceName = helpers.NewServiceInstanceName()
			helpers.TargetOrgAndSpace(org1, space1)
			Eventually(helpers.CF("create-service", brokerWithUnavailablePlan.Services[0].Name, brokerWithUnavailablePlan.Services[0].Plans[0].Name, serviceInstanceName)).Should(Exit(0))

			brokerWithUnavailablePlan.Services[0].Plans[0].Name = helpers.NewPlanName()
			brokerWithUnavailablePlan.Services[0].Plans[0].ID = helpers.RandomName()
			brokerWithUnavailablePlan.Services[0].Plans[0].Free = false
			brokerWithUnavailablePlan.Services[0].Plans[0].Costs = []config.Cost{{
				Amount: map[string]float64{"gbp": 600.00},
				Unit:   "MONTHLY",
			}}
			brokerWithUnavailablePlan.Configure().Register().EnableServiceAccess()
		})

		AfterEach(func() {
			helpers.LoginCF()

			helpers.TargetOrgAndSpace(org1, space1)
			Eventually(helpers.CF("delete-service", "-f", serviceInstanceName)).Should(Exit(0))

			helpers.QuickDeleteOrg(org1)
			helpers.QuickDeleteOrg(org2)

			brokerWithPublicPlans.Forget()
			brokerWithSomePrivatePlans.Forget()
			brokerWithSameOfferingNameAndCosts.Forget()
			brokerWithUnavailablePlan.Forget()
		})

		Context("no service name filter", func() {
			expectHeaders := func(expected string, args ...interface{}) {
				ExpectWithOffset(1, session).To(SatisfyAll(
					Say(expected, args...),
					Say(`\n\n`),
					Say(`offering\s+plans\s+description\s+broker`),
					Say(`\n\n`),
					Say(`TIP: Use 'cf marketplace -e SERVICE_OFFERING' to view descriptions of individual plans of a given service offering\.`),
				))
			}

			expectEntry := func(offering, plans, description, broker string) {
				ExpectWithOffset(1, BufferWithBytes(session.Out.Contents())).To(SatisfyAll(
					Say(`offering\s+plans\s+description\s+broker`),
					Say(`%s\s+%s\s+%s\s+%s`, offering, plans, description, broker),
				))
			}

			When("not logged in", func() {
				BeforeEach(func() {
					helpers.LogoutCF()
				})

				It("displays all public offerings and plans", func() {
					session = helpers.CF("marketplace")
					Eventually(session).Should(Exit(0))

					expectHeaders(`Getting all service offerings from marketplace\.\.\.`)
					expectEntry(brokerWithPublicPlans.FirstServiceOfferingName(), planNamesOf(brokerWithPublicPlans), brokerWithPublicPlans.FirstServiceOfferingDescription(), brokerWithPublicPlans.Name)
					expectEntry(brokerWithSomePrivatePlans.FirstServiceOfferingName(), brokerWithSomePrivatePlans.Services[0].Plans[2].Name, brokerWithSomePrivatePlans.FirstServiceOfferingDescription(), brokerWithSomePrivatePlans.Name)
					expectEntry(brokerWithSomePrivatePlans.Services[1].Name, brokerWithSomePrivatePlans.Services[1].Plans[0].Name, brokerWithSomePrivatePlans.Services[1].Description, brokerWithSomePrivatePlans.Name)
					expectEntry(brokerWithSameOfferingNameAndCosts.FirstServiceOfferingName(), planNamesOf(brokerWithSameOfferingNameAndCosts), brokerWithSameOfferingNameAndCosts.FirstServiceOfferingDescription(), brokerWithSameOfferingNameAndCosts.Name)

					Expect(string(session.Out.Contents())).NotTo(ContainSubstring(unavailablePlanName))
				})

				It("can filter by service broker name", func() {
					session = helpers.CF("marketplace", "-b", brokerWithSomePrivatePlans.Name)
					Eventually(session).Should(Exit(0))

					expectHeaders(`Getting all service offerings from marketplace for service broker %s\.\.\.`, brokerWithSomePrivatePlans.Name)

					expectEntry(brokerWithSomePrivatePlans.FirstServiceOfferingName(), brokerWithSomePrivatePlans.Services[0].Plans[2].Name, brokerWithSomePrivatePlans.FirstServiceOfferingDescription(), brokerWithSomePrivatePlans.Name)
					expectEntry(brokerWithSomePrivatePlans.Services[1].Name, brokerWithSomePrivatePlans.Services[1].Plans[0].Name, brokerWithSomePrivatePlans.Services[1].Description, brokerWithSomePrivatePlans.Name)

					Expect(string(session.Out.Contents())).NotTo(SatisfyAny(
						ContainSubstring(brokerWithPublicPlans.Name),
						ContainSubstring(brokerWithSameOfferingNameAndCosts.Name),
					))
				})

				It("can show unavailable plans", func() {
					session = helpers.CF("marketplace", "--show-unavailable")
					Eventually(session).Should(Exit(0))

					plans := strings.Join([]string{
						unavailablePlanName,
						brokerWithUnavailablePlan.Services[0].Plans[0].Name},
						", ",
					)

					expectHeaders(`Getting all service offerings from marketplace\.\.\.`)
					expectEntry(brokerWithUnavailablePlan.FirstServiceOfferingName(), plans, brokerWithUnavailablePlan.FirstServiceOfferingDescription(), brokerWithUnavailablePlan.Name)
				})
			})

			When("logged in and targeting a space", func() {
				var username string

				BeforeEach(func() {
					helpers.TargetOrgAndSpace(org1, space1)
					username, _ = helpers.GetCredentials()
				})

				It("displays public offerings and plans, and those enabled for that space", func() {
					session = helpers.CF("marketplace")
					Eventually(session).Should(Exit(0))

					broker2Plans := strings.Join([]string{
						brokerWithSomePrivatePlans.Services[0].Plans[0].Name,
						brokerWithSomePrivatePlans.Services[0].Plans[2].Name},
						", ",
					)

					expectHeaders(`Getting all service offerings from marketplace in org %s / space %s as %s\.\.\.`, org1, space1, username)
					expectEntry(brokerWithPublicPlans.FirstServiceOfferingName(), planNamesOf(brokerWithPublicPlans), brokerWithPublicPlans.FirstServiceOfferingDescription(), brokerWithPublicPlans.Name)
					expectEntry(brokerWithSomePrivatePlans.FirstServiceOfferingName(), broker2Plans, brokerWithSomePrivatePlans.FirstServiceOfferingDescription(), brokerWithSomePrivatePlans.Name)
					expectEntry(brokerWithSomePrivatePlans.Services[1].Name, brokerWithSomePrivatePlans.Services[1].Plans[0].Name, brokerWithSomePrivatePlans.Services[1].Description, brokerWithSomePrivatePlans.Name)
					expectEntry(brokerWithSameOfferingNameAndCosts.FirstServiceOfferingName(), planNamesOf(brokerWithSameOfferingNameAndCosts), brokerWithSameOfferingNameAndCosts.FirstServiceOfferingDescription(), brokerWithSameOfferingNameAndCosts.Name)

					Expect(string(session.Out.Contents())).NotTo(ContainSubstring(unavailablePlanName))
				})

				It("can filter by service broker name", func() {
					session = helpers.CF("marketplace", "-b", brokerWithSomePrivatePlans.Name)
					Eventually(session).Should(Exit(0))

					broker2Plans := strings.Join([]string{
						brokerWithSomePrivatePlans.Services[0].Plans[0].Name,
						brokerWithSomePrivatePlans.Services[0].Plans[2].Name},
						", ",
					)

					expectHeaders(`Getting all service offerings from marketplace for service broker %s in org %s / space %s as %s\.\.\.`, brokerWithSomePrivatePlans.Name, org1, space1, username)
					expectEntry(brokerWithSomePrivatePlans.FirstServiceOfferingName(), broker2Plans, brokerWithSomePrivatePlans.FirstServiceOfferingDescription(), brokerWithSomePrivatePlans.Name)
					expectEntry(brokerWithSomePrivatePlans.Services[1].Name, brokerWithSomePrivatePlans.Services[1].Plans[0].Name, brokerWithSomePrivatePlans.Services[1].Description, brokerWithSomePrivatePlans.Name)

					Expect(string(session.Out.Contents())).NotTo(SatisfyAny(
						ContainSubstring(brokerWithPublicPlans.Name),
						ContainSubstring(brokerWithSameOfferingNameAndCosts.Name),
					))
				})

				It("can show unavailable plans", func() {
					session = helpers.CF("marketplace", "--show-unavailable")
					Eventually(session).Should(Exit(0))

					plans := strings.Join([]string{
						unavailablePlanName,
						brokerWithUnavailablePlan.Services[0].Plans[0].Name},
						", ",
					)

					expectHeaders(`Getting all service offerings from marketplace in org %s / space %s as %s\.\.\.`, org1, space1, username)
					expectEntry(brokerWithUnavailablePlan.FirstServiceOfferingName(), plans, brokerWithUnavailablePlan.FirstServiceOfferingDescription(), brokerWithUnavailablePlan.Name)
				})
			})
		})

		Context("filtering by service offering name", func() {
			expectEntry := func(broker, plan, description, free, costs string) {
				ExpectWithOffset(1, BufferWithBytes(session.Out.Contents())).To(SatisfyAll(
					Say(`\n\n`),
					Say(`broker: %s`, broker),
					Say(`plan\s+description\s+free or paid\s+costs`),
					Not(Say(`available`)),
					Say(`%s\s+%s\s+%s\s+%s`, plan, description, free, costs),
				))
			}

			expectEntryWithAvailability := func(broker, plan, description, free, costs, available string) {
				ExpectWithOffset(1, BufferWithBytes(session.Out.Contents())).To(SatisfyAll(
					Say(`\n\n`),
					Say(`broker: %s`, broker),
					Say(`plan\s+description\s+free or paid\s+costs\s+available`),
					Say(`%s\s+%s\s+%s\s+%s\s+%s`, plan, description, free, costs, available),
				))
			}

			When("not logged in", func() {
				BeforeEach(func() {
					helpers.LogoutCF()
				})

				It("filters by service offering name", func() {
					session = helpers.CF("marketplace", "-e", brokerWithSomePrivatePlans.Services[0].Name)
					Eventually(session).Should(Exit(0))

					Expect(session).To(Say(`Getting service plan information for service offering %s\.\.\.`, brokerWithSomePrivatePlans.Services[0].Name))

					expectEntry(
						brokerWithSomePrivatePlans.Name,
						brokerWithSomePrivatePlans.Services[0].Plans[2].Name,
						brokerWithSomePrivatePlans.Services[0].Plans[2].Description,
						"free",
						"",
					)

					expectEntry(
						brokerWithSameOfferingNameAndCosts.Name,
						brokerWithSameOfferingNameAndCosts.Services[0].Plans[0].Name,
						brokerWithSameOfferingNameAndCosts.Services[0].Plans[0].Description,
						"paid",
						"GBP 600.00/MONTHLY, USD 649.00/MONTHLY, USD 1.00/1GB of messages over 20GB",
					)

					expectEntry(
						brokerWithSameOfferingNameAndCosts.Name,
						brokerWithSameOfferingNameAndCosts.Services[0].Plans[1].Name,
						brokerWithSameOfferingNameAndCosts.Services[0].Plans[1].Description,
						"paid",
						"",
					)

					Expect(string(session.Out.Contents())).NotTo(SatisfyAny(
						ContainSubstring(brokerWithPublicPlans.Name),
						ContainSubstring(brokerWithSomePrivatePlans.Services[1].Name),
						ContainSubstring(brokerWithUnavailablePlan.Name),
					))
				})

				It("can also filter by service broker name", func() {
					session = helpers.CF("marketplace", "-e", brokerWithSomePrivatePlans.Services[0].Name, "-b", brokerWithSomePrivatePlans.Name)
					Eventually(session).Should(Exit(0))

					Expect(session).To(Say(`Getting service plan information for service offering %s from service broker %s\.\.\.`, brokerWithSomePrivatePlans.Services[0].Name, brokerWithSomePrivatePlans.Name))

					expectEntry(
						brokerWithSomePrivatePlans.Name,
						brokerWithSomePrivatePlans.Services[0].Plans[2].Name,
						brokerWithSomePrivatePlans.Services[0].Plans[2].Description,
						"free",
						"",
					)

					Expect(string(session.Out.Contents())).NotTo(SatisfyAny(
						ContainSubstring(brokerWithPublicPlans.Name),
						ContainSubstring(brokerWithSomePrivatePlans.Services[0].Plans[0].Name),
						ContainSubstring(brokerWithSomePrivatePlans.Services[1].Plans[0].Name),
						ContainSubstring(brokerWithSameOfferingNameAndCosts.Name),
					))
				})

				It("can show unavailable plans", func() {
					session = helpers.CF("marketplace", "-e", brokerWithUnavailablePlan.Services[0].Name, "--show-unavailable")
					Eventually(session).Should(Exit(0))

					Expect(session).To(Say(`Getting service plan information for service offering %s\.\.\.`, brokerWithUnavailablePlan.Services[0].Name))

					expectEntryWithAvailability(
						brokerWithUnavailablePlan.Name,
						brokerWithUnavailablePlan.Services[0].Plans[0].Name,
						brokerWithUnavailablePlan.Services[0].Plans[0].Description,
						"paid",
						"GBP 600.00/MONTHLY",
						"yes",
					)

					expectEntryWithAvailability(
						brokerWithUnavailablePlan.Name,
						unavailablePlanName,
						brokerWithUnavailablePlan.Services[0].Plans[0].Description,
						"free",
						"",
						"no",
					)
				})
			})

			When("logged in and targeting a space", func() {
				var username string

				BeforeEach(func() {
					helpers.TargetOrgAndSpace(org1, space1)
					username, _ = helpers.GetCredentials()
				})

				It("filters by service offering name", func() {
					session = helpers.CF("marketplace", "-e", brokerWithSomePrivatePlans.Services[0].Name)
					Eventually(session).Should(Exit(0))

					Expect(session).To(Say(`Getting service plan information for service offering %s in org %s / space %s as %s\.\.\.`, brokerWithSomePrivatePlans.Services[0].Name, org1, space1, username))

					expectEntry(
						brokerWithSomePrivatePlans.Name,
						brokerWithSomePrivatePlans.Services[0].Plans[0].Name,
						brokerWithSomePrivatePlans.Services[0].Plans[0].Description,
						"free",
						"",
					)

					expectEntry(
						brokerWithSomePrivatePlans.Name,
						brokerWithSomePrivatePlans.Services[0].Plans[2].Name,
						brokerWithSomePrivatePlans.Services[0].Plans[2].Description,
						"free",
						"",
					)

					expectEntry(
						brokerWithSameOfferingNameAndCosts.Name,
						brokerWithSameOfferingNameAndCosts.Services[0].Plans[0].Name,
						brokerWithSameOfferingNameAndCosts.Services[0].Plans[0].Description,
						"paid",
						"GBP 600.00/MONTHLY, USD 649.00/MONTHLY, USD 1.00/1GB of messages over 20GB",
					)

					expectEntry(
						brokerWithSameOfferingNameAndCosts.Name,
						brokerWithSameOfferingNameAndCosts.Services[0].Plans[1].Name,
						brokerWithSameOfferingNameAndCosts.Services[0].Plans[1].Description,
						"paid",
						"",
					)

					Expect(string(session.Out.Contents())).NotTo(ContainSubstring(brokerWithUnavailablePlan.Name))
				})

				It("can also filter by service broker name", func() {
					session = helpers.CF("marketplace", "-e", brokerWithSomePrivatePlans.Services[0].Name, "-b", brokerWithSomePrivatePlans.Name)
					Eventually(session).Should(Exit(0))

					Expect(session).To(Say(`Getting service plan information for service offering %s from service broker %s in org %s / space %s as %s\.\.\.`, brokerWithSomePrivatePlans.Services[0].Name, brokerWithSomePrivatePlans.Name, org1, space1, username))

					expectEntry(
						brokerWithSomePrivatePlans.Name,
						brokerWithSomePrivatePlans.Services[0].Plans[0].Name,
						brokerWithSomePrivatePlans.Services[0].Plans[0].Description,
						"free",
						"",
					)

					expectEntry(
						brokerWithSomePrivatePlans.Name,
						brokerWithSomePrivatePlans.Services[0].Plans[2].Name,
						brokerWithSomePrivatePlans.Services[0].Plans[2].Description,
						"free",
						"",
					)

					Expect(string(session.Out.Contents())).NotTo(SatisfyAny(
						ContainSubstring(brokerWithPublicPlans.Name),
						ContainSubstring(brokerWithSameOfferingNameAndCosts.Name),
					))
				})

				It("can show unavailable plans", func() {
					session = helpers.CF("marketplace", "-e", brokerWithUnavailablePlan.Services[0].Name, "--show-unavailable")
					Eventually(session).Should(Exit(0))

					Expect(session).To(Say(`Getting service plan information for service offering %s in org %s / space %s as %s\.\.\.`, brokerWithUnavailablePlan.Services[0].Name, org1, space1, username))

					expectEntryWithAvailability(
						brokerWithUnavailablePlan.Name,
						brokerWithUnavailablePlan.Services[0].Plans[0].Name,
						brokerWithUnavailablePlan.Services[0].Plans[0].Description,
						"paid",
						"GBP 600.00/MONTHLY",
						"yes",
					)

					expectEntryWithAvailability(
						brokerWithUnavailablePlan.Name,
						unavailablePlanName,
						brokerWithUnavailablePlan.Services[0].Plans[0].Description,
						"free",
						"",
						"no",
					)
				})
			})
		})
	})
})

func planNamesOf(broker *servicebrokerstub.ServiceBrokerStub) string {
	var planNames []string
	for _, p := range broker.Services[0].Plans {
		planNames = append(planNames, p.Name)
	}
	return strings.Join(planNames, ", ")
}
