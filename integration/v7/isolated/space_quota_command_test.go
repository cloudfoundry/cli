package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("space-quota command", func() {
	var (
		quotaName string
		orgName   string
	)

	BeforeEach(func() {
		quotaName = helpers.QuotaName()
		orgName = helpers.NewOrgName()
	})

	Describe("help", func() {
		When("--help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("space-quota", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("space-quota - Show space quota info"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf space-quota QUOTA"))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("space, space-quotas"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, false, ReadOnlyOrg, "space-quota", quotaName)
		})
	})

	When("the environment is set up correctly", func() {
		var userName string

		BeforeEach(func() {
			userName = helpers.LoginCF()
			helpers.CreateOrg(orgName)
			helpers.TargetOrg(orgName)
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		When("the quota does not exist", func() {
			It("displays quota not found and exits 1", func() {
				session := helpers.CF("space-quota", quotaName)
				userName, _ := helpers.GetCredentials()
				Eventually(session).Should(Say(`Getting space quota %s for org %s as %s\.\.\.`, quotaName, orgName, userName))
				Eventually(session.Err).Should(Say("Space quota with name '%s' not found.", quotaName))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the quota exists", func() {
			BeforeEach(func() {
				session := helpers.CF(
					"create-space-quota",
					quotaName,
					"--allow-paid-service-plans",
					"-i", "6M",
					"-m", "7M",
					"-r", "8",
					"--reserved-route-ports", "0",
					"-s", "-1")
				Eventually(session).Should(Exit(0))
			})

			It("displays a table with quota names and their values and exits 0", func() {
				session := helpers.CF("space-quota", quotaName)

				Eventually(session).Should(Say(`Getting space quota %s for org %s as %s\.\.\.`, quotaName, orgName, userName))
				Eventually(session).Should(Say(`total memory:\s+7M`))
				Eventually(session).Should(Say(`instance memory:\s+6M`))
				Eventually(session).Should(Say(`routes:\s+8`))
				Eventually(session).Should(Say(`service instances:\s+unlimited`))
				Eventually(session).Should(Say(`paid service plans:\s+allowed`))
				Eventually(session).Should(Say(`app instances:\s+unlimited`))
				Eventually(session).Should(Say(`route ports:\s+0`))

				Eventually(session).Should(Exit(0))
			})
		})
	})
})
