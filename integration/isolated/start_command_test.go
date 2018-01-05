package isolated

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("start command", func() {
	Describe("help", func() {
		Context("when --help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("start", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("start - Start an app"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf start APP_NAME"))
				Eventually(session).Should(Say("ALIAS:"))
				Eventually(session).Should(Say("st"))
				Eventually(session).Should(Say("ENVIRONMENT:"))
				Eventually(session).Should(Say("CF_STAGING_TIMEOUT=15\\s+Max wait time for buildpack staging, in minutes"))
				Eventually(session).Should(Say("CF_STARTUP_TIMEOUT=5\\s+Max wait time for app instance startup, in minutes"))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("apps, logs, restart, run-task, scale, ssh, stop"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("when the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "start", "app-name")
		})
	})

	Context("when the environment is set up correctly", func() {
		var (
			orgName   string
			spaceName string
		)

		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()

			setupCF(orgName, spaceName)
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		Context("when the app does not exist", func() {
			It("tells the user that the start is not found and exits 1", func() {
				appName := helpers.PrefixedRandomName("app")
				session := helpers.CF("start", appName)

				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("App %s not found", appName))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the app does exist", func() {
			var (
				domainName string
				appName    string
			)

			Context("when the app is started", func() {
				BeforeEach(func() {
					appName = helpers.PrefixedRandomName("app")
					domainName = defaultSharedDomain()
					helpers.WithHelloWorldApp(func(appDir string) {
						Eventually(helpers.CF("push", appName, "-p", appDir, "-b", "staticfile_buildpack")).Should(Exit(0))
					})
				})

				It("only displays the app already started message", func() {
					userName, _ := helpers.GetCredentials()
					session := helpers.CF("start", appName)
					Eventually(session).Should(Say("Starting app %s in org %s / space %s as %s...", appName, orgName, spaceName, userName))
					Eventually(session).Should(Say("App %s is already started", appName))
					Eventually(session).Should(Exit(0))
				})
			})

			Context("when the app is stopped", func() {
				Context("when the app has been staged", func() {
					BeforeEach(func() {
						appName = helpers.PrefixedRandomName("app")
						domainName = defaultSharedDomain()
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

					It("displays the app information with instances table", func() {
						userName, _ := helpers.GetCredentials()
						session := helpers.CF("start", appName)
						Eventually(session).Should(Say("Starting app %s in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, userName))
						Consistently(session).ShouldNot(Say("Staging app and tracing logs\\.\\.\\."))

						Eventually(session).Should(Say("Waiting for app to start\\.\\.\\."))

						Eventually(session).Should(Say("name:\\s+%s", appName))
						Eventually(session).Should(Say("requested state:\\s+started"))
						Eventually(session).Should(Say("instances:\\s+2/2"))
						Eventually(session).Should(Say("usage:\\s+128M x 2 instances"))
						Eventually(session).Should(Say("routes:\\s+%s.%s", appName, domainName))
						Eventually(session).Should(Say("last uploaded:"))
						Eventually(session).Should(Say("stack:\\s+cflinuxfs2"))
						Eventually(session).Should(Say("buildpack:\\s+staticfile_buildpack"))
						Eventually(session).Should(Say("start command:"))

						Eventually(session).Should(Say("state\\s+since\\s+cpu\\s+memory\\s+disk\\s+details"))
						Eventually(session).Should(Say("#0\\s+(running|starting)\\s+.*\\d+\\.\\d+%.*of 128M.*of 128M"))
						Eventually(session).Should(Say("#1\\s+(running|starting)\\s+.*\\d+\\.\\d+%.*of 128M.*of 128M"))
						Eventually(session).Should(Exit(0))
					})
				})

				Context("when the app has *not* yet been staged", func() {
					Context("when the app does *not* stage properly because the app was not detected by any buildpacks", func() {
						BeforeEach(func() {
							appName = helpers.PrefixedRandomName("app")
							domainName = defaultSharedDomain()
							helpers.WithHelloWorldApp(func(appDir string) {
								err := os.Remove(filepath.Join(appDir, "Staticfile"))
								Expect(err).ToNot(HaveOccurred())
								Eventually(helpers.CF("push", appName, "-p", appDir, "--no-start")).Should(Exit(0))
							})
						})

						It("fails and displays the staging failure message", func() {
							userName, _ := helpers.GetCredentials()
							session := helpers.CF("start", appName)
							Eventually(session).Should(Say("Starting app %s in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, userName))

							// The staticfile_buildback does compile an index.html file. However, it requires a "Staticfile" during buildpack detection.
							Eventually(session.Err).Should(Say("Error staging application: An app was not successfully detected by any available buildpack"))
							Eventually(session.Err).Should(Say(`TIP: Use 'cf buildpacks' to see a list of supported buildpacks.`))
							Eventually(session).Should(Exit(1))
						})
					})

					Context("when the app stages properly", func() {
						Context("when the app does *not* start properly", func() {
							BeforeEach(func() {
								appName = helpers.PrefixedRandomName("app")
								helpers.WithHelloWorldApp(func(appDir string) {
									Eventually(helpers.CF("push", appName, "-p", appDir, "--no-start", "-b", "staticfile_buildpack", "-c", "gibberish")).Should(Exit(0))
								})
							})

							It("fails and displays the start failure message", func() {
								userName, _ := helpers.GetCredentials()
								session := helpers.CF("start", appName)
								Eventually(session).Should(Say("Starting app %s in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, userName))

								Eventually(session).Should(Say("Staging app and tracing logs\\.\\.\\."))

								Eventually(session.Err).Should(Say("Start unsuccessful"))
								Eventually(session.Err).Should(Say("TIP: use 'cf logs .* --recent' for more information"))
								Eventually(session).Should(Exit(1))
							})
						})

						Context("when the app starts properly", func() {
							BeforeEach(func() {
								Eventually(helpers.CF("create-isolation-segment", RealIsolationSegment)).Should(Exit(0))
								Eventually(helpers.CF("enable-org-isolation", orgName, RealIsolationSegment)).Should(Exit(0))
								Eventually(helpers.CF("set-space-isolation-segment", spaceName, RealIsolationSegment)).Should(Exit(0))
								appName = helpers.PrefixedRandomName("app")
								domainName = defaultSharedDomain()
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
								session := helpers.CF("start", appName)
								Eventually(session).Should(Say("Starting app %s in org %s / space %s as %s\\.\\.\\.", appName, orgName, spaceName, userName))

								// Display Staging Logs
								Eventually(session).Should(Say("Uploading droplet\\.\\.\\."))
								Eventually(session).Should(Say("Waiting for app to start\\.\\.\\."))

								Eventually(session).Should(Say("name:\\s+%s", appName))
								Eventually(session).Should(Say("requested state:\\s+started"))
								Eventually(session).Should(Say("instances:\\s+2/2"))
								Eventually(session).Should(Say("isolation segment:\\s+%s", RealIsolationSegment))
								Eventually(session).Should(Say("usage:\\s+128M x 2 instances"))
								Eventually(session).Should(Say("routes:\\s+%s.%s", appName, domainName))
								Eventually(session).Should(Say("last uploaded:"))
								Eventually(session).Should(Say("stack:\\s+cflinuxfs2"))
								Eventually(session).Should(Say("buildpack:\\s+staticfile_buildpack"))
								Eventually(session).Should(Say("start command:"))

								Eventually(session).Should(Say("state\\s+since\\s+cpu\\s+memory\\s+disk\\s+details"))

								Eventually(session).Should(Say("#0\\s+(running|starting)\\s+.*\\d+\\.\\d+%.*of 128M.*of 128M"))
								Eventually(session).Should(Say("#1\\s+(running|starting)\\s+.*\\d+\\.\\d+%.*of 128M.*of 128M"))
								Eventually(session).Should(Exit(0))
							})
						})
					})
				})
			})
		})
	})
})
