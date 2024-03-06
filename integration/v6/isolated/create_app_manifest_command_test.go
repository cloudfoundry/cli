package isolated

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/util/manifest"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	yaml "gopkg.in/yaml.v2"
)

func createManifest(appName string) (manifest.Manifest, string, error) {
	tmpDir, err := ioutil.TempDir("", "")
	defer os.RemoveAll(tmpDir)
	if err != nil {
		return manifest.Manifest{}, "", err
	}

	manifestPath := filepath.Join(tmpDir, "manifest.yml")
	Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: tmpDir}, "create-app-manifest", appName, "-p", manifestPath)).Should(Exit(0))

	manifestContents, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		return manifest.Manifest{}, "", err
	}

	var appsManifest manifest.Manifest
	err = yaml.Unmarshal(manifestContents, &appsManifest)
	if err != nil {
		return manifest.Manifest{}, "", err
	}

	return appsManifest, string(manifestContents), nil
}

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

			domainName string
		)

		BeforeEach(func() {
			orgName = helpers.NewOrgName()
			spaceName = helpers.NewSpaceName()

			helpers.SetupCF(orgName, spaceName)
			userName, _ = helpers.GetCredentials()
			domainName = helpers.DefaultSharedDomain()
		})

		AfterEach(func() {
			helpers.QuickDeleteOrg(orgName)
		})

		When("the app does not exist", func() {
			It("displays a usage error", func() {
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: tempDir}, "create-app-manifest", appName)
				Eventually(session).Should(Say(`Creating an app manifest from current settings of app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("App '%s' not found", appName))

				Eventually(session).Should(Exit(1))
			})
		})

		When("the app exists", func() {
			When("the app does not have routes", func() {
				BeforeEach(func() {
					helpers.WithHelloWorldApp(func(appDir string) {
						Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "push", appName, "--no-route")).Should(Exit(0))
					})
				})

				It("creates the manifest with no-route set to true", func() {
					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: tempDir}, "create-app-manifest", appName)
					Eventually(session).Should(Say(`Creating an app manifest from current settings of app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
					Eventually(session).Should(Say("OK"))
					expectedFilePath := helpers.ConvertPathToRegularExpression(fmt.Sprintf(".%s%s_manifest.yml", string(os.PathSeparator), appName))
					Eventually(session).Should(Say("Manifest file created successfully at %s", expectedFilePath))

					expectedFile := fmt.Sprintf(`applications:
- name: %s
  disk_quota: 1G
  instances: 1
  memory: 32M
  no-route: true
  stack: cflinuxfs\d
`, appName)

					createdFile, err := ioutil.ReadFile(manifestFilePath)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(createdFile)).To(MatchRegexp(expectedFile))

					Eventually(session).Should(Exit(0))
				})
			})

			When("the app has routes", func() {
				BeforeEach(func() {
					helpers.WithHelloWorldApp(func(appDir string) {
						Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "push", appName)).Should(Exit(0))
					})
				})

				It("creates the manifest", func() {
					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: tempDir}, "create-app-manifest", appName)
					Eventually(session).Should(Say(`Creating an app manifest from current settings of app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
					Eventually(session).Should(Say("OK"))
					expectedFilePath := helpers.ConvertPathToRegularExpression(fmt.Sprintf(".%s%s_manifest.yml", string(os.PathSeparator), appName))
					Eventually(session).Should(Say("Manifest file created successfully at %s", expectedFilePath))

					expectedFile := fmt.Sprintf(`applications:
- name: %s
  disk_quota: 1G
  instances: 1
  memory: 32M
  routes:
  - route: %s.%s
  stack: cflinuxfs\d
`, appName, strings.ToLower(appName), domainName)

					createdFile, err := ioutil.ReadFile(manifestFilePath)
					Expect(err).ToNot(HaveOccurred())
					Expect(string(createdFile)).To(MatchRegexp(expectedFile))

					Eventually(session).Should(Exit(0))
				})

				When("the -p flag is provided", func() {
					When("the specified file is a directory", func() {
						It("displays a file creation error", func() {
							session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: tempDir}, "create-app-manifest", appName, "-p", tempDir)
							Eventually(session).Should(Say(`Creating an app manifest from current settings of app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
							Eventually(session).Should(Say("FAILED"))
							Eventually(session.Err).Should(Say("Error creating manifest file: open %s: is a directory", helpers.ConvertPathToRegularExpression(tempDir)))

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
							Eventually(session).Should(Say("OK"))
							Eventually(session).Should(Say("Manifest file created successfully at %s", helpers.ConvertPathToRegularExpression(newFile)))

							expectedFile := fmt.Sprintf(`applications:
- name: %s
  disk_quota: 1G
  instances: 1
  memory: 32M
  routes:
  - route: %s.%s
  stack: cflinuxfs\d
`, appName, strings.ToLower(appName), domainName)

							createdFile, err := ioutil.ReadFile(newFile)
							Expect(err).ToNot(HaveOccurred())
							Expect(string(createdFile)).To(MatchRegexp(expectedFile))

							Eventually(session).Should(Exit(0))
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
							Eventually(session).Should(Say("OK"))
							Eventually(session).Should(Say("Manifest file created successfully at %s", helpers.ConvertPathToRegularExpression(existingFile)))

							expectedFile := fmt.Sprintf(`applications:
- name: %s
  disk_quota: 1G
  instances: 1
  memory: 32M
  routes:
  - route: %s.%s
  stack: cflinuxfs\d
`, appName, strings.ToLower(appName), domainName)

							createdFile, err := ioutil.ReadFile(existingFile)
							Expect(err).ToNot(HaveOccurred())
							Expect(string(createdFile)).To(MatchRegexp(expectedFile))

							Eventually(session).Should(Exit(0))
						})
					})
				})
			})
		})

		When("app was created with docker image", func() {
			BeforeEach(func() {
				Eventually(helpers.CF("push", appName, "-o", DockerImage)).Should(Exit(0))
			})

			It("creates the manifest", func() {
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: tempDir}, "create-app-manifest", appName, "-v")
				Eventually(session).Should(Say(`Creating an app manifest from current settings of app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
				Eventually(session).Should(Say("OK"))
				expectedFilePath := helpers.ConvertPathToRegularExpression(fmt.Sprintf(".%s%s_manifest.yml", string(os.PathSeparator), appName))
				Eventually(session).Should(Say("Manifest file created successfully at %s", expectedFilePath))
				Eventually(session).Should(Exit(0))

				expectedFile := fmt.Sprintf(`applications:
- name: %s
  disk_quota: 1G
  docker:
    image: %s
  instances: 1
  memory: 32M
  routes:
  - route: %s.%s
  stack: cflinuxfs\d
`, appName, DockerImage, strings.ToLower(appName), domainName)

				createdFile, err := ioutil.ReadFile(manifestFilePath)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(createdFile)).To(MatchRegexp(expectedFile))
			})
		})

		When("the API supports docker credentials", func() {
			var oldDockerPassword string

			BeforeEach(func() {
				oldDockerPassword = os.Getenv("CF_DOCKER_PASSWORD")
				Expect(os.Setenv("CF_DOCKER_PASSWORD", "my-docker-password")).To(Succeed())

				Eventually(helpers.CF("push", appName, "-o", DockerImage, "--docker-username", "some-docker-username")).Should(Exit())
			})

			AfterEach(func() {
				Expect(os.Setenv("CF_DOCKER_PASSWORD", oldDockerPassword)).To(Succeed())
			})

			It("creates the manifest", func() {
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: tempDir}, "create-app-manifest", appName, "-v")
				Eventually(session).Should(Say(`Creating an app manifest from current settings of app %s in org %s / space %s as %s\.\.\.`, appName, orgName, spaceName, userName))
				Eventually(session).Should(Say("OK"))
				expectedFilePath := helpers.ConvertPathToRegularExpression(fmt.Sprintf(".%s%s_manifest.yml", string(os.PathSeparator), appName))
				Eventually(session).Should(Say("Manifest file created successfully at %s", expectedFilePath))

				expectedFile := fmt.Sprintf(`applications:
- name: %s
  disk_quota: 1G
  docker:
    image: %s
    username: some-docker-username
  instances: 1
  memory: 32M
  routes:
  - route: %s.%s
  stack: cflinuxfs\d
`, appName, DockerImage, strings.ToLower(appName), domainName)

				createdFile, err := ioutil.ReadFile(manifestFilePath)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(createdFile)).To(MatchRegexp(expectedFile))

				Eventually(session).Should(Exit(0))
			})
		})

		When("app has no hostname", func() {
			var domain helpers.Domain

			BeforeEach(func() {
				domain = helpers.NewDomain(orgName, helpers.NewDomainName(""))
				domain.Create()

				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "push", appName, "--no-start", "-b", "staticfile_buildpack", "--no-hostname", "-d", domain.Name)).Should(Exit(0))
				})
			})

			It("contains routes without hostnames", func() {
				appManifest, _, err := createManifest(appName)
				Expect(err).ToNot(HaveOccurred())

				Expect(appManifest.Applications).To(HaveLen(1))
				Expect(appManifest.Applications[0].Routes).To(HaveLen(1))
				Expect(appManifest.Applications[0].Routes[0]).To(Equal(domain.Name))
			})
		})

		When("the app has a buildpack", func() {
			BeforeEach(func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "push", appName, "--no-start", "-b", "staticfile_buildpack")).Should(Exit(0))
				})
			})

			It("returns a manifest with one buildpack under buildpacks", func() {
				appManifest, rawManifest, err := createManifest(appName)
				Expect(err).ToNot(HaveOccurred())

				Expect(appManifest.Applications).To(HaveLen(1))
				Expect(appManifest.Applications[0].Buildpacks).To(ConsistOf("staticfile_buildpack"), fmt.Sprintf("Manifest should have a staticfile_buildpack:\n%s\n", rawManifest))
			})
		})

		When("the app has multiple buildpacks", func() {
			BeforeEach(func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "push", appName, "--no-start", "-b", "ruby_buildpack", "-b", "staticfile_buildpack")).Should(Exit(0))
				})
			})

			It("returns a manifest with multiple buildpacks", func() {
				appManifest, rawManifest, err := createManifest(appName)
				Expect(err).ToNot(HaveOccurred())

				Expect(appManifest.Applications).To(HaveLen(1))
				Expect(appManifest.Applications[0].Buildpacks).To(ConsistOf("ruby_buildpack", "staticfile_buildpack"), fmt.Sprintf("Manifest should have ruby and staticfile:\n%s\n", rawManifest))
			})
		})
	})
})
