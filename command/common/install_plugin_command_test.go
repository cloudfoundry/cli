package common_test

import (
	"errors"
	"io/ioutil"

	"code.cloudfoundry.org/cli/actor/pluginaction"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/common"
	"code.cloudfoundry.org/cli/command/common/commonfakes"
	"code.cloudfoundry.org/cli/command/plugin/shared"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = FDescribe("install-plugin command", func() {
	var (
		cmd         InstallPluginCommand
		testUI      *ui.UI
		input       *Buffer
		fakeConfig  *commandfakes.FakeConfig
		fakeActor   *commonfakes.FakeInstallPluginActor
		executeErr  error
		expectedErr error
	)

	BeforeEach(func() {
		input = NewBuffer()
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeActor = new(commonfakes.FakeInstallPluginActor)

		cmd = InstallPluginCommand{
			UI:     testUI,
			Config: fakeConfig,
			Actor:  fakeActor,
		}

		fakeActor.FileExistsReturns(true)
		fakeActor.CreateExecutableCopyReturns("copy-path", nil)
		fakeConfig.ExperimentalReturns(true)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	Describe("when the path/url argument is empty", func() {
		BeforeEach(func() {
			cmd.OptionalArgs.PathURLOrPluginName = ""
		})

		It("returns a RequiredArgumentError ", func() {
			Expect(executeErr).To(MatchError(command.RequiredArgumentError{
				ArgumentName: "PATH_URL_PLUGIN_NAME",
			}))
		})
	})

	Describe("installing from a local file", func() {
		BeforeEach(func() {
			cmd.OptionalArgs.PathURLOrPluginName = "some-path"
		})

		Context("when the local file does not exist", func() {
			BeforeEach(func() {
				fakeActor.FileExistsReturns(false)
			})

			It("does not print installation messages and returns a FileNotFoundError", func() {
				Expect(executeErr).To(MatchError(shared.FileNotFoundError{Path: "some-path"}))

				Expect(testUI.Out).ToNot(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
				Expect(testUI.Out).ToNot(Say("Installing plugin some-path\\.\\.\\."))
			})
		})

		Context("when the -f argument is given", func() {
			BeforeEach(func() {
				cmd.Force = true
			})

			Context("when the plugin is invalid", func() {
				var returnedErr error

				BeforeEach(func() {
					returnedErr = pluginaction.PluginInvalidError{Path: "copy-path"}
					fakeActor.GetAndValidatePluginReturns(configv3.Plugin{}, returnedErr)
				})

				It("returns an error", func() {
					Expect(executeErr).To(MatchError(shared.PluginInvalidError{
						Path:              "some-path",
						WrappedErrMessage: returnedErr.Error()}))

					Expect(testUI.Out).ToNot(Say("Installing plugin"))
				})
			})

			Context("when the plugin is already installed", func() {
				var plugin configv3.Plugin

				BeforeEach(func() {
					plugin = configv3.Plugin{
						Name: "some-plugin",
						Version: configv3.PluginVersion{
							Major: 1,
							Minor: 2,
							Build: 3,
						},
					}
					fakeActor.GetAndValidatePluginReturns(plugin, nil)
					fakeActor.IsPluginInstalledReturns(true)
				})

				Context("when an error is encountered uninstalling the existing plugin", func() {
					BeforeEach(func() {
						expectedErr = errors.New("uninstall plugin error")
						fakeActor.UninstallPluginReturns(expectedErr)
					})

					It("returns the error", func() {
						Expect(executeErr).To(MatchError(expectedErr))

						Expect(testUI.Out).ToNot(Say("Plugin some-plugin successfully uninstalled\\."))
					})
				})

				Context("when no errors are encountered uninstalling the existing plugin", func() {
					It("uninstalls the existing plugin and installs the current plugin", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
						Expect(testUI.Out).To(Say("Install and use plugins at your own risk\\."))
						Expect(testUI.Out).To(Say("Plugin some-plugin 1\\.2\\.3 is already installed\\. Uninstalling existing plugin\\.\\.\\."))
						Expect(testUI.Out).To(Say("OK"))
						Expect(testUI.Out).To(Say("Plugin some-plugin successfully uninstalled\\."))
						Expect(testUI.Out).To(Say("Installing plugin some-plugin\\.\\.\\."))
						Expect(testUI.Out).To(Say("OK"))
						Expect(testUI.Out).To(Say("Plugin some-plugin 1\\.2\\.3 successfully installed\\."))

						Expect(fakeActor.FileExistsCallCount()).To(Equal(1))
						Expect(fakeActor.FileExistsArgsForCall(0)).To(Equal("some-path"))

						Expect(fakeActor.GetAndValidatePluginCallCount()).To(Equal(1))
						_, _, path := fakeActor.GetAndValidatePluginArgsForCall(0)
						Expect(path).To(Equal("copy-path"))

						Expect(fakeActor.IsPluginInstalledCallCount()).To(Equal(1))
						Expect(fakeActor.IsPluginInstalledArgsForCall(0)).To(Equal("some-plugin"))

						Expect(fakeActor.UninstallPluginCallCount()).To(Equal(1))
						_, pluginName := fakeActor.UninstallPluginArgsForCall(0)
						Expect(pluginName).To(Equal("some-plugin"))

						Expect(fakeActor.InstallPluginFromPathCallCount()).To(Equal(1))
						path, installedPlugin := fakeActor.InstallPluginFromPathArgsForCall(0)
						Expect(path).To(Equal("copy-path"))
						Expect(installedPlugin).To(Equal(plugin))
					})

					Context("when an error is encountered installing the plugin", func() {
						BeforeEach(func() {
							expectedErr = errors.New("install plugin error")
							fakeActor.InstallPluginFromPathReturns(expectedErr)
						})

						It("returns the error", func() {
							Expect(executeErr).To(MatchError(expectedErr))

							Expect(testUI.Out).ToNot(Say("Plugin some-plugin 1\\.2\\.3 successfully installed\\."))
						})
					})
				})
			})

			Context("when the plugin is not already installed", func() {
				var plugin configv3.Plugin

				BeforeEach(func() {
					plugin = configv3.Plugin{
						Name: "some-plugin",
						Version: configv3.PluginVersion{
							Major: 1,
							Minor: 2,
							Build: 3,
						},
					}
					fakeActor.GetAndValidatePluginReturns(plugin, nil)
				})

				It("installs the plugin", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).To(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
					Expect(testUI.Out).To(Say("Install and use plugins at your own risk\\."))
					Expect(testUI.Out).To(Say("Installing plugin some-plugin\\.\\.\\."))
					Expect(testUI.Out).To(Say("OK"))
					Expect(testUI.Out).To(Say("Plugin some-plugin 1\\.2\\.3 successfully installed\\."))

					Expect(fakeActor.FileExistsCallCount()).To(Equal(1))
					Expect(fakeActor.FileExistsArgsForCall(0)).To(Equal("some-path"))

					Expect(fakeActor.CreateExecutableCopyCallCount()).To(Equal(1))
					Expect(fakeActor.CreateExecutableCopyArgsForCall(0)).To(Equal("some-path"))

					Expect(fakeActor.GetAndValidatePluginCallCount()).To(Equal(1))
					_, _, path := fakeActor.GetAndValidatePluginArgsForCall(0)
					Expect(path).To(Equal("copy-path"))

					Expect(fakeActor.IsPluginInstalledCallCount()).To(Equal(1))
					Expect(fakeActor.IsPluginInstalledArgsForCall(0)).To(Equal("some-plugin"))

					Expect(fakeActor.InstallPluginFromPathCallCount()).To(Equal(1))
					path, installedPlugin := fakeActor.InstallPluginFromPathArgsForCall(0)
					Expect(path).To(Equal("copy-path"))
					Expect(installedPlugin).To(Equal(plugin))

					Expect(fakeActor.UninstallPluginCallCount()).To(Equal(0))
				})

				Context("when there is an error making an executable copy of the plugin binary", func() {
					BeforeEach(func() {
						expectedErr = errors.New("create executable copy error")
						fakeActor.CreateExecutableCopyReturns("", expectedErr)
					})

					It("returns the error", func() {
						Expect(executeErr).To(MatchError(expectedErr))
					})
				})

				Context("when an error is encountered installing the plugin", func() {
					BeforeEach(func() {
						expectedErr = errors.New("install plugin error")
						fakeActor.InstallPluginFromPathReturns(expectedErr)
					})

					It("returns the error", func() {
						Expect(executeErr).To(MatchError(expectedErr))

						Expect(testUI.Out).ToNot(Say("Plugin some-plugin 1\\.2\\.3 successfully installed\\."))
					})
				})
			})
		})

		Context("when the -f argument is not given (user is prompted for confirmation)", func() {
			BeforeEach(func() {
				cmd.Force = false
			})

			Context("when the user chooses no", func() {
				BeforeEach(func() {
					input.Write([]byte("n\n"))
				})

				It("cancels plugin installation", func() {
					Expect(executeErr).To(MatchError(shared.PluginInstallationCancelled{}))

					Expect(testUI.Out).ToNot(Say("Installing plugin"))
				})
			})

			Context("when the user chooses the default", func() {
				BeforeEach(func() {
					input.Write([]byte("\n"))
				})

				It("cancels plugin installation", func() {
					Expect(executeErr).To(MatchError(shared.PluginInstallationCancelled{}))

					Expect(testUI.Out).ToNot(Say("Installing plugin"))
				})
			})

			Context("when the user input is invalid", func() {
				BeforeEach(func() {
					input.Write([]byte("e\n"))
				})

				It("returns an error", func() {
					Expect(executeErr).To(HaveOccurred())

					Expect(testUI.Out).ToNot(Say("Installing plugin"))
				})
			})

			Context("when the user chooses yes", func() {
				BeforeEach(func() {
					input.Write([]byte("y\n"))
				})

				Context("when the plugin is not already installed", func() {
					var plugin configv3.Plugin

					BeforeEach(func() {
						plugin = configv3.Plugin{
							Name: "some-plugin",
							Version: configv3.PluginVersion{
								Major: 1,
								Minor: 2,
								Build: 3,
							},
						}
						fakeActor.GetAndValidatePluginReturns(plugin, nil)
					})

					It("installs the plugin", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
						Expect(testUI.Out).To(Say("Install and use plugins at your own risk\\."))
						Expect(testUI.Out).To(Say("Do you want to install the plugin some-path\\? \\[yN\\]"))
						Expect(testUI.Out).To(Say("Installing plugin some-plugin\\.\\.\\."))
						Expect(testUI.Out).To(Say("OK"))
						Expect(testUI.Out).To(Say("Plugin some-plugin 1\\.2\\.3 successfully installed\\."))

						Expect(fakeActor.FileExistsCallCount()).To(Equal(1))
						Expect(fakeActor.FileExistsArgsForCall(0)).To(Equal("some-path"))

						Expect(fakeActor.GetAndValidatePluginCallCount()).To(Equal(1))
						_, _, path := fakeActor.GetAndValidatePluginArgsForCall(0)
						Expect(path).To(Equal("copy-path"))

						Expect(fakeActor.IsPluginInstalledCallCount()).To(Equal(1))
						Expect(fakeActor.IsPluginInstalledArgsForCall(0)).To(Equal("some-plugin"))

						Expect(fakeActor.InstallPluginFromPathCallCount()).To(Equal(1))
						path, plugin := fakeActor.InstallPluginFromPathArgsForCall(0)
						Expect(path).To(Equal("copy-path"))
						Expect(plugin).To(Equal(plugin))

						Expect(fakeActor.UninstallPluginCallCount()).To(Equal(0))
					})
				})

				Context("when the plugin is already installed", func() {
					BeforeEach(func() {
						plugin := configv3.Plugin{
							Name: "some-plugin",
							Version: configv3.PluginVersion{
								Major: 1,
								Minor: 2,
								Build: 3,
							},
						}
						fakeActor.GetAndValidatePluginReturns(plugin, nil)
						fakeActor.IsPluginInstalledReturns(true)
					})

					It("returns PluginAlreadyInstalledError", func() {
						Expect(executeErr).To(MatchError(shared.PluginAlreadyInstalledError{
							Name:    "some-plugin",
							Version: "1.2.3",
							Path:    "some-path",
						}))
					})
				})
			})
		})
	})

	Describe("installing from a bad URL", func() {
		BeforeEach(func() {
			cmd.OptionalArgs.PathURLOrPluginName = "ftp://some-url"
		})

		It("returns an error indicating an unsupported URL scheme", func() {
			Expect(executeErr).To(MatchError(command.UnsupportedURLSchemeError{
				UnsupportedURL: string(cmd.OptionalArgs.PathURLOrPluginName),
			}))
		})
	})

	Describe("installing from a URL", func() {
		var (
			plugin               configv3.Plugin
			pluginName           string
			downloadedPluginPath string
		)

		BeforeEach(func() {
			cmd.OptionalArgs.PathURLOrPluginName = "http://some-url"
			pluginName = "some-plugin"
			plugin = configv3.Plugin{
				Name: pluginName,
				Version: configv3.PluginVersion{
					Major: 1,
					Minor: 2,
					Build: 3,
				},
			}
			fakeActor.GetAndValidatePluginReturns(plugin, nil)

			tempFile, err := ioutil.TempFile("", "")
			err = tempFile.Close()
			Expect(err).ToNot(HaveOccurred())
			downloadedPluginPath = tempFile.Name()
			err = ioutil.WriteFile(downloadedPluginPath, []byte("abcd"), 0700)
			Expect(err).ToNot(HaveOccurred())

			fakeActor.FetchPluginFromURLReturns(downloadedPluginPath, nil)
		})

		AfterEach(func() {
			// downloadedPluginPath is cleaned up in Execute
		})

		Context("when the -f argument is given", func() {
			BeforeEach(func() {
				cmd.Force = true
			})

			Context("when the plugin is invalid", func() {
				var returnedErr error

				BeforeEach(func() {
					returnedErr = pluginaction.PluginInvalidError{Path: string(cmd.OptionalArgs.PathURLOrPluginName)}
					fakeActor.GetAndValidatePluginReturns(configv3.Plugin{}, returnedErr)
				})

				It("returns an error", func() {
					Expect(executeErr).To(MatchError(shared.PluginInvalidError{
						Path:              string(cmd.OptionalArgs.PathURLOrPluginName),
						WrappedErrMessage: returnedErr.Error()}))

					Expect(testUI.Out).ToNot(Say("Installing plugin"))
				})
			})

			Context("when the plugin is already installed", func() {
				BeforeEach(func() {
					fakeActor.IsPluginInstalledReturns(true)
				})

				Context("when an error is encountered uninstalling the existing plugin", func() {
					BeforeEach(func() {
						expectedErr = errors.New("uninstall plugin error")
						fakeActor.UninstallPluginReturns(expectedErr)
					})

					It("returns the error", func() {
						Expect(executeErr).To(MatchError(expectedErr))

						Expect(testUI.Out).ToNot(Say("Plugin some-plugin successfully uninstalled\\."))
					})
				})

				Context("when no errors are encountered uninstalling the existing plugin", func() {
					It("uninstalls the existing plugin and installs the current plugin", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say("Plugin %s 1\\.2\\.3 is already installed\\. Uninstalling existing plugin\\.\\.\\.", pluginName))
						Expect(testUI.Out).To(Say("OK"))
						Expect(testUI.Out).To(Say("Plugin %s successfully uninstalled\\.", pluginName))
						Expect(testUI.Out).To(Say("Installing plugin %s\\.\\.\\.", pluginName))
						Expect(testUI.Out).To(Say("OK"))
						Expect(testUI.Out).To(Say("Plugin %s 1\\.2\\.3 successfully installed\\.", pluginName))

						Expect(fakeActor.FileExistsCallCount()).To(Equal(0))
						Expect(fakeActor.CreateExecutableCopyCallCount()).To(Equal(0))

						Expect(fakeActor.FetchPluginFromURLCallCount()).To(Equal(1))
						requestURL := fakeActor.FetchPluginFromURLArgsForCall(0)
						Expect(requestURL).To(BeEquivalentTo(cmd.OptionalArgs.PathURLOrPluginName))

						Expect(fakeActor.GetAndValidatePluginCallCount()).To(Equal(1))
						_, _, path := fakeActor.GetAndValidatePluginArgsForCall(0)
						Expect(path).To(Equal(downloadedPluginPath))

						Expect(fakeActor.IsPluginInstalledCallCount()).To(Equal(1))
						Expect(fakeActor.IsPluginInstalledArgsForCall(0)).To(Equal(pluginName))

						Expect(fakeActor.UninstallPluginCallCount()).To(Equal(1))
						_, uninstalledPluginName := fakeActor.UninstallPluginArgsForCall(0)
						Expect(uninstalledPluginName).To(Equal(pluginName))

						Expect(fakeActor.InstallPluginFromPathCallCount()).To(Equal(1))
						path, installedPlugin := fakeActor.InstallPluginFromPathArgsForCall(0)
						Expect(path).To(Equal(downloadedPluginPath))
						Expect(installedPlugin).To(Equal(plugin))
					})

					Context("when an error is encountered installing the plugin", func() {
						BeforeEach(func() {
							expectedErr = errors.New("install plugin error")
							fakeActor.InstallPluginFromPathReturns(expectedErr)
						})

						It("returns the error", func() {
							Expect(executeErr).To(MatchError(expectedErr))

							Expect(testUI.Out).ToNot(Say("Plugin some-plugin 1\\.2\\.3 successfully installed\\."))
						})
					})
				})
			})

			Context("when the plugin is not already installed", func() {
				It("installs the plugin", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).To(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
					Expect(testUI.Out).To(Say("Install and use plugins at your own risk\\."))
					Expect(testUI.Out).To(Say("Starting download of plugin binary from URL\\.\\.\\."))
					Expect(testUI.Out).To(Say("4 bytes downloaded\\.\\.\\."))
					Expect(testUI.Out).To(Say("Installing plugin %s\\.\\.\\.", pluginName))
					Expect(testUI.Out).To(Say("OK"))
					Expect(testUI.Out).To(Say("Plugin %s 1\\.2\\.3 successfully installed\\.", pluginName))

					Expect(fakeActor.FileExistsCallCount()).To(Equal(0))
					Expect(fakeActor.CreateExecutableCopyCallCount()).To(Equal(0))

					Expect(fakeActor.FetchPluginFromURLCallCount()).To(Equal(1))
					requestURL := fakeActor.FetchPluginFromURLArgsForCall(0)
					Expect(requestURL).To(BeEquivalentTo(cmd.OptionalArgs.PathURLOrPluginName))

					Expect(fakeActor.GetAndValidatePluginCallCount()).To(Equal(1))
					_, _, path := fakeActor.GetAndValidatePluginArgsForCall(0)
					Expect(path).To(Equal(downloadedPluginPath))

					Expect(fakeActor.IsPluginInstalledCallCount()).To(Equal(1))
					Expect(fakeActor.IsPluginInstalledArgsForCall(0)).To(Equal(pluginName))

					Expect(fakeActor.InstallPluginFromPathCallCount()).To(Equal(1))
					path, installedPlugin := fakeActor.InstallPluginFromPathArgsForCall(0)
					Expect(path).To(Equal(downloadedPluginPath))
					Expect(installedPlugin).To(Equal(plugin))

					Expect(fakeActor.UninstallPluginCallCount()).To(Equal(0))
				})
			})

			Context("when the -f argument is not given (user is prompted for confirmation)", func() {
				BeforeEach(func() {
					cmd.Force = false
				})

				Context("when the user chooses no", func() {
					BeforeEach(func() {
						input.Write([]byte("n\n"))
					})

					It("cancels plugin installation", func() {
						Expect(executeErr).To(MatchError(shared.PluginInstallationCancelled{}))

						Expect(testUI.Out).ToNot(Say("Installing plugin"))
					})
				})

				Context("when the user chooses the default", func() {
					BeforeEach(func() {
						input.Write([]byte("\n"))
					})

					It("cancels plugin installation", func() {
						Expect(executeErr).To(MatchError(shared.PluginInstallationCancelled{}))

						Expect(testUI.Out).ToNot(Say("Installing plugin"))
					})
				})

				Context("when the user input is invalid", func() {
					BeforeEach(func() {
						input.Write([]byte("e\n"))
					})

					It("returns an error", func() {
						Expect(executeErr).To(HaveOccurred())

						Expect(testUI.Out).ToNot(Say("Installing plugin"))
					})
				})

				Context("when the user chooses yes", func() {
					BeforeEach(func() {
						input.Write([]byte("y\n"))
					})

					Context("when the plugin is not already installed", func() {
						var (
							plugin               configv3.Plugin
							pluginName           string
							downloadedPluginPath string
						)

						BeforeEach(func() {
							pluginName = "some-plugin"
							plugin = configv3.Plugin{
								Name: pluginName,
								Version: configv3.PluginVersion{
									Major: 1,
									Minor: 2,
									Build: 3,
								},
							}
							fakeActor.GetAndValidatePluginReturns(plugin, nil)

							tempFile, err := ioutil.TempFile("", "")
							err = tempFile.Close()
							Expect(err).ToNot(HaveOccurred())
							downloadedPluginPath = tempFile.Name()
							err = ioutil.WriteFile(downloadedPluginPath, []byte("abcd"), 0700)
							Expect(err).ToNot(HaveOccurred())

							fakeActor.FetchPluginFromURLReturns(downloadedPluginPath, nil)
						})

						AfterEach(func() {
							// downloadedPluginPath is cleaned up in Execute
						})

						It("installs the plugin", func() {
							Expect(executeErr).ToNot(HaveOccurred())

							Expect(testUI.Out).To(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
							Expect(testUI.Out).To(Say("Install and use plugins at your own risk\\."))
							Expect(testUI.Out).To(Say("Do you want to install the plugin from %s\\? \\[yN\\]", cmd.OptionalArgs.PathURLOrPluginName))
							Expect(testUI.Out).To(Say("Starting download of plugin binary from URL\\.\\.\\."))
							Expect(testUI.Out).To(Say("4 bytes downloaded\\.\\.\\."))
							Expect(testUI.Out).To(Say("Installing plugin %s\\.\\.\\.", pluginName))
							Expect(testUI.Out).To(Say("OK"))
							Expect(testUI.Out).To(Say("Plugin %s 1\\.2\\.3 successfully installed\\.", pluginName))

							Expect(fakeActor.FileExistsCallCount()).To(Equal(0))
							Expect(fakeActor.CreateExecutableCopyCallCount()).To(Equal(0))

							Expect(fakeActor.FetchPluginFromURLCallCount()).To(Equal(1))
							requestURL := fakeActor.FetchPluginFromURLArgsForCall(0)
							Expect(requestURL).To(BeEquivalentTo(cmd.OptionalArgs.PathURLOrPluginName))

							Expect(fakeActor.GetAndValidatePluginCallCount()).To(Equal(1))
							_, _, path := fakeActor.GetAndValidatePluginArgsForCall(0)
							Expect(path).To(Equal(downloadedPluginPath))

							Expect(fakeActor.IsPluginInstalledCallCount()).To(Equal(1))
							Expect(fakeActor.IsPluginInstalledArgsForCall(0)).To(Equal(pluginName))

							Expect(fakeActor.InstallPluginFromPathCallCount()).To(Equal(1))
							path, installedPlugin := fakeActor.InstallPluginFromPathArgsForCall(0)
							Expect(path).To(Equal(downloadedPluginPath))
							Expect(installedPlugin).To(Equal(plugin))

							Expect(fakeActor.UninstallPluginCallCount()).To(Equal(0))
						})
					})

					Context("when the plugin is already installed", func() {
						BeforeEach(func() {
							plugin := configv3.Plugin{
								Name: pluginName,
								Version: configv3.PluginVersion{
									Major: 1,
									Minor: 2,
									Build: 3,
								},
							}
							fakeActor.GetAndValidatePluginReturns(plugin, nil)
							fakeActor.IsPluginInstalledReturns(true)
						})

						It("returns PluginAlreadyInstalledError", func() {
							Expect(executeErr).To(MatchError(shared.PluginAlreadyInstalledError{
								Name:    pluginName,
								Version: "1.2.3",
								Path:    string(cmd.OptionalArgs.PathURLOrPluginName),
							}))
						})
					})
				})
			})
		})
	})
})
