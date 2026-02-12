package isolated

import (
	"fmt"
	"regexp"

	. "code.cloudfoundry.org/cli/v8/cf/util/testhelpers/matchers"

	"code.cloudfoundry.org/cli/v8/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("stack command", func() {
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
			It("appears in cf help -a", func() {
				session := helpers.CF("help", "-a")
				Eventually(session).Should(Exit(0))
				Expect(session).To(HaveCommandInCategoryWithDescription("stack", "APPS", "Show information for a stack (a stack is a pre-built file system, including an operating system, that can run apps) and current state"))
			})

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
			helpers.CheckEnvironmentTargetedCorrectly(false, false, ReadOnlyOrg, "stack", stackName)
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
		})

		When("the stack does not exist", func() {
			It("Fails", func() {
				session := helpers.CF("stack", stackName)

				Eventually(session).Should(Say(`Getting info for stack %s as %s\.\.\.`, stackName, username))
				Eventually(session.Err).Should(Say("Stack '%s' not found", stackName))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the stack exists with valid state", func() {
			var stackGUID string

			BeforeEach(func() {
				jsonBody := fmt.Sprintf(`{"name": "%s", "description": "%s", "state": "ACTIVE"}`, stackName, stackDescription)
				session := helpers.CF("curl", "-d", jsonBody, "-X", "POST", "/v3/stacks")
				Eventually(session).Should(Exit(0))

				r := regexp.MustCompile(`[a-z0-9]{8}-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{12}`)
				stackGUID = string(r.Find(session.Out.Contents()))
			})

			AfterEach(func() {
				session := helpers.CF("curl", "-X", "DELETE", fmt.Sprintf("/v3/stacks/%s", stackGUID))
				Eventually(session).Should(Exit(0))
			})

			It("Shows the details for the stack with state", func() {
				session := helpers.CF("stack", stackName)

				Eventually(session).Should(Say(`Getting info for stack %s as %s\.\.\.`, stackName, username))
				Eventually(session).Should(Say(`name:\s+%s`, stackName))
				Eventually(session).Should(Say(`description:\s+%s`, stackDescription))
				Eventually(session).Should(Say(`state:\s+ACTIVE`))
				Consistently(session).ShouldNot(Say(`reason:`))
				Eventually(session).Should(Exit(0))
			})

			It("does not show reason for an active stack", func() {
				session := helpers.CF("stack", stackName)

				Eventually(session).Should(Say(`state:\s+ACTIVE`))
				Consistently(session).ShouldNot(Say(`reason:`))
				Eventually(session).Should(Exit(0))
			})

			It("prints nothing but the guid when --guid flag is passed", func() {
				session := helpers.CF("stack", stackName, "--guid")

				Consistently(session).ShouldNot(Say(`Getting info for stack %s as %s\.\.\.`, stackName, username))
				Consistently(session).ShouldNot(Say(`name:\s+%s`, stackName))
				Consistently(session).ShouldNot(Say(`description:\s+%s`, stackDescription))
				Consistently(session).ShouldNot(Say(`state:`))
				Eventually(session).Should(Say(`^[a-z0-9]{8}-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{12}`))
				Eventually(session).Should(Exit(0))
			})

			When("the stack is in a non-active state without a reason", func() {
				BeforeEach(func() {
					session := helpers.CF("update-stack", stackName, "--state", "deprecated")
					Eventually(session).Should(Exit(0))
				})

				It("shows an empty reason field", func() {
					session := helpers.CF("stack", stackName)

					Eventually(session).Should(Say(`Getting info for stack %s as %s\.\.\.`, stackName, username))
					Eventually(session).Should(Say(`name:\s+%s`, stackName))
					Eventually(session).Should(Say(`description:\s+%s`, stackDescription))
					Eventually(session).Should(Say(`state:\s+DEPRECATED`))
					Eventually(session).Should(Say(`reason:\s*$`))
					Eventually(session).Should(Exit(0))
				})
			})

			When("the stack is in a non-active state with a reason", func() {
				BeforeEach(func() {
					session := helpers.CF("update-stack", stackName, "--state", "disabled", "--reason", "This stack is no longer supported.")
					Eventually(session).Should(Exit(0))
				})

				It("shows the reason in the output", func() {
					session := helpers.CF("stack", stackName)

					Eventually(session).Should(Say(`Getting info for stack %s as %s\.\.\.`, stackName, username))
					Eventually(session).Should(Say(`name:\s+%s`, stackName))
					Eventually(session).Should(Say(`description:\s+%s`, stackDescription))
					Eventually(session).Should(Say(`state:\s+DISABLED`))
					Eventually(session).Should(Say(`reason:\s+This stack is no longer supported\.`))
					Eventually(session).Should(Exit(0))
				})
			})
		})
	})
})
