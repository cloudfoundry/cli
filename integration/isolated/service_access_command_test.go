package isolated

import (
	. "code.cloudfoundry.org/cli/integration/helpers"
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

			serviceBroker ServiceBroker
		)

		BeforeEach(func() {
			orgName = NewOrgName()
			spaceName = PrefixedRandomName("space")
			setupCF(orgName, spaceName)

			serviceBroker = NewServiceBroker(
				PrefixedRandomName("broker"),
				NewAssets().ServiceBroker,
				defaultSharedDomain(),
				PrefixedRandomName("service"),
				PrefixedRandomName("plan"),
			)

			serviceBroker.Push()
			serviceBroker.Configure()
			serviceBroker.Create()
		})

		It("sets visibility", func() {
			// initial access is none
			session := CF("service-access")
			Eventually(session).Should(Exit(0))
			Expect(session).To(Say("%s\\s+%s\\s+none",
				serviceBroker.Service.Name,
				serviceBroker.SyncPlans[0].Name,
			))

			// enable access for org and plan
			session = CF("enable-service-access",
				serviceBroker.Service.Name,
				"-o", orgName,
				"-p", serviceBroker.SyncPlans[0].Name)
			Eventually(session).Should(Exit(0))

			session = CF("service-access")
			Eventually(session).Should(Exit(0))
			Expect(session).To(Say("%s\\s+%s\\s+limited\\s+%s",
				serviceBroker.Service.Name,
				serviceBroker.SyncPlans[0].Name,
				orgName))

			// enable access for all
			session = CF("enable-service-access", serviceBroker.Service.Name)
			Eventually(session).Should(Exit(0))

			session = CF("service-access", "-e", serviceBroker.Service.Name)
			Eventually(session).Should(Exit(0))
			Expect(session).To(Say("%s\\s+%s\\s+all",
				serviceBroker.Service.Name,
				serviceBroker.SyncPlans[0].Name,
			))

			// disable access
			session = CF("disable-service-access",
				serviceBroker.Service.Name,
				"-p", serviceBroker.SyncPlans[0].Name,
			)
			Eventually(session).Should(Exit(0))

			session = CF("service-access", "-b", serviceBroker.Name)
			Eventually(session).Should(Exit(0))
			Expect(session).To(Say("%s\\s+%s\\s+none",
				serviceBroker.Service.Name,
				serviceBroker.SyncPlans[0].Name,
			))
		})
	})
})
