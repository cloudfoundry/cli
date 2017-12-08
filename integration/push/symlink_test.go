package push

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("push with symlink path", func() {
	var (
		appName       string
		runningDir    string
		symlinkedPath string
	)

	BeforeEach(func() {
		appName = helpers.NewAppName()

		var err error
		runningDir, err = ioutil.TempDir("", "push-with-symlink")
		Expect(err).ToNot(HaveOccurred())
		symlinkedPath = filepath.Join(runningDir, "symlink-dir")
	})

	AfterEach(func() {
		Expect(os.RemoveAll(runningDir)).ToNot(HaveOccurred())
	})

	Context("push with flag options", func() {
		Context("when pushing from a symlinked current directory", func() {
			It("should push with the absolute path of the app", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					Expect(os.Symlink(dir, symlinkedPath)).ToNot(HaveOccurred())

					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: symlinkedPath}, PushCommandName, appName)
					Eventually(session).Should(Say("path:\\s+(\\/private)?%s", regexp.QuoteMeta(dir)))
					Eventually(session).Should(Exit(0))
				})
			})
		})

		Context("when pushing a symlinked path with the '-p' flag", func() {
			It("should push with the absolute path of the app", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					Expect(os.Symlink(dir, symlinkedPath)).ToNot(HaveOccurred())

					session := helpers.CF(PushCommandName, appName, "-p", symlinkedPath)
					Eventually(session).Should(Say("path:\\s+(\\/private)?%s", regexp.QuoteMeta(dir)))
					Eventually(session).Should(Exit(0))
				})
			})
		})

		Context("when pushing an symlinked archive with the '-p' flag", func() {
			var archive string

			BeforeEach(func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					tmpfile, err := ioutil.TempFile("", "push-archive-integration")
					Expect(err).ToNot(HaveOccurred())
					archive = tmpfile.Name()
					Expect(tmpfile.Close()).ToNot(HaveOccurred())

					err = helpers.Zipit(appDir, archive, "")
					Expect(err).ToNot(HaveOccurred())
				})
			})

			AfterEach(func() {
				Expect(os.RemoveAll(archive)).ToNot(HaveOccurred())
			})

			It("should push with the absolute path of the archive", func() {
				Expect(os.Symlink(archive, symlinkedPath)).ToNot(HaveOccurred())

				session := helpers.CF(PushCommandName, appName, "-p", symlinkedPath)
				Eventually(session).Should(Say("path:\\s+(\\/private)?%s", regexp.QuoteMeta(archive)))
				Eventually(session).Should(Exit(0))
			})
		})

		Context("push with a single app manifest", func() {
			Context("when the path property is a symlinked path", func() {
				It("should push with the absolute path of the app", func() {
					helpers.WithHelloWorldApp(func(dir string) {
						Expect(os.Symlink(dir, symlinkedPath)).ToNot(HaveOccurred())

						helpers.WriteManifest(filepath.Join(runningDir, "manifest.yml"), map[string]interface{}{
							"applications": []map[string]string{
								{
									"name": appName,
									"path": symlinkedPath,
								},
							},
						})

						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: runningDir}, PushCommandName)
						Eventually(session).Should(Say("path:\\s+(\\/private)?%s", regexp.QuoteMeta(dir)))
						Eventually(session).Should(Exit(0))
					})
				})
			})
		})
	})
})

var _ = Describe("push with symlinked resources", func() {
	var (
		appName string
	)

	BeforeEach(func() {
		appName = helpers.NewAppName()
	})

	Context("when pushing a directory", func() {
		Context("when the directory contains a symlink to a file in the directory", func() {
			Context("when the file exists", func() {
				It("should push the symlink", func() {
					helpers.WithHelloWorldApp(func(dir string) {
						targetFile := filepath.Join(dir, "targetFile")
						Expect(ioutil.WriteFile(targetFile, []byte("foo bar baz"), 0777)).ToNot(HaveOccurred())
						relativePath, err := filepath.Rel(dir, targetFile)

						err = os.Symlink(relativePath, filepath.Join(dir, "symlinkFile"))
						Expect(err).ToNot(HaveOccurred())

						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--no-start")

						Eventually(session).Should(Exit(0))
					})

					helpers.VerifyAppPackageContents(appName, "symlinkFile", "targetFile", "Staticfile", "index.html")
				})
			})

			Context("when the file doesn't exists", func() {
				It("should push the symlink", func() {
					helpers.WithHelloWorldApp(func(dir string) {
						tempFile, err := ioutil.TempFile(dir, "tempFile")
						Expect(err).ToNot(HaveOccurred())
						tempFile.Close()
						relativePath, err := filepath.Rel(dir, tempFile.Name())

						err = os.Symlink(relativePath, filepath.Join(dir, "symlinkFile"))
						Expect(err).ToNot(HaveOccurred())
						Expect(os.Remove(tempFile.Name()))

						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--no-start")

						Eventually(session).Should(Exit(0))
					})

					helpers.VerifyAppPackageContents(appName, "symlinkFile", "Staticfile", "index.html")
				})
			})
		})

		Context("when the directory contains a symlink to subdirectory in the directory", func() {
			It("should push the symlink", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					targetDir, err := ioutil.TempDir(dir, "target-dir")
					Expect(err).ToNot(HaveOccurred())
					relativePath, err := filepath.Rel(dir, targetDir)

					err = os.Symlink(relativePath, filepath.Join(dir, "symlinkFile"))
					Expect(err).ToNot(HaveOccurred())
					Expect(os.RemoveAll(targetDir)).ToNot(HaveOccurred())

					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--no-start")

					Eventually(session).Should(Exit(0))
				})

				helpers.VerifyAppPackageContents(appName, "symlinkFile", "Staticfile", "index.html")
			})
		})

		PContext("when the directory contains a symlink to a file outside the directory", func() {
			var targetPath string

			BeforeEach(func() {
				tmpFile, err := ioutil.TempFile("", "push-symlink-integration-")
				Expect(err).ToNot(HaveOccurred())
				tmpFile.Close()

				targetPath = tmpFile.Name()
			})

			AfterEach(func() {
				Expect(os.Remove(targetPath))
			})

			It("it should fail with an upload invalid error", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					err := os.Symlink(targetPath, filepath.Join(dir, "symlinkFile"))
					Expect(err).ToNot(HaveOccurred())

					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--no-start")

					Eventually(session.Err).Should(Say("The app upload is invalid: Symlink\\(s\\) point outside of root folder"))
					Eventually(session).Should(Exit(1))
				})
			})
		})
	})

	Context("when pushing an archive", func() {
		var archive string

		AfterEach(func() {
			Expect(os.RemoveAll(archive)).ToNot(HaveOccurred())
		})

		Context("when the archive contains a symlink to a file in the directory", func() {
			BeforeEach(func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					helpers.WithHelloWorldApp(func(appDir string) {
						tmpfile, err := ioutil.TempFile("", "push-archive-integration")
						Expect(err).ToNot(HaveOccurred())
						archive = tmpfile.Name()
						Expect(tmpfile.Close()).ToNot(HaveOccurred())

						targetFile := filepath.Join(appDir, "targetFile")
						Expect(ioutil.WriteFile(targetFile, []byte("some random data"), 0777)).ToNot(HaveOccurred())
						relativePath, err := filepath.Rel(appDir, targetFile)

						err = os.Symlink(relativePath, filepath.Join(appDir, "symlinkFile"))
						Expect(err).ToNot(HaveOccurred())

						err = helpers.Zipit(appDir, archive, "")
						Expect(err).ToNot(HaveOccurred())
					})
				})
			})

			It("should push the symlink", func() {
				session := helpers.CF(PushCommandName, appName, "--no-start", "-p", archive)

				Eventually(session).Should(Exit(0))
				helpers.VerifyAppPackageContents(appName, "symlinkFile", "targetFile", "Staticfile", "index.html")
			})
		})
	})
})
