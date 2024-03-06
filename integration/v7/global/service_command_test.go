package global

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/integration/helpers/servicebrokerstub"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("service command", func() {

	var (
		serviceInstanceName string
		orgName             string
		sourceSpaceName     string

		service     string
		servicePlan string
		broker      *servicebrokerstub.ServiceBrokerStub
	)

	BeforeEach(func() {
		serviceInstanceName = helpers.PrefixedRandomName("SI")
		orgName = helpers.NewOrgName()
		sourceSpaceName = helpers.NewSpaceName()
		helpers.SetupCF(orgName, sourceSpaceName)

		broker = servicebrokerstub.EnableServiceAccess()
		service = broker.FirstServiceOfferingName()
		servicePlan = broker.FirstServicePlanName()

		Eventually(helpers.CF("create-service", service, servicePlan, serviceInstanceName)).Should(Exit(0))
	})

	AfterEach(func() {
		broker.Forget()
		helpers.QuickDeleteOrg(orgName)
	})

	Context("service is shared between two spaces", func() {
		var (
			targetSpaceName string
		)

		BeforeEach(func() {
			targetSpaceName = helpers.NewSpaceName()
			helpers.CreateOrgAndSpace(orgName, targetSpaceName)
			helpers.TargetOrgAndSpace(orgName, sourceSpaceName)
			Eventually(helpers.CF("share-service", serviceInstanceName, "-s", targetSpaceName)).Should(Exit(0))
		})

		Context("due to global settings of service sharing disabled", func() {
			BeforeEach(func() {
				helpers.DisableFeatureFlag("service_instance_sharing")
			})

			AfterEach(func() {
				helpers.EnableFeatureFlag("service_instance_sharing")
			})

			It("should display that the service instance feature flag is disabled", func() {
				session := helpers.CF("service", serviceInstanceName)
				Eventually(session).Should(Say(`The "service_instance_sharing" feature flag is disabled for this Cloud Foundry platform.`))
				Eventually(session).Should(Exit(0))
			})

			Context("AND service broker does not allow service instance sharing", func() {
				BeforeEach(func() {
					broker.Services[0].Shareable = false
					broker.Configure().Register()
				})

				It("should display that service instance sharing is disabled for this service (global message)", func() {
					session := helpers.CF("service", serviceInstanceName)
					Eventually(session).Should(Say(`The "service_instance_sharing" feature flag is disabled for this Cloud Foundry platform. Also, service instance sharing is disabled for this service.`))
					Eventually(session).Should(Exit(0))
				})
			})
		})
	})
})
