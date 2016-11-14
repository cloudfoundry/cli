package integration

import (
	"fmt"

	. "code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("terminate-task command", func() {
	var (
		orgName   string
		spaceName string
		appName   string
	)

	BeforeEach(func() {
		orgName = PrefixedRandomName("ORG")
		spaceName = PrefixedRandomName("SPACE")
		appName = PrefixedRandomName("APP")

		setupCF(orgName, spaceName)
	})

	AfterEach(func() {
		setAPI()
		loginCF()
		Eventually(CF("delete-org", "-f", orgName), CFLongTimeout).Should(Exit(0))
	})

	It("should display the command level help", func() {
		session := CF("terminate-task", "-h")
		Eventually(session).Should(Exit(0))
		Expect(session.Out).To(Say(`NAME:
   terminate-task - Terminate a running task of an app

USAGE:
   cf terminate-task APP_NAME TASK_ID

EXAMPLES:
   cf terminate-task my-app 3

SEE ALSO:
   tasks`))
	})

	Context("when the environment is not setup correctly", func() {
		Context("when no API endpoint is set", func() {
			BeforeEach(func() {
				unsetAPI()
			})

			It("fails with no API endpoint set message", func() {
				session := CF("terminate-task", "app-name", "3")
				Eventually(session).Should(Exit(1))
				Expect(session.Out).To(Say("FAILED"))
				Expect(session.Err).To(Say("No API endpoint set. Use 'cf login' or 'cf api' to target an endpoint."))
			})
		})

		Context("when not logged in", func() {
			BeforeEach(func() {
				logoutCF()
			})

			It("fails with not logged in message", func() {
				session := CF("terminate-task", "app-name", "3")
				Eventually(session).Should(Exit(1))
				Expect(session.Out).To(Say("FAILED"))
				Expect(session.Err).To(Say("Not logged in. Use 'cf login' to log in."))
			})
		})

		Context("when there no org set", func() {
			BeforeEach(func() {
				logoutCF()
				loginCF()
			})

			It("fails with no targeted org error message", func() {
				session := CF("terminate-task", "app-name", "3")
				Eventually(session).Should(Exit(1))
				Expect(session.Out).To(Say("FAILED"))
				Expect(session.Err).To(Say("No org targeted, use 'cf target -o ORG' to target an org."))
			})
		})

		Context("when there no space set", func() {
			BeforeEach(func() {
				// create a another space, because if the org has only one space it
				// will be automatically targetted
				createSpace(PrefixedRandomName("SPACE"))
				logoutCF()
				loginCF()
				targetOrg(orgName)
			})

			It("fails with no space targeted error message", func() {
				session := CF("terminate-task", "app-name", "3")
				Eventually(session).Should(Exit(1))
				Expect(session.Out).To(Say("FAILED"))
				Expect(session.Err).To(Say("No space targeted, use 'cf target -s SPACE' to target a space"))
			})
		})
	})

	Context("when the environment is setup correctly", func() {
		Context("when the task id argument is not an integer", func() {
			It("outputs an error message to the user and exits 1", func() {
				session := CF("terminate-task", appName, "not-an-integer")
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Incorrect usage: Value for TASK_ID must be integer"))
				Expect(session.Out).To(Say("FAILED"))
			})
		})

		Context("when the application does not exist", func() {
			It("fails to terminate task and outputs an error message", func() {
				session := CF("terminate-task", appName, "1")
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say(fmt.Sprintf("App %s not found", appName)))
				Expect(session.Out).To(Say("FAILED"))
			})
		})

		Context("when the application exists", func() {
			BeforeEach(func() {
				WithSimpleApp(func(appDir string) {
					Eventually(CF("push", appName, "-p", appDir, "-b", "staticfile_buildpack"), CFLongTimeout).Should(Exit(0))
				})
			})

			Context("when the wrong data type is provided to terminate-task", func() {
				It("fails to terminate task and outputs an error message", func() {
					session := CF("terminate-task", appName, "foo")
					Eventually(session).Should(Exit(1))
					Expect(session.Err).To(Say("Unexpected error: strconv.ParseInt: parsing \"foo\": invalid syntax"))
				})
			})

			Context("when the task is in the RUNNING state", func() {
				BeforeEach(func() {
					WithSimpleApp(func(appDir string) {
						Eventually(CF("run-task", appName, "sleep 1000")).Should(Exit(0))
					})
				})

				It("terminates the task", func() {
					tasksSession := CF("tasks", appName)
					Eventually(tasksSession).Should(Exit(0))
					Expect(tasksSession.Out).To(Say("1\\s+[a-zA-Z-0-9]+\\s+RUNNING"))

					session := CF("terminate-task", appName, "1")
					Eventually(session).Should(Exit(0))
					userName, _ := getCredentials()
					Expect(session.Out).To(Say(
						fmt.Sprintf("Terminating task 1 of app %s in org %s / space %s as %s..", appName, orgName, spaceName, userName)))
					Expect(session.Out).To(Say("OK"))
				})
			})

			Context("when the task is in the SUCCEEDED state", func() {
				BeforeEach(func() {
					WithSimpleApp(func(appDir string) {
						Eventually(CF("run-task", appName, "echo test")).Should(Exit(0))
					})
				})

				It("fails to terminate the task and prints an error", func() {
					Eventually(func() *Buffer {
						taskSession := CF("tasks", appName)
						Eventually(taskSession).Should(Exit(0))
						return taskSession.Out
					}).Should(Say("1\\s+[a-zA-Z-0-9]+\\s+SUCCEEDED"))

					session := CF("terminate-task", appName, "1")
					Eventually(session).Should(Exit(1))
					Expect(session.Err).To(Say("The request is semantically invalid: Task state is SUCCEEDED and therefore cannot be canceled"))
					Expect(session.Out).To(Say("FAILED"))
				})
			})

			Context("when the task is in the FAILED state", func() {
				BeforeEach(func() {
					WithSimpleApp(func(appDir string) {
						Eventually(CF("run-task", appName, "false")).Should(Exit(0))
					})
				})

				It("fails to terminate the task and prints an error", func() {
					Eventually(func() *Buffer {
						taskSession := CF("tasks", appName)
						Eventually(taskSession).Should(Exit(0))
						return taskSession.Out
					}).Should(Say("1\\s+[a-zA-Z-0-9]+\\s+FAILED"))

					session := CF("terminate-task", appName, "1")
					Eventually(session).Should(Exit(1))
					Expect(session.Err).To(Say("The request is semantically invalid: Task state is FAILED and therefore cannot be canceled"))
					Expect(session.Out).To(Say("FAILED"))
				})
			})

			Context("when the task ID does not exist", func() {
				It("fails to terminate the task and prints an error", func() {
					session := CF("terminate-task", appName, "1")
					Eventually(session).Should(Exit(1))
					Expect(session.Err).To(Say("Task sequence ID 1 not found."))
					Expect(session.Out).To(Say("FAILED"))
				})
			})
		})
	})
})
