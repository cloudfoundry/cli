//go:build !windows
// +build !windows

package plugin

import (
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/util/generic"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("uninstall-plugin command", func() {
	When("the plugin is not executable", func() {
		var binaryPath string

		BeforeEach(func() {
			helpers.InstallConfigurablePlugin("configurable_plugin", "banana-plugin-name-1", "2.0.1", []helpers.PluginCommand{
				{Name: "banana-command-1", Help: "banana-command-1"},
			})

			binaryPath = generic.ExecutableFilename(
				filepath.Join(homeDir, ".cf", "plugins", "banana-plugin-name-1"))
			Expect(os.Chmod(binaryPath, 0644)).ToNot(HaveOccurred())
		})

		It("exits with an error, and does not remove the plugin", func() {
			session := helpers.CF("uninstall-plugin", "banana-plugin-name-1")
			Eventually(session).Should(Say(`Uninstalling plugin banana-plugin-name-1\.\.\.`))
			Eventually(session).Should(Say("FAILED"))
			Eventually(session.Err).Should(Say(`The plugin's uninstall method returned an unexpected error\.`))
			Eventually(session.Err).Should(Say(`The plugin uninstall will proceed\. Contact the plugin author if you need help\.`))
			Eventually(session).Should(Exit(1))

			_, err := os.Stat(binaryPath)
			Expect(os.IsNotExist(err)).To(BeTrue())

			session = helpers.CF("plugins")
			Consistently(session).ShouldNot(Say("banana-plugin-name-1"))
			Eventually(session).Should(Exit(0))
		})
	})
})
