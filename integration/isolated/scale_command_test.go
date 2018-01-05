package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = XDescribe("scale command", func() {
	var (
		orgName   string
		spaceName string
		appName   string
	)

	BeforeEach(func() {
		orgName = helpers.NewOrgName()
		spaceName = helpers.NewSpaceName()
		appName = helpers.NewAppName()
	})

	Describe("help", func() {
		Context("when --help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("scale", "--help")

				Eventually(session).Should(Say("scale - Change or view the instance count, disk space limit, and memory limit for an app"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf scale APP_NAME [-i INSTANCES] [-k DISK] [-m MEMORY] [-f]"))
				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say("-f\\s+Force restart of app without prompt"))
				Eventually(session).Should(Say("-i\\s+Number of instances"))
				Eventually(session).Should(Say("-k\\s+Disk limit (e.g. 256M, 1024M, 1G)"))
				Eventually(session).Should(Say("-m\\s+Memory limit (e.g. 256M, 1024M, 1G)"))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("push"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("when the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "scale", "app-name")
		})
	})

	Context("when the environment is set up correctly", func() {
		var userName string

		BeforeEach(func() {
			setupCF(orgName, spaceName)
			userName, _ = helpers.GetCredentials()
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		Context("when the app name is not provided", func() {
			It("tells the user that the app name is required, prints help text, and exits 1", func() {
				session := helpers.CF("scale")

				Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `APP_NAME` was not provided"))
				Eventually(session.Out).Should(Say("NAME:"))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the app does not exist", func() {
			It("tells the user that the app is not found and exits 1", func() {
				session := helpers.CF("scale", appName)

				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("App %s not found", appName))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the app does exist", func() {
			var (
				domainName string
			)

			BeforeEach(func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "v2-push", appName)).Should(Exit(0))
				})

				domainName = defaultSharedDomain()
			})

			AfterEach(func() {
				Eventually(helpers.CF("delete", appName, "-f", "-r")).Should(Exit(0))
			})

			Context("when scaling number of instances", func() {
				Context("when the wrong data type is provided to -i", func() {
					It("outputs an error message to the user, provides help text, and exits 1", func() {
						session := helpers.CF("scale", appName, "-i", "not-an-integer")
						Eventually(session.Err).Should(Say("Incorrect Usage: invalid argument for flag `-i' \\(expected int\\)"))
						Eventually(session.Out).Should(Say("cf scale APP_NAME")) // help
						Eventually(session).Should(Exit(1))
					})
				})

				Context("when correct data is provided to -i", func() {
					It("scales application to specified number of instances", func() {
						session := helpers.CF("scale", appName, "-i", "2")
						Eventually(session.Out).Should(Say("Scaling app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName)) // help
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Exit(0))
					})
				})
			})

			Context("when scaling memory", func() {
				Context("when the wrong data type is provided to -m", func() {
					It("outputs an error message to the user, provides help text, and exits 1", func() {
						session := helpers.CF("scale", appName, "-m", "not-a-memory")
						Eventually(session.Err).Should(Say("Incorrect Usage: invalid argument for flag `-m`"))
						Eventually(session.Out).Should(Say("cf scale APP_NAME")) // help
						Eventually(session).Should(Exit(1))
					})
				})

				Context("when correct data is provided to -m", func() {
					Context("when -f flag is not provided", func() {
						var buffer *Buffer

						BeforeEach(func() {
							buffer = NewBuffer()
						})

						Context("when user enters y", func() {
							It("scales application to specified memory with restart", func() {
								buffer.Write([]byte("y\n"))
								session := helpers.CFWithStdin(buffer, "scale", appName, "-m", "256M")
								Eventually(session.Out).Should(Say("This will cause the app to restart. Are you sure you want to scale %s\\? \\[yN]]", appName))
								Eventually(session.Out).Should(Say("Scaling app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName))
								Eventually(session.Out).Should(Say("Stopping app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName))
								Eventually(session.Out).Should(Say("Starting app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName))
								Eventually(session.Out).Should(Say("Waiting for app to start\\.\\.\\."))

								Eventually(session.Out).Should(Say("name:\\s+%s", appName))
								Eventually(session.Out).Should(Say("requested state:\\s+started"))
								Eventually(session.Out).Should(Say("instances:\\s+1/1"))
								Eventually(session.Out).Should(Say("usage:\\s+256M x 1 instances"))
								Eventually(session.Out).Should(Say("routes:\\s+%s.%s", appName, domainName))
								Eventually(session.Out).Should(Say("last uploaded:"))
								Eventually(session.Out).Should(Say("stack:\\s+cflinuxfs2"))
								Eventually(session.Out).Should(Say("buildpack:\\s+staticfile_buildpack"))
								Eventually(session.Out).Should(Say("start command:"))

								Eventually(session.Out).Should(Say("state\\s+since\\s+cpu\\s+memory\\s+disk\\s+details"))
								Eventually(session.Out).Should(Say("#0\\s+(running|starting)\\s+.*\\d+\\.\\d+%.*of 256M.*of 1G"))
								Eventually(session.Out).Should(Exit(0))
							})
						})

						Context("when user enters n", func() {
							It("does not scale the app", func() {
								buffer.Write([]byte("n\n"))
								session := helpers.CFWithStdin(buffer, "scale", appName, "-m", "256M")
								Eventually(session.Out).Should(Say("This will cause the app to restart. Are you sure you want to scale %s\\? \\[yN]]", appName))
								Eventually(session.Out).Should(Say("Scale cancelled"))
								Eventually(session).Should(Exit(0))
								session = helpers.CF("scale", appName)
								Eventually(session.Out).Should(Say("memoty:\\s+128M"))
								Eventually(session).Should(Exit(0))
							})
						})
					})

					Context("when -f flag provided", func() {
						It("scales without prompt", func() {
							session := helpers.CF("scale", appName, "-m", "256M", "-f")
							Eventually(session.Out).Should(Say("Scaling app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName))
							Eventually(session).Should(Exit(0))
						})
					})
				})
			})

			Context("when scaling disk", func() {
				Context("when the wrong data type is provided to -k", func() {
					It("outputs an error message to the user, provides help text, and exits 1", func() {
						session := helpers.CF("scale", appName, "-k", "not-a-disk")
						Eventually(session.Err).Should(Say("Incorrect Usage: invalid argument for flag `-k`"))
						Eventually(session.Out).Should(Say("cf scale APP_NAME")) // help
						Eventually(session).Should(Exit(1))
					})
				})

				It("scales application to specified disk with restart", func() {
					session := helpers.CF("scale", appName, "-k", "512M", "-f")
					Eventually(session.Out).Should(Say("Scaling app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName))
					Eventually(session.Out).Should(Say("Stopping app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName))
					Eventually(session.Out).Should(Say("Starting app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName))
					Eventually(session.Out).Should(Say("Waiting for app to start\\.\\.\\."))

					Eventually(session.Out).Should(Say("name:\\s+%s", appName))
					Eventually(session.Out).Should(Say("requested state:\\s+started"))
					Eventually(session.Out).Should(Say("instances:\\s+1/1"))
					Eventually(session.Out).Should(Say("usage:\\s+128M x 1 instances"))
					Eventually(session.Out).Should(Say("routes:\\s+%s.%s", appName, domainName))
					Eventually(session.Out).Should(Say("last uploaded:"))
					Eventually(session.Out).Should(Say("stack:\\s+cflinuxfs2"))
					Eventually(session.Out).Should(Say("buildpack:\\s+staticfile_buildpack"))
					Eventually(session.Out).Should(Say("start command:"))

					Eventually(session.Out).Should(Say("state\\s+since\\s+cpu\\s+memory\\s+disk\\s+details"))
					Eventually(session.Out).Should(Say("#0\\s+(running|starting)\\s+.*\\d+\\.\\d+%.*of 128M.*of 512M"))
					Eventually(session.Out).Should(Exit(0))
				})
			})

			Context("when scaling all of them", func() {
				It("scales application to specified number of instances, memory and disk with restart", func() {
					session := helpers.CF("scale", appName, "-i", "2", "-m", "256M", "-k", "512M", "-f")
					Eventually(session.Out).Should(Say("Scaling app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName))
					Eventually(session.Out).Should(Say("Stopping app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName))
					Eventually(session.Out).Should(Say("Starting app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName))
					Eventually(session.Out).Should(Say("Waiting for app to start\\.\\.\\."))

					Eventually(session.Out).Should(Say("name:\\s+%s", appName))
					Eventually(session.Out).Should(Say("requested state:\\s+started"))
					Eventually(session.Out).Should(Say("instances:\\s+2/2"))
					Eventually(session.Out).Should(Say("usage:\\s+256M x 2 instances"))
					Eventually(session.Out).Should(Say("routes:\\s+%s.%s", appName, domainName))
					Eventually(session.Out).Should(Say("last uploaded:"))
					Eventually(session.Out).Should(Say("stack:\\s+cflinuxfs2"))
					Eventually(session.Out).Should(Say("buildpack:\\s+staticfile_buildpack"))
					Eventually(session.Out).Should(Say("start command:"))

					Eventually(session.Out).Should(Say("state\\s+since\\s+cpu\\s+memory\\s+disk\\s+details"))
					Eventually(session.Out).Should(Say("#0\\s+(running|starting)\\s+.*\\d+\\.\\d+%.*of 256M.*of 512M"))
					Eventually(session.Out).Should(Say("#1\\s+(running|starting)\\s+.*\\d+\\.\\d+%.*of 256M.*of 512M"))
					Eventually(session.Out).Should(Exit(0))
				})
			})

			Context("when scaling argument is not provided", func() {
				It("outputs current scaling information", func() {
					session := helpers.CF("scale", appName)
					Eventually(session).Should(Say("Showing current scale of app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName))

					Eventually(session).Should(Say("memory: 128M"))
					Eventually(session).Should(Say("disk: 1G"))
					Eventually(session).Should(Say("instances: 1"))

					Eventually(session.Out).Should(Exit(0))
				})
			})
		})
	})
})
