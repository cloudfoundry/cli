package global

import (
	"strings"

	"code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("marketplace command", func() {
	When("an API endpoint is set", func() {
		When("not logged in", func() {
			When("there are no accessible services", func() {
				BeforeEach(func() {
					helpers.LogoutCF()
				})

				It("displays a message that no services are available", func() {
					session := helpers.CF("marketplace")
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Say("\n\n"))
					Eventually(session).Should(Say("No service offerings found"))
					Eventually(session).Should(Exit(0))
				})
			})
		})

		When("logged in", func() {
			var user string

			BeforeEach(func() {
				helpers.LoginCF()
				user, _ = helpers.GetCredentials()
			})

			When("a space is targeted", func() {
				var org, space string

				BeforeEach(func() {
					org = helpers.NewOrgName()
					space = helpers.NewSpaceName()
					helpers.SetupCF(org, space)
				})

				AfterEach(func() {
					helpers.QuickDeleteOrg(org)
				})

				When("there are no accessible services", func() {
					It("displays a message saying no services exist", func() {
						session := helpers.CF("marketplace")
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Say("\n\n"))
						Eventually(session).Should(Say("No service offerings found"))
						Eventually(session).Should(Exit(0))
					})
				})

				When("a service is accessible but not in the currently targeted space", func() {
					var (
						broker1      helpers.ServiceBroker
						org1, space1 string

						broker2      helpers.ServiceBroker
						org2, space2 string
					)

					BeforeEach(func() {
						org1 = helpers.NewOrgName()
						space1 = helpers.NewSpaceName()
						helpers.SetupCF(org1, space1)
						helpers.TargetOrgAndSpace(org1, space1)

						domain := helpers.DefaultSharedDomain()

						broker1 = createBroker(domain, "SERVICE-1", "SERVICE-PLAN-1")
						enableServiceAccessForOrg(broker1, org1)

						org2 = helpers.NewOrgName()
						space2 = helpers.NewSpaceName()
						helpers.CreateOrgAndSpace(org2, space2)
						helpers.TargetOrgAndSpace(org2, space2)

						broker2 = createBroker(domain, "SERVICE-2", "SERVICE-PLAN-2")
						enableServiceAccess(broker2)
					})

					AfterEach(func() {
						helpers.TargetOrgAndSpace(org2, space2)
						broker2.Destroy()
						helpers.QuickDeleteOrg(org2)

						helpers.TargetOrgAndSpace(org1, space1)
						broker1.Destroy()
						helpers.QuickDeleteOrg(org1)
					})

					It("displays a table and tip that does not include that service", func() {
						session := helpers.CF("marketplace")
						Eventually(session).Should(Say("Getting services from marketplace in org %s / space %s as %s\\.\\.\\.", org2, space2, user))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Say("\n\n"))
						Eventually(session).Should(Say("service\\s+plans\\s+description"))
						Consistently(session).ShouldNot(Say("%s\\s+%s\\s+fake service", getServiceName(broker1), getBrokerPlanNames(broker1)))
						Eventually(session).Should(Say("%s\\s+%s\\s+fake service", getServiceName(broker2), getBrokerPlanNames(broker2)))
						Eventually(session).Should(Say("TIP: Use 'cf marketplace -s SERVICE' to view descriptions of individual plans of a given service."))
						Eventually(session).Should(Exit(0))
					})
				})
			})
		})
	})
})

func createBroker(domain, serviceName, planName string) helpers.ServiceBroker {
	service := helpers.PrefixedRandomName(serviceName)
	servicePlan := helpers.PrefixedRandomName(planName)
	broker := helpers.NewServiceBroker(helpers.NewServiceBrokerName(), helpers.NewAssets().ServiceBroker, domain, service, servicePlan)
	broker.Push()
	broker.Configure(true)
	broker.Create()

	return broker
}

func enableServiceAccess(broker helpers.ServiceBroker) {
	Eventually(helpers.CF("enable-service-access", getServiceName(broker))).Should(Exit(0))
}

func enableServiceAccessForOrg(broker helpers.ServiceBroker, orgName string) {
	Eventually(helpers.CF("enable-service-access", getServiceName(broker), "-o", orgName)).Should(Exit(0))
}

func getServiceName(broker helpers.ServiceBroker) string {
	return broker.Service.Name
}

func getPlanName(broker helpers.ServiceBroker) string {
	return broker.SyncPlans[0].Name
}

func getBrokerPlanNames(broker helpers.ServiceBroker) string {
	return strings.Join(plansToNames(append(broker.SyncPlans, broker.AsyncPlans...)), ", ")
}

func plansToNames(plans []helpers.Plan) []string {
	planNames := []string{}
	for _, plan := range plans {
		planNames = append(planNames, plan.Name)
	}
	return planNames
}
