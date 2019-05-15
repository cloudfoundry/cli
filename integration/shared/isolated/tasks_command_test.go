package isolated

import (
	"fmt"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("tasks command", func() {
	var (
		appName string
	)

	BeforeEach(func() {
		appName = helpers.PrefixedRandomName("APP")
	})

	When("--help flag is set", func() {
		It("Displays command usage to output", func() {
			session := helpers.CF("tasks", "--help")
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Say("   tasks - List tasks of an app"))
			Eventually(session).Should(Say("USAGE:"))
			Eventually(session).Should(Say("   cf tasks APP_NAME"))
			Eventually(session).Should(Say("SEE ALSO:"))
			Eventually(session).Should(Say("   apps, logs, run-task, terminate-task"))
			Eventually(session).Should(Exit(0))
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "tasks", "app-name")
		})
	})

	When("the environment is setup correctly", func() {
		var (
			orgName   string
			spaceName string
		)

		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()

			helpers.SetupCF(orgName, spaceName)
		})

		AfterEach(func() {
			helpers.LoginCF()
			helpers.QuickDeleteOrg(orgName)
		})

		When("the application does not exist", func() {
			It("fails and outputs an app not found message", func() {
				session := helpers.CF("tasks", appName)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say(fmt.Sprintf("App '%s' not found", appName)))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the application exists", func() {
			BeforeEach(func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CF("push", appName, "-p", appDir, "-b", "staticfile_buildpack")).Should(Exit(0))
				})
			})

			When("the application does not have associated tasks", func() {
				It("displays an empty table", func() {
					session := helpers.CF("tasks", appName)
					Eventually(session).Should(Say(`
id   name   state   start time   command
`,
					))
					Consistently(session).ShouldNot(Say("1"))
					Eventually(session).Should(Exit(0))
				})
			})

			When("the application has associated tasks", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("run-task", appName, "echo hello world")).Should(Exit(0))
					Eventually(helpers.CF("run-task", appName, "echo foo bar")).Should(Exit(0))
				})

				It("displays all the tasks in descending order", func() {
					session := helpers.CF("tasks", appName)
					userName, _ := helpers.GetCredentials()
					Eventually(session).Should(Say(fmt.Sprintf("Getting tasks for app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName)))
					Eventually(session).Should(Say("OK\n"))
					Eventually(session).Should(Say(`id\s+name\s+state\s+start time\s+command
2\s+[a-zA-Z-0-9 ,:]+echo foo bar
1\s+[a-zA-Z-0-9 ,:]+echo hello world`))
					Eventually(session).Should(Exit(0))
				})

				When("the logged in user does not have authorization to see task commands", func() {
					var user string

					BeforeEach(func() {
						user = helpers.NewUsername()
						password := helpers.NewPassword()
						Eventually(helpers.CF("create-user", user, password)).Should(Exit(0))
						Eventually(helpers.CF("set-space-role", user, orgName, spaceName, "SpaceAuditor")).Should(Exit(0))
						helpers.LogoutCF()
						env := map[string]string{
							"CF_USERNAME": user,
							"CF_PASSWORD": password,
						}
						Eventually(helpers.CFWithEnv(env, "auth")).Should(Exit(0))
						Eventually(helpers.CF("target", "-o", orgName, "-s", spaceName)).Should(Exit(0))
					})

					It("does not display task commands", func() {
						session := helpers.CF("tasks", appName)
						Eventually(session).Should(Say(`2\s+[a-zA-Z-0-9 ,:]+\[hidden\]`))
						Eventually(session).Should(Say(`1\s+[a-zA-Z-0-9 ,:]+\[hidden\]`))
						Eventually(session).Should(Exit(0))
					})
				})
			})
		})
	})
})
