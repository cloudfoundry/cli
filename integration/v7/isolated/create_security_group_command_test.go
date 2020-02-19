package isolated

import (
	"io/ioutil"
	"os"

	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = FDescribe("create-security-group command", func() {
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
				Expect(session).To(HaveCommandInCategoryWithDescription("create-security-group", "SECURITY GROUP", "Create a security group"))
			})

			It("Displays command usage to output", func() {
				session := helpers.CF("create-security-group", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("create-security-group - Create a security group"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say(`cf create-security-group SECURITY_GROUP PATH_TO_JSON_RULES_FILE`))
				Eventually(session).Should(Say(`The provided path can be an absolute or relative path to a file. The file should have`))
				Eventually(session).Should(Say(`a single array with JSON objects inside describing the rules. The JSON Base Object is`))
				Eventually(session).Should(Say(`omitted and only the square brackets and associated child object are required in the file.`))

				Eventually(session).Should(Say(`Valid json file example:`))
				Eventually(session).Should(Say(`\[`))
				Eventually(session).Should(Say(`  {`))
				Eventually(session).Should(Say(`    "protocol": "tcp",`))
				Eventually(session).Should(Say(`    "destination": "10.0.11.0/24",`))
				Eventually(session).Should(Say(`    "ports": "80,443",`))
				Eventually(session).Should(Say(`    "description": "Allow http and https traffic from ZoneA"`))
				Eventually(session).Should(Say(`  }`))
				Eventually(session).Should(Say(`\]`))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("bind-running-security-group, bind-security-group, bind-staging-security-group, security-groups"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("not logged in", func() {
		var tmpfile *os.File

		BeforeEach(func() {
			var err error
			tmpfile, err = ioutil.TempFile("", "push-archive-integration")
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			Expect(os.RemoveAll(tmpfile.Name())).ToNot(HaveOccurred())
		})

		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(false, false, "", "create-security-group", "some-group", tmpfile.Name())
		})
	})

	XWhen("the environment is set up correctly", func() {
		BeforeEach(func() {
			userName = helpers.LoginCF()
			orgName = helpers.CreateAndTargetOrg()
			spaceQuotaName = helpers.QuotaName()
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
