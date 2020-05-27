package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/integration/helpers/servicebrokerstub"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("delete-service command", func() {
	When("an api is targeted and the user is logged in", func() {
		BeforeEach(func() {
			helpers.LoginCF()
		})

		When("the environment is not setup correctly", func() {
			It("fails with the appropriate errors", func() {
				By("checking the org is targeted correctly")
				session := helpers.CF("delete-service", "service-name", "-f")
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Out).Should(Say("No org and space targeted, use 'cf target -o ORG -s SPACE' to target an org and space"))
				Eventually(session).Should(Exit(1))

				By("checking the space is targeted correctly")
				helpers.TargetOrg(ReadOnlyOrg)
				session = helpers.CF("delete-service", "service-name", "-f")
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Out).Should(Say(`No space targeted, use 'cf target -s' to target a space\.`))
				Eventually(session).Should(Exit(1))
			})
		})

		When("an org and space are targeted", func() {
			var (
				orgName   string
				spaceName string
			)

			BeforeEach(func() {
				orgName = helpers.NewOrgName()
				spaceName = helpers.NewSpaceName()
				helpers.CreateOrgAndSpace(orgName, spaceName)
				helpers.TargetOrgAndSpace(orgName, spaceName)
			})

			AfterEach(func() {
				helpers.QuickDeleteOrg(orgName)
			})

			When("there is a service instance and it is bound to an app", func() {
				var (
					service     string
					servicePlan string
					broker      *servicebrokerstub.ServiceBrokerStub

					serviceInstanceName string
					appName             string
				)

				BeforeEach(func() {
					broker = servicebrokerstub.EnableServiceAccess()
					service = broker.FirstServiceOfferingName()
					servicePlan = broker.FirstServicePlanName()

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
					broker.Forget()
				})

				It("should display an error message that the service instance's keys, bindings, and shares must first be deleted", func() {
					session := helpers.CF("delete-service", serviceInstanceName, "-f")
					Eventually(session).Should(Say(`Cannot delete service instance. Service keys, bindings, and shares must first be deleted\.`))
					Eventually(session).Should(Exit(1))
				})
			})
		})
	})
})
