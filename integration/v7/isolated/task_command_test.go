package isolated

import (
	"fmt"

	"code.cloudfoundry.org/cli/v8/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("task command", func() {
	var (
		appName string
	)

	BeforeEach(func() {
		appName = helpers.PrefixedRandomName("APP")
	})

	When("--help flag is set", func() {
		It("Displays command usage to output", func() {
			session := helpers.CF("task", "--help")
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Say("   task - Display a task of an app"))
			Eventually(session).Should(Say("USAGE:"))
			Eventually(session).Should(Say("   cf task APP_NAME TASK_ID"))
			Eventually(session).Should(Say("SEE ALSO:"))
			Eventually(session).Should(Say("   apps, logs, run-task, tasks, terminate-task"))
			Eventually(session).Should(Exit(0))
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "task", "app-name", "1")
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
				session := helpers.CF("task", appName, "1")
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

			When("the application does not have the associated task", func() {
				It("displays an erro", func() {
					session := helpers.CF("task", appName, "1000")
					Eventually(session.Err).Should(Say(`Task sequence ID 1000 not found`))
					Eventually(session).Should(Exit(1))
				})
			})

			When("the application has associated tasks", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("run-task", appName, "--command", "echo hello world")).Should(Exit(0))
				})

				It("displays the task", func() {
					session := helpers.CF("task", appName, "1")
					userName, _ := helpers.GetCredentials()
					Eventually(session).Should(Say(fmt.Sprintf("Getting task 1 for app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName)))

					Eventually(session).Should(Say(`id:\s+1`))
					Eventually(session).Should(Say(`name:\s+`))
					Eventually(session).Should(Say(`state:\s+`))
					Eventually(session).Should(Say(`command:\s+echo hello world`))

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

					It("displays [hidden] as tasks command", func() {
						session := helpers.CF("task", appName, "1")
						Eventually(session).Should(Say(`.*command:\s+\[hidden\]`))
						Eventually(session).Should(Exit(0))
					})
				})
			})
		})
	})
})
