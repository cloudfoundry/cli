package global

import (
	"encoding/json"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/resources"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("update-security-group command", func() {
	When("--help flag is set", func() {
		It("Displays command usage to output", func() {
			session := helpers.CF("update-security-group", "--help")
			displaysUpdateSecurityGroupHelpText(session)
			Eventually(session).Should(Exit(0))
		})
	})

	var (
		securityGroup             resources.SecurityGroup
		securityGroupName         string
		updatedSecurityGroupRules []resources.Rule
		tmpRulesDir               string
		updatedRulesPath          string
		username                  string
		err                       error
	)

	BeforeEach(func() {
		username = helpers.LoginCF()
		securityGroupName = helpers.NewSecurityGroupName()
		securityGroup = resources.SecurityGroup{
			Name:  securityGroupName,
			Rules: []resources.Rule{},
		}
		helpers.CreateSecurityGroup(securityGroup)
	})

	BeforeEach(func() {
		// create the JSON rules file
		ports := "25,465,587"
		description := "Give me a place to stand on, and I shall spam the world!"
		updatedSecurityGroupRules = []resources.Rule{
			{
				Protocol:    "tcp",
				Destination: "0.0.0.0/0",
				Ports:       &ports,
				Description: &description,
			},
		}

		tmpRulesDir, err = os.MkdirTemp("", "spam-security-group")
		Expect(err).ToNot(HaveOccurred())

		updatedRulesPath = filepath.Join(tmpRulesDir, "spam-security-group.json")

		securityGroup, err := json.Marshal(updatedSecurityGroupRules)
		Expect(err).ToNot(HaveOccurred())

		err = os.WriteFile(updatedRulesPath, securityGroup, 0666)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		helpers.DeleteSecurityGroup(securityGroup)
		err = os.RemoveAll(tmpRulesDir)
		Expect(err).ToNot(HaveOccurred())
	})

	It("updates the security group", func() {
		session := helpers.CF("update-security-group", securityGroupName, updatedRulesPath)
		Eventually(session).Should(Say(`Updating security group %s as %s`, securityGroupName, username))
		Eventually(session).Should(Say("OK"))
		Eventually(session).Should(Say(`TIP: If Dynamic ASG's are enabled, changes will automatically apply for running and staging applications. Otherwise, changes will require an app restart \(for running\) or restage \(for staging\) to apply to existing applications\.`))
		Eventually(session).Should(Exit(0))

		session = helpers.CF("security-group", securityGroupName)
		Eventually(session).Should(Say(`Getting info for security group %s as %s\.\.\.`, securityGroupName, username))
		Eventually(session).Should(Say(`name:\s+%s`, securityGroupName))
		Eventually(session).Should(Say("rules:"))
		Eventually(session).Should(Say("\\["))
		Eventually(session).Should(Say("{"))
		Eventually(session).Should(Say(`"protocol": "tcp",`))
		Eventually(session).Should(Say(`"destination": "0.0.0.0/0",`))
		Eventually(session).Should(Say(`"ports": "25,465,587",`))
		Eventually(session).Should(Say(`"description": "Give me a place to stand on, and I shall spam the world!"`))
		Eventually(session).Should(Say("}"))
		Eventually(session).Should(Say("]"))
		Eventually(session).Should(Say("No spaces assigned"))

		Eventually(session).Should(Exit(0))
	})

	When("the security group does not exist or is not visible to the user", func() {
		BeforeEach(func() {
			helpers.DeleteSecurityGroup(securityGroup)
		})
		It("displays a 'not-found' security group error message and fails", func() {
			session := helpers.CF("update-security-group", securityGroupName, updatedRulesPath)
			Eventually(session).Should(Say("FAILED"))
			Eventually(session.Err).Should(Say("Security group '%s' not found.", securityGroupName))
			Eventually(session).Should(Exit(1))
		})
	})

	When("the JSON file with the security group's rules does not exist", func() {
		It("displays a does-not-exist error message and displays help text", func() {
			session := helpers.CF("update-security-group", securityGroupName, "/non/existent/path")
			Eventually(session.Err).Should(Say("Incorrect Usage: The specified path '/non/existent/path' does not exist."))
			displaysUpdateSecurityGroupHelpText(session)
			Eventually(session).Should(Exit(1))
		})
	})

	When("the JSON file with the security group's rules is not JSON", func() {
		BeforeEach(func() {
			err = os.WriteFile(updatedRulesPath, []byte("I'm definitely not JSON"), 0666)
			Expect(err).ToNot(HaveOccurred())
		})
		It("displays a an incorrect JSON format message and fails", func() {
			session := helpers.CF("update-security-group", securityGroupName, updatedRulesPath)
			Eventually(session).Should(Say("FAILED"))
			Eventually(session.Err).Should(helpers.SayPath("Incorrect json format: %s", updatedRulesPath))
			Eventually(session).Should(Exit(1))
		})
	})

	When("there are unexpected arguments", func() {
		It("complains that there are unexpected arguments, fails, and prints the help text", func() {
			session := helpers.CF("update-security-group", securityGroupName, updatedRulesPath, "unexpected-argument")
			Eventually(session.Err).Should(Say(`Incorrect Usage: unexpected argument "unexpected-argument"`))
			Eventually(session).Should(Say("FAILED"))
			displaysUpdateSecurityGroupHelpText(session)
			Eventually(session).Should(Exit(1))
		})
	})
})

func displaysUpdateSecurityGroupHelpText(session *Session) {
	Eventually(session).Should(Say("NAME:"))
	Eventually(session).Should(Say("update-security-group - Update a security group"))
	Eventually(session).Should(Say("USAGE:"))
	Eventually(session).Should(Say("cf update-security-group SECURITY_GROUP PATH_TO_JSON_RULES_FILE"))
	Eventually(session).Should(Say("The provided path can be an absolute or relative path to a file. The file should have"))
	Eventually(session).Should(Say("a single array with JSON objects inside describing the rules. The JSON Base Object is"))
	Eventually(session).Should(Say("omitted and only the square brackets and associated child object are required in the file."))
	Eventually(session).Should(Say("Valid json file example:"))
	Eventually(session).Should(Say("\\["))
	Eventually(session).Should(Say("{"))
	Eventually(session).Should(Say(`"protocol": "tcp",`))
	Eventually(session).Should(Say(`"destination": "10.0.11.0/24",`))
	Eventually(session).Should(Say(`"ports": "80,443",`))
	Eventually(session).Should(Say(`"description": "Allow http and https traffic from ZoneA"`))
	Eventually(session).Should(Say("}"))
	Eventually(session).Should(Say("]"))
	Eventually(session).Should(Say(`TIP: If Dynamic ASG's are enabled, changes will automatically apply for running and staging applications. Otherwise, changes will require an app restart \(for running\) or restage \(for staging\) to apply to existing applications\.`))
	Eventually(session).Should(Say("SEE ALSO:"))
	Eventually(session).Should(Say("restage, security-groups"))
}
