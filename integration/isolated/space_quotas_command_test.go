package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("space-quotas command", func() {
	var (
		quotaName string

		totalMemory        string
		instanceMemory     string
		routes             string
		serviceInstances   string
		appInstances       string
		reservedRoutePorts string
	)
	BeforeEach(func() {
		setupCF(ReadOnlyOrg, ReadOnlySpace)
		quotaName = helpers.QuotaName()
		totalMemory = "24M"
		instanceMemory = "6M"
		routes = "8"
		serviceInstances = "2"
		appInstances = "3"
		reservedRoutePorts = "1"
		session := helpers.CF("create-space-quota", quotaName, "-m", totalMemory, "-i", instanceMemory, "-r", routes, "-s", serviceInstances, "-a", appInstances, "--allow-paid-service-plans", "--reserved-route-ports", reservedRoutePorts)
		Eventually(session).Should(Exit(0))
	})

	It("lists the space quotas", func() {
		session := helpers.CF("space-quotas")
		Eventually(session).Should(Say("name\\s+total memory\\s+instance memory\\s+routes\\s+service instances\\s+paid plans\\s+app instances\\s+route ports"))
		Eventually(session).Should(Say("%s\\s+%s\\s+%s\\s+%s\\s+%s\\s+%s\\s+%s\\s+%s", quotaName, totalMemory, instanceMemory, routes, serviceInstances, "allowed", appInstances, reservedRoutePorts))
		Eventually(session).Should(Exit(0))
	})
})
