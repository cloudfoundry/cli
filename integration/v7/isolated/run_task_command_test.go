package isolated

import (
	"fmt"
	"path"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("run-task command", func() {
	When("--help flag is set", func() {
		It("Displays command usage to output", func() {
			session := helpers.CF("run-task", "--help")
			Eventually(session).Should(Exit(0))
			Expect(session).To(Say("NAME:"))
			Expect(session).To(Say("   run-task - Run a one-off task on an app"))
			Expect(session).To(Say("USAGE:"))
			Expect(session).To(Say(`   cf run-task APP_NAME \[--command COMMAND\] \[-k DISK] \[-m MEMORY\] \[-l LOG_RATE_LIMIT\] \[--name TASK_NAME\] \[--process PROCESS_TYPE\]`))
			Expect(session).To(Say("TIP:"))
			Expect(session).To(Say("   Use 'cf logs' to display the logs of the app and all its tasks. If your task name is unique, grep this command's output for the task name to view task-specific logs."))
			Expect(session).To(Say("EXAMPLES:"))
			Expect(session).To(Say(`   cf run-task my-app --command "bundle exec rake db:migrate" --name migrate`))
			Expect(session).To(Say("ALIAS:"))
			Expect(session).To(Say("   rt"))
			Expect(session).To(Say("OPTIONS:"))
			Expect(session).To(Say(`   --command, -c\s+The command to execute`))
			Expect(session).To(Say(`   -k                 Disk limit \(e\.g\. 256M, 1024M, 1G\)`))
			Expect(session).To(Say(`   -l                 Log rate limit per second, in bytes \(e\.g\. 128B, 4K, 1M\). -l=-1 represents unlimited`))
			Expect(session).To(Say(`   -m                 Memory limit \(e\.g\. 256M, 1024M, 1G\)`))
			Expect(session).To(Say(`   --name             Name to give the task \(generated if omitted\)`))
			Expect(session).To(Say(`   --process          Process type to use as a template for command, memory, and disk for the created task`))
			Expect(session).To(Say("SEE ALSO:"))
			Expect(session).To(Say("   logs, tasks, terminate-task"))
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "run-task", "app-name", "--command", "some-command")
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

		When("the application exists", func() {

			When("the app has a default task process", func() {
				BeforeEach(func() {
					helpers.WithTaskApp(func(appDir string) {
						Eventually(helpers.CF("push", appName, "-p", appDir, "-b", "staticfile_buildpack", "-f", path.Join(appDir, "manifest.yml"))).Should(Exit(0))
					}, appName)
				})

				It("creates a new task", func() {
					session := helpers.CF("run-task", appName)
					userName, _ := helpers.GetCredentials()
					Eventually(session).Should(Exit(0))
					Expect(session).To(Say("Creating task for app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName))
					Expect(session).To(Say("Task has been submitted successfully for execution."))
					Expect(session).To(Say("OK"))
					Expect(session).To(Say(`task name:\s+.+`))
					Expect(session).To(Say(`task id:\s+1`))
				})
			})

			When("the app is given a command flag", func() {

				BeforeEach(func() {
					helpers.WithHelloWorldApp(func(appDir string) {
						Eventually(helpers.CF("push", appName, "-p", appDir, "-b", "staticfile_buildpack")).Should(Exit(0))
					})
				})
				When("the task name is provided", func() {
					It("creates a new task with the provided name", func() {
						session := helpers.CF("run-task", appName, "--command", "echo hi", "--name", "some-task-name")
						userName, _ := helpers.GetCredentials()
						Eventually(session).Should(Exit(0))
						Expect(session).To(Say("Creating task for app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName))
						Expect(session).To(Say("Task has been submitted successfully for execution."))
						Expect(session).To(Say("OK"))
						Expect(session).To(Say(`task name:\s+some-task-name`))
						Expect(session).To(Say(`task id:\s+1`))

						taskSession := helpers.CF("tasks", appName)
						Eventually(taskSession).Should(Exit(0))
						Expect(taskSession).Should(Say(`1\s+some-task-name`))
					})
				})

				When("disk space is provided", func() {
					When("the provided disk space is invalid", func() {
						It("displays error and exits 1", func() {
							session := helpers.CF("run-task", appName, "--command", "echo hi", "-k", "invalid")
							Eventually(session).Should(Exit(1))
							Expect(session.Err).Should(Say("Byte quantity must be an integer with a unit of measurement like M, MB, G, or GB"))

						})
					})

					When("the provided disk space is valid", func() {
						It("runs the task with the provided disk space", func() {
							diskSpace := 123
							session := helpers.CF("run-task", appName, "--command", "echo hi", "-k", fmt.Sprintf("%dM", diskSpace))
							Eventually(session).Should(Exit(0))

							session = helpers.CF("tasks", appName, "-v")
							Eventually(session).Should(Exit(0))
							Expect(session).To(Say("\"disk_in_mb\": %d", diskSpace))
						})
					})
				})

				When("log rate limit is provided", func() {
					BeforeEach(func() {
						helpers.SkipIfVersionLessThan(ccversion.MinVersionLogRateLimitingV3)
					})
					When("the provided log rate limit is invalid", func() {
						It("displays error and exits 1", func() {
							session := helpers.CF("run-task", appName, "--command", "echo hi", "-l", "invalid")
							Eventually(session).Should(Exit(1))
							Expect(session.Err).Should(Say("Byte quantity must be an integer with a unit of measurement like B, K, KB, M, MB, G, or GB"))

						})
					})

					When("the provided log rate limit is valid", func() {
						It("runs the task with the provided log rate limit", func() {
							logRateLimit := 1024
							session := helpers.CF("run-task", appName, "--command", "echo hi", "-l", fmt.Sprintf("%dB", logRateLimit))
							Eventually(session).Should(Exit(0))

							session = helpers.CF("tasks", appName, "-v")
							Eventually(session).Should(Exit(0))
							Expect(session).To(Say("\"log_rate_limit_in_bytes_per_second\": %d", logRateLimit))
						})
					})
				})

				When("task memory is provided", func() {
					When("the provided memory is invalid", func() {
						It("displays error and exits 1", func() {
							session := helpers.CF("run-task", appName, "--command", "echo hi", "-m", "invalid")
							Eventually(session).Should(Exit(1))
							Expect(session.Err).Should(Say("Byte quantity must be an integer with a unit of measurement like M, MB, G, or GB"))
						})
					})

					When("the provided memory is valid", func() {
						It("runs the task with the provided memory", func() {
							taskMemory := 123
							session := helpers.CF("run-task", appName, "--command", " echo hi", "-m", fmt.Sprintf("%dM", taskMemory))
							Eventually(session).Should(Exit(0))

							session = helpers.CF("tasks", appName, "-v")
							Eventually(session).Should(Exit(0))
							Expect(session).To(Say("\"memory_in_mb\": %d", taskMemory))
						})
					})
				})
			})

		})

		When("the application is not staged", func() {
			BeforeEach(func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CF("push", appName, "--no-start", "-p", appDir, "-b", "staticfile_buildpack")).Should(Exit(0))
				})
			})

			It("fails and outputs task must have a droplet message", func() {
				session := helpers.CF("run-task", appName, "--command", "echo hi")
				Eventually(session).Should(Exit(1))
				Expect(session).To(Say("FAILED"))
				Expect(session.Err).To(Say(`Error running task: App is not staged.`))
			})
		})

		When("the application is staged but stopped", func() {
			BeforeEach(func() {
				helpers.WithTaskApp(func(appDir string) {
					Eventually(helpers.CF("push", appName, "-p", appDir, "-b", "staticfile_buildpack", "-f", path.Join(appDir, "manifest.yml"))).Should(Exit(0))
				}, appName)
				session := helpers.CF("stop", appName)
				Eventually(session).Should(Exit(0))
			})

			It("creates a new task", func() {
				session := helpers.CF("run-task", appName)
				Eventually(session).Should(Exit(0))
				userName, _ := helpers.GetCredentials()
				Expect(session).To(Say("Creating task for app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName))
				Expect(session).To(Say("Task has been submitted successfully for execution."))
				Expect(session).To(Say("OK"))
				Expect(session).To(Say(`task name:\s+.+`))
				Expect(session).To(Say(`task id:\s+1`))
			})
		})

		When("the application does not exist", func() {
			It("fails and outputs an app not found message", func() {
				session := helpers.CF("run-task", appName)
				Eventually(session).Should(Exit(1))
				Expect(session).To(Say("FAILED"))
				Expect(session.Err).To(Say(fmt.Sprintf("App '%s' not found", appName)))
			})
		})
	})
})
