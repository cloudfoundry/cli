package experimental

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

var _ = Describe("create-app-manifest command", func() {
	var appName string
	var manifestFilePath string
	var tempDir string

	BeforeEach(func() {
		appName = helpers.NewAppName()
		var err error
		tempDir, err = ioutil.TempDir("", "create-manifest")
		Expect(err).ToNot(HaveOccurred())

		manifestFilePath = filepath.Join(tempDir, fmt.Sprintf("%s_manifest.yml", appName))
	})

	AfterEach(func() {
		os.RemoveAll(tempDir)
	})

	Context("Help", func() {
		It("displays the help information", func() {
			session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: tempDir}, "create-route", "--help")
			Eventually(session.Out).Should(Say("NAME:"))
			Eventually(session.Out).Should(Say("create-app-manifest - Create an app manifest for an app that has been pushed successfully"))
			Eventually(session.Out).Should(Say("USAGE:"))
			Eventually(session.Out).Should(Say("cf create-app-manifest APP_NAME [-p /path/to/<app-name>-manifest.yml]"))
			Eventually(session.Out).Should(Say(""))
			Eventually(session.Out).Should(Say("OPTIONS:"))
			Eventually(session.Out).Should(Say("-p      Specify a path for file creation. If path not specified, manifest file is created in current working directory."))
			Eventually(session.Out).Should(Say("SEE ALSO:"))
			Eventually(session.Out).Should(Say("apps, push"))

		})
	})

	Context("when the environment is not setup correctly", func() {
		Context("when no API endpoint is set", func() {
			BeforeEach(func() {
				helpers.UnsetAPI()
			})

			It("fails with no API endpoint set message", func() {
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: tempDir}, "create-app-manifest", appName)
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("No API endpoint set. Use 'cf login' or 'cf api' to target an endpoint."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when not logged in", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
			})

			It("fails with not logged in message", func() {
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: tempDir}, "create-app-manifest", appName)
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("Not logged in. Use 'cf login' to log in."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when there is no org set", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
				helpers.LoginCF()
			})

			It("fails with no targeted org error message", func() {
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: tempDir}, "create-app-manifest", appName)
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("No org targeted, use 'cf target -o ORG' to target an org."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when there is no space set", func() {
			BeforeEach(func() {
				helpers.LogoutCF()
				helpers.LoginCF()
				helpers.TargetOrg(ReadOnlyOrg)
			})

			It("fails with no targeted space error message", func() {
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: tempDir}, "create-app-manifest", appName)
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("No space targeted, use 'cf target -s SPACE' to target a space."))
				Eventually(session).Should(Exit(1))
			})
		})
	})

	Context("when app name not provided", func() {
		It("displays a usage error", func() {
			session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: tempDir}, "create-app-manifest")
			Eventually(session.Out).Should(Say("Incorrect Usage: the required argument `APP_NAME` was not provided"))
			Eventually(session.Out).Should(Say("USAGE:"))
		})
	})

	Context("when the environment is setup correctly", func() {
		var (
			orgName   string
			spaceName string

			domainName string
		)

		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()

			setupCF(orgName, spaceName)
			domainName = defaultSharedDomain()
		})

		Context("when the app does not exist", func() {
			It("displays a usage error", func() {
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: tempDir}, "create-app-manifest", appName)
				Eventually(session.Out).Should(Say("Creating an app manifest from current settings of app %s ...", appName))
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Out).Should(Say("App %s not found", appName))
			})
		})

		Context("when the app exists", func() {
			BeforeEach(func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "v2-push", appName)).Should(Exit(0))
				})
			})

			FIt("creates the manifest", func() {
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: tempDir}, "create-app-manifest", appName)
				Eventually(session.Out).Should(Say("Creating an app manifest from current settings of app %s ...", appName))
				Eventually(session.Out).Should(Say("OK"))
				Eventually(session.Out).Should(Say("Manifest file created successfully at ./%s_manifest.yml", appName))

				expectedFile := fmt.Sprintf(`applications:
- name: %s
  instances: 1
  memory: 32M
  disk_quota: 1024M
  routes:
  - route: %s.%s
  stack: cflinuxfs2
`, appName, appName, domainName)

				createdFile, err := ioutil.ReadFile(manifestFilePath)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(createdFile)).To(Equal(expectedFile))
			})

			Context("when the -p flag is provided", func() {
				Context("when the specified file is a directory", func() {
					It("displays a file creation error", func() {
						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: tempDir}, "create-app-manifest", appName, "-p", tempDir)
						Eventually(session.Out).Should(Say("Creating an app manifest from current settings of app %s ...", appName))
						Eventually(session.Out).Should(Say("FAILED"))
						Eventually(session.Err).Should(Say("Error creating manifest file: open ./: is a directory"))
					})
				})

				Context("when you don't have permissions to write the specified file", func() {
					var tempFile string

					BeforeEach(func() {
						tempFile = filepath.Join(tempDir, "some-file")
						f, err := os.OpenFile(tempFile, os.O_CREATE, 0000)
						Expect(err).ToNot(HaveOccurred())
						Expect(f.Close()).To(Succeed())
					})

					It("displays a file creation error", func() {
						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: tempDir}, "create-app-manifest", appName, "-p", tempFile)
						Eventually(session.Out).Should(Say("Creating an app manifest from current settings of app %s ...", appName))
						Eventually(session.Out).Should(Say("FAILED"))
						Eventually(session.Err).Should(Say("Error creating manifest file: open /core: permission denied"))
					})
				})

				Context("when the specified file does not exist", func() {
					var newFile string

					BeforeEach(func() {
						newFile = filepath.Join(tempDir, "new-file.yml")
					})

					It("creates the manifest in the file", func() {
						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: tempDir}, "create-app-manifest", appName, "-p", newFile)
						Eventually(session.Out).Should(Say("Creating an app manifest from current settings of app %s ...", appName))
						Eventually(session.Out).Should(Say("OK"))
						Eventually(session.Out).Should(Say("Manifest file created successfully at %s", newFile))

						expectedFile := fmt.Sprintf(`applications:
- name: %s
  instances: 1
  memory: 32M
  disk_quota: 1024M
  routes:
  - route: %s.%s
  stack: cflinuxfs2
`, appName, appName, domainName)

						createdFile, err := ioutil.ReadFile(newFile)
						Expect(err).ToNot(HaveOccurred())
						Expect(string(createdFile)).To(Equal(expectedFile))
					})
				})

				Context("when the specified file exists", func() {
					var existingFile string

					BeforeEach(func() {
						existingFile = filepath.Join(tempDir, "some-file")
						f, err := os.Create(existingFile)
						Expect(err).ToNot(HaveOccurred())
						Expect(f.Close()).To(Succeed())
					})

					It("overrides the previous file with the new manifest", func() {
						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: tempDir}, "create-app-manifest", appName, "-p", existingFile)
						Eventually(session.Out).Should(Say("Creating an app manifest from current settings of app %s ...", appName))
						Eventually(session.Out).Should(Say("OK"))
						Eventually(session.Out).Should(Say("Manifest file created successfully at %s", existingFile))

						expectedFile := fmt.Sprintf(`applications:
- name: %s
  instances: 1
  memory: 32M
  disk_quota: 1024M
  routes:
  - route: %s.%s
  stack: cflinuxfs2
`, appName, appName, domainName)

						createdFile, err := ioutil.ReadFile(existingFile)
						Expect(err).ToNot(HaveOccurred())
						Expect(string(createdFile)).To(Equal(expectedFile))
					})
				})
			})
		})
	})
})
