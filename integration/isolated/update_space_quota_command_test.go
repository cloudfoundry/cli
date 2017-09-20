package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("update-space-quota command", func() {
	var (
		orgName   string
		spaceName string
		quotaName string
	)
	BeforeEach(func() {
		orgName = helpers.NewOrgName()
		spaceName = helpers.NewSpaceName()

		setupCF(orgName, spaceName)
		quotaName = helpers.QuotaName()
		totalMemory := "24M"
		instanceMemory := "6M"
		routes := "8"
		serviceInstances := "2"
		appInstances := "3"
		reservedRoutePorts := "1"
		session := helpers.CF("create-space-quota", quotaName, "-m", totalMemory, "-i", instanceMemory, "-r", routes, "-s", serviceInstances, "-a", appInstances, "--allow-paid-service-plans", "--reserved-route-ports", reservedRoutePorts)
		Eventually(session).Should(Exit(0))
	})

	AfterEach(func() {
		helpers.QuickDeleteOrg(orgName)
	})

	It("updates a space quota", func() {
		totalMemory := "25M"
		instanceMemory := "5M"
		serviceInstances := "1"
		appInstances := "2"
		reservedRoutePorts := "0"
		session := helpers.CF("update-space-quota", quotaName, "-m", totalMemory, "-i", instanceMemory, "-s", serviceInstances, "-a", appInstances, "--allow-paid-service-plans", "--reserved-route-ports", reservedRoutePorts)
		Eventually(session).Should(Say("Updating space quota %s", quotaName))
		Eventually(session).Should(Exit(0))

		session = helpers.CF("space-quota", quotaName)
		Eventually(session).Should(Say("total memory limit\\s+%s", totalMemory))
		Eventually(session).Should(Say("instance memory limit\\s+%s", instanceMemory))
		Eventually(session).Should(Say("routes\\s+%s", "8"))
		Eventually(session).Should(Say("services\\s+%s", serviceInstances))
		//TODO: Uncomment when #134821331 is complete
		// Eventually(session).Should(Say("Paid service plans\\s+%s", "allowed"))
		Eventually(session).Should(Say("app instance limit\\s+%s", appInstances))
		Eventually(session).Should(Say("reserved route ports\\s+%s", reservedRoutePorts))
		Eventually(session).Should(Exit(0))
	})
})
