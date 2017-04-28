package plugin

import (
	"path/filepath"

	"code.cloudfoundry.org/cli/integration/helpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("install-plugin command", func() {
	BeforeEach(func() {
		helpers.RunIfExperimental("install-plugin refactor is experimental")
	})

	It("displays the experimental warning", func() {
	})

	Describe("help", func() {
		Context("when the --help flag is given", func() {
			It("displays command usage to stdout", func() {
				session := helpers.CF("install-plugin", "--help")

				Eventually(session.Out).Should(Say("NAME:"))
				Eventually(session.Out).Should(Say("install-plugin - Install CLI plugin"))
				Eventually(session.Out).Should(Say("USAGE:"))
				Eventually(session.Out).Should(Say("cf install-plugin \\(LOCAL-PATH/TO/PLUGIN | URL | -r REPO_NAME PLUGIN_NAME\\) \\[-f\\]"))
				Eventually(session.Out).Should(Say("EXAMPLES:"))
				Eventually(session.Out).Should(Say("cf install-plugin ~/Downloads/plugin-foobar"))
				Eventually(session.Out).Should(Say("cf install-plugin https://example.com/plugin-foobar_linux_amd64"))
				Eventually(session.Out).Should(Say("cf install-plugin -r My-Repo plugin-echo"))
				Eventually(session.Out).Should(Say("OPTIONS:"))
				Eventually(session.Out).Should(Say("-f      Force install of plugin without confirmation"))
				Eventually(session.Out).Should(Say("-r      Name of a registered repository where the specified plugin is located"))
				Eventually(session.Out).Should(Say("SEE ALSO:"))
				Eventually(session.Out).Should(Say("add-plugin-repo, list-plugin-repos, plugins"))

				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("installing a plugin from a local file", func() {
		var pluginPath string

		BeforeEach(func() {
			pluginPath = helpers.BuildConfigurablePlugin("configurable_plugin", "some-plugin", "1.0.0",
				[]helpers.PluginCommand{
					{Name: "some-command", Help: "some-command-help"},
				},
			)
		})

		FContext("when the -f flag is given", func() {
			FIt("installs the plugin", func() {
				session := helpers.CF("install-plugin", pluginPath, "-f")

				Eventually(session.Out).Should(Say("Attention: Plugins are binaries written by potentially untrusted authors."))
				Eventually(session.Out).Should(Say("Install and use plugins at your own risk."))
				Eventually(session.Out).Should(Say("Installing plugin some-plugin..."))
				Eventually(session.Out).Should(Say("OK"))
				Eventually(session.Out).Should(Say("Plugin some-plugin successfully installed."))

				Eventually(session).Should(Exit(0))

				pluginsSession := helpers.CF("plugins", "--checksum")
				expectedSha := helpers.Sha1Sum(
					filepath.Join(homeDir, ".cf/plugins/configurable_plugin.some-plugin"))
				Eventually(pluginsSession.Out).Should(Say("some-plugin\\s+1.0.0\\s+%s", expectedSha))
				Eventually(pluginsSession).Should(Exit(0))

				Eventually(helpers.CF("some-command")).Should(Exit(0))
				helpSession := helpers.CF("help")
				Eventually(helpSession.Out).Should(Say("some-command"))
				Eventually(helpSession).Should(Exit(0))
			})

			Context("when the file is not executable", func() {
				It("installs the plugin", func() {
					// check output
					// cf plugins --shasum shows plugin with correct sha sum
					// cf can run command from plugin
					// cf help shows plugins commands
				})
			})

			Context("when the plugin is already installed", func() {
				It("uninstalls the existing plugin and installs the plugin", func() {
				})
			})

			Context("when the file does not exist", func() {
				It("tells the user that the file was not found and fails", func() {
					// should not display the warning about untrusted authors
				})
			})

			Context("when the file is not a plugin", func() {
				It("tells the user that the file is not a plugin and fails", func() {
				})
			})

			Context("command conflict", func() {
				Context("when the plugin has a command that is the same as a built-in command", func() {
					It("tells the user about the conflict and fails", func() {
					})
				})

				Context("when the plugin has a command that is the same as a built-in alias", func() {
					It("tells the user about the conflict and fails", func() {
					})
				})

				Context("when the plugin has a command that is the same as another plugin command", func() {
					It("tells the user about the conflict and fails", func() {
					})
				})

				Context("when the plugin has a command that is the same as another plugin alias", func() {
					It("tells the user about the conflict and fails", func() {
					})
				})
			})

			Context("alias conflict", func() {
				Context("when the plugin has an alias that is the same as a built-in command", func() {
					It("tells the user about the conflict and fails", func() {
					})
				})

				Context("when the plugin has an alias that is the same as a built-in alias", func() {
					It("tells the user about the conflict and fails", func() {
					})
				})

				Context("when the plugin has an alias that is the same as another plugin command", func() {
					It("tells the user about the conflict and fails", func() {
					})
				})

				Context("when the plugin has an alias that is the same as another plugin alias", func() {
					It("tells the user about the conflict and fails", func() {
					})
				})
			})
		})

		Context("when the -f flag is not given", func() {
			Context("when the user says yes", func() {
				It("installs the plugin", func() {
					// check output
					// cf plugins --shasum shows plugin with correct sha sum
					// cf can run command from plugin
					// cf help shows plugins commands
				})
			})

			Context("when the user says no", func() {
				It("does not install the plugin", func() {
				})
			})

			Context("when the user interrupts with control-c", func() {
				It("does not install the plugin and does not create a bad state", func() {
					// cf plugins still works
					// cf install plugin -f of the same plugin works
				})
			})

			Context("when the plugin is already installed", func() {
				It("tells the user that the plugin is already installed and fails", func() {
				})
			})
		})

		Context("when the file is not executable", func() {
		})
	})

	Context("when the plugin contains command names that match core commands", func() {
		It("displays an error on installation", func() {
			session := helpers.CF("install-plugin", "-f", overrideTestPluginPath)
			Eventually(session).Should(Say("Command `push` in the plugin being installed is a native CF command/alias."))
			Eventually(session).Should(Exit(1))
		})
	})
})
