package isolated

import (
	"fmt"
	"strings"
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
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
			It("appears in cf help -a", func() {
				session := helpers.CF("help", "-a")
				Eventually(session).Should(Exit(0))
				Expect(session).To(HaveCommandInCategoryWithDescription("scale", "APPS", "Change or view the instance count, disk space limit, memory limit, and log rate limit for an app"))
			})

			It("displays command usage to output", func() {
				session := helpers.CF("scale", "--help")

				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("scale - Change or view the instance count, disk space limit, memory limit, and log rate limit for an app"))

				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say(`cf scale APP_NAME \[--process PROCESS\] \[-i INSTANCES\] \[-k DISK\] \[-m MEMORY\] \[-l LOG_RATE_LIMIT\]`))
				Eventually(session).Should(Say("Modifying the app's disk, memory, or log rate will cause the app to restart."))

				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say(`-f\s+Force restart of app without prompt`))
				Eventually(session).Should(Say(`-i\s+Number of instances`))
				Eventually(session).Should(Say(`-k\s+Disk limit \(e\.g\. 256M, 1024M, 1G\)`))
				Eventually(session).Should(Say(`-l\s+Log rate limit per second, in bytes \(e\.g\. 128B, 4K, 1M\). -l=-1 represents unlimited`))
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
				Eventually(session).Should(Exit(1))

				Expect(session.Err).To(Say("Incorrect Usage: the required argument `APP_NAME` was not provided"))
				Expect(session).To(Say("NAME:"))
			})
		})

		When("the app does not exist", func() {
			It("displays app not found and exits 1", func() {
				invalidAppName := "invalid-app-name"
				session := helpers.CF("scale", invalidAppName)
				Eventually(session).Should(Exit(1))

				Expect(session.Err).To(Say("App '%s' not found", invalidAppName))
				Expect(session).To(Say("FAILED"))
			})
		})

		When("the app exists", func() {
			BeforeEach(func() {
				helpers.WithProcfileApp(func(appDir string) {
					Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "push", appName)).Should(Exit(0))
				})
				helpers.WaitForAppMemoryToTakeEffect(appName, 0, 0, false, "1G")
			})

			When("scale option flags are not provided", func() {
				It("displays the current scale properties for all processes", func() {
					session := helpers.CF("scale", appName)

					Eventually(session).Should(Exit(0))

					Expect(session).To(Say(`Showing current scale of app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
					Expect(session).To(Say(`name:\s+%s`, appName))
					Expect(session).To(Say(`requested state:\s+started`))

					Consistently(session).ShouldNot(Say("Scaling"))
					Consistently(session).ShouldNot(Say("This will cause the app to restart"))
					Consistently(session).ShouldNot(Say("Stopping"))
					Consistently(session).ShouldNot(Say("Starting"))
					Consistently(session).ShouldNot(Say("Waiting"))

					appTable := helpers.ParseV3AppProcessTable(session.Out.Contents())
					Expect(len(appTable.Processes)).To(Equal(2))

					processSummary := appTable.Processes[0]
					Expect(processSummary.Type).To(Equal("web"))
					Expect(processSummary.InstanceCount).To(Equal("1/1"))

					instanceSummary := processSummary.Instances[0]
					Expect(instanceSummary.Memory).To(MatchRegexp(`\d+(\.\d+)?[KMG]? of \d+?[KMG]?`))
					Expect(instanceSummary.Disk).To(MatchRegexp(`\d+(\.\d+)?[KMG]? of \d+?[KMG]?`))

					Expect(appTable.Processes[1].Type).To(Equal("console"))
					Expect(appTable.Processes[1].InstanceCount).To(Equal("0/0"))
				})
			})

			When("only one scale option flag is provided", func() {
				When("scaling the number of instances", func() {
					It("Scales to the correct number of instances", func() {
						By("Verifying we start with one instance")
						session := helpers.CF("scale", appName)
						Eventually(session).Should(Exit(0))

						appTable := helpers.ParseV3AppProcessTable(session.Out.Contents())
						Expect(appTable.Processes).To(HaveLen(2))

						processSummary := appTable.Processes[0]
						Expect(processSummary.Type).To(Equal("web"))
						Expect(processSummary.InstanceCount).To(Equal("1/1"))

						By("then scaling to 3 instances")
						session = helpers.CF("scale", appName, "-i", "3")
						Eventually(session).Should(Exit(0))
						Expect(session).To(Say(`Scaling app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
						Consistently(session).ShouldNot(Say("This will cause the app to restart"))
						Consistently(session).ShouldNot(Say("Stopping"))
						Consistently(session).ShouldNot(Say("Starting"))

						helpers.WaitForAppMemoryToTakeEffect(appName, 0, 0, true, "1G")

						session = helpers.CF("app", appName)
						Eventually(session).Should(Exit(0))

						updatedAppTable := helpers.ParseV3AppProcessTable(session.Out.Contents())
						Expect(updatedAppTable.Processes).To(HaveLen(2))

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
						_, err := buffer.Write([]byte("y\n"))
						Expect(err).ToNot(HaveOccurred())
						session := helpers.CFWithStdin(buffer, "scale", appName, "-m", "64M")
						Eventually(session).Should(Exit(0))
						Expect(session).To(Say(`Scaling app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
						Expect(session).To(Say(`This will cause the app to restart\. Are you sure you want to scale %s\? \[yN\]:`, appName))
						Expect(session).To(Say(`Stopping app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
						Expect(session).To(Say(`Starting app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))

						helpers.WaitForAppMemoryToTakeEffect(appName, 0, 0, false, "64M")
					})

					When("-f flag provided", func() {
						It("scales without prompt", func() {
							session := helpers.CF("scale", appName, "-m", "64M", "-f")
							Eventually(session).Should(Exit(0))
							Expect(session).To(Say("Scaling app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName))

							helpers.WaitForAppMemoryToTakeEffect(appName, 0, 0, false, "64M")
						})
					})
				})

				When("Scaling the disk space", func() {
					It("scales disk to 512M", func() {
						buffer := NewBuffer()
						_, err := buffer.Write([]byte("y\n"))
						Expect(err).ToNot(HaveOccurred())
						session := helpers.CFWithStdin(buffer, "scale", appName, "-k", "512M")
						Eventually(session).Should(Exit(0))
						Expect(session).To(Say(`Scaling app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
						Expect(session).To(Say(`This will cause the app to restart\. Are you sure you want to scale %s\? \[yN\]:`, appName))
						Expect(session).To(Say(`Stopping app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
						Expect(session).To(Say(`Starting app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
						Expect(session).To(Say(`Instances starting\.\.\.`))

						helpers.WaitForAppDiskToTakeEffect(appName, 0, 0, false, "512M")
					})

					When("-f flag provided", func() {
						It("scales without prompt", func() {
							session := helpers.CF("scale", appName, "-k", "512M", "-f")
							Eventually(session).Should(Exit(0))
							Expect(session).To(Say("Scaling app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName))

							helpers.WaitForAppDiskToTakeEffect(appName, 0, 0, false, "512M")
						})
					})
				})

				When("Scaling the log rate limit", func() {
					BeforeEach(func() {
						helpers.SkipIfVersionLessThan(ccversion.MinVersionLogRateLimitingV3)
					})

					It("scales log rate limit to 1M", func() {
						buffer := NewBuffer()
						_, err := buffer.Write([]byte("y\n"))
						Expect(err).ToNot(HaveOccurred())
						session := helpers.CFWithStdin(buffer, "scale", appName, "-l", "1M")
						Eventually(session).Should(Exit(0))
						Expect(session).To(Say(`Scaling app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
						Expect(session).To(Say(`This will cause the app to restart\. Are you sure you want to scale %s\? \[yN\]:`, appName))
						Expect(session).To(Say(`Stopping app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
						Expect(session).To(Say(`Starting app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
						Expect(session).To(Say(`Instances starting\.\.\.`))

						helpers.WaitForLogRateLimitToTakeEffect(appName, 0, 0, false, "1M")
					})

					When("-f flag provided", func() {
						It("scales without prompt", func() {
							session := helpers.CF("scale", appName, "-l", "1M", "-f")
							Eventually(session).Should(Exit(0))
							Expect(session).To(Say("Scaling app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName))

							helpers.WaitForLogRateLimitToTakeEffect(appName, 0, 0, false, "1M")
						})
					})
				})

				When("Scaling to 0 instances", func() {
					It("scales to 0 instances", func() {
						session := helpers.CF("scale", appName, "-i", "0")
						Eventually(session).Should(Exit(0))
						Expect(session).To(Say(`Scaling app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
						Consistently(session).ShouldNot(Say(`This will cause the app to restart|Stopping|Starting`))
						updatedAppTable := helpers.ParseV3AppProcessTable(session.Out.Contents())
						Expect(updatedAppTable.Processes[0].InstanceCount).To(Equal("0/0"))
						Expect(updatedAppTable.Processes[0].Instances).To(BeEmpty())
					})
				})

				When("the user chooses not to restart the app", func() {
					It("cancels the scale", func() {
						buffer := NewBuffer()
						_, err := buffer.Write([]byte("n\n"))
						session := helpers.CFWithStdin(buffer, "scale", appName, "-i", "2", "-k", "90M")
						Eventually(session).Should(Exit(0))
						Expect(err).ToNot(HaveOccurred())
						Expect(session).To(Say("This will cause the app to restart"))
						Expect(session).To(Say("Scaling cancelled"))

						Consistently(session).ShouldNot(Say("Stopping"))
						Consistently(session).ShouldNot(Say("Starting"))
						Consistently(session).ShouldNot(Say(`Waiting for app to start\.\.\.`))

						appTable := helpers.ParseV3AppProcessTable(session.Out.Contents())
						Expect(appTable.Processes).To(BeEmpty())
					})
				})
			})

			When("all scale option flags are provided", func() {
				When("the app starts successfully", func() {
					It("scales the app accordingly", func() {
						buffer := NewBuffer()
						_, err := buffer.Write([]byte("y\n"))
						Expect(err).ToNot(HaveOccurred())

						//Delay to reduce flakiness
						time.Sleep(3 * time.Second)

						session := helpers.CFWithStdin(buffer, "scale", appName, "-i", "2", "-k", "512M", "-m", "60M")
						Eventually(session).Should(Exit(0))
						Expect(session).To(Say(`Scaling app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
						Expect(session).To(Say(`This will cause the app to restart\. Are you sure you want to scale %s\? \[yN\]:`, appName))
						Expect(session).To(Say(`Stopping app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
						Expect(session).To(Say(`Starting app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))

						helpers.WaitForAppMemoryToTakeEffect(appName, 0, 0, false, "60M")

						//Delay to reduce flakiness
						time.Sleep(5 * time.Second)

						session = helpers.CF("app", appName)
						Eventually(session).Should(Exit(0))

						appTable := helpers.ParseV3AppProcessTable(session.Out.Contents())
						Expect(appTable.Processes).To(HaveLen(2))

						processSummary := appTable.Processes[0]
						instanceSummary := processSummary.Instances[0]
						Expect(processSummary.Type).To(Equal("web"))
						Expect(processSummary.InstanceCount).To(MatchRegexp(`\d/2`))
						Expect(instanceSummary.State).To(MatchRegexp(`running|starting`))
						Expect(instanceSummary.Memory).To(MatchRegexp(`\d+(\.\d+)?[KMG]? of 60M`))
						Expect(instanceSummary.Disk).To(MatchRegexp(`\d+(\.\d+)?[KMG]? of 512M`))
					})
				})

				When("the app does not start successfully", func() {
					It("scales the app and displays the app summary", func() {
						buffer := NewBuffer()
						_, err := buffer.Write([]byte("y\n"))
						Expect(err).ToNot(HaveOccurred())
						session := helpers.CFWithStdin(buffer, "scale", appName, "-i", "2", "-k", "10M", "-m", "6M")
						Eventually(session).Should(Exit(1))
						Expect(session).To(Say(`Scaling app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
						Expect(session).To(Say(`This will cause the app to restart\. Are you sure you want to scale %s\? \[yN\]:`, appName))
						Expect(session).To(Say(`Stopping app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
						Expect(session).To(Say(`Starting app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))

						helpers.WaitForAppMemoryToTakeEffect(appName, 0, 0, false, "6M")

						session = helpers.CF("restart", appName)
						Eventually(session).Should(Exit(1))

						session = helpers.CF("app", appName)
						Eventually(session).Should(Exit(0))

						appTable := helpers.ParseV3AppProcessTable(session.Out.Contents())
						Expect(appTable.Processes).To(HaveLen(2))

						processSummary := appTable.Processes[0]
						instanceSummary := processSummary.Instances[0]
						Expect(processSummary.Type).To(Equal("web"))
						Expect(processSummary.InstanceCount).To(MatchRegexp(`\d/2`))
						Expect(instanceSummary.State).To(MatchRegexp(`crashed`))
						Expect(instanceSummary.Memory).To(MatchRegexp(`\d+(\.\d+)?[KMG]? of 6M`))
						Expect(instanceSummary.Disk).To(MatchRegexp(`\d+(\.\d+)?[KMG]? of 10M`))
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
					currentInstances = fmt.Sprint(len(appTable.Processes[0].Instances))
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
					Eventually(session).Should(Exit(0))
					Expect(session).To(Say(`Scaling app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))

					session = helpers.CF("app", appName)
					Eventually(session).Should(Exit(0))

					appTable := helpers.ParseV3AppProcessTable(session.Out.Contents())
					Expect(appTable.Processes).To(HaveLen(2))

					processSummary := appTable.Processes[1]
					Expect(processSummary.Instances).To(HaveLen(2))
					Expect(processSummary.Type).To(Equal("console"))
					Expect(processSummary.InstanceCount).To(MatchRegexp(`\d/2`))
				})
			})
		})
	})

	When("invalid scale option values are provided", func() {
		When("a negative value is passed to a flag argument", func() {
			It("outputs an error message to the user, provides help text, and exits 1", func() {
				session := helpers.CF("scale", "some-app", "-i=-5")
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say(`Incorrect Usage: invalid argument for flag '-i' \(expected int > 0\)`))
				Expect(session).To(Say("cf scale APP_NAME"))
				session = helpers.CF("scale", "some-app", "-k=-5")
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Byte quantity must be an integer with a unit of measurement like M, MB, G, or GB"))
				Expect(session).To(Say("cf scale APP_NAME"))
				session = helpers.CF("scale", "some-app", "-m=-5")
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Byte quantity must be an integer with a unit of measurement like M, MB, G, or GB"))
				Expect(session).To(Say("cf scale APP_NAME"))
			})
		})

		When("a non-integer value is passed to a flag argument", func() {
			It("outputs an error message to the user, provides help text, and exits 1", func() {
				session := helpers.CF("scale", "some-app", "-i", "not-an-integer")
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say(`Incorrect Usage: invalid argument for flag '-i' \(expected int > 0\)`))
				Expect(session).To(Say("cf scale APP_NAME"))

				session = helpers.CF("scale", "some-app", "-k", "not-an-integer")
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Byte quantity must be an integer with a unit of measurement like M, MB, G, or GB"))
				Expect(session).To(Say("cf scale APP_NAME"))

				session = helpers.CF("scale", "some-app", "-m", "not-an-integer")
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Byte quantity must be an integer with a unit of measurement like M, MB, G, or GB"))
				Expect(session).To(Say("cf scale APP_NAME"))
			})
		})

		When("the unit of measurement is not provided", func() {
			It("outputs an error message to the user, provides help text, and exits 1", func() {
				session := helpers.CF("scale", "some-app", "-k", "9")
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Byte quantity must be an integer with a unit of measurement like M, MB, G, or GB"))
				Expect(session).To(Say("cf scale APP_NAME"))

				session = helpers.CF("scale", "some-app", "-m", "7")
				Eventually(session).Should(Exit(1))
				Expect(session.Err).To(Say("Byte quantity must be an integer with a unit of measurement like M, MB, G, or GB"))
				Expect(session).To(Say("cf scale APP_NAME"))
			})
		})
	})
})
