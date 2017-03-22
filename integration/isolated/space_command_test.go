package isolated

import (
	"fmt"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("space command", func() {
	var (
		orgName   string
		spaceName string
	)

	BeforeEach(func() {
		helpers.RunIfExperimental("space command refactor is still experimental")

		orgName = helpers.NewOrgName()
		spaceName = helpers.NewSpaceName()
	})

	Describe("help", func() {
		Context("when --help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("space", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("space - Show space info"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf space SPACE \\[--guid\\] \\[--security-group-rules\\]"))
				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say("--guid\\s+Retrieve and display the given space's guid\\.  All other output for the space is suppressed\\."))
				Eventually(session).Should(Say("--security-group-rules\\s+Retrieve the rules for all the security groups associated with the space\\."))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("space-users"))
			})
		})
	})

	Context("when the environment is not setup correctly", func() {
		Context("when no API endpoint is set", func() {
			BeforeEach(func() {
				helpers.UnsetAPI()
			})

			It("fails with no API endpoint set message", func() {
				session := helpers.CF("space", "some-space")
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("No API endpoint set. Use 'cf login' or 'cf api' to target an endpoint."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when not logged in", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
			})

			It("fails with not logged in message", func() {
				session := helpers.CF("space", "some-space")
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("Not logged in. Use 'cf login' to log in."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when no organization is targeted", func() {
			BeforeEach(func() {
				helpers.LoginCF()
			})

			AfterEach(func() {
				helpers.LogoutCF()
			})

			It("fails with no organization targeted message and exits 1", func() {
				session := helpers.CF("space", spaceName)
				_, _ = helpers.GetCredentials()
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("No org targeted, use 'cf target -o ORG' to target an org."))
				Eventually(session).Should(Exit(1))
			})
		})
	})

	Context("when the environment is set up correctly", func() {
		BeforeEach(func() {
			helpers.LoginCF()
			setupCF(orgName, spaceName)
		})

		AfterEach(func() {
			helpers.ClearTarget()
		})

		Context("when the space does not exist", func() {
			It("displays not found and exits 1", func() {
				badSpaceName := fmt.Sprintf("%s-1", spaceName)
				session := helpers.CF("space", badSpaceName)
				userName, _ := helpers.GetCredentials()
				Eventually(session.Out).Should(Say("Getting info for space %s in org %s as %s...", badSpaceName, orgName, userName))
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("Space '%s' not found.", badSpaceName))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the space exists", func() {
			Context("when the --guid flag is used", func() {
				It("displays the space guid", func() {
					session := helpers.CF("space", "--guid", spaceName)
					Eventually(session.Out).Should(Say("[\\da-f]{8}-[\\da-f]{4}-[\\da-f]{4}-[\\da-f]{4}-[\\da-f]{12}"))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when no flags are used", func() {
				var (
					appName              string
					spaceQuotaName       string
					serviceInstance      string
					isolationSegmentName string
				)

				BeforeEach(func() {
					appName = helpers.PrefixedRandomName("app")
					helpers.WithHelloWorldApp(func(appDir string) {
						Eventually(helpers.CF("push", appName, "-p", appDir, "-b", "staticfile_buildpack")).Should(Exit(0))
					})
					serviceInstance = helpers.PrefixedRandomName("si")
					Eventually(helpers.CF("create-user-provided-service", serviceInstance, "-p", "{}")).Should(Exit(0))
					Eventually(helpers.CF("bind-service", appName, serviceInstance)).Should(Exit(0))
					spaceQuotaName = helpers.PrefixedRandomName("space-quota")
					Eventually(helpers.CF("create-space-quota", spaceQuotaName)).Should(Exit(0))
					Eventually(helpers.CF("set-space-quota", spaceName, spaceQuotaName)).Should(Exit(0))
					isolationSegmentName = helpers.IsolationSegmentName()
					Eventually(helpers.CF("create-isolation-segment", isolationSegmentName)).Should(Exit(0))
					Eventually(helpers.CF("enable-org-isolation", orgName, isolationSegmentName)).Should(Exit(0))
					Eventually(helpers.CF("set-space-isolation-segment", spaceName, isolationSegmentName)).Should(Exit(0))
				})

				It("displays a table with space name, org, apps, services, isolation segment, space quota and security groups", func() {
					session := helpers.CF("space", spaceName)
					userName, _ := helpers.GetCredentials()
					Eventually(session.Out).Should(Say("Getting info for space %s in org %s as %s...", spaceName, orgName, userName))

					Eventually(session.Out).Should(Say("name:\\s+%s", spaceName))
					Eventually(session.Out).Should(Say("org:\\s+%s", orgName))
					Eventually(session.Out).Should(Say("apps:\\s+%s", appName))
					Eventually(session.Out).Should(Say("services:\\s+%s", serviceInstance))
					Eventually(session.Out).Should(Say("isolation segment:\\s+%s", isolationSegmentName))
					Eventually(session.Out).Should(Say("space quota:\\s+%s", spaceQuotaName))
					Eventually(session.Out).Should(Say("security groups:\\s+dns, load_balancer, public_networks"))
				})
			})

			Context("when the security group rules flag is used", func() {
				It("displays the space information as well as all security group rules", func() {
					session := helpers.CF("space", "--security-group-rules", spaceName)
					userName, _ := helpers.GetCredentials()
					Eventually(session.Out).Should(Say("Getting info for space %s in org %s as %s...", spaceName, orgName, userName))

					Eventually(session.Out).Should(Say("name:\\s+%s", spaceName))
					Eventually(session.Out).Should(Say("org:\\s+%s", orgName))
					Eventually(session.Out).Should(Say("apps:"))
					Eventually(session.Out).Should(Say("services:"))
					Eventually(session.Out).Should(Say("isolation segment:"))
					Eventually(session.Out).Should(Say("space quota:"))
					Eventually(session.Out).Should(Say("security groups:\\s+dns, load_balancer, public_networks"))
					Eventually(session.Out).Should(Say("\n\n"))

					Eventually(session.Out).Should(Say("security group\\s+destination\\s+ports\\s+protocol\\s+lifecycle\\s+description"))
					Eventually(session.Out).Should(Say("#0\\s+dns\\s+0.0.0.0/0\\s+53\\s+tcp\\s+running"))
					Eventually(session.Out).Should(Say("\\s+dns\\s+0.0.0.0/0\\s+53\\s+udp\\s+running"))
					Eventually(session.Out).Should(Say("\\s+dns\\s+0.0.0.0/0\\s+53\\s+tcp\\s+staging"))
					Eventually(session.Out).Should(Say("\\s+dns\\s+0.0.0.0/0\\s+53\\s+udp\\s+staging"))

					Eventually(session.Out).Should(Say("#1\\s+load_balancer\\s+10.244.0.34\\s+all\\s+running"))

					Eventually(session.Out).Should(Say("#2\\s+public_networks\\s+0.0.0.0-9.255.255.255\\s+all\\s+running"))
					Eventually(session.Out).Should(Say("\\s+public_networks\\s+0.0.0.0-9.255.255.255\\s+all\\s+staging"))
					Eventually(session.Out).Should(Say("\\s+public_networks\\s+11.0.0.0-169.253.255.255\\s+all\\s+running"))
					Eventually(session.Out).Should(Say("\\s+public_networks\\s+11.0.0.0-169.253.255.255\\s+all\\s+staging"))
					Eventually(session.Out).Should(Say("\\s+public_networks\\s+169.255.0.0-172.15.255.255\\s+all\\s+running"))
					Eventually(session.Out).Should(Say("\\s+public_networks\\s+169.255.0.0-172.15.255.255\\s+all\\s+staging"))
					Eventually(session.Out).Should(Say("\\s+public_networks\\s+172.32.0.0-192.167.255.255\\s+all\\s+running"))
					Eventually(session.Out).Should(Say("\\s+public_networks\\s+172.32.0.0-192.167.255.255\\s+all\\s+staging"))
					Eventually(session.Out).Should(Say("\\s+public_networks\\s+192.169.0.0-255.255.255.255\\s+all\\s+running"))
					Eventually(session.Out).Should(Say("\\s+public_networks\\s+192.169.0.0-255.255.255.255\\s+all\\s+staging"))

				})
			})
		})
	})
})
