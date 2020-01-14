package global

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("org-quotas command", func() {
	var (
		quotaName string

		totalMemory        string
		instanceMemory     string
		routes             string
		serviceInstances   string
		appInstances       string
		reservedRoutePorts string
	)

	Describe("help", func() {
		When("--help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("org-quotas", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("org-quotas - List available organization quotas"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf org-quotas"))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("org-quota"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(false, false, ReadOnlyOrg, "org", "org-name")
		})
	})

	When("the environment is set up correctly", func() {
		BeforeEach(func() {
			helpers.LoginCF()
			helpers.TargetOrgAndSpace(ReadOnlyOrg, ReadOnlySpace)
			quotaName = helpers.QuotaName()
			totalMemory = "24M"
			instanceMemory = "6M"
			routes = "8"
			serviceInstances = "2"
			appInstances = "3"
			reservedRoutePorts = "1"
			session := helpers.CF("create-quota", quotaName, "-m", totalMemory, "-i", instanceMemory, "-r", routes, "-s", serviceInstances, "-a", appInstances, "--allow-paid-service-plans", "--reserved-route-ports", reservedRoutePorts)
			Eventually(session).Should(Exit(0))
		})

		It("lists the org quotas", func() {
			session := helpers.CF("org-quotas")
			Eventually(session).Should(Say(`name\s+total memory\s+instance memory\s+routes\s+service instances\s+paid service plans\s+app instances\s+route ports`))
			Eventually(session).Should(Say(`%s\s+24\s+6\s+%s\s+%s\s+%s\s+%s\s+%s`, quotaName, routes, serviceInstances, "allowed", appInstances, reservedRoutePorts))
			Eventually(session).Should(Exit(0))
		})
	})
})
