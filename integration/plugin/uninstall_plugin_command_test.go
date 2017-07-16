package plugin

import (
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/util/generic"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("uninstall-plugin command", func() {
	Describe("help", func() {
		Context("when --help flag is provided", func() {
			It("displays command usage to output", func() {
				session := helpers.CF("uninstall-plugin", "--help")
				Eventually(session.Out).Should(Say("NAME:"))
				Eventually(session.Out).Should(Say("uninstall-plugin - Uninstall CLI plugin"))
				Eventually(session.Out).Should(Say("USAGE:"))
				Eventually(session.Out).Should(Say("cf uninstall-plugin PLUGIN-NAME"))
				Eventually(session.Out).Should(Say("SEE ALSO:"))
				Eventually(session.Out).Should(Say("plugins"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("when the plugin is not installed", func() {
		It("informs the user that no such plugin is present and exits 1", func() {
			session := helpers.CF("uninstall-plugin", "bananarama")
			Eventually(session.Err).Should(Say("Plugin bananarama does not exist\\."))
			Eventually(session).Should(Exit(1))
		})
	})

	Context("when the plugin is installed", func() {
		BeforeEach(func() {
			helpers.InstallConfigurablePlugin("banana-plugin-name-1", "2.0.1", []helpers.PluginCommand{
				{Name: "banana-command-1", Help: "banana-command-1"},
			})
			helpers.InstallConfigurablePlugin("banana-plugin-name-2", "1.4.3", []helpers.PluginCommand{
				{Name: "banana-command-2", Help: "banana-command-2"},
			})
		})

		Context("when no errors are encountered", func() {
			It("does not list the plugin after it is uninstalled", func() {
				session := helpers.CF("uninstall-plugin", "banana-plugin-name-1")
				Eventually(session.Out).Should(Say("Uninstalling plugin banana-plugin-name-1\\.\\.\\."))
				// Test that RPC works
				Eventually(session.Out).Should(Say("[0-9]{1,5} CLI-MESSAGE-UNINSTALL"))
				Eventually(session.Out).Should(Say("OK"))
				Eventually(session.Out).Should(Say("Plugin banana-plugin-name-1 2\\.0\\.1 successfully uninstalled\\."))
				Eventually(session).Should(Exit(0))

				session = helpers.CF("plugins")
				Consistently(session.Out).ShouldNot(Say("banana-plugin-name-1"))
				Eventually(session.Out).Should(Say("banana-plugin-name-2"))
				Eventually(session).Should(Exit(0))
			})

			It("matches the plugin name case insensitive", func() {
				session := helpers.CF("uninstall-plugin", "BaNaNa-PlUgIn-NaMe-1")
				Eventually(session.Out).Should(Say("Uninstalling plugin banana-plugin-name-1\\.\\.\\."))
				Eventually(session.Out).Should(Say("OK"))
				Eventually(session.Out).Should(Say("Plugin banana-plugin-name-1 2\\.0\\.1 successfully uninstalled\\."))
				Eventually(session).Should(Exit(0))
			})
		})

		Context("when the plugin encounters an error during cleanup", func() {
			BeforeEach(func() {
				helpers.InstallConfigurablePluginFailsUninstall("failing-plugin", "2.0.1", []helpers.PluginCommand{
					{Name: "failing-command-1", Help: "failing-command-1"},
				})
			})

			It("exits with an error but still uninstalls the plugin", func() {
				session := helpers.CF("uninstall-plugin", "failing-plugin")
				Eventually(session.Out).Should(Say("Uninstalling plugin failing-plugin\\.\\.\\."))
				Eventually(session.Err).Should(Say("I'm failing...I'm failing..."))
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("The plugin's uninstall method returned an unexpected error\\."))
				Eventually(session.Err).Should(Say("The plugin uninstall will proceed\\. Contact the plugin author if you need help\\."))
				Eventually(session.Err).Should(Say("exit status 1"))
				Eventually(session).Should(Exit(1))

				binaryPath := generic.ExecutableFilename(
					filepath.Join(homeDir, ".cf", "plugins", "failing-plugin"))
				_, err := os.Stat(binaryPath)
				Expect(os.IsNotExist(err)).To(BeTrue())

				session = helpers.CF("plugins")
				Eventually(session.Out).Should(Say("banana-plugin-name-1"))
				Eventually(session.Out).Should(Say("banana-plugin-name-2"))
				Consistently(session.Out).ShouldNot(Say("failing-plugin"))
				Eventually(session).Should(Exit(0))
			})
		})

		Context("when the plugin binary has been deleted", func() {
			BeforeEach(func() {
				helpers.InstallConfigurablePlugin(
					"banana-plugin-name-1",
					"2.0.1",
					[]helpers.PluginCommand{
						{
							Name: "banana-command-1",
							Help: "banana-command-1"},
					})

				binaryPath := generic.ExecutableFilename(
					filepath.Join(homeDir, ".cf", "plugins", "banana-plugin-name-1"))
				Expect(os.Remove(binaryPath)).ToNot(HaveOccurred())
			})

			It("uninstalls the plugin with no warning or error and exits 0", func() {
				session := helpers.CF("uninstall-plugin", "banana-plugin-name-1")
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
