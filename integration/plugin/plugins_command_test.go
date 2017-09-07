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
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("plugins command", func() {
	Describe("help", func() {
		Context("when --help flag is provided", func() {
			It("displays command usage to output", func() {
				session := helpers.CF("plugins", "--help")
				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("plugins - List commands of installed plugins"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say("cf plugins [--checksum | --outdated]"))
				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say("--checksum\\s+Compute and show the sha1 value of the plugin binary file"))
				Eventually(session).Should(Say("--outdated\\s+Search the plugin repositories for new versions of installed plugins"))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("install-plugin, repo-plugins, uninstall-plugin"))
				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("when no plugins are installed", func() {
		It("displays an empty table", func() {
			session := helpers.CF("plugins")
			Eventually(session.Out).Should(Say("Listing installed plugins..."))
			Eventually(session.Out).Should(Say(""))
			Eventually(session.Out).Should(Say("plugin\\s+version\\s+command name\\s+command help"))
			Eventually(session.Out).Should(Say(""))
			Eventually(session.Out).Should(Say("Use 'cf repo-plugins' to list plugins in registered repos available to install\\."))
			Consistently(session.Out).ShouldNot(Say("[a-za-z0-9]+"))
			Eventually(session).Should(Exit(0))
		})

		Context("when the --checksum flag is provided", func() {
			It("displays an empty checksum table", func() {
				session := helpers.CF("plugins", "--checksum")
				Eventually(session.Out).Should(Say("Computing sha1 for installed plugins, this may take a while..."))
				Eventually(session.Out).Should(Say(""))
				Eventually(session.Out).Should(Say("plugin\\s+version\\s+sha1"))
				Consistently(session.Out).ShouldNot(Say("[a-za-z0-9]+"))
				Eventually(session).Should(Exit(0))
			})
		})

		Context("when the --outdated flag is provided", func() {
			It("errors with no repositories", func() {
				session := helpers.CF("plugins", "--outdated")
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("No plugin repositories registered to search for plugin updates."))

				Eventually(session).Should(Exit(1))
			})
		})
	})

	Context("when plugins are installed", func() {
		Context("when there are multiple plugins", func() {
			BeforeEach(func() {
				helpers.InstallConfigurablePlugin("I-should-be-sorted-first", "1.2.0", []helpers.PluginCommand{
					{Name: "command-1", Help: "some-command-1"},
					{Name: "Better-command", Help: "some-better-command"},
					{Name: "command-2", Help: "some-command-2"},
				})
				helpers.InstallConfigurablePlugin("sorted-third", "2.0.1", []helpers.PluginCommand{
					{Name: "banana-command", Help: "banana-command"},
				})
				helpers.InstallConfigurablePlugin("i-should-be-sorted-second", "1.0.0", []helpers.PluginCommand{
					{Name: "some-command", Help: "some-command"},
					{Name: "Some-other-command", Help: "some-other-command"},
				})
			})

			It("displays the installed plugins in alphabetical order", func() {
				session := helpers.CF("plugins")
				Eventually(session).Should(Say("Listing installed plugins..."))
				Eventually(session).Should(Say(""))
				Eventually(session).Should(Say("plugin\\s+version\\s+command name\\s+command help"))
				Eventually(session).Should(Say("I-should-be-sorted-first\\s+1\\.2\\.0\\s+Better-command\\s+some-better-command"))
				Eventually(session).Should(Say("I-should-be-sorted-first\\s+1\\.2\\.0\\s+command-1\\s+some-command-1"))
				Eventually(session).Should(Say("I-should-be-sorted-first\\s+1\\.2\\.0\\s+command-2\\s+some-command-2"))
				Eventually(session).Should(Say("i-should-be-sorted-second\\s+1\\.0\\.0\\s+some-command\\s+some-command"))
				Eventually(session).Should(Say("i-should-be-sorted-second\\s+1\\.0\\.0\\s+Some-other-command\\s+some-other-command"))
				Eventually(session).Should(Say("sorted-third\\s+2\\.0\\.1\\s+banana-command\\s+banana-command"))
				Eventually(session.Out).Should(Say(""))
				Eventually(session.Out).Should(Say("Use 'cf repo-plugins' to list plugins in registered repos available to install\\."))
				Eventually(session).Should(Exit(0))
			})
		})

		Context("when plugin version information is 0.0.0", func() {
			BeforeEach(func() {
				helpers.InstallConfigurablePlugin("some-plugin", "0.0.0", []helpers.PluginCommand{
					{Name: "banana-command", Help: "banana-command"},
				})
			})

			It("displays N/A for the plugin's version", func() {
				session := helpers.CF("plugins")
				Eventually(session.Out).Should(Say("some-plugin\\s+N/A"))
				Eventually(session).Should(Exit(0))
			})
		})

		Context("when a plugin command has an alias", func() {
			BeforeEach(func() {
				helpers.InstallConfigurablePlugin("some-plugin", "1.0.0", []helpers.PluginCommand{
					{Name: "banana-command", Alias: "bc", Help: "banana-command"},
				})
			})

			It("displays the command name and it's alias", func() {
				session := helpers.CF("plugins")
				Eventually(session.Out).Should(Say("some-plugin\\s+1\\.0\\.0\\s+banana-command, bc"))
				Eventually(session).Should(Exit(0))
			})
		})

		Context("when the --checksum flag is provided", func() {
			var installedPluginPath string

			BeforeEach(func() {
				helpers.InstallConfigurablePlugin("some-plugin", "1.0.0", []helpers.PluginCommand{
					{Name: "banana-command", Help: "banana-command"},
				})
				installedPluginPath = generic.ExecutableFilename(filepath.Join(homeDir, ".cf", "plugins", "some-plugin"))
			})

			It("displays the sha1 value for each installed plugin", func() {
				calculatedSha := helpers.Sha1Sum(installedPluginPath)
				session := helpers.CF("plugins", "--checksum")
				Eventually(session.Out).Should(Say("Computing sha1 for installed plugins, this may take a while..."))
				Eventually(session.Out).Should(Say(""))
				Eventually(session.Out).Should(Say("plugin\\s+version\\s+sha1"))
				Eventually(session.Out).Should(Say("some-plugin\\s+1\\.0\\.0\\s+%s", calculatedSha))
				Eventually(session).Should(Exit(0))
			})

			Context("when an error is encountered calculating the sha1 value", func() {
				It("displays N/A for the plugin's sha1", func() {
					err := os.Remove(installedPluginPath)
					Expect(err).NotTo(HaveOccurred())

					session := helpers.CF("plugins", "--checksum")
					Eventually(session.Out).Should(Say("some-plugin\\s+1\\.0\\.0\\s+N/A"))
					Eventually(session).Should(Exit(0))
				})
			})
		})

		Context("when the --outdated flag is provided", func() {
			Context("when there are no repos", func() {
				BeforeEach(func() {
					helpers.InstallConfigurablePlugin("some-plugin", "1.0.0", []helpers.PluginCommand{
						{Name: "banana-command", Alias: "bc", Help: "banana-command"},
					})
				})

				It("aborts with error 'No plugin repositories added' and exit code 1", func() {
					session := helpers.CF("plugins", "--outdated")
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("No plugin repositories registered to search for plugin updates."))
					Eventually(session).Should(Exit(1))
				})
			})

			Context("when there is 1 repository", func() {
				var (
					server1 *Server
				)

				BeforeEach(func() {
					server1 = helpers.NewPluginRepositoryTLSServer(helpers.PluginRepository{
						Plugins: []helpers.Plugin{
							{Name: "plugin-1", Version: "1.0.0"},
							{Name: "plugin-2", Version: "2.0.0"},
						},
					})

					Eventually(helpers.CF("add-plugin-repo", "repo1", server1.URL(), "-k")).Should(Exit(0))
					// TODO: re-add when refactor repo-plugins
					// session := helpers.CF("repo-plugins")
					// Eventually(session).Should(Say("plugin-1\\s+1\\.0\\.0"))
					// Eventually(session).Should(Say("plugin-2\\s+2\\.0\\.0"))
					// Eventually(session).Should(Exit(0))
				})

				AfterEach(func() {
					server1.Close()
				})

				Context("when nothing is outdated", func() {
					BeforeEach(func() {
						helpers.InstallConfigurablePlugin("plugin-1", "1.0.0", []helpers.PluginCommand{
							{Name: "banana-command-1", Help: "banana-command"},
						})
						helpers.InstallConfigurablePlugin("plugin-2", "2.0.0", []helpers.PluginCommand{
							{Name: "banana-command-2", Help: "banana-command"},
						})
					})

					AfterEach(func() {
						Eventually(helpers.CF("uninstall-plugin", "plugin-1")).Should(Exit(0))
						Eventually(helpers.CF("uninstall-plugin", "plugin-2")).Should(Exit(0))
					})

					It("displays an empty table", func() {
						session := helpers.CF("plugins", "--outdated", "-k")
						Eventually(session).Should(Say("Searching repo1 for newer versions of installed plugins..."))
						Eventually(session).Should(Say(""))
						Eventually(session).Should(Say("plugin\\s+version\\s+latest version\\n\\nUse 'cf install-plugin' to update a plugin to the latest version\\."))
						Eventually(session).Should(Exit(0))
					})
				})

				Context("when the plugins are outdated", func() {
					BeforeEach(func() {
						helpers.InstallConfigurablePlugin("plugin-1", "0.9.0", []helpers.PluginCommand{
							{Name: "banana-command-1", Help: "banana-command"},
						})
						helpers.InstallConfigurablePlugin("plugin-2", "1.9.0", []helpers.PluginCommand{
							{Name: "banana-command-2", Help: "banana-command"},
						})
					})

					AfterEach(func() {
						Eventually(helpers.CF("uninstall-plugin", "plugin-1")).Should(Exit(0))
						Eventually(helpers.CF("uninstall-plugin", "plugin-2")).Should(Exit(0))
					})

					It("displays the table with outdated plugin and new version", func() {
						session := helpers.CF("plugins", "--outdated", "-k")
						Eventually(session).Should(Say("Searching repo1 for newer versions of installed plugins..."))
						Eventually(session).Should(Say(""))
						Eventually(session).Should(Say("plugin\\s+version\\s+latest version"))
						Eventually(session).Should(Say("plugin-1\\s+0\\.9\\.0\\s+1\\.0\\.0"))
						Eventually(session).Should(Say("plugin-2\\s+1\\.9\\.0\\s+2\\.0\\.0"))
						Eventually(session).Should(Say(""))
						Eventually(session).Should(Say("Use 'cf install-plugin' to update a plugin to the latest version\\."))
						Eventually(session).Should(Exit(0))
					})
				})
			})

			Context("when multiple repositories are registered", func() {
				var (
					server1 *Server
					server2 *Server
				)

				BeforeEach(func() {
					server1 = helpers.NewPluginRepositoryTLSServer(helpers.PluginRepository{
						Plugins: []helpers.Plugin{
							{Name: "plugin-1", Version: "1.0.0"},
							{Name: "plugin-3", Version: "3.5.0"},
						},
					})

					server2 = helpers.NewPluginRepositoryTLSServer(helpers.PluginRepository{
						Plugins: []helpers.Plugin{
							{Name: "plugin-2", Version: "2.0.0"},
							{Name: "plugin-3", Version: "3.0.0"},
						},
					})

					Eventually(helpers.CF("add-plugin-repo", "repo1", server1.URL(), "-k")).Should(Exit(0))
					Eventually(helpers.CF("add-plugin-repo", "repo2", server2.URL(), "-k")).Should(Exit(0))
				})

				AfterEach(func() {
					server1.Close()
					server2.Close()
				})

				Context("when plugins are outdated", func() {
					BeforeEach(func() {
						helpers.InstallConfigurablePlugin("plugin-1", "0.9.0", []helpers.PluginCommand{
							{Name: "banana-command-1", Help: "banana-command"},
						})
						helpers.InstallConfigurablePlugin("plugin-2", "1.9.0", []helpers.PluginCommand{
							{Name: "banana-command-2", Help: "banana-command"},
						})
					})

					It("displays the table with outdated plugin and new version", func() {
						session := helpers.CF("plugins", "--outdated", "-k")
						Eventually(session).Should(Say("Searching repo1, repo2 for newer versions of installed plugins..."))
						Eventually(session).Should(Say("plugin\\s+version\\s+latest version"))
						Eventually(session).Should(Say("plugin-1\\s+0\\.9\\.0\\s+1\\.0\\.0"))
						Eventually(session).Should(Say("plugin-2\\s+1\\.9\\.0\\s+2\\.0\\.0"))
						Eventually(session).Should(Say(""))
						Eventually(session).Should(Say("Use 'cf install-plugin' to update a plugin to the latest version\\."))
						Eventually(session).Should(Exit(0))
					})
				})

				Context("when the same plugin is outdated from multiple repositories", func() {
					BeforeEach(func() {
						helpers.InstallConfigurablePlugin("plugin-1", "0.9.0", []helpers.PluginCommand{
							{Name: "banana-command-1", Help: "banana-command"},
						})
						helpers.InstallConfigurablePlugin("plugin-2", "1.9.0", []helpers.PluginCommand{
							{Name: "banana-command-2", Help: "banana-command"},
						})
						helpers.InstallConfigurablePlugin("plugin-3", "2.9.0", []helpers.PluginCommand{
							{Name: "banana-command-3", Help: "banana-command"},
						})
					})

					It("only displays the newest version of the plugin found in the repositories", func() {
						session := helpers.CF("plugins", "--outdated", "-k")
						Eventually(session).Should(Say("Searching repo1, repo2 for newer versions of installed plugins..."))
						Eventually(session).Should(Say(""))
						Eventually(session).Should(Say("plugin\\s+version\\s+latest version"))
						Eventually(session).Should(Say("plugin-1\\s+0\\.9\\.0\\s+1\\.0\\.0"))
						Eventually(session).Should(Say("plugin-2\\s+1\\.9\\.0\\s+2\\.0\\.0"))
						Eventually(session).Should(Say("plugin-3\\s+2\\.9\\.0\\s+3\\.5\\.0"))
						Eventually(session).Should(Say(""))
						Eventually(session).Should(Say("Use 'cf install-plugin' to update a plugin to the latest version\\."))
						Eventually(session).Should(Exit(0))
					})
				})
			})
		})
	})
})
