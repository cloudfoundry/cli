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

var _ = Describe("update-stack command", func() {
	var (
		orgName          string
		spaceName        string
		stackName        string
		stackDescription string
		stackGUID        string
		username         string
	)

	BeforeEach(func() {
		orgName = helpers.NewOrgName()
		spaceName = helpers.NewSpaceName()
		stackName = helpers.PrefixedRandomName("stack")
		stackDescription = "test stack for update"
	})

	Describe("help", func() {
		When("--help flag is set", func() {
			It("appears in cf help -a", func() {
				session := helpers.CF("help", "-a")
				Eventually(session).Should(Exit(0))
				Expect(session).To(HaveCommandInCategoryWithDescription("update-stack", "APPS", "Transition a stack between the defined states"))
			})

			It("displays command usage to output", func() {
				session := helpers.CF("update-stack", "--help")

				Eventually(session).Should(Say(`NAME:`))
				Eventually(session).Should(Say(`update-stack - Transition a stack between the defined states`))
				Eventually(session).Should(Say(`USAGE:`))
				Eventually(session).Should(Say(`cf update-stack STACK_NAME \[--state \(active \| restricted \| deprecated \| disabled\)\] \[--reason REASON\]`))
				Eventually(session).Should(Say(`EXAMPLES:`))
				Eventually(session).Should(Say(`cf update-stack cflinuxfs3 --state disabled`))
				Eventually(session).Should(Say(`cf update-stack cflinuxfs3 --state deprecated --reason 'Use cflinuxfs4 instead'`))
				Eventually(session).Should(Say(`OPTIONS:`))
				Eventually(session).Should(Say(`--state\s+State to transition the stack to`))
				Eventually(session).Should(Say(`SEE ALSO:`))
				Eventually(session).Should(Say(`stack, stacks`))

				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the stack name is not provided", func() {
		It("tells the user that the stack name is required, prints help text, and exits 1", func() {
			session := helpers.CF("update-stack", "--state", "active")

			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `STACK_NAME` was not provided"))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	When("the state flag is not provided", func() {
		It("tells the user that the state flag is required, prints help text, and exits 1", func() {
			session := helpers.CF("update-stack", "some-stack")

			Eventually(session.Err).Should(Say("Incorrect Usage: the required flag `--state' was not specified"))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(false, false, ReadOnlyOrg, "update-stack", stackName, "--state", "active")
		})
	})

	When("the environment is set up correctly", func() {
		BeforeEach(func() {
			helpers.SetupCF(orgName, spaceName)
			username, _ = helpers.GetCredentials()
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		When("the stack does not exist", func() {
			It("fails with stack not found error", func() {
				session := helpers.CF("update-stack", stackName, "--state", "active")

				Eventually(session).Should(Say(`Updating stack %s as %s\.\.\.`, stackName, username))
				Eventually(session.Err).Should(Say("Stack '%s' not found", stackName))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})

		When("an invalid state is provided", func() {
			BeforeEach(func() {
				// Create a stack first
				jsonBody := fmt.Sprintf(`{"name": "%s", "description": "%s", "state": "ACTIVE"}`, stackName, stackDescription)
				session := helpers.CF("curl", "-d", jsonBody, "-X", "POST", "/v3/stacks")
				Eventually(session).Should(Exit(0))

				r := regexp.MustCompile(`[a-z0-9]{8}-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{12}`)
				stackGUID = string(r.Find(session.Out.Contents()))
			})

			AfterEach(func() {
				if stackGUID != "" {
					session := helpers.CF("curl", "-X", "DELETE", fmt.Sprintf("/v3/stacks/%s", stackGUID))
					Eventually(session).Should(Exit(0))
				}
			})

			It("fails with invalid state error", func() {
				session := helpers.CF("update-stack", stackName, "--state", "invalid-state")

				Eventually(session.Err).Should(Say("Invalid state: invalid-state"))
				Eventually(session.Err).Should(Say("Must be one of: active, restricted, deprecated, disabled"))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the stack exists", func() {
			BeforeEach(func() {
				// Create a stack with initial state
				jsonBody := fmt.Sprintf(`{"name": "%s", "description": "%s", "state": "ACTIVE"}`, stackName, stackDescription)
				session := helpers.CF("curl", "-d", jsonBody, "-X", "POST", "/v3/stacks")
				Eventually(session).Should(Exit(0))

				r := regexp.MustCompile(`[a-z0-9]{8}-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{12}`)
				stackGUID = string(r.Find(session.Out.Contents()))
			})

			AfterEach(func() {
				if stackGUID != "" {
					session := helpers.CF("curl", "-X", "DELETE", fmt.Sprintf("/v3/stacks/%s", stackGUID))
					Eventually(session).Should(Exit(0))
				}
			})

			When("updating to deprecated state", func() {
				It("successfully updates the stack state", func() {
					session := helpers.CF("update-stack", stackName, "--state", "deprecated")

					Eventually(session).Should(Say(`Updating stack %s as %s\.\.\.`, stackName, username))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Say(`name:\s+%s`, stackName))
					Eventually(session).Should(Say(`description:\s+%s`, stackDescription))
					Eventually(session).Should(Say(`state:\s+DEPRECATED`))
					Eventually(session).Should(Exit(0))

					// Verify the state was actually updated
					verifySession := helpers.CF("stack", stackName)
					Eventually(verifySession).Should(Say(`state:\s+DEPRECATED`))
					Eventually(verifySession).Should(Exit(0))
				})
			})

			When("updating to restricted state", func() {
				It("successfully updates the stack state", func() {
					session := helpers.CF("update-stack", stackName, "--state", "restricted")

					Eventually(session).Should(Say(`Updating stack %s as %s\.\.\.`, stackName, username))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Say(`state:\s+RESTRICTED`))
					Eventually(session).Should(Exit(0))
				})
			})

			When("updating to disabled state", func() {
				It("successfully updates the stack state", func() {
					session := helpers.CF("update-stack", stackName, "--state", "disabled")

					Eventually(session).Should(Say(`Updating stack %s as %s\.\.\.`, stackName, username))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Say(`state:\s+DISABLED`))
					Eventually(session).Should(Exit(0))
				})
			})

			When("updating back to active state", func() {
				BeforeEach(func() {
					// First set it to deprecated
					session := helpers.CF("update-stack", stackName, "--state", "deprecated")
					Eventually(session).Should(Exit(0))
				})

				It("successfully updates the stack back to active", func() {
					session := helpers.CF("update-stack", stackName, "--state", "active")

					Eventually(session).Should(Say(`Updating stack %s as %s\.\.\.`, stackName, username))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Say(`state:\s+ACTIVE`))
					Eventually(session).Should(Exit(0))
				})
			})

			When("state value is provided in different cases", func() {
				It("accepts lowercase state value", func() {
					session := helpers.CF("update-stack", stackName, "--state", "deprecated")
					Eventually(session).Should(Exit(0))
					Eventually(session).Should(Say(`state:\s+DEPRECATED`))
				})

				It("accepts uppercase state value", func() {
					session := helpers.CF("update-stack", stackName, "--state", "RESTRICTED")
					Eventually(session).Should(Exit(0))
					Eventually(session).Should(Say(`state:\s+RESTRICTED`))
				})

				It("accepts mixed case state value", func() {
					session := helpers.CF("update-stack", stackName, "--state", "Disabled")
					Eventually(session).Should(Exit(0))
					Eventually(session).Should(Say(`state:\s+DISABLED`))
				})
			})
		})
	})
})

