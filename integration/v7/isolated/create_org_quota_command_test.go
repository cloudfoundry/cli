package isolated

import (
	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("create-org-quota command", func() {
	var (
		quotaOrgName string
	)

	BeforeEach(func() {
		_ = helpers.LoginCF()
		quotaOrgName = helpers.NewOrgQuotaName()
	})

	AfterEach(func() {
		// TODO: delete `quotaOrgName` when the `delete-org-quota` is implemented
	})

	Describe("help", func() {
		When("--help flag is set", func() {
			It("appears in cf help -a", func() {
				session := helpers.CF("help", "-a")
				Eventually(session).Should(Exit(0))
				Expect(session).To(HaveCommandInCategoryWithDescription("create-org-quota", "ORG ADMIN", "Define a new quota for an organization"))
			})

			It("Displays command usage to output", func() {
				session := helpers.CF("create-org-quota", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("create-org-quota - Define a new quota for an organization"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say(`cf create-org-quota ORG_QUOTA \[-m TOTAL_MEMORY\] \[-i INSTANCE_MEMORY\] \[-r ROUTES\] \[-s SERVICE_INSTANCES\] \[-a APP_INSTANCES\] \[--allow-paid-service-plans\] \[--reserved-route-ports RESERVED_ROUTE_PORTS\]`))
				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say(`-a\s+Total number of application instances. \(Default: unlimited\)`))
				Eventually(session).Should(Say(`--allow-paid-service-plans\s+Allow provisioning instances of paid service plans. \(Default: disallowed\)`))
				Eventually(session).Should(Say(`-i\s+Maximum amount of memory a process can have \(e.g. 1024M, 1G, 10G\). \(Default: unlimited\)`))
				Eventually(session).Should(Say(`-m\s+Total amount of memory all processes can have \(e.g. 1024M, 1G, 10G\).  -1 represents an unlimited amount. \(Default: 0\)`))
				Eventually(session).Should(Say(`-r\s+Total number of routes. -1 represents an unlimited amount. \(Default: 0\)`))
				Eventually(session).Should(Say(`--reserved-route-ports\s+Maximum number of routes that may be created with ports. -1 represents an unlimited amount. \(Default: 0\)`))
				Eventually(session).Should(Say(`-s\s+Total number of service instances. -1 represents an unlimited amount. \(Default: 0\)`))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("create-org, org-quotas, set-org-quota"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the quota name is not provided", func() {
		It("tells the user that the quota name is required, prints help text, and exits 1", func() {
			session := helpers.CF("create-org-quota")

			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `ORG_QUOTA_NAME` was not provided"))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(false, false, ReadOnlyOrg, "create-org-quota", quotaOrgName)
		})
	})

	When("the environment is set up correctly", func() {
		When("the quota does not exist", func() {
			When("no flags are provided", func() {
				It("creates the quota with the default values", func() {
					session := helpers.CF("create-org-quota", quotaOrgName)
					userName, _ := helpers.GetCredentials()
					Eventually(session).Should(Say("Creating org quota %s as %s...", quotaOrgName, userName))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))

					// TODO: find sustainable way to retrieve the created-quota to make sure it exists
					// session = helpers.CF("org-quota", quotaOrgName)
					//Eventually(session).Should(Say("Getting quota %s info as %s...", quotaOrgName, userName))
					// The following defaults are correct, but the ordering and naming may change; they were copied from V6
					//Eventually(session).Should(Say(`Total Memory\s+0`))
					//Eventually(session).Should(Say(`Instance Memory\s+unlimited`))
					//Eventually(session).Should(Say(`Routes\s+0`))
					//Eventually(session).Should(Say(`Services\s+0`))
					//Eventually(session).Should(Say(`Paid service plans\s+disallowed`))
					//Eventually(session).Should(Say(`App instance limit\s+0`))
					//Eventually(session).Should(Say(`Reserved Route Ports\s+unlimited`))
					//Eventually(session).Should(Say("OK"))
					//Eventually(session).Should(Exit(0))
				})
			})

			When("all the optional flags are provided", func() {
				It("creates the quota with the specified values", func() {
					userName, _ := helpers.GetCredentials()
					session := helpers.CF("create-org-quota", quotaOrgName,
						"-a", "2",
						"--allow-paid-service-plans",
						"-i", "3M",
						"-m", "4M",
						"-r", "5",
						"--reserved-route-ports", "6",
						"-s", "7")
					Eventually(session).Should(Say("Creating org quota %s as %s...", quotaOrgName, userName))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))

					// TODO: find sustainable way to retrieve the created-quota to make sure it exists
					// session = helpers.CF("org-quota", quotaOrgName)
					//Eventually(session).Should(Say("Getting quota %s info as %s...", quotaOrgName, userName))
					// The following defaults are correct, but the ordering and naming may change; they were copied from V6
					//Eventually(session).Should(Say(`Total Memory\s+4M`))
					//Eventually(session).Should(Say(`Instance Memory\s+3M`))
					//Eventually(session).Should(Say(`Routes\s+5`))
					//Eventually(session).Should(Say(`Services\s+7`))
					//Eventually(session).Should(Say(`Paid service plans\s+allowed`))
					//Eventually(session).Should(Say(`App instance limit\s+2`))
					//Eventually(session).Should(Say(`Reserved Route Ports\s+6`))
					//Eventually(session).Should(Say("OK"))
					//Eventually(session).Should(Exit(0))
				})
			})
		})

		When("the quota already exists", func() {
			BeforeEach(func() {
				Eventually(helpers.CF("create-org-quota", quotaOrgName)).Should(Exit(0))
			})

			It("notifies the user that the quota already exists and exits 0", func() {
				session := helpers.CF("create-org-quota", quotaOrgName)
				userName, _ := helpers.GetCredentials()
				Eventually(session).Should(Say("Creating org quota %s as %s...", quotaOrgName, userName))
				Eventually(session).Should(Say(`Organization quota '%s' already exists\.`, quotaOrgName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
