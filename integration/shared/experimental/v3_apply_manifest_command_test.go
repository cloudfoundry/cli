// +build !partialPush

package experimental

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("v3-apply-manifest command", func() {
	var (
		orgName      string
		spaceName    string
		appName      string
		manifestPath string
		appDir       string
	)

	BeforeEach(func() {
		orgName = helpers.NewOrgName()
		spaceName = helpers.NewSpaceName()
		appName = helpers.PrefixedRandomName("app")
		appDir, _ = ioutil.TempDir("", "simple-app")
		manifestPath = filepath.Join(appDir, "manifest.yml")
		// Ensure the file exists at the minimum
		helpers.WriteManifest(manifestPath, map[string]interface{}{})
	})

	AfterEach(func() {
		Expect(os.RemoveAll(appDir)).ToNot(HaveOccurred())
	})

	Describe("help", func() {
		When("--help flag is set", func() {
			It("displays command usage to output", func() {
				session := helpers.CF("v3-apply-manifest", "--help")

				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("v3-apply-manifest - Applies manifest properties to an application"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf v3-apply-manifest -f APP_MANIFEST_PATH"))

				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the -f flag is not given an arg", func() {
		It("tells the user that the flag requires an arg, prints help text, and exits 1", func() {
			session := helpers.CF("v3-apply-manifest", "-f")

			Eventually(session.Err).Should(Say("Incorrect Usage: expected argument for flag `-f'"))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	When("the -f flag path does not exist", func() {
		It("tells the user that the provided path doesn't exist, prints help text, and exits 1", func() {
			session := helpers.CF("v3-apply-manifest", "-f", "path/that/does/not/exist")

			Eventually(session.Err).Should(Say("Incorrect Usage: The specified path 'path/that/does/not/exist' does not exist."))
			Eventually(session).Should(Say("NAME:"))
			Eventually(session).Should(Exit(1))
		})
	})

	When("the environment is not setup correctly", func() {
		When("the v3 api version is lower than the minimum version", func() {
			var server *Server

			BeforeEach(func() {
				server = helpers.StartAndTargetServerWithAPIVersions(helpers.DefaultV2Version, ccversion.MinV3ClientVersion)
			})

			AfterEach(func() {
				server.Close()
			})

			It("fails with error message that the minimum version is not met", func() {
				session := helpers.CF("v3-apply-manifest", "-f", manifestPath)
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say(`This command requires CF API version 3\.27\.0 or higher\.`))
				Eventually(session).Should(Exit(1))
			})
		})

		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "v3-apply-manifest", "-f", manifestPath)
		})
	})

	When("the environment is set up correctly", func() {
		BeforeEach(func() {
			helpers.SetupCF(orgName, spaceName)
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		When("the app exists", func() {
			BeforeEach(func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "v3-push", appName)).Should(Exit(0))
				})
			})

			When("the app name in the manifest is missing", func() {
				BeforeEach(func() {
					helpers.WriteManifest(manifestPath, map[string]interface{}{
						"applications": []map[string]interface{}{
							{
								"instances": 3,
							},
						},
					})
				})

				It("reports an error", func() {
					session := helpers.CF("v3-apply-manifest", "-f", manifestPath)
					Eventually(session.Err).Should(Say("Found an application with no name specified"))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("the app name in the manifest doesn't exist", func() {
				var invalidAppName string
				BeforeEach(func() {
					invalidAppName = "no-such-app"
					helpers.WriteManifest(manifestPath, map[string]interface{}{
						"applications": []map[string]interface{}{
							{
								"name":      invalidAppName,
								"instances": 3,
							},
						},
					})
				})

				It("reports an error", func() {
					session := helpers.CF("v3-apply-manifest", "-f", manifestPath)
					Eventually(session.Err).Should(Say("App %s not found", invalidAppName))
					Eventually(session).Should(Say("FAILED"))

					Eventually(session).Should(Exit(1))
				})
			})

			When("the app name in the manifest does exist", func() {
				When("the instances value is negative", func() {
					BeforeEach(func() {
						helpers.WriteManifest(manifestPath, map[string]interface{}{
							"applications": []map[string]interface{}{
								{
									"name":      appName,
									"instances": -1,
								},
							},
						})
					})

					It("reports an error", func() {
						session := helpers.CF("v3-apply-manifest", "-f", manifestPath)
						Eventually(session.Err).Should(Say("Instances must be greater than or equal to 0"))
						Eventually(session).Should(Say("FAILED"))

						Eventually(session).Should(Exit(1))
					})
				})

				When("the instances value is more than the space quota limit", func() {
					BeforeEach(func() {
						Eventually(helpers.CF("create-space-quota", "some-space-quota-name", "-a", "4")).Should(Exit(0))
						Eventually(helpers.CF("set-space-quota", spaceName, "some-space-quota-name")).Should(Exit(0))
						helpers.WriteManifest(manifestPath, map[string]interface{}{
							"applications": []map[string]interface{}{
								{
									"name":      appName,
									"instances": 5,
								},
							},
						})
					})

					It("reports an error", func() {
						session := helpers.CF("v3-apply-manifest", "-f", manifestPath)
						Eventually(session.Err).Should(Say("memory space_quota_exceeded, app_instance_limit space_app_instance_limit_exceeded"))
						Eventually(session).Should(Say("FAILED"))

						Eventually(session).Should(Exit(1))
					})
				})

				When("instances are specified correctly", func() {
					BeforeEach(func() {
						helpers.WriteManifest(manifestPath, map[string]interface{}{
							"applications": []map[string]interface{}{
								{
									"name":      appName,
									"instances": 3,
								},
							},
						})
					})

					It("displays the experimental warning", func() {
						session := helpers.CF("v3-apply-manifest", "-f", manifestPath)
						Eventually(session.Err).Should(Say("This command is in EXPERIMENTAL stage and may change without notice"))
						Eventually(session).Should(Exit())
					})

					It("rescales the app", func() {
						session := helpers.CF("app", appName)
						userName, _ := helpers.GetCredentials()
						Eventually(session).Should(Say(`instances:\s+%s`, "1/1"))
						Eventually(session).Should(Exit())

						session = helpers.CF("v3-apply-manifest", "-f", manifestPath)
						Eventually(session).Should(Say("Applying manifest %s in org %s / space %s as %s...", regexp.QuoteMeta(manifestPath), orgName, spaceName, userName))
						Eventually(session).Should(Exit())

						session = helpers.CF("app", appName)
						Eventually(session).Should(Say(`instances:\s+%s`, `\d/3`))
						Eventually(session).Should(Exit())
					})
				})
			})

		})
	})
})
