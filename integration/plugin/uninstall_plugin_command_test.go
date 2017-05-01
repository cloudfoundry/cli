package plugin

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cli/integration/helpers"
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
		})

		Context("when the plugin encounters an error during cleanup", func() {
			BeforeEach(func() {
				helpers.InstallConfigurablePluginFailsUninstall("failing-plugin", "2.0.1", []helpers.PluginCommand{
					{Name: "failing-command-1", Help: "failing-command-1"},
				})
			})

			It("exits with an error, and does not remove the plugin", func() {
				session := helpers.CF("uninstall-plugin", "failing-plugin")
				Eventually(session.Out).Should(Say("Uninstalling plugin failing-plugin\\.\\.\\."))
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("I'm failing...I'm failing..."))
				Eventually(session).Should(Exit(1))

				session = helpers.CF("plugins")
				Eventually(session.Out).Should(Say("banana-plugin-name-1"))
				Eventually(session.Out).Should(Say("banana-plugin-name-2"))
				Eventually(session.Out).Should(Say("failing-plugin"))
				Eventually(session).Should(Exit(0))
			})
		})

		Context("and the user cannot remove the plugin's executable (NON_WINDOWS)", func() {
			var executablePath string
			var pluginsRootDir string

			BeforeEach(func() {
				pluginsRootDir = filepath.Join(homeDir, ".cf", "plugins")
				executablePath = filepath.Join(pluginsRootDir, "some-dir")
				rawConfig := fmt.Sprintf(`
				{
					"Plugins": {
						"banana-plugin-name-1": {
							"Location": "%s",
							"Version": {
								"Major": 1,
								"Minor": 0,
								"Build": 1
							},
							"Commands": [
								{
									"Name": "enable-diego",
									"Alias": "",
									"HelpText": "enable Diego support for an app",
									"UsageDetails": {
										"Usage": "cf enable-diego APP_NAME",
										"Options": null
									}
								}
							]
						},
						"banana-plugin-name-2": {
							"Location": "%s",
							"Version": {
								"Major": 1,
								"Minor": 0,
								"Build": 1
							},
							"Commands": [
								{
									"Name": "enable-banana",
									"Alias": "",
									"HelpText": "enable banana support for an app",
									"UsageDetails": {
										"Usage": "cf enable-banana APP_NAME",
										"Options": null
									}
								}
							]
						}
					}
				}`, executablePath, executablePath)

				os.MkdirAll(pluginsRootDir, 0700)
				err := ioutil.WriteFile(filepath.Join(pluginsRootDir, "config.json"), []byte(rawConfig), 644)
				Expect(err).ToNot(HaveOccurred())
				err = os.MkdirAll(executablePath, 0700)
				Expect(err).ToNot(HaveOccurred())
				err = ioutil.WriteFile(filepath.Join(executablePath, "foooooooo"), nil, 644)
				Expect(err).ToNot(HaveOccurred())
			})

			It("exits with an error, and does not remove the plugin", func() {
				session := helpers.CF("uninstall-plugin", "banana-plugin-name-1")
				Eventually(session.Out).Should(Say("Uninstalling plugin banana-plugin-name-1\\.\\.\\."))
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("some-dir: permission denied"))
				Eventually(session).Should(Exit(1))

				session = helpers.CF("plugins")
				Eventually(session.Out).Should(Say("banana-plugin-name-1"))
				Eventually(session.Out).Should(Say("banana-plugin-name-2"))
				Eventually(session).Should(Exit(0))
			})
		})
	})
})
