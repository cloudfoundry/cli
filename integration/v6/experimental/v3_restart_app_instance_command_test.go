package experimental

import (
	"fmt"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("v3-restart-app-instance command", func() {
	var (
		orgName   string
		spaceName string
		appName   string
	)

	BeforeEach(func() {
		orgName = helpers.NewOrgName()
		spaceName = helpers.NewSpaceName()
		appName = helpers.PrefixedRandomName("app")
	})

	When("--help flag is set", func() {
		It("Displays command usage to output", func() {
			session := helpers.CF("v3-restart-app-instance", "--help")
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Say("v3-restart-app-instance - Terminate, then instantiate an app instance"))
			Eventually(session).Should(Say("USAGE:"))
			Eventually(session).Should(Say(`cf v3-restart-app-instance APP_NAME INDEX [--process PROCESS]`))
			Eventually(session).Should(Say("SEE ALSO:"))
			Eventually(session).Should(Say("v3-restart"))
			Eventually(session).Should(Exit(0))
		})
	})

	When("the app name is not provided", func() {
		It("tells the user that the app name is required, prints help text, and exits 1", func() {
			session := helpers.CF("v3-restart-app-instance")

			Eventually(session.Err).Should(Say("Incorrect Usage: the required arguments `APP_NAME` and `INDEX` were not provided"))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	When("the index is not provided", func() {
		It("tells the user that the index is required, prints help text, and exits 1", func() {
			session := helpers.CF("v3-restart-app-instance", appName)

			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `INDEX` was not provided"))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	It("displays the experimental warning", func() {
		session := helpers.CF("v3-restart-app-instance", appName, "1")
		Eventually(session.Err).Should(Say("This command is in EXPERIMENTAL stage and may change without notice"))
		Eventually(session).Should(Exit())
	})

	When("the environment is not setup correctly", func() {
		When("no API endpoint is set", func() {
			BeforeEach(func() {
				helpers.UnsetAPI()
			})

			It("fails with no API endpoint set message", func() {
				session := helpers.CF("v3-restart-app-instance", appName, "1")
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("No API endpoint set. Use 'cf login' or 'cf api' to target an endpoint."))
				Eventually(session).Should(Exit(1))
			})
		})

		When("not logged in", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
			})

			It("fails with not logged in message", func() {
				session := helpers.CF("v3-restart-app-instance", appName, "1")
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("Not logged in. Use 'cf login' or 'cf login --sso' to log in."))
				Eventually(session).Should(Exit(1))
			})
		})

		When("there is no org set", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
				helpers.LoginCF()
			})

			It("fails with no targeted org error message", func() {
				session := helpers.CF("v3-restart-app-instance", appName, "1")
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("No org targeted, use 'cf target -o ORG' to target an org."))
				Eventually(session).Should(Exit(1))
			})
		})

		When("there is no space set", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
				helpers.LoginCF()
				helpers.TargetOrg(ReadOnlyOrg)
			})

			It("fails with no targeted space error message", func() {
				session := helpers.CF("v3-restart-app-instance", appName, "1")
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("No space targeted, use 'cf target -s SPACE' to target a space."))
				Eventually(session).Should(Exit(1))
			})
		})
	})

	When("the environment is setup correctly", func() {
		var userName string

		BeforeEach(func() {
			helpers.SetupCF(orgName, spaceName)
			userName, _ = helpers.GetCredentials()
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		When("app does not exist", func() {
			It("fails with error", func() {
				session := helpers.CF("v3-restart-app-instance", appName, "0", "--process", "some-process")
				Eventually(session).Should(Say("Restarting instance 0 of process some-process of app %s in org %s / space %s as %s", appName, orgName, spaceName, userName))
				Eventually(session.Err).Should(Say("App '%s' not found", appName))
				Eventually(session).Should(Exit(1))
			})
		})

		When("app exists", func() {
			BeforeEach(func() {
				helpers.WithProcfileApp(func(appDir string) {
					Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "v3-push", appName)).Should(Exit(0))
				})
			})

			When("process type is not provided", func() {
				It("defaults to web process", func() {
					appOutputSession := helpers.CF("app", appName)
					Eventually(appOutputSession).Should(Exit(0))
					firstAppTable := helpers.ParseV3AppProcessTable(appOutputSession.Out.Contents())

					session := helpers.CF("v3-restart-app-instance", appName, "0")
					Eventually(session).Should(Say("Restarting instance 0 of process web of app %s in org %s / space %s as %s", appName, orgName, spaceName, userName))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Exit(0))

					Eventually(func() string {
						var restartedAppTable helpers.AppTable
						Eventually(func() string {
							appOutputSession := helpers.CF("app", appName)
							Eventually(appOutputSession).Should(Exit(0))
							restartedAppTable = helpers.ParseV3AppProcessTable(appOutputSession.Out.Contents())

							if len(restartedAppTable.Processes) > 0 {
								return fmt.Sprintf("%s, %s", restartedAppTable.Processes[0].Type, restartedAppTable.Processes[0].InstanceCount)
							}

							return ""
						}).Should(Equal(`web, 1/1`))
						Expect(restartedAppTable.Processes[0].Instances).ToNot(BeEmpty())
						return restartedAppTable.Processes[0].Instances[0].Since
					}).ShouldNot(Equal(firstAppTable.Processes[0].Instances[0].Since))
				})
			})

			When("a process type is provided", func() {
				When("the process type does not exist", func() {
					It("fails with error", func() {
						session := helpers.CF("v3-restart-app-instance", appName, "0", "--process", "unknown-process")
						Eventually(session).Should(Say("Restarting instance 0 of process unknown-process of app %s in org %s / space %s as %s", appName, orgName, spaceName, userName))
						Eventually(session.Err).Should(Say("Process unknown-process not found"))
						Eventually(session).Should(Exit(1))
					})
				})

				When("the process type exists", func() {
					When("instance index exists", func() {
						findConsoleProcess := func(appTable helpers.AppTable) (helpers.AppProcessTable, bool) {
							for _, process := range appTable.Processes {
								if process.Type == "console" {
									return process, true
								}
							}
							return helpers.AppProcessTable{}, false
						}

						It("defaults to requested process", func() {
							By("scaling worker process to 1 instance")
							session := helpers.CF("v3-scale", appName, "--process", "console", "-i", "1")
							Eventually(session).Should(Exit(0))

							By("waiting for worker process to come up")
							var firstAppTableConsoleProcess helpers.AppProcessTable
							Eventually(func() string {
								appOutputSession := helpers.CF("app", appName)
								Eventually(appOutputSession).Should(Exit(0))
								firstAppTable := helpers.ParseV3AppProcessTable(appOutputSession.Out.Contents())

								var found bool
								firstAppTableConsoleProcess, found = findConsoleProcess(firstAppTable)
								Expect(found).To(BeTrue())
								return fmt.Sprintf("%s, %s", firstAppTableConsoleProcess.Type, firstAppTableConsoleProcess.InstanceCount)
							}).Should(MatchRegexp(`console, 1/1`))

							By("restarting worker process instance")
							session = helpers.CF("v3-restart-app-instance", appName, "0", "--process", "console")
							Eventually(session).Should(Say("Restarting instance 0 of process console of app %s in org %s / space %s as %s", appName, orgName, spaceName, userName))
							Eventually(session).Should(Say("OK"))
							Eventually(session).Should(Exit(0))

							By("waiting for restarted process instance to come up")
							Eventually(func() string {
								var restartedAppTableConsoleProcess helpers.AppProcessTable

								Eventually(func() string {
									appOutputSession := helpers.CF("app", appName)
									Eventually(appOutputSession).Should(Exit(0))

									restartedAppTable := helpers.ParseV3AppProcessTable(appOutputSession.Out.Contents())
									var found bool
									restartedAppTableConsoleProcess, found = findConsoleProcess(restartedAppTable)
									Expect(found).To(BeTrue())

									return fmt.Sprintf("%s, %s", restartedAppTableConsoleProcess.Type, restartedAppTableConsoleProcess.InstanceCount)
								}).Should(MatchRegexp(`console, 1/1`))

								return restartedAppTableConsoleProcess.Instances[0].Since
							}).ShouldNot(Equal(firstAppTableConsoleProcess.Instances[0].Since))
						})
					})

					When("instance index does not exist", func() {
						It("fails with error", func() {
							session := helpers.CF("v3-restart-app-instance", appName, "42", "--process", constant.ProcessTypeWeb)
							Eventually(session).Should(Say("Restarting instance 42 of process web of app %s in org %s / space %s as %s", appName, orgName, spaceName, userName))
							Eventually(session.Err).Should(Say("Instance 42 of process web not found"))
							Eventually(session).Should(Exit(1))
						})
					})
				})
			})
		})
	})
})
