package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("delete-service command", func() {
	Context("when an api is targeted, the user is logged in, and an org and space are targeted", func() {
		var (
			orgName   string
			spaceName string
		)

		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()
			setupCF(orgName, spaceName)
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		Context("when there is a service instance and it is bound to an app", func() {
			var (
				domain      string
				service     string
				servicePlan string
				broker      helpers.ServiceBroker

				serviceInstanceName string
				appName             string
			)

			BeforeEach(func() {
				domain = defaultSharedDomain()
				service = helpers.PrefixedRandomName("SERVICE")
				servicePlan = helpers.PrefixedRandomName("SERVICE-PLAN")
				broker = helpers.NewServiceBroker(helpers.NewServiceBrokerName(), helpers.NewAssets().ServiceBroker, domain, service, servicePlan)
				broker.Push()
				broker.Configure()
				broker.Create()

				Eventually(helpers.CF("enable-service-access", service)).Should(Exit(0))

				serviceInstanceName = helpers.PrefixedRandomName("SI")
				Eventually(helpers.CF("create-service", service, servicePlan, serviceInstanceName)).Should(Exit(0))

				appName = helpers.NewAppName()
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CF("push", appName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack", "--no-route")).Should(Exit(0))
				})

				Eventually(helpers.CF("bind-service", appName, serviceInstanceName)).Should(Exit(0))
			})

			AfterEach(func() {
				Eventually(helpers.CF("unbind-service", appName, serviceInstanceName)).Should(Exit(0))
				Eventually(helpers.CF("delete", appName, "-f")).Should(Exit(0))
				Eventually(helpers.CF("delete-service", serviceInstanceName, "-f")).Should(Exit(0))
				broker.Destroy()
			})

			It("should display an error message that the service instance's keys, bindings, and shares must first be deleted", func() {
				session := helpers.CF("delete-service", serviceInstanceName, "-f")
				Eventually(session.Out).Should(Say("Cannot delete service instance. Service keys, bindings, and shares must first be deleted."))
				Eventually(session).Should(Exit(1))
			})
		})
	})
})
