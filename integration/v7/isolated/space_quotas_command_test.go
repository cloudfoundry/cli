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
		orgName   string
		userName  string

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
				session := helpers.CF("space-quotas", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("space-quotas - List available space quotas"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf space-quotas"))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("set-space-quota, space-quota"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, false, ReadOnlyOrg, "space-quotas")
		})
	})

	When("the environment is setup correctly", func() {
		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			userName = helpers.LoginCF()
			helpers.SetupCFWithOrgOnly(orgName)
			quotaName = helpers.QuotaName()
			totalMemory = "24M"
			instanceMemory = "6M"
			routes = "8"
			serviceInstances = "2"
			appInstances = "-1"
			reservedRoutePorts = "1"
			session := helpers.CF("create-space-quota", quotaName, "-m", totalMemory, "-i", instanceMemory, "-r", routes, "-s", serviceInstances, "-a", appInstances, "--allow-paid-service-plans", "--reserved-route-ports", reservedRoutePorts)
			Eventually(session).Should(Exit(0))
		})

		It("lists the space quotas", func() {
			session := helpers.CF("space-quotas")
			Eventually(session).Should(Say(`Getting space quotas for org %s as %s\.\.\.`, orgName, userName))
			Eventually(session).Should(Say(`name\s+total memory\s+instance memory\s+routes\s+service instances\s+paid service plans\s+app instances\s+route ports`))
			Eventually(session).Should(Say(`%s\s+%s\s+%s\s+%s\s+%s\s+%s\s+%s\s+%s`, quotaName, totalMemory, instanceMemory, routes, serviceInstances, "allowed", "unlimited", reservedRoutePorts))
			Eventually(session).Should(Exit(0))
		})

	})
})
