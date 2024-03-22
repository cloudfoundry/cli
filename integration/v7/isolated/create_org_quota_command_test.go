package isolated

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("create-org-quota command", func() {
	var (
		orgQuotaName string
	)

	BeforeEach(func() {
		_ = helpers.LoginCF()
		orgQuotaName = helpers.NewOrgQuotaName()
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
				Eventually(session).Should(Say(`cf create-org-quota ORG_QUOTA \[-m TOTAL_MEMORY\] \[-i INSTANCE_MEMORY\] \[-r ROUTES\] \[-s SERVICE_INSTANCES\] \[-a APP_INSTANCES\] \[--allow-paid-service-plans\] \[--reserved-route-ports RESERVED_ROUTE_PORTS\] \[-l LOG_VOLUME\]`))
				Eventually(session).Should(Say("ALIAS:"))
				Eventually(session).Should(Say("create-quota"))
				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say(`-a\s+Total number of application instances. \(Default: unlimited\)`))
				Eventually(session).Should(Say(`--allow-paid-service-plans\s+Allow provisioning instances of paid service plans. \(Default: disallowed\)`))
				Eventually(session).Should(Say(`-i\s+Maximum amount of memory a process can have \(e.g. 1024M, 1G, 10G\). \(Default: unlimited\)`))
				Eventually(session).Should(Say(`-m\s+Total amount of memory all processes can have \(e.g. 1024M, 1G, 10G\).  -1 represents an unlimited amount. \(Default: 0\)`))
				Eventually(session).Should(Say(`-r\s+Total number of routes. -1 represents an unlimited amount. \(Default: 0\)`))
				Eventually(session).Should(Say(`--reserved-route-ports\s+Maximum number of routes that may be created with ports. -1 represents an unlimited amount. \(Default: 0\)`))
				Eventually(session).Should(Say(`-s\s+Total number of service instances. -1 represents an unlimited amount. \(Default: 0\)`))
				Eventually(session).Should(Say(`-l\s+Total log volume per second all processes can have, in bytes \(e.g. 128B, 4K, 1M\). -1 represents an unlimited amount. \(Default: -1\)`))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("create-org, org-quotas, set-org-quota"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(false, false, ReadOnlyOrg, "create-org-quota", orgQuotaName)
		})
	})

	When("the environment is set up correctly", func() {
		When("the quota name is not provided", func() {
			It("tells the user that the quota name is required, prints help text, and exits 1", func() {
				session := helpers.CF("create-org-quota")

				Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `ORG_QUOTA_NAME` was not provided"))
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Exit(1))
			})
		})

		When("a nonexistent flag is provided", func() {
			It("tells the user that the flag is invalid", func() {
				session := helpers.CF("create-org-quota", "--nonexistent")

				Eventually(session.Err).Should(Say("Incorrect Usage: unknown flag `nonexistent'"))
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Exit(1))
			})

		})

		When("an invalid flag value is provided", func() {
			It("tells the user that the flag value is invalid", func() {
				session := helpers.CF("create-org-quota", orgQuotaName,
					"-a", "hello")

				Eventually(session.Err).Should(Say(`Incorrect Usage: invalid integer limit \(expected int >= -1\)`))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the quota does not exist", func() {
			When("no flags are provided", func() {
				It("creates the quota with the default values", func() {
					session := helpers.CF("create-org-quota", orgQuotaName)
					userName, _ := helpers.GetCredentials()
					Eventually(session).Should(Say("Creating org quota %s as %s...", orgQuotaName, userName))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))

					session = helpers.CF("org-quota", orgQuotaName)
					Eventually(session).Should(Say("Getting org quota %s as %s...", orgQuotaName, userName))
					Eventually(session).Should(Say(`total memory:\s+0`))
					Eventually(session).Should(Say(`instance memory:\s+unlimited`))
					Eventually(session).Should(Say(`service instances:\s+0`))
					Eventually(session).Should(Say(`paid service plans:\s+disallowed`))
					Eventually(session).Should(Say(`app instances:\s+unlimited`))
					Eventually(session).Should(Say(`route ports:\s+0`))
					Eventually(session).Should(Exit(0))
				})
			})

			When("all the optional flags are provided", func() {
				It("creates the quota with the specified values", func() {
					userName, _ := helpers.GetCredentials()
					session := helpers.CF("create-org-quota", orgQuotaName,
						"-a", "2",
						"--allow-paid-service-plans",
						"-i", "2M",
						"-m", "4M",
						"-r", "6",
						"--reserved-route-ports", "5",
						"-s", "7",
					)
					Eventually(session).Should(Say("Creating org quota %s as %s...", orgQuotaName, userName))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))

					session = helpers.CF("org-quota", orgQuotaName)
					Eventually(session).Should(Say("Getting org quota %s as %s...", orgQuotaName, userName))
					Eventually(session).Should(Say(`total memory:\s+4M`))
					Eventually(session).Should(Say(`instance memory:\s+2M`))
					Eventually(session).Should(Say(`routes:\s+6`))
					Eventually(session).Should(Say(`service instances:\s+7`))
					Eventually(session).Should(Say(`paid service plans:\s+allowed`))
					Eventually(session).Should(Say(`app instances:\s+2`))
					Eventually(session).Should(Say(`route ports:\s+5`))
					Eventually(session).Should(Say(`log volume per second:\s+unlimited`))
					Eventually(session).Should(Exit(0))
				})

				When("CAPI supports log rate limits", func() {
					BeforeEach(func() {
						helpers.SkipIfVersionLessThan(ccversion.MinVersionLogRateLimitingV3)
					})

					It("creates the quota with the specified values", func() {

						userName, _ := helpers.GetCredentials()
						session := helpers.CF("create-org-quota", orgQuotaName,
							"-l", "32K",
						)
						Eventually(session).Should(Say("Creating org quota %s as %s...", orgQuotaName, userName))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Exit(0))

						session = helpers.CF("org-quota", orgQuotaName)
						Eventually(session).Should(Say(`log volume per second:\s+32K`))
						Eventually(session).Should(Exit(0))
					})
				})
			})

			When("The flags are all set to -1", func() {
				It("creates the quota with the specified values", func() {
					userName, _ := helpers.GetCredentials()
					session := helpers.CF("create-org-quota", orgQuotaName,
						"-a", "-1",
						"-i", "-1",
						"-m", "-1",
						"-r", "-1",
						"-s", "-1",
						"--reserved-route-ports", "-1",
					)
					Eventually(session).Should(Say("Creating org quota %s as %s...", orgQuotaName, userName))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))

					session = helpers.CF("org-quota", orgQuotaName)
					Eventually(session).Should(Say(`total memory:\s+unlimited`))
					Eventually(session).Should(Say(`instance memory:\s+unlimited`))
					Eventually(session).Should(Say(`routes:\s+unlimited`))
					Eventually(session).Should(Say(`service instances:\s+unlimited`))
					Eventually(session).Should(Say(`app instances:\s+unlimited`))
					Eventually(session).Should(Say(`route ports:\s+unlimited`))
					Eventually(session).Should(Say(`log volume per second:\s+unlimited`))
					Eventually(session).Should(Exit(0))
				})

				When("CAPI supports log rate limits", func() {
					BeforeEach(func() {
						helpers.SkipIfVersionLessThan(ccversion.MinVersionLogRateLimitingV3)
					})

					It("creates the quota with the specified values", func() {
						userName, _ := helpers.GetCredentials()
						session := helpers.CF("create-org-quota", orgQuotaName,
							"-l", "-1",
						)
						Eventually(session).Should(Say("Creating org quota %s as %s...", orgQuotaName, userName))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Exit(0))

						session = helpers.CF("org-quota", orgQuotaName)
						Eventually(session).Should(Say(`log volume per second:\s+unlimited`))
						Eventually(session).Should(Exit(0))
					})
				})
			})
		})

		When("the quota already exists", func() {
			BeforeEach(func() {
				Eventually(helpers.CF("create-org-quota", orgQuotaName)).Should(Exit(0))
			})

			It("notifies the user that the quota already exists and exits 0", func() {
				session := helpers.CF("create-org-quota", orgQuotaName)
				userName, _ := helpers.GetCredentials()
				Eventually(session).Should(Say("Creating org quota %s as %s...", orgQuotaName, userName))
				Eventually(session.Err).Should(Say(`Organization Quota '%s' already exists\.`, orgQuotaName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
