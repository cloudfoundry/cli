package isolated

import (
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
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
		When("--help flag is set", func() {
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

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "scale", "app-name")
		})
	})

	When("the environment is set up correctly", func() {
		var userName string

		BeforeEach(func() {
			helpers.SetupCF(orgName, spaceName)
			userName, _ = helpers.GetCredentials()
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		When("the app name is not provided", func() {
			It("tells the user that the app name is required, prints help text, and exits 1", func() {
				session := helpers.CF("scale")

				Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `APP_NAME` was not provided"))
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the app does not exist", func() {
			It("tells the user that the app is not found and exits 1", func() {
				session := helpers.CF("scale", appName)

				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("App '%s' not found", appName))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the app does exist", func() {
			var (
				domainName string
			)

			BeforeEach(func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "push", appName)).Should(Exit(0))
				})

				domainName = helpers.DefaultSharedDomain()
			})

			AfterEach(func() {
				Eventually(helpers.CF("delete", appName, "-f", "-r")).Should(Exit(0))
			})

			When("scaling number of instances", func() {
				When("the wrong data type is provided to -i", func() {
					It("outputs an error message to the user, provides help text, and exits 1", func() {
						session := helpers.CF("scale", appName, "-i", "not-an-integer")
						Eventually(session.Err).Should(Say("Incorrect Usage: invalid argument for flag `-i' \\(expected int\\)"))
						Eventually(session).Should(Say("cf scale APP_NAME")) // help
						Eventually(session).Should(Exit(1))
					})
				})

				When("correct data is provided to -i", func() {
					It("scales application to specified number of instances", func() {
						session := helpers.CF("scale", appName, "-i", "2")
						Eventually(session).Should(Say("Scaling app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName)) // help
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Exit(0))
					})
				})
			})

			When("scaling memory", func() {
				When("the wrong data type is provided to -m", func() {
					It("outputs an error message to the user, provides help text, and exits 1", func() {
						session := helpers.CF("scale", appName, "-m", "not-a-memory")
						Eventually(session.Err).Should(Say("Incorrect Usage: invalid argument for flag `-m`"))
						Eventually(session).Should(Say("cf scale APP_NAME")) // help
						Eventually(session).Should(Exit(1))
					})
				})

				When("correct data is provided to -m", func() {
					When("-f flag is not provided", func() {
						var buffer *Buffer

						BeforeEach(func() {
							buffer = NewBuffer()
						})

						When("user enters y", func() {
							It("scales application to specified memory with restart", func() {
								_, err := buffer.Write([]byte("y\n"))
								Expect(err).NotTo(HaveOccurred())

								session := helpers.CFWithStdin(buffer, "scale", appName, "-m", "256M")
								Eventually(session).Should(Say("This will cause the app to restart. Are you sure you want to scale %s\\? \\[yN]]", appName))
								Eventually(session).Should(Say("Scaling app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName))
								Eventually(session).Should(Say("Stopping app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName))
								Eventually(session).Should(Say("Starting app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName))
								Eventually(session).Should(Say("Waiting for app to start\\.\\.\\."))

								Eventually(session).Should(Say("name:\\s+%s", appName))
								Eventually(session).Should(Say("requested state:\\s+started"))
								Eventually(session).Should(Say("instances:\\s+1/1"))
								Eventually(session).Should(Say("usage:\\s+256M x 1 instances"))
								Eventually(session).Should(Say("routes:\\s+%s.%s", appName, domainName))
								Eventually(session).Should(Say("last uploaded:"))
								Eventually(session).Should(Say("stack:\\s+cflinuxfs"))
								Eventually(session).Should(Say("buildpack:\\s+staticfile_buildpack"))
								Eventually(session).Should(Say("start command:"))

								Eventually(session).Should(Say("state\\s+since\\s+cpu\\s+memory\\s+disk\\s+details"))
								Eventually(session).Should(Say("#0\\s+(running|starting)\\s+.*\\d+\\.\\d+%.*of 256M.*of 1G"))
								Eventually(session).Should(Exit(0))
							})
						})

						When("user enters n", func() {
							It("does not scale the app", func() {
								_, err := buffer.Write([]byte("n\n"))
								Expect(err).NotTo(HaveOccurred())

								session := helpers.CFWithStdin(buffer, "scale", appName, "-m", "256M")
								Eventually(session).Should(Say("This will cause the app to restart. Are you sure you want to scale %s\\? \\[yN]]", appName))
								Eventually(session).Should(Say("Scale cancelled"))
								Eventually(session).Should(Exit(0))
								session = helpers.CF("scale", appName)
								Eventually(session).Should(Say("memoty:\\s+128M"))
								Eventually(session).Should(Exit(0))
							})
						})
					})

					When("-f flag provided", func() {
						It("scales without prompt", func() {
							session := helpers.CF("scale", appName, "-m", "256M", "-f")
							Eventually(session).Should(Say("Scaling app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName))
							Eventually(session).Should(Exit(0))
						})
					})
				})
			})

			When("scaling disk", func() {
				When("the wrong data type is provided to -k", func() {
					It("outputs an error message to the user, provides help text, and exits 1", func() {
						session := helpers.CF("scale", appName, "-k", "not-a-disk")
						Eventually(session.Err).Should(Say("Incorrect Usage: invalid argument for flag `-k`"))
						Eventually(session).Should(Say("cf scale APP_NAME")) // help
						Eventually(session).Should(Exit(1))
					})
				})

				It("scales application to specified disk with restart", func() {
					session := helpers.CF("scale", appName, "-k", "512M", "-f")
					Eventually(session).Should(Say("Scaling app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName))
					Eventually(session).Should(Say("Stopping app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName))
					Eventually(session).Should(Say("Starting app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName))
					Eventually(session).Should(Say("Waiting for app to start\\.\\.\\."))

					Eventually(session).Should(Say("name:\\s+%s", appName))
					Eventually(session).Should(Say("requested state:\\s+started"))
					Eventually(session).Should(Say("instances:\\s+1/1"))
					Eventually(session).Should(Say("usage:\\s+128M x 1 instances"))
					Eventually(session).Should(Say("routes:\\s+%s.%s", appName, domainName))
					Eventually(session).Should(Say("last uploaded:"))
					Eventually(session).Should(Say("stack:\\s+cflinuxfs"))
					Eventually(session).Should(Say("buildpack:\\s+staticfile_buildpack"))
					Eventually(session).Should(Say("start command:"))

					Eventually(session).Should(Say("state\\s+since\\s+cpu\\s+memory\\s+disk\\s+details"))
					Eventually(session).Should(Say("#0\\s+(running|starting)\\s+.*\\d+\\.\\d+%.*of 128M.*of 512M"))
					Eventually(session).Should(Exit(0))
				})
			})

			When("scaling all of them", func() {
				It("scales application to specified number of instances, memory and disk with restart", func() {
					session := helpers.CF("scale", appName, "-i", "2", "-m", "256M", "-k", "512M", "-f")
					Eventually(session).Should(Say("Scaling app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName))
					Eventually(session).Should(Say("Stopping app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName))
					Eventually(session).Should(Say("Starting app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName))
					Eventually(session).Should(Say("Waiting for app to start\\.\\.\\."))

					Eventually(session).Should(Say("name:\\s+%s", appName))
					Eventually(session).Should(Say("requested state:\\s+started"))
					Eventually(session).Should(Say("instances:\\s+2/2"))
					Eventually(session).Should(Say("usage:\\s+256M x 2 instances"))
					Eventually(session).Should(Say("routes:\\s+%s.%s", appName, domainName))
					Eventually(session).Should(Say("last uploaded:"))
					Eventually(session).Should(Say("stack:\\s+cflinuxfs"))
					Eventually(session).Should(Say("buildpack:\\s+staticfile_buildpack"))
					Eventually(session).Should(Say("start command:"))

					Eventually(session).Should(Say("state\\s+since\\s+cpu\\s+memory\\s+disk\\s+details"))
					Eventually(session).Should(Say("#0\\s+(running|starting)\\s+.*\\d+\\.\\d+%.*of 256M.*of 512M"))
					Eventually(session).Should(Say("#1\\s+(running|starting)\\s+.*\\d+\\.\\d+%.*of 256M.*of 512M"))
					Eventually(session).Should(Exit(0))
				})
			})

			When("scaling argument is not provided", func() {
				It("outputs current scaling information", func() {
					session := helpers.CF("scale", appName)
					Eventually(session).Should(Say("Showing current scale of app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName))

					Eventually(session).Should(Say("memory: 128M"))
					Eventually(session).Should(Say("disk: 1G"))
					Eventually(session).Should(Say("instances: 1"))

					Eventually(session).Should(Exit(0))
				})
			})
		})
	})
})
