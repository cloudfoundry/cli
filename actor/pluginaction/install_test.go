package pluginaction_test

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cli/actor/actionerror"
	. "code.cloudfoundry.org/cli/actor/pluginaction"
	"code.cloudfoundry.org/cli/actor/pluginaction/pluginactionfakes"
	"code.cloudfoundry.org/cli/api/plugin"
	"code.cloudfoundry.org/cli/api/plugin/pluginfakes"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/generic"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("install actions", func() {
	var (
		actor         *Actor
		fakeConfig    *pluginactionfakes.FakeConfig
		fakeClient    *pluginactionfakes.FakePluginClient
		tempPluginDir string
	)

	BeforeEach(func() {
		fakeConfig = new(pluginactionfakes.FakeConfig)
		fakeClient = new(pluginactionfakes.FakePluginClient)
		actor = NewActor(fakeConfig, fakeClient)

		var err error
		tempPluginDir, err = ioutil.TempDir("", "")
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		err := os.RemoveAll(tempPluginDir)
		Expect(err).ToNot(HaveOccurred())
	})

	Describe("CreateExecutableCopy", func() {
		Context("when the file exists", func() {
			var pluginPath string

			BeforeEach(func() {
				tempFile, err := ioutil.TempFile("", "")
				Expect(err).ToNot(HaveOccurred())

				_, err = tempFile.WriteString("cthulhu")
				Expect(err).ToNot(HaveOccurred())
				err = tempFile.Close()
				Expect(err).ToNot(HaveOccurred())

				pluginPath = tempFile.Name()
			})

			AfterEach(func() {
				err := os.Remove(pluginPath)
				Expect(err).ToNot(HaveOccurred())
			})

			It("creates a copy of a file in plugin home", func() {
				copyPath, err := actor.CreateExecutableCopy(pluginPath, tempPluginDir)
				Expect(err).ToNot(HaveOccurred())

				contents, err := ioutil.ReadFile(copyPath)
				Expect(err).ToNot(HaveOccurred())
				Expect(contents).To(BeEquivalentTo("cthulhu"))
			})
		})

		Context("when the file does not exist", func() {
			It("returns an os.PathError", func() {
				_, err := actor.CreateExecutableCopy("i-don't-exist", tempPluginDir)
				_, isPathError := err.(*os.PathError)
				Expect(isPathError).To(BeTrue())
			})
		})
	})

	Describe("DownloadExecutableBinaryFromURL", func() {
		var (
			path            string
			downloadErr     error
			fakeProxyReader *pluginfakes.FakeProxyReader
		)

		JustBeforeEach(func() {
			fakeProxyReader = new(pluginfakes.FakeProxyReader)
			path, downloadErr = actor.DownloadExecutableBinaryFromURL("some-plugin-url.com", tempPluginDir, fakeProxyReader)
		})

		Context("when the downloaded is successful", func() {
			var (
				data []byte
			)

			BeforeEach(func() {
				data = []byte("some test data")
				fakeClient.DownloadPluginStub = func(_ string, path string, _ plugin.ProxyReader) error {
					err := ioutil.WriteFile(path, data, 0700)
					Expect(err).ToNot(HaveOccurred())
					return nil
				}
			})
			It("returns the path to the file and the size", func() {
				Expect(downloadErr).ToNot(HaveOccurred())
				fileData, err := ioutil.ReadFile(path)
				Expect(err).ToNot(HaveOccurred())
				Expect(fileData).To(Equal(data))

				Expect(fakeClient.DownloadPluginCallCount()).To(Equal(1))
				pluginURL, downloadPath, proxyReader := fakeClient.DownloadPluginArgsForCall(0)
				Expect(pluginURL).To(Equal("some-plugin-url.com"))
				Expect(downloadPath).To(Equal(path))
				Expect(proxyReader).To(Equal(fakeProxyReader))
			})
		})

		Context("when there is an error downloading file", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("some error")
				fakeClient.DownloadPluginReturns(expectedErr)
			})

			It("returns the error", func() {
				Expect(downloadErr).To(MatchError(expectedErr))
			})
		})
	})

	Describe("FileExists", func() {
		var pluginPath string

		Context("when the file exists", func() {
			BeforeEach(func() {
				pluginFile, err := ioutil.TempFile("", "")
				Expect(err).NotTo(HaveOccurred())
				err = pluginFile.Close()
				Expect(err).NotTo(HaveOccurred())

				pluginPath = pluginFile.Name()
			})

			AfterEach(func() {
				err := os.Remove(pluginPath)
				Expect(err).NotTo(HaveOccurred())
			})

			It("returns true", func() {
				Expect(actor.FileExists(pluginPath)).To(BeTrue())
			})
		})

		Context("when the file does not exist", func() {
			It("returns false", func() {
				Expect(actor.FileExists("/some/path/that/does/not/exist")).To(BeFalse())
			})
		})
	})

	Describe("GetAndValidatePlugin", func() {
		var (
			fakePluginMetadata *pluginactionfakes.FakePluginMetadata
			fakeCommandList    *pluginactionfakes.FakeCommandList
			plugin             configv3.Plugin
			validateErr        error
		)

		BeforeEach(func() {
			fakePluginMetadata = new(pluginactionfakes.FakePluginMetadata)
			fakeCommandList = new(pluginactionfakes.FakeCommandList)
		})

		JustBeforeEach(func() {
			plugin, validateErr = actor.GetAndValidatePlugin(fakePluginMetadata, fakeCommandList, "some-plugin-path")
		})

		Context("when getting the plugin metadata returns an error", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("error getting metadata")
				fakePluginMetadata.GetMetadataReturns(configv3.Plugin{}, expectedErr)
			})

			It("returns a PluginInvalidError", func() {
				Expect(validateErr).To(MatchError(actionerror.PluginInvalidError{Err: expectedErr}))
			})
		})

		Context("when the plugin name is missing", func() {
			BeforeEach(func() {
				fakePluginMetadata.GetMetadataReturns(configv3.Plugin{}, nil)
			})

			It("returns a PluginInvalidError", func() {
				Expect(validateErr).To(MatchError(actionerror.PluginInvalidError{}))
			})
		})

		Context("when the plugin does not have any commands", func() {
			BeforeEach(func() {
				fakePluginMetadata.GetMetadataReturns(configv3.Plugin{Name: "some-plugin"}, nil)
			})

			It("returns a PluginInvalidError", func() {
				Expect(validateErr).To(MatchError(actionerror.PluginInvalidError{}))
			})
		})

		Context("when there are command conflicts", func() {
			BeforeEach(func() {
				fakePluginMetadata.GetMetadataReturns(configv3.Plugin{
					Name: "some-plugin",
					Version: configv3.PluginVersion{
						Major: 1,
						Minor: 1,
						Build: 1,
					},
					Commands: []configv3.PluginCommand{
						{Name: "some-other-command", Alias: "soc"},
						{Name: "some-command", Alias: "sc"},
						{Name: "version", Alias: "v"},
						{Name: "p", Alias: "push"},
					},
				}, nil)
			})

			Context("when the plugin has command names that conflict with native command names", func() {
				BeforeEach(func() {
					fakeCommandList.HasCommandStub = func(commandName string) bool {
						switch commandName {
						case "version":
							return true
						default:
							return false
						}
					}
				})

				It("returns a PluginCommandsConflictError containing all conflicting command names", func() {
					Expect(validateErr).To(MatchError(actionerror.PluginCommandsConflictError{
						PluginName:     "some-plugin",
						PluginVersion:  "1.1.1",
						CommandNames:   []string{"version"},
						CommandAliases: []string{},
					}))
				})
			})

			Context("when the plugin has command names that conflict with native command aliases", func() {
				BeforeEach(func() {
					fakeCommandList.HasAliasStub = func(commandAlias string) bool {
						switch commandAlias {
						case "p":
							return true
						default:
							return false
						}
					}
				})

				It("returns a PluginCommandsConflictError containing all conflicting command names", func() {
					Expect(validateErr).To(MatchError(actionerror.PluginCommandsConflictError{
						PluginName:     "some-plugin",
						PluginVersion:  "1.1.1",
						CommandNames:   []string{"p"},
						CommandAliases: []string{},
					}))
				})
			})

			Context("when the plugin has command aliases that conflict with native command names", func() {
				BeforeEach(func() {
					fakeCommandList.HasCommandStub = func(commandName string) bool {
						switch commandName {
						case "push":
							return true
						default:
							return false
						}
					}
				})

				It("returns a PluginCommandsConflictError containing all conflicting command aliases", func() {
					Expect(validateErr).To(MatchError(actionerror.PluginCommandsConflictError{
						PluginName:     "some-plugin",
						PluginVersion:  "1.1.1",
						CommandAliases: []string{"push"},
						CommandNames:   []string{},
					}))
				})
			})

			Context("when the plugin has command aliases that conflict with native command aliases", func() {
				BeforeEach(func() {
					fakeCommandList.HasAliasStub = func(commandAlias string) bool {
						switch commandAlias {
						case "v":
							return true
						default:
							return false
						}
					}
				})

				It("returns a PluginCommandsConflictError containing all conflicting command aliases", func() {
					Expect(validateErr).To(MatchError(actionerror.PluginCommandsConflictError{
						PluginName:     "some-plugin",
						PluginVersion:  "1.1.1",
						CommandAliases: []string{"v"},
						CommandNames:   []string{},
					}))
				})
			})

			Context("when the plugin has command names that conflict with existing plugin command names", func() {
				BeforeEach(func() {
					fakeConfig.PluginsReturns([]configv3.Plugin{{
						Name:     "installed-plugin-2",
						Commands: []configv3.PluginCommand{{Name: "some-command"}},
					}})
				})

				It("returns a PluginCommandsConflictError containing all conflicting command names", func() {
					Expect(validateErr).To(MatchError(actionerror.PluginCommandsConflictError{
						PluginName:     "some-plugin",
						PluginVersion:  "1.1.1",
						CommandNames:   []string{"some-command"},
						CommandAliases: []string{},
					}))
				})
			})

			Context("when the plugin has command names that conflict with existing plugin command aliases", func() {
				BeforeEach(func() {
					fakeConfig.PluginsReturns([]configv3.Plugin{{
						Name:     "installed-plugin-2",
						Commands: []configv3.PluginCommand{{Alias: "some-command"}}},
					})
				})

				It("returns a PluginCommandsConflictError containing all conflicting command names", func() {
					Expect(validateErr).To(MatchError(actionerror.PluginCommandsConflictError{
						PluginName:     "some-plugin",
						PluginVersion:  "1.1.1",
						CommandNames:   []string{"some-command"},
						CommandAliases: []string{},
					}))
				})
			})

			Context("when the plugin has command aliases that conflict with existing plugin command names", func() {
				BeforeEach(func() {
					fakeConfig.PluginsReturns([]configv3.Plugin{{
						Name:     "installed-plugin-2",
						Commands: []configv3.PluginCommand{{Name: "sc"}}},
					})
				})

				It("returns a PluginCommandsConflictError containing all conflicting command aliases", func() {
					Expect(validateErr).To(MatchError(actionerror.PluginCommandsConflictError{
						PluginName:     "some-plugin",
						PluginVersion:  "1.1.1",
						CommandNames:   []string{},
						CommandAliases: []string{"sc"},
					}))
				})
			})

			Context("when the plugin has command aliases that conflict with existing plugin command aliases", func() {
				BeforeEach(func() {
					fakeConfig.PluginsReturns([]configv3.Plugin{{
						Name:     "installed-plugin-2",
						Commands: []configv3.PluginCommand{{Alias: "sc"}}},
					})
				})

				It("returns a PluginCommandsConflictError containing all conflicting command aliases", func() {
					Expect(validateErr).To(MatchError(actionerror.PluginCommandsConflictError{
						PluginName:     "some-plugin",
						PluginVersion:  "1.1.1",
						CommandAliases: []string{"sc"},
						CommandNames:   []string{},
					}))
				})
			})

			Context("when the plugin has command names and aliases that conflict with existing native and plugin command names and aliases", func() {
				BeforeEach(func() {
					fakeConfig.PluginsReturns([]configv3.Plugin{
						{
							Name: "installed-plugin-1",
							Commands: []configv3.PluginCommand{
								{Name: "some-command"},
								{Alias: "some-other-command"},
							},
						},
						{
							Name: "installed-plugin-2",
							Commands: []configv3.PluginCommand{
								{Name: "sc"},
								{Alias: "soc"},
							},
						},
					})

					fakeCommandList.HasCommandStub = func(commandName string) bool {
						switch commandName {
						case "version", "p":
							return true
						default:
							return false
						}
					}

					fakeCommandList.HasAliasStub = func(commandAlias string) bool {
						switch commandAlias {
						case "v", "push":
							return true
						default:
							return false
						}
					}
				})

				It("returns a PluginCommandsConflictError with all conflicting command names and aliases", func() {
					Expect(validateErr).To(MatchError(actionerror.PluginCommandsConflictError{
						PluginName:     "some-plugin",
						PluginVersion:  "1.1.1",
						CommandNames:   []string{"p", "some-command", "some-other-command", "version"},
						CommandAliases: []string{"push", "sc", "soc", "v"},
					}))
				})
			})

			Context("when the plugin is already installed", func() {
				BeforeEach(func() {
					fakeConfig.PluginsReturns([]configv3.Plugin{{
						Name: "some-plugin",
						Commands: []configv3.PluginCommand{
							{Name: "some-command", Alias: "sc"},
							{Name: "some-other-command", Alias: "soc"},
						},
					}})
				})

				It("does not return any errors due to command name or alias conflict", func() {
					Expect(validateErr).ToNot(HaveOccurred())
				})
			})
		})

		Context("when the plugin is valid", func() {
			var pluginToBeInstalled configv3.Plugin

			BeforeEach(func() {
				pluginToBeInstalled = configv3.Plugin{
					Name: "some-plugin",
					Version: configv3.PluginVersion{
						Major: 1,
						Minor: 1,
						Build: 1,
					},
					Commands: []configv3.PluginCommand{
						{
							Name:  "some-command",
							Alias: "sc",
						},
						{
							Name:  "some-other-command",
							Alias: "soc",
						},
					},
				}
				fakePluginMetadata.GetMetadataReturns(pluginToBeInstalled, nil)
				fakeConfig.PluginsReturns([]configv3.Plugin{
					{
						Name: "installed-plugin-1",
						Commands: []configv3.PluginCommand{
							{
								Name:  "unique-command-1",
								Alias: "uc1",
							},
						},
					},
					{
						Name: "installed-plugin-2",
						Commands: []configv3.PluginCommand{
							{
								Name:  "unique-command-2",
								Alias: "uc2",
							},
							{
								Name:  "unique-command-3",
								Alias: "uc3",
							},
						},
					},
				})
			})

			It("returns the plugin and no errors", func() {
				Expect(validateErr).ToNot(HaveOccurred())
				Expect(plugin).To(Equal(pluginToBeInstalled))

				Expect(fakePluginMetadata.GetMetadataCallCount()).To(Equal(1))
				Expect(fakePluginMetadata.GetMetadataArgsForCall(0)).To(Equal("some-plugin-path"))

				Expect(fakeCommandList.HasCommandCallCount()).To(Equal(4))
				Expect(fakeCommandList.HasCommandArgsForCall(0)).To(Equal("some-command"))
				Expect(fakeCommandList.HasCommandArgsForCall(1)).To(Equal("sc"))
				Expect(fakeCommandList.HasCommandArgsForCall(2)).To(Equal("some-other-command"))
				Expect(fakeCommandList.HasCommandArgsForCall(3)).To(Equal("soc"))

				Expect(fakeCommandList.HasAliasCallCount()).To(Equal(4))
				Expect(fakeCommandList.HasAliasArgsForCall(0)).To(Equal("some-command"))
				Expect(fakeCommandList.HasAliasArgsForCall(1)).To(Equal("sc"))
				Expect(fakeCommandList.HasAliasArgsForCall(2)).To(Equal("some-other-command"))
				Expect(fakeCommandList.HasAliasArgsForCall(3)).To(Equal("soc"))

				Expect(fakeConfig.PluginsCallCount()).To(Equal(1))
			})
		})
	})

	Describe("InstallPluginFromLocalPath", func() {
		var (
			plugin     configv3.Plugin
			installErr error

			pluginHomeDir string
			pluginPath    string
			tempDir       string
		)

		BeforeEach(func() {
			plugin = configv3.Plugin{
				Name: "some-plugin",
				Commands: []configv3.PluginCommand{
					{Name: "some-command"},
				},
			}

			pluginFile, err := ioutil.TempFile("", "")
			Expect(err).NotTo(HaveOccurred())
			err = pluginFile.Close()
			Expect(err).NotTo(HaveOccurred())

			pluginPath = pluginFile.Name()

			tempDir, err = ioutil.TempDir("", "")
			Expect(err).ToNot(HaveOccurred())

			pluginHomeDir = filepath.Join(tempDir, ".cf", "plugin")
		})

		AfterEach(func() {
			err := os.Remove(pluginPath)
			Expect(err).NotTo(HaveOccurred())

			err = os.RemoveAll(tempDir)
			Expect(err).NotTo(HaveOccurred())
		})

		JustBeforeEach(func() {
			installErr = actor.InstallPluginFromPath(pluginPath, plugin)
		})

		Context("when an error is encountered copying the plugin to the plugin directory", func() {
			BeforeEach(func() {
				fakeConfig.PluginHomeReturns(pluginPath)
			})

			It("returns the error", func() {
				_, isPathError := installErr.(*os.PathError)
				Expect(isPathError).To(BeTrue())
			})
		})

		Context("when an error is encountered writing the plugin config to disk", func() {
			var (
				expectedErr error
			)

			BeforeEach(func() {
				fakeConfig.PluginHomeReturns(pluginHomeDir)

				expectedErr = errors.New("write config error")
				fakeConfig.WritePluginConfigReturns(expectedErr)
			})

			It("returns the error", func() {
				Expect(installErr).To(MatchError(expectedErr))
			})
		})

		Context("when no errors are encountered", func() {
			BeforeEach(func() {
				fakeConfig.PluginHomeReturns(pluginHomeDir)
			})

			It("makes an executable copy of the plugin file in the plugin directory, updates the plugin config, and writes the config to disk", func() {
				Expect(installErr).ToNot(HaveOccurred())

				installedPluginPath := generic.ExecutableFilename(filepath.Join(pluginHomeDir, "some-plugin"))

				Expect(fakeConfig.PluginHomeCallCount()).To(Equal(1))

				Expect(fakeConfig.AddPluginCallCount()).To(Equal(1))
				Expect(fakeConfig.AddPluginArgsForCall(0)).To(Equal(configv3.Plugin{
					Name: "some-plugin",
					Commands: []configv3.PluginCommand{
						{Name: "some-command"},
					},
					Location: installedPluginPath,
				}))

				Expect(fakeConfig.WritePluginConfigCallCount()).To(Equal(1))
			})
		})
	})
})
