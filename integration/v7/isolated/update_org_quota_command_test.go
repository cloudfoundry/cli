package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("update-org-quota command", func() {
	var (
		orgName   string
		spaceName string
		quotaName string
		username  string
	)
	BeforeEach(func() {
		orgName = helpers.NewOrgName()
		spaceName = helpers.NewSpaceName()
		username, _ = helpers.GetCredentials()

		helpers.SetupCF(orgName, spaceName)
		quotaName = helpers.QuotaName()
		totalMemory := "24M"
		instanceMemory := "6M"
		routes := "8"
		serviceInstances := "2"
		appInstances := "3"
		reservedRoutePorts := "1"
		session := helpers.CF("create-quota", quotaName, "-m", totalMemory, "-i", instanceMemory, "-r", routes, "-s", serviceInstances, "-a", appInstances, "--allow-paid-service-plans", "--reserved-route-ports", reservedRoutePorts)
		Eventually(session).Should(Exit(0))
	})

	AfterEach(func() {
		helpers.QuickDeleteOrg(orgName)
	})

	It("updates a quota", func() {
		totalMemory := "25M"
		instanceMemory := "5M"
		serviceInstances := "1"
		appInstances := "2"
		reservedRoutePorts := "0"
		session := helpers.CF("update-org-quota", quotaName, "-m", totalMemory, "-i", instanceMemory, "-s", serviceInstances, "-a", appInstances, "--allow-paid-service-plans", "--reserved-route-ports", reservedRoutePorts)
		Eventually(session).Should(Say(`Updating quota %s as %s\.\.\.`, quotaName, username))
		Eventually(session).Should(Exit(0))

		session = helpers.CF("quota", quotaName)
		Eventually(session).Should(Say(`Total Memory\s+%s`, totalMemory))
		Eventually(session).Should(Say(`Instance Memory\s+%s`, instanceMemory))
		Eventually(session).Should(Say(`Routes\s+%s`, "8"))
		Eventually(session).Should(Say(`Services\s+%s`, serviceInstances))
		Eventually(session).Should(Say(`Paid service plans\s+%s`, "allowed"))
		Eventually(session).Should(Say(`App instance limit\s+%s`, appInstances))
		Eventually(session).Should(Say(`Reserved Route Ports\s+%s`, reservedRoutePorts))
		Eventually(session).Should(Say("OK"))
		Eventually(session).Should(Exit(0))
	})

	When("the -n rename flag is provided", func() {
		var (
			newQuotaName string
		)

		BeforeEach(func() {
			newQuotaName = helpers.QuotaName()
		})
		It("renames the quota", func() {
			session := helpers.CF("update-org-quota", quotaName, "-n", newQuotaName)
			Eventually(session).Should(Say(`Updating quota %s as %s\.\.\.`, quotaName, username))
			Eventually(session.Err).Should(Say("CAPI ERROR: Organization quota %s already exists", quotaName))
			Eventually(session).Should(Say("FAILED"))
			Eventually(session).Should(Exit(1))

		})

		When("an org quota with the new name already exists", func() {

		})
	})
	When("the named quota does not exist", func() {
		It("displays a missing quota error message and fails", func() {
			session := helpers.CF("update-org-quota", "bogus-org-quota")
			Eventually(session).Should(Say(`Updating quota %s as %s\.\.\.`, quotaName, username))
			Eventually(session).Should(Say("FAILED"))
			Eventually(session.Err).Should(Say(`Quota %s not found`, "bogus-org-quota"))
			Eventually(session).Should(Exit(1))
		})
	})

	When("no user-provided updates to the quota are specified", func() {
		It("behaves idempotently and succeeds", func() {
			session := helpers.CF("update-org-quota", quotaName)
			Eventually(session).Should(Say(`Updating quota %s as %s\.\.\.`, quotaName, username))
			Eventually(session).Should(Say("OK"))
			Eventually(session).Should(Exit(0))
		})
	})

	When("conflicting flags are specified", func() {
		It("displays a flag conflict error message and fails", func() {
			session := helpers.CF("update-org-quota", quotaName, "--allow-paid-service-plans", "--disallow-paid-service-plans")
			Eventually(session).Should(Say("FAILED"))
			Eventually(session.Err).Should(Say(`Please choose either allow or disallow. Both flags are not permitted to be passed in the same command.`))
			Eventually(session).Should(Exit(1))
		})
	})
})
