package common_test

import (
	"errors"
	"io/ioutil"
	"os"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/api/plugin/pluginerror"
	"code.cloudfoundry.org/cli/api/plugin/pluginfakes"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/common"
	"code.cloudfoundry.org/cli/command/common/commonfakes"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
)

var _ = Describe("install-plugin command", func() {
	var (
		cmd             InstallPluginCommand
		testUI          *ui.UI
		input           *Buffer
		fakeConfig      *commandfakes.FakeConfig
		fakeActor       *commonfakes.FakeInstallPluginActor
		fakeProgressBar *pluginfakes.FakeProxyReader
		executeErr      error
		expectedErr     error
		pluginHome      string
	)

	BeforeEach(func() {
		input = NewBuffer()
		testUI = ui.NewTestUI(input, NewBuffer(), NewBuffer())
		fakeConfig = new(commandfakes.FakeConfig)
		fakeActor = new(commonfakes.FakeInstallPluginActor)
		fakeProgressBar = new(pluginfakes.FakeProxyReader)

		cmd = InstallPluginCommand{
			UI:          testUI,
			Config:      fakeConfig,
			Actor:       fakeActor,
			ProgressBar: fakeProgressBar,
		}

		var err error
		pluginHome, err = ioutil.TempDir("", "some-pluginhome")
		Expect(err).NotTo(HaveOccurred())

		fakeConfig.PluginHomeReturns(pluginHome)
		fakeConfig.BinaryNameReturns("faceman")
	})

	AfterEach(func() {
		os.RemoveAll(pluginHome)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	Describe("installing from a local file", func() {
		BeforeEach(func() {
			cmd.OptionalArgs.PluginNameOrLocation = "some-path"
		})

		Context("when the local file does not exist", func() {
			BeforeEach(func() {
				fakeActor.FileExistsReturns(false)
			})

			It("does not print installation messages and returns a FileNotFoundError", func() {
				Expect(executeErr).To(MatchError(translatableerror.PluginNotFoundOnDiskOrInAnyRepositoryError{PluginName: "some-path", BinaryName: "faceman"}))

				Expect(testUI.Out).ToNot(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
				Expect(testUI.Out).ToNot(Say("Installing plugin some-path\\.\\.\\."))
			})
		})

		Context("when the file exists", func() {
			BeforeEach(func() {
				fakeActor.CreateExecutableCopyReturns("copy-path", nil)
				fakeActor.FileExistsReturns(true)
			})

			Context("when the -f argument is given", func() {
				BeforeEach(func() {
					cmd.Force = true
				})

				Context("when the plugin is invalid", func() {
					var returnedErr error

					BeforeEach(func() {
						returnedErr = actionerror.PluginInvalidError{}
						fakeActor.GetAndValidatePluginReturns(configv3.Plugin{}, returnedErr)
					})

					It("returns an error", func() {
						Expect(executeErr).To(MatchError(returnedErr))

						Expect(testUI.Out).ToNot(Say("Installing plugin"))
					})
				})

				Context("when the plugin is valid but generates an error when fetching metadata", func() {
					var wrappedErr error

					BeforeEach(func() {
						wrappedErr = errors.New("some-error")
						fakeActor.GetAndValidatePluginReturns(configv3.Plugin{}, actionerror.PluginInvalidError{Err: wrappedErr})
					})

					It("returns an error", func() {
						Expect(executeErr).To(MatchError(actionerror.PluginInvalidError{Err: wrappedErr}))

						Expect(testUI.Out).ToNot(Say("Installing plugin"))
					})
				})

				Context("when the plugin is already installed", func() {
					var (
						plugin    configv3.Plugin
						newPlugin configv3.Plugin
					)
					BeforeEach(func() {
						plugin = configv3.Plugin{
							Name: "some-plugin",
							Version: configv3.PluginVersion{
								Major: 1,
								Minor: 2,
								Build: 2,
							},
						}
						newPlugin = configv3.Plugin{
							Name: "some-plugin",
							Version: configv3.PluginVersion{
								Major: 1,
								Minor: 2,
								Build: 3,
							},
						}
						fakeActor.GetAndValidatePluginReturns(newPlugin, nil)
						fakeConfig.GetPluginCaseInsensitiveReturns(plugin, true)
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
							Expect(testUI.Out).To(Say("Plugin some-plugin 1\\.2\\.2 is already installed\\. Uninstalling existing plugin\\.\\.\\."))
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

							Expect(fakeConfig.GetPluginCaseInsensitiveCallCount()).To(Equal(1))
							Expect(fakeConfig.GetPluginCaseInsensitiveArgsForCall(0)).To(Equal("some-plugin"))

							Expect(fakeActor.UninstallPluginCallCount()).To(Equal(1))
							_, pluginName := fakeActor.UninstallPluginArgsForCall(0)
							Expect(pluginName).To(Equal("some-plugin"))

							Expect(fakeActor.InstallPluginFromPathCallCount()).To(Equal(1))
							path, installedPlugin := fakeActor.InstallPluginFromPathArgsForCall(0)
							Expect(path).To(Equal("copy-path"))
							Expect(installedPlugin).To(Equal(newPlugin))
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
						pathArg, pluginDirArg := fakeActor.CreateExecutableCopyArgsForCall(0)
						Expect(pathArg).To(Equal("some-path"))
						Expect(pluginDirArg).To(ContainSubstring("some-pluginhome"))
						Expect(pluginDirArg).To(ContainSubstring("temp"))

						Expect(fakeActor.GetAndValidatePluginCallCount()).To(Equal(1))
						_, _, path := fakeActor.GetAndValidatePluginArgsForCall(0)
						Expect(path).To(Equal("copy-path"))

						Expect(fakeConfig.GetPluginCaseInsensitiveCallCount()).To(Equal(1))
						Expect(fakeConfig.GetPluginCaseInsensitiveArgsForCall(0)).To(Equal("some-plugin"))

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
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say("Plugin installation cancelled\\."))
					})
				})

				Context("when the user chooses the default", func() {
					BeforeEach(func() {
						input.Write([]byte("\n"))
					})

					It("cancels plugin installation", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say("Plugin installation cancelled\\."))
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

							Expect(fakeConfig.GetPluginCaseInsensitiveCallCount()).To(Equal(1))
							Expect(fakeConfig.GetPluginCaseInsensitiveArgsForCall(0)).To(Equal("some-plugin"))

							Expect(fakeActor.InstallPluginFromPathCallCount()).To(Equal(1))
							path, plugin := fakeActor.InstallPluginFromPathArgsForCall(0)
							Expect(path).To(Equal("copy-path"))
							Expect(plugin).To(Equal(plugin))

							Expect(fakeActor.UninstallPluginCallCount()).To(Equal(0))
						})
					})

					Context("when the plugin is already installed", func() {
						BeforeEach(func() {
							fakeConfig.GetPluginCaseInsensitiveReturns(configv3.Plugin{
								Name: "some-plugin",
								Version: configv3.PluginVersion{
									Major: 1,
									Minor: 2,
									Build: 2,
								},
							}, true)
							fakeActor.GetAndValidatePluginReturns(configv3.Plugin{
								Name: "some-plugin",
								Version: configv3.PluginVersion{
									Major: 1,
									Minor: 2,
									Build: 3,
								},
							}, nil)
						})

						It("returns PluginAlreadyInstalledError", func() {
							Expect(executeErr).To(MatchError(translatableerror.PluginAlreadyInstalledError{
								BinaryName: "faceman",
								Name:       "some-plugin",
								Version:    "1.2.3",
							}))
						})
					})
				})
			})
		})
	})

	Describe("installing from an unsupported URL scheme", func() {
		BeforeEach(func() {
			cmd.OptionalArgs.PluginNameOrLocation = "ftp://some-url"
		})

		It("returns an error indicating an unsupported URL scheme", func() {
			Expect(executeErr).To(MatchError(translatableerror.UnsupportedURLSchemeError{
				UnsupportedURL: string(cmd.OptionalArgs.PluginNameOrLocation),
			}))
		})
	})

	Describe("installing from an HTTP URL", func() {
		var (
			plugin               configv3.Plugin
			pluginName           string
			executablePluginPath string
		)

		BeforeEach(func() {
			cmd.OptionalArgs.PluginNameOrLocation = "http://some-url"
			pluginName = "some-plugin"
			executablePluginPath = "executable-path"
		})

		It("displays the plugin warning", func() {
			Expect(testUI.Out).To(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
			Expect(testUI.Out).To(Say("Install and use plugins at your own risk\\."))
		})

		Context("when the -f argument is given", func() {
			BeforeEach(func() {
				cmd.Force = true
			})

			It("begins downloading the plugin", func() {
				Expect(testUI.Out).To(Say("Starting download of plugin binary from URL\\.\\.\\."))

				Expect(fakeActor.DownloadExecutableBinaryFromURLCallCount()).To(Equal(1))
				url, tempPluginDir, proxyReader := fakeActor.DownloadExecutableBinaryFromURLArgsForCall(0)
				Expect(url).To(Equal(cmd.OptionalArgs.PluginNameOrLocation.String()))
				Expect(tempPluginDir).To(ContainSubstring("some-pluginhome"))
				Expect(tempPluginDir).To(ContainSubstring("temp"))
				Expect(proxyReader).To(Equal(fakeProgressBar))
			})

			Context("When getting the binary fails", func() {
				BeforeEach(func() {
					expectedErr = errors.New("some-error")
					fakeActor.DownloadExecutableBinaryFromURLReturns("", expectedErr)
				})

				It("returns the error", func() {
					Expect(executeErr).To(MatchError(expectedErr))

					Expect(testUI.Out).ToNot(Say("downloaded"))
					Expect(fakeActor.GetAndValidatePluginCallCount()).To(Equal(0))
				})

				Context("when a 4xx or 5xx status is encountered while downloading the plugin", func() {
					BeforeEach(func() {
						fakeActor.DownloadExecutableBinaryFromURLReturns("", pluginerror.RawHTTPStatusError{Status: "some-status"})
					})

					It("returns a DownloadPluginHTTPError", func() {
						Expect(executeErr).To(MatchError(pluginerror.RawHTTPStatusError{Status: "some-status"}))
					})
				})

				Context("when a SSL error is encountered while downloading the plugin", func() {
					BeforeEach(func() {
						fakeActor.DownloadExecutableBinaryFromURLReturns("", pluginerror.UnverifiedServerError{})
					})

					It("returns a DownloadPluginHTTPError", func() {
						Expect(executeErr).To(MatchError(pluginerror.UnverifiedServerError{}))
					})
				})
			})

			Context("when getting the binary succeeds", func() {
				BeforeEach(func() {
					fakeActor.DownloadExecutableBinaryFromURLReturns("some-path", nil)
					fakeActor.CreateExecutableCopyReturns(executablePluginPath, nil)
				})

				It("sets up the progress bar", func() {
					Expect(fakeActor.GetAndValidatePluginCallCount()).To(Equal(1))
					_, _, path := fakeActor.GetAndValidatePluginArgsForCall(0)
					Expect(path).To(Equal(executablePluginPath))

					Expect(fakeActor.DownloadExecutableBinaryFromURLCallCount()).To(Equal(1))
					urlArg, pluginDirArg, proxyReader := fakeActor.DownloadExecutableBinaryFromURLArgsForCall(0)
					Expect(urlArg).To(Equal("http://some-url"))
					Expect(pluginDirArg).To(ContainSubstring("some-pluginhome"))
					Expect(pluginDirArg).To(ContainSubstring("temp"))
					Expect(proxyReader).To(Equal(fakeProgressBar))

					Expect(fakeActor.CreateExecutableCopyCallCount()).To(Equal(1))
					pathArg, pluginDirArg := fakeActor.CreateExecutableCopyArgsForCall(0)
					Expect(pathArg).To(Equal("some-path"))
					Expect(pluginDirArg).To(ContainSubstring("some-pluginhome"))
					Expect(pluginDirArg).To(ContainSubstring("temp"))
				})

				Context("when the plugin is invalid", func() {
					var returnedErr error

					BeforeEach(func() {
						returnedErr = actionerror.PluginInvalidError{}
						fakeActor.GetAndValidatePluginReturns(configv3.Plugin{}, returnedErr)
					})

					It("returns an error", func() {
						Expect(executeErr).To(MatchError(returnedErr))

						Expect(fakeConfig.GetPluginCaseInsensitiveCallCount()).To(Equal(0))
					})
				})

				Context("when the plugin is valid", func() {
					var newPlugin configv3.Plugin

					BeforeEach(func() {
						plugin = configv3.Plugin{
							Name: pluginName,
							Version: configv3.PluginVersion{
								Major: 1,
								Minor: 2,
								Build: 2,
							},
						}
						newPlugin = configv3.Plugin{
							Name: pluginName,
							Version: configv3.PluginVersion{
								Major: 1,
								Minor: 2,
								Build: 3,
							},
						}
						fakeActor.GetAndValidatePluginReturns(newPlugin, nil)
					})

					Context("when the plugin is already installed", func() {
						BeforeEach(func() {
							fakeConfig.GetPluginCaseInsensitiveReturns(plugin, true)
						})

						It("displays uninstall message", func() {
							Expect(testUI.Out).To(Say("Plugin %s 1\\.2\\.2 is already installed\\. Uninstalling existing plugin\\.\\.\\.", pluginName))
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
							It("displays uninstall message", func() {
								Expect(testUI.Out).To(Say("Plugin %s successfully uninstalled\\.", pluginName))
							})

							Context("when no errors are encountered installing the plugin", func() {
								It("uninstalls the existing plugin and installs the current plugin", func() {
									Expect(executeErr).ToNot(HaveOccurred())

									Expect(testUI.Out).To(Say("Installing plugin %s\\.\\.\\.", pluginName))
									Expect(testUI.Out).To(Say("OK"))
									Expect(testUI.Out).To(Say("Plugin %s 1\\.2\\.3 successfully installed\\.", pluginName))
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

					Context("when the plugin is not already installed", func() {
						It("installs the plugin", func() {
							Expect(executeErr).ToNot(HaveOccurred())

							Expect(testUI.Out).To(Say("Installing plugin %s\\.\\.\\.", pluginName))
							Expect(testUI.Out).To(Say("OK"))
							Expect(testUI.Out).To(Say("Plugin %s 1\\.2\\.3 successfully installed\\.", pluginName))

							Expect(fakeActor.UninstallPluginCallCount()).To(Equal(0))
						})
					})
				})
			})
		})

		Context("when the -f argument is not given (user is prompted for confirmation)", func() {
			BeforeEach(func() {
				plugin = configv3.Plugin{
					Name: pluginName,
					Version: configv3.PluginVersion{
						Major: 1,
						Minor: 2,
						Build: 3,
					},
				}

				cmd.Force = false
				fakeActor.DownloadExecutableBinaryFromURLReturns("some-path", nil)
				fakeActor.CreateExecutableCopyReturns("executable-path", nil)
			})

			Context("when the user chooses no", func() {
				BeforeEach(func() {
					input.Write([]byte("n\n"))
				})

				It("cancels plugin installation", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).To(Say("Plugin installation cancelled\\."))
				})
			})

			Context("when the user chooses the default", func() {
				BeforeEach(func() {
					input.Write([]byte("\n"))
				})

				It("cancels plugin installation", func() {
					Expect(executeErr).ToNot(HaveOccurred())

					Expect(testUI.Out).To(Say("Plugin installation cancelled\\."))
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
					BeforeEach(func() {
						fakeActor.GetAndValidatePluginReturns(plugin, nil)
					})

					It("installs the plugin", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
						Expect(testUI.Out).To(Say("Install and use plugins at your own risk\\."))
						Expect(testUI.Out).To(Say("Do you want to install the plugin %s\\? \\[yN\\]", cmd.OptionalArgs.PluginNameOrLocation))
						Expect(testUI.Out).To(Say("Starting download of plugin binary from URL\\.\\.\\."))
						Expect(testUI.Out).To(Say("Installing plugin %s\\.\\.\\.", pluginName))
						Expect(testUI.Out).To(Say("OK"))
						Expect(testUI.Out).To(Say("Plugin %s 1\\.2\\.3 successfully installed\\.", pluginName))

						Expect(fakeActor.DownloadExecutableBinaryFromURLCallCount()).To(Equal(1))
						url, tempPluginDir, proxyReader := fakeActor.DownloadExecutableBinaryFromURLArgsForCall(0)
						Expect(url).To(Equal(cmd.OptionalArgs.PluginNameOrLocation.String()))
						Expect(tempPluginDir).To(ContainSubstring("some-pluginhome"))
						Expect(tempPluginDir).To(ContainSubstring("temp"))
						Expect(proxyReader).To(Equal(fakeProgressBar))

						Expect(fakeActor.CreateExecutableCopyCallCount()).To(Equal(1))
						path, tempPluginDir := fakeActor.CreateExecutableCopyArgsForCall(0)
						Expect(path).To(Equal("some-path"))
						Expect(tempPluginDir).To(ContainSubstring("some-pluginhome"))
						Expect(tempPluginDir).To(ContainSubstring("temp"))

						Expect(fakeActor.GetAndValidatePluginCallCount()).To(Equal(1))
						_, _, path = fakeActor.GetAndValidatePluginArgsForCall(0)
						Expect(path).To(Equal(executablePluginPath))

						Expect(fakeConfig.GetPluginCaseInsensitiveCallCount()).To(Equal(1))
						Expect(fakeConfig.GetPluginCaseInsensitiveArgsForCall(0)).To(Equal(pluginName))

						Expect(fakeActor.InstallPluginFromPathCallCount()).To(Equal(1))
						path, installedPlugin := fakeActor.InstallPluginFromPathArgsForCall(0)
						Expect(path).To(Equal(executablePluginPath))
						Expect(installedPlugin).To(Equal(plugin))

						Expect(fakeActor.UninstallPluginCallCount()).To(Equal(0))
					})
				})

				Context("when the plugin is already installed", func() {
					BeforeEach(func() {
						fakeConfig.GetPluginCaseInsensitiveReturns(configv3.Plugin{
							Name: "some-plugin",
							Version: configv3.PluginVersion{
								Major: 1,
								Minor: 2,
								Build: 2,
							},
						}, true)
						fakeActor.GetAndValidatePluginReturns(configv3.Plugin{
							Name: "some-plugin",
							Version: configv3.PluginVersion{
								Major: 1,
								Minor: 2,
								Build: 3,
							},
						}, nil)
					})

					It("returns PluginAlreadyInstalledError", func() {
						Expect(executeErr).To(MatchError(translatableerror.PluginAlreadyInstalledError{
							BinaryName: "faceman",
							Name:       pluginName,
							Version:    "1.2.3",
						}))
					})
				})
			})
		})
	})
})
