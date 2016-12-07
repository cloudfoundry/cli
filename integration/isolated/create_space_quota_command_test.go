package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("create-space-quota command", func() {
	BeforeEach(func() {
		setupCF(ReadOnlyOrg, ReadOnlySpace)
	})

	It("creates a space quota", func() {
		quotaName := helpers.QuotaName()
		totalMemory := "24M"
		instanceMemory := "6M"
		routes := "8"
		serviceInstances := "2"
		appInstances := "3"
		reservedRoutePorts := "1"
		session := helpers.CF("create-space-quota", quotaName, "-m", totalMemory, "-i", instanceMemory, "-r", routes, "-s", serviceInstances, "-a", appInstances, "--allow-paid-service-plans", "--reserved-route-ports", reservedRoutePorts)
		Eventually(session).Should(Say("Creating space quota %s", quotaName))
		Eventually(session).Should(Exit(0))

		session = helpers.CF("space-quota", quotaName)
		Eventually(session).Should(Say("total memory limit\\s+%s", totalMemory))
		Eventually(session).Should(Say("instance memory limit\\s+%s", instanceMemory))
		Eventually(session).Should(Say("routes\\s+%s", routes))
		Eventually(session).Should(Say("services\\s+%s", serviceInstances))
		//TODO: uncomment when #134821331 is complete
		//Eventually(session).Should(Say("paid service plans\\s+%s", "allowed"))
		Eventually(session).Should(Say("app instance limit\\s+%s", appInstances))
		Eventually(session).Should(Say("reserved route ports\\s+%s", reservedRoutePorts))
		Eventually(session).Should(Exit(0))
	})
})
