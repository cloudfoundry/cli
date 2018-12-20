package push

import (
	"io/ioutil"
	"path/filepath"

	"code.cloudfoundry.org/cli/integration/helpers"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("ignoring files while gathering resources", func() {
	var (
		firstApp string
	)

	BeforeEach(func() {
		firstApp = helpers.NewAppName()
	})

	When("the .cfignore file is in the app source directory", func() {
		When("the .cfignore file doesn't exclude any files", func() {
			It("pushes all the files", func() {
				helpers.WithHelloWorldApp(func(dir string) {
					file1 := filepath.Join(dir, "file1")
					err := ioutil.WriteFile(file1, nil, 0666)
					Expect(err).ToNot(HaveOccurred())

					file2 := filepath.Join(dir, "file2")
					err = ioutil.WriteFile(file2, nil, 0666)
					Expect(err).ToNot(HaveOccurred())

					cfIgnoreFilePath := filepath.Join(dir, ".cfignore")
					err = ioutil.WriteFile(cfIgnoreFilePath, nil, 0666)
					Expect(err).ToNot(HaveOccurred())

					session := helpers.DebugCustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, firstApp, "--no-start")

					Eventually(session.Err).Should(Say("zipped_file_count=4"))
					Eventually(session).Should(Exit(0))
				})
			})
		})

		When("the .cfignore file excludes some files", func() {
			Context("ignored files are relative paths", func() {
				It("does not push those files", func() {
					helpers.WithHelloWorldApp(func(dir string) {
						file1 := filepath.Join(dir, "file1")
						err := ioutil.WriteFile(file1, nil, 0666)
						Expect(err).ToNot(HaveOccurred())

						file2 := filepath.Join(dir, "file2")
						err = ioutil.WriteFile(file2, nil, 0666)
						Expect(err).ToNot(HaveOccurred())

						cfIgnoreFilePath := filepath.Join(dir, ".cfignore")
						err = ioutil.WriteFile(cfIgnoreFilePath, []byte("file*"), 0666)
						Expect(err).ToNot(HaveOccurred())

						session := helpers.DebugCustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, firstApp, "--no-start")

						Eventually(session.Err).Should(Say("zipped_file_count=2"))
						Eventually(session).Should(Exit(0))
					})
				})
			})

			Context("with absolute paths - where '/' == appDir", func() {
				It("does not push those files", func() {
					helpers.WithHelloWorldApp(func(dir string) {
						file1 := filepath.Join(dir, "file1")
						err := ioutil.WriteFile(file1, nil, 0666)
						Expect(err).ToNot(HaveOccurred())

						file2 := filepath.Join(dir, "file2")
						err = ioutil.WriteFile(file2, nil, 0666)
						Expect(err).ToNot(HaveOccurred())

						cfIgnoreFilePath := filepath.Join(dir, ".cfignore")
						err = ioutil.WriteFile(cfIgnoreFilePath, []byte("/file*"), 0666)
						Expect(err).ToNot(HaveOccurred())

						session := helpers.DebugCustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, firstApp, "--no-start")

						Eventually(session.Err).Should(Say("zipped_file_count=2"))
						Eventually(session).Should(Exit(0))
					})
				})
			})
		})
	})

	When("the CF_TRACE file is in the app source directory", func() {
		It("does not push it", func() {
			helpers.WithHelloWorldApp(func(dir string) {
				traceFilePath := filepath.Join(dir, "i-am-trace.txt")
				err := ioutil.WriteFile(traceFilePath, nil, 0666)
				Expect(err).ToNot(HaveOccurred())

				session := helpers.DebugCustomCF(helpers.CFEnv{
					WorkingDirectory: dir,
					EnvVars: map[string]string{
						"CF_TRACE": traceFilePath,
					},
				},
					PushCommandName, firstApp, "--no-start",
				)

				Eventually(session.Err).Should(Say("zipped_file_count=2"))
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
