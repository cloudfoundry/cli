package push

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccversion"
	"code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("push with symlinked resources", func() {
	var (
		appName string
	)

	BeforeEach(func() {
		helpers.SkipIfVersionLessThan(ccversion.MinVersionSymlinkedFilesV2)
		appName = helpers.NewAppName()
	})

	When("pushing a directory", func() {
		When("the directory contains a symlink to a file in the directory", func() {
			When("the file exists", func() {
				It("should push the symlink", func() {
					helpers.WithHelloWorldApp(func(dir string) {
						targetFile := filepath.Join(dir, "targetFile")
						Expect(ioutil.WriteFile(targetFile, []byte("foo bar baz"), 0777)).ToNot(HaveOccurred())
						relativePath, err := filepath.Rel(dir, targetFile)
						Expect(err).ToNot(HaveOccurred())

						err = os.Symlink(relativePath, filepath.Join(dir, "symlinkFile"))
						Expect(err).ToNot(HaveOccurred())

						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--no-start")

						Eventually(session).Should(Exit(0))
					})

					helpers.VerifyAppPackageContentsV2(appName, "symlinkFile", "targetFile", "Staticfile", "index.html")
				})
			})

			When("the file doesn't exists", func() {
				It("should push the symlink", func() {
					helpers.WithHelloWorldApp(func(dir string) {
						tempFile, err := ioutil.TempFile(dir, "tempFile")
						Expect(err).ToNot(HaveOccurred())
						tempFile.Close()
						relativePath, err := filepath.Rel(dir, tempFile.Name())
						Expect(err).ToNot(HaveOccurred())

						err = os.Symlink(relativePath, filepath.Join(dir, "symlinkFile"))
						Expect(err).ToNot(HaveOccurred())
						Expect(os.Remove(tempFile.Name()))

						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--no-start")

						Eventually(session).Should(Exit(0))
					})

					helpers.VerifyAppPackageContentsV2(appName, "symlinkFile", "Staticfile", "index.html")
				})
			})
		})

		When("the directory contains a symlink to subdirectory in the directory", func() {
			It("should push the symlink", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					targetDir, err := ioutil.TempDir(dir, "target-dir")
					Expect(err).ToNot(HaveOccurred())
					relativePath, err := filepath.Rel(dir, targetDir)
					Expect(err).ToNot(HaveOccurred())

					err = os.Symlink(relativePath, filepath.Join(dir, "symlinkFile"))
					Expect(err).ToNot(HaveOccurred())
					Expect(os.RemoveAll(targetDir)).ToNot(HaveOccurred())

					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, appName, "--no-start")

					Eventually(session).Should(Exit(0))
				})

				helpers.VerifyAppPackageContentsV2(appName, "symlinkFile", "Staticfile", "index.html")
			})
		})
	})

	When("pushing an archive", func() {
		var archive string

		AfterEach(func() {
			Expect(os.RemoveAll(archive)).ToNot(HaveOccurred())
		})

		When("the archive contains a symlink to a file in the directory", func() {
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
						Expect(err).ToNot(HaveOccurred())

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
				helpers.VerifyAppPackageContentsV2(appName, "symlinkFile", "targetFile", "Staticfile", "index.html")
			})
		})
	})
})
