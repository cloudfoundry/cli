package experimental

import (
	"fmt"
	"strings"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("v3-scale command", func() {
	var (
		orgName   string
		spaceName string
		appName   string
		userName  string
	)

	BeforeEach(func() {
		orgName = helpers.NewOrgName()
		spaceName = helpers.NewSpaceName()
		appName = helpers.PrefixedRandomName("app")
		userName, _ = helpers.GetCredentials()
	})

	Describe("help", func() {
		Context("when --help flag is set", func() {
			It("displays command usage to output", func() {
				session := helpers.CF("v3-scale", "--help")

				Eventually(session.Out).Should(Say("NAME:"))
				Eventually(session.Out).Should(Say("v3-scale - Change or view the instance count, disk space limit, and memory limit for an app"))

				Eventually(session.Out).Should(Say("USAGE:"))
				Eventually(session.Out).Should(Say("cf v3-scale APP_NAME \\[--process PROCESS\\] \\[-i INSTANCES\\] \\[-k DISK\\] \\[-m MEMORY\\]"))

				Eventually(session.Out).Should(Say("OPTIONS:"))
				Eventually(session.Out).Should(Say("-f\\s+Force restart of app without prompt"))
				Eventually(session.Out).Should(Say("-i\\s+Number of instances"))
				Eventually(session.Out).Should(Say("-k\\s+Disk limit \\(e\\.g\\. 256M, 1024M, 1G\\)"))
				Eventually(session.Out).Should(Say("-m\\s+Memory limit \\(e\\.g\\. 256M, 1024M, 1G\\)"))
				Eventually(session.Out).Should(Say("--process\\s+App process to scale \\(Default: web\\)"))

				Eventually(session.Out).Should(Say("ENVIRONMENT:"))
				Eventually(session.Out).Should(Say("CF_STARTUP_TIMEOUT=5\\s+Max wait time for app instance startup, in minutes"))

				Eventually(session).Should(Exit(0))
			})
		})
	})

	It("displays the experimental warning", func() {
		session := helpers.CF("v3-scale", appName)
		Eventually(session.Out).Should(Say("This command is in EXPERIMENTAL stage and may change without notice"))
		Eventually(session).Should(Exit())
	})

	Context("when the environment is not setup correctly", func() {
		Context("when no API endpoint is set", func() {
			BeforeEach(func() {
				helpers.UnsetAPI()
			})

			It("fails with no API endpoint set message", func() {
				session := helpers.CF("v3-scale", appName)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("No API endpoint set\\. Use 'cf login' or 'cf api' to target an endpoint\\."))
				Eventually(session).Should(Exit(1))
			})
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
				session := helpers.CF("v3-scale", appName)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("This command requires CF API version 3\\.27\\.0 or higher\\."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the v3 api version is lower than the minimum version", func() {
			var server *Server

			BeforeEach(func() {
				server = helpers.StartAndTargetServerWithV3Version("3.0.0")
			})

			AfterEach(func() {
				server.Close()
			})

			It("fails with error message that the minimum version is not met", func() {
				session := helpers.CF("v3-scale", appName)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("This command requires CF API version 3\\.27\\.0 or higher\\."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when not logged in", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
			})

			It("fails with not logged in message", func() {
				session := helpers.CF("v3-scale", appName)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("Not logged in\\. Use 'cf login' to log in\\."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when there is no org set", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
				helpers.LoginCF()
			})

			It("fails with no org targeted error message", func() {
				session := helpers.CF("v3-scale", appName)
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("No org targeted, use 'cf target -o ORG' to target an org\\."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when there is no space set", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
				helpers.LoginCF()
				helpers.TargetOrg(ReadOnlyOrg)
			})

			It("fails with no space targeted error message", func() {
				session := helpers.CF("v3-scale", appName)
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("No space targeted, use 'cf target -s SPACE' to target a space\\."))
				Eventually(session).Should(Exit(1))
			})
		})
	})

	Context("when the environment is set up correctly", func() {
		BeforeEach(func() {
			setupCF(orgName, spaceName)
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		Context("when the app name is not provided", func() {
			It("tells the user that the app name is required, prints help text, and exits 1", func() {
				session := helpers.CF("v3-scale")

				Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `APP_NAME` was not provided"))
				Eventually(session.Out).Should(Say("NAME:"))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the app does not exist", func() {
			It("displays app not found and exits 1", func() {
				invalidAppName := "invalid-app-name"
				session := helpers.CF("v3-scale", invalidAppName)
				Eventually(session.Err).Should(Say("App %s not found", invalidAppName))
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the app exists", func() {
			BeforeEach(func() {
				helpers.WithProcfileApp(func(appDir string) {
					Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "v3-push", appName)).Should(Exit(0))
				})
			})

			Context("when scale option flags are not provided", func() {
				It("displays the current scale properties for all processes", func() {
					session := helpers.CF("v3-scale", appName)

					Eventually(session.Out).Should(Say("Showing current scale of app %s in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, userName))
					Consistently(session.Out).ShouldNot(Say("Scaling"))
					Consistently(session.Out).ShouldNot(Say("This will cause the app to restart"))
					Consistently(session.Out).ShouldNot(Say("Stopping"))
					Consistently(session.Out).ShouldNot(Say("Starting"))
					Consistently(session.Out).ShouldNot(Say("Waiting"))
					Eventually(session).Should(Exit(0))

					appTable := helpers.ParseV3AppProcessTable(session.Out.Contents())
					Expect(len(appTable.Processes)).To(Equal(3))

					processSummary := appTable.Processes[0]
					Expect(processSummary.Title).To(Equal("web:1/1"))

					instanceSummary := processSummary.Instances[0]
					Expect(instanceSummary.Memory).To(MatchRegexp(`\d+(\.\d+)?[KMG]? of \d+[KMG]`))
					Expect(instanceSummary.Disk).To(MatchRegexp(`\d+(\.\d+)?[KMG]? of \d+[KMG]`))

					Expect(appTable.Processes[1].Title).To(Equal("console:0/0"))
					Expect(appTable.Processes[2].Title).To(Equal("rake:0/0"))
				})
			})

			Context("when only one scale option flag is provided", func() {
				It("scales the app accordingly", func() {
					By("verifying we start with a single instance")
					session := helpers.CF("v3-scale", appName)
					Eventually(session).Should(Exit(0))
					appTable := helpers.ParseV3AppProcessTable(session.Out.Contents())
					Expect(appTable.Processes).To(HaveLen(3))

					By("scaling to 3 instances")
					session = helpers.CF("v3-scale", appName, "-i", "3")
					Eventually(session.Out).Should(Say("Scaling app %s in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, userName))
					Consistently(session.Out).ShouldNot(Say("This will cause the app to restart"))
					Consistently(session.Out).ShouldNot(Say("Stopping"))
					Consistently(session.Out).ShouldNot(Say("Starting"))
					Eventually(session).Should(Exit(0))

					updatedAppTable := helpers.ParseV3AppProcessTable(session.Out.Contents())
					Expect(updatedAppTable.Processes).To(HaveLen(3))

					processSummary := updatedAppTable.Processes[0]
					instanceSummary := processSummary.Instances[0]
					Expect(processSummary.Title).To(MatchRegexp(`web:\d/3`))
					Expect(instanceSummary.Memory).To(MatchRegexp(`\d+(\.\d+)?[KMG]? of \d+[KMG]`))
					Expect(instanceSummary.Disk).To(MatchRegexp(`\d+(\.\d+)?[KMG]? of \d+[KMG]`))

					By("scaling memory to 64M")
					buffer := NewBuffer()
					buffer.Write([]byte("y\n"))
					session = helpers.CFWithStdin(buffer, "v3-scale", appName, "-m", "64M")
					Eventually(session.Out).Should(Say("Scaling app %s in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, userName))
					Eventually(session.Out).Should(Say("This will cause the app to restart\\. Are you sure you want to scale %s\\? \\[yN\\]:", appName))
					Eventually(session.Out).Should(Say("Stopping app %s in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, userName))
					Eventually(session.Out).Should(Say("Starting app %s in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, userName))
					Eventually(session).Should(Exit(0))

					updatedAppTable = helpers.ParseV3AppProcessTable(session.Out.Contents())
					Expect(updatedAppTable.Processes).To(HaveLen(3))

					processSummary = updatedAppTable.Processes[0]
					instanceSummary = processSummary.Instances[0]
					Expect(processSummary.Title).To(MatchRegexp(`web:\d/3`))
					Expect(instanceSummary.Memory).To(MatchRegexp(`\d+(\.\d+)?[KMG]? of 64M`))
					Expect(instanceSummary.Disk).To(MatchRegexp(`\d+(\.\d+)?[KMG]? of \d+[KMG]`))

					By("scaling disk to 92M")
					buffer = NewBuffer()
					buffer.Write([]byte("y\n"))
					session = helpers.CFWithStdin(buffer, "v3-scale", appName, "-k", "92M")
					Eventually(session.Out).Should(Say("Scaling app %s in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, userName))
					Eventually(session.Out).Should(Say("This will cause the app to restart\\. Are you sure you want to scale %s\\? \\[yN\\]:", appName))
					Eventually(session.Out).Should(Say("Stopping app %s in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, userName))
					Eventually(session.Out).Should(Say("Starting app %s in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, userName))
					Eventually(session).Should(Exit(0))

					updatedAppTable = helpers.ParseV3AppProcessTable(session.Out.Contents())
					Expect(updatedAppTable.Processes).To(HaveLen(3))

					processSummary = updatedAppTable.Processes[0]
					instanceSummary = processSummary.Instances[0]
					Expect(processSummary.Title).To(MatchRegexp(`web:\d/3`))
					Expect(instanceSummary.Memory).To(MatchRegexp(`\d+(\.\d+)?[KMG]? of 64M`))
					Expect(instanceSummary.Disk).To(MatchRegexp(`\d+(\.\d+)?[KMG]? of 92M`))

					By("scaling to 0 instances")
					session = helpers.CF("v3-scale", appName, "-i", "0")
					Eventually(session.Out).Should(Say("Scaling app %s in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, userName))
					Consistently(session.Out).ShouldNot(Say("This will cause the app to restart"))
					Consistently(session.Out).ShouldNot(Say("Stopping"))
					Consistently(session.Out).ShouldNot(Say("Starting"))
					Eventually(session).Should(Exit(0))

					updatedAppTable = helpers.ParseV3AppProcessTable(session.Out.Contents())
					Expect(updatedAppTable.Processes).To(BeEmpty())
				})

				Context("when the user chooses not to restart the app", func() {
					It("cancels the scale", func() {
						buffer := NewBuffer()
						buffer.Write([]byte("n\n"))
						session := helpers.CFWithStdin(buffer, "v3-scale", appName, "-i", "2", "-k", "90M")
						Eventually(session.Out).Should(Say("This will cause the app to restart"))
						Consistently(session.Out).ShouldNot(Say("Stopping"))
						Consistently(session.Out).ShouldNot(Say("Starting"))
						Eventually(session.Out).Should(Say("Scaling cancelled"))
						Consistently(session.Out).ShouldNot(Say("Waiting for app to start\\.\\.\\."))
						Eventually(session).Should(Exit(0))

						appTable := helpers.ParseV3AppProcessTable(session.Out.Contents())
						Expect(appTable.Processes).To(BeEmpty())
					})
				})
			})

			Context("when all scale option flags are provided", func() {
				Context("when the app starts successfully", func() {
					It("scales the app accordingly", func() {
						buffer := NewBuffer()
						buffer.Write([]byte("y\n"))
						session := helpers.CFWithStdin(buffer, "v3-scale", appName, "-i", "2", "-k", "120M", "-m", "60M")
						Eventually(session.Out).Should(Say("Scaling app %s in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, userName))
						Eventually(session.Out).Should(Say("This will cause the app to restart\\. Are you sure you want to scale %s\\? \\[yN\\]:", appName))
						Eventually(session.Out).Should(Say("Stopping app %s in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, userName))
						Eventually(session.Out).Should(Say("Starting app %s in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, userName))
						Eventually(session).Should(Exit(0))

						appTable := helpers.ParseV3AppProcessTable(session.Out.Contents())
						Expect(appTable.Processes).To(HaveLen(3))

						processSummary := appTable.Processes[0]
						instanceSummary := processSummary.Instances[0]
						Expect(processSummary.Title).To(MatchRegexp(`web:\d/2`))
						Expect(instanceSummary.State).To(MatchRegexp(`running|starting`))
						Expect(instanceSummary.Memory).To(MatchRegexp(`\d+(\.\d+)?[KMG]? of 60M`))
						Expect(instanceSummary.Disk).To(MatchRegexp(`\d+(\.\d+)?[KMG]? of 120M`))
					})
				})

				Context("when the app does not start successfully", func() {
					It("scales the app and displays the app summary", func() {
						buffer := NewBuffer()
						buffer.Write([]byte("y\n"))
						session := helpers.CFWithStdin(buffer, "v3-scale", appName, "-i", "2", "-k", "120M", "-m", "6M")
						Eventually(session.Out).Should(Say("Scaling app %s in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, userName))
						Eventually(session.Out).Should(Say("This will cause the app to restart\\. Are you sure you want to scale %s\\? \\[yN\\]:", appName))
						Eventually(session.Out).Should(Say("Stopping app %s in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, userName))
						Eventually(session.Out).Should(Say("Starting app %s in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, userName))
						Eventually(session).Should(Exit(0))

						appTable := helpers.ParseV3AppProcessTable(session.Out.Contents())
						Expect(appTable.Processes).To(HaveLen(3))

						processSummary := appTable.Processes[0]
						instanceSummary := processSummary.Instances[0]
						Expect(processSummary.Title).To(MatchRegexp(`web:\d/2`))
						Expect(instanceSummary.State).To(MatchRegexp(`crashed`))
						Expect(instanceSummary.Memory).To(MatchRegexp(`\d+(\.\d+)?[KMG]? of 6M`))
						Expect(instanceSummary.Disk).To(MatchRegexp(`\d+(\.\d+)?[KMG]? of 120M`))
					})
				})
			})

			PContext("when the provided scale options are the same as the existing scale properties", func() {
				var (
					session          *Session
					currentInstances string
					maxMemory        string
					maxDiskSize      string
				)

				BeforeEach(func() {
					session = helpers.CF("v3-scale", appName)
					Eventually(session).Should(Exit(0))

					appTable := helpers.ParseV3AppProcessTable(session.Out.Contents())
					instanceSummary := appTable.Processes[0].Instances[0]
					currentInstances = string(len(appTable.Processes[0].Instances))
					maxMemory = strings.Fields(instanceSummary.Memory)[2]
					maxDiskSize = strings.Fields(instanceSummary.Disk)[2]
				})

				It("the action should be a no-op", func() {
					session = helpers.CF("v3-scale", appName, "-i", currentInstances, "-m", maxMemory, "-k", maxDiskSize)
					Eventually(session.Out).Should(Say("Scaling app %s in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, userName))
					Consistently(session.Out).ShouldNot(Say("This will cause the app to restart"))
					Consistently(session.Out).ShouldNot(Say("Stopping"))
					Consistently(session.Out).ShouldNot(Say("Starting"))
					Consistently(session.Out).ShouldNot(Say("Waiting for app to start"))
					Eventually(session).Should(Exit(0))

					appTable := helpers.ParseV3AppProcessTable(session.Out.Contents())
					Expect(appTable.Processes).To(HaveLen(1))

					newProcessSummary := appTable.Processes[0]
					newInstanceSummary := newProcessSummary.Instances[0]
					Expect(newProcessSummary.Title).To(MatchRegexp(fmt.Sprintf(`web:\d/%s`, currentInstances)))
					Expect(newInstanceSummary.Memory).To(MatchRegexp(fmt.Sprintf(`\d+(\.\d+)?[KMG]? of %s`, maxMemory)))
					Expect(newInstanceSummary.Disk).To(MatchRegexp(fmt.Sprintf(`\d+(\.\d+)?[KMG]? of %s`, maxDiskSize)))
				})
			})

			Context("when the process flag is provided", func() {
				It("scales the requested process", func() {
					session := helpers.CF("v3-scale", appName, "-i", "2", "--process", "console")
					Eventually(session.Out).Should(Say("Scaling app %s in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, userName))
					Eventually(session).Should(Exit(0))

					appTable := helpers.ParseV3AppProcessTable(session.Out.Contents())
					Expect(appTable.Processes).To(HaveLen(3))

					processSummary := appTable.Processes[1]
					instanceSummary := processSummary.Instances[0]
					Expect(processSummary.Instances).To(HaveLen(2))
					Expect(processSummary.Title).To(MatchRegexp(`console:\d/2`))
					Expect(instanceSummary.Memory).To(MatchRegexp(`\d+(\.\d+)?[KMG]? of \d+[KMG]`))
					Expect(instanceSummary.Disk).To(MatchRegexp(`\d+(\.\d+)?[KMG]? of \d+[KMG]`))
				})
			})
		})
	})

	Context("when invalid scale option values are provided", func() {
		Context("when a negative value is passed to a flag argument", func() {
			It("outputs an error message to the user, provides help text, and exits 1", func() {
				session := helpers.CF("v3-scale", "some-app", "-i=-5")
				Eventually(session.Err).Should(Say("Incorrect Usage: invalid argument for flag '-i' \\(expected int > 0\\)"))
				Eventually(session.Out).Should(Say("cf v3-scale APP_NAME")) // help
				Eventually(session).Should(Exit(1))

				session = helpers.CF("v3-scale", "some-app", "-k=-5")
				Eventually(session.Err).Should(Say("Byte quantity must be an integer with a unit of measurement like M, MB, G, or GB"))
				Eventually(session.Out).Should(Say("cf v3-scale APP_NAME")) // help
				Eventually(session).Should(Exit(1))

				session = helpers.CF("v3-scale", "some-app", "-m=-5")
				Eventually(session.Err).Should(Say("Byte quantity must be an integer with a unit of measurement like M, MB, G, or GB"))
				Eventually(session.Out).Should(Say("cf v3-scale APP_NAME")) // help
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when a non-integer value is passed to a flag argument", func() {
			It("outputs an error message to the user, provides help text, and exits 1", func() {
				session := helpers.CF("v3-scale", "some-app", "-i", "not-an-integer")
				Eventually(session.Err).Should(Say("Incorrect Usage: invalid argument for flag '-i' \\(expected int > 0\\)"))
				Eventually(session.Out).Should(Say("cf v3-scale APP_NAME")) // help
				Eventually(session).Should(Exit(1))

				session = helpers.CF("v3-scale", "some-app", "-k", "not-an-integer")
				Eventually(session.Err).Should(Say("Byte quantity must be an integer with a unit of measurement like M, MB, G, or GB"))
				Eventually(session.Out).Should(Say("cf v3-scale APP_NAME")) // help
				Eventually(session).Should(Exit(1))

				session = helpers.CF("v3-scale", "some-app", "-m", "not-an-integer")
				Eventually(session.Err).Should(Say("Byte quantity must be an integer with a unit of measurement like M, MB, G, or GB"))
				Eventually(session.Out).Should(Say("cf v3-scale APP_NAME")) // help
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the unit of measurement is not provided", func() {
			It("outputs an error message to the user, provides help text, and exits 1", func() {
				session := helpers.CF("v3-scale", "some-app", "-k", "9")
				Eventually(session.Err).Should(Say("Byte quantity must be an integer with a unit of measurement like M, MB, G, or GB"))
				Eventually(session.Out).Should(Say("cf v3-scale APP_NAME")) // help
				Eventually(session).Should(Exit(1))

				session = helpers.CF("v3-scale", "some-app", "-m", "7")
				Eventually(session.Err).Should(Say("Byte quantity must be an integer with a unit of measurement like M, MB, G, or GB"))
				Eventually(session.Out).Should(Say("cf v3-scale APP_NAME")) // help
				Eventually(session).Should(Exit(1))
			})
		})
	})
})
