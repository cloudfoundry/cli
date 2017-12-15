package plugin

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
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

var _ = Describe("install-plugin command", func() {
	var (
		buffer     *Buffer
		pluginPath string
	)

	AfterEach(func() {
		pluginsHomeDirContents, err := ioutil.ReadDir(filepath.Join(homeDir, ".cf", "plugins"))
		if os.IsNotExist(err) {
			return
		}

		Expect(err).ToNot(HaveOccurred())

		for _, entry := range pluginsHomeDirContents {
			Expect(entry.Name()).NotTo(ContainSubstring("temp"))
		}
	})

	Describe("help", func() {
		Context("when the --help flag is given", func() {
			It("displays command usage to stdout", func() {
				session := helpers.CF("install-plugin", "--help")

				Eventually(session.Out).Should(Say("NAME:"))
				Eventually(session.Out).Should(Say("install-plugin - Install CLI plugin"))
				Eventually(session.Out).Should(Say("USAGE:"))
				Eventually(session.Out).Should(Say("cf install-plugin PLUGIN_NAME \\[-r REPO_NAME\\] \\[-f\\]"))
				Eventually(session.Out).Should(Say("cf install-plugin LOCAL-PATH/TO/PLUGIN | URL \\[-f\\]"))
				Eventually(session.Out).Should(Say(""))
				Eventually(session.Out).Should(Say("WARNING:"))
				Eventually(session.Out).Should(Say("Plugins are binaries written by potentially untrusted authors."))
				Eventually(session.Out).Should(Say("Install and use plugins at your own risk."))
				Eventually(session.Out).Should(Say(""))
				Eventually(session.Out).Should(Say("EXAMPLES:"))
				Eventually(session.Out).Should(Say("cf install-plugin ~/Downloads/plugin-foobar"))
				Eventually(session.Out).Should(Say("cf install-plugin https://example.com/plugin-foobar_linux_amd64"))
				Eventually(session.Out).Should(Say("cf install-plugin -r My-Repo plugin-echo"))
				Eventually(session.Out).Should(Say("OPTIONS:"))
				Eventually(session.Out).Should(Say("-f\\s+Force install of plugin without confirmation"))
				Eventually(session.Out).Should(Say("-r\\s+Restrict search for plugin to this registered repository"))
				Eventually(session.Out).Should(Say("SEE ALSO:"))
				Eventually(session.Out).Should(Say("add-plugin-repo, list-plugin-repos, plugins"))

				Eventually(session).Should(Exit(0))
			})
		})
	})

	Context("when the user does not provide a plugin name or location", func() {
		It("errors and displays usage", func() {
			session := helpers.CF("install-plugin")
			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `PLUGIN_NAME_OR_LOCATION` was not provided"))
			Eventually(session.Out).Should(Say("USAGE:"))

			Eventually(session).Should(Exit(1))
		})
	})

	Describe("installing a plugin from a local file", func() {
		Context("when the file is compiled for a different os and architecture", func() {
			BeforeEach(func() {
				goos := os.Getenv("GOOS")
				goarch := os.Getenv("GOARCH")

				err := os.Setenv("GOOS", "openbsd")
				Expect(err).ToNot(HaveOccurred())
				err = os.Setenv("GOARCH", "amd64")
				Expect(err).ToNot(HaveOccurred())

				pluginPath = helpers.BuildConfigurablePlugin("configurable_plugin", "some-plugin", "1.0.0",
					[]helpers.PluginCommand{
						{Name: "some-command", Help: "some-command-help"},
					},
				)

				err = os.Setenv("GOOS", goos)
				Expect(err).ToNot(HaveOccurred())
				err = os.Setenv("GOARCH", goarch)
				Expect(err).ToNot(HaveOccurred())
			})

			It("fails and reports the file is not a valid CLI plugin", func() {
				session := helpers.CF("install-plugin", pluginPath, "-f")

				Eventually(session.Out).Should(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
				Eventually(session.Out).Should(Say("Install and use plugins at your own risk\\."))
				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("File is not a valid cf CLI plugin binary\\."))

				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the file is compiled for the correct os and architecture", func() {
			BeforeEach(func() {
				pluginPath = helpers.BuildConfigurablePlugin("configurable_plugin", "some-plugin", "1.0.0",
					[]helpers.PluginCommand{
						{Name: "some-command", Help: "some-command-help"},
					},
				)
			})

			Context("when the -f flag is given", func() {
				It("installs the plugin and cleans up all temp files", func() {
					session := helpers.CF("install-plugin", pluginPath, "-f")

					Eventually(session.Out).Should(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
					Eventually(session.Out).Should(Say("Install and use plugins at your own risk\\."))
					Eventually(session.Out).Should(Say("Installing plugin some-plugin\\.\\.\\."))
					Eventually(session.Out).Should(Say("OK"))
					Eventually(session.Out).Should(Say("Plugin some-plugin 1\\.0\\.0 successfully installed\\."))

					Eventually(session).Should(Exit(0))

					installedPath := generic.ExecutableFilename(filepath.Join(homeDir, ".cf", "plugins", "some-plugin"))

					pluginsSession := helpers.CF("plugins", "--checksum")
					expectedSha := helpers.Sha1Sum(installedPath)

					Eventually(pluginsSession.Out).Should(Say("some-plugin\\s+1\\.0\\.0\\s+%s", expectedSha))
					Eventually(pluginsSession).Should(Exit(0))

					Eventually(helpers.CF("some-command")).Should(Exit(0))

					helpSession := helpers.CF("help")
					Eventually(helpSession.Out).Should(Say("some-command"))
					Eventually(helpSession).Should(Exit(0))
				})

				Context("when the file does not have executable permissions", func() {
					BeforeEach(func() {
						Expect(os.Chmod(pluginPath, 0666)).ToNot(HaveOccurred())
					})

					It("installs the plugin", func() {
						session := helpers.CF("install-plugin", pluginPath, "-f")
						Eventually(session.Out).Should(Say("Plugin some-plugin 1\\.0\\.0 successfully installed\\."))
						Eventually(session).Should(Exit(0))
					})
				})

				Context("when the plugin is already installed", func() {
					BeforeEach(func() {
						Eventually(helpers.CF("install-plugin", pluginPath, "-f")).Should(Exit(0))
					})

					It("uninstalls the existing plugin and installs the plugin", func() {
						session := helpers.CF("install-plugin", pluginPath, "-f")

						Eventually(session.Out).Should(Say("Plugin some-plugin 1\\.0\\.0 is already installed\\. Uninstalling existing plugin\\.\\.\\."))
						Eventually(session.Out).Should(Say("CLI-MESSAGE-UNINSTALL"))
						Eventually(session.Out).Should(Say("Plugin some-plugin successfully uninstalled\\."))
						Eventually(session.Out).Should(Say("Plugin some-plugin 1\\.0\\.0 successfully installed\\."))

						Eventually(session).Should(Exit(0))
					})
				})

				Context("when the file does not exist", func() {
					It("tells the user that the file was not found and fails", func() {
						session := helpers.CF("install-plugin", "some/path/that/does/not/exist", "-f")
						Eventually(session.Err).Should(Say("Plugin some/path/that/does/not/exist not found on disk or in any registered repo\\."))
						Eventually(session.Err).Should(Say("Use 'cf repo-plugins' to list plugins available in the repos\\."))

						Consistently(session.Out).ShouldNot(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
						Consistently(session.Out).ShouldNot(Say("Install and use plugins at your own risk\\."))

						Eventually(session).Should(Exit(1))
					})
				})

				Context("when the file is not an executable", func() {
					BeforeEach(func() {
						badPlugin, err := ioutil.TempFile("", "")
						Expect(err).ToNot(HaveOccurred())
						pluginPath = badPlugin.Name()
						err = badPlugin.Close()
						Expect(err).ToNot(HaveOccurred())
					})

					AfterEach(func() {
						err := os.Remove(pluginPath)
						Expect(err).ToNot(HaveOccurred())
					})

					It("tells the user that the file is not a plugin and fails", func() {
						session := helpers.CF("install-plugin", pluginPath, "-f")
						Eventually(session.Err).Should(Say("File is not a valid cf CLI plugin binary\\."))

						Eventually(session).Should(Exit(1))
					})
				})

				Context("when the file is not a plugin", func() {
					BeforeEach(func() {
						var err error
						pluginPath, err = Build("code.cloudfoundry.org/cli/integration/assets/non_plugin")
						Expect(err).ToNot(HaveOccurred())
					})

					It("tells the user that the file is not a plugin and fails", func() {
						session := helpers.CF("install-plugin", pluginPath, "-f")
						Eventually(session.Err).Should(Say("File is not a valid cf CLI plugin binary\\."))

						Eventually(session).Should(Exit(1))
					})
				})

				Context("when getting metadata from the plugin errors", func() {
					BeforeEach(func() {
						var err error
						pluginPath, err = Build("code.cloudfoundry.org/cli/integration/assets/test_plugin_fails_metadata")
						Expect(err).ToNot(HaveOccurred())
					})

					It("displays the error to stderr", func() {
						session := helpers.CF("install-plugin", pluginPath, "-f")
						Eventually(session.Err).Should(Say("exit status 51"))
						Eventually(session.Err).Should(Say("File is not a valid cf CLI plugin binary\\."))

						Eventually(session).Should(Exit(1))
					})
				})

				Context("when there is a command conflict", func() {
					Context("when the plugin has a command that is the same as a built-in command", func() {
						BeforeEach(func() {
							pluginPath = helpers.BuildConfigurablePlugin(
								"configurable_plugin", "some-plugin", "1.1.1",
								[]helpers.PluginCommand{
									{Name: "version"},
								})
						})

						It("tells the user about the conflict and fails", func() {
							session := helpers.CF("install-plugin", "-f", pluginPath)

							Eventually(session.Out).Should(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
							Eventually(session.Out).Should(Say("Install and use plugins at your own risk\\."))

							Eventually(session.Out).Should(Say("FAILED"))
							Eventually(session.Err).Should(Say("Plugin some-plugin v1\\.1\\.1 could not be installed as it contains commands with names that are already used: version"))

							Eventually(session).Should(Exit(1))
						})
					})

					Context("when the plugin has a command that is the same as a built-in alias", func() {
						BeforeEach(func() {
							pluginPath = helpers.BuildConfigurablePlugin(
								"configurable_plugin", "some-plugin", "1.1.1",
								[]helpers.PluginCommand{
									{Name: "cups"},
								})
						})

						It("tells the user about the conflict and fails", func() {
							session := helpers.CF("install-plugin", "-f", pluginPath)

							Eventually(session.Out).Should(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
							Eventually(session.Out).Should(Say("Install and use plugins at your own risk\\."))

							Eventually(session.Out).Should(Say("FAILED"))
							Eventually(session.Err).Should(Say("Plugin some-plugin v1\\.1\\.1 could not be installed as it contains commands with names that are already used: cups"))

							Eventually(session).Should(Exit(1))
						})
					})

					Context("when the plugin has a command that is the same as another plugin command", func() {
						BeforeEach(func() {
							helpers.InstallConfigurablePlugin("existing-plugin", "1.1.1",
								[]helpers.PluginCommand{
									{Name: "existing-command"},
								})

							pluginPath = helpers.BuildConfigurablePlugin(
								"configurable_plugin", "new-plugin", "1.1.1",
								[]helpers.PluginCommand{
									{Name: "existing-command"},
								})
						})

						It("tells the user about the conflict and fails", func() {
							session := helpers.CF("install-plugin", "-f", pluginPath)

							Eventually(session.Out).Should(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
							Eventually(session.Out).Should(Say("Install and use plugins at your own risk\\."))

							Eventually(session.Out).Should(Say("FAILED"))
							Eventually(session.Err).Should(Say("Plugin new-plugin v1\\.1\\.1 could not be installed as it contains commands with names that are already used: existing-command\\."))

							Eventually(session).Should(Exit(1))
						})
					})

					Context("when the plugin has a command that is the same as another plugin alias", func() {
						BeforeEach(func() {
							helpers.InstallConfigurablePlugin("existing-plugin", "1.1.1",
								[]helpers.PluginCommand{
									{Name: "existing-command"},
								})

							pluginPath = helpers.BuildConfigurablePlugin(
								"configurable_plugin", "new-plugin", "1.1.1",
								[]helpers.PluginCommand{
									{Name: "new-command", Alias: "existing-command"},
								})
						})

						It("tells the user about the conflict and fails", func() {
							session := helpers.CF("install-plugin", "-f", pluginPath)

							Eventually(session.Out).Should(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
							Eventually(session.Out).Should(Say("Install and use plugins at your own risk\\."))

							Eventually(session.Out).Should(Say("FAILED"))
							Eventually(session.Err).Should(Say("Plugin new-plugin v1\\.1\\.1 could not be installed as it contains commands with aliases that are already used: existing-command\\."))

							Eventually(session).Should(Exit(1))
						})
					})
				})

				Context("alias conflict", func() {
					Context("when the plugin has an alias that is the same as a built-in command", func() {

						BeforeEach(func() {
							pluginPath = helpers.BuildConfigurablePlugin(
								"configurable_plugin", "some-plugin", "1.1.1",
								[]helpers.PluginCommand{
									{Name: "some-command", Alias: "version"},
								})
						})

						It("tells the user about the conflict and fails", func() {
							session := helpers.CF("install-plugin", "-f", pluginPath)

							Eventually(session.Out).Should(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
							Eventually(session.Out).Should(Say("Install and use plugins at your own risk\\."))

							Eventually(session.Out).Should(Say("FAILED"))
							Eventually(session.Err).Should(Say("Plugin some-plugin v1\\.1\\.1 could not be installed as it contains commands with aliases that are already used: version"))

							Eventually(session).Should(Exit(1))
						})
					})

					Context("when the plugin has an alias that is the same as a built-in alias", func() {
						BeforeEach(func() {
							pluginPath = helpers.BuildConfigurablePlugin(
								"configurable_plugin", "some-plugin", "1.1.1",
								[]helpers.PluginCommand{
									{Name: "some-command", Alias: "cups"},
								})
						})

						It("tells the user about the conflict and fails", func() {
							session := helpers.CF("install-plugin", "-f", pluginPath)

							Eventually(session.Out).Should(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
							Eventually(session.Out).Should(Say("Install and use plugins at your own risk\\."))

							Eventually(session.Out).Should(Say("FAILED"))
							Eventually(session.Err).Should(Say("Plugin some-plugin v1\\.1\\.1 could not be installed as it contains commands with aliases that are already used: cups"))

							Eventually(session).Should(Exit(1))
						})
					})

					Context("when the plugin has an alias that is the same as another plugin command", func() {
						BeforeEach(func() {
							helpers.InstallConfigurablePlugin("existing-plugin", "1.1.1",
								[]helpers.PluginCommand{
									{Name: "existing-command"},
								})

							pluginPath = helpers.BuildConfigurablePlugin(
								"configurable_plugin", "new-plugin", "1.1.1",
								[]helpers.PluginCommand{
									{Name: "new-command", Alias: "existing-command"},
								})
						})

						It("tells the user about the conflict and fails", func() {
							session := helpers.CF("install-plugin", "-f", pluginPath)

							Eventually(session.Out).Should(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
							Eventually(session.Out).Should(Say("Install and use plugins at your own risk\\."))

							Eventually(session.Out).Should(Say("FAILED"))
							Eventually(session.Err).Should(Say("Plugin new-plugin v1\\.1\\.1 could not be installed as it contains commands with aliases that are already used: existing-command\\."))

							Eventually(session).Should(Exit(1))
						})
					})

					Context("when the plugin has an alias that is the same as another plugin alias", func() {
						BeforeEach(func() {
							helpers.InstallConfigurablePlugin("existing-plugin", "1.1.1",
								[]helpers.PluginCommand{
									{Name: "existing-command", Alias: "existing-alias"},
								})

							pluginPath = helpers.BuildConfigurablePlugin(
								"configurable_plugin", "new-plugin", "1.1.1",
								[]helpers.PluginCommand{
									{Name: "new-command", Alias: "existing-alias"},
								})
						})

						It("tells the user about the conflict and fails", func() {
							session := helpers.CF("install-plugin", "-f", pluginPath)

							Eventually(session.Out).Should(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
							Eventually(session.Out).Should(Say("Install and use plugins at your own risk\\."))

							Eventually(session.Out).Should(Say("FAILED"))
							Eventually(session.Err).Should(Say("Plugin new-plugin v1\\.1\\.1 could not be installed as it contains commands with aliases that are already used: existing-alias\\."))

							Eventually(session).Should(Exit(1))
						})
					})
				})

				Context("alias and command conflicts", func() {
					Context("when the plugin has a command and an alias that are both taken by another plugin", func() {
						BeforeEach(func() {
							helpers.InstallConfigurablePlugin("existing-plugin", "1.1.1",
								[]helpers.PluginCommand{
									{Name: "existing-command", Alias: "existing-alias"},
								})

							pluginPath = helpers.BuildConfigurablePlugin(
								"configurable_plugin", "new-plugin", "1.1.1",
								[]helpers.PluginCommand{
									{Name: "existing-command", Alias: "existing-alias"},
								})
						})

						It("tells the user about the conflict and fails", func() {
							session := helpers.CF("install-plugin", "-f", pluginPath)

							Eventually(session.Out).Should(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
							Eventually(session.Out).Should(Say("Install and use plugins at your own risk\\."))

							Eventually(session.Out).Should(Say("FAILED"))
							Eventually(session.Err).Should(Say("Plugin new-plugin v1\\.1\\.1 could not be installed as it contains commands with names and aliases that are already used: existing-command, existing-alias\\."))

							Eventually(session).Should(Exit(1))
						})
					})
				})
			})

			Context("when the -f flag is not given", func() {
				Context("when the user says yes", func() {
					BeforeEach(func() {
						buffer = NewBuffer()
						_, _ = buffer.Write([]byte("y\n"))
					})

					It("installs the plugin", func() {
						session := helpers.CFWithStdin(buffer, "install-plugin", pluginPath)

						Eventually(session.Out).Should(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
						Eventually(session.Out).Should(Say("Install and use plugins at your own risk\\."))
						Eventually(session.Out).Should(Say("Do you want to install the plugin %s\\? \\[yN\\]: y", helpers.ConvertPathToRegularExpression(pluginPath)))
						Eventually(session.Out).Should(Say("Installing plugin some-plugin\\.\\.\\."))
						Eventually(session.Out).Should(Say("OK"))
						Eventually(session.Out).Should(Say("Plugin some-plugin 1\\.0\\.0 successfully installed\\."))

						Eventually(session).Should(Exit(0))

						pluginsSession := helpers.CF("plugins", "--checksum")
						expectedSha := helpers.Sha1Sum(
							generic.ExecutableFilename(filepath.Join(homeDir, ".cf/plugins/some-plugin")))
						Eventually(pluginsSession.Out).Should(Say("some-plugin\\s+1.0.0\\s+%s", expectedSha))
						Eventually(pluginsSession).Should(Exit(0))

						Eventually(helpers.CF("some-command")).Should(Exit(0))

						helpSession := helpers.CF("help")
						Eventually(helpSession.Out).Should(Say("some-command"))
						Eventually(helpSession).Should(Exit(0))
					})

					Context("when the plugin is already installed", func() {
						BeforeEach(func() {
							Eventually(helpers.CF("install-plugin", pluginPath, "-f")).Should(Exit(0))
						})

						It("fails and tells the user how to force a reinstall", func() {
							session := helpers.CFWithStdin(buffer, "install-plugin", pluginPath)

							Eventually(session.Out).Should(Say("FAILED"))
							Eventually(session.Err).Should(Say("Plugin some-plugin 1\\.0\\.0 could not be installed\\. A plugin with that name is already installed\\."))
							Eventually(session.Err).Should(Say("TIP: Use 'cf install-plugin -f' to force a reinstall\\."))

							Eventually(session).Should(Exit(1))
						})
					})
				})

				Context("when the user says no", func() {
					BeforeEach(func() {
						buffer = NewBuffer()
						_, _ = buffer.Write([]byte("n\n"))
					})

					It("does not install the plugin", func() {
						session := helpers.CFWithStdin(buffer, "install-plugin", pluginPath)

						Eventually(session.Out).Should(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
						Eventually(session.Out).Should(Say("Install and use plugins at your own risk\\."))
						Eventually(session.Out).Should(Say("Do you want to install the plugin %s\\? \\[yN\\]: n", helpers.ConvertPathToRegularExpression(pluginPath)))
						Eventually(session.Out).Should(Say("Plugin installation cancelled\\."))

						Eventually(session).Should(Exit(0))
					})

					Context("when the plugin is already installed", func() {
						BeforeEach(func() {
							Eventually(helpers.CF("install-plugin", pluginPath, "-f")).Should(Exit(0))
						})

						It("does not uninstall the existing plugin", func() {
							session := helpers.CFWithStdin(buffer, "install-plugin", pluginPath)

							Eventually(session.Out).Should(Say("Plugin installation cancelled\\."))

							Consistently(session.Out).ShouldNot(Say("Plugin some-plugin 1\\.0\\.0 is already installed\\. Uninstalling existing plugin\\.\\.\\."))
							Consistently(session.Out).ShouldNot(Say("CLI-MESSAGE-UNINSTALL"))
							Consistently(session.Out).ShouldNot(Say("Plugin some-plugin successfully uninstalled\\."))

							Eventually(session).Should(Exit(0))
						})
					})
				})

				Context("when the user interrupts with control-c", func() {
					BeforeEach(func() {
						buffer = NewBuffer()
						_, _ = buffer.Write([]byte("y")) // but not enter
					})

					It("does not install the plugin and does not create a bad state", func() {
						session := helpers.CFWithStdin(buffer, "install-plugin", pluginPath)

						Eventually(session.Out).Should(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
						Eventually(session.Out).Should(Say("Install and use plugins at your own risk\\."))
						Eventually(session.Out).Should(Say("Do you want to install the plugin %s\\? \\[yN\\]:", helpers.ConvertPathToRegularExpression(pluginPath)))

						session.Interrupt()

						Eventually(session.Out).Should(Say("FAILED"))

						// There is a timing issue -- the exit code may be either 1 (processed error), 2 (config writing error), or 130 (Ctrl-C)
						Eventually(session).Should(SatisfyAny(Exit(1), Exit(2), Exit(130)))

						// make sure cf plugins did not break
						Eventually(helpers.CF("plugins", "--checksum")).Should(Exit(0))

						// make sure a retry of the plugin install works
						retrySession := helpers.CF("install-plugin", pluginPath, "-f")
						Eventually(retrySession.Out).Should(Say("Plugin some-plugin 1\\.0\\.0 successfully installed\\."))
						Eventually(retrySession).Should(Exit(0))
					})
				})
			})
		})
	})

	Describe("installing a plugin from a URL", func() {
		var (
			server *Server
		)

		BeforeEach(func() {
			server = NewTLSServer()
			// Suppresses ginkgo server logs
			server.HTTPTestServer.Config.ErrorLog = log.New(&bytes.Buffer{}, "", 0)
		})

		AfterEach(func() {
			server.Close()
		})

		Context("when a URL and the -f flag are provided", func() {
			Context("when an executable is available for download at the URL", func() {
				var (
					pluginData []byte
				)

				BeforeEach(func() {
					pluginPath = helpers.BuildConfigurablePlugin("configurable_plugin", "some-plugin", "1.0.0",
						[]helpers.PluginCommand{
							{Name: "some-command", Help: "some-command-help"},
						},
					)

					var err error
					pluginData, err = ioutil.ReadFile(pluginPath)
					Expect(err).ToNot(HaveOccurred())
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/"),
							RespondWith(http.StatusOK, pluginData),
						),
					)
				})

				AfterEach(func() {
					err := os.Remove(pluginPath)
					Expect(err).ToNot(HaveOccurred())
				})

				It("installs the plugin", func() {
					session := helpers.CF("install-plugin", "-f", server.URL(), "-k")

					Eventually(session.Out).Should(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
					Eventually(session.Out).Should(Say("Install and use plugins at your own risk\\."))

					Eventually(session.Out).Should(Say("Starting download of plugin binary from URL\\.\\.\\."))
					Eventually(session.Out).Should(Say("\\d.* .*B / ?"))

					Eventually(session.Out).Should(Say("Installing plugin some-plugin\\.\\.\\."))
					Eventually(session.Out).Should(Say("OK"))
					Eventually(session.Out).Should(Say("Plugin some-plugin 1\\.0\\.0 successfully installed\\."))

					Eventually(session).Should(Exit(0))
				})

				Context("when the URL redirects", func() {
					BeforeEach(func() {
						server.Reset()
						server.AppendHandlers(
							CombineHandlers(
								VerifyRequest(http.MethodGet, "/redirect"),
								RespondWith(http.StatusMovedPermanently, nil, http.Header{"Location": []string{server.URL()}}),
							),
							CombineHandlers(
								VerifyRequest(http.MethodGet, "/"),
								RespondWith(http.StatusOK, pluginData),
							))
					})

					It("installs the plugin", func() {
						session := helpers.CF("install-plugin", "-f", fmt.Sprintf("%s/redirect", server.URL()), "-k")

						Eventually(session.Out).Should(Say("Installing plugin some-plugin\\.\\.\\."))
						Eventually(session.Out).Should(Say("OK"))
						Eventually(session.Out).Should(Say("Plugin some-plugin 1\\.0\\.0 successfully installed\\."))

						Eventually(session).Should(Exit(0))
					})
				})

				Context("when the plugin has already been installed", func() {
					BeforeEach(func() {
						Eventually(helpers.CF("install-plugin", pluginPath, "-f")).Should(Exit(0))
					})

					It("uninstalls and reinstalls the plugin", func() {
						session := helpers.CF("install-plugin", "-f", server.URL(), "-k")

						Eventually(session.Out).Should(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
						Eventually(session.Out).Should(Say("Install and use plugins at your own risk\\."))

						Eventually(session.Out).Should(Say("Starting download of plugin binary from URL\\.\\.\\."))
						Eventually(session.Out).Should(Say("\\d.* .*B / ?"))

						Eventually(session.Out).Should(Say("Plugin some-plugin 1\\.0\\.0 is already installed\\. Uninstalling existing plugin\\.\\.\\."))
						Eventually(session.Out).Should(Say("CLI-MESSAGE-UNINSTALL"))
						Eventually(session.Out).Should(Say("Plugin some-plugin successfully uninstalled\\."))
						Eventually(session.Out).Should(Say("OK"))
						Eventually(session.Out).Should(Say("Plugin some-plugin 1\\.0\\.0 successfully installed\\."))

						Eventually(session).Should(Exit(0))
					})
				})
			})

			Context("when a 4xx or 5xx HTTP response status is encountered", func() {
				BeforeEach(func() {
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/"),
							RespondWith(http.StatusNotFound, nil),
						),
					)
				})

				It("displays an appropriate error", func() {
					session := helpers.CF("install-plugin", "-f", server.URL(), "-k")

					Eventually(session.Out).Should(Say("Starting download of plugin binary from URL\\.\\.\\."))
					Eventually(session.Out).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("Download attempt failed; server returned 404 Not Found"))
					Eventually(session.Err).Should(Say("Unable to install; plugin is not available from the given URL\\."))

					Eventually(session).Should(Exit(1))
				})
			})

			Context("when the file is not a plugin", func() {
				BeforeEach(func() {
					var err error
					pluginPath, err = Build("code.cloudfoundry.org/cli/integration/assets/non_plugin")
					Expect(err).ToNot(HaveOccurred())

					pluginData, err := ioutil.ReadFile(pluginPath)
					Expect(err).ToNot(HaveOccurred())
					server.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/"),
							RespondWith(http.StatusOK, pluginData),
						),
					)
				})

				AfterEach(func() {
					err := os.Remove(pluginPath)
					Expect(err).ToNot(HaveOccurred())
				})

				It("tells the user that the file is not a plugin and fails", func() {
					session := helpers.CF("install-plugin", "-f", server.URL(), "-k")

					Eventually(session.Out).Should(Say("Starting download of plugin binary from URL\\.\\.\\."))
					Eventually(session.Out).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("File is not a valid cf CLI plugin binary\\."))

					Eventually(session).Should(Exit(1))
				})
			})
		})

		Context("when the -f flag is not provided", func() {
			var (
				pluginData []byte
			)

			BeforeEach(func() {
				pluginPath = helpers.BuildConfigurablePlugin("configurable_plugin", "some-plugin", "1.0.0",
					[]helpers.PluginCommand{
						{Name: "some-command", Help: "some-command-help"},
					},
				)

				var err error
				pluginData, err = ioutil.ReadFile(pluginPath)
				Expect(err).ToNot(HaveOccurred())
				server.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/"),
						RespondWith(http.StatusOK, pluginData),
					),
				)
			})

			AfterEach(func() {
				err := os.Remove(pluginPath)
				Expect(err).ToNot(HaveOccurred())
			})

			Context("when the user says yes", func() {
				BeforeEach(func() {
					buffer = NewBuffer()
					_, _ = buffer.Write([]byte("y\n"))
				})

				It("installs the plugin", func() {
					session := helpers.CFWithStdin(buffer, "install-plugin", server.URL(), "-k")

					Eventually(session.Out).Should(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
					Eventually(session.Out).Should(Say("Install and use plugins at your own risk\\."))
					Eventually(session.Out).Should(Say("Do you want to install the plugin %s\\? \\[yN\\]: y", server.URL()))

					Eventually(session.Out).Should(Say("Starting download of plugin binary from URL\\.\\.\\."))
					Eventually(session.Out).Should(Say("\\d.* .*B / ?"))

					Eventually(session.Out).Should(Say("Installing plugin some-plugin\\.\\.\\."))
					Eventually(session.Out).Should(Say("OK"))
					Eventually(session.Out).Should(Say("Plugin some-plugin 1\\.0\\.0 successfully installed\\."))

					Eventually(session).Should(Exit(0))
				})

				Context("when the plugin is already installed", func() {
					BeforeEach(func() {
						Eventually(helpers.CF("install-plugin", pluginPath, "-f")).Should(Exit(0))
					})

					It("fails and tells the user how to force a reinstall", func() {
						session := helpers.CFWithStdin(buffer, "install-plugin", server.URL(), "-k")

						Eventually(session.Out).Should(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
						Eventually(session.Out).Should(Say("Install and use plugins at your own risk\\."))
						Eventually(session.Out).Should(Say("Do you want to install the plugin %s\\? \\[yN\\]: y", server.URL()))

						Eventually(session.Out).Should(Say("Starting download of plugin binary from URL\\.\\.\\."))
						Eventually(session.Out).Should(Say("\\d.* .*B / ?"))

						Eventually(session.Out).Should(Say("FAILED"))
						Eventually(session.Err).Should(Say("Plugin some-plugin 1\\.0\\.0 could not be installed\\. A plugin with that name is already installed\\."))
						Eventually(session.Err).Should(Say("TIP: Use 'cf install-plugin -f' to force a reinstall\\."))
						Eventually(session).Should(Exit(1))
					})
				})
			})

			Context("when the user says no", func() {
				BeforeEach(func() {
					buffer = NewBuffer()
					_, _ = buffer.Write([]byte("n\n"))
				})

				It("does not install the plugin", func() {
					session := helpers.CFWithStdin(buffer, "install-plugin", server.URL())

					Eventually(session.Out).Should(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
					Eventually(session.Out).Should(Say("Install and use plugins at your own risk\\."))
					Eventually(session.Out).Should(Say("Do you want to install the plugin %s\\? \\[yN\\]: n", server.URL()))
					Eventually(session.Out).Should(Say("Plugin installation cancelled\\."))

					Eventually(session).Should(Exit(0))

					Expect(server.ReceivedRequests()).To(HaveLen(0))
				})
			})

			Context("when the user interrupts with control-c", func() {
				BeforeEach(func() {
					buffer = NewBuffer()
					_, _ = buffer.Write([]byte("y")) // but not enter
				})

				It("does not install the plugin and does not create a bad state", func() {
					session := helpers.CFWithStdin(buffer, "install-plugin", pluginPath)

					Eventually(session.Out).Should(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
					Eventually(session.Out).Should(Say("Install and use plugins at your own risk\\."))
					Eventually(session.Out).Should(Say("Do you want to install the plugin %s\\? \\[yN\\]:", helpers.ConvertPathToRegularExpression(pluginPath)))

					session.Interrupt()

					Eventually(session.Out).Should(Say("FAILED"))

					// There is a timing issue -- the exit code may be either 1 (processed error), 2 (config writing error), or 130 (Ctrl-C)
					Eventually(session).Should(SatisfyAny(Exit(1), Exit(2), Exit(130)))

					Expect(server.ReceivedRequests()).To(HaveLen(0))

					// make sure cf plugins did not break
					Eventually(helpers.CF("plugins", "--checksum")).Should(Exit(0))

					// make sure a retry of the plugin install works
					retrySession := helpers.CF("install-plugin", pluginPath, "-f")
					Eventually(retrySession.Out).Should(Say("Plugin some-plugin 1\\.0\\.0 successfully installed\\."))
					Eventually(retrySession).Should(Exit(0))
				})
			})
		})
	})
})
