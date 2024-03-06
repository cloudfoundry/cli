package isolated

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("restart command", func() {
	Describe("help", func() {
		When("--help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("restart", "--help")

				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("restart - Stop all instances of the app, then start them again. This causes downtime."))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf restart APP_NAME"))
				Eventually(session).Should(Say("ALIAS:"))
				Eventually(session).Should(Say("rs"))
				Eventually(session).Should(Say("ENVIRONMENT:"))
				Eventually(session).Should(Say(`CF_STAGING_TIMEOUT=15\s+Max wait time for buildpack staging, in minutes`))
				Eventually(session).Should(Say(`CF_STARTUP_TIMEOUT=5\s+Max wait time for app instance startup, in minutes`))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("restage, restart-app-instance"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "restart", "app-name")
		})
	})

	When("the environment is set up correctly", func() {
		var (
			orgName   string
			spaceName string
		)

		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()

			helpers.SetupCF(orgName, spaceName)
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		When("the app does not exist", func() {
			It("tells the user that the start is not found and exits 1", func() {
				appName := helpers.PrefixedRandomName("app")
				session := helpers.CF("restart", appName)

				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("App '%s' not found", appName))
				Eventually(session).Should(Exit(1))
			})
		})

		When("the app does exist", func() {
			var (
				domainName string
				appName    string
			)

			BeforeEach(func() {
				appName = helpers.PrefixedRandomName("app")
				domainName = helpers.DefaultSharedDomain()
			})

			When("the app is started", func() {
				BeforeEach(func() {
					helpers.WithHelloWorldApp(func(appDir string) {
						Eventually(helpers.CF("push", appName, "-p", appDir, "-b", "staticfile_buildpack")).Should(Exit(0))
					})
				})

				It("stops the app and starts it again", func() {
					userName, _ := helpers.GetCredentials()
					session := helpers.CF("restart", appName)
					Eventually(session).Should(Say(`Restarting app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
					Eventually(session).Should(Say(`Stopping app\.\.\.`))
					Consistently(session).ShouldNot(Say(`Staging app and tracing logs\.\.\.`))
					Eventually(session).Should(Say(`Waiting for app to start\.\.\.`))
					Eventually(session).Should(Exit(0))
				})
			})

			When("the app is stopped", func() {
				When("the app has been staged", func() {
					BeforeEach(func() {
						helpers.WithHelloWorldApp(func(appDir string) {
							manifestContents := []byte(fmt.Sprintf(`
---
applications:
- name: %s
  memory: 128M
  instances: 2
  disk_quota: 128M
  routes:
  - route: %s.%s
`, appName, appName, domainName))
							manifestPath := filepath.Join(appDir, "manifest.yml")
							err := ioutil.WriteFile(manifestPath, manifestContents, 0666)
							Expect(err).ToNot(HaveOccurred())

							Eventually(helpers.CF("push", appName, "-p", appDir, "-f", manifestPath, "-b", "staticfile_buildpack")).Should(Exit(0))
						})
						Eventually(helpers.CF("stop", appName)).Should(Exit(0))
					})

					Describe("multiprocess display", func() {
						It("uses the multiprocess display", func() {
							userName, _ := helpers.GetCredentials()

							session := helpers.CF("start", appName)

							Eventually(session).Should(Say(`Starting app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))

							Eventually(session).Should(Say(`name:\s+%s`, appName))
							Eventually(session).Should(Say(`requested state:\s+started`))
							Eventually(session).Should(Say(`routes:\s+%s\.%s`, appName, domainName))
							Eventually(session).Should(Say(`last uploaded:\s+%s`, helpers.ReadableDateTimeRegex))
							Eventually(session).Should(Say(`stack:\s+cflinuxfs`))
							Eventually(session).Should(Say(`buildpacks:\s+staticfile`))
							Eventually(session).Should(Say(`type:\s+web`))
							Eventually(session).Should(Say(`instances:\s+\d/2`))
							Eventually(session).Should(Say(`memory usage:\s+128M`))
							Eventually(session).Should(Say(`\s+state\s+since\s+cpu\s+memory\s+disk`))
							Eventually(session).Should(Say(`#0\s+(starting|running)\s+\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z`))

							Eventually(session).Should(Exit(0))
						})
					})
				})

				When("the app does *not* stage properly because the app was not detected by any buildpacks", func() {
					BeforeEach(func() {
						helpers.WithHelloWorldApp(func(appDir string) {
							err := os.Remove(filepath.Join(appDir, "Staticfile"))
							Expect(err).ToNot(HaveOccurred())
							Eventually(helpers.CF("push", appName, "-p", appDir, "--no-start")).Should(Exit(0))
						})
					})

					It("fails and displays the staging failure message", func() {
						userName, _ := helpers.GetCredentials()
						session := helpers.CF("restart", appName)
						Eventually(session).Should(Say(`Restarting app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
						Eventually(session).Should(Say(`Staging app and tracing logs\.\.\.`))

						// The staticfile_buildback does compile an index.html file. However, it requires a "Staticfile" during buildpack detection.
						Eventually(session.Err).Should(Say("Error staging application: An app was not successfully detected by any available buildpack"))
						Eventually(session.Err).Should(Say(`TIP: Use 'cf buildpacks' to see a list of supported buildpacks.`))
						Eventually(session).Should(Exit(1))
					})
				})

				When("the app stages properly", func() {
					When("the app does *not* start properly", func() {
						BeforeEach(func() {
							appName = helpers.PrefixedRandomName("app")
							helpers.WithHelloWorldApp(func(appDir string) {
								Eventually(helpers.CF("push", appName, "-p", appDir, "--no-start", "-b", "staticfile_buildpack", "-c", "gibberish")).Should(Exit(0))
							})
						})

						It("fails and displays the start failure message", func() {
							userName, _ := helpers.GetCredentials()
							session := helpers.CF("restart", appName)
							Eventually(session).Should(Say(`Restarting app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))

							Eventually(session.Err).Should(Say("Start unsuccessful"))
							Eventually(session.Err).Should(Say("TIP: use 'cf logs .* --recent' for more information"))
							Eventually(session).Should(Exit(1))
						})
					})

					When("the app starts properly", func() {
						BeforeEach(func() {
							helpers.WithHelloWorldApp(func(appDir string) {
								manifestContents := []byte(fmt.Sprintf(`
---
applications:
- name: %s
  memory: 128M
  instances: 2
  disk_quota: 128M
  routes:
  - route: %s.%s
`, appName, appName, domainName))
								manifestPath := filepath.Join(appDir, "manifest.yml")
								err := ioutil.WriteFile(manifestPath, manifestContents, 0666)
								Expect(err).ToNot(HaveOccurred())

								Eventually(helpers.CF("push", appName, "-p", appDir, "-f", manifestPath, "-b", "staticfile_buildpack", "--no-start")).Should(Exit(0))
							})
							Eventually(helpers.CF("stop", appName)).Should(Exit(0))
						})

						It("displays the app logs and information with instances table", func() {
							userName, _ := helpers.GetCredentials()
							session := helpers.CF("restart", appName)
							Eventually(session).Should(Say(`Restarting app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
							Consistently(session).ShouldNot(Say(`Stopping app\.\.\.`))

							helpers.ConfirmStagingLogs(session)

							Eventually(session).Should(Say(`name:\s+%s`, appName))
							Eventually(session).Should(Say(`memory usage:\s+128M`))
							Eventually(session).Should(Exit(0))
						})
					})

					When("isolation segments are available", func() {
						BeforeEach(func() {
							Eventually(helpers.CF("create-isolation-segment", RealIsolationSegment)).Should(Exit(0))
							Eventually(helpers.CF("enable-org-isolation", orgName, RealIsolationSegment)).Should(Exit(0))
							Eventually(helpers.CF("set-space-isolation-segment", spaceName, RealIsolationSegment)).Should(Exit(0))

							helpers.WithHelloWorldApp(func(appDir string) {
								Eventually(helpers.CF("push", appName, "-p", appDir, "--no-start")).Should(Exit(0))
							})
						})

						It("displays the isolation segment information", func() {
							session := helpers.CF("restart", appName)

							Eventually(session).Should(Say(`isolation segment:\s+%s`, RealIsolationSegment))
							Eventually(session).Should(Exit(0))
						})
					})
				})
			})
		})
	})
})
