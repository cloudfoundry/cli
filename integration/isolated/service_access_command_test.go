package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("service-access command", func() {
	Context("when the environment is setup correctly", func() {
		var (
			orgName   string
			spaceName string

			serviceBroker helpers.ServiceBroker
		)

		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()
			setupCF(orgName, spaceName)

			serviceBroker = helpers.NewServiceBroker(
				helpers.NewServiceBrokerName(),
				helpers.NewAssets().ServiceBroker,
				defaultSharedDomain(),
				helpers.PrefixedRandomName("service"),
				helpers.PrefixedRandomName("plan"),
			)

			serviceBroker.Push()
			serviceBroker.Configure()
			serviceBroker.Create()
		})

		AfterEach(func() {
			serviceBroker.Destroy()
			helpers.QuickDeleteOrg(orgName)
		})

		It("sets visibility", func() {
			// initial access is none
			session := helpers.CF("service-access")
			Eventually(session).Should(Say("%s\\s+%s\\s+none",
				serviceBroker.Service.Name,
				serviceBroker.SyncPlans[0].Name,
			))
			Eventually(session).Should(Exit(0))

			// enable access for org and plan
			session = helpers.CF("enable-service-access",
				serviceBroker.Service.Name,
				"-o", orgName,
				"-p", serviceBroker.SyncPlans[0].Name)
			Eventually(session).Should(Exit(0))

			session = helpers.CF("service-access")
			Eventually(session).Should(Say("%s\\s+%s\\s+limited\\s+%s",
				serviceBroker.Service.Name,
				serviceBroker.SyncPlans[0].Name,
				orgName))
			Eventually(session).Should(Exit(0))

			// enable access for all
			session = helpers.CF("enable-service-access", serviceBroker.Service.Name)
			Eventually(session).Should(Exit(0))

			session = helpers.CF("service-access", "-e", serviceBroker.Service.Name)
			Eventually(session).Should(Say("%s\\s+%s\\s+all",
				serviceBroker.Service.Name,
				serviceBroker.SyncPlans[0].Name,
			))
			Eventually(session).Should(Exit(0))

			// disable access
			session = helpers.CF("disable-service-access",
				serviceBroker.Service.Name,
				"-p", serviceBroker.SyncPlans[0].Name,
			)
			Eventually(session).Should(Exit(0))

			session = helpers.CF("service-access", "-b", serviceBroker.Name)
			Eventually(session).Should(Say("%s\\s+%s\\s+none",
				serviceBroker.Service.Name,
				serviceBroker.SyncPlans[0].Name,
			))
			Eventually(session).Should(Exit(0))
		})
	})
})
