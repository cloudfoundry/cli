package plugin_test

import (
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Command", func() {
	var (
		validPluginPath string
		OLD_PLUGINS_DIR string
	)

	BeforeEach(func() {
		OLD_PLUGINS_DIR = os.Getenv("CF_PLUGINS_DIR")

		dir, err := os.Getwd()
		Expect(err).NotTo(HaveOccurred())

		fullDir := filepath.Join(dir, "..", "fixtures", "config", "plugin-config")
		err = os.Setenv("CF_PLUGINS_DIR", fullDir)
		Expect(err).NotTo(HaveOccurred())

		validPluginPath = filepath.Join(fullDir, ".cf", "plugins", "test_1.exe")
	})

	AfterEach(func() {
		err := os.Setenv("CF_PLUGINS_DIR", OLD_PLUGINS_DIR)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe(".ServeCommand", func() {
		It("prints a warning if a plugin does not implement the rpc interface", func() {
			//This would seem like a valid test, but the plugin itself will not compile
		})

		It("Exits with status 1 if it cannot ping the host port passed as an argument", func() {
			args := []string{"0", "0"}
			session, err := Start(exec.Command(validPluginPath, args...), GinkgoWriter, GinkgoWriter)
			Expect(err).ToNot(HaveOccurred())
			Eventually(session, 2).Should(Exit(1))
		})
	})
})
