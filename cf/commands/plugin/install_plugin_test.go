package plugin_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/rpc"
	"os"
	"path/filepath"
	"runtime"

	"github.com/cloudfoundry/cli/cf/actors/plugin_repo/fakes"
	"github.com/cloudfoundry/cli/cf/command"
	testCommand "github.com/cloudfoundry/cli/cf/command/fakes"
	"github.com/cloudfoundry/cli/cf/command_metadata"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/configuration/plugin_config"
	testPluginConfig "github.com/cloudfoundry/cli/cf/configuration/plugin_config/fakes"
	"github.com/cloudfoundry/cli/cf/models"
	"github.com/cloudfoundry/cli/plugin"
	cliRpc "github.com/cloudfoundry/cli/plugin/rpc"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	testChecksum "github.com/cloudfoundry/cli/utils/fakes"

	clipr "github.com/cloudfoundry-incubator/cli-plugin-repo/models"

	. "github.com/cloudfoundry/cli/cf/commands/plugin"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Install", func() {
	var (
		ui                  *testterm.FakeUI
		requirementsFactory *testreq.FakeReqFactory
		config              core_config.ReadWriter
		pluginConfig        *testPluginConfig.FakePluginConfiguration
		fakePluginRepo      *fakes.FakePluginRepo
		fakeChecksum        *testChecksum.FakeSha1Checksum

		coreCmds   map[string]command.Command
		pluginFile *os.File
		homeDir    string
		pluginDir  string
		curDir     string

		test_1                    string
		test_2                    string
		test_curDir               string
		test_with_help            string
		test_with_push            string
		test_with_push_short_name string
		aliasConflicts            string
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		requirementsFactory = &testreq.FakeReqFactory{}
		pluginConfig = &testPluginConfig.FakePluginConfiguration{}
		config = testconfig.NewRepositoryWithDefaults()
		fakePluginRepo = &fakes.FakePluginRepo{}
		fakeChecksum = &testChecksum.FakeSha1Checksum{}
		coreCmds = make(map[string]command.Command)

		dir, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		test_1 = filepath.Join(dir, "..", "..", "..", "fixtures", "plugins", "test_1.exe")
		test_2 = filepath.Join(dir, "..", "..", "..", "fixtures", "plugins", "test_2.exe")
		test_curDir = filepath.Join("test_1.exe")
		test_with_help = filepath.Join(dir, "..", "..", "..", "fixtures", "plugins", "test_with_help.exe")
		test_with_push = filepath.Join(dir, "..", "..", "..", "fixtures", "plugins", "test_with_push.exe")
		test_with_push_short_name = filepath.Join(dir, "..", "..", "..", "fixtures", "plugins", "test_with_push_short_name.exe")
		aliasConflicts = filepath.Join(dir, "..", "..", "..", "fixtures", "plugins", "alias_conflicts.exe")

		homeDir, err = ioutil.TempDir(os.TempDir(), "plugins")
		Expect(err).ToNot(HaveOccurred())

		pluginDir = filepath.Join(homeDir, ".cf", "plugins")
		pluginConfig.GetPluginPathReturns(pluginDir)

		curDir, err = os.Getwd()
		Expect(err).ToNot(HaveOccurred())
		pluginFile, err = ioutil.TempFile("./", "test_plugin")
		Expect(err).ToNot(HaveOccurred())

		if runtime.GOOS != "windows" {
			err = os.Chmod(test_1, 0700)
			Expect(err).ToNot(HaveOccurred())
		}
	})

	AfterEach(func() {
		os.Remove(filepath.Join(curDir, pluginFile.Name()))
		os.Remove(homeDir)
	})

	runCommand := func(args ...string) bool {
		//reset rpc registration, each service can only be registered once
		rpc.DefaultServer = rpc.NewServer()
		rpcService, _ := cliRpc.NewRpcService(nil, nil, nil, nil)
		cmd := NewPluginInstall(ui, config, pluginConfig, coreCmds, fakePluginRepo, fakeChecksum, rpcService)
		return testcmd.RunCommand(cmd, args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("fails with usage when not provided a path to the plugin executable", func() {
			Expect(runCommand()).ToNot(HavePassedRequirements())
		})
	})

	Describe("Locating binary file", func() {

		Describe("install from plugin repository when '-r' provided", func() {
			Context("gets metadata of the plugin from repo", func() {
				Context("when repo is not found in config", func() {
					It("informs user repo is not found", func() {
						runCommand("plugin1", "-r", "repo1")
						Ω(ui.Outputs).To(ContainSubstrings([]string{"Looking up 'plugin1' from repository 'repo1'"}))
						Ω(ui.Outputs).To(ContainSubstrings([]string{"repo1 not found"}))
					})
				})

				Context("when repo is found in config", func() {
					Context("when repo endpoint returns an error", func() {
						It("informs user about the error", func() {
							config.SetPluginRepo(models.PluginRepo{Name: "repo1", Url: ""})
							fakePluginRepo.GetPluginsReturns(nil, []string{"repo error1"})
							runCommand("plugin1", "-r", "repo1")

							Ω(ui.Outputs).To(ContainSubstrings([]string{"Error getting plugin metadata from repo"}))
							Ω(ui.Outputs).To(ContainSubstrings([]string{"repo error1"}))
						})
					})

					Context("when plugin metadata is available and desired plugin is not found", func() {
						It("informs user about the error", func() {
							config.SetPluginRepo(models.PluginRepo{Name: "repo1", Url: ""})
							fakePluginRepo.GetPluginsReturns(nil, nil)
							runCommand("plugin1", "-r", "repo1")

							Ω(ui.Outputs).To(ContainSubstrings([]string{"plugin1 is not available in repo 'repo1'"}))
						})
					})

					It("ignore cases in repo name", func() {
						config.SetPluginRepo(models.PluginRepo{Name: "repo1", Url: ""})
						fakePluginRepo.GetPluginsReturns(nil, nil)
						runCommand("plugin1", "-r", "REPO1")

						Ω(ui.Outputs).NotTo(ContainSubstrings([]string{"REPO1 not found"}))
					})
				})
			})

			Context("downloads the binary for the machine's OS", func() {
				Context("when binary is not available", func() {
					It("informs user when binary is not available for OS", func() {
						p := clipr.Plugin{
							Name: "plugin1",
						}
						result := make(map[string][]clipr.Plugin)
						result["repo1"] = []clipr.Plugin{p}

						config.SetPluginRepo(models.PluginRepo{Name: "repo1", Url: ""})
						fakePluginRepo.GetPluginsReturns(result, nil)
						runCommand("plugin1", "-r", "repo1")

						Ω(ui.Outputs).To(ContainSubstrings([]string{"Plugin requested has no binary available"}))
					})
				})

				Context("when binary is available", func() {
					var (
						testServer *httptest.Server
					)

					BeforeEach(func() {
						h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							fmt.Fprintln(w, "abc")
						})

						testServer = httptest.NewServer(h)

						fakeChecksum.CheckSha1Returns(true)

						p := clipr.Plugin{
							Name: "plugin1",
							Binaries: []clipr.Binary{
								clipr.Binary{
									Platform: "osx",
									Url:      testServer.URL + "/test.exe",
								},
								clipr.Binary{
									Platform: "win64",
									Url:      testServer.URL + "/test.exe",
								},
								clipr.Binary{
									Platform: "win32",
									Url:      testServer.URL + "/test.exe",
								},
								clipr.Binary{
									Platform: "linux32",
									Url:      testServer.URL + "/test.exe",
								},
								clipr.Binary{
									Platform: "linux64",
									Url:      testServer.URL + "/test.exe",
								},
							},
						}
						result := make(map[string][]clipr.Plugin)
						result["repo1"] = []clipr.Plugin{p}

						config.SetPluginRepo(models.PluginRepo{Name: "repo1", Url: ""})
						fakePluginRepo.GetPluginsReturns(result, nil)
					})

					AfterEach(func() {
						testServer.Close()
					})

					It("performs sha1 checksum validation on the downloaded binary", func() {
						runCommand("plugin1", "-r", "repo1")
						Ω(fakeChecksum.CheckSha1CallCount()).To(Equal(1))
					})

					It("reports error downloaded file's sha1 does not match the sha1 in metadata", func() {
						fakeChecksum.CheckSha1Returns(false)

						runCommand("plugin1", "-r", "repo1")
						Ω(ui.Outputs).To(ContainSubstrings(
							[]string{"FAILED"},
							[]string{"checksum does not match"},
						))
					})

					It("downloads and installs binary when it is available and checksum matches", func() {
						runCommand("plugin1", "-r", "repo1")

						Ω(ui.Outputs).To(ContainSubstrings([]string{"4 bytes downloaded..."}))
						Ω(ui.Outputs).To(ContainSubstrings([]string{"FAILED"}))
						Ω(ui.Outputs).To(ContainSubstrings([]string{"Installing plugin"}))
					})
				})
			})
		})

		Describe("install from plugin repository with no '-r' provided", func() {
			Context("downloads file from internet if path prefix with 'http','ftp' etc...", func() {
				It("will not try locate file locally", func() {
					runCommand("http://127.0.0.1/plugin.exe")

					Expect(ui.Outputs).ToNot(ContainSubstrings(
						[]string{"File not found locally"},
					))
					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"download binary file from internet address"},
					))
				})

				It("informs users when binary is not downloadable from net", func() {
					runCommand("http://path/to/not/a/thing.exe")

					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Download attempt failed"},
						[]string{"Unable to install"},
						[]string{"FAILED"},
					))
				})

				It("downloads and installs binary when it is available", func() {
					h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						fmt.Fprintln(w, "hi")
					})

					testServer := httptest.NewServer(h)
					defer testServer.Close()

					runCommand(testServer.URL + "/testfile.exe")

					Ω(ui.Outputs).To(ContainSubstrings([]string{"3 bytes downloaded..."}))
					Ω(ui.Outputs).To(ContainSubstrings([]string{"FAILED"}))
					Ω(ui.Outputs).To(ContainSubstrings([]string{"Installing plugin"}))
				})
			})

			Context("tries to locate binary file at local path if path has no internet prefix", func() {
				It("reports error if local file is not found at given path", func() {
					runCommand("./install_plugin.go")

					Expect(ui.Outputs).ToNot(ContainSubstrings(
						[]string{"download binary file from internet"},
					))
					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Installing plugin", "./install_plugin.go"},
					))
				})

				It("installs the plugin from a local file if found", func() {
					runCommand("./no/file/is/here.exe")

					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"File not found locally"},
					))
				})
			})
		})

	})

	Describe("install failures", func() {
		Context("when the plugin contains a 'help' command", func() {
			It("fails", func() {
				runCommand(test_with_help)

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Command `help` in the plugin being installed is a native CF command/alias.  Rename the `help` command in the plugin being installed in order to enable its installation and use."},
					[]string{"FAILED"},
				))
			})
		})

		Context("when the plugin's command conflicts with a core command", func() {
			It("fails if is shares a command name", func() {
				coreCmds["push"] = &testCommand.FakeCommand{}
				runCommand(test_with_push)

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Command `push` in the plugin being installed is a native CF command/alias.  Rename the `push` command in the plugin being installed in order to enable its installation and use."},
					[]string{"FAILED"},
				))
			})

			It("fails if it shares a command short name", func() {
				push := &testCommand.FakeCommand{}
				push.MetadataReturns(command_metadata.CommandMetadata{
					ShortName: "p",
				})

				coreCmds["push"] = push
				runCommand(test_with_push_short_name)

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Command `p` in the plugin being installed is a native CF command/alias.  Rename the `p` command in the plugin being installed in order to enable its installation and use."},
					[]string{"FAILED"},
				))
			})
		})

		Context("when the plugin's alias conflicts with a core command/alias", func() {
			It("fails if is shares a command name", func() {
				coreCmds["conflict-alias"] = &testCommand.FakeCommand{}
				runCommand(aliasConflicts)

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Alias `conflict-alias` in the plugin being installed is a native CF command/alias.  Rename the `conflict-alias` command in the plugin being installed in order to enable its installation and use."},
					[]string{"FAILED"},
				))
			})

			It("fails if it shares a command short name", func() {
				push := &testCommand.FakeCommand{}
				push.MetadataReturns(command_metadata.CommandMetadata{
					ShortName: "conflict-alias",
				})

				coreCmds["push"] = push
				runCommand(aliasConflicts)

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Alias `conflict-alias` in the plugin being installed is a native CF command/alias.  Rename the `conflict-alias` command in the plugin being installed in order to enable its installation and use."},
					[]string{"FAILED"},
				))
			})
		})

		Context("when the plugin's alias conflicts with other installed plugin", func() {
			It("fails if it shares a command name", func() {
				pluginsMap := make(map[string]plugin_config.PluginMetadata)
				pluginsMap["AliasCollision"] = plugin_config.PluginMetadata{
					Location: "location/to/config.exe",
					Commands: []plugin.Command{
						{
							Name:     "conflict-alias",
							HelpText: "Hi!",
						},
					},
				}
				pluginConfig.PluginsReturns(pluginsMap)

				runCommand(aliasConflicts)

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Alias `conflict-alias` is a command/alias in plugin 'AliasCollision'.  You could try uninstalling plugin 'AliasCollision' and then install this plugin in order to invoke the `conflict-alias` command.  However, you should first fully understand the impact of uninstalling the existing 'AliasCollision' plugin."},
					[]string{"FAILED"},
				))
			})

			It("fails if it shares a command alias", func() {
				pluginsMap := make(map[string]plugin_config.PluginMetadata)
				pluginsMap["AliasCollision"] = plugin_config.PluginMetadata{
					Location: "location/to/alias.exe",
					Commands: []plugin.Command{
						{
							Name:     "non-conflict-cmd",
							Alias:    "conflict-alias",
							HelpText: "Hi!",
						},
					},
				}
				pluginConfig.PluginsReturns(pluginsMap)

				runCommand(aliasConflicts)

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Alias `conflict-alias` is a command/alias in plugin 'AliasCollision'.  You could try uninstalling plugin 'AliasCollision' and then install this plugin in order to invoke the `conflict-alias` command.  However, you should first fully understand the impact of uninstalling the existing 'AliasCollision' plugin."},
					[]string{"FAILED"},
				))
			})
		})

		Context("when the plugin's command conflicts with other installed plugin", func() {
			It("fails if it shares a command name", func() {
				pluginsMap := make(map[string]plugin_config.PluginMetadata)
				pluginsMap["Test1Collision"] = plugin_config.PluginMetadata{
					Location: "location/to/config.exe",
					Commands: []plugin.Command{
						{
							Name:     "test_1_cmd1",
							HelpText: "Hi!",
						},
					},
				}
				pluginConfig.PluginsReturns(pluginsMap)

				runCommand(test_1)

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Command `test_1_cmd1` is a command/alias in plugin 'Test1Collision'.  You could try uninstalling plugin 'Test1Collision' and then install this plugin in order to invoke the `test_1_cmd1` command.  However, you should first fully understand the impact of uninstalling the existing 'Test1Collision' plugin."},
					[]string{"FAILED"},
				))
			})

			It("fails if it shares a command alias", func() {
				pluginsMap := make(map[string]plugin_config.PluginMetadata)
				pluginsMap["AliasCollision"] = plugin_config.PluginMetadata{
					Location: "location/to/alias.exe",
					Commands: []plugin.Command{
						{
							Name:     "non-conflict-cmd",
							Alias:    "conflict-cmd",
							HelpText: "Hi!",
						},
					},
				}
				pluginConfig.PluginsReturns(pluginsMap)

				runCommand(aliasConflicts)

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Command `conflict-cmd` is a command/alias in plugin 'AliasCollision'.  You could try uninstalling plugin 'AliasCollision' and then install this plugin in order to invoke the `conflict-cmd` command.  However, you should first fully understand the impact of uninstalling the existing 'AliasCollision' plugin."},
					[]string{"FAILED"},
				))
			})
		})

		It("if plugin name is already taken", func() {
			pluginConfig.PluginsReturns(map[string]plugin_config.PluginMetadata{"Test1": plugin_config.PluginMetadata{}})
			runCommand(test_1)

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Plugin name", "Test1", "is already taken"},
				[]string{"FAILED"},
			))
		})

		Context("io", func() {
			BeforeEach(func() {
				err := os.MkdirAll(pluginDir, 0700)
				Expect(err).NotTo(HaveOccurred())
			})

			It("if a file with the plugin name already exists under ~/.cf/plugin/", func() {
				pluginConfig.PluginsReturns(map[string]plugin_config.PluginMetadata{"useless": plugin_config.PluginMetadata{}})
				pluginConfig.GetPluginPathReturns(curDir)

				runCommand(filepath.Join(curDir, pluginFile.Name()))
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Installing plugin"},
					[]string{"The file", pluginFile.Name(), "already exists"},
					[]string{"FAILED"},
				))
			})
		})
	})

	Describe("install success", func() {
		BeforeEach(func() {
			err := os.MkdirAll(pluginDir, 0700)
			Expect(err).ToNot(HaveOccurred())
			pluginConfig.GetPluginPathReturns(pluginDir)
		})

		It("finds plugin in the current directory without having to specify `./`", func() {
			curDir, err := os.Getwd()
			Expect(err).ToNot(HaveOccurred())

			err = os.Chdir("../../../fixtures/plugins")
			Expect(err).ToNot(HaveOccurred())

			runCommand(test_curDir)
			_, err = os.Stat(filepath.Join(pluginDir, "test_1.exe"))
			Expect(err).ToNot(HaveOccurred())

			err = os.Chdir(curDir)
			Expect(err).ToNot(HaveOccurred())
		})

		It("copies the plugin into directory <FAKE_HOME_DIR>/.cf/plugins/PLUGIN_FILE_NAME", func() {
			runCommand(test_1)

			_, err := os.Stat(test_1)
			Expect(err).ToNot(HaveOccurred())
			_, err = os.Stat(filepath.Join(pluginDir, "test_1.exe"))
			Expect(err).ToNot(HaveOccurred())
		})

		if runtime.GOOS != "windows" {
			It("Chmods the plugin so it is executable", func() {
				runCommand(test_1)

				fileInfo, err := os.Stat(filepath.Join(pluginDir, "test_1.exe"))
				Expect(err).ToNot(HaveOccurred())
				Expect(int(fileInfo.Mode())).To(Equal(0700))
			})
		}

		It("populate the configuration with plugin metadata", func() {
			runCommand(test_1)

			pluginName, pluginMetadata := pluginConfig.SetPluginArgsForCall(0)

			Expect(pluginName).To(Equal("Test1"))
			Expect(pluginMetadata.Location).To(Equal(filepath.Join(pluginDir, "test_1.exe")))
			Expect(pluginMetadata.Version.Major).To(Equal(1))
			Expect(pluginMetadata.Version.Minor).To(Equal(2))
			Expect(pluginMetadata.Version.Build).To(Equal(4))
			Expect(pluginMetadata.Commands[0].Name).To(Equal("test_1_cmd1"))
			Expect(pluginMetadata.Commands[0].HelpText).To(Equal("help text for test_1_cmd1"))
			Expect(pluginMetadata.Commands[1].Name).To(Equal("test_1_cmd2"))
			Expect(pluginMetadata.Commands[1].HelpText).To(Equal("help text for test_1_cmd2"))
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Installing plugin", test_1},
				[]string{"OK"},
				[]string{"Plugin", "Test1", "v1.2.4", "successfully installed"},
			))
		})

		It("installs multiple plugins with no aliases", func() {
			Expect(runCommand(test_1)).To(Equal(true))
			Expect(runCommand(test_2)).To(Equal(true))
		})
	})
})
