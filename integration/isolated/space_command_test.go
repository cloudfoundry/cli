package isolated

import (
	"fmt"
	"io/ioutil"
	"os"

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
				Eventually(session).Should(Say("set-space-isolation-segment, space-quota, space-users"))
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
					securityGroupName    string
					securityGroupRules   *os.File
					err                  error
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

					securityGroupName = helpers.PrefixedRandomName("aaaaaaaaaaa")
					securityGroupRules, err = ioutil.TempFile("", "security-group-rules")
					Expect(err).ToNot(HaveOccurred())

					securityGroupRules.Write([]byte(`[]`))

					Eventually(helpers.CF("create-security-group", securityGroupName, securityGroupRules.Name())).Should(Exit(0))
					Eventually(helpers.CF("bind-security-group", securityGroupName, orgName, spaceName)).Should(Exit(0))
				})

				AfterEach(func() {
					Eventually(helpers.CF("delete-security-group", securityGroupName, "-f")).Should(Exit(0))
					os.Remove(securityGroupRules.Name())
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
					Eventually(session.Out).Should(Say("security groups:\\s+.*%s", securityGroupName))
				})
			})

			Context("when the security group rules flag is used", func() {
				var (
					securityGroupName   string
					securityGroupRules  *os.File
					securityGroupName2  string
					securityGroupRules2 *os.File
					err                 error
				)

				BeforeEach(func() {
					securityGroupName = helpers.PrefixedRandomName("aaaaaaaaaaa")
					securityGroupRules, err = ioutil.TempFile("", "security-group-rules")
					Expect(err).ToNot(HaveOccurred())

					securityGroupRules.Write([]byte(`
						[
							{
								"protocol": "tcp",
								"destination": "1.2.3.4/24",
								"ports": "80,443",
								"description": "some security group"
							}
						]
					`))

					Eventually(helpers.CF("create-security-group", securityGroupName, securityGroupRules.Name())).Should(Exit(0))
					Eventually(helpers.CF("bind-security-group", securityGroupName, orgName, spaceName)).Should(Exit(0))

					securityGroupName2 = helpers.PrefixedRandomName("aaaaaaaaaaz")
					securityGroupRules2, err = ioutil.TempFile("", "security-group-rules")
					Expect(err).ToNot(HaveOccurred())

					securityGroupRules2.Write([]byte(`
						[
							{
								"protocol": "udp",
								"destination": "92.0.0.1/24",
								"ports": "80,443",
								"description": "some other other security group"
							},
							{
								"protocol": "tcp",
								"destination": "5.7.9.11/24",
								"ports": "80,443",
								"description": "some other security group"
							}
						]
					`))

					Eventually(helpers.CF("create-security-group", securityGroupName2, securityGroupRules2.Name())).Should(Exit(0))
					Eventually(helpers.CF("bind-security-group", securityGroupName2, orgName, spaceName)).Should(Exit(0))
				})

				AfterEach(func() {
					Eventually(helpers.CF("delete-security-group", securityGroupName, "-f")).Should(Exit(0))
					os.Remove(securityGroupRules.Name())
					Eventually(helpers.CF("delete-security-group", securityGroupName2, "-f")).Should(Exit(0))
					os.Remove(securityGroupRules2.Name())
				})

				It("displays the space information as well as all security group rules", func() {
					session := helpers.CF("space", "--security-group-rules", spaceName)
					userName, _ := helpers.GetCredentials()
					Eventually(session.Out).Should(Say("Getting info for space %s in org %s as %s...", spaceName, orgName, userName))

					Eventually(session.Out).Should(Say("name:"))
					Eventually(session.Out).Should(Say("org:"))
					Eventually(session.Out).Should(Say("apps:"))
					Eventually(session.Out).Should(Say("services:"))
					Eventually(session.Out).Should(Say("isolation segment:"))
					Eventually(session.Out).Should(Say("space quota:"))
					Eventually(session.Out).Should(Say("security groups:"))
					Eventually(session.Out).Should(Say("\n\n"))

					Eventually(session.Out).Should(Say("security group\\s+destination\\s+ports\\s+protocol\\s+lifecycle\\s+description"))
					Eventually(session.Out).Should(Say("#0\\s+%s\\s+1.2.3.4/24\\s+80,443\\s+tcp\\s+running\\s+some security group", securityGroupName))
					Eventually(session.Out).Should(Say("#1\\s+%s\\s+5.7.9.11/24\\s+80,443\\s+tcp\\s+running\\s+some other security group", securityGroupName2))
					Eventually(session.Out).Should(Say("\\s+%s\\s+92.0.0.1/24\\s+80,443\\s+udp\\s+running\\s+some other other security group", securityGroupName2))
				})
			})
		})
	})
})
