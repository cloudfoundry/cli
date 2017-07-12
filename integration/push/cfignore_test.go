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

var _ = Describe("ignores files matching patterns in cfignore", func() {
	var (
		firstApp string
	)

	BeforeEach(func() {
		firstApp = helpers.NewAppName()
	})

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

				Eventually(session).Should(Say("502 B / 502 B"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("when the .cfignore file excludes some files", func() {
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

				Eventually(session).Should(Say("288 B / 288 B"))
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
