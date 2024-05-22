package isolated

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	. "code.cloudfoundry.org/cli/cf/util/testhelpers/matchers"
	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("apply-manifest command", func() {
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

		appDir = helpers.TempDirAbsolutePath("", "simple-app")

		manifestPath = filepath.Join(appDir, "manifest.yml")
		// Ensure the file exists at the minimum
		helpers.WriteManifest(manifestPath, map[string]interface{}{})
	})

	AfterEach(func() {
		Expect(os.RemoveAll(appDir)).ToNot(HaveOccurred())
	})

	Describe("help", func() {
		It("appears in cf help -a", func() {
			session := helpers.CF("help", "-a")
			Eventually(session).Should(Exit(0))
			Expect(session).To(HaveCommandInCategoryWithDescription("apply-manifest", "SPACES", "Apply manifest properties to a space"))
		})

		When("--help flag is set", func() {
			It("displays command usage to output", func() {
				session := helpers.CF("apply-manifest", "--help")

				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("apply-manifest - Apply manifest properties to a space"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf apply-manifest -f APP_MANIFEST_PATH"))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("create-app, create-app-manifest, push"))

				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "apply-manifest", "-f", manifestPath)
		})
	})

	When("the environment is set up correctly", func() {
		BeforeEach(func() {
			helpers.SetupCF(orgName, spaceName)
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
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
				session := helpers.CF("apply-manifest", "-f", manifestPath)
				Eventually(session.Err).Should(Say("For application at index 0: Name must not be empty"))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session).Should(Exit(1))
			})
		})

		When("there is a CC error", func() {
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

			It("displays the error", func() {
				session := helpers.CF("apply-manifest", "-f", manifestPath)
				Eventually(session.Err).Should(Say("Instances must be greater than or equal to 0"))
				Eventually(session).Should(Say("FAILED"))

				Eventually(session).Should(Exit(1))
			})
		})

		When("-f is provided", func() {
			When("the -f flag is not given an arg", func() {
				It("tells the user that the flag requires an arg, prints help text, and exits 1", func() {
					session := helpers.CF("apply-manifest", "-f")

					Eventually(session.Err).Should(Say("Incorrect Usage: expected argument for flag `-f'"))
					Eventually(session).Should(Say("NAME:"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("the -f flag points to a directory that does not have a manifest.yml file", func() {
				var (
					emptyDir string
				)

				BeforeEach(func() {
					emptyDir = helpers.TempDirAbsolutePath("", "empty")
				})

				AfterEach(func() {
					Expect(os.RemoveAll(emptyDir)).ToNot(HaveOccurred())
				})

				It("tells the user that the provided path doesn't exist, prints help text, and exits 1", func() {
					session := helpers.CF("apply-manifest", "-f", emptyDir)

					Eventually(session.Err).Should(helpers.SayPath("Incorrect Usage: The specified directory '%s' does not contain a file named 'manifest.yml'.", emptyDir))
					Eventually(session).Should(Say("NAME:"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("the -f flag points to a file that does not exist", func() {
				It("tells the user that the provided path doesn't exist, prints help text, and exits 1", func() {
					session := helpers.CF("apply-manifest", "-f", "path/that/does/not/exist")

					Eventually(session.Err).Should(Say("Incorrect Usage: The specified path 'path/that/does/not/exist' does not exist."))
					Eventually(session).Should(Say("NAME:"))
					Eventually(session).Should(Exit(1))
				})
			})

			When("the manifest exists where -f points", func() {
				It("applies the manifest successfully", func() {
					userName, _ := helpers.GetCredentials()
					helpers.WriteManifest(filepath.Join(appDir, "manifest.yml"), map[string]interface{}{
						"applications": []map[string]interface{}{
							{
								"name":      appName,
								"instances": 3,
							},
						},
					})

					session := helpers.CF("apply-manifest", "-f", appDir)
					Eventually(session).Should(Say("Applying manifest %s in org %s / space %s as %s...", regexp.QuoteMeta(manifestPath), orgName, spaceName, userName))
					Eventually(session).Should(Exit())

					session = helpers.CF("app", appName)
					Eventually(session).Should(Say(`instances:\s+%s`, `\d/3`))
					Eventually(session).Should(Exit())
				})
			})

		})

		When("-f is not provided", func() {
			When("a properly formatted manifest is present in the pwd", func() {
				It("autodetects and applies the manifest", func() {
					userName, _ := helpers.GetCredentials()
					helpers.WriteManifest(filepath.Join(appDir, "manifest.yml"), map[string]interface{}{
						"applications": []map[string]interface{}{
							{
								"name":      appName,
								"instances": 3,
							},
						},
					})

					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "apply-manifest")
					formatString := fmt.Sprintf("Applying manifest %%s in org %s / space %s as %s...", orgName, spaceName, userName)
					Eventually(session).Should(helpers.SayPath(formatString, manifestPath))
					Eventually(session).Should(Exit())

					session = helpers.CF("app", appName)
					Eventually(session).Should(Say(`instances:\s+%s`, `\d/3`))
					Eventually(session).Should(Exit())
				})
			})

			When("the current directory does not have a manifest", func() {
				It("fails nicely", func() {
					currentDir, err := os.Getwd()
					Expect(err).NotTo(HaveOccurred())
					session := helpers.CF("apply-manifest")

					Eventually(session.Err).Should(helpers.SayPath(`Could not find 'manifest.yml' file in %s`, currentDir))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session).Should(Exit(1))
				})
			})
		})

		When("testing manifest diffing output", func() {
			var (
				userName string
			)

			BeforeEach(func() {
				userName, _ = helpers.GetCredentials()
			})

			When("there are no changes in the manifest", func() {
				It("shows no changes", func() {
					helpers.WithHelloWorldApp(func(dir string) {
						manifest, manifestPath := pushAppAndGenerateManifest(appName, dir)
						helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), manifest)
						session := helpers.CF("apply-manifest", "-f", manifestPath)
						Eventually(session).Should(Say("Applying manifest %s in org %s / space %s as %s...", regexp.QuoteMeta(manifestPath), orgName, spaceName, userName))
						Eventually(session).Should(Say("Updating with these attributes..."))
						Consistently(session).ShouldNot(Say(`^\+ `))
						Consistently(session).ShouldNot(Say(`^- `))
						Eventually(session).Should(Exit(0))
					})
				})
			})

			When("there are changes in the manifest", func() {
				It("shows changes", func() {
					helpers.WithHelloWorldApp(func(dir string) {
						manifest, manifestPath := pushAppAndGenerateManifest(appName, dir)
						helpers.WriteManifest(filepath.Join(dir, "manifest.yml"), manifest)

						session := helpers.CF("scale", appName, "-i", "3")
						Eventually(session).Should(Exit(0))

						session = helpers.CF("apply-manifest", "-f", manifestPath)
						Eventually(session).Should(Say("Applying manifest %s in org %s / space %s as %s...", regexp.QuoteMeta(manifestPath), orgName, spaceName, userName))
						Eventually(session).Should(Say("Updating with these attributes..."))
						Eventually(session).Should(Say("applications"))
						Eventually(session).Should(Say(`\n-\s+instances: 3`))
						Eventually(session).Should(Say(`\n\+\s+instances: 1`))
						Eventually(session).Should(Exit(0))
					})
				})
			})
		})
		When("--redact-env flag is provided", func() {
			var (
				tempDir        string
				pathToManifest string
			)

			BeforeEach(func() {
				var err error
				tempDir, err = os.MkdirTemp("", "simple-manifest-test")
				Expect(err).ToNot(HaveOccurred())
				pathToManifest = filepath.Join(tempDir, "manifest.yml")
				helpers.WriteManifest(pathToManifest, map[string]interface{}{
					"applications": []map[string]interface{}{
						{
							"name": appName,
							"env": map[string]interface{}{
								"key1": "((var1))",
								"key4": "((var2))",
							},
						},
					},
				})
			})

			It("uses the manifest with substituted variables", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					session := helpers.CF("apply-manifest", "--redact-env", "-f", pathToManifest, "--var=var1=secret-key", "--var=var2=foobar")
					Eventually(session).Should(Exit(0))
					Eventually(session).Should(Say(`key1: <redacted>`))
					Consistently(session).ShouldNot(Say(`foobar`))
					Consistently(session).ShouldNot(Say(`secret-key`))
				})

			})
		})

		When("--vars are provided", func() {
			var (
				tempDir        string
				pathToManifest string
			)

			BeforeEach(func() {
				var err error
				tempDir, err = os.MkdirTemp("", "simple-manifest-test")
				Expect(err).ToNot(HaveOccurred())
				pathToManifest = filepath.Join(tempDir, "manifest.yml")
				helpers.WriteManifest(pathToManifest, map[string]interface{}{
					"applications": []map[string]interface{}{
						{
							"name": appName,
							"env": map[string]interface{}{
								"key1": "((var1))",
								"key4": "((var2))",
							},
						},
					},
				})
			})

			It("uses the manifest with substituted variables", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					session := helpers.CF("apply-manifest", "-f", pathToManifest, "--var=var1=secret-key", "--var=var2=foobar")
					Eventually(session).Should(Exit(0))
				})

				session := helpers.CF("env", appName)
				Eventually(session).Should(Say(`key1:\s+secret-key`))
				Eventually(session).Should(Say(`key4:\s+foobar`))
				Eventually(session).Should(Exit(0))
			})
		})
	})
})

func pushAppAndGenerateManifest(appName, dir string) (map[string]interface{}, string) {
	session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName)
	Eventually(session).Should(Exit(0))
	manifestPath := filepath.Join(dir, "manifest.yml")
	session = helpers.CF("create-app-manifest", appName, "-p", manifestPath)
	Eventually(session).Should(helpers.SayPath(`Manifest file created successfully at %s`, manifestPath))
	Eventually(session).Should(Exit(0))
	manifest := helpers.ReadManifest(manifestPath)

	return manifest, manifestPath
}
