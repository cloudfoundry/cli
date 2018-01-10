package push

import (
	"io/ioutil"
	"os"
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

	Context("when the .cfignore file is in the app source directory", func() {
		Context("when the .cfignore file doesn't exclude any files", func() {
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

					session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, firstApp)

					Eventually(session).Should(Say("50[0-9] B / 50[0-9] B"))
					Eventually(session).Should(Exit(0))
				})
			})
		})

		Context("when the .cfignore file excludes some files", func() {
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

						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, firstApp)

						Eventually(session).Should(Say("28[0-9] B / 28[0-9] B"))
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

						session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, firstApp)

						Eventually(session).Should(Say("28[0-9] B / 28[0-9] B"))
						Eventually(session).Should(Exit(0))
					})
				})
			})
		})
	})

	Context("when the CF_TRACE file is in the app source directory", func() {
		var previousEnv string

		AfterEach(func() {
			err := os.Setenv("CF_TRACE", previousEnv)
			Expect(err).ToNot(HaveOccurred())
		})

		It("does not push it", func() {
			helpers.WithHelloWorldApp(func(dir string) {
				traceFilePath := filepath.Join(dir, "i-am-trace.txt")
				err := ioutil.WriteFile(traceFilePath, nil, 0666)
				Expect(err).ToNot(HaveOccurred())

				previousEnv = os.Getenv("CF_TRACE")
				err = os.Setenv("CF_TRACE", traceFilePath)
				Expect(err).ToNot(HaveOccurred())

				session := helpers.CustomCF(helpers.CFEnv{WorkingDirectory: dir}, PushCommandName, firstApp)

				Eventually(session).Should(Say("28[0-9] B / 28[0-9] B"))
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
