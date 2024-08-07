package isolated

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("app command", func() {
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

	Describe("help", func() {
		When("--help flag is set", func() {
			It("appears in cf help -a", func() {
				session := helpers.CF("help", "-a")
				Eventually(session).Should(Exit(0))
				Expect(session).To(HaveCommandInCategoryWithDescription("app", "APPS", "Display health and status for an app"))
			})

			It("Displays command usage to output", func() {
				session := helpers.CF("app", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("app - Display health and status for an app"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf app APP_NAME"))
				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say("--guid      Retrieve and display the given app's guid.  All other health and status output for the app is suppressed."))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("apps, events, logs, map-route, push, unmap-route"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "app", appName)
		})
	})

	When("the environment is set up correctly", func() {
		BeforeEach(func() {
			helpers.SetupCF(orgName, spaceName)
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		Describe("version dependent display", func() {
			When("the app is created but not pushed", func() {
				BeforeEach(func() {
					Eventually(helpers.CF("create-app", appName)).Should(Exit(0))
				})

				It("displays blank fields for unpopulated fields", func() {
					session := helpers.CF("app", appName)
					Eventually(session).Should(Say(`name:\s+%s`, appName))
					Eventually(session).Should(Say(`requested state:\s+stopped`))
					Eventually(session).Should(Say(`routes:\s+\n`))
					Eventually(session).Should(Say(`last uploaded:\s+\n`))
					Eventually(session).Should(Say(`stack:\s+\n`))
					Eventually(session).Should(Say(`buildpacks:\s+\n`))
					Eventually(session).Should(Exit(0))
				})
			})

			When("the app is a buildpack app", func() {
				var domainName string

				BeforeEach(func() {
					domainName = helpers.DefaultSharedDomain()
				})

				When("the app is started and has 2 instances", func() {
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
							err := os.WriteFile(manifestPath, manifestContents, 0666)
							Expect(err).ToNot(HaveOccurred())

							// Create manifest
							Eventually(helpers.CF("push", appName, "-p", appDir, "-f", manifestPath, "-b", "staticfile_buildpack")).Should(Exit(0))
						})
					})

					// TODO: use multiprocess
					It("uses the multiprocess display", func() {
						userName, _ := helpers.GetCredentials()

						session := helpers.CF("app", appName)

						Eventually(session).Should(Say(`Showing health and status for app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))

						Eventually(session).ShouldNot(Say("This command is in EXPERIMENTAL stage and may change without notice"))
						Eventually(session).Should(Say(`name:\s+%s`, appName))
						Eventually(session).Should(Say(`requested state:\s+started`))
						Eventually(session).Should(Say(`routes:\s+%s\.%s`, appName, domainName))
						Eventually(session).Should(Say(`last uploaded:\s+%s`, helpers.ReadableDateTimeRegex))
						Eventually(session).Should(Say(`stack:\s+cflinuxfs`))
						Eventually(session).Should(Say(`staticfile_buildpack\s+\d+.\d+.\d+\s+staticfile\s+staticfile`))
						Eventually(session).Should(Say(`type:\s+web`))
						Eventually(session).Should(Say(`instances:\s+\d/2`))
						Eventually(session).Should(Say(`memory usage:\s+128M`))
						Eventually(session).Should(Say(`\s+state\s+since\s+cpu entitlement\s+memory\s+disk\s+logging\s+details`))
						Eventually(session).Should(Say(`#0\s+(starting|running)\s+\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z`))

						Eventually(session).Should(Exit(0))
					})
				})
			})

			When("the app is stopped", func() {
				BeforeEach(func() {
					helpers.WithHelloWorldApp(func(appDir string) {
						Eventually(helpers.CF("push", appName, "-p", appDir, "-b", "staticfile_buildpack", "--no-start")).Should(Exit(0))
					})
				})

				It("displays that there are no running instances of the app", func() {
					session := helpers.CF("app", appName)

					userName, _ := helpers.GetCredentials()
					Eventually(session).Should(Say(`Showing health and status for app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
					Eventually(session).Should(Say(`name:\s+%s`, appName))
					Eventually(session).Should(Say(`requested state:\s+stopped`))
					Eventually(session).Should(Say(`type:\s+web`))
					Eventually(session).Should(Say("#0\\s+down\\s+\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}Z"))
					Eventually(session).Should(Exit(0))
				})
			})

			When("all instances of the app are down", func() {
				BeforeEach(func() {
					infiniteMemQuota := helpers.QuotaName()
					Eventually(helpers.CF("create-quota", infiniteMemQuota, "-i", "-1", "-r", "-1", "-m", "2000G")).Should(Exit(0))
					Eventually(helpers.CF("set-quota", orgName, infiniteMemQuota)).Should(Exit(0))
					helpers.WithHelloWorldApp(func(appDir string) {
						Eventually(helpers.CF("push", appName, "-p", appDir)).Should(Exit(0))
					})
					Eventually(helpers.CFWithEnv(map[string]string{"CF_STAGING_TIMEOUT": "0.1", "CF_STARTUP_TIMEOUT": "0.1"}, "scale", appName, "-m", "1000G", "-f")).Should(Exit(1))
				})

				It("displays the down app instances", func() {
					session := helpers.CF("app", appName)
					userName, _ := helpers.GetCredentials()
					Eventually(session).Should(Say(`Showing health and status for app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
					Eventually(session).Should(Say("name:\\s+%s", appName))
					Eventually(session).Should(Say("requested state:\\s+started"))
					Eventually(session).Should(Say("type:\\s+web"))
					Eventually(session).Should(Say("#0\\s+down\\s+\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}Z"))
					Eventually(session).Should(Exit(0))
				})
			})

			When("the app has 0 instances", func() {
				BeforeEach(func() {
					helpers.WithHelloWorldApp(func(appDir string) {
						Eventually(helpers.CF("push", appName, "-p", appDir, "-b", "staticfile_buildpack", "-i", "0")).Should(Exit(0))
					})
				})

				It("displays the app information", func() {
					session := helpers.CF("app", appName)
					userName, _ := helpers.GetCredentials()

					Eventually(session).Should(Say(`Showing health and status for app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
					Eventually(session).Should(Say(`name:\s+%s`, appName))
					Eventually(session).Should(Say(`requested state:\s+started`))
					Eventually(session).Should(Say(`type:\s+web`))
					Eventually(session).Should(Say("There are no running instances of this process"))
					Eventually(session).Should(Exit(0))
				})
			})

			When("the --guid flag is given", func() {
				var appGUID string

				BeforeEach(func() {
					helpers.WithHelloWorldApp(func(appDir string) {
						Eventually(helpers.CF("push", appName, "-p", appDir, "-b", "staticfile_buildpack", "--no-start")).Should(Exit(0))
					})

					var AppInfo struct {
						Resources []struct {
							GUID string `json:"guid"`
						} `json:"resources"`
					}

					helpers.Curl(&AppInfo, "/v3/apps?names=%s", appName)
					appGUID = AppInfo.Resources[0].GUID
				})

				It("displays the app guid", func() {
					session := helpers.CF("app", "--guid", appName)
					Eventually(session).Should(Say(appGUID))
					Eventually(session).Should(Exit(0))
				})
			})

			When("the app uses multiple buildpacks", func() {
				BeforeEach(func() {
					helpers.WithMultiBuildpackApp(func(appDir string) {
						Eventually(helpers.CF("push", appName, "-p", appDir, "-b", "ruby_buildpack", "-b", "go_buildpack")).Should(Exit(0))
					})
				})

				It("displays the app buildpacks", func() {
					session := helpers.CF("app", appName)
					Eventually(session).Should(Say(`ruby_buildpack\s+\d+.\d+.\d+\s+ruby`))
					Eventually(session).Should(Exit(0))
				})
			})

			When("the app uses sidecars", func() {
				BeforeEach(func() {
					helpers.WithSidecarApp(func(appDir string) {
						Eventually(helpers.CF("push", appName, "-p", appDir, "-f", path.Join(appDir, "manifest.yml"), "-b", "staticfile_buildpack", "--no-start")).Should(Exit(0))
					}, appName)
				})

				It("displays the sidecars", func() {
					session := helpers.CF("app", appName)
					Eventually(session).Should(Say(`sidecars:\s+sidecar_name\s+`))
					Eventually(session).Should(Exit(0))
				})
			})

			When("there is an active deployment", func() {
				BeforeEach(func() {
					helpers.WithHelloWorldApp(func(appDir string) {
						Eventually(helpers.CF("push", appName, "-p", appDir, "-b", "staticfile_buildpack")).Should(Exit(0))
					})
				})

				When("the deployment strategy is rolling", func() {
					When("the deployment is in progress", func() {
						It("displays the message", func() {
							session := helpers.CF("restart", appName, "--strategy", "rolling")

							session1 := helpers.CF("app", appName)
							Eventually(session1).Should(Say("Rolling deployment currently DEPLOYING."))
							Eventually(session).Should(Exit(0))
							Eventually(session1).Should(Exit(0))
						})
					})
					When("the deployment is cancelled", func() {
						It("displays the message", func() {
							helpers.CF("restart", appName, "--strategy", "rolling")
							Eventually(func() *Session {
								return helpers.CF("cancel-deployment", appName).Wait()
							}).Should(Exit(0))

							session2 := helpers.CF("app", appName)
							Eventually(session2).Should(Say("Rolling deployment currently CANCELING."))
							Eventually(session2).Should(Exit(0))
						})
					})
				})

				When("the deployment strategy is canary", func() {
					When("the deployment is paused", func() {
						It("displays the message", func() {
							Eventually(helpers.CF("restart", appName, "--strategy", "canary")).Should(Exit(0))

							session1 := helpers.CF("app", appName)
							Eventually(session1).Should(Say("Canary deployment currently PAUSED."))
							Eventually(session1).Should(Exit(0))
						})
					})

					When("the deployment is cancelled after it is paused", func() {
						It("no deployment information is displayed", func() {
							Eventually(helpers.CF("restart", appName, "--strategy", "canary")).Should(Exit(0))
							Eventually(helpers.CF("cancel-deployment", appName)).Should(Exit(0))

							session2 := helpers.CF("app", appName)
							Eventually(session2).ShouldNot(Say("Canary deployment currently CANCELING."))
							Eventually(session2).Should(Exit(0))
						})
					})
				})
			})

			When("there is no active deployment", func() {
				BeforeEach(func() {
					helpers.WithHelloWorldApp(func(appDir string) {
						Eventually(helpers.CF("push", appName, "-p", appDir, "-b", "staticfile_buildpack")).Should(Exit(0))
					})
				})

				It("doesn not display the message", func() {
					session := helpers.CF("app", appName)
					Eventually(session).Should(Exit(0))
					Eventually(session).ShouldNot(Say(`\w+ deployment currently \w+`))
				})
			})
		})

		Describe("version independent display", func() {
			When("the app name is not provided", func() {
				It("tells the user that the app name is required, prints help text, and exits 1", func() {
					session := helpers.CF("app")

					Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `APP_NAME` was not provided"))
					Eventually(session).Should(Say("NAME:"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("the app does not exist", func() {
				When("no flags are given", func() {
					It("tells the user that the app is not found and exits 1", func() {
						session := helpers.CF("app", appName)

						Eventually(session).Should(Say("FAILED"))
						Eventually(session.Err).Should(Say("App '%s' not found", appName))
						Eventually(session).Should(Exit(1))
					})
				})

				When("the --guid flag is given", func() {
					It("tells the user that the app is not found and exits 1", func() {
						session := helpers.CF("app", "--guid", appName)

						Eventually(session).Should(Say("FAILED"))
						Eventually(session.Err).Should(Say("App '%s' not found", appName))
						Eventually(session).Should(Exit(1))
					})
				})
			})

			When("the app exists", func() {
				When("isolation segments are available", func() {
					BeforeEach(func() {
						Eventually(helpers.CF("create-isolation-segment", RealIsolationSegment)).Should(Exit(0))
						Eventually(helpers.CF("enable-org-isolation", orgName, RealIsolationSegment)).Should(Exit(0))
						Eventually(helpers.CF("set-space-isolation-segment", spaceName, RealIsolationSegment)).Should(Exit(0))

						helpers.WithHelloWorldApp(func(appDir string) {
							Eventually(helpers.CF("push", appName, "-p", appDir, "-b", "staticfile_buildpack")).Should(Exit(0))
						})
					})

					It("displays the app isolation segment information", func() {
						session := helpers.CF("app", appName)

						Eventually(session).Should(Say(`isolation segment:\s+%s`, RealIsolationSegment))
						Eventually(session).Should(Exit(0))
					})
				})

				When("isolation segment is not set for the application", func() {
					BeforeEach(func() {
						helpers.WithHelloWorldApp(func(appDir string) {
							Eventually(helpers.CF("push", appName, "-p", appDir, "-b", "staticfile_buildpack")).Should(Exit(0))
						})
					})

					It("displays the app isolation segment information", func() {
						session := helpers.CF("app", appName)

						Consistently(session).ShouldNot(Say("isolation segment:"))
						Eventually(session).Should(Exit(0))
					})
				})

				When("the app is a Docker app", func() {
					BeforeEach(func() {
						Eventually(helpers.CF("push", appName, "-o", DockerImage)).Should(Exit())
					})

					It("displays the docker image", func() {
						session := helpers.CF("app", appName)
						Eventually(session).Should(Say(`name:\s+%s`, appName))
						Eventually(session).Should(Say(`docker image:\s+%s`, DockerImage))
						Eventually(session).Should(Exit(0))
					})

					It("does not display buildpack", func() {
						session := helpers.CF("app", appName)
						Consistently(session).ShouldNot(Say("buildpacks?:"))
						Eventually(session).Should(Exit(0))
					})
				})

				When("the app has tcp routes", func() {
					var tcpDomain helpers.Domain

					BeforeEach(func() {
						tcpDomain = helpers.NewDomain(orgName, helpers.NewDomainName("tcp"))
						tcpDomain.CreateWithRouterGroup(helpers.FindOrCreateTCPRouterGroup(GinkgoParallelProcess()))
						helpers.WithHelloWorldApp(func(appDir string) {
							manifestContents := []byte(fmt.Sprintf(`
---
applications:
- name: %s
  routes:
  - route: %s:1024
`, appName, tcpDomain.Name))
							manifestPath := filepath.Join(appDir, "manifest.yml")
							err := os.WriteFile(manifestPath, manifestContents, 0666)
							Expect(err).ToNot(HaveOccurred())

							// Create manifest
							Eventually(helpers.CF("push", appName, "-p", appDir, "-f", manifestPath, "-b", "staticfile_buildpack")).Should(Exit(0))
						})

						It("displays the app information", func() {
							session := helpers.CF("app", appName)
							Eventually(session).Should(Say(`routes:\s+[\w\d-]+\.%s`, tcpDomain))
							Eventually(session).Should(Exit(0))
						})
					})
				})
			})
		})
	})
})
