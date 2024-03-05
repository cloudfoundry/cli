package isolated

import (
	"fmt"
	"io/ioutil"
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

var _ = Describe("create-app-manifest command", func() {
	var appName string
	var manifestFilePath string
	var tempDir string

	BeforeEach(func() {
		appName = helpers.NewAppName()
		tempDir = helpers.TempDirAbsolutePath("", "create-manifest")

		manifestFilePath = filepath.Join(tempDir, fmt.Sprintf("%s_manifest.yml", appName))
	})

	AfterEach(func() {
		os.RemoveAll(tempDir)
	})

	Context("Help", func() {
		When("--help flag is set", func() {
			It("appears in cf help -a", func() {
				session := helpers.CF("help", "-a")
				Eventually(session).Should(Exit(0))
				Expect(session).To(HaveCommandInCategoryWithDescription("create-app-manifest", "APPS", "Create an app manifest for an app that has been pushed successfully"))
			})

			It("displays the help information", func() {
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: tempDir}, "create-app-manifest", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("create-app-manifest - Create an app manifest for an app that has been pushed successfully"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say(`cf create-app-manifest APP_NAME \[-p \/path\/to\/<app-name>_manifest\.yml\]`))
				Eventually(session).Should(Say(""))
				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say("-p      Specify a path for file creation. If path not specified, manifest file is created in current working directory."))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("apps, push"))

				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the environment is not setup correctly", func() {
		It("fails with the appropriate errors", func() {
			helpers.CheckEnvironmentTargetedCorrectly(true, true, ReadOnlyOrg, "create-app-manifest", "some-app-name")
		})
	})

	When("app name not provided", func() {
		It("displays a usage error", func() {
			session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: tempDir}, "create-app-manifest")
			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `APP_NAME` was not provided"))
			Eventually(session).Should(Say("USAGE:"))

			Eventually(session).Should(Exit(1))
		})
	})

	When("the environment is setup correctly", func() {
		var (
			orgName   string
			spaceName string
			userName  string
		)

		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()

			helpers.SetupCF(orgName, spaceName)
			userName, _ = helpers.GetCredentials()
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		When("the app does not exist", func() {
			It("displays an app not found error", func() {
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: tempDir}, "create-app-manifest", appName)
				Eventually(session).Should(Say(`Creating an app manifest from current settings of app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
				Eventually(session.Err).Should(Say("App '%s' not found", appName))
				Eventually(session).Should(Say("FAILED"))

				Eventually(session).Should(Exit(1))
			})
		})

		When("the app exists", func() {
			BeforeEach(func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "push", appName, "--no-start")).Should(Exit(0))
				})
			})

			It("creates the manifest", func() {
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: tempDir}, "create-app-manifest", appName)
				Eventually(session).Should(Say(`Creating an app manifest from current settings of app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
				path := filepath.Join(tempDir, fmt.Sprintf("%s_manifest.yml", appName))
				Eventually(session).Should(helpers.SayPath("Manifest file created successfully at %s", path))
				Eventually(session).Should(Say("OK"))
				Eventually(session).Should(Exit(0))

				createdFile, err := ioutil.ReadFile(manifestFilePath)
				Expect(err).ToNot(HaveOccurred())
				Expect(createdFile).To(MatchRegexp("---"))
				Expect(createdFile).To(MatchRegexp("applications:"))
				Expect(createdFile).To(MatchRegexp("name: %s", appName))
			})

			When("the -p flag is provided", func() {
				When("the specified file is a directory", func() {
					It("displays a file creation error", func() {
						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: tempDir}, "create-app-manifest", appName, "-p", tempDir)
						Eventually(session).Should(Say(`Creating an app manifest from current settings of app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
						Eventually(session).Should(Say("FAILED"))
						Eventually(session.Err).Should(Say("Error creating file: open %s: is a directory", regexp.QuoteMeta(tempDir)))

						Eventually(session).Should(Exit(1))
					})
				})

				When("the specified path is invalid", func() {
					It("displays a file creation error", func() {
						invalidPath := filepath.Join(tempDir, "invalid", "path.yml")
						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: tempDir}, "create-app-manifest", appName, "-p", invalidPath)
						Eventually(session).Should(Say(`Creating an app manifest from current settings of app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
						Eventually(session).Should(Say("FAILED"))
						Eventually(session.Err).Should(Say("Error creating file: open %s:.*", regexp.QuoteMeta(invalidPath)))

						Eventually(session).Should(Exit(1))
					})
				})

				When("the specified file does not exist", func() {
					var newFile string

					BeforeEach(func() {
						newFile = filepath.Join(tempDir, "new-file.yml")
					})

					It("creates the manifest in the file", func() {
						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: tempDir}, "create-app-manifest", appName, "-p", newFile)
						Eventually(session).Should(Say(`Creating an app manifest from current settings of app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
						Eventually(session).Should(Say("Manifest file created successfully at %s", helpers.ConvertPathToRegularExpression(newFile)))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Exit(0))

						createdFile, err := ioutil.ReadFile(newFile)
						Expect(err).ToNot(HaveOccurred())
						Expect(createdFile).To(MatchRegexp("---"))
						Expect(createdFile).To(MatchRegexp("applications:"))
						Expect(createdFile).To(MatchRegexp("name: %s", appName))
					})
				})

				When("the specified file exists", func() {
					var existingFile string

					BeforeEach(func() {
						existingFile = filepath.Join(tempDir, "some-file")
						f, err := os.Create(existingFile)
						Expect(err).ToNot(HaveOccurred())
						Expect(f.Close()).To(Succeed())
					})

					It("overrides the previous file with the new manifest", func() {
						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: tempDir}, "create-app-manifest", appName, "-p", existingFile)
						Eventually(session).Should(Say(`Creating an app manifest from current settings of app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
						Eventually(session).Should(Say("Manifest file created successfully at %s", helpers.ConvertPathToRegularExpression(existingFile)))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Exit(0))

						createdFile, err := ioutil.ReadFile(existingFile)
						Expect(err).ToNot(HaveOccurred())
						Expect(createdFile).To(MatchRegexp("---"))
						Expect(createdFile).To(MatchRegexp("applications:"))
						Expect(createdFile).To(MatchRegexp("name: %s", appName))
					})
				})
			})
		})
	})
})
