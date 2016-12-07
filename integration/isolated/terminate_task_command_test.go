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
	Context("when the environment is not setup correctly", func() {
		Context("when no API endpoint is set", func() {
			BeforeEach(func() {
				helpers.UnsetAPI()
			})

			It("fails with no API endpoint set message", func() {
				session := helpers.CF("terminate-task", "app-name", "3")
				Eventually(session).Should(Exit(1))
				Expect(session.Out).To(Say("FAILED"))
				Expect(session.Err).To(Say("No API endpoint set. Use 'cf login' or 'cf api' to target an endpoint."))
			})
		})

		Context("when not logged in", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
			})

			It("fails with not logged in message", func() {
				session := helpers.CF("terminate-task", "app-name", "3")
				Eventually(session).Should(Exit(1))
				Expect(session.Out).To(Say("FAILED"))
				Expect(session.Err).To(Say("Not logged in. Use 'cf login' to log in."))
			})
		})

		Context("when there no org set", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
				helpers.LoginCF()
			})

			It("fails with no targeted org error message", func() {
				session := helpers.CF("terminate-task", "app-name", "3")
				Eventually(session).Should(Exit(1))
				Expect(session.Out).To(Say("FAILED"))
				Expect(session.Err).To(Say("No org targeted, use 'cf target -o ORG' to target an org."))
			})
		})

		Context("when there no space set", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
				helpers.LoginCF()
				helpers.TargetOrg(ReadOnlyOrg)
			})

			It("fails with no space targeted error message", func() {
				session := helpers.CF("terminate-task", "app-name", "3")
				Eventually(session).Should(Exit(1))
				Expect(session.Out).To(Say("FAILED"))
				Expect(session.Err).To(Say("No space targeted, use 'cf target -s SPACE' to target a space"))
			})
		})
	})

	Context("when the environment is setup correctly", func() {
		var (
			orgName   string
			spaceName string
			appName   string
		)

		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			spaceName = helpers.PrefixedRandomName("SPACE")
			appName = helpers.PrefixedRandomName("APP")

			setupCF(orgName, spaceName)
		})

		Context("when the application does not exist", func() {
			It("fails to terminate task and outputs an error message", func() {
				session := helpers.CF("terminate-task", appName, "1")
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say(fmt.Sprintf("App %s not found", appName)))
				Expect(session.Out).To(Say("FAILED"))
			})
		})

		Context("when the application exists", func() {
			BeforeEach(func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CF("push", appName, "-p", appDir, "-b", "staticfile_buildpack")).Should(Exit(0))
				})
			})

			Context("when the wrong data type is provided to terminate-task", func() {
				It("outputs an error message to the user, provides help text, and exits 1", func() {
					session := helpers.CF("terminate-task", appName, "not-an-integer")
					Eventually(session).Should(Exit(1))
					Expect(session.Err).To(Say("Incorrect usage: Value for TASK_ID must be integer"))
					Expect(session.Out).To(Say("FAILED"))
					Expect(session.Out).To(Say("terminate-task APP_NAME TASK_ID")) // help
				})
			})

			Context("when the task is in the RUNNING state", func() {
				BeforeEach(func() {
					helpers.WithHelloWorldApp(func(appDir string) {
						Eventually(helpers.CF("run-task", appName, "sleep 1000")).Should(Exit(0))
					})
				})

				It("terminates the task", func() {
					tasksSession := helpers.CF("tasks", appName)
					Eventually(tasksSession).Should(Exit(0))
					Expect(tasksSession.Out).To(Say("1\\s+[a-zA-Z-0-9]+\\s+RUNNING"))

					session := helpers.CF("terminate-task", appName, "1")
					Eventually(session).Should(Exit(0))
					userName, _ := helpers.GetCredentials()
					Expect(session.Out).To(Say(
						fmt.Sprintf("Terminating task 1 of app %s in org %s / space %s as %s..", appName, orgName, spaceName, userName)))
					Expect(session.Out).To(Say("OK"))
				})
			})

			Context("when the task is in the SUCCEEDED state", func() {
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
					}).Should(Say("1\\s+[a-zA-Z-0-9]+\\s+SUCCEEDED"))

					session := helpers.CF("terminate-task", appName, "1")
					Eventually(session).Should(Exit(1))
					Expect(session.Err).To(Say("The request is semantically invalid: Task state is SUCCEEDED and therefore cannot be canceled"))
					Expect(session.Out).To(Say("FAILED"))
				})
			})

			Context("when the task is in the FAILED state", func() {
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
					}).Should(Say("1\\s+[a-zA-Z-0-9]+\\s+FAILED"))

					session := helpers.CF("terminate-task", appName, "1")
					Eventually(session).Should(Exit(1))
					Expect(session.Err).To(Say("The request is semantically invalid: Task state is FAILED and therefore cannot be canceled"))
					Expect(session.Out).To(Say("FAILED"))
				})
			})

			Context("when the task ID does not exist", func() {
				It("fails to terminate the task and prints an error", func() {
					session := helpers.CF("terminate-task", appName, "1")
					Eventually(session).Should(Exit(1))
					Expect(session.Err).To(Say("Task sequence ID 1 not found."))
					Expect(session.Out).To(Say("FAILED"))
				})
			})
		})
	})
})
