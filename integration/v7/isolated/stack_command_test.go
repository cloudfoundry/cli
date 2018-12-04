package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = FDescribe("stack command", func() {
	var (
		orgName          string
		spaceName        string
		stackName        string
		stackDescription string
	)

	BeforeEach(func() {
		orgName = helpers.NewOrgName()
		spaceName = helpers.NewSpaceName()
		stackName = helpers.PrefixedRandomName("stack")
		stackDescription = "this is a test stack"
	})

	Describe("help", func() {
		When("--help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("stack", "--help")

				Eventually(session).Should(Say(`NAME:`))
				Eventually(session).Should(Say(`stack - Show information for a stack \(a stack is a pre-built file system, including an operating system, that can run apps\)`))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf stack STACK_NAME"))

				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the stack name is not provided", func() {
		It("tells the user that the stack name is required, prints help text, and exits 1", func() {
			session := helpers.CF("stack")

			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `STACK_NAME` was not provided"))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "stack", stackName)
		})
	})

	When("the environment is set up correctly", func() {
		var username string

		BeforeEach(func() {
			helpers.SetupCF(orgName, spaceName)
			username, _ = helpers.GetCredentials()
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		When("the input is invalid", func() {
			When("there are not enough arguments", func() {
				It("outputs the usage and exits 1", func() {
					session := helpers.CF("stack")

					Eventually(session.Err).Should(Say("Incorrect Usage:"))
					Eventually(session).Should(Say("NAME:"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("there too many arguments", func() {
				It("ignores the extra arguments", func() {
					session := helpers.CF("stack", stackName, "extra")

					Eventually(session).Should(Say(`Getting stack %s in org %s / space %s as %s\.\.\.`, stackName, orgName, spaceName, username))
					Eventually(session.Err).Should(Say("Stack %s not found", stackName))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				})
			})
		})

		When("the stack does not exist", func() {
			It("Fails", func() {
				session := helpers.CF("stack", stackName)

				Eventually(session).Should(Say(`Getting stack %s in org %s / space %s as %s\.\.\.`, stackName, orgName, spaceName, username))
				Eventually(session.Err).Should(Say("Stack %s not found", stackName))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the stack exists", func() {
			BeforeEach(func() {
				jsonBody := fmt.Sprintf(`{"name": "%s", "description": "%s"}`, stackName, stackDescription)
				session := helpers.CF("curl", "-d", jsonBody, "-X", "POST", "/v3/stacks")
				Eventually(session).Should(Exit(0))
			})
			It("Shows the details for the stack", func() {
				session := helpers.CF("stack", stackName)

				Eventually(session).Should(Say(`Getting stack %s in org %s / space %s as %s\.\.\.`, stackName, orgName, spaceName, username))
				Eventually(session).Should(Say(`name:\s+%s`, stackName))
				Eventually(session).Should(Say(`description:\s+%s`, stackDescription))
				Eventually(session).Should(Exit(0))
			})

			When("the stack exists and the --guid flag is passed", func() {
				It("prints nothing but the guid", func() {
					session := helpers.CF("stack", stackName, "--guid")

					Consistently(session).ShouldNot(Say(`Getting stack %s in org %s / space %s as %s\.\.\.`, stackName, orgName, spaceName, username))
					Consistently(session).ShouldNot(Say(`name:\s+%s`, stackName))
					Consistently(session).ShouldNot(Say(`description:\s+%s`, stackDescription))
					Eventually(session).Should(Say(`^[a-z0-9]{8}-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{12}`))
					Eventually(session).Should(Exit(0))
				})
			})
		})
	})
})
