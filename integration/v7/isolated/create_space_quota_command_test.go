package isolated

import (
	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("create-space-quota command", func() {
	var (
		userName       string
		orgName        string
		spaceQuotaName string
	)

	Describe("help", func() {
		When("--help flag is set", func() {
			It("appears in cf help -a", func() {
				session := helpers.CF("help", "-a")
				Eventually(session).Should(Exit(0))
				Expect(session).To(HaveCommandInCategoryWithDescription("create-space-quota", "SPACE ADMIN", "Define a new quota for a space"))
			})

			It("Displays command usage to output", func() {
				session := helpers.CF("create-space-quota", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("create-space-quota - Define a new quota for a space"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say(`cf create-space-quota QUOTA \[-m TOTAL_MEMORY\] \[-i INSTANCE_MEMORY\] \[-r ROUTES\] \[-s SERVICE_INSTANCES\] \[-a APP_INSTANCES\] \[--allow-paid-service-plans\] \[--reserved-route-ports RESERVED_ROUTE_PORTS\]`))
				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say(`-a\s+Total number of application instances. \(Default: unlimited\)`))
				Eventually(session).Should(Say(`--allow-paid-service-plans\s+Allow provisioning instances of paid service plans. \(Default: disallowed\)`))
				Eventually(session).Should(Say(`-i\s+Maximum amount of memory a process can have \(e.g. 1024M, 1G, 10G\). \(Default: unlimited\)`))
				Eventually(session).Should(Say(`-m\s+Total amount of memory all processes can have \(e.g. 1024M, 1G, 10G\). -1 represents an unlimited amount. \(Default: 0\)`))
				Eventually(session).Should(Say(`-r\s+Total number of routes. -1 represents an unlimited amount. \(Default: 0\)`))
				Eventually(session).Should(Say(`--reserved-route-ports\s+Maximum number of routes that may be created with ports. -1 represents an unlimited amount. \(Default: 0\)`))
				Eventually(session).Should(Say(`-s\s+Total number of service instances. -1 represents an unlimited amount. \(Default: 0\)`))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("create-space, set-space-quota, space-quotas"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, false, orgName, "create-space-quota", spaceQuotaName)
		})
	})

	When("the environment is set up correctly", func() {
		BeforeEach(func() {
			userName = helpers.LoginCF()
			orgName = helpers.CreateAndTargetOrg()
			spaceQuotaName = helpers.QuotaName()
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		When("the quota name is not provided", func() {
			It("tells the user that the quota name is required, prints help text, and exits 1", func() {
				session := helpers.CF("create-space-quota")

				Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `SPACE_QUOTA_NAME` was not provided"))
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Exit(1))
			})
		})

		When("a nonexistent flag is provided", func() {
			It("tells the user that the flag is invalid", func() {
				session := helpers.CF("create-space-quota", "--nonexistent")

				Eventually(session.Err).Should(Say("Incorrect Usage: unknown flag `nonexistent'"))
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the quota does not exist", func() {
			When("no flags are provided", func() {
				It("creates the quota with the default values", func() {
					session := helpers.CF("create-space-quota", spaceQuotaName)
					Eventually(session).Should(Say("Creating space quota %s for org %s as %s...", spaceQuotaName, orgName, userName))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))

					session = helpers.CF("space-quota", spaceQuotaName)
					Eventually(session).Should(Say("Getting space quota %s for org %s as %s...", spaceQuotaName, orgName, userName))
					Eventually(session).Should(Say(`total memory:\s+0`))
					Eventually(session).Should(Say(`instance memory:\s+unlimited`))
					Eventually(session).Should(Say(`routes:\s+0`))
					Eventually(session).Should(Say(`service instances:\s+0`))
					Eventually(session).Should(Say(`paid service plans:\s+disallowed`))
					Eventually(session).Should(Say(`app instances:\s+unlimited`))
					Eventually(session).Should(Say(`route ports:\s+0`))
					Eventually(session).Should(Exit(0))
				})
			})

			When("all the optional flags are provided", func() {
				It("creates the quota with the specified values", func() {
					session := helpers.CF("create-space-quota", spaceQuotaName,
						"-a", "2",
						"--allow-paid-service-plans",
						"-i", "3M",
						"-m", "4M",
						"-r", "15",
						"--reserved-route-ports", "6",
						"-s", "7")
					Eventually(session).Should(Say("Creating space quota %s for org %s as %s...", spaceQuotaName, orgName, userName))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))

					session = helpers.CF("space-quota", spaceQuotaName)
					Eventually(session).Should(Say("Getting space quota %s for org %s as %s...", spaceQuotaName, orgName, userName))
					Eventually(session).Should(Say(`total memory:\s+4M`))
					Eventually(session).Should(Say(`instance memory:\s+3M`))
					Eventually(session).Should(Say(`routes:\s+15`))
					Eventually(session).Should(Say(`service instances:\s+7`))
					Eventually(session).Should(Say(`paid service plans:\s+allowed`))
					Eventually(session).Should(Say(`app instances:\s+2`))
					Eventually(session).Should(Say(`route ports:\s+6`))
					Eventually(session).Should(Exit(0))
				})
			})
		})

		When("the quota already exists", func() {
			BeforeEach(func() {
				Eventually(helpers.CF("create-space-quota", spaceQuotaName)).Should(Exit(0))
			})

			It("notifies the user that the quota already exists and exits 0", func() {
				session := helpers.CF("create-space-quota", spaceQuotaName)
				Eventually(session).Should(Say("Creating space quota %s for org %s as %s...", spaceQuotaName, orgName, userName))
				Eventually(session.Err).Should(Say(`Space Quota '%s' already exists\.`, spaceQuotaName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
