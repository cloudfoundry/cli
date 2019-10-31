package isolated

import (
	"fmt"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("terminate-task command", func() {
	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "terminate-task", "app-name", "3")
		})
	})

	When("the environment is setup correctly", func() {
		var (
			orgName   string
			spaceName string
			appName   string
		)

		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()
			appName = helpers.PrefixedRandomName("APP")

			helpers.SetupCF(orgName, spaceName)
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		When("the application does not exist", func() {
			It("fails to terminate task and outputs an error message", func() {
				session := helpers.CF("terminate-task", appName, "1")
				Eventually(session.Err).Should(Say(fmt.Sprintf("App '%s' not found", appName)))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the application exists", func() {
			BeforeEach(func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CF("push", appName, "-p", appDir, "-b", "staticfile_buildpack")).Should(Exit(0))
				})
			})

			When("the wrong data type is provided to terminate-task", func() {
				It("outputs an error message to the user, provides help text, and exits 1", func() {
					session := helpers.CF("terminate-task", appName, "not-an-integer")
					Eventually(session.Err).Should(Say("Incorrect usage: Value for TASK_ID must be integer"))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Say("terminate-task APP_NAME TASK_ID")) // help
					Eventually(session).Should(Exit(1))
				})
			})

			When("the task is in the RUNNING state", func() {
				BeforeEach(func() {
					helpers.WithHelloWorldApp(func(appDir string) {
						Eventually(helpers.CF("run-task", appName, "sleep 1000")).Should(Exit(0))
					})
				})

				It("terminates the task", func() {
					tasksSession := helpers.CF("tasks", appName)
					Eventually(tasksSession).Should(Exit(0))
					Expect(tasksSession).To(Say(`1\s+[a-zA-Z-0-9]+\s+RUNNING`))

					session := helpers.CF("terminate-task", appName, "1")
					userName, _ := helpers.GetCredentials()
					Eventually(session).Should(Say(
						fmt.Sprintf("Terminating task 1 of app %s in org %s / space %s as %s..", appName, orgName, spaceName, userName)))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))
				})
			})

			When("the task is in the SUCCEEDED state", func() {
				BeforeEach(func() {
					helpers.WithHelloWorldApp(func(appDir string) {
						Eventually(helpers.CF("run-task", appName, "echo test")).Should(Exit(0))
					})
				})

				It("fails to terminate the task and prints an error", func() {
					Eventually(func() *Buffer {
						taskSession := helpers.CF("tasks", appName)
						Eventually(taskSession).Should(Exit(0))
						return taskSession.Out
					}).Should(Say(`1\s+[a-zA-Z-0-9]+\s+SUCCEEDED`))

					session := helpers.CF("terminate-task", appName, "1")
					Eventually(session.Err).Should(Say("Task state is SUCCEEDED and therefore cannot be canceled"))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("the task is in the FAILED state", func() {
				BeforeEach(func() {
					helpers.WithHelloWorldApp(func(appDir string) {
						Eventually(helpers.CF("run-task", appName, "false")).Should(Exit(0))
					})
				})

				It("fails to terminate the task and prints an error", func() {
					Eventually(func() *Buffer {
						taskSession := helpers.CF("tasks", appName)
						Eventually(taskSession).Should(Exit(0))
						return taskSession.Out
					}).Should(Say(`1\s+[a-zA-Z-0-9]+\s+FAILED`))

					session := helpers.CF("terminate-task", appName, "1")
					Eventually(session.Err).Should(Say("Task state is FAILED and therefore cannot be canceled"))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("the task ID does not exist", func() {
				It("fails to terminate the task and prints an error", func() {
					session := helpers.CF("terminate-task", appName, "1")
					Eventually(session.Err).Should(Say("Task sequence ID 1 not found."))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				})
			})
		})
	})
})
