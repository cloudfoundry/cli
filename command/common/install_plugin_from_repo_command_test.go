package common_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"strconv"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/pluginaction"
	"code.cloudfoundry.org/cli/api/plugin/pluginerror"
	"code.cloudfoundry.org/cli/api/plugin/pluginfakes"
	"code.cloudfoundry.org/cli/command/commandfakes"
	. "code.cloudfoundry.org/cli/command/common"
	"code.cloudfoundry.org/cli/command/common/commonfakes"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/ui"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
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
		binaryName      string
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

		tmpDirectorySeed := strconv.Itoa(int(rand.Int63()))
		pluginHome = fmt.Sprintf("some-pluginhome-%s", tmpDirectorySeed)
		fakeConfig.PluginHomeReturns(pluginHome)
		binaryName = helpers.PrefixedRandomName("bin")
		fakeConfig.BinaryNameReturns(binaryName)
	})

	AfterEach(func() {
		os.RemoveAll(pluginHome)
	})

	JustBeforeEach(func() {
		executeErr = cmd.Execute(nil)
	})

	Describe("installing from a specific repo", func() {
		var (
			pluginName string
			pluginURL  string
			repoName   string
			repoURL    string
		)

		BeforeEach(func() {
			pluginName = helpers.PrefixedRandomName("plugin")
			pluginURL = helpers.PrefixedRandomName("http://")
			repoName = helpers.PrefixedRandomName("repo")
			repoURL = helpers.PrefixedRandomName("http://")
			cmd.OptionalArgs.PluginNameOrLocation = flag.Path(pluginName)
			cmd.RegisteredRepository = repoName
		})

		Context("when the repo is not registered", func() {
			BeforeEach(func() {
				fakeActor.GetPluginRepositoryReturns(configv3.PluginRepository{}, actionerror.RepositoryNotRegisteredError{Name: repoName})
			})

			It("returns a RepositoryNotRegisteredError", func() {
				Expect(executeErr).To(MatchError(actionerror.RepositoryNotRegisteredError{Name: repoName}))

				Expect(fakeActor.GetPluginRepositoryCallCount()).To(Equal(1))
				repositoryNameArg := fakeActor.GetPluginRepositoryArgsForCall(0)
				Expect(repositoryNameArg).To(Equal(repoName))
			})
		})

		Context("when the repository is registered", func() {
			var platform string

			BeforeEach(func() {
				platform = helpers.PrefixedRandomName("platform")
				fakeActor.GetPlatformStringReturns(platform)
				fakeActor.GetPluginRepositoryReturns(configv3.PluginRepository{Name: repoName, URL: repoURL}, nil)
			})

			Context("when getting repository information returns a json syntax error", func() {
				var jsonErr error
				BeforeEach(func() {
					jsonErr = &json.SyntaxError{}
					fakeActor.GetPluginInfoFromRepositoriesForPlatformReturns(pluginaction.PluginInfo{}, nil, jsonErr)
				})

				It("returns a JSONSyntaxError", func() {
					Expect(executeErr).To(MatchError(jsonErr))
				})
			})

			Context("when getting the repository information errors", func() {
				Context("with a generic error", func() {
					BeforeEach(func() {
						expectedErr = errors.New("some-client-error")
						fakeActor.GetPluginInfoFromRepositoriesForPlatformReturns(pluginaction.PluginInfo{}, nil, actionerror.FetchingPluginInfoFromRepositoryError{
							RepositoryName: "some-repo",
							Err:            expectedErr,
						})
					})

					It("returns the wrapped client(request/http status) error", func() {
						Expect(executeErr).To(MatchError(expectedErr))
					})
				})

				Context("with a RawHTTPStatusError error", func() {
					var returnedErr pluginerror.RawHTTPStatusError

					BeforeEach(func() {
						returnedErr = pluginerror.RawHTTPStatusError{Status: "some-status"}
						fakeActor.GetPluginInfoFromRepositoriesForPlatformReturns(pluginaction.PluginInfo{}, nil, actionerror.FetchingPluginInfoFromRepositoryError{
							RepositoryName: "some-repo",
							Err:            returnedErr,
						})
					})

					It("returns the wrapped client(request/http status) error", func() {
						Expect(executeErr).To(MatchError(returnedErr))
					})
				})
			})

			Context("when the plugin can't be found in the repository", func() {
				BeforeEach(func() {
					fakeActor.GetPluginInfoFromRepositoriesForPlatformReturns(pluginaction.PluginInfo{}, nil, actionerror.PluginNotFoundInAnyRepositoryError{PluginName: pluginName})
				})

				It("returns the PluginNotFoundInRepositoryError", func() {
					Expect(executeErr).To(MatchError(translatableerror.PluginNotFoundInRepositoryError{BinaryName: binaryName, PluginName: pluginName, RepositoryName: repoName}))

					Expect(fakeActor.GetPlatformStringCallCount()).To(Equal(1))
					platformGOOS, platformGOARCH := fakeActor.GetPlatformStringArgsForCall(0)
					Expect(platformGOOS).To(Equal(runtime.GOOS))
					Expect(platformGOARCH).To(Equal(runtime.GOARCH))

					Expect(fakeActor.GetPluginInfoFromRepositoriesForPlatformCallCount()).To(Equal(1))
					pluginNameArg, pluginRepositoriesArg, pluginPlatform := fakeActor.GetPluginInfoFromRepositoriesForPlatformArgsForCall(0)
					Expect(pluginNameArg).To(Equal(pluginName))
					Expect(pluginRepositoriesArg).To(Equal([]configv3.PluginRepository{{Name: repoName, URL: repoURL}}))
					Expect(pluginPlatform).To(Equal(platform))
				})
			})

			Context("when a compatible binary can't be found in the repository", func() {
				BeforeEach(func() {
					fakeActor.GetPluginInfoFromRepositoriesForPlatformReturns(pluginaction.PluginInfo{}, nil, actionerror.NoCompatibleBinaryError{})
				})

				It("returns the NoCompatibleBinaryError", func() {
					Expect(executeErr).To(MatchError(actionerror.NoCompatibleBinaryError{}))
				})
			})

			Context("when the plugin is found", func() {
				var (
					checksum                string
					downloadedVersionString string
				)

				BeforeEach(func() {
					checksum = helpers.PrefixedRandomName("checksum")
					downloadedVersionString = helpers.PrefixedRandomName("version")

					fakeActor.GetPluginInfoFromRepositoriesForPlatformReturns(pluginaction.PluginInfo{Name: pluginName, Version: downloadedVersionString, URL: pluginURL, Checksum: checksum}, []string{repoName}, nil)
				})

				Context("when the -f argument is given", func() {
					BeforeEach(func() {
						cmd.Force = true
					})

					Context("when the plugin is already installed", func() {
						BeforeEach(func() {
							plugin := configv3.Plugin{
								Name:    pluginName,
								Version: configv3.PluginVersion{Major: 1, Minor: 2, Build: 2},
							}
							fakeConfig.GetPluginReturns(plugin, true)
							fakeConfig.GetPluginCaseInsensitiveReturns(plugin, true)
						})

						Context("when getting the binary errors", func() {
							BeforeEach(func() {
								expectedErr = errors.New("some-error")
								fakeActor.DownloadExecutableBinaryFromURLReturns("", expectedErr)
							})

							It("returns the error", func() {
								Expect(executeErr).To(MatchError(expectedErr))

								Expect(testUI.Out).To(Say("Searching %s for plugin %s\\.\\.\\.", repoName, pluginName))
								Expect(testUI.Out).To(Say("Plugin %s %s found in: %s", pluginName, downloadedVersionString, repoName))
								Expect(testUI.Out).To(Say("Plugin %s 1\\.2\\.2 is already installed\\.", pluginName))
								Expect(testUI.Out).To(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
								Expect(testUI.Out).To(Say("Install and use plugins at your own risk\\."))
								Expect(testUI.Out).To(Say("Starting download of plugin binary from repository %s\\.\\.\\.", repoName))

								Expect(testUI.Out).ToNot(Say("downloaded"))
								Expect(fakeActor.GetAndValidatePluginCallCount()).To(Equal(0))

								Expect(fakeConfig.GetPluginCallCount()).To(Equal(1))
								Expect(fakeConfig.GetPluginArgsForCall(0)).To(Equal(pluginName))

								Expect(fakeActor.GetPlatformStringCallCount()).To(Equal(1))
								platformGOOS, platformGOARCH := fakeActor.GetPlatformStringArgsForCall(0)
								Expect(platformGOOS).To(Equal(runtime.GOOS))
								Expect(platformGOARCH).To(Equal(runtime.GOARCH))

								Expect(fakeActor.GetPluginInfoFromRepositoriesForPlatformCallCount()).To(Equal(1))
								pluginNameArg, pluginRepositoriesArg, pluginPlatform := fakeActor.GetPluginInfoFromRepositoriesForPlatformArgsForCall(0)
								Expect(pluginNameArg).To(Equal(pluginName))
								Expect(pluginRepositoriesArg).To(Equal([]configv3.PluginRepository{{Name: repoName, URL: repoURL}}))
								Expect(pluginPlatform).To(Equal(platform))

								Expect(fakeActor.DownloadExecutableBinaryFromURLCallCount()).To(Equal(1))
								urlArg, dirArg, proxyReader := fakeActor.DownloadExecutableBinaryFromURLArgsForCall(0)
								Expect(urlArg).To(Equal(pluginURL))
								Expect(dirArg).To(ContainSubstring("temp"))
								Expect(proxyReader).To(Equal(fakeProgressBar))
							})
						})

						Context("when getting the binary succeeds", func() {
							var execPath string

							BeforeEach(func() {
								execPath = helpers.PrefixedRandomName("some-path")
								fakeActor.DownloadExecutableBinaryFromURLReturns(execPath, nil)
							})

							Context("when the checksum fails", func() {
								BeforeEach(func() {
									fakeActor.ValidateFileChecksumReturns(false)
								})

								It("returns the checksum error", func() {
									Expect(executeErr).To(MatchError(translatableerror.InvalidChecksumError{}))

									Expect(testUI.Out).To(Say("Searching %s for plugin %s\\.\\.\\.", repoName, pluginName))
									Expect(testUI.Out).To(Say("Plugin %s %s found in: %s", pluginName, downloadedVersionString, repoName))
									Expect(testUI.Out).To(Say("Plugin %s 1\\.2\\.2 is already installed\\.", pluginName))
									Expect(testUI.Out).To(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
									Expect(testUI.Out).To(Say("Install and use plugins at your own risk\\."))
									Expect(testUI.Out).To(Say("Starting download of plugin binary from repository %s\\.\\.\\.", repoName))
									Expect(testUI.Out).ToNot(Say("Installing plugin"))

									Expect(fakeActor.ValidateFileChecksumCallCount()).To(Equal(1))
									pathArg, checksumArg := fakeActor.ValidateFileChecksumArgsForCall(0)
									Expect(pathArg).To(Equal(execPath))
									Expect(checksumArg).To(Equal(checksum))
								})
							})

							Context("when the checksum succeeds", func() {
								BeforeEach(func() {
									fakeActor.ValidateFileChecksumReturns(true)
								})

								Context("when creating an executable copy errors", func() {
									BeforeEach(func() {
										fakeActor.CreateExecutableCopyReturns("", errors.New("some-error"))
									})

									It("returns the error", func() {
										Expect(executeErr).To(MatchError(errors.New("some-error")))
										Expect(testUI.Out).ToNot(Say("Installing plugin"))

										Expect(fakeActor.CreateExecutableCopyCallCount()).To(Equal(1))
										pathArg, tempDirArg := fakeActor.CreateExecutableCopyArgsForCall(0)
										Expect(pathArg).To(Equal(execPath))
										Expect(tempDirArg).To(ContainSubstring("temp"))
									})
								})

								Context("when creating an exectuable copy succeeds", func() {
									BeforeEach(func() {
										fakeActor.CreateExecutableCopyReturns("copy-path", nil)
									})

									Context("when validating the new plugin errors", func() {
										BeforeEach(func() {
											fakeActor.GetAndValidatePluginReturns(configv3.Plugin{}, actionerror.PluginInvalidError{})
										})

										It("returns the error", func() {
											Expect(executeErr).To(MatchError(actionerror.PluginInvalidError{}))
											Expect(testUI.Out).ToNot(Say("Installing plugin"))

											Expect(fakeActor.GetAndValidatePluginCallCount()).To(Equal(1))
											_, commandsArg, tempDirArg := fakeActor.GetAndValidatePluginArgsForCall(0)
											Expect(commandsArg).To(Equal(Commands))
											Expect(tempDirArg).To(Equal("copy-path"))
										})
									})

									Context("when validating the new plugin succeeds", func() {
										var (
											pluginVersion      configv3.PluginVersion
											pluginVersionRegex string
										)

										BeforeEach(func() {
											major := rand.Int()
											minor := rand.Int()
											build := rand.Int()
											pluginVersion = configv3.PluginVersion{Major: major, Minor: minor, Build: build}
											pluginVersionRegex = fmt.Sprintf("%d\\.%d\\.%d", major, minor, build)

											fakeActor.GetAndValidatePluginReturns(configv3.Plugin{
												Name:    pluginName,
												Version: pluginVersion,
											}, nil)
										})

										Context("when uninstalling the existing errors", func() {
											BeforeEach(func() {
												expectedErr = errors.New("uninstall plugin error")
												fakeActor.UninstallPluginReturns(expectedErr)
											})

											It("returns the error", func() {
												Expect(executeErr).To(MatchError(expectedErr))

												Expect(testUI.Out).To(Say("Uninstalling existing plugin\\.\\.\\."))
												Expect(testUI.Out).ToNot(Say("Plugin %s successfully uninstalled\\.", pluginName))

												Expect(fakeActor.UninstallPluginCallCount()).To(Equal(1))
												_, pluginNameArg := fakeActor.UninstallPluginArgsForCall(0)
												Expect(pluginNameArg).To(Equal(pluginName))
											})
										})

										Context("when uninstalling the existing plugin succeeds", func() {
											Context("when installing the new plugin errors", func() {
												BeforeEach(func() {
													expectedErr = errors.New("install plugin error")
													fakeActor.InstallPluginFromPathReturns(expectedErr)
												})

												It("returns the error", func() {
													Expect(executeErr).To(MatchError(expectedErr))

													Expect(testUI.Out).To(Say("Plugin %s successfully uninstalled\\.", pluginName))
													Expect(testUI.Out).To(Say("Installing plugin %s\\.\\.\\.", pluginName))
													Expect(testUI.Out).ToNot(Say("successfully installed"))

													Expect(fakeActor.InstallPluginFromPathCallCount()).To(Equal(1))
													pathArg, pluginArg := fakeActor.InstallPluginFromPathArgsForCall(0)
													Expect(pathArg).To(Equal("copy-path"))
													Expect(pluginArg).To(Equal(configv3.Plugin{
														Name:    pluginName,
														Version: pluginVersion,
													}))
												})
											})

											Context("when installing the new plugin succeeds", func() {
												It("uninstalls the existing plugin and installs the new one", func() {
													Expect(executeErr).ToNot(HaveOccurred())

													Expect(testUI.Out).To(Say("Searching %s for plugin %s\\.\\.\\.", repoName, pluginName))
													Expect(testUI.Out).To(Say("Plugin %s %s found in: %s", pluginName, downloadedVersionString, repoName))
													Expect(testUI.Out).To(Say("Plugin %s 1\\.2\\.2 is already installed\\.", pluginName))
													Expect(testUI.Out).To(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
													Expect(testUI.Out).To(Say("Install and use plugins at your own risk\\."))
													Expect(testUI.Out).To(Say("Starting download of plugin binary from repository %s\\.\\.\\.", repoName))
													Expect(testUI.Out).To(Say("Uninstalling existing plugin\\.\\.\\."))
													Expect(testUI.Out).To(Say("OK"))
													Expect(testUI.Out).To(Say("Plugin %s successfully uninstalled\\.", pluginName))
													Expect(testUI.Out).To(Say("Installing plugin %s\\.\\.\\.", pluginName))
													Expect(testUI.Out).To(Say("OK"))
													Expect(testUI.Out).To(Say("%s %s successfully installed\\.", pluginName, pluginVersionRegex))
												})
											})
										})
									})
								})
							})
						})
					})

					Context("when the plugin is NOT already installed", func() {
						Context("when getting the binary errors", func() {
							BeforeEach(func() {
								expectedErr = errors.New("some-error")
								fakeActor.DownloadExecutableBinaryFromURLReturns("", expectedErr)
							})

							It("returns the error", func() {
								Expect(executeErr).To(MatchError(expectedErr))

								Expect(testUI.Out).To(Say("Searching %s for plugin %s\\.\\.\\.", repoName, pluginName))
								Expect(testUI.Out).To(Say("Plugin %s %s found in: %s", pluginName, downloadedVersionString, repoName))
								Expect(testUI.Out).To(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
								Expect(testUI.Out).To(Say("Install and use plugins at your own risk\\."))
								Expect(testUI.Out).To(Say("Starting download of plugin binary from repository %s\\.\\.\\.", repoName))

								Expect(testUI.Out).ToNot(Say("downloaded"))
								Expect(fakeActor.GetAndValidatePluginCallCount()).To(Equal(0))
							})
						})

						Context("when getting the binary succeeds", func() {
							BeforeEach(func() {
								fakeActor.DownloadExecutableBinaryFromURLReturns("some-path", nil)
							})

							Context("when the checksum fails", func() {
								BeforeEach(func() {
									fakeActor.ValidateFileChecksumReturns(false)
								})

								It("returns the checksum error", func() {
									Expect(executeErr).To(MatchError(translatableerror.InvalidChecksumError{}))

									Expect(testUI.Out).To(Say("Searching %s for plugin %s\\.\\.\\.", repoName, pluginName))
									Expect(testUI.Out).To(Say("Plugin %s %s found in: %s", pluginName, downloadedVersionString, repoName))
									Expect(testUI.Out).To(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
									Expect(testUI.Out).To(Say("Install and use plugins at your own risk\\."))
									Expect(testUI.Out).To(Say("Starting download of plugin binary from repository %s\\.\\.\\.", repoName))
									Expect(testUI.Out).ToNot(Say("Installing plugin"))
								})
							})

							Context("when the checksum succeeds", func() {
								BeforeEach(func() {
									fakeActor.ValidateFileChecksumReturns(true)
								})

								Context("when creating an executable copy errors", func() {
									BeforeEach(func() {
										fakeActor.CreateExecutableCopyReturns("", errors.New("some-error"))
									})

									It("returns the error", func() {
										Expect(executeErr).To(MatchError(errors.New("some-error")))
										Expect(testUI.Out).ToNot(Say("Installing plugin"))
									})
								})

								Context("when creating an executable copy succeeds", func() {
									BeforeEach(func() {
										fakeActor.CreateExecutableCopyReturns("copy-path", nil)
									})

									Context("when validating the plugin errors", func() {
										BeforeEach(func() {
											fakeActor.GetAndValidatePluginReturns(configv3.Plugin{}, actionerror.PluginInvalidError{})
										})

										It("returns the error", func() {
											Expect(executeErr).To(MatchError(actionerror.PluginInvalidError{}))
											Expect(testUI.Out).ToNot(Say("Installing plugin"))
										})
									})

									Context("when validating the plugin succeeds", func() {
										BeforeEach(func() {
											fakeActor.GetAndValidatePluginReturns(configv3.Plugin{
												Name:    pluginName,
												Version: configv3.PluginVersion{Major: 1, Minor: 2, Build: 3},
											}, nil)
										})

										Context("when installing the plugin errors", func() {
											BeforeEach(func() {
												expectedErr = errors.New("install plugin error")
												fakeActor.InstallPluginFromPathReturns(expectedErr)
											})

											It("returns the error", func() {
												Expect(executeErr).To(MatchError(expectedErr))

												Expect(testUI.Out).To(Say("Installing plugin %s\\.\\.\\.", pluginName))
												Expect(testUI.Out).ToNot(Say("successfully installed"))
											})
										})

										Context("when installing the plugin succeeds", func() {
											It("uninstalls the existing plugin and installs the new one", func() {
												Expect(executeErr).ToNot(HaveOccurred())

												Expect(testUI.Out).To(Say("Searching %s for plugin %s\\.\\.\\.", repoName, pluginName))
												Expect(testUI.Out).To(Say("Plugin %s %s found in: %s", pluginName, downloadedVersionString, repoName))
												Expect(testUI.Out).To(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
												Expect(testUI.Out).To(Say("Install and use plugins at your own risk\\."))
												Expect(testUI.Out).To(Say("Starting download of plugin binary from repository %s\\.\\.\\.", repoName))
												Expect(testUI.Out).To(Say("Installing plugin %s\\.\\.\\.", pluginName))
												Expect(testUI.Out).To(Say("OK"))
												Expect(testUI.Out).To(Say("%s 1\\.2\\.3 successfully installed", pluginName))
											})
										})
									})
								})
							})
						})
					})
				})

				Context("when the -f argument is not given (user is prompted for confirmation)", func() {
					BeforeEach(func() {
						fakeActor.ValidateFileChecksumReturns(true)
					})

					Context("when the plugin is already installed", func() {
						BeforeEach(func() {
							plugin := configv3.Plugin{
								Name:    pluginName,
								Version: configv3.PluginVersion{Major: 1, Minor: 2, Build: 2},
							}
							fakeConfig.GetPluginReturns(plugin, true)
							fakeConfig.GetPluginCaseInsensitiveReturns(plugin, true)
						})

						Context("when the user chooses no", func() {
							BeforeEach(func() {
								input.Write([]byte("n\n"))
							})

							It("cancels plugin installation", func() {
								Expect(executeErr).ToNot(HaveOccurred())

								Expect(testUI.Out).To(Say("Searching %s for plugin %s\\.\\.\\.", repoName, pluginName))
								Expect(testUI.Out).To(Say("Plugin %s %s found in: %s", pluginName, downloadedVersionString, repoName))
								Expect(testUI.Out).To(Say("Plugin %s 1\\.2\\.2 is already installed\\.", pluginName))
								Expect(testUI.Out).To(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
								Expect(testUI.Out).To(Say("Install and use plugins at your own risk\\."))
								Expect(testUI.Out).To(Say("Do you want to uninstall the existing plugin and install %s %s\\? \\[yN\\]", pluginName, downloadedVersionString))
								Expect(testUI.Out).To(Say("Plugin installation cancelled\\."))
							})
						})

						Context("when the user chooses the default", func() {
							BeforeEach(func() {
								input.Write([]byte("\n"))
							})

							It("cancels plugin installation", func() {
								Expect(executeErr).ToNot(HaveOccurred())

								Expect(testUI.Out).To(Say("Do you want to uninstall the existing plugin and install %s %s\\? \\[yN\\]", pluginName, downloadedVersionString))
								Expect(testUI.Out).To(Say("Plugin installation cancelled\\."))
							})
						})

						Context("when the user input is invalid", func() {
							BeforeEach(func() {
								input.Write([]byte("e\n"))
							})

							It("returns an error", func() {
								Expect(executeErr).To(HaveOccurred())

								Expect(testUI.Out).To(Say("Do you want to uninstall the existing plugin and install %s %s\\? \\[yN\\]", pluginName, downloadedVersionString))
								Expect(testUI.Out).ToNot(Say("Plugin installation cancelled\\."))
								Expect(testUI.Out).ToNot(Say("Installing plugin"))
							})
						})

						Context("when the user chooses yes", func() {
							BeforeEach(func() {
								input.Write([]byte("y\n"))
								fakeActor.DownloadExecutableBinaryFromURLReturns("some-path", nil)
								fakeActor.CreateExecutableCopyReturns("copy-path", nil)
								fakeActor.GetAndValidatePluginReturns(configv3.Plugin{
									Name:    pluginName,
									Version: configv3.PluginVersion{Major: 1, Minor: 2, Build: 3},
								}, nil)
							})

							It("installs the plugin", func() {
								Expect(testUI.Out).To(Say("Searching %s for plugin %s\\.\\.\\.", repoName, pluginName))
								Expect(testUI.Out).To(Say("Plugin %s %s found in: %s", pluginName, downloadedVersionString, repoName))
								Expect(testUI.Out).To(Say("Plugin %s 1\\.2\\.2 is already installed\\.", pluginName))
								Expect(testUI.Out).To(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
								Expect(testUI.Out).To(Say("Install and use plugins at your own risk\\."))
								Expect(testUI.Out).To(Say("Do you want to uninstall the existing plugin and install %s %s\\? \\[yN\\]", pluginName, downloadedVersionString))
								Expect(testUI.Out).To(Say("Starting download of plugin binary from repository %s\\.\\.\\.", repoName))
								Expect(testUI.Out).To(Say("Uninstalling existing plugin\\.\\.\\."))
								Expect(testUI.Out).To(Say("OK"))
								Expect(testUI.Out).To(Say("Plugin %s successfully uninstalled\\.", pluginName))
								Expect(testUI.Out).To(Say("Installing plugin %s\\.\\.\\.", pluginName))
								Expect(testUI.Out).To(Say("OK"))
								Expect(testUI.Out).To(Say("%s 1\\.2\\.3 successfully installed", pluginName))
							})
						})
					})

					Context("when the plugin is NOT already installed", func() {
						Context("when the user chooses no", func() {
							BeforeEach(func() {
								input.Write([]byte("n\n"))
							})

							It("cancels plugin installation", func() {
								Expect(executeErr).ToNot(HaveOccurred())

								Expect(testUI.Out).To(Say("Searching %s for plugin %s\\.\\.\\.", repoName, pluginName))
								Expect(testUI.Out).To(Say("Plugin %s %s found in: %s", pluginName, downloadedVersionString, repoName))
								Expect(testUI.Out).To(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
								Expect(testUI.Out).To(Say("Install and use plugins at your own risk\\."))
								Expect(testUI.Out).To(Say("Do you want to install the plugin %s\\? \\[yN\\]", pluginName))
								Expect(testUI.Out).To(Say("Plugin installation cancelled\\."))
							})
						})

						Context("when the user chooses the default", func() {
							BeforeEach(func() {
								input.Write([]byte("\n"))
							})

							It("cancels plugin installation", func() {
								Expect(executeErr).ToNot(HaveOccurred())

								Expect(testUI.Out).To(Say("Do you want to install the plugin %s\\? \\[yN\\]", pluginName))
								Expect(testUI.Out).To(Say("Plugin installation cancelled\\."))
							})
						})

						Context("when the user input is invalid", func() {
							BeforeEach(func() {
								input.Write([]byte("e\n"))
							})

							It("returns an error", func() {
								Expect(executeErr).To(HaveOccurred())

								Expect(testUI.Out).To(Say("Do you want to install the plugin %s\\? \\[yN\\]", pluginName))
								Expect(testUI.Out).ToNot(Say("Plugin installation cancelled\\."))
								Expect(testUI.Out).ToNot(Say("Installing plugin"))
							})
						})

						Context("when the user chooses yes", func() {
							var execPath string

							BeforeEach(func() {
								input.Write([]byte("y\n"))
								execPath = helpers.PrefixedRandomName("some-path")
								fakeActor.DownloadExecutableBinaryFromURLReturns(execPath, nil)
								fakeActor.CreateExecutableCopyReturns("copy-path", nil)
								fakeActor.GetAndValidatePluginReturns(configv3.Plugin{
									Name:    pluginName,
									Version: configv3.PluginVersion{Major: 1, Minor: 2, Build: 3},
								}, nil)
							})

							It("installs the plugin", func() {
								Expect(executeErr).ToNot(HaveOccurred())

								Expect(testUI.Out).To(Say("Searching %s for plugin %s\\.\\.\\.", repoName, pluginName))
								Expect(testUI.Out).To(Say("Plugin %s %s found in: %s", pluginName, downloadedVersionString, repoName))
								Expect(testUI.Out).To(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
								Expect(testUI.Out).To(Say("Install and use plugins at your own risk\\."))
								Expect(testUI.Out).To(Say("Do you want to install the plugin %s\\? \\[yN\\]", pluginName))
								Expect(testUI.Out).To(Say("Starting download of plugin binary from repository %s\\.\\.\\.", repoName))
								Expect(testUI.Out).To(Say("Installing plugin %s\\.\\.\\.", pluginName))
								Expect(testUI.Out).To(Say("OK"))
								Expect(testUI.Out).To(Say("%s 1\\.2\\.3 successfully installed", pluginName))
							})
						})
					})
				})
			})
		})
	})

	Describe("installing from any registered repo", func() {
		var (
			pluginName string
			pluginURL  string
			repoName   string
			repoURL    string
			repo2Name  string
			repo2URL   string
			repo3Name  string
			repo3URL   string
			platform   string
		)

		BeforeEach(func() {
			pluginName = helpers.PrefixedRandomName("plugin")
			pluginURL = helpers.PrefixedRandomName("http://")
			repoName = helpers.PrefixedRandomName("repoA")
			repoURL = helpers.PrefixedRandomName("http://")
			repo2Name = helpers.PrefixedRandomName("repoB")
			repo2URL = helpers.PrefixedRandomName("http://")
			repo3Name = helpers.PrefixedRandomName("repoC")
			repo3URL = helpers.PrefixedRandomName("http://")
			cmd.OptionalArgs.PluginNameOrLocation = flag.Path(pluginName)

			platform = helpers.PrefixedRandomName("platform")
			fakeActor.GetPlatformStringReturns(platform)
		})

		Context("when there are no registered repos", func() {
			BeforeEach(func() {
				fakeConfig.PluginRepositoriesReturns([]configv3.PluginRepository{})
			})

			It("returns PluginNotFoundOnDiskOrInAnyRepositoryError", func() {
				Expect(executeErr).To(MatchError(translatableerror.PluginNotFoundOnDiskOrInAnyRepositoryError{PluginName: pluginName, BinaryName: binaryName}))
			})
		})

		Context("when there is one registered repo", func() {
			BeforeEach(func() {
				fakeConfig.PluginRepositoriesReturns([]configv3.PluginRepository{{Name: repoName, URL: repoURL}})
			})

			Context("when there is an error getting the plugin", func() {
				BeforeEach(func() {
					fakeActor.GetPluginInfoFromRepositoriesForPlatformReturns(pluginaction.PluginInfo{}, nil, errors.New("some-error"))
				})

				It("returns the error", func() {
					Expect(executeErr).To(MatchError(errors.New("some-error")))
				})
			})

			Context("when the plugin is not found", func() {
				BeforeEach(func() {
					fakeActor.GetPluginInfoFromRepositoriesForPlatformReturns(pluginaction.PluginInfo{}, nil, actionerror.PluginNotFoundInAnyRepositoryError{PluginName: pluginName})
				})

				It("returns the plugin not found error", func() {
					Expect(executeErr).To(MatchError(translatableerror.PluginNotFoundOnDiskOrInAnyRepositoryError{PluginName: pluginName, BinaryName: binaryName}))
				})
			})

			Context("when the plugin is found", func() {
				var (
					checksum                string
					downloadedVersionString string
					execPath                string
				)

				BeforeEach(func() {
					fakeActor.GetPluginInfoFromRepositoriesForPlatformReturns(pluginaction.PluginInfo{Name: pluginName, Version: downloadedVersionString, URL: pluginURL, Checksum: checksum}, []string{repoName}, nil)

					plugin := configv3.Plugin{
						Name:    pluginName,
						Version: configv3.PluginVersion{Major: 1, Minor: 2, Build: 2},
					}
					fakeConfig.GetPluginReturns(plugin, true)
					fakeConfig.GetPluginCaseInsensitiveReturns(plugin, true)

					execPath = helpers.PrefixedRandomName("some-path")
					fakeActor.DownloadExecutableBinaryFromURLReturns(execPath, nil)
				})

				Context("when the -f flag is provided, the plugin has already been installed, getting the binary succeeds, validating the checksum succeeds, creating an executable copy succeeds, validating the new plugin succeeds, uninstalling the existing plugin succeeds, and installing the plugin is succeeds", func() {
					var (
						pluginVersion      configv3.PluginVersion
						pluginVersionRegex string
					)

					BeforeEach(func() {
						cmd.Force = true

						fakeActor.ValidateFileChecksumReturns(true)
						checksum = helpers.PrefixedRandomName("checksum")

						fakeActor.CreateExecutableCopyReturns("copy-path", nil)

						major := rand.Int()
						minor := rand.Int()
						build := rand.Int()
						pluginVersion = configv3.PluginVersion{Major: major, Minor: minor, Build: build}
						pluginVersionRegex = fmt.Sprintf("%d\\.%d\\.%d", major, minor, build)

						fakeActor.GetAndValidatePluginReturns(configv3.Plugin{
							Name:    pluginName,
							Version: pluginVersion,
						}, nil)
					})

					It("uninstalls the existing plugin and installs the new one", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say("Searching %s for plugin %s\\.\\.\\.", repoName, pluginName))
						Expect(testUI.Out).To(Say("Plugin %s %s found in: %s", pluginName, downloadedVersionString, repoName))
						Expect(testUI.Out).To(Say("Plugin %s 1\\.2\\.2 is already installed\\.", pluginName))
						Expect(testUI.Out).To(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
						Expect(testUI.Out).To(Say("Install and use plugins at your own risk\\."))
						Expect(testUI.Out).To(Say("Starting download of plugin binary from repository %s\\.\\.\\.", repoName))
						Expect(testUI.Out).To(Say("Uninstalling existing plugin\\.\\.\\."))
						Expect(testUI.Out).To(Say("OK"))
						Expect(testUI.Out).To(Say("Plugin %s successfully uninstalled\\.", pluginName))
						Expect(testUI.Out).To(Say("Installing plugin %s\\.\\.\\.", pluginName))
						Expect(testUI.Out).To(Say("OK"))
						Expect(testUI.Out).To(Say("%s %s successfully installed\\.", pluginName, pluginVersionRegex))
					})
				})

				Context("when the -f flag is not provided, the plugin has already been installed, getting the binary succeeds, but validating the checksum fails", func() {
					BeforeEach(func() {
						fakeActor.DownloadExecutableBinaryFromURLReturns("some-path", nil)
					})

					Context("when the checksum fails", func() {
						BeforeEach(func() {
							cmd.Force = false
							fakeActor.ValidateFileChecksumReturns(false)
							input.Write([]byte("y\n"))
						})

						It("returns the checksum error", func() {
							Expect(executeErr).To(MatchError(translatableerror.InvalidChecksumError{}))

							Expect(testUI.Out).To(Say("Searching %s for plugin %s\\.\\.\\.", repoName, pluginName))
							Expect(testUI.Out).To(Say("Plugin %s %s found in: %s", pluginName, downloadedVersionString, repoName))
							Expect(testUI.Out).To(Say("Plugin %s 1\\.2\\.2 is already installed\\.", pluginName))
							Expect(testUI.Out).To(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
							Expect(testUI.Out).To(Say("Install and use plugins at your own risk\\."))
							Expect(testUI.Out).To(Say("Do you want to uninstall the existing plugin and install %s %s\\? \\[yN\\]", pluginName, downloadedVersionString))
							Expect(testUI.Out).To(Say("Starting download of plugin binary from repository %s\\.\\.\\.", repoName))
						})
					})
				})
			})
		})

		Context("when there are many registered repos", func() {
			BeforeEach(func() {
				fakeConfig.PluginRepositoriesReturns([]configv3.PluginRepository{{Name: repoName, URL: repoURL}, {Name: repo2Name, URL: repo2URL}, {Name: repo3Name, URL: repo3URL}})
			})

			Context("when getting the repository information errors", func() {
				DescribeTable("properly propagates errors",
					func(clientErr error, expectedErr error) {
						fakeActor.GetPluginInfoFromRepositoriesForPlatformReturns(
							pluginaction.PluginInfo{},
							nil,
							clientErr)

						executeErr = cmd.Execute(nil)

						Expect(executeErr).To(MatchError(expectedErr))
					},

					Entry("when the error is a RawHTTPStatusError",
						actionerror.FetchingPluginInfoFromRepositoryError{
							RepositoryName: "some-repo",
							Err:            pluginerror.RawHTTPStatusError{Status: "some-status"},
						},
						translatableerror.FetchingPluginInfoFromRepositoriesError{Message: "some-status", RepositoryName: "some-repo"},
					),

					Entry("when the error is a SSLValidationHostnameError",
						actionerror.FetchingPluginInfoFromRepositoryError{
							RepositoryName: "some-repo",
							Err:            pluginerror.SSLValidationHostnameError{Message: "some-status"},
						},

						translatableerror.FetchingPluginInfoFromRepositoriesError{Message: "Hostname does not match SSL Certificate (some-status)", RepositoryName: "some-repo"},
					),

					Entry("when the error is an UnverifiedServerError",
						actionerror.FetchingPluginInfoFromRepositoryError{
							RepositoryName: "some-repo",
							Err:            pluginerror.UnverifiedServerError{URL: "some-url"},
						},
						translatableerror.FetchingPluginInfoFromRepositoriesError{Message: "x509: certificate signed by unknown authority", RepositoryName: "some-repo"},
					),

					Entry("when the error is generic",
						actionerror.FetchingPluginInfoFromRepositoryError{
							RepositoryName: "some-repo",
							Err:            errors.New("generic-error"),
						},
						errors.New("generic-error"),
					),
				)
			})

			Context("when the plugin can't be found in any repos", func() {
				BeforeEach(func() {
					fakeActor.GetPluginInfoFromRepositoriesForPlatformReturns(pluginaction.PluginInfo{}, nil, actionerror.PluginNotFoundInAnyRepositoryError{PluginName: pluginName})
				})

				It("returns PluginNotFoundOnDiskOrInAnyRepositoryError", func() {
					Expect(executeErr).To(MatchError(translatableerror.PluginNotFoundOnDiskOrInAnyRepositoryError{PluginName: pluginName, BinaryName: binaryName}))
				})
			})

			Context("when the plugin is found in one repo", func() {
				var (
					checksum                string
					downloadedVersionString string
					execPath                string
				)

				BeforeEach(func() {
					fakeActor.GetPluginInfoFromRepositoriesForPlatformReturns(pluginaction.PluginInfo{Name: pluginName, Version: downloadedVersionString, URL: pluginURL, Checksum: checksum}, []string{repo2Name}, nil)

					plugin := configv3.Plugin{
						Name:    pluginName,
						Version: configv3.PluginVersion{Major: 1, Minor: 2, Build: 2},
					}
					fakeConfig.GetPluginReturns(plugin, true)
					fakeConfig.GetPluginCaseInsensitiveReturns(plugin, true)

					execPath = helpers.PrefixedRandomName("some-path")
				})

				Context("when the -f flag is provided, the plugin has already been installed, getting the binary succeeds, validating the checksum succeeds, creating an executable copy succeeds, validating the new plugin succeeds, uninstalling the existing plugin succeeds, and installing the plugin is succeeds", func() {
					var pluginVersion configv3.PluginVersion

					BeforeEach(func() {
						cmd.Force = true

						fakeActor.DownloadExecutableBinaryFromURLReturns(execPath, nil)

						fakeActor.ValidateFileChecksumReturns(true)
						checksum = helpers.PrefixedRandomName("checksum")

						fakeActor.CreateExecutableCopyReturns("copy-path", nil)

						major := rand.Int()
						minor := rand.Int()
						build := rand.Int()

						pluginVersion = configv3.PluginVersion{Major: major, Minor: minor, Build: build}

						fakeActor.GetAndValidatePluginReturns(configv3.Plugin{
							Name:    pluginName,
							Version: pluginVersion,
						}, nil)
					})

					It("uninstalls the existing plugin and installs the new one", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say("Searching %s, %s, %s for plugin %s\\.\\.\\.", repoName, repo2Name, repo3Name, pluginName))
						Expect(testUI.Out).To(Say("Plugin %s %s found in: %s", pluginName, downloadedVersionString, repo2Name))
					})
				})

				Context("when the -f flag is not provided, the plugin has already been installed, getting the binary succeeds fails", func() {

					BeforeEach(func() {
						cmd.Force = false
						fakeActor.DownloadExecutableBinaryFromURLReturns("", errors.New("some-error"))
						input.Write([]byte("y\n"))
					})

					It("returns the checksum error", func() {
						Expect(executeErr).To(MatchError(errors.New("some-error")))

						Expect(testUI.Out).To(Say("Searching %s, %s, %s for plugin %s\\.\\.\\.", repoName, repo2Name, repo3Name, pluginName))
						Expect(testUI.Out).To(Say("Plugin %s %s found in: %s", pluginName, downloadedVersionString, repo2Name))
						Expect(testUI.Out).To(Say("Plugin %s 1\\.2\\.2 is already installed\\.", pluginName))
						Expect(testUI.Out).To(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
						Expect(testUI.Out).To(Say("Install and use plugins at your own risk\\."))
						Expect(testUI.Out).To(Say("Do you want to uninstall the existing plugin and install %s %s\\? \\[yN\\]", pluginName, downloadedVersionString))
						Expect(testUI.Out).To(Say("Starting download of plugin binary from repository %s\\.\\.\\.", repo2Name))
					})
				})
			})

			Context("when the plugin is found in multiple repos", func() {
				var (
					checksum                string
					downloadedVersionString string
					execPath                string
				)

				BeforeEach(func() {
					downloadedVersionString = helpers.PrefixedRandomName("version")

					fakeActor.GetPluginInfoFromRepositoriesForPlatformReturns(pluginaction.PluginInfo{Name: pluginName, Version: downloadedVersionString, URL: pluginURL, Checksum: checksum}, []string{repo2Name, repo3Name}, nil)

					plugin := configv3.Plugin{
						Name:    pluginName,
						Version: configv3.PluginVersion{Major: 1, Minor: 2, Build: 2},
					}
					fakeConfig.GetPluginReturns(plugin, true)
					fakeConfig.GetPluginCaseInsensitiveReturns(plugin, true)

					execPath = helpers.PrefixedRandomName("some-path")
					fakeActor.DownloadExecutableBinaryFromURLReturns(execPath, nil)
				})

				Context("when the -f flag is provided, the plugin has already been installed, getting the binary succeeds, validating the checksum succeeds, creating an executable copy succeeds, validating the new plugin succeeds, uninstalling the existing plugin succeeds, and installing the plugin is succeeds", func() {
					var (
						pluginVersion      configv3.PluginVersion
						pluginVersionRegex string
					)

					BeforeEach(func() {
						cmd.Force = true

						fakeActor.ValidateFileChecksumReturns(true)
						checksum = helpers.PrefixedRandomName("checksum")

						fakeActor.CreateExecutableCopyReturns("copy-path", nil)

						major := rand.Int()
						minor := rand.Int()
						build := rand.Int()
						pluginVersion = configv3.PluginVersion{Major: major, Minor: minor, Build: build}
						pluginVersionRegex = fmt.Sprintf("%d\\.%d\\.%d", major, minor, build)

						fakeActor.GetAndValidatePluginReturns(configv3.Plugin{
							Name:    pluginName,
							Version: pluginVersion,
						}, nil)
					})

					It("uninstalls the existing plugin and installs the new one", func() {
						Expect(executeErr).ToNot(HaveOccurred())

						Expect(testUI.Out).To(Say("Searching %s, %s, %s for plugin %s\\.\\.\\.", repoName, repo2Name, repo3Name, pluginName))
						Expect(testUI.Out).To(Say("Plugin %s %s found in: %s, %s", pluginName, downloadedVersionString, repo2Name, repo3Name))
						Expect(testUI.Out).To(Say("Plugin %s 1\\.2\\.2 is already installed\\.", pluginName))
						Expect(testUI.Out).To(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
						Expect(testUI.Out).To(Say("Install and use plugins at your own risk\\."))
						Expect(testUI.Out).To(Say("Starting download of plugin binary from repository %s\\.\\.\\.", repo2Name))
						Expect(testUI.Out).To(Say("Uninstalling existing plugin\\.\\.\\."))
						Expect(testUI.Out).To(Say("OK"))
						Expect(testUI.Out).To(Say("Plugin %s successfully uninstalled\\.", pluginName))
						Expect(testUI.Out).To(Say("Installing plugin %s\\.\\.\\.", pluginName))
						Expect(testUI.Out).To(Say("OK"))
						Expect(testUI.Out).To(Say("%s %s successfully installed\\.", pluginName, pluginVersionRegex))
					})
				})

				Context("when the -f flag is not provided, the plugin has already been installed, getting the binary succeeds, validating the checksum succeeds, but creating an executable copy fails", func() {

					BeforeEach(func() {
						cmd.Force = false
						fakeActor.ValidateFileChecksumReturns(true)
						checksum = helpers.PrefixedRandomName("checksum")

						fakeActor.CreateExecutableCopyReturns("", errors.New("some-error"))
						input.Write([]byte("y\n"))
					})

					It("returns the error", func() {
						Expect(executeErr).To(MatchError(errors.New("some-error")))

						Expect(testUI.Out).To(Say("Searching %s, %s, %s for plugin %s\\.\\.\\.", repoName, repo2Name, repo3Name, pluginName))
						Expect(testUI.Out).To(Say("Plugin %s %s found in: %s, %s", pluginName, downloadedVersionString, repo2Name, repo3Name))
						Expect(testUI.Out).To(Say("Plugin %s 1\\.2\\.2 is already installed\\.", pluginName))
						Expect(testUI.Out).To(Say("Attention: Plugins are binaries written by potentially untrusted authors\\."))
						Expect(testUI.Out).To(Say("Install and use plugins at your own risk\\."))
						Expect(testUI.Out).To(Say("Do you want to uninstall the existing plugin and install %s %s\\? \\[yN\\]", pluginName, downloadedVersionString))
						Expect(testUI.Out).To(Say("Starting download of plugin binary from repository %s\\.\\.\\.", repo2Name))
					})
				})
			})
		})
	})
})
