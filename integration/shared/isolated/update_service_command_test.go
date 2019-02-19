package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("update-service command", func() {
	When("an api is targeted, the user is logged in, and an org and space are targeted", func() {
		var (
			orgName string
		)

		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			var spaceName = helpers.NewSpaceName()
			helpers.SetupCF(orgName, spaceName)
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		When("there is a service instance", func() {
			var (
				service             string
				servicePlan         string
				broker              helpers.ServiceBroker
				serviceInstanceName string
				username            string
			)

			BeforeEach(func() {
				var domain = helpers.DefaultSharedDomain()
				service = helpers.PrefixedRandomName("SERVICE")
				servicePlan = helpers.PrefixedRandomName("SERVICE-PLAN")
				broker = helpers.CreateBroker(domain, service, servicePlan)

				Eventually(helpers.CF("enable-service-access", service)).Should(Exit(0))

				serviceInstanceName = helpers.PrefixedRandomName("SI")
				Eventually(helpers.CF("create-service", service, servicePlan, serviceInstanceName)).Should(Exit(0))

				username, _ = helpers.GetCredentials()
			})

			AfterEach(func() {
				Eventually(helpers.CF("delete-service", serviceInstanceName, "-f")).Should(Exit(0))
				broker.Destroy()
			})

			When("updating to a service plan that does not exist", func() {
				It("displays an informative error message, exits 1", func() {
					session := helpers.CF("update-service", serviceInstanceName, "-p", "non-existing-service-plan")
					Eventually(session).Should(Say("Plan does not exist for the %s service", service))
					Eventually(session).Should(Exit(1))
				})
			})

			When("updating to the same service plan (no-op)", func() {
				It("displays an informative success message, exits 0", func() {
					session := helpers.CF("update-service", serviceInstanceName, "-p", servicePlan)
					Eventually(session).Should(Say("Updating service instance %s as %s...", serviceInstanceName, username))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))
				})
			})
		})
	})
})
