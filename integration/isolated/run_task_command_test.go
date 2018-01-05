package isolated

import (
	"fmt"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("run-task command", func() {
	Context("when --help flag is set", func() {
		It("Displays command usage to output", func() {
			session := helpers.CF("run-task", "--help")
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Say("   run-task - Run a one-off task on an app"))
			Eventually(session).Should(Say("USAGE:"))
			Eventually(session).Should(Say("   cf run-task APP_NAME COMMAND \\[-k DISK] \\[-m MEMORY\\] \\[--name TASK_NAME\\]"))
			Eventually(session).Should(Say("TIP:"))
			Eventually(session).Should(Say("   Use 'cf logs' to display the logs of the app and all its tasks. If your task name is unique, grep this command's output for the task name to view task-specific logs."))
			Eventually(session).Should(Say("EXAMPLES:"))
			Eventually(session).Should(Say(`   cf run-task my-app "bundle exec rake db:migrate" --name migrate`))
			Eventually(session).Should(Say("ALIAS:"))
			Eventually(session).Should(Say("   rt"))
			Eventually(session).Should(Say("OPTIONS:"))
			Eventually(session).Should(Say("   -k          Disk limit \\(e\\.g\\. 256M, 1024M, 1G\\)"))
			Eventually(session).Should(Say("   -m          Memory limit \\(e\\.g\\. 256M, 1024M, 1G\\)"))
			Eventually(session).Should(Say("   --name      Name to give the task \\(generated if omitted\\)"))
			Eventually(session).Should(Say("SEE ALSO:"))
			Eventually(session).Should(Say("   logs, tasks, terminate-task"))
			Eventually(session).Should(Exit(0))
		})
	})

	Context("when the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "run-task", "app-name", "some-command")
		})

		Context("when the v3 api does not exist", func() {
			var server *Server

			BeforeEach(func() {
				server = helpers.StartAndTargetServerWithoutV3API()
			})

			AfterEach(func() {
				server.Close()
			})

			It("fails with error message that the minimum version is not met", func() {
				session := helpers.CF("run-task", "app-name", "some command")
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("This command requires CF API version 3\\.0\\.0 or higher\\."))
				Eventually(session).Should(Exit(1))
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
			spaceName = helpers.NewSpaceName()
			appName = helpers.PrefixedRandomName("APP")

			setupCF(orgName, spaceName)
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
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
					userName, _ := helpers.GetCredentials()
					Eventually(session).Should(Say("Creating task for app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Say("Task has been submitted successfully for execution."))
					Eventually(session).Should(Say("task name:\\s+.+"))
					Eventually(session).Should(Say("task id:\\s+1"))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when the task name is provided", func() {
				It("creates a new task with the provided name", func() {
					session := helpers.CF("run-task", appName, "echo hi", "--name", "some-task-name")
					userName, _ := helpers.GetCredentials()
					Eventually(session).Should(Say("Creating task for app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Say("Task has been submitted successfully for execution."))
					Eventually(session).Should(Say("task name:\\s+some-task-name"))
					Eventually(session).Should(Say("task id:\\s+1"))
					Eventually(session).Should(Exit(0))

					taskSession := helpers.CF("tasks", appName)
					Eventually(taskSession).Should(Say("1\\s+some-task-name"))
					Eventually(taskSession).Should(Exit(0))
				})
			})

			Context("when disk space is provided", func() {
				Context("when the provided disk space is invalid", func() {
					It("displays error and exits 1", func() {
						session := helpers.CF("run-task", appName, "echo hi", "-k", "invalid")
						Eventually(session.Err).Should(Say("Byte quantity must be an integer with a unit of measurement like M, MB, G, or GB"))

						Eventually(session).Should(Exit(1))
					})
				})

				Context("when the provided disk space is valid", func() {
					It("runs the task with the provided disk space", func() {
						diskSpace := 123
						session := helpers.CF("run-task", appName, "echo hi", "-k", fmt.Sprintf("%dM", diskSpace))
						Eventually(session).Should(Exit(0))

						session = helpers.CF("tasks", appName, "-v")
						Eventually(session).Should(Say("\"disk_in_mb\": %d", diskSpace))
						Eventually(session).Should(Exit(0))
					})
				})
			})

			Context("when task memory is provided", func() {
				Context("when the provided memory is invalid", func() {
					It("displays error and exits 1", func() {
						session := helpers.CF("run-task", appName, "echo hi", "-m", "invalid")
						Eventually(session.Err).Should(Say("Byte quantity must be an integer with a unit of measurement like M, MB, G, or GB"))
						Eventually(session).Should(Exit(1))
					})
				})

				Context("when the provided memory is valid", func() {
					It("runs the task with the provided memory", func() {
						taskMemory := 123
						session := helpers.CF("run-task", appName, "echo hi", "-m", fmt.Sprintf("%dM", taskMemory))
						Eventually(session).Should(Exit(0))

						session = helpers.CF("tasks", appName, "-v")
						Eventually(session).Should(Say("\"memory_in_mb\": %d", taskMemory))
						Eventually(session).Should(Exit(0))
					})
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
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say(`Error running task: App is not staged.`))
				Eventually(session).Should(Exit(1))
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
				userName, _ := helpers.GetCredentials()
				Eventually(session).Should(Say("Creating task for app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Say("Task has been submitted successfully for execution."))
				Eventually(session).Should(Say("task name:\\s+.+"))
				Eventually(session).Should(Say("task id:\\s+1"))
				Eventually(session).Should(Exit(0))
			})
		})

		Context("when the application does not exist", func() {
			It("fails and outputs an app not found message", func() {
				session := helpers.CF("run-task", appName, "echo hi")
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say(fmt.Sprintf("App %s not found", appName)))
				Eventually(session).Should(Exit(1))
			})
		})
	})
})
