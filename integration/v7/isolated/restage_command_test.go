package isolated

import (
	"fmt"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("restage command", func() {
	Describe("help", func() {
		When("--help flag is set", func() {
			It("Displays command usage to output", func() {
				session := helpers.CF("restage", "--help")

				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say(`restage - Stage the app's latest package into a droplet and restart the app with this new droplet and updated configuration \(environment variables, service bindings, buildpack, stack, etc.\).`))
				Eventually(session).ShouldNot(Say(`This action will cause app downtime.`))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf restage APP_NAME"))
				Eventually(session).Should(Say("This command will cause downtime unless you use '--strategy"))
				Eventually(session).Should(Say("EXAMPLES:"))
				Eventually(session).Should(Say("cf restage APP_NAME"))
				Eventually(session).Should(Say("cf restage APP_NAME --strategy rolling"))
				Eventually(session).Should(Say("cf restage APP_NAME --strategy canary --no-wait"))
				Eventually(session).Should(Say("ALIAS:"))
				Eventually(session).Should(Say("rg"))
				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say("--instance-steps"))
				Eventually(session).Should(Say("--max-in-flight"))
				Eventually(session).Should(Say(`--no-wait\s+Exit when the first instance of the web process is healthy`))
				Eventually(session).Should(Say(`--strategy\s+Deployment strategy can be canary, rolling or null`))
				Eventually(session).Should(Say("ENVIRONMENT:"))
				Eventually(session).Should(Say(`CF_STAGING_TIMEOUT=15\s+Max wait time for staging, in minutes`))
				Eventually(session).Should(Say(`CF_STARTUP_TIMEOUT=5\s+Max wait time for app instance startup, in minutes`))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("restart"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "restage", "app-name")
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
			It("tells the user that the app is not found and exits 1", func() {
				appName := helpers.PrefixedRandomName("app")
				session := helpers.CF("restage", appName)

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
			When("there are no packages for the app to restage", func() {
				BeforeEach(func() {
					appName = helpers.PrefixedRandomName("app")
					Eventually(helpers.CF("create-app", appName)).Should(Exit(0))
				})

				It("fails and displays the package not found failure message", func() {
					userName, _ := helpers.GetCredentials()
					session := helpers.CF("restage", appName)
					Eventually(session).Should(Say(`Restaging app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))

					Eventually(session.Err).Should(Say(`App '%s' has no eligible packages\.`, appName))
					Eventually(session.Err).Should(Say(`TIP: Use 'cf packages %s' to list packages in your app. Use 'cf create-package' to create one\.`, appName))
					Eventually(session).Should(Exit(1))
				})
			})

			When("there is an error in staging the app", func() {
				BeforeEach(func() {
					appName = helpers.PrefixedRandomName("app")
					domainName = helpers.DefaultSharedDomain()
					helpers.WithHelloWorldApp(func(appDir string) {
						err := os.Remove(filepath.Join(appDir, "Staticfile"))
						Expect(err).ToNot(HaveOccurred())
						Eventually(helpers.CF("push", appName, "-p", appDir)).Should(Exit(1))
					})
				})

				It("fails and displays the staging failure message", func() {
					userName, _ := helpers.GetCredentials()
					session := helpers.CF("restage", appName)
					Eventually(session).Should(Say(`Restaging app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))

					// The staticfile_buildpack does compile an index.html file. However, it requires a "Staticfile" during buildpack detection.
					Eventually(session.Err).Should(Say("Error staging application: NoAppDetectedError - An app was not successfully detected by any available buildpack"))
					Eventually(session.Err).Should(Say(`TIP: Use 'cf buildpacks' to see a list of supported buildpacks.`))
					Eventually(session).Should(Exit(1))
				})
			})

			When("the app does *not* start properly", func() {
				BeforeEach(func() {
					appName = helpers.PrefixedRandomName("app")
					helpers.WithHelloWorldApp(func(appDir string) {
						Eventually(helpers.CF("push", appName, "-p", appDir, "-b", "staticfile_buildpack", "-c", "gibberish")).Should(Exit(1))
					})
				})

				It("fails and displays the start failure message", func() {
					userName, _ := helpers.GetCredentials()
					session := helpers.CF("restage", appName)
					Eventually(session).Should(Say(`Restaging app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))

					Eventually(session.Err).Should(Say("Start unsuccessful"))
					Eventually(session.Err).Should(Say("TIP: use 'cf logs .* --recent' for more information"))
					Eventually(session).Should(Exit(1))
				})

				When("a deployment strategy is used", func() {
					It("fails and displays the deployment failure message", func() {
						userName, _ := helpers.GetCredentials()
						session := helpers.CustomCF(helpers.CFEnv{
							EnvVars: map[string]string{"CF_STARTUP_TIMEOUT": "0.1"},
						}, "restage", appName, "--strategy", "canary", "--max-in-flight", "3")
						Consistently(session.Err).ShouldNot(Say(`This action will cause app downtime\.`))
						Eventually(session).Should(Say(`Restaging app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
						Eventually(session).Should(Say(`Creating deployment for app %s\.\.\.`, appName))
						Eventually(session).Should(Say(`Waiting for app to deploy\.\.\.`))
						Eventually(session.Err).Should(Say(`Start app timeout`))
						Eventually(session.Err).Should(Say(`TIP: Application must be listening on the right port\.`))
						Eventually(session).Should(Say("FAILED"))
						Eventually(session).Should(Exit(1))

						appGUID := helpers.AppGUID(appName)
						Eventually(func() *Buffer {
							session_deployment := helpers.CF("curl", fmt.Sprintf("/v3/deployments?app_guids=%s", appGUID))
							Eventually(session_deployment).Should(Exit(0))
							return session_deployment.Out
						}).Should(Say(`"reason":\s*"CANCELED"`))
					})
				})
			})

			When("the app stages and starts properly", func() {
				BeforeEach(func() {
					appName = helpers.PrefixedRandomName("app")
					domainName = helpers.DefaultSharedDomain()
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

						Eventually(helpers.CF("push", appName, "-p", appDir, "-f", manifestPath, "-b", "staticfile_buildpack")).Should(Exit(0))
					})
				})

				It("uses the multiprocess display", func() {
					userName, _ := helpers.GetCredentials()
					session := helpers.CF("restage", appName)
					Eventually(session.Err).Should(Say(`This action will cause app downtime\.`))
					Eventually(session).Should(Say(`Restaging app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))

					helpers.ConfirmStagingLogs(session)

					Eventually(session).Should(Say(`name:\s+%s`, appName))
					Eventually(session).Should(Say(`requested state:\s+started`))
					Eventually(session).Should(Say(`routes:\s+%s\.%s`, appName, domainName))
					Eventually(session).Should(Say(`last uploaded:\s+%s`, helpers.ReadableDateTimeRegex))
					Eventually(session).Should(Say(`stack:\s+cflinuxfs`))
					Eventually(session).Should(Say(`buildpacks:\s+\n`))
					Eventually(session).Should(Say(`staticfile_buildpack\s+\d+.\d+.\d+\s+`))
					Eventually(session).Should(Say(`type:\s+web`))
					Eventually(session).Should(Say(`sidecars:`))
					Eventually(session).Should(Say(`instances:\s+\d/2`))
					Eventually(session).Should(Say(`memory usage:\s+128M`))
					Eventually(session).Should(Say(`\s+state\s+since\s+cpu entitlement\s+memory\s+disk\s+logging`))
					Eventually(session).Should(Say(`#0\s+(starting|running)\s+\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z`))
					Eventually(session).Should(Exit(0))
				})

				When("strategy rolling is given", func() {
					It("creates a deploy", func() {
						userName, _ := helpers.GetCredentials()
						session := helpers.CF("restage", appName, "--strategy=rolling")
						Consistently(session.Err).ShouldNot(Say(`This action will cause app downtime\.`))
						Eventually(session).Should(Say(`Restaging app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
						Eventually(session).Should(Say(`Creating deployment for app %s\.\.\.`, appName))
						Eventually(session).Should(Say(`Waiting for app to deploy\.\.\.`))
						Eventually(session).Should(Say(`name:\s+%s`, appName))
						Eventually(session).Should(Say(`requested state:\s+started`))
						Eventually(session).Should(Say(`routes:\s+%s\.%s`, appName, domainName))
						Eventually(session).Should(Say(`last uploaded:\s+%s`, helpers.ReadableDateTimeRegex))
						Eventually(session).Should(Say(`stack:\s+cflinuxfs`))
						Eventually(session).Should(Say(`buildpacks:\s+\n`))
						Eventually(session).Should(Say(`staticfile_buildpack\s+\d+.\d+.\d+\s+`))
						Eventually(session).Should(Say(`type:\s+web`))
						Eventually(session).Should(Say(`sidecars:`))
						Eventually(session).Should(Say(`instances:\s+\d/2`))
						Eventually(session).Should(Say(`memory usage:\s+128M`))
						Eventually(session).Should(Say(`\s+state\s+since\s+cpu entitlement\s+memory\s+disk\s+logging`))
						Eventually(session).Should(Say(`#0\s+(starting|running)\s+\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z`))
						Eventually(session).Should(Exit(0))
					})
				})

				When("strategy canary is given without the max-in-flight flag", func() {
					It("restages successfully and notes the max-in-flight value", func() {
						userName, _ := helpers.GetCredentials()
						session := helpers.CF("restage", appName, "--strategy", "canary")
						Consistently(session.Err).ShouldNot(Say(`This action will cause app downtime\.`))
						Eventually(session).Should(Say(`Restaging app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
						Eventually(session).Should(Say(`Creating deployment for app %s\.\.\.`, appName))
						Eventually(session).Should(Say(`Waiting for app to deploy\.\.\.`))
						Eventually(session).Should(Say("Active deployment with status PAUSED"))
						Eventually(session).Should(Say("strategy:        canary"))
						Eventually(session).Should(Say("max-in-flight:   1"))
						Eventually(session).Should(Say("Please run `cf continue-deployment %s` to promote the canary deployment, or `cf cancel-deployment %s` to rollback to the previous version.", appName, appName))
						Eventually(session).Should(Exit(0))
					})
				})

				When("strategy canary is given with a non-default max-in-flight value and instance-steps", func() {
					It("restages successfully and notes the max-in-flight value", func() {
						userName, _ := helpers.GetCredentials()
						session := helpers.CF("restage", appName, "--strategy", "canary", "--max-in-flight", "2", "--instance-steps", "1,20")
						Consistently(session.Err).ShouldNot(Say(`This action will cause app downtime\.`))
						Eventually(session).Should(Say(`Restaging app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
						Eventually(session).Should(Say(`Creating deployment for app %s\.\.\.`, appName))
						Eventually(session).Should(Say(`Waiting for app to deploy\.\.\.`))
						Eventually(session).Should(Say("Active deployment with status PAUSED"))
						Eventually(session).Should(Say("strategy:        canary"))
						Eventually(session).Should(Say("max-in-flight:   2"))
						Eventually(session).Should(Say("Please run `cf continue-deployment %s` to promote the canary deployment, or `cf cancel-deployment %s` to rollback to the previous version.", appName, appName))
						Eventually(session).Should(Exit(0))
					})
				})

				XWhen("isolation segments are available", func() {
					BeforeEach(func() {
						Eventually(helpers.CF("create-isolation-segment", RealIsolationSegment)).Should(Exit(0))
						Eventually(helpers.CF("enable-org-isolation", orgName, RealIsolationSegment)).Should(Exit(0))
						Eventually(helpers.CF("set-space-isolation-segment", spaceName, RealIsolationSegment)).Should(Exit(0))
					})

					It("displays app isolation segment information", func() {
						session := helpers.CF("restage", appName)
						Eventually(session).Should(Say(`isolation segment:\s+%s`, RealIsolationSegment))
					})
				})
			})
		})
	})
})
