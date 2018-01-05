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
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("when the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, false, ReadOnlyOrg, "space", "space-name")
		})
	})

	Context("when the environment is set up correctly", func() {
		BeforeEach(func() {
			helpers.LoginCF()
			setupCF(orgName, spaceName)
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
			helpers.ClearTarget()
		})

		Context("when the space does not exist", func() {
			It("displays not found and exits 1", func() {
				badSpaceName := fmt.Sprintf("%s-1", spaceName)
				session := helpers.CF("space", badSpaceName)
				userName, _ := helpers.GetCredentials()
				Eventually(session).Should(Say("Getting info for space %s in org %s as %s\\.\\.\\.", badSpaceName, orgName, userName))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("Space '%s' not found.", badSpaceName))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the space exists", func() {
			Context("when the --guid flag is used", func() {
				It("displays the space guid", func() {
					session := helpers.CF("space", "--guid", spaceName)
					Eventually(session).Should(Say("[\\da-f]{8}-[\\da-f]{4}-[\\da-f]{4}-[\\da-f]{4}-[\\da-f]{12}"))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when the --guid flag is not used", func() {
				var (
					securityGroup0      helpers.SecurityGroup
					securityGroup1      helpers.SecurityGroup
					securityGroupName2  string
					securityGroupRules2 *os.File
					err                 error
				)

				BeforeEach(func() {
					securityGroup0 = helpers.NewSecurityGroup(helpers.NewSecurityGroupName("0"), "tcp", "4.3.2.1/24", "80,443", "foo security group")
					securityGroup0.Create()
					Eventually(helpers.CF("bind-security-group", securityGroup0.Name, orgName, spaceName, "--lifecycle", "staging")).Should(Exit(0))

					securityGroup1 = helpers.NewSecurityGroup(helpers.NewSecurityGroupName("1"), "tcp", "1.2.3.4/24", "80,443", "some security group")
					securityGroup1.Create()
					Eventually(helpers.CF("bind-security-group", securityGroup1.Name, orgName, spaceName)).Should(Exit(0))

					securityGroupName2 = helpers.NewSecurityGroupName("2")
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
					os.Remove(securityGroupRules2.Name())

					Eventually(helpers.CF("bind-security-group", securityGroupName2, orgName, spaceName)).Should(Exit(0))
					Eventually(helpers.CF("bind-security-group", securityGroupName2, orgName, spaceName, "--lifecycle", "staging")).Should(Exit(0))
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
						isolationSegmentName = helpers.NewIsolationSegmentName()
						Eventually(helpers.CF("create-isolation-segment", isolationSegmentName)).Should(Exit(0))
						Eventually(helpers.CF("enable-org-isolation", orgName, isolationSegmentName)).Should(Exit(0))
						Eventually(helpers.CF("set-space-isolation-segment", spaceName, isolationSegmentName)).Should(Exit(0))

						Eventually(helpers.CF("bind-security-group", securityGroup1.Name, orgName, spaceName)).Should(Exit(0))
					})

					It("displays a table with space name, org, apps, services, isolation segment, space quota and security groups", func() {
						session := helpers.CF("space", spaceName)
						userName, _ := helpers.GetCredentials()
						Eventually(session).Should(Say("Getting info for space %s in org %s as %s\\.\\.\\.", spaceName, orgName, userName))

						Eventually(session).Should(Say("name:\\s+%s", spaceName))
						Eventually(session).Should(Say("org:\\s+%s", orgName))
						Eventually(session).Should(Say("apps:\\s+%s", appName))
						Eventually(session).Should(Say("services:\\s+%s", serviceInstance))
						Eventually(session).Should(Say("isolation segment:\\s+%s", isolationSegmentName))
						Eventually(session).Should(Say("space quota:\\s+%s", spaceQuotaName))
						Eventually(session).Should(Say("running security groups:\\s+.*%s,.* %s", securityGroup1.Name, securityGroupName2))
						Eventually(session).Should(Say("staging security groups:\\s+.*%s,.* %s", securityGroup0.Name, securityGroupName2))
						Eventually(session).Should(Exit(0))
					})
				})

				Context("when the space does not have an isolation segment and its org has a default isolation segment", func() {
					var orgIsolationSegmentName string

					BeforeEach(func() {
						orgIsolationSegmentName = helpers.NewIsolationSegmentName()
						Eventually(helpers.CF("create-isolation-segment", orgIsolationSegmentName)).Should(Exit(0))
						Eventually(helpers.CF("enable-org-isolation", orgName, orgIsolationSegmentName)).Should(Exit(0))
						Eventually(helpers.CF("set-org-default-isolation-segment", orgName, orgIsolationSegmentName)).Should(Exit(0))
					})

					It("shows the org default isolation segment", func() {
						session := helpers.CF("space", spaceName)
						Eventually(session).Should(Say("isolation segment:\\s+%s", orgIsolationSegmentName))
						Eventually(session).Should(Exit(0))
					})
				})

				Context("when the security group rules flag is used", func() {
					It("displays the space information as well as all security group rules", func() {
						session := helpers.CF("space", "--security-group-rules", spaceName)
						userName, _ := helpers.GetCredentials()
						Eventually(session).Should(Say("Getting info for space %s in org %s as %s\\.\\.\\.", spaceName, orgName, userName))

						Eventually(session).Should(Say("name:"))
						Eventually(session).Should(Say("org:"))
						Eventually(session).Should(Say("apps:"))
						Eventually(session).Should(Say("services:"))
						Eventually(session).Should(Say("isolation segment:"))
						Eventually(session).Should(Say("space quota:"))
						Eventually(session).Should(Say("running security groups:"))
						Eventually(session).Should(Say("staging security groups:"))
						Eventually(session).Should(Say("\n\n"))

						Eventually(session).Should(Say("security group\\s+destination\\s+ports\\s+protocol\\s+lifecycle\\s+description"))
						Eventually(session).Should(Say("#\\d+\\s+%s\\s+4.3.2.1/24\\s+80,443\\s+tcp\\s+staging\\s+foo security group", securityGroup0.Name))
						Eventually(session).Should(Say("#\\d+\\s+%s\\s+1.2.3.4/24\\s+80,443\\s+tcp\\s+running\\s+some security group", securityGroup1.Name))
						Eventually(session).Should(Say("#\\d+\\s+%s\\s+5.7.9.11/24\\s+80,443\\s+tcp\\s+running\\s+some other security group", securityGroupName2))
						Eventually(session).Should(Say("\\s+%s\\s+5.7.9.11/24\\s+80,443\\s+tcp\\s+staging\\s+some other security group", securityGroupName2))
						Eventually(session).Should(Say("\\s+%s\\s+92.0.0.1/24\\s+80,443\\s+udp\\s+running\\s+some other other security group", securityGroupName2))
						Eventually(session).Should(Say("\\s+%s\\s+92.0.0.1/24\\s+80,443\\s+udp\\s+staging\\s+some other other security group", securityGroupName2))
						Eventually(session).Should(Exit(0))
					})
				})
			})
		})
	})
})
