package isolated

import (
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("create-package with .cfignore", func() {
	var (
		orgName   string
		spaceName string
		appName   string
	)

	BeforeEach(func() {
		orgName = helpers.NewOrgName()
		spaceName = helpers.NewSpaceName()
		appName = helpers.PrefixedRandomName("app")
		helpers.SetupCF(orgName, spaceName)

	})

	AfterEach(func() {
		helpers.QuickDeleteOrg(orgName)
	})

	When(".cfignore file exists", func() {
		When("the .cfignore file doesn't exclude any files", func() {
			It("pushes all the files except .cfignore and files ignored by default", func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					file1 := filepath.Join(appDir, "file1")
					err := os.WriteFile(file1, nil, 0666)
					Expect(err).ToNot(HaveOccurred())

					file2 := filepath.Join(appDir, "file2")
					err = os.WriteFile(file2, nil, 0666)
					Expect(err).ToNot(HaveOccurred())

					cfIgnoreFilePath := filepath.Join(appDir, ".cfignore")
					err = os.WriteFile(cfIgnoreFilePath, nil, 0666)
					Expect(err).ToNot(HaveOccurred())

					darcsFile := filepath.Join(appDir, "_darcs")
					err = os.WriteFile(darcsFile, nil, 0666)
					Expect(err).ToNot(HaveOccurred())

					dsFile := filepath.Join(appDir, ".DS_Store")
					err = os.WriteFile(dsFile, nil, 0666)
					Expect(err).ToNot(HaveOccurred())

					gitFile := filepath.Join(appDir, ".git")
					err = os.WriteFile(gitFile, nil, 0666)
					Expect(err).ToNot(HaveOccurred())

					gitIgnoreFile := filepath.Join(appDir, ".gitignore")
					err = os.WriteFile(gitIgnoreFile, nil, 0666)
					Expect(err).ToNot(HaveOccurred())

					hgFile := filepath.Join(appDir, ".hg")
					err = os.WriteFile(hgFile, nil, 0666)
					Expect(err).ToNot(HaveOccurred())

					manifestFile := filepath.Join(appDir, "manifest.yml")
					err = os.WriteFile(manifestFile, nil, 0666)
					Expect(err).ToNot(HaveOccurred())

					svnFile := filepath.Join(appDir, ".svn")
					err = os.WriteFile(svnFile, nil, 0666)
					Expect(err).ToNot(HaveOccurred())

					Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "create-app", appName)).Should(Exit(0))
					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "create-package", appName)

					Eventually(session).Should(Exit(0))
					helpers.VerifyAppPackageContentsV3(appName, "file1", "file2", "Staticfile", "index.html")
				})
			})
		})

		When("the .cfignore file excludes some files", func() {
			When("pushing from the current directory", func() {
				It("does not push those files", func() {
					helpers.WithHelloWorldApp(func(appDir string) {
						file1 := filepath.Join(appDir, "file1")
						err := os.WriteFile(file1, nil, 0666)
						Expect(err).ToNot(HaveOccurred())

						file2 := filepath.Join(appDir, "file2")
						err = os.WriteFile(file2, nil, 0666)
						Expect(err).ToNot(HaveOccurred())

						cfIgnoreFilePath := filepath.Join(appDir, ".cfignore")
						err = os.WriteFile(cfIgnoreFilePath, []byte("file*"), 0666)
						Expect(err).ToNot(HaveOccurred())

						Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "create-app", appName)).Should(Exit(0))
						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "create-package", appName)

						Eventually(session).Should(Exit(0))
						helpers.VerifyAppPackageContentsV3(appName, "Staticfile", "index.html")
					})
				})
			})

			When("pushing from a different directory", func() {
				It("does not push those files", func() {
					helpers.WithHelloWorldApp(func(appDir string) {
						file1 := filepath.Join(appDir, "file1")
						err := os.WriteFile(file1, nil, 0666)
						Expect(err).ToNot(HaveOccurred())

						file2 := filepath.Join(appDir, "file2")
						err = os.WriteFile(file2, nil, 0666)
						Expect(err).ToNot(HaveOccurred())

						cfIgnoreFilePath := filepath.Join(appDir, ".cfignore")
						err = os.WriteFile(cfIgnoreFilePath, []byte("file*"), 0666)
						Expect(err).ToNot(HaveOccurred())

						Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "create-app", appName)).Should(Exit(0))
						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "create-package", appName, "-p", appDir)

						Eventually(session).Should(Exit(0))
						helpers.VerifyAppPackageContentsV3(appName, "Staticfile", "index.html")
					})
				})
			})

			When("pushing a zip file", func() {
				var archive string
				BeforeEach(func() {
					helpers.WithHelloWorldApp(func(appDir string) {
						file1 := filepath.Join(appDir, "file1")
						err := os.WriteFile(file1, nil, 0666)
						Expect(err).ToNot(HaveOccurred())

						file2 := filepath.Join(appDir, "file2")
						err = os.WriteFile(file2, nil, 0666)
						Expect(err).ToNot(HaveOccurred())

						cfIgnoreFilePath := filepath.Join(appDir, ".cfignore")
						err = os.WriteFile(cfIgnoreFilePath, []byte("file*"), 0666)
						Expect(err).ToNot(HaveOccurred())

						tmpfile, err := os.CreateTemp("", "push-archive-integration")
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

				It("does not push those files", func() {
					Eventually(helpers.CF("create-app", appName)).Should(Exit(0))
					session := helpers.CF("create-package", appName, "-p", archive)

					Eventually(session).Should(Exit(0))
					helpers.VerifyAppPackageContentsV3(appName, "Staticfile", "index.html")
				})
			})
		})

		When("the CF_TRACE file is in the app source directory", func() {
			var previousEnv string

			AfterEach(func() {
				err := os.Setenv("CF_TRACE", previousEnv)
				Expect(err).ToNot(HaveOccurred())
			})

			It("does not push it", func() {
				helpers.WithHelloWorldApp(func(appDir string) {
					traceFilePath := filepath.Join(appDir, "i-am-trace.txt")
					err := os.WriteFile(traceFilePath, nil, 0666)
					Expect(err).ToNot(HaveOccurred())

					previousEnv = os.Getenv("CF_TRACE")
					err = os.Setenv("CF_TRACE", traceFilePath)
					Expect(err).ToNot(HaveOccurred())

					Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "create-app", appName)).Should(Exit(0))
					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "create-package", appName)
					Eventually(session).Should(Exit(0))
					helpers.VerifyAppPackageContentsV3(appName, "Staticfile", "index.html")
				})
			})
		})
	})

	When(".cfignore file does not exist", func() {
		It("pushes all the files except for the files ignored by default", func() {
			helpers.WithHelloWorldApp(func(appDir string) {
				file1 := filepath.Join(appDir, "file1")
				err := os.WriteFile(file1, nil, 0666)
				Expect(err).ToNot(HaveOccurred())

				file2 := filepath.Join(appDir, "file2")
				err = os.WriteFile(file2, nil, 0666)
				Expect(err).ToNot(HaveOccurred())

				darcsFile := filepath.Join(appDir, "_darcs")
				err = os.WriteFile(darcsFile, nil, 0666)
				Expect(err).ToNot(HaveOccurred())

				dsFile := filepath.Join(appDir, ".DS_Store")
				err = os.WriteFile(dsFile, nil, 0666)
				Expect(err).ToNot(HaveOccurred())

				gitFile := filepath.Join(appDir, ".git")
				err = os.WriteFile(gitFile, nil, 0666)
				Expect(err).ToNot(HaveOccurred())

				gitIgnoreFile := filepath.Join(appDir, ".gitignore")
				err = os.WriteFile(gitIgnoreFile, nil, 0666)
				Expect(err).ToNot(HaveOccurred())

				hgFile := filepath.Join(appDir, ".hg")
				err = os.WriteFile(hgFile, nil, 0666)
				Expect(err).ToNot(HaveOccurred())

				manifestFile := filepath.Join(appDir, "manifest.yml")
				err = os.WriteFile(manifestFile, nil, 0666)
				Expect(err).ToNot(HaveOccurred())

				svnFile := filepath.Join(appDir, ".svn")
				err = os.WriteFile(svnFile, nil, 0666)
				Expect(err).ToNot(HaveOccurred())

				Eventually(helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "create-app", appName)).Should(Exit(0))
				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: appDir}, "create-package", appName)

				Eventually(session).Should(Exit(0))
				helpers.VerifyAppPackageContentsV3(appName, "file1", "file2", "Staticfile", "index.html")
			})
		})
	})
})
