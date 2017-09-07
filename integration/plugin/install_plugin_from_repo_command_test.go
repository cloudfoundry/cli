package plugin

import (
	"net/http"
	"runtime"

	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/util/generic"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
	. "github.com/onsi/gomega/ghttp"
)

var _ = Describe("install-plugin (from repo) command", func() {
	Describe("installing a plugin from a specific repo", func() {
		Context("when the repo and the plugin name are swapped", func() {
			var repoServer *Server

			BeforeEach(func() {
				repoServer = helpers.NewPluginRepositoryServer(helpers.PluginRepository{})
				Eventually(helpers.CF("add-plugin-repo", "kaka", repoServer.URL(), "-k")).Should(Exit(0))
			})

			AfterEach(func() {
				repoServer.Close()
			})

			It("it parses the arguments correctly", func() {
				session := helpers.CF("install-plugin", "-f", "some-plugin", "-r", "kaka", "-k")

				Eventually(session.Err).Should(Say("Plugin some-plugin not found in repository kaka\\."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the repo is not registered", func() {
			It("fails with an error message", func() {
				session := helpers.CF("install-plugin", "-f", "-r", "repo-that-does-not-exist", "some-plugin")

				Eventually(session.Err).Should(Say("Plugin repository repo-that-does-not-exist not found\\."))
				Eventually(session.Err).Should(Say("Use 'cf list-plugin-repos' to list registered repos\\."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when fetching a list of plugins from a repo returns a 4xx/5xx status or a SSL error", func() {
			var repoServer *Server

			BeforeEach(func() {
				repoServer = NewTLSServer()

				repoServer.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/list"),
						RespondWith(http.StatusOK, `{"plugins":[]}`),
					),
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/list"),
						RespondWith(http.StatusTeapot, nil),
					),
				)

				Eventually(helpers.CF("add-plugin-repo", "kaka", repoServer.URL(), "-k")).Should(Exit(0))
			})

			AfterEach(func() {
				repoServer.Close()
			})

			It("fails with an error message", func() {
				session := helpers.CF("install-plugin", "-f", "-r", "kaka", "some-plugin", "-k")

				Eventually(session.Err).Should(Say("Download attempt failed; server returned 418 I'm a teapot"))
				Eventually(session.Err).Should(Say("Unable to install; plugin is not available from the given URL\\."))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the repo returns invalid json", func() {
			var repoServer *Server

			BeforeEach(func() {
				repoServer = NewTLSServer()

				repoServer.AppendHandlers(
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/list"),
						RespondWith(http.StatusOK, `{"plugins":[]}`),
					),
					CombineHandlers(
						VerifyRequest(http.MethodGet, "/list"),
						RespondWith(http.StatusOK, `{"foo":}`),
					),
				)

				Eventually(helpers.CF("add-plugin-repo", "kaka", repoServer.URL(), "-k")).Should(Exit(0))
			})

			AfterEach(func() {
				repoServer.Close()
			})

			It("fails with an error message", func() {
				session := helpers.CF("install-plugin", "-f", "-r", "kaka", "some-plugin", "-k")

				Eventually(session.Err).Should(Say("Invalid JSON content from server: invalid character '}' looking for beginning of value"))
				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the repo does not contain the specified plugin", func() {
			var repoServer *Server

			BeforeEach(func() {
				repoServer = helpers.NewPluginRepositoryServer(helpers.PluginRepository{})
				Eventually(helpers.CF("add-plugin-repo", "kaka", repoServer.URL(), "-k")).Should(Exit(0))
			})

			AfterEach(func() {
				repoServer.Close()
			})

			It("fails with an error message", func() {
				session := helpers.CF("install-plugin", "-f", "-r", "kaka", "plugin-that-does-not-exist", "-k")

				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("Plugin plugin-that-does-not-exist not found in repository kaka\\."))
				Eventually(session.Err).Should(Say("Use 'cf repo-plugins -r kaka' to list plugins available in the repo\\."))

				Eventually(session).Should(Exit(1))
			})
		})

		Context("when the repo contains the specified plugin", func() {
			var repoServer *helpers.PluginRepositoryServerWithPlugin

			Context("when no compatible binary is found in the repo", func() {
				BeforeEach(func() {
					repoServer = helpers.NewPluginRepositoryServerWithPlugin("some-plugin", "1.0.0", "not-me-platform", true)
					Eventually(helpers.CF("add-plugin-repo", "kaka", repoServer.URL())).Should(Exit(0))
				})

				AfterEach(func() {
					repoServer.Cleanup()
				})

				It("returns plugin not found", func() {
					session := helpers.CF("install-plugin", "-f", "-r", "kaka", "some-plugin", "-k")
					Eventually(session.Out).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("Plugin requested has no binary available for your platform\\."))

					Eventually(session).Should(Exit(1))
				})
			})

			Context("when -f is specified", func() {
				Context("when the plugin is not already installed", func() {
					Context("when the plugin checksum is valid", func() {
						BeforeEach(func() {
							repoServer = helpers.NewPluginRepositoryServerWithPlugin("some-plugin", "1.0.0", generic.GeneratePlatform(runtime.GOOS, runtime.GOARCH), true)
							Eventually(helpers.CF("add-plugin-repo", "kaka", repoServer.URL())).Should(Exit(0))
						})

						AfterEach(func() {
							repoServer.Cleanup()
						})

						It("installs the plugin case-insensitively", func() {
							session := helpers.CF("install-plugin", "-f", "-r", "kAkA", "some-plugin", "-k")
							Eventually(session.Out).Should(Say("Searching kaka for plugin some-plugin\\.\\.\\."))
							Eventually(session.Out).Should(Say("Plugin some-plugin 1\\.0\\.0 found in: kaka\n"))
							Eventually(session.Out).Should(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
							Eventually(session.Out).Should(Say("Install and use plugins at your own risk\\."))
							Eventually(session.Out).Should(Say("Starting download of plugin binary from repository kaka\\.\\.\\."))
							Eventually(session.Out).Should(Say("\\d.* .*B / ?"))
							Eventually(session.Out).Should(Say("Installing plugin some-plugin\\.\\.\\."))
							Eventually(session.Out).Should(Say("OK"))
							Eventually(session.Out).Should(Say("Plugin some-plugin 1\\.0\\.0 successfully installed\\."))

							Eventually(session).Should(Exit(0))
						})
					})

					Context("when the plugin checksum is invalid", func() {
						BeforeEach(func() {
							repoServer = helpers.NewPluginRepositoryServerWithPlugin("some-plugin", "1.0.0", generic.GeneratePlatform(runtime.GOOS, runtime.GOARCH), false)
							Eventually(helpers.CF("add-plugin-repo", "kaka", repoServer.URL())).Should(Exit(0))

						})

						AfterEach(func() {
							repoServer.Cleanup()
						})

						It("fails with an error message", func() {
							session := helpers.CF("install-plugin", "-f", "-r", "kaka", "some-plugin", "-k")
							Eventually(session.Out).Should(Say("FAILED"))
							Eventually(session.Err).Should(Say("Downloaded plugin binary's checksum does not match repo metadata\\."))
							Eventually(session.Err).Should(Say("Please try again or contact the plugin author\\."))
							Eventually(session).Should(Exit(1))
						})
					})
				})

				Context("when the plugin is already installed", func() {
					BeforeEach(func() {
						pluginPath := helpers.BuildConfigurablePlugin("configurable_plugin", "some-plugin", "1.0.0",
							[]helpers.PluginCommand{
								{Name: "some-command", Help: "some-command-help"},
							},
						)
						Eventually(helpers.CF("install-plugin", pluginPath, "-f", "-k")).Should(Exit(0))
					})

					Context("when the plugin checksum is valid", func() {
						BeforeEach(func() {
							repoServer = helpers.NewPluginRepositoryServerWithPlugin("some-plugin", "2.0.0", generic.GeneratePlatform(runtime.GOOS, runtime.GOARCH), true)
							Eventually(helpers.CF("add-plugin-repo", "kaka", repoServer.URL())).Should(Exit(0))
						})

						AfterEach(func() {
							repoServer.Cleanup()
						})

						It("reinstalls the plugin", func() {
							session := helpers.CF("install-plugin", "-f", "-r", "kaka", "some-plugin", "-k")

							Eventually(session.Out).Should(Say("Searching kaka for plugin some-plugin\\.\\.\\."))
							Eventually(session.Out).Should(Say("Plugin some-plugin 2\\.0\\.0 found in: kaka\n"))
							Eventually(session.Out).Should(Say("Plugin some-plugin 1\\.0\\.0 is already installed\\."))
							Eventually(session.Out).Should(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
							Eventually(session.Out).Should(Say("Install and use plugins at your own risk\\."))
							Eventually(session.Out).Should(Say("Starting download of plugin binary from repository kaka\\.\\.\\."))
							Eventually(session.Out).Should(Say("\\d.* .*B / ?"))
							Eventually(session.Out).Should(Say("Uninstalling existing plugin\\.\\.\\."))
							Eventually(session.Out).Should(Say("OK"))
							Eventually(session.Out).Should(Say("Plugin some-plugin successfully uninstalled\\."))
							Eventually(session.Out).Should(Say("Installing plugin some-plugin\\.\\.\\."))
							Eventually(session.Out).Should(Say("OK"))
							Eventually(session.Out).Should(Say("Plugin some-plugin 2\\.0\\.0 successfully installed\\."))

							Eventually(session).Should(Exit(0))
						})
					})

					Context("when the plugin checksum is invalid", func() {
						BeforeEach(func() {
							repoServer = helpers.NewPluginRepositoryServerWithPlugin("some-plugin", "2.0.0", generic.GeneratePlatform(runtime.GOOS, runtime.GOARCH), false)
							Eventually(helpers.CF("add-plugin-repo", "kaka", repoServer.URL())).Should(Exit(0))
						})

						AfterEach(func() {
							repoServer.Cleanup()
						})

						It("fails with an error message", func() {
							session := helpers.CF("install-plugin", "-f", "-r", "kaka", "some-plugin", "-k")

							Eventually(session.Out).Should(Say("Searching kaka for plugin some-plugin\\.\\.\\."))
							Eventually(session.Out).Should(Say("Plugin some-plugin 2\\.0\\.0 found in: kaka\n"))
							Eventually(session.Out).Should(Say("Plugin some-plugin 1\\.0\\.0 is already installed\\."))
							Eventually(session.Out).Should(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
							Eventually(session.Out).Should(Say("Install and use plugins at your own risk\\."))
							Eventually(session.Out).Should(Say("Starting download of plugin binary from repository kaka\\.\\.\\."))
							Eventually(session.Out).Should(Say("\\d.* .*B / ?"))

							Eventually(session.Out).Should(Say("FAILED"))
							Eventually(session.Err).Should(Say("Downloaded plugin binary's checksum does not match repo metadata\\."))
							Eventually(session.Err).Should(Say("Please try again or contact the plugin author\\."))

							Eventually(session).Should(Exit(1))
						})
					})
				})
			})

			Context("when -f is not specified", func() {
				var buffer *Buffer

				BeforeEach(func() {
					buffer = NewBuffer()
				})

				Context("when the plugin is not already installed", func() {
					BeforeEach(func() {
						repoServer = helpers.NewPluginRepositoryServerWithPlugin("some-plugin", "1.2.3", generic.GeneratePlatform(runtime.GOOS, runtime.GOARCH), true)
						Eventually(helpers.CF("add-plugin-repo", "kaka", repoServer.URL())).Should(Exit(0))
					})

					AfterEach(func() {
						repoServer.Cleanup()
					})

					Context("when the user says yes", func() {
						BeforeEach(func() {
							_, _ = buffer.Write([]byte("y\n"))
						})

						It("installs the plugin", func() {
							session := helpers.CFWithStdin(buffer, "install-plugin", "-r", "kaka", "some-plugin", "-k")
							Eventually(session.Out).Should(Say("Searching kaka for plugin some-plugin\\.\\.\\."))
							Eventually(session.Out).Should(Say("Plugin some-plugin 1\\.2\\.3 found in: kaka\n"))
							Eventually(session.Out).Should(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
							Eventually(session.Out).Should(Say("Install and use plugins at your own risk\\."))
							Eventually(session.Out).Should(Say("Do you want to install the plugin some-plugin\\? \\[yN\\]: y"))
							Eventually(session.Out).Should(Say("Starting download of plugin binary from repository kaka\\.\\.\\."))
							Eventually(session.Out).Should(Say("\\d.* .*B / ?"))
							Eventually(session.Out).Should(Say("Installing plugin some-plugin\\.\\.\\."))
							Eventually(session.Out).Should(Say("OK"))
							Eventually(session.Out).Should(Say("Plugin some-plugin 1\\.2\\.3 successfully installed\\."))

							Eventually(session).Should(Exit(0))
						})
					})

					Context("when the user says no", func() {
						BeforeEach(func() {
							_, _ = buffer.Write([]byte("n\n"))
						})

						It("does not install the plugin", func() {
							session := helpers.CFWithStdin(buffer, "install-plugin", "-r", "kaka", "some-plugin", "-k")
							Eventually(session.Out).Should(Say("Searching kaka for plugin some-plugin\\.\\.\\."))
							Eventually(session.Out).Should(Say("Plugin some-plugin 1\\.2\\.3 found in: kaka\n"))
							Eventually(session.Out).Should(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
							Eventually(session.Out).Should(Say("Install and use plugins at your own risk\\."))
							Eventually(session.Out).Should(Say("Do you want to install the plugin some-plugin\\? \\[yN\\]: n"))

							Eventually(session.Out).Should(Say("Plugin installation cancelled"))

							Eventually(session).Should(Exit(0))
						})
					})
				})

				Context("when the plugin is already installed", func() {
					BeforeEach(func() {
						pluginPath := helpers.BuildConfigurablePlugin("configurable_plugin", "some-plugin", "1.2.2",
							[]helpers.PluginCommand{
								{Name: "some-command", Help: "some-command-help"},
							},
						)
						Eventually(helpers.CF("install-plugin", pluginPath, "-f", "-k")).Should(Exit(0))
					})

					Context("when the user chooses yes", func() {
						BeforeEach(func() {
							_, _ = buffer.Write([]byte("y\n"))
						})

						Context("when the plugin checksum is valid", func() {
							BeforeEach(func() {
								repoServer = helpers.NewPluginRepositoryServerWithPlugin("some-plugin", "1.2.3", generic.GeneratePlatform(runtime.GOOS, runtime.GOARCH), true)
								Eventually(helpers.CF("add-plugin-repo", "kaka", repoServer.URL())).Should(Exit(0))
							})

							AfterEach(func() {
								repoServer.Cleanup()
							})

							It("installs the plugin", func() {
								session := helpers.CFWithStdin(buffer, "install-plugin", "-r", "kaka", "some-plugin", "-k")

								Eventually(session.Out).Should(Say("Searching kaka for plugin some-plugin\\.\\.\\."))
								Eventually(session.Out).Should(Say("Plugin some-plugin 1\\.2\\.3 found in: kaka\n"))
								Eventually(session.Out).Should(Say("Plugin some-plugin 1\\.2\\.2 is already installed\\."))
								Eventually(session.Out).Should(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
								Eventually(session.Out).Should(Say("Install and use plugins at your own risk\\."))
								Eventually(session.Out).Should(Say("Do you want to uninstall the existing plugin and install some-plugin 1\\.2\\.3\\? \\[yN\\]: y"))
								Eventually(session.Out).Should(Say("Starting download of plugin binary from repository kaka\\.\\.\\."))
								Eventually(session.Out).Should(Say("\\d.* .*B / ?"))
								Eventually(session.Out).Should(Say("Uninstalling existing plugin\\.\\.\\."))
								Eventually(session.Out).Should(Say("OK"))
								Eventually(session.Out).Should(Say("Plugin some-plugin successfully uninstalled\\."))
								Eventually(session.Out).Should(Say("Installing plugin some-plugin\\.\\.\\."))
								Eventually(session.Out).Should(Say("OK"))
								Eventually(session.Out).Should(Say("Plugin some-plugin 1\\.2\\.3 successfully installed\\."))

								Eventually(session).Should(Exit(0))
							})
						})

						Context("when the plugin checksum is invalid", func() {
							BeforeEach(func() {
								repoServer = helpers.NewPluginRepositoryServerWithPlugin("some-plugin", "1.2.3", generic.GeneratePlatform(runtime.GOOS, runtime.GOARCH), false)

								Eventually(helpers.CF("add-plugin-repo", "kaka", repoServer.URL())).Should(Exit(0))
							})

							AfterEach(func() {
								repoServer.Cleanup()
							})

							It("fails with an error message", func() {
								session := helpers.CFWithStdin(buffer, "install-plugin", "-r", "kaka", "some-plugin", "-k")
								Eventually(session.Out).Should(Say("Searching kaka for plugin some-plugin\\.\\.\\."))
								Eventually(session.Out).Should(Say("Plugin some-plugin 1\\.2\\.3 found in: kaka\n"))
								Eventually(session.Out).Should(Say("Plugin some-plugin 1\\.2\\.2 is already installed\\."))
								Eventually(session.Out).Should(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
								Eventually(session.Out).Should(Say("Install and use plugins at your own risk\\."))
								Eventually(session.Out).Should(Say("Do you want to uninstall the existing plugin and install some-plugin 1\\.2\\.3\\? \\[yN\\]: y"))
								Eventually(session.Out).Should(Say("Starting download of plugin binary from repository kaka\\.\\.\\."))
								Eventually(session.Out).Should(Say("\\d.* .*B / ?"))
								Eventually(session.Out).Should(Say("FAILED"))
								Eventually(session.Err).Should(Say("Downloaded plugin binary's checksum does not match repo metadata\\."))
								Eventually(session.Err).Should(Say("Please try again or contact the plugin author\\."))

								Eventually(session).Should(Exit(1))
							})
						})
					})

					Context("when the user chooses no", func() {
						BeforeEach(func() {
							repoServer = helpers.NewPluginRepositoryServerWithPlugin("some-plugin", "1.2.3", generic.GeneratePlatform(runtime.GOOS, runtime.GOARCH), false)
							Eventually(helpers.CF("add-plugin-repo", "kaka", repoServer.URL())).Should(Exit(0))

							_, _ = buffer.Write([]byte("n\n"))
						})

						AfterEach(func() {
							repoServer.Cleanup()
						})

						It("does not install the plugin", func() {
							session := helpers.CFWithStdin(buffer, "install-plugin", "-r", "kaka", "some-plugin", "-k")

							Eventually(session.Out).Should(Say("Searching kaka for plugin some-plugin\\.\\.\\."))
							Eventually(session.Out).Should(Say("Plugin some-plugin 1\\.2\\.3 found in: kaka\n"))
							Eventually(session.Out).Should(Say("Plugin some-plugin 1\\.2\\.2 is already installed\\."))
							Eventually(session.Out).Should(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
							Eventually(session.Out).Should(Say("Install and use plugins at your own risk\\."))
							Eventually(session.Out).Should(Say("Do you want to uninstall the existing plugin and install some-plugin 1\\.2\\.3\\? \\[yN\\]: n"))
							Eventually(session.Out).Should(Say("Plugin installation cancelled"))

							Eventually(session).Should(Exit(0))
						})
					})
				})
			})
		})
	})

	Describe("installing a plugin from any repo", func() {
		Context("when there are no repositories registered", func() {
			It("fails and displays the plugin not found message", func() {
				session := helpers.CF("install-plugin", "some-plugin")

				Eventually(session.Out).Should(Say("FAILED"))
				Eventually(session.Err).Should(Say("Plugin some-plugin not found on disk or in any registered repo\\."))
				Eventually(session.Err).Should(Say("Use 'cf repo-plugins' to list plugins available in the repos\\."))

				Eventually(session).Should(Exit(1))
			})
		})

		Context("when there are repositories registered", func() {
			Context("when fetching a list of plugins from a repo returns a 4xx/5xx status or a SSL error", func() {
				var repoServer *Server

				BeforeEach(func() {
					repoServer = NewTLSServer()

					repoServer.AppendHandlers(
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/list"),
							RespondWith(http.StatusOK, `{"plugins":[]}`),
						),
						CombineHandlers(
							VerifyRequest(http.MethodGet, "/list"),
							RespondWith(http.StatusTeapot, nil),
						),
					)

					Eventually(helpers.CF("add-plugin-repo", "kaka", repoServer.URL(), "-k")).Should(Exit(0))
				})

				AfterEach(func() {
					repoServer.Close()
				})

				It("fails with an error message", func() {
					session := helpers.CF("install-plugin", "-f", "some-plugin", "-k")

					Eventually(session.Err).Should(Say("Plugin list download failed; repository kaka returned 418 I'm a teapot"))
					Consistently(session.Err).ShouldNot(Say("Unable to install; plugin is not available from the given URL\\."))
					Eventually(session).Should(Exit(1))
				})
			})

			Context("when the plugin isn't found in any of the repositories", func() {
				var repoServer1 *Server
				var repoServer2 *Server

				BeforeEach(func() {
					repoServer1 = helpers.NewPluginRepositoryServer(helpers.PluginRepository{})
					repoServer2 = helpers.NewPluginRepositoryServer(helpers.PluginRepository{})
					Eventually(helpers.CF("add-plugin-repo", "kaka1", repoServer1.URL(), "-k")).Should(Exit(0))
					Eventually(helpers.CF("add-plugin-repo", "kaka2", repoServer2.URL(), "-k")).Should(Exit(0))
				})

				AfterEach(func() {
					repoServer1.Close()
					repoServer2.Close()
				})

				It("fails and displays the plugin not found message", func() {
					session := helpers.CF("install-plugin", "some-plugin", "-k")

					Eventually(session.Out).Should(Say("Searching kaka1, kaka2 for plugin some-plugin\\.\\.\\."))
					Eventually(session.Out).Should(Say("FAILED"))
					Eventually(session.Err).Should(Say("Plugin some-plugin not found on disk or in any registered repo\\."))
					Eventually(session.Err).Should(Say("Use 'cf repo-plugins' to list plugins available in the repos\\."))

					Eventually(session).Should(Exit(1))
				})
			})

			Context("when -f is specified", func() {
				Context("when identical versions of the plugin are found in the repositories", func() {
					var repoServer1 *helpers.PluginRepositoryServerWithPlugin
					var repoServer2 *helpers.PluginRepositoryServerWithPlugin

					BeforeEach(func() {
						repoServer1 = helpers.NewPluginRepositoryServerWithPlugin("some-plugin", "1.2.3", generic.GeneratePlatform(runtime.GOOS, runtime.GOARCH), true)
						Eventually(helpers.CF("add-plugin-repo", "kaka1", repoServer1.URL())).Should(Exit(0))

						repoServer2 = helpers.NewPluginRepositoryServerWithPlugin("some-plugin", "1.2.3", generic.GeneratePlatform(runtime.GOOS, runtime.GOARCH), true)
						Eventually(helpers.CF("add-plugin-repo", "kaka2", repoServer2.URL())).Should(Exit(0))
					})

					AfterEach(func() {
						repoServer1.Cleanup()
						repoServer2.Cleanup()
					})

					Context("when the plugin is not already installed", func() {
						Context("when the checksum is valid", func() {
							It("installs the plugin", func() {
								session := helpers.CF("install-plugin", "some-plugin", "-f", "-k")

								Eventually(session.Out).Should(Say("Searching kaka1, kaka2 for plugin some-plugin\\.\\.\\."))
								Eventually(session.Out).Should(Say("Plugin some-plugin 1\\.2\\.3 found in: kaka1, kaka2"))
								Eventually(session.Out).Should(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
								Eventually(session.Out).Should(Say("Install and use plugins at your own risk\\."))
								Eventually(session.Out).Should(Say("Starting download of plugin binary from repository kaka1\\.\\.\\."))
								Eventually(session.Out).Should(Say("\\d.* .*B / ?"))
								Eventually(session.Out).Should(Say("Installing plugin some-plugin\\.\\.\\."))
								Eventually(session.Out).Should(Say("OK"))
								Eventually(session.Out).Should(Say("Plugin some-plugin 1\\.2\\.3 successfully installed\\."))

								Eventually(session).Should(Exit(0))
							})
						})

						Context("when the plugin checksum is invalid", func() {
							var repoServer3 *helpers.PluginRepositoryServerWithPlugin
							var repoServer4 *helpers.PluginRepositoryServerWithPlugin

							BeforeEach(func() {
								repoServer3 = helpers.NewPluginRepositoryServerWithPlugin("some-plugin-with-bad-checksum", "2.2.3", generic.GeneratePlatform(runtime.GOOS, runtime.GOARCH), false)
								Eventually(helpers.CF("add-plugin-repo", "kaka3", repoServer3.URL())).Should(Exit(0))

								repoServer4 = helpers.NewPluginRepositoryServerWithPlugin("some-plugin-with-bad-checksum", "2.2.3", generic.GeneratePlatform(runtime.GOOS, runtime.GOARCH), true)
								Eventually(helpers.CF("add-plugin-repo", "kaka4", repoServer4.URL())).Should(Exit(0))
							})

							AfterEach(func() {
								repoServer3.Cleanup()
								repoServer4.Cleanup()
							})

							It("fails with the invalid checksum message", func() {
								session := helpers.CF("install-plugin", "some-plugin-with-bad-checksum", "-f", "-k")

								Eventually(session.Out).Should(Say("Searching kaka1, kaka2, kaka3, kaka4 for plugin some-plugin-with-bad-checksum\\.\\.\\."))
								Eventually(session.Out).Should(Say("Plugin some-plugin-with-bad-checksum 2\\.2\\.3 found in: kaka3, kaka4"))
								Eventually(session.Out).Should(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
								Eventually(session.Out).Should(Say("Install and use plugins at your own risk\\."))
								Eventually(session.Out).Should(Say("Starting download of plugin binary from repository kaka3\\.\\.\\."))
								Eventually(session.Out).Should(Say("\\d.* .*B / ?"))
								Eventually(session.Out).Should(Say("FAILED"))
								Eventually(session.Err).Should(Say("Downloaded plugin binary's checksum does not match repo metadata\\."))
								Eventually(session.Err).Should(Say("Please try again or contact the plugin author\\."))

								Eventually(session).Should(Exit(1))
							})
						})
					})

					Context("when the plugin is already installed", func() {
						Context("when the checksum is valid", func() {
							BeforeEach(func() {
								pluginPath := helpers.BuildConfigurablePlugin("configurable_plugin", "some-plugin", "1.0.0",
									[]helpers.PluginCommand{
										{Name: "some-command", Help: "some-command-help"},
									},
								)
								Eventually(helpers.CF("install-plugin", pluginPath, "-f", "-k")).Should(Exit(0))
							})

							It("reinstalls the plugin", func() {
								session := helpers.CF("install-plugin", "-f", "some-plugin", "-k")

								Eventually(session.Out).Should(Say("Searching kaka1, kaka2 for plugin some-plugin\\.\\.\\."))
								Eventually(session.Out).Should(Say("Plugin some-plugin 1\\.2\\.3 found in: kaka1, kaka2"))
								Eventually(session.Out).Should(Say("Plugin some-plugin 1\\.0\\.0 is already installed\\."))
								Eventually(session.Out).Should(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))

								Eventually(session).Should(Exit(0))
							})
						})
					})
				})

				Context("when the binary version for the current GOOS/GOARCH is exists in only one repo", func() {
					var repoServer1 *helpers.PluginRepositoryServerWithPlugin
					var repoServer2 *helpers.PluginRepositoryServerWithPlugin

					BeforeEach(func() {
						repoServer1 = helpers.NewPluginRepositoryServerWithPlugin("some-plugin", "1.2.3", "solaris", false)
						Eventually(helpers.CF("add-plugin-repo", "kaka1", repoServer1.URL())).Should(Exit(0))

						repoServer2 = helpers.NewPluginRepositoryServerWithPlugin("some-plugin", "1.2.3", generic.GeneratePlatform(runtime.GOOS, runtime.GOARCH), true)
						Eventually(helpers.CF("add-plugin-repo", "kaka2", repoServer2.URL())).Should(Exit(0))
					})

					AfterEach(func() {
						repoServer1.Cleanup()
						repoServer2.Cleanup()
					})

					It("installs the plugin from the correct repo", func() {
						session := helpers.CF("install-plugin", "-f", "some-plugin", "-k")

						Eventually(session.Out).Should(Say("Searching kaka1, kaka2 for plugin some-plugin\\.\\.\\."))
						Eventually(session.Out).Should(Say("Plugin some-plugin 1\\.2\\.3 found in: kaka2"))
						Eventually(session.Out).Should(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
						Eventually(session.Out).Should(Say("Install and use plugins at your own risk\\."))
						Eventually(session.Out).Should(Say("Starting download of plugin binary from repository kaka2\\.\\.\\."))

						Eventually(session).Should(Exit(0))
					})
				})

				Context("when different versions of the plugin are found in the repositories", func() {
					Context("when the checksum is valid", func() {
						var repoServer1 *helpers.PluginRepositoryServerWithPlugin
						var repoServer2 *helpers.PluginRepositoryServerWithPlugin

						BeforeEach(func() {
							repoServer1 = helpers.NewPluginRepositoryServerWithPlugin("some-plugin", "1.2.3", generic.GeneratePlatform(runtime.GOOS, runtime.GOARCH), true)
							Eventually(helpers.CF("add-plugin-repo", "kaka1", repoServer1.URL())).Should(Exit(0))

							repoServer2 = helpers.NewPluginRepositoryServerWithPlugin("some-plugin", "1.2.4", generic.GeneratePlatform(runtime.GOOS, runtime.GOARCH), true)
							Eventually(helpers.CF("add-plugin-repo", "kaka2", repoServer2.URL())).Should(Exit(0))
						})

						AfterEach(func() {
							repoServer1.Cleanup()
							repoServer2.Cleanup()
						})

						It("installs the newest plugin", func() {
							session := helpers.CF("install-plugin", "some-plugin", "-f", "-k")

							Eventually(session.Out).Should(Say("Searching kaka1, kaka2 for plugin some-plugin\\.\\.\\."))
							Eventually(session.Out).Should(Say("Plugin some-plugin 1.2.4 found in: kaka2"))
							Eventually(session.Out).Should(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
							Eventually(session.Out).Should(Say("Install and use plugins at your own risk\\."))
							Eventually(session.Out).Should(Say("Starting download of plugin binary from repository kaka2\\.\\.\\."))
							Eventually(session.Out).Should(Say("\\d.* .*B / ?"))
							Eventually(session.Out).Should(Say("Installing plugin some-plugin\\.\\.\\."))
							Eventually(session.Out).Should(Say("OK"))
							Eventually(session.Out).Should(Say("Plugin some-plugin 1\\.2\\.4 successfully installed\\."))

							Eventually(session).Should(Exit(0))
						})
					})

					Context("when the checksum for the latest version is invalid", func() {
						var repoServer1 *helpers.PluginRepositoryServerWithPlugin
						var repoServer2 *helpers.PluginRepositoryServerWithPlugin

						BeforeEach(func() {
							repoServer1 = helpers.NewPluginRepositoryServerWithPlugin("some-plugin", "1.2.3", generic.GeneratePlatform(runtime.GOOS, runtime.GOARCH), true)
							Eventually(helpers.CF("add-plugin-repo", "kaka1", repoServer1.URL())).Should(Exit(0))

							repoServer2 = helpers.NewPluginRepositoryServerWithPlugin("some-plugin", "1.2.4", generic.GeneratePlatform(runtime.GOOS, runtime.GOARCH), false)
							Eventually(helpers.CF("add-plugin-repo", "kaka2", repoServer2.URL())).Should(Exit(0))
						})

						AfterEach(func() {
							repoServer1.Cleanup()
							repoServer2.Cleanup()
						})

						It("prints the invalid checksum error", func() {
							session := helpers.CF("install-plugin", "some-plugin", "-f", "-k")

							Eventually(session.Out).Should(Say("Searching kaka1, kaka2 for plugin some-plugin\\.\\.\\."))
							Eventually(session.Out).Should(Say("Plugin some-plugin 1.2.4 found in: kaka2"))
							Eventually(session.Out).Should(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
							Eventually(session.Out).Should(Say("Install and use plugins at your own risk\\."))
							Eventually(session.Out).Should(Say("Starting download of plugin binary from repository kaka2\\.\\.\\."))
							Eventually(session.Out).Should(Say("\\d.* .*B / ?"))
							Eventually(session.Out).Should(Say("FAILED"))
							Eventually(session.Err).Should(Say("Downloaded plugin binary's checksum does not match repo metadata\\."))
							Eventually(session.Err).Should(Say("Please try again or contact the plugin author\\."))

							Eventually(session).Should(Exit(1))
						})
					})
				})
			})

			Context("when -f is not specified", func() {
				var buffer *Buffer

				BeforeEach(func() {
					buffer = NewBuffer()
				})

				Context("when identical versions of the plugin are found in the repositories", func() {
					var repoServer1 *helpers.PluginRepositoryServerWithPlugin
					var repoServer2 *helpers.PluginRepositoryServerWithPlugin

					BeforeEach(func() {
						repoServer1 = helpers.NewPluginRepositoryServerWithPlugin("some-plugin", "1.2.3", generic.GeneratePlatform(runtime.GOOS, runtime.GOARCH), true)
						Eventually(helpers.CF("add-plugin-repo", "kaka1", repoServer1.URL())).Should(Exit(0))

						repoServer2 = helpers.NewPluginRepositoryServerWithPlugin("some-plugin", "1.2.3", generic.GeneratePlatform(runtime.GOOS, runtime.GOARCH), true)
						Eventually(helpers.CF("add-plugin-repo", "kaka2", repoServer2.URL())).Should(Exit(0))
					})

					AfterEach(func() {
						repoServer1.Cleanup()
						repoServer2.Cleanup()
					})

					Context("when the plugin is not already installed", func() {
						Context("when the user says yes", func() {
							BeforeEach(func() {
								_, _ = buffer.Write([]byte("y\n"))
							})

							Context("when the checksum is valid", func() {
								It("installs the plugin", func() {
									session := helpers.CFWithStdin(buffer, "install-plugin", "some-plugin", "-k")

									Eventually(session.Out).Should(Say("Searching kaka1, kaka2 for plugin some-plugin\\.\\.\\."))
									Eventually(session.Out).Should(Say("Plugin some-plugin 1\\.2\\.3 found in: kaka1, kaka2"))
									Eventually(session.Out).Should(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
									Eventually(session.Out).Should(Say("Install and use plugins at your own risk\\."))
									Eventually(session.Out).Should(Say("Do you want to install the plugin some-plugin\\? \\[yN\\]: y"))
									Eventually(session.Out).Should(Say("Starting download of plugin binary from repository kaka1\\.\\.\\."))
									Eventually(session.Out).Should(Say("\\d.* .*B / ?"))
									Eventually(session.Out).Should(Say("Installing plugin some-plugin\\.\\.\\."))
									Eventually(session.Out).Should(Say("OK"))
									Eventually(session.Out).Should(Say("Plugin some-plugin 1\\.2\\.3 successfully installed\\."))

									Eventually(session).Should(Exit(0))
								})
							})

							Context("when the checksum is invalid", func() {
								var repoServer3 *helpers.PluginRepositoryServerWithPlugin
								var repoServer4 *helpers.PluginRepositoryServerWithPlugin

								BeforeEach(func() {
									repoServer3 = helpers.NewPluginRepositoryServerWithPlugin("some-plugin-with-bad-checksum", "2.2.3", generic.GeneratePlatform(runtime.GOOS, runtime.GOARCH), false)
									Eventually(helpers.CF("add-plugin-repo", "kaka3", repoServer3.URL())).Should(Exit(0))

									repoServer4 = helpers.NewPluginRepositoryServerWithPlugin("some-plugin-with-bad-checksum", "2.2.3", generic.GeneratePlatform(runtime.GOOS, runtime.GOARCH), false)
									Eventually(helpers.CF("add-plugin-repo", "kaka4", repoServer4.URL())).Should(Exit(0))
								})

								AfterEach(func() {
									repoServer3.Cleanup()
									repoServer4.Cleanup()
								})

								It("fails with the invalid checksum message", func() {
									session := helpers.CFWithStdin(buffer, "install-plugin", "some-plugin-with-bad-checksum", "-k")

									Eventually(session.Out).Should(Say("FAILED"))
									Eventually(session.Err).Should(Say("Downloaded plugin binary's checksum does not match repo metadata\\."))
									Eventually(session.Err).Should(Say("Please try again or contact the plugin author\\."))

									Eventually(session).Should(Exit(1))
								})
							})
						})

						Context("when the user says no", func() {
							BeforeEach(func() {
								_, _ = buffer.Write([]byte("n\n"))
							})

							It("does not install the plugin", func() {
								session := helpers.CFWithStdin(buffer, "install-plugin", "some-plugin", "-k")

								Eventually(session.Out).Should(Say("Do you want to install the plugin some-plugin\\? \\[yN\\]: n"))
								Eventually(session.Out).Should(Say("Plugin installation cancelled"))

								Eventually(session).Should(Exit(0))
							})
						})
					})

					Context("when the plugin is already installed", func() {
						BeforeEach(func() {
							pluginPath := helpers.BuildConfigurablePlugin("configurable_plugin", "some-plugin", "1.0.0",
								[]helpers.PluginCommand{
									{Name: "some-command", Help: "some-command-help"},
								},
							)
							Eventually(helpers.CF("install-plugin", pluginPath, "-f", "-k")).Should(Exit(0))
						})

						Context("when the user says yes", func() {
							BeforeEach(func() {
								_, _ = buffer.Write([]byte("y\n"))
							})

							Context("when the checksum is valid", func() {
								It("installs the plugin", func() {
									session := helpers.CFWithStdin(buffer, "install-plugin", "some-plugin", "-k")

									Eventually(session.Out).Should(Say("Searching kaka1, kaka2 for plugin some-plugin\\.\\.\\."))
									Eventually(session.Out).Should(Say("Plugin some-plugin 1\\.2\\.3 found in: kaka1, kaka2"))
									Eventually(session.Out).Should(Say("Plugin some-plugin 1\\.0\\.0 is already installed."))
									Eventually(session.Out).Should(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
									Eventually(session.Out).Should(Say("Install and use plugins at your own risk\\."))
									Eventually(session.Out).Should(Say("Do you want to uninstall the existing plugin and install some-plugin 1\\.2\\.3\\? \\[yN\\]: y"))
									Eventually(session.Out).Should(Say("Starting download of plugin binary from repository kaka1\\.\\.\\."))
									Eventually(session.Out).Should(Say("\\d.* .*B / ?"))
									Eventually(session.Out).Should(Say("Uninstalling existing plugin\\.\\.\\."))
									Eventually(session.Out).Should(Say("OK"))
									Eventually(session.Out).Should(Say("Plugin some-plugin successfully uninstalled\\."))
									Eventually(session.Out).Should(Say("Installing plugin some-plugin\\.\\.\\."))
									Eventually(session.Out).Should(Say("OK"))
									Eventually(session.Out).Should(Say("Plugin some-plugin 1\\.2\\.3 successfully installed\\."))

									Eventually(session).Should(Exit(0))
								})
							})
						})

						Context("when the user says no", func() {
							BeforeEach(func() {
								_, _ = buffer.Write([]byte("n\n"))
							})

							It("does not install the plugin", func() {
								session := helpers.CFWithStdin(buffer, "install-plugin", "some-plugin", "-k")

								Eventually(session.Out).Should(Say("Searching kaka1, kaka2 for plugin some-plugin\\.\\.\\."))
								Eventually(session.Out).Should(Say("Plugin some-plugin 1\\.2\\.3 found in: kaka1, kaka2"))
								Eventually(session.Out).Should(Say("Plugin some-plugin 1\\.0\\.0 is already installed."))
								Eventually(session.Out).Should(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
								Eventually(session.Out).Should(Say("Install and use plugins at your own risk\\."))
								Eventually(session.Out).Should(Say("Do you want to uninstall the existing plugin and install some-plugin 1\\.2\\.3\\? \\[yN\\]: n"))
								Eventually(session.Out).Should(Say("Plugin installation cancelled"))

								Eventually(session).Should(Exit(0))
							})
						})
					})
				})
			})
		})
	})
})
