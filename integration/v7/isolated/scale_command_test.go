package isolated

import (
	"fmt"
	"strings"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("scale command", func() {
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
		When("--help flag is set", func() {
			It("displays command usage to output", func() {
				session := helpers.CF("scale", "--help")

				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("scale - Change or view the instance count, disk space limit, and memory limit for an app"))

				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say(`cf scale APP_NAME \[--process PROCESS\] \[-i INSTANCES\] \[-k DISK\] \[-m MEMORY\]`))

				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say(`-f\s+Force restart of app without prompt`))
				Eventually(session).Should(Say(`-i\s+Number of instances`))
				Eventually(session).Should(Say(`-k\s+Disk limit \(e\.g\. 256M, 1024M, 1G\)`))
				Eventually(session).Should(Say(`-m\s+Memory limit \(e\.g\. 256M, 1024M, 1G\)`))
				Eventually(session).Should(Say(`--process\s+App process to scale \(Default: web\)`))

				Eventually(session).Should(Say("ENVIRONMENT:"))
				Eventually(session).Should(Say(`CF_STARTUP_TIMEOUT=5\s+Max wait time for app instance startup, in minutes`))

				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "scale", appName)
		})
	})

	When("the environment is set up correctly", func() {
		BeforeEach(func() {
			helpers.SetupCF(orgName, spaceName)
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
			It("displays app not found and exits 1", func() {
				invalidAppName := "invalid-app-name"
				session := helpers.CF("scale", invalidAppName)
				Eventually(session.Err).Should(Say("App %s not found", invalidAppName))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the app exists", func() {
			BeforeEach(func() {
				helpers.WithProcfileApp(func(appDir string) {
					Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "push", appName)).Should(Exit(0))
				})
			})

			When("scale option flags are not provided", func() {
				It("displays the current scale properties for all processes", func() {
					session := helpers.CF("scale", appName)

					Eventually(session).Should(Say(`Showing current scale of app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
					Consistently(session).ShouldNot(Say("Scaling"))
					Consistently(session).ShouldNot(Say("This will cause the app to restart"))
					Consistently(session).ShouldNot(Say("Stopping"))
					Consistently(session).ShouldNot(Say("Starting"))
					Consistently(session).ShouldNot(Say("Waiting"))
					Eventually(session).Should(Say(`name:\s+%s`, appName))
					Eventually(session).Should(Say(`requested state:\s+started`))
					Eventually(session).Should(Exit(0))

					appTable := helpers.ParseV3AppProcessTable(session.Out.Contents())
					Expect(len(appTable.Processes)).To(Equal(3))

					processSummary := appTable.Processes[0]
					Expect(processSummary.Type).To(Equal("web"))
					Expect(processSummary.InstanceCount).To(Equal("1/1"))

					instanceSummary := processSummary.Instances[0]
					Expect(instanceSummary.Memory).To(MatchRegexp(`\d+(\.\d+)?[KMG]? of \d+[KMG]`))
					Expect(instanceSummary.Disk).To(MatchRegexp(`\d+(\.\d+)?[KMG]? of \d+[KMG]`))

					Expect(appTable.Processes[1].Type).To(Equal("console"))
					Expect(appTable.Processes[1].InstanceCount).To(Equal("0/0"))

					Expect(appTable.Processes[2].Type).To(Equal("rake"))
					Expect(appTable.Processes[2].InstanceCount).To(Equal("0/0"))
				})
			})

			When("only one scale option flag is provided", func() {
				When("scaling the number of instances", func() {
					It("Scales to the correct number of instances", func() {
						By("Verifying we start with one instance")
						session := helpers.CF("scale", appName)
						Eventually(session).Should(Exit(0))
						appTable := helpers.ParseV3AppProcessTable(session.Out.Contents())
						Expect(appTable.Processes).To(HaveLen(3))
						processSummary := appTable.Processes[0]
						Expect(processSummary.Type).To(Equal("web"))
						Expect(processSummary.InstanceCount).To(Equal("1/1"))

						By("then scaling to 3 instances")
						session = helpers.CF("scale", appName, "-i", "3")
						Eventually(session).Should(Say(`Scaling app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
						Consistently(session).ShouldNot(Say("This will cause the app to restart"))
						Consistently(session).ShouldNot(Say("Stopping"))
						Consistently(session).ShouldNot(Say("Starting"))
						Eventually(session).Should(Exit(0))
						updatedAppTable := helpers.ParseV3AppProcessTable(session.Out.Contents())
						Expect(updatedAppTable.Processes).To(HaveLen(3))
						processSummary = updatedAppTable.Processes[0]
						instanceSummary := processSummary.Instances[0]
						Expect(processSummary.Type).To(Equal("web"))
						Expect(processSummary.InstanceCount).To(MatchRegexp(`\d/3`))
						Expect(instanceSummary.Disk).To(MatchRegexp(`\d+(\.\d+)?[KMG]? of \d+[KMG]`))
						Expect(instanceSummary.Memory).To(MatchRegexp(`\d+(\.\d+)?[KMG]? of \d+[KMG]`))
					})
				})

				When("Scaling the memory", func() {
					It("scales memory to 64M", func() {
						buffer := NewBuffer()
						buffer.Write([]byte("y\n"))
						session := helpers.CFWithStdin(buffer, "scale", appName, "-m", "64M")
						Eventually(session).Should(Say(`Scaling app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
						Eventually(session).Should(Say(`This will cause the app to restart\. Are you sure you want to scale %s\? \[yN\]:`, appName))
						Eventually(session).Should(Say(`Stopping app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
						Eventually(session).Should(Say(`Starting app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
						Eventually(session).Should(Exit(0))

						updatedAppTable := helpers.ParseV3AppProcessTable(session.Out.Contents())
						Expect(updatedAppTable.Processes).To(HaveLen(3))

						processSummary := updatedAppTable.Processes[0]
						instanceSummary := processSummary.Instances[0]
						Expect(processSummary.Type).To(Equal("web"))
						Expect(processSummary.InstanceCount).To(MatchRegexp(`\d/1`))
						Expect(instanceSummary.Memory).To(MatchRegexp(`\d+(\.\d+)?[KMG]? of 64M`))
						Expect(instanceSummary.Disk).To(MatchRegexp(`\d+(\.\d+)?[KMG]? of \d+[KMG]`))
					})

					When("-f flag provided", func() {
						It("scales without prompt", func() {
							session := helpers.CF("scale", appName, "-m", "64M", "-f")
							Eventually(session).Should(Say("Scaling app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName))
							Eventually(session).Should(Exit(0))

							updatedAppTable := helpers.ParseV3AppProcessTable(session.Out.Contents())
							Expect(updatedAppTable.Processes).To(HaveLen(3))

							processSummary := updatedAppTable.Processes[0]
							instanceSummary := processSummary.Instances[0]
							Expect(processSummary.Type).To(Equal("web"))
							Expect(processSummary.InstanceCount).To(MatchRegexp(`\d/1`))
							Expect(instanceSummary.Memory).To(MatchRegexp(`\d+(\.\d+)?[KMG]? of 64M`))
							Expect(instanceSummary.Disk).To(MatchRegexp(`\d+(\.\d+)?[KMG]? of \d+[KMG]`))
						})
					})
				})

				When("Scaling the disk space", func() {
					It("scales disk to 92M", func() {
						buffer := NewBuffer()
						buffer.Write([]byte("y\n"))
						session := helpers.CFWithStdin(buffer, "scale", appName, "-k", "92M")
						Eventually(session).Should(Say(`Scaling app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
						Eventually(session).Should(Say(`This will cause the app to restart\. Are you sure you want to scale %s\? \[yN\]:`, appName))
						Eventually(session).Should(Say(`Stopping app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
						Eventually(session).Should(Say(`Starting app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
						Eventually(session).Should(Exit(0))

						updatedAppTable := helpers.ParseV3AppProcessTable(session.Out.Contents())
						Expect(updatedAppTable.Processes).To(HaveLen(3))

						processSummary := updatedAppTable.Processes[0]
						instanceSummary := processSummary.Instances[0]
						Expect(processSummary.Type).To(Equal("web"))
						Expect(processSummary.InstanceCount).To(MatchRegexp(`\d/1`))
						Expect(instanceSummary.Memory).To(MatchRegexp(`\d+(\.\d+)?[KMG]? of \d+[KMG]`))
						Expect(instanceSummary.Disk).To(MatchRegexp(`\d+(\.\d+)?[KMG]? of 92M`))
					})

					When("-f flag provided", func() {
						It("scales without prompt", func() {
							session := helpers.CF("scale", appName, "-k", "92M", "-f")
							Eventually(session).Should(Say("Scaling app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName))
							Eventually(session).Should(Exit(0))

							updatedAppTable := helpers.ParseV3AppProcessTable(session.Out.Contents())
							Expect(updatedAppTable.Processes).To(HaveLen(3))

							processSummary := updatedAppTable.Processes[0]
							instanceSummary := processSummary.Instances[0]
							Expect(processSummary.Type).To(Equal("web"))
							Expect(processSummary.InstanceCount).To(MatchRegexp(`\d/1`))
							Expect(instanceSummary.Memory).To(MatchRegexp(`\d+(\.\d+)?[KMG]? of \d+[KMG]`))
							Expect(instanceSummary.Disk).To(MatchRegexp(`\d+(\.\d+)?[KMG]? of 92M`))
						})
					})
				})

				When("Scaling to 0 instances", func() {
					It("scales to 0 instances", func() {
						session := helpers.CF("scale", appName, "-i", "0")
						Eventually(session).Should(Say(`Scaling app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
						Consistently(session).ShouldNot(Say(`This will cause the app to restart|Stopping|Starting`))
						Eventually(session).Should(Exit(0))
						updatedAppTable := helpers.ParseV3AppProcessTable(session.Out.Contents())
						Expect(updatedAppTable.Processes[0].InstanceCount).To(Equal("0/0"))
						Expect(updatedAppTable.Processes[0].Instances).To(BeEmpty())
					})
				})

				When("the user chooses not to restart the app", func() {
					It("cancels the scale", func() {
						buffer := NewBuffer()
						buffer.Write([]byte("n\n"))
						session := helpers.CFWithStdin(buffer, "scale", appName, "-i", "2", "-k", "90M")
						Eventually(session).Should(Say("This will cause the app to restart"))
						Consistently(session).ShouldNot(Say("Stopping"))
						Consistently(session).ShouldNot(Say("Starting"))
						Eventually(session).Should(Say("Scaling cancelled"))
						Consistently(session).ShouldNot(Say(`Waiting for app to start\.\.\.`))
						Eventually(session).Should(Exit(0))

						appTable := helpers.ParseV3AppProcessTable(session.Out.Contents())
						Expect(appTable.Processes).To(BeEmpty())
					})
				})
			})

			When("all scale option flags are provided", func() {
				When("the app starts successfully", func() {
					It("scales the app accordingly", func() {
						buffer := NewBuffer()
						buffer.Write([]byte("y\n"))
						session := helpers.CFWithStdin(buffer, "scale", appName, "-i", "2", "-k", "120M", "-m", "60M")
						Eventually(session).Should(Say(`Scaling app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
						Eventually(session).Should(Say(`This will cause the app to restart\. Are you sure you want to scale %s\? \[yN\]:`, appName))
						Eventually(session).Should(Say(`Stopping app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
						Eventually(session).Should(Say(`Starting app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
						Eventually(session).Should(Exit(0))

						appTable := helpers.ParseV3AppProcessTable(session.Out.Contents())
						Expect(appTable.Processes).To(HaveLen(3))

						processSummary := appTable.Processes[0]
						instanceSummary := processSummary.Instances[0]
						Expect(processSummary.Type).To(Equal("web"))
						Expect(processSummary.InstanceCount).To(MatchRegexp(`\d/2`))
						Expect(instanceSummary.State).To(MatchRegexp(`running|starting`))
						Expect(instanceSummary.Memory).To(MatchRegexp(`\d+(\.\d+)?[KMG]? of 60M`))
						Expect(instanceSummary.Disk).To(MatchRegexp(`\d+(\.\d+)?[KMG]? of 120M`))
					})
				})

				When("the app does not start successfully", func() {
					It("scales the app and displays the app summary", func() {
						buffer := NewBuffer()
						buffer.Write([]byte("y\n"))
						session := helpers.CFWithStdin(buffer, "scale", appName, "-i", "2", "-k", "120M", "-m", "6M")
						Eventually(session).Should(Say(`Scaling app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
						Eventually(session).Should(Say(`This will cause the app to restart\. Are you sure you want to scale %s\? \[yN\]:`, appName))
						Eventually(session).Should(Say(`Stopping app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
						Eventually(session).Should(Say(`Starting app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
						Eventually(session).Should(Exit(0))

						appTable := helpers.ParseV3AppProcessTable(session.Out.Contents())
						Expect(appTable.Processes).To(HaveLen(3))

						processSummary := appTable.Processes[0]
						instanceSummary := processSummary.Instances[0]
						Expect(processSummary.Type).To(Equal("web"))
						Expect(processSummary.InstanceCount).To(MatchRegexp(`\d/2`))
						Expect(instanceSummary.State).To(MatchRegexp(`crashed`))
						Expect(instanceSummary.Memory).To(MatchRegexp(`\d+(\.\d+)?[KMG]? of 6M`))
						Expect(instanceSummary.Disk).To(MatchRegexp(`\d+(\.\d+)?[KMG]? of 120M`))
					})
				})
			})

			PWhen("the provided scale options are the same as the existing scale properties", func() {
				var (
					session          *Session
					currentInstances string
					maxMemory        string
					maxDiskSize      string
				)

				BeforeEach(func() {
					session = helpers.CF("scale", appName)
					Eventually(session).Should(Exit(0))

					appTable := helpers.ParseV3AppProcessTable(session.Out.Contents())
					instanceSummary := appTable.Processes[0].Instances[0]
					currentInstances = string(len(appTable.Processes[0].Instances))
					maxMemory = strings.Fields(instanceSummary.Memory)[2]
					maxDiskSize = strings.Fields(instanceSummary.Disk)[2]
				})

				It("the action should be a no-op", func() {
					session = helpers.CF("scale", appName, "-i", currentInstances, "-m", maxMemory, "-k", maxDiskSize)
					Eventually(session).Should(Say(`Scaling app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
					Consistently(session).ShouldNot(Say("This will cause the app to restart"))
					Consistently(session).ShouldNot(Say("Stopping"))
					Consistently(session).ShouldNot(Say("Starting"))
					Consistently(session).ShouldNot(Say("Waiting for app to start"))
					Eventually(session).Should(Exit(0))

					appTable := helpers.ParseV3AppProcessTable(session.Out.Contents())
					Expect(appTable.Processes).To(HaveLen(1))

					newProcessSummary := appTable.Processes[0]
					newInstanceSummary := newProcessSummary.Instances[0]
					Expect(newProcessSummary.Type).To(Equal("web"))
					Expect(newProcessSummary.InstanceCount).To(MatchRegexp(fmt.Sprintf(`\d/%s`, currentInstances)))
					Expect(newInstanceSummary.Memory).To(MatchRegexp(fmt.Sprintf(`\d+(\.\d+)?[KMG]? of %s`, maxMemory)))
					Expect(newInstanceSummary.Disk).To(MatchRegexp(fmt.Sprintf(`\d+(\.\d+)?[KMG]? of %s`, maxDiskSize)))
				})
			})

			When("the process flag is provided", func() {
				It("scales the requested process", func() {
					session := helpers.CF("scale", appName, "-i", "2", "--process", "console")
					Eventually(session).Should(Say(`Scaling app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
					Eventually(session).Should(Exit(0))

					appTable := helpers.ParseV3AppProcessTable(session.Out.Contents())
					Expect(appTable.Processes).To(HaveLen(3))

					processSummary := appTable.Processes[1]
					instanceSummary := processSummary.Instances[0]
					Expect(processSummary.Instances).To(HaveLen(2))
					Expect(processSummary.Type).To(Equal("console"))
					Expect(processSummary.InstanceCount).To(MatchRegexp(`\d/2`))
					Expect(instanceSummary.Memory).To(MatchRegexp(`\d+(\.\d+)?[KMG]? of \d+[KMG]`))
					Expect(instanceSummary.Disk).To(MatchRegexp(`\d+(\.\d+)?[KMG]? of \d+[KMG]`))
				})
			})
		})
	})

	When("invalid scale option values are provided", func() {
		When("a negative value is passed to a flag argument", func() {
			It("outputs an error message to the user, provides help text, and exits 1", func() {
				session := helpers.CF("scale", "some-app", "-i=-5")
				Eventually(session.Err).Should(Say(`Incorrect Usage: invalid argument for flag '-i' \(expected int > 0\)`))
				Eventually(session).Should(Say("cf scale APP_NAME")) // help
				Eventually(session).Should(Exit(1))

				session = helpers.CF("scale", "some-app", "-k=-5")
				Eventually(session.Err).Should(Say("Byte quantity must be an integer with a unit of measurement like M, MB, G, or GB"))
				Eventually(session).Should(Say("cf scale APP_NAME")) // help
				Eventually(session).Should(Exit(1))

				session = helpers.CF("scale", "some-app", "-m=-5")
				Eventually(session.Err).Should(Say("Byte quantity must be an integer with a unit of measurement like M, MB, G, or GB"))
				Eventually(session).Should(Say("cf scale APP_NAME")) // help
				Eventually(session).Should(Exit(1))
			})
		})

		When("a non-integer value is passed to a flag argument", func() {
			It("outputs an error message to the user, provides help text, and exits 1", func() {
				session := helpers.CF("scale", "some-app", "-i", "not-an-integer")
				Eventually(session.Err).Should(Say(`Incorrect Usage: invalid argument for flag '-i' \(expected int > 0\)`))
				Eventually(session).Should(Say("cf scale APP_NAME")) // help
				Eventually(session).Should(Exit(1))

				session = helpers.CF("scale", "some-app", "-k", "not-an-integer")
				Eventually(session.Err).Should(Say("Byte quantity must be an integer with a unit of measurement like M, MB, G, or GB"))
				Eventually(session).Should(Say("cf scale APP_NAME")) // help
				Eventually(session).Should(Exit(1))

				session = helpers.CF("scale", "some-app", "-m", "not-an-integer")
				Eventually(session.Err).Should(Say("Byte quantity must be an integer with a unit of measurement like M, MB, G, or GB"))
				Eventually(session).Should(Say("cf scale APP_NAME")) // help
				Eventually(session).Should(Exit(1))
			})
		})

		When("the unit of measurement is not provided", func() {
			It("outputs an error message to the user, provides help text, and exits 1", func() {
				session := helpers.CF("scale", "some-app", "-k", "9")
				Eventually(session.Err).Should(Say("Byte quantity must be an integer with a unit of measurement like M, MB, G, or GB"))
				Eventually(session).Should(Say("cf scale APP_NAME")) // help
				Eventually(session).Should(Exit(1))

				session = helpers.CF("scale", "some-app", "-m", "7")
				Eventually(session.Err).Should(Say("Byte quantity must be an integer with a unit of measurement like M, MB, G, or GB"))
				Eventually(session).Should(Say("cf scale APP_NAME")) // help
				Eventually(session).Should(Exit(1))
			})
		})
	})
})
