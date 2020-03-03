package isolated

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("create-security-group command", func() {
	var (
		dir               string
		securityGroupName string
		tempPath          string
		userName          string
	)

	BeforeEach(func() {
		userName = helpers.LoginCF()
		securityGroupName = helpers.NewSecurityGroupName()
		dir = helpers.TempDirAbsolutePath("", "security-group")

		tempPath = filepath.Join(dir, "tmpfile.json")
		err := ioutil.WriteFile(tempPath, []byte(`[{
			"protocol": "tcp",
			"destination": "10.0.11.0/24",
			"ports": "80,443",
			"description": "Allow http and https traffic from ZoneA"
		}]`), 0666)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		err := os.RemoveAll(dir)
		Expect(err).NotTo(HaveOccurred())
	})

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
				Eventually(session).Should(Say(`\s+\[`))
				Eventually(session).Should(Say(`\s+{`))
				Eventually(session).Should(Say(`\s+"protocol": "tcp",`))
				Eventually(session).Should(Say(`\s+"destination": "10.0.11.0/24",`))
				Eventually(session).Should(Say(`\s+"ports": "80,443",`))
				Eventually(session).Should(Say(`\s+"description": "Allow http and https traffic from ZoneA"`))
				Eventually(session).Should(Say(`\s+}`))
				Eventually(session).Should(Say(`\s+\]`))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("bind-running-security-group, bind-security-group, bind-staging-security-group, security-groups"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("not logged in", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(false, false, "", "create-security-group", "some-group", tempPath)
		})
	})

	When("the environment is set up correctly", func() {
		When("the security group name is not provided", func() {
			It("tells the user that the security group name is required, prints help text, and exits 1", func() {
				session := helpers.CF("create-security-group")

				Eventually(session.Err).Should(Say("Incorrect Usage: the required arguments `SECURITY_GROUP` and `PATH_TO_JSON_RULES_FILE` were not provided"))
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the security group path is not valid", func() {
			It("tells the user that the security group path does not exist, prints help text, and exits 1", func() {
				session := helpers.CF("create-security-group", securityGroupName, "/invalid/path")

				Eventually(session.Err).Should(Say("Incorrect Usage: The specified path '/invalid/path' does not exist."))
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the security group path has invalid JSON", func() {
			BeforeEach(func() {
				err := ioutil.WriteFile(tempPath, []byte("Invalid JSON!"), 0666)
				Expect(err).ToNot(HaveOccurred())
			})

			It("tells the user that the security group path is required, prints help text, and exits 1", func() {
				session := helpers.CF("create-security-group", securityGroupName, tempPath)
				Eventually(session).Should(Say("Creating security group %s as %s...", securityGroupName, userName))
				Eventually(session.Err).Should(Say("Incorrect json format: %s", strings.ReplaceAll(tempPath, "\\", "\\\\")))

				Eventually(session.Err).Should(Say("Valid json file example:"))
				Eventually(session.Err).Should(Say(`\[`))
				Eventually(session.Err).Should(Say(`\s+{`))
				Eventually(session.Err).Should(Say(`\s+"protocol": "tcp",`))
				Eventually(session.Err).Should(Say(`\s+"destination": "10.244.1.18",`))
				Eventually(session.Err).Should(Say(`\s+"ports": "3306"`))
				Eventually(session.Err).Should(Say(`\s+}`))
				Eventually(session.Err).Should(Say(`\]`))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the security group does not exist", func() {
			It("creates the security group", func() {
				session := helpers.CF("create-security-group", securityGroupName, tempPath)
				Eventually(session).Should(Say("Creating security group %s as %s...", securityGroupName, userName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))

				session = helpers.CF("security-group", securityGroupName)
				Eventually(session).Should(Say(`Name\s+%s`, securityGroupName))
			})
		})

		When("the security group already exists", func() {
			BeforeEach(func() {
				Eventually(helpers.CF("create-security-group", securityGroupName, tempPath)).Should(Exit(0))
			})

			It("notifies the user that the security group already exists and exits 0", func() {
				session := helpers.CF("create-security-group", securityGroupName, tempPath)
				Eventually(session).Should(Say("Creating security group %s as %s...", securityGroupName, userName))
				Eventually(session.Err).Should(Say(`Security group with name '%s' already exists\.`, securityGroupName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
