package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("create-quota command", func() {
	BeforeEach(func() {
		setupCF(ReadOnlyOrg, ReadOnlySpace)
	})

	It("creates a quota", func() {
		quotaName := helpers.QuotaName()
		totalMemory := "24M"
		instanceMemory := "6M"
		routes := "8"
		serviceInstances := "2"
		appInstances := "3"
		reservedRoutePorts := "1"
		session := helpers.CF("create-quota", quotaName, "-m", totalMemory, "-i", instanceMemory, "-r", routes, "-s", serviceInstances, "-a", appInstances, "--allow-paid-service-plans", "--reserved-route-ports", reservedRoutePorts)
		Eventually(session).Should(Say("Creating quota %s", quotaName))
		Eventually(session).Should(Exit(0))

		session = helpers.CF("quota", quotaName)
		Eventually(session).Should(Say("Total Memory\\s+%s", totalMemory))
		Eventually(session).Should(Say("Instance Memory\\s+%s", instanceMemory))
		Eventually(session).Should(Say("Routes\\s+%s", routes))
		Eventually(session).Should(Say("Services\\s+%s", serviceInstances))
		Eventually(session).Should(Say("Paid service plans\\s+%s", "allowed"))
		Eventually(session).Should(Say("App instance limit\\s+%s", appInstances))
		Eventually(session).Should(Say("Reserved Route Ports\\s+%s", reservedRoutePorts))
		Eventually(session).Should(Exit(0))
	})
})
