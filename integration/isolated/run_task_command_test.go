package isolated

import (
	"fmt"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("run-task command", func() {
	Context("when --help flag is set", func() {
		It("Displays command usage to output", func() {
			session := helpers.CF("run-task", "--help")
			Eventually(session).Should(Exit(0))
			Expect(session.Out).To(Say("NAME:"))
			Expect(session.Out).To(Say("   run-task - Run a one-off task on an app"))
			Expect(session.Out).To(Say("USAGE:"))
			Expect(session.Out).To(Say("   cf run-task APP_NAME COMMAND [--name TASK_NAME]"))
			Expect(session.Out).To(Say("EXAMPLES:"))
			Expect(session.Out).To(Say(`   cf run-task my-app "bundle exec rake db:migrate" --name migrate`))
			Expect(session.Out).To(Say("ALIAS:"))
			Expect(session.Out).To(Say("   rt"))
			Expect(session.Out).To(Say("OPTIONS:"))
			Expect(session.Out).To(Say("   --name      Name to give the task \\(generated if omitted\\)"))
			Expect(session.Out).To(Say("SEE ALSO:"))
			Expect(session.Out).To(Say("   logs, tasks, terminate-task"))
		})
	})

	Context("when the environment is not setup correctly", func() {
		Context("when no API endpoint is set", func() {
			BeforeEach(func() {
				helpers.UnsetAPI()
			})

			It("fails with no API endpoint set message", func() {
				session := helpers.CF("run-task", "app-name", "some command")
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
				session := helpers.CF("run-task", "app-name", "some command")
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
				session := helpers.CF("run-task", "app-name", "some command")
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
				session := helpers.CF("run-task", "app-name", "some command")
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

		Context("when the application exists", func() {
			BeforeEach(func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CF("push", appName, "-p", appDir, "-b", "staticfile_buildpack")).Should(Exit(0))
				})
			})

			Context("when the task name is not provided", func() {
				It("creates a new task", func() {
					session := helpers.CF("run-task", appName, "echo hi")
					Eventually(session).Should(Exit(0))
					userName, _ := helpers.GetCredentials()
					Expect(session.Out).To(Say(fmt.Sprintf("Creating task for app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName)))
					Expect(session.Out).To(Say(`OK

Task 1 has been submitted successfully for execution.`,
					))
				})
			})

			Context("when the task name is provided", func() {
				It("creates a new task with the provided name", func() {
					session := helpers.CF("run-task", appName, "echo hi", "--name", "some-task-name")
					Eventually(session).Should(Exit(0))
					userName, _ := helpers.GetCredentials()
					Expect(session.Out).To(Say(fmt.Sprintf("Creating task for app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName)))
					Expect(session.Out).To(Say(`OK

Task 1 has been submitted successfully for execution.`,
					))

					Eventually(func() *Buffer {
						taskSession := helpers.CF("tasks", appName)
						Eventually(taskSession).Should(Exit(0))
						return taskSession.Out
					}).Should(Say("1\\s+some-task-name"))
				})
			})
		})

		Context("when the application is not staged", func() {
			BeforeEach(func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CF("push", appName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack")).Should(Exit(0))
				})
			})

			It("fails and outputs task must have a droplet message", func() {
				session := helpers.CF("run-task", appName, "echo hi")
				Eventually(session).Should(Exit(1))
				Expect(session.Out).To(Say("FAILED"))
				Expect(session.Err).To(Say(`Error running task: App is not staged.`))
			})
		})

		Context("when the application is staged but stopped", func() {
			BeforeEach(func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CF("push", appName, "-p", appDir, "-b", "staticfile_buildpack")).Should(Exit(0))
				})
				session := helpers.CF("stop", appName)
				Eventually(session).Should(Exit(0))
			})

			It("creates a new task", func() {
				session := helpers.CF("run-task", appName, "echo hi")
				Eventually(session).Should(Exit(0))
				userName, _ := helpers.GetCredentials()
				Expect(session.Out).To(Say(fmt.Sprintf("Creating task for app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName)))
				Expect(session.Out).To(Say(`OK

Task 1 has been submitted successfully for execution.`,
				))
			})
		})

		Context("when the application does not exist", func() {
			It("fails and outputs an app not found message", func() {
				session := helpers.CF("run-task", appName, "echo hi")
				Eventually(session).Should(Exit(1))
				Expect(session.Out).To(Say("FAILED"))
				Expect(session.Err).To(Say(fmt.Sprintf("App %s not found", appName)))
			})
		})
	})
})
