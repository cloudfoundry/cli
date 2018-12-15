// +build !partialPush

package experimental

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("service-access command", func() {
	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(false, false, ReadOnlyOrg, "service-access")
		})
	})

	When("the environment is setup correctly", func() {
		var (
			orgName1   string
			spaceName1 string

			serviceBroker1   helpers.ServiceBroker
			servicePlanName1 string
		)

		BeforeEach(func() {
			orgName1 = helpers.NewOrgName()
			spaceName1 = helpers.NewSpaceName()

			helpers.SetupCF(orgName1, spaceName1)

			servicePlanName1 = helpers.NewPlanName()
			serviceBroker1 = helpers.NewServiceBroker(
				helpers.NewServiceBrokerName(),
				helpers.NewAssets().ServiceBroker,
				helpers.DefaultSharedDomain(),
				helpers.PrefixedRandomName("service"),
				servicePlanName1,
			)
			serviceBroker1.SyncPlans[1].Name = helpers.GenerateHigherName(helpers.NewPlanName, servicePlanName1)

			serviceBroker1.Push()
			serviceBroker1.Configure(true)
			serviceBroker1.Create()
		})

		AfterEach(func() {
			serviceBroker1.Destroy()
			helpers.QuickDeleteOrg(orgName1)
		})

		Describe("service plan visibility", func() {
			It("displays the correct state for the visibility", func() {
				By("having nothing enabled")
				session := helpers.CF("service-access", "-e", serviceBroker1.Service.Name)
				Eventually(session).Should(Say(`broker:\s+%s`, serviceBroker1.Name))
				Eventually(session).Should(Say(`%s\s+%s\s+none`,
					serviceBroker1.Service.Name,
					servicePlanName1,
				))
				Eventually(session).Should(Exit(0))

				By("having the plan enabled for just the org and space")
				Eventually(
					helpers.CF("enable-service-access",
						serviceBroker1.Service.Name,
						"-o", orgName1,
						"-p", servicePlanName1)).Should(Exit(0))

				session = helpers.CF("service-access", "-e", serviceBroker1.Service.Name)
				Eventually(session).Should(Say(`broker:\s+%s`, serviceBroker1.Name))
				Eventually(session).Should(Say(`%s\s+%s\s+limited\s+%s`,
					serviceBroker1.Service.Name,
					servicePlanName1,
					orgName1))
				Eventually(session).Should(Exit(0))

				By("having the plan enabled for everyone")
				Eventually(helpers.CF("enable-service-access", serviceBroker1.Service.Name)).Should(Exit(0))

				session = helpers.CF("service-access", "-e", serviceBroker1.Service.Name)
				Eventually(session).Should(Say(`broker:\s+%s`, serviceBroker1.Name))
				Eventually(session).Should(Say(`%s\s+%s\s+all`,
					serviceBroker1.Service.Name,
					servicePlanName1,
				))
				Eventually(session).Should(Exit(0))
			})
		})

		Describe("narrowing display with flags", func() {
			var (
				serviceBroker2   helpers.ServiceBroker
				servicePlanName2 string

				orgName2 string
			)

			BeforeEach(func() {
				servicePlanName2 = helpers.NewPlanName()
				serviceBroker2 = helpers.NewServiceBroker(
					helpers.GenerateLowerName(helpers.NewServiceBrokerName, serviceBroker1.Name),
					helpers.NewAssets().ServiceBroker,
					helpers.DefaultSharedDomain(),
					helpers.PrefixedRandomName("service"),
					servicePlanName2,
				)
				serviceBroker2.SyncPlans[1].Name = helpers.GenerateLowerName(helpers.NewPlanName, servicePlanName2)

				serviceBroker2.Push()
				serviceBroker2.Configure(true)
				serviceBroker2.Create()

				Eventually(
					helpers.CF("enable-service-access",
						serviceBroker1.Service.Name,
						"-o", orgName1,
						"-p", servicePlanName1)).Should(Exit(0))

				orgName2 = helpers.GenerateLowerName(helpers.NewOrgName, orgName1)
				helpers.CreateOrg(orgName2)

				Eventually(
					helpers.CF("enable-service-access",
						serviceBroker1.Service.Name,
						"-o", orgName2,
						"-p", servicePlanName1)).Should(Exit(0))
				Eventually(helpers.CF("enable-service-access", serviceBroker2.Service.Name)).Should(Exit(0))
			})

			AfterEach(func() {
				helpers.QuickDeleteOrg(orgName2)
				serviceBroker2.Destroy()
			})

			It("only shows broker/service information based on the flags provided", func() {
				By("by showing the brokers and plans in alphabetical order when no flags are provided")
				session := helpers.CF("service-access")
				Eventually(session).Should(Say(`broker:\s+%s`, serviceBroker2.Name))
				Eventually(session).Should(Say(`%s\s+%s\s+all`,
					serviceBroker2.Service.Name,
					serviceBroker2.SyncPlans[1].Name,
				))
				Eventually(session).Should(Say(`%s\s+%s\s+all`,
					serviceBroker2.Service.Name,
					serviceBroker2.SyncPlans[0].Name,
				))
				Eventually(session).Should(Say(`broker:\s+%s`, serviceBroker1.Name))
				Eventually(session).Should(Say(`%s\s+%s\s+limited\s+%s,%s`,
					serviceBroker1.Service.Name,
					serviceBroker1.SyncPlans[0].Name,
					orgName2,
					orgName1,
				))
				Eventually(session).Should(Say(`%s\s+%s\s+none`,
					serviceBroker1.Service.Name,
					serviceBroker1.SyncPlans[1].Name,
				))
				Eventually(session).Should(Exit(0))

				By("by showing the specified broker and it's plans in alphabetical order when the -b flag is provided")
				session = helpers.CF("service-access", "-b", serviceBroker2.Name)
				Eventually(session).Should(Say(`broker:\s+%s`, serviceBroker2.Name))
				Eventually(session).Should(Say(`%s\s+%s\s+all`,
					serviceBroker2.Service.Name,
					serviceBroker2.SyncPlans[1].Name,
				))
				Eventually(session).Should(Say(`%s\s+%s\s+all`,
					serviceBroker2.Service.Name,
					serviceBroker2.SyncPlans[0].Name,
				))
				Consistently(session).ShouldNot(Say(`broker:\s+%s`, serviceBroker1.Name))
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
