package push

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("handle path in manifest and flag override", func() {
	var (
		appName string

		secondName string
		tempDir    string
	)

	BeforeEach(func() {
		appName = helpers.NewAppName()
		secondName = helpers.NewAppName()
		var err error
		tempDir, err = ioutil.TempDir("", "simple-manifest-test")
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		Expect(os.RemoveAll(tempDir)).ToNot(HaveOccurred())
	})

	When("manifest specifies paths", func() {
		It("pushes the apps using the relative path to the manifest specified", func() {
			helpers.WithHelloWorldApp(func(dir string) {
				nestedDir := filepath.Join(dir, "nested")
				err := os.Mkdir(nestedDir, os.FileMode(0777))
				if err != nil {
					Expect(err).NotTo(HaveOccurred())
				}
				manifestPath := filepath.Join(nestedDir, "manifest.yml")
				helpers.WriteManifest(manifestPath, map[string]interface{}{
					"applications": []map[string]interface{}{
						{
							"name": appName,
							"path": "..",
						},
						{
							"name": secondName,
							"path": dir,
						},
					},
				})
				session := helpers.CustomCF(
					helpers.CFEnv{
						EnvVars: map[string]string{"CF_LOG_LEVEL": "debug"},
					},
					PushCommandName,
					appName,
					"-f", manifestPath,
				)

				if runtime.GOOS == "windows" {
					// The paths in windows logging have extra escaping that is difficult
					// to match. Instead match on uploading the right number of files.
					Eventually(session.Err).Should(Say("zipped_file_count=3"))
				} else {
					Eventually(session.Err).Should(helpers.SayPath(`msg="creating archive"\s+Path="?%s"?`, dir))
				}
				Eventually(session).Should(Exit(0))
			})
		})

		When("a single path is not valid", func() {
			It("errors", func() {
				expandedTempDir, err := filepath.EvalSymlinks(tempDir)
				Expect(err).NotTo(HaveOccurred())
				manifestPath := filepath.Join(tempDir, "manifest.yml")
				helpers.WriteManifest(manifestPath, map[string]interface{}{
					"applications": []map[string]interface{}{
						{
							"name": appName,
							"path": "potato",
						},
						{
							"name": secondName,
							"path": "/baboaboaboaobao/foo",
						},
					},
				})
				session := helpers.CF(PushCommandName, appName, "-f", manifestPath)
				Eventually(session.Err).Should(helpers.SayPath("File not found locally, make sure the file exists at given path %s", filepath.Join(expandedTempDir, "potato")))
				Eventually(session).Should(Exit(1))
			})
		})
	})

	When("manifest does not specify a path and there is no flag override", func() {
		When("no droplet or docker specified", func() {
			It("defaults to the current working directory", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					workingDir, err := os.Getwd()
					Expect(err).ToNot(HaveOccurred())

					err = os.Chdir(dir)
					Expect(err).ToNot(HaveOccurred())

					nestedDir := filepath.Join(dir, "nested")
					err = os.Mkdir(nestedDir, os.FileMode(0777))
					if err != nil {
						Expect(err).NotTo(HaveOccurred())
					}
					manifestPath := filepath.Join(nestedDir, "manifest.yml")
					helpers.WriteManifest(manifestPath, map[string]interface{}{
						"applications": []map[string]interface{}{
							{
								"name": appName,
							},
						},
					})
					session := helpers.CustomCF(
						helpers.CFEnv{
							EnvVars: map[string]string{"CF_LOG_LEVEL": "debug"},
						},
						PushCommandName,
						appName,
						"-f", manifestPath,
					)

					if runtime.GOOS == "windows" {
						// The paths in windows logging have extra escaping that is difficult
						// to match. Instead match on uploading the right number of files.
						Eventually(session.Err).Should(Say("zipped_file_count=3"))
					} else {
						Eventually(session.Err).Should(helpers.SayPath(`msg="creating archive"\s+Path="?%s"?`, dir))
					}
					Eventually(session).Should(Exit(0))

					err = os.Chdir(workingDir)
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})

		When("docker is specified", func() {
			It("it uses the docker config", func() {
				manifestPath := filepath.Join(tempDir, "manifest.yml")
				helpers.WriteManifest(manifestPath, map[string]interface{}{
					"applications": []map[string]interface{}{
						{
							"name": appName,
							"docker": map[string]string{
								"image":    "bad-docker-image",
								"username": "bad-docker-username",
							},
						},
					},
				})
				session := helpers.CustomCF(
					helpers.CFEnv{
						EnvVars: map[string]string{"CF_LOG_LEVEL": "debug", "CF_DOCKER_PASSWORD": "bad-docker-password"},
					},
					PushCommandName,
					appName,
					"-f", manifestPath,
				)

				Eventually(session).Should(Say("docker"))
				Eventually(session.Err).Should(Say("staging failed"))
				Eventually(session).Should(Exit(1))
			})
		})
	})

	When("the -p flag is provided", func() {
		var (
			appName string
		)

		BeforeEach(func() {
			appName = helpers.PrefixedRandomName("app")
		})

		When("the path is a directory", func() {
			When("the directory contains files", func() {
				It("pushes the app from the directory", func() {
					helpers.WithHelloWorldApp(func(appDir string) {
						session := helpers.CF(PushCommandName, appName, "-p", appDir)
						Eventually(session).Should(Say(`name:\s+%s`, appName))
						Eventually(session).Should(Say(`requested state:\s+started`))
						Eventually(session).Should(Exit(0))
					})
				})

				When("The manifest is in a different directory than the app's source", func() {
					It("pushes the app with a relative path to the app directory", func() {
						manifestDir := helpers.TempDirAbsolutePath("", "manifest-dir")
						defer os.RemoveAll(manifestDir)

						err := ioutil.WriteFile(
							filepath.Join(manifestDir, "manifest.yml"),
							[]byte(fmt.Sprintf(`---
applications:
  - name: %s`,
								appName)), 0666)
						Expect(err).ToNot(HaveOccurred())

						helpers.WithHelloWorldApp(func(appDir string) {
							err = os.Chdir("/")
							Expect(err).ToNot(HaveOccurred())
							session := helpers.CF(PushCommandName, appName, "-p", path.Join(".", appDir), "-f", path.Join(manifestDir, "manifest.yml"))
							Eventually(session).Should(Say(`name:\s+%s`, appName))
							Eventually(session).Should(Say(`requested state:\s+started`))
							Eventually(session).Should(Exit(0))
						})
					})
				})
			})

			When("the directory is empty", func() {
				var emptyDir string

				BeforeEach(func() {
					sympath, err := ioutil.TempDir("", "integration-push-path-empty")
					Expect(err).ToNot(HaveOccurred())
					emptyDir, err = filepath.EvalSymlinks(sympath)
					Expect(err).ToNot(HaveOccurred())
				})

				AfterEach(func() {
					Expect(os.RemoveAll(emptyDir)).ToNot(HaveOccurred())
				})

				It("returns an error", func() {
					session := helpers.CF(PushCommandName, appName, "-p", emptyDir)
					Eventually(session.Err).Should(helpers.SayPath("No app files found in '%s'", emptyDir))
					Eventually(session).Should(Exit(1))
				})
			})
		})

		When("the path is a zip file", func() {
			Context("pushing a zip file", func() {
				var archive string

				BeforeEach(func() {
					helpers.WithHelloWorldApp(func(appDir string) {
						tmpfile, err := ioutil.TempFile("", "push-archive-integration")
						Expect(err).ToNot(HaveOccurred())
						archive = tmpfile.Name()
						Expect(tmpfile.Close())

						err = helpers.Zipit(appDir, archive, "")
						Expect(err).ToNot(HaveOccurred())
					})
				})

				AfterEach(func() {
					Expect(os.RemoveAll(archive)).ToNot(HaveOccurred())
				})

				It("pushes the app from the zip file", func() {
					session := helpers.CF(PushCommandName, appName, "-p", archive)

					Eventually(session).Should(Say(`name:\s+%s`, appName))
					Eventually(session).Should(Say(`requested state:\s+started`))
					Eventually(session).Should(Exit(0))
				})
			})
		})
	})
})
