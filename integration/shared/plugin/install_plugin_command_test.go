package plugin

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/util/generic"
	. "github.com/onsi/ginkgo/v2"
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
		pluginsHomeDirContents, err := os.ReadDir(filepath.Join(homeDir, ".cf", "plugins"))
		if os.IsNotExist(err) {
			return
		}

		Expect(err).ToNot(HaveOccurred())

		for _, entry := range pluginsHomeDirContents {
			Expect(entry.Name()).NotTo(ContainSubstring("temp"))
		}
	})

	Describe("help", func() {
		When("the --help flag is given", func() {
			It("displays command usage to stdout", func() {
				session := helpers.CF("install-plugin", "--help")

				Eventually(session).Should(Say("NAME:"))
				Eventually(session).Should(Say("install-plugin - Install CLI plugin"))
				Eventually(session).Should(Say("USAGE:"))
				Eventually(session).Should(Say(`cf install-plugin PLUGIN_NAME \[-r REPO_NAME\] \[-f\]`))
				Eventually(session).Should(Say(`cf install-plugin LOCAL-PATH/TO/PLUGIN | URL \[-f\]`))
				Eventually(session).Should(Say(""))
				Eventually(session).Should(Say("WARNING:"))
				Eventually(session).Should(Say("Plugins are binaries written by potentially untrusted authors."))
				Eventually(session).Should(Say("Install and use plugins at your own risk."))
				Eventually(session).Should(Say(""))
				Eventually(session).Should(Say("EXAMPLES:"))
				Eventually(session).Should(Say("cf install-plugin ~/Downloads/plugin-foobar"))
				Eventually(session).Should(Say("cf install-plugin https://example.com/plugin-foobar_linux_amd64"))
				Eventually(session).Should(Say("cf install-plugin -r My-Repo plugin-echo"))
				Eventually(session).Should(Say("OPTIONS:"))
				Eventually(session).Should(Say(`-f\s+Force install of plugin without confirmation`))
				Eventually(session).Should(Say(`-r\s+Restrict search for plugin to this registered repository`))
				Eventually(session).Should(Say("SEE ALSO:"))
				Eventually(session).Should(Say("add-plugin-repo, list-plugin-repos, plugins"))

				Eventually(session).Should(Exit(0))
			})
		})
	})

	When("the user does not provide a plugin name or location", func() {
		It("errors and displays usage", func() {
			session := helpers.CF("install-plugin")
			Eventually(session.Err).Should(Say("Incorrect Usage: the required argument `PLUGIN_NAME_OR_LOCATION` was not provided"))
			Eventually(session).Should(Say("USAGE:"))

			Eventually(session).Should(Exit(1))
		})
	})

	When("the plugin dir does not exist", func() {
		var (
			newPluginHome string

			cfPluginHome string
		)

		BeforeEach(func() {
			cfPluginHome = os.Getenv("CF_PLUGIN_HOME")

			var err error
			newPluginHome, err = os.MkdirTemp("", "plugin-temp-dir")
			Expect(err).ToNot(HaveOccurred())
			Expect(os.RemoveAll(newPluginHome))

			Expect(os.Setenv("CF_PLUGIN_HOME", newPluginHome)).ToNot(HaveOccurred())

			pluginPath = helpers.BuildConfigurablePlugin("configurable_plugin", "some-plugin", "1.0.0",
				[]helpers.PluginCommand{
					{Name: "some-command", Help: "some-command-help"},
				},
			)
		})

		AfterEach(func() {
			Expect(os.RemoveAll(newPluginHome)).ToNot(HaveOccurred())
			Expect(os.Setenv("CF_PLUGIN_HOME", cfPluginHome)).ToNot(HaveOccurred())
		})

		It("creates the new directory, and continues as normal", func() {
			session := helpers.CF("install-plugin", pluginPath, "-f")
			Eventually(session).Should(Exit(0))

			log.Println(newPluginHome)
			_, err := os.Stat(newPluginHome)
			Expect(os.IsNotExist(err)).To(Equal(false))
		})
	})

	Describe("installing a plugin from a local file", func() {
		When("the file is compiled for a different os and architecture", func() {
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

				Eventually(session).Should(Say(`Attention: Plugins are binaries written by potentially untrusted authors\.`))
				Eventually(session).Should(Say(`Install and use plugins at your own risk\.`))
				Eventually(session).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say(`File is not a valid cf CLI plugin binary\.`))

				Eventually(session).Should(Exit(1))
			})
		})

		When("the file is compiled for the correct os and architecture", func() {
			BeforeEach(func() {
				pluginPath = helpers.BuildConfigurablePlugin("configurable_plugin", "some-plugin", "1.0.0",
					[]helpers.PluginCommand{
						{Name: "some-command", Help: "some-command-help"},
					},
				)
			})

			When("the -f flag is given", func() {
				It("installs the plugin and cleans up all temp files", func() {
					session := helpers.CF("install-plugin", pluginPath, "-f")

					Eventually(session).Should(Say(`Attention: Plugins are binaries written by potentially untrusted authors\.`))
					Eventually(session).Should(Say(`Install and use plugins at your own risk\.`))
					Eventually(session).Should(Say(`Installing plugin some-plugin\.\.\.`))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Say(`Plugin some-plugin 1\.0\.0 successfully installed\.`))

					Eventually(session).Should(Exit(0))

					installedPath := generic.ExecutableFilename(filepath.Join(homeDir, ".cf", "plugins", "some-plugin"))

					pluginsSession := helpers.CF("plugins", "--checksum")
					expectedSha := helpers.Sha1Sum(installedPath)

					Eventually(pluginsSession).Should(Say(`some-plugin\s+1\.0\.0\s+%s`, expectedSha))
					Eventually(pluginsSession).Should(Exit(0))

					Eventually(helpers.CF("some-command")).Should(Exit(0))

					helpSession := helpers.CF("help")
					Eventually(helpSession).Should(Say("some-command"))
					Eventually(helpSession).Should(Exit(0))
				})

				When("the file does not have executable permissions", func() {
					BeforeEach(func() {
						Expect(os.Chmod(pluginPath, 0666)).ToNot(HaveOccurred())
					})

					It("installs the plugin", func() {
						session := helpers.CF("install-plugin", pluginPath, "-f")
						Eventually(session).Should(Say(`Plugin some-plugin 1\.0\.0 successfully installed\.`))
						Eventually(session).Should(Exit(0))
					})
				})

				When("the plugin is already installed", func() {
					BeforeEach(func() {
						Eventually(helpers.CF("install-plugin", pluginPath, "-f")).Should(Exit(0))
					})

					It("uninstalls the existing plugin and installs the plugin", func() {
						session := helpers.CF("install-plugin", pluginPath, "-f")

						Eventually(session).Should(Say(`Plugin some-plugin 1\.0\.0 is already installed\. Uninstalling existing plugin\.\.\.`))
						Eventually(session).Should(Say("CLI-MESSAGE-UNINSTALL"))
						Eventually(session).Should(Say(`Plugin some-plugin successfully uninstalled\.`))
						Eventually(session).Should(Say(`Plugin some-plugin 1\.0\.0 successfully installed\.`))

						Eventually(session).Should(Exit(0))
					})
				})

				When("the file does not exist", func() {
					It("tells the user that the file was not found and fails", func() {
						session := helpers.CF("install-plugin", "some/path/that/does/not/exist", "-f")
						Eventually(session.Err).Should(Say(`Plugin some/path/that/does/not/exist not found on disk or in any registered repo\.`))
						Eventually(session.Err).Should(Say(`Use 'cf repo-plugins' to list plugins available in the repos\.`))

						Consistently(session).ShouldNot(Say(`Attention: Plugins are binaries written by potentially untrusted authors\.`))
						Consistently(session).ShouldNot(Say(`Install and use plugins at your own risk\.`))

						Eventually(session).Should(Exit(1))
					})
				})

				When("the file is not an executable", func() {
					BeforeEach(func() {
						badPlugin, err := os.CreateTemp("", "")
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
						Eventually(session.Err).Should(Say(`File is not a valid cf CLI plugin binary\.`))

						Eventually(session).Should(Exit(1))
					})
				})

				When("the file is not a plugin", func() {
					BeforeEach(func() {
						var err error
						pluginPath, err = Build("code.cloudfoundry.org/cli/integration/assets/non_plugin")
						Expect(err).ToNot(HaveOccurred())
					})

					It("tells the user that the file is not a plugin and fails", func() {
						session := helpers.CF("install-plugin", pluginPath, "-f")
						Eventually(session.Err).Should(Say(`File is not a valid cf CLI plugin binary\.`))

						Eventually(session).Should(Exit(1))
					})
				})

				When("getting metadata from the plugin errors", func() {
					BeforeEach(func() {
						var err error
						pluginPath, err = Build("code.cloudfoundry.org/cli/integration/assets/test_plugin_fails_metadata")
						Expect(err).ToNot(HaveOccurred())
					})

					It("displays the error to stderr", func() {
						session := helpers.CF("install-plugin", pluginPath, "-f")
						Eventually(session.Err).Should(Say("exit status 51"))
						Eventually(session.Err).Should(Say(`File is not a valid cf CLI plugin binary\.`))

						Eventually(session).Should(Exit(1))
					})
				})

				When("there is a command conflict", func() {
					When("the plugin has a command that is the same as a built-in command", func() {
						BeforeEach(func() {
							pluginPath = helpers.BuildConfigurablePlugin(
								"configurable_plugin", "some-plugin", "1.1.1",
								[]helpers.PluginCommand{
									{Name: "version"},
								})
						})

						It("tells the user about the conflict and fails", func() {
							session := helpers.CF("install-plugin", "-f", pluginPath)

							Eventually(session).Should(Say(`Attention: Plugins are binaries written by potentially untrusted authors\.`))
							Eventually(session).Should(Say(`Install and use plugins at your own risk\.`))

							Eventually(session).Should(Say("FAILED"))
							Eventually(session.Err).Should(Say(`Plugin some-plugin v1\.1\.1 could not be installed as it contains commands with names that are already used: version`))

							Eventually(session).Should(Exit(1))
						})
					})

					When("the plugin has a command that is the same as a built-in alias", func() {
						BeforeEach(func() {
							pluginPath = helpers.BuildConfigurablePlugin(
								"configurable_plugin", "some-plugin", "1.1.1",
								[]helpers.PluginCommand{
									{Name: "cups"},
								})
						})

						It("tells the user about the conflict and fails", func() {
							session := helpers.CF("install-plugin", "-f", pluginPath)

							Eventually(session).Should(Say(`Attention: Plugins are binaries written by potentially untrusted authors\.`))
							Eventually(session).Should(Say(`Install and use plugins at your own risk\.`))

							Eventually(session).Should(Say("FAILED"))
							Eventually(session.Err).Should(Say(`Plugin some-plugin v1\.1\.1 could not be installed as it contains commands with names that are already used: cups`))

							Eventually(session).Should(Exit(1))
						})
					})

					When("the plugin has a command that is the same as another plugin command", func() {
						BeforeEach(func() {
							helpers.InstallConfigurablePlugin("configurable_plugin", "existing-plugin", "1.1.1",
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

							Eventually(session).Should(Say(`Attention: Plugins are binaries written by potentially untrusted authors\.`))
							Eventually(session).Should(Say(`Install and use plugins at your own risk\.`))

							Eventually(session).Should(Say("FAILED"))
							Eventually(session.Err).Should(Say(`Plugin new-plugin v1\.1\.1 could not be installed as it contains commands with names that are already used: existing-command\.`))

							Eventually(session).Should(Exit(1))
						})
					})

					When("the plugin has a command that is the same as another plugin alias", func() {
						BeforeEach(func() {
							helpers.InstallConfigurablePlugin("configurable_plugin", "existing-plugin", "1.1.1",
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

							Eventually(session).Should(Say(`Attention: Plugins are binaries written by potentially untrusted authors\.`))
							Eventually(session).Should(Say(`Install and use plugins at your own risk\.`))

							Eventually(session).Should(Say("FAILED"))
							Eventually(session.Err).Should(Say(`Plugin new-plugin v1\.1\.1 could not be installed as it contains commands with aliases that are already used: existing-command\.`))

							Eventually(session).Should(Exit(1))
						})
					})
				})

				Context("alias conflict", func() {
					When("the plugin has an alias that is the same as a built-in command", func() {

						BeforeEach(func() {
							pluginPath = helpers.BuildConfigurablePlugin(
								"configurable_plugin", "some-plugin", "1.1.1",
								[]helpers.PluginCommand{
									{Name: "some-command", Alias: "version"},
								})
						})

						It("tells the user about the conflict and fails", func() {
							session := helpers.CF("install-plugin", "-f", pluginPath)

							Eventually(session).Should(Say(`Attention: Plugins are binaries written by potentially untrusted authors\.`))
							Eventually(session).Should(Say(`Install and use plugins at your own risk\.`))

							Eventually(session).Should(Say("FAILED"))
							Eventually(session.Err).Should(Say(`Plugin some-plugin v1\.1\.1 could not be installed as it contains commands with aliases that are already used: version`))

							Eventually(session).Should(Exit(1))
						})
					})

					When("the plugin has an alias that is the same as a built-in alias", func() {
						BeforeEach(func() {
							pluginPath = helpers.BuildConfigurablePlugin(
								"configurable_plugin", "some-plugin", "1.1.1",
								[]helpers.PluginCommand{
									{Name: "some-command", Alias: "cups"},
								})
						})

						It("tells the user about the conflict and fails", func() {
							session := helpers.CF("install-plugin", "-f", pluginPath)

							Eventually(session).Should(Say(`Attention: Plugins are binaries written by potentially untrusted authors\.`))
							Eventually(session).Should(Say(`Install and use plugins at your own risk\.`))

							Eventually(session).Should(Say("FAILED"))
							Eventually(session.Err).Should(Say(`Plugin some-plugin v1\.1\.1 could not be installed as it contains commands with aliases that are already used: cups`))

							Eventually(session).Should(Exit(1))
						})
					})

					When("the plugin has an alias that is the same as another plugin command", func() {
						BeforeEach(func() {
							helpers.InstallConfigurablePlugin("configurable_plugin", "existing-plugin", "1.1.1",
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

							Eventually(session).Should(Say(`Attention: Plugins are binaries written by potentially untrusted authors\.`))
							Eventually(session).Should(Say(`Install and use plugins at your own risk\.`))

							Eventually(session).Should(Say("FAILED"))
							Eventually(session.Err).Should(Say(`Plugin new-plugin v1\.1\.1 could not be installed as it contains commands with aliases that are already used: existing-command\.`))

							Eventually(session).Should(Exit(1))
						})
					})

					When("the plugin has an alias that is the same as another plugin alias", func() {
						BeforeEach(func() {
							helpers.InstallConfigurablePlugin("configurable_plugin", "existing-plugin", "1.1.1",
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

							Eventually(session).Should(Say(`Attention: Plugins are binaries written by potentially untrusted authors\.`))
							Eventually(session).Should(Say(`Install and use plugins at your own risk\.`))

							Eventually(session).Should(Say("FAILED"))
							Eventually(session.Err).Should(Say(`Plugin new-plugin v1\.1\.1 could not be installed as it contains commands with aliases that are already used: existing-alias\.`))

							Eventually(session).Should(Exit(1))
						})
					})
				})

				Context("alias and command conflicts", func() {
					When("the plugin has a command and an alias that are both taken by another plugin", func() {
						BeforeEach(func() {
							helpers.InstallConfigurablePlugin("configurable_plugin", "existing-plugin", "1.1.1",
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

							Eventually(session).Should(Say(`Attention: Plugins are binaries written by potentially untrusted authors\.`))
							Eventually(session).Should(Say(`Install and use plugins at your own risk\.`))

							Eventually(session).Should(Say("FAILED"))
							Eventually(session.Err).Should(Say(`Plugin new-plugin v1\.1\.1 could not be installed as it contains commands with names and aliases that are already used: existing-command, existing-alias\.`))

							Eventually(session).Should(Exit(1))
						})
					})
				})
			})

			When("the -f flag is not given", func() {
				When("the user says yes", func() {
					BeforeEach(func() {
						buffer = NewBuffer()
						_, err := buffer.Write([]byte("y\n"))
						Expect(err).ToNot(HaveOccurred())
					})

					It("installs the plugin", func() {
						session := helpers.CFWithStdin(buffer, "install-plugin", pluginPath)

						Eventually(session).Should(Say(`Attention: Plugins are binaries written by potentially untrusted authors\.`))
						Eventually(session).Should(Say(`Install and use plugins at your own risk\.`))
						Eventually(session).Should(Say(`Do you want to install the plugin %s\? \[yN\]: y`, helpers.ConvertPathToRegularExpression(pluginPath)))
						Eventually(session).Should(Say(`Installing plugin some-plugin\.\.\.`))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Say(`Plugin some-plugin 1\.0\.0 successfully installed\.`))

						Eventually(session).Should(Exit(0))

						pluginsSession := helpers.CF("plugins", "--checksum")
						expectedSha := helpers.Sha1Sum(
							generic.ExecutableFilename(filepath.Join(homeDir, ".cf/plugins/some-plugin")))
						Eventually(pluginsSession).Should(Say(`some-plugin\s+1.0.0\s+%s`, expectedSha))
						Eventually(pluginsSession).Should(Exit(0))

						Eventually(helpers.CF("some-command")).Should(Exit(0))

						helpSession := helpers.CF("help")
						Eventually(helpSession).Should(Say("some-command"))
						Eventually(helpSession).Should(Exit(0))
					})

					When("the plugin is already installed", func() {
						BeforeEach(func() {
							Eventually(helpers.CF("install-plugin", pluginPath, "-f")).Should(Exit(0))
						})

						It("fails and tells the user how to force a reinstall", func() {
							session := helpers.CFWithStdin(buffer, "install-plugin", pluginPath)

							Eventually(session).Should(Say("FAILED"))
							Eventually(session.Err).Should(Say(`Plugin some-plugin 1\.0\.0 could not be installed\. A plugin with that name is already installed\.`))
							Eventually(session.Err).Should(Say(`TIP: Use 'cf install-plugin -f' to force a reinstall\.`))

							Eventually(session).Should(Exit(1))
						})
					})
				})

				When("the user says no", func() {
					BeforeEach(func() {
						buffer = NewBuffer()
						_, err := buffer.Write([]byte("n\n"))
						Expect(err).ToNot(HaveOccurred())
					})

					It("does not install the plugin", func() {
						session := helpers.CFWithStdin(buffer, "install-plugin", pluginPath)

						Eventually(session).Should(Say(`Attention: Plugins are binaries written by potentially untrusted authors\.`))
						Eventually(session).Should(Say(`Install and use plugins at your own risk\.`))
						Eventually(session).Should(Say(`Do you want to install the plugin %s\? \[yN\]: n`, helpers.ConvertPathToRegularExpression(pluginPath)))
						Eventually(session).Should(Say(`Plugin installation cancelled\.`))

						Eventually(session).Should(Exit(0))
					})

					When("the plugin is already installed", func() {
						BeforeEach(func() {
							Eventually(helpers.CF("install-plugin", pluginPath, "-f")).Should(Exit(0))
						})

						It("does not uninstall the existing plugin", func() {
							session := helpers.CFWithStdin(buffer, "install-plugin", pluginPath)

							Eventually(session).Should(Say(`Plugin installation cancelled\.`))

							Consistently(session).ShouldNot(Say(`Plugin some-plugin 1\.0\.0 is already installed\. Uninstalling existing plugin\.\.\.`))
							Consistently(session).ShouldNot(Say("CLI-MESSAGE-UNINSTALL"))
							Consistently(session).ShouldNot(Say(`Plugin some-plugin successfully uninstalled\.`))

							Eventually(session).Should(Exit(0))
						})
					})
				})

				When("the user interrupts with control-c", func() {
					BeforeEach(func() {
						buffer = NewBuffer()
						_, err := buffer.Write([]byte("y")) // but not enter
						Expect(err).ToNot(HaveOccurred())
					})

					It("does not install the plugin and does not create a bad state", func() {
						session := helpers.CFWithStdin(buffer, "install-plugin", pluginPath)

						Eventually(session).Should(Say(`Attention: Plugins are binaries written by potentially untrusted authors\.`))
						Eventually(session).Should(Say(`Install and use plugins at your own risk\.`))
						Eventually(session).Should(Say(`Do you want to install the plugin %s\? \[yN\]:`, helpers.ConvertPathToRegularExpression(pluginPath)))

						session.Interrupt()
						// There is a timing issue -- the exit code may be either 1 (processed error), 2 (config writing error), or 130 (Ctrl-C)
						Eventually(session).Should(SatisfyAny(Exit(1), Exit(2), Exit(130)))

						Expect(session).Should(Say("FAILED"))

						// make sure cf plugins did not break
						Eventually(helpers.CF("plugins", "--checksum")).Should(Exit(0))

						// make sure a retry of the plugin install works
						retrySession := helpers.CF("install-plugin", pluginPath, "-f")
						Eventually(retrySession).Should(Exit(0))
						Expect(retrySession).To(Say(`Plugin some-plugin 1\.0\.0 successfully installed\.`))
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

		When("a URL and the -f flag are provided", func() {
			When("an executable is available for download at the URL", func() {
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
					pluginData, err = os.ReadFile(pluginPath)
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

					Eventually(session).Should(Say(`Attention: Plugins are binaries written by potentially untrusted authors\.`))
					Eventually(session).Should(Say(`Install and use plugins at your own risk\.`))

					Eventually(session).Should(Say(`Starting download of plugin binary from URL\.\.\.`))
					Eventually(session).Should(Say(`\d.* .*B / ?`))

					Eventually(session).Should(Say(`Installing plugin some-plugin\.\.\.`))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Say(`Plugin some-plugin 1\.0\.0 successfully installed\.`))

					Eventually(session).Should(Exit(0))
				})

				When("the URL redirects", func() {
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

						Eventually(session).Should(Say(`Installing plugin some-plugin\.\.\.`))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Say(`Plugin some-plugin 1\.0\.0 successfully installed\.`))

						Eventually(session).Should(Exit(0))
					})
				})

				When("the plugin has already been installed", func() {
					BeforeEach(func() {
						Eventually(helpers.CF("install-plugin", pluginPath, "-f")).Should(Exit(0))
					})

					It("uninstalls and reinstalls the plugin", func() {
						session := helpers.CF("install-plugin", "-f", server.URL(), "-k")

						Eventually(session).Should(Say(`Attention: Plugins are binaries written by potentially untrusted authors\.`))
						Eventually(session).Should(Say(`Install and use plugins at your own risk\.`))

						Eventually(session).Should(Say(`Starting download of plugin binary from URL\.\.\.`))
						Eventually(session).Should(Say(`\d.* .*B / ?`))

						Eventually(session).Should(Say(`Plugin some-plugin 1\.0\.0 is already installed\. Uninstalling existing plugin\.\.\.`))
						Eventually(session).Should(Say("CLI-MESSAGE-UNINSTALL"))
						Eventually(session).Should(Say(`Plugin some-plugin successfully uninstalled\.`))
						Eventually(session).Should(Say("OK"))
						Eventually(session).Should(Say(`Plugin some-plugin 1\.0\.0 successfully installed\.`))

						Eventually(session).Should(Exit(0))
					})
				})
			})

			When("a 4xx or 5xx HTTP response status is encountered", func() {
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

					Eventually(session).Should(Say(`Starting download of plugin binary from URL\.\.\.`))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("Download attempt failed; server returned 404 Not Found"))
					Eventually(session.Err).Should(Say(`Unable to install; plugin is not available from the given URL\.`))

					Eventually(session).Should(Exit(1))
				})
			})

			When("the file is not a plugin", func() {
				BeforeEach(func() {
					var err error
					pluginPath, err = Build("code.cloudfoundry.org/cli/integration/assets/non_plugin")
					Expect(err).ToNot(HaveOccurred())

					pluginData, err := os.ReadFile(pluginPath)
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

					Eventually(session).Should(Say(`Starting download of plugin binary from URL\.\.\.`))
					Eventually(session).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say(`File is not a valid cf CLI plugin binary\.`))

					Eventually(session).Should(Exit(1))
				})
			})
		})

		When("the -f flag is not provided", func() {
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
				pluginData, err = os.ReadFile(pluginPath)
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

			When("the user says yes", func() {
				BeforeEach(func() {
					buffer = NewBuffer()
					_, err := buffer.Write([]byte("y\n"))
					Expect(err).ToNot(HaveOccurred())
				})

				It("installs the plugin", func() {
					session := helpers.CFWithStdin(buffer, "install-plugin", server.URL(), "-k")

					Eventually(session).Should(Say(`Attention: Plugins are binaries written by potentially untrusted authors\.`))
					Eventually(session).Should(Say(`Install and use plugins at your own risk\.`))
					Eventually(session).Should(Say(`Do you want to install the plugin %s\? \[yN\]: y`, server.URL()))

					Eventually(session).Should(Say(`Starting download of plugin binary from URL\.\.\.`))
					Eventually(session).Should(Say(`\d.* .*B / ?`))

					Eventually(session).Should(Say(`Installing plugin some-plugin\.\.\.`))
					Eventually(session).Should(Say("OK"))
					Eventually(session).Should(Say(`Plugin some-plugin 1\.0\.0 successfully installed\.`))

					Eventually(session).Should(Exit(0))
				})

				When("the plugin is already installed", func() {
					BeforeEach(func() {
						Eventually(helpers.CF("install-plugin", pluginPath, "-f")).Should(Exit(0))
					})

					It("fails and tells the user how to force a reinstall", func() {
						session := helpers.CFWithStdin(buffer, "install-plugin", server.URL(), "-k")

						Eventually(session).Should(Say(`Attention: Plugins are binaries written by potentially untrusted authors\.`))
						Eventually(session).Should(Say(`Install and use plugins at your own risk\.`))
						Eventually(session).Should(Say(`Do you want to install the plugin %s\? \[yN\]: y`, server.URL()))

						Eventually(session).Should(Say(`Starting download of plugin binary from URL\.\.\.`))
						Eventually(session).Should(Say(`\d.* .*B / ?`))

						Eventually(session).Should(Say("FAILED"))
						Eventually(session.Err).Should(Say(`Plugin some-plugin 1\.0\.0 could not be installed\. A plugin with that name is already installed\.`))
						Eventually(session.Err).Should(Say(`TIP: Use 'cf install-plugin -f' to force a reinstall\.`))
						Eventually(session).Should(Exit(1))
					})
				})
			})

			When("the user says no", func() {
				BeforeEach(func() {
					buffer = NewBuffer()
					_, err := buffer.Write([]byte("n\n"))
					Expect(err).ToNot(HaveOccurred())
				})

				It("does not install the plugin", func() {
					session := helpers.CFWithStdin(buffer, "install-plugin", server.URL())

					Eventually(session).Should(Say(`Attention: Plugins are binaries written by potentially untrusted authors\.`))
					Eventually(session).Should(Say(`Install and use plugins at your own risk\.`))
					Eventually(session).Should(Say(`Do you want to install the plugin %s\? \[yN\]: n`, server.URL()))
					Eventually(session).Should(Say(`Plugin installation cancelled\.`))

					Eventually(session).Should(Exit(0))

					Expect(server.ReceivedRequests()).To(HaveLen(0))
				})
			})

			When("the user interrupts with control-c", func() {
				BeforeEach(func() {
					buffer = NewBuffer()
					_, err := buffer.Write([]byte("y")) // but not enter
					Expect(err).ToNot(HaveOccurred())
				})

				It("does not install the plugin and does not create a bad state", func() {
					session := helpers.CFWithStdin(buffer, "install-plugin", pluginPath)

					Eventually(session).Should(Say(`Attention: Plugins are binaries written by potentially untrusted authors\.`))
					Eventually(session).Should(Say(`Install and use plugins at your own risk\.`))
					Eventually(session).Should(Say(`Do you want to install the plugin %s\? \[yN\]:`, helpers.ConvertPathToRegularExpression(pluginPath)))

					session.Interrupt()

					// There is a timing issue -- the exit code may be either 1 (processed error), 2 (config writing error), or 130 (Ctrl-C)
					Eventually(session).Should(SatisfyAny(Exit(1), Exit(2), Exit(130)))

					Expect(session).To(Say("FAILED"))

					Expect(server.ReceivedRequests()).To(HaveLen(0))

					// make sure cf plugins did not break
					Eventually(helpers.CF("plugins", "--checksum")).Should(Exit(0))

					// make sure a retry of the plugin install works
					retrySession := helpers.CF("install-plugin", pluginPath, "-f")
					Eventually(retrySession).Should(Say(`Plugin some-plugin 1\.0\.0 successfully installed\.`))
					Eventually(retrySession).Should(Exit(0))
				})
			})
		})
	})
})
