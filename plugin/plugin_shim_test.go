package plugin_test

import (
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Command", func() {
	var (
		validPluginPath = filepath.Join("..", "fixtures", "plugins", "test_1.exe")
	)

	Describe(".Start", func() {
		It("Exits with status 1 if it cannot ping the host port passed as an argument", func() {
			args := []string{"0", "0"}
			session, err := Start(exec.Command(validPluginPath, args...), GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session, 2).Should(Exit(1))
		})
	})
})
