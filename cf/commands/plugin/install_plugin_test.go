package plugin_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"

	"code.cloudfoundry.org/cli/cf/actors/pluginrepo/pluginrepofakes"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/commandregistry/commandregistryfakes"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/configuration/pluginconfig"
	"code.cloudfoundry.org/cli/cf/configuration/pluginconfig/pluginconfigfakes"
	"code.cloudfoundry.org/cli/cf/flags"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/cf/requirements"
	"code.cloudfoundry.org/cli/cf/requirements/requirementsfakes"
	"code.cloudfoundry.org/cli/plugin"
	testcmd "code.cloudfoundry.org/cli/util/testhelpers/commands"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"
	"code.cloudfoundry.org/cli/util/utilfakes"

	clipr "github.com/cloudfoundry/cli-plugin-repo/web"

	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Install", func() {
	var (
		ui                  *testterm.FakeUI
		requirementsFactory *requirementsfakes.FakeFactory
		config              coreconfig.Repository
		pluginConfig        *pluginconfigfakes.FakePluginConfiguration
		fakePluginRepo      *pluginrepofakes.FakePluginRepo
		fakeChecksum        *utilfakes.FakeSha1Checksum

		pluginFile *os.File
		homeDir    string
		pluginDir  string
		curDir     string

		test_1                    string
		test_2                    string
		test_curDir               string
		test_with_help            string
		test_with_orgs            string
		test_with_orgs_short_name string
		aliasConflicts            string
		deps                      commandregistry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.UI = ui
		deps.Config = config
		deps.PluginConfig = pluginConfig
		deps.PluginRepo = fakePluginRepo
		deps.ChecksumUtil = fakeChecksum
		commandregistry.Commands.SetCommand(commandregistry.Commands.FindCommand("install-plugin").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		requirementsFactory = new(requirementsfakes.FakeFactory)
		pluginConfig = new(pluginconfigfakes.FakePluginConfiguration)
		config = testconfig.NewRepositoryWithDefaults()
		fakePluginRepo = new(pluginrepofakes.FakePluginRepo)
		fakeChecksum = new(utilfakes.FakeSha1Checksum)

		dir, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		test_1 = filepath.Join(dir, "..", "..", "..", "fixtures", "plugins", "test_1.exe")
		test_2 = filepath.Join(dir, "..", "..", "..", "fixtures", "plugins", "test_2.exe")
		test_curDir = filepath.Join("test_1.exe")
		test_with_help = filepath.Join(dir, "..", "..", "..", "fixtures", "plugins", "test_with_help.exe")
		test_with_orgs = filepath.Join(dir, "..", "..", "..", "fixtures", "plugins", "test_with_orgs.exe")
		test_with_orgs_short_name = filepath.Join(dir, "..", "..", "..", "fixtures", "plugins", "test_with_orgs_short_name.exe")
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
		os.RemoveAll(filepath.Join(curDir, pluginFile.Name()))
		os.RemoveAll(homeDir)
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCLICommand("install-plugin", args, requirementsFactory, updateCommandDependency, false, ui)
	}

	Describe("requirements", func() {
		It("fails with usage when not provided a path to the plugin executable", func() {
			Expect(runCommand()).ToNot(HavePassedRequirements())
		})
	})

	Context("when the -f flag is not provided", func() {
		Context("and the user responds with 'y'", func() {
			It("continues to install the plugin", func() {
				ui.Inputs = []string{"y"}
				runCommand("pluggy", "-r", "somerepo")
				Expect(ui.Outputs()).To(ContainSubstrings([]string{"Looking up 'pluggy' from repository 'somerepo'"}))
			})
		})

		Context("but the user responds with 'n'", func() {
			It("quits with a message", func() {
				ui.Inputs = []string{"n"}
				runCommand("pluggy", "-r", "somerepo")
				Expect(ui.Outputs()).To(ContainSubstrings([]string{"Plugin installation cancelled"}))
			})
		})
	})

	Describe("Locating binary file", func() {

		Describe("install from plugin repository when '-r' provided", func() {
			Context("gets metadata of the plugin from repo", func() {
				Context("when repo is not found in config", func() {
					It("informs user repo is not found", func() {
						runCommand("plugin1", "-r", "repo1", "-f")
						Expect(ui.Outputs()).To(ContainSubstrings([]string{"Looking up 'plugin1' from repository 'repo1'"}))
						Expect(ui.Outputs()).To(ContainSubstrings([]string{"repo1 not found"}))
					})
				})

				Context("when repo is found in config", func() {
					Context("when repo endpoint returns an error", func() {
						It("informs user about the error", func() {
							config.SetPluginRepo(models.PluginRepo{Name: "repo1", URL: ""})
							fakePluginRepo.GetPluginsReturns(nil, []string{"repo error1"})
							runCommand("plugin1", "-r", "repo1", "-f")

							Expect(ui.Outputs()).To(ContainSubstrings([]string{"Error getting plugin metadata from repo"}))
							Expect(ui.Outputs()).To(ContainSubstrings([]string{"repo error1"}))
						})
					})

					Context("when plugin metadata is available and desired plugin is not found", func() {
						It("informs user about the error", func() {
							config.SetPluginRepo(models.PluginRepo{Name: "repo1", URL: ""})
							fakePluginRepo.GetPluginsReturns(nil, nil)
							runCommand("plugin1", "-r", "repo1", "-f")

							Expect(ui.Outputs()).To(ContainSubstrings([]string{"plugin1 is not available in repo 'repo1'"}))
						})
					})

					It("ignore cases in repo name", func() {
						config.SetPluginRepo(models.PluginRepo{Name: "repo1", URL: ""})
						fakePluginRepo.GetPluginsReturns(nil, nil)
						runCommand("plugin1", "-r", "REPO1", "-f")

						Expect(ui.Outputs()).NotTo(ContainSubstrings([]string{"REPO1 not found"}))
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

						config.SetPluginRepo(models.PluginRepo{Name: "repo1", URL: ""})
						fakePluginRepo.GetPluginsReturns(result, nil)
						runCommand("plugin1", "-r", "repo1", "-f")

						Expect(ui.Outputs()).To(ContainSubstrings([]string{"Plugin requested has no binary available"}))
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
								{
									Platform: "osx",
									Url:      testServer.URL + "/test.exe",
								},
								{
									Platform: "win64",
									Url:      testServer.URL + "/test.exe",
								},
								{
									Platform: "win32",
									Url:      testServer.URL + "/test.exe",
								},
								{
									Platform: "linux32",
									Url:      testServer.URL + "/test.exe",
								},
								{
									Platform: "linux64",
									Url:      testServer.URL + "/test.exe",
								},
							},
						}
						result := make(map[string][]clipr.Plugin)
						result["repo1"] = []clipr.Plugin{p}

						config.SetPluginRepo(models.PluginRepo{Name: "repo1", URL: ""})
						fakePluginRepo.GetPluginsReturns(result, nil)
					})

					AfterEach(func() {
						testServer.Close()
					})

					It("performs sha1 checksum validation on the downloaded binary", func() {
						runCommand("plugin1", "-r", "repo1", "-f")
						Expect(fakeChecksum.CheckSha1CallCount()).To(Equal(1))
					})

					It("reports error downloaded file's sha1 does not match the sha1 in metadata", func() {
						fakeChecksum.CheckSha1Returns(false)

						runCommand("plugin1", "-r", "repo1", "-f")
						Expect(ui.Outputs()).To(ContainSubstrings(
							[]string{"FAILED"},
							[]string{"checksum does not match"},
						))

					})

					It("downloads and installs binary when it is available and checksum matches", func() {
						runCommand("plugin1", "-r", "repo1", "-f")

						Expect(ui.Outputs()).To(ContainSubstrings([]string{"4 bytes downloaded..."}))
						Expect(ui.Outputs()).To(ContainSubstrings([]string{"FAILED"}))
						Expect(ui.Outputs()).To(ContainSubstrings([]string{"Installing plugin"}))
					})
				})
			})
		})

		Describe("install from plugin repository with no '-r' provided", func() {
			Context("downloads file from internet if path prefix with 'http','ftp' etc...", func() {
				It("will not try locate file locally", func() {
					runCommand("http://127.0.0.1/plugin.exe", "-f")

					Expect(ui.Outputs()).ToNot(ContainSubstrings(
						[]string{"File not found locally"},
					))
					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"download binary file from internet address"},
					))
				})

				It("informs users when binary is not downloadable from net", func() {
					runCommand("http://path/to/not/a/thing.exe", "-f")

					Expect(ui.Outputs()).To(ContainSubstrings(
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

					runCommand(testServer.URL+"/testfile.exe", "-f")

					Expect(ui.Outputs()).To(ContainSubstrings([]string{"3 bytes downloaded..."}))
					Expect(ui.Outputs()).To(ContainSubstrings([]string{"FAILED"}))
					Expect(ui.Outputs()).To(ContainSubstrings([]string{"Installing plugin"}))
				})
			})

			Context("tries to locate binary file at local path if path has no internet prefix", func() {
				It("installs the plugin from a local file if found", func() {
					runCommand("./install_plugin.go", "-f")

					Expect(ui.Outputs()).ToNot(ContainSubstrings(
						[]string{"download binary file from internet"},
					))
					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"Installing plugin install_plugin.go"},
					))
				})

				It("reports error if local file is not found at given path", func() {
					runCommand("./no/file/is/here.exe", "-f")

					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"File not found locally",
							"./no/file/is/here.exe",
						},
					))
				})
			})
		})

	})

	Describe("install failures", func() {
		Context("when the plugin contains a 'help' command", func() {
			It("fails", func() {
				runCommand(test_with_help, "-f")

				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Command `help` in the plugin being installed is a native CF command/alias.  Rename the `help` command in the plugin being installed in order to enable its installation and use."},
					[]string{"FAILED"},
				))
			})
		})

		Context("when the plugin's command conflicts with a core command/alias", func() {
			var originalCommand commandregistry.Command

			BeforeEach(func() {
				originalCommand = commandregistry.Commands.FindCommand("org")

				commandregistry.Register(testOrgsCmd{})
			})

			AfterEach(func() {
				if originalCommand != nil {
					commandregistry.Register(originalCommand)
				}
			})

			It("fails if is shares a command name", func() {
				runCommand(test_with_orgs, "-f")

				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Command `orgs` in the plugin being installed is a native CF command/alias.  Rename the `orgs` command in the plugin being installed in order to enable its installation and use."},
					[]string{"FAILED"},
				))
			})

			It("fails if it shares a command short name", func() {
				runCommand(test_with_orgs_short_name, "-f")

				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Command `o` in the plugin being installed is a native CF command/alias.  Rename the `o` command in the plugin being installed in order to enable its installation and use."},
					[]string{"FAILED"},
				))
			})
		})

		Context("when the plugin's alias conflicts with a core command/alias", func() {
			var fakeCmd *commandregistryfakes.FakeCommand
			BeforeEach(func() {
				fakeCmd = new(commandregistryfakes.FakeCommand)
			})

			AfterEach(func() {
				commandregistry.Commands.RemoveCommand("non-conflict-cmd")
				commandregistry.Commands.RemoveCommand("conflict-alias")
			})

			It("fails if it shares a command name", func() {
				fakeCmd.MetaDataReturns(commandregistry.CommandMetadata{Name: "conflict-alias"})
				commandregistry.Register(fakeCmd)

				runCommand(aliasConflicts, "-f")

				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Alias `conflict-alias` in the plugin being installed is a native CF command/alias.  Rename the `conflict-alias` command in the plugin being installed in order to enable its installation and use."},
					[]string{"FAILED"},
				))
			})

			It("fails if it shares a command short name", func() {
				fakeCmd.MetaDataReturns(commandregistry.CommandMetadata{Name: "non-conflict-cmd", ShortName: "conflict-alias"})
				commandregistry.Register(fakeCmd)

				runCommand(aliasConflicts, "-f")

				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Alias `conflict-alias` in the plugin being installed is a native CF command/alias.  Rename the `conflict-alias` command in the plugin being installed in order to enable its installation and use."},
					[]string{"FAILED"},
				))
			})
		})

		Context("when the plugin's alias conflicts with other installed plugin", func() {
			It("fails if it shares a command name", func() {
				pluginsMap := make(map[string]pluginconfig.PluginMetadata)
				pluginsMap["AliasCollision"] = pluginconfig.PluginMetadata{
					Location: "location/to/config.exe",
					Commands: []plugin.Command{
						{
							Name:     "conflict-alias",
							HelpText: "Hi!",
						},
					},
				}
				pluginConfig.PluginsReturns(pluginsMap)

				runCommand(aliasConflicts, "-f")

				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Alias `conflict-alias` is a command/alias in plugin 'AliasCollision'.  You could try uninstalling plugin 'AliasCollision' and then install this plugin in order to invoke the `conflict-alias` command.  However, you should first fully understand the impact of uninstalling the existing 'AliasCollision' plugin."},
					[]string{"FAILED"},
				))
			})

			It("fails if it shares a command alias", func() {
				pluginsMap := make(map[string]pluginconfig.PluginMetadata)
				pluginsMap["AliasCollision"] = pluginconfig.PluginMetadata{
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

				runCommand(aliasConflicts, "-f")

				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Alias `conflict-alias` is a command/alias in plugin 'AliasCollision'.  You could try uninstalling plugin 'AliasCollision' and then install this plugin in order to invoke the `conflict-alias` command.  However, you should first fully understand the impact of uninstalling the existing 'AliasCollision' plugin."},
					[]string{"FAILED"},
				))
			})
		})

		Context("when the plugin's command conflicts with other installed plugin", func() {
			It("fails if it shares a command name", func() {
				pluginsMap := make(map[string]pluginconfig.PluginMetadata)
				pluginsMap["Test1Collision"] = pluginconfig.PluginMetadata{
					Location: "location/to/config.exe",
					Commands: []plugin.Command{
						{
							Name:     "test_1_cmd1",
							HelpText: "Hi!",
						},
					},
				}
				pluginConfig.PluginsReturns(pluginsMap)

				runCommand(test_1, "-f")

				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Command `test_1_cmd1` is a command/alias in plugin 'Test1Collision'.  You could try uninstalling plugin 'Test1Collision' and then install this plugin in order to invoke the `test_1_cmd1` command.  However, you should first fully understand the impact of uninstalling the existing 'Test1Collision' plugin."},
					[]string{"FAILED"},
				))
			})

			It("fails if it shares a command alias", func() {
				pluginsMap := make(map[string]pluginconfig.PluginMetadata)
				pluginsMap["AliasCollision"] = pluginconfig.PluginMetadata{
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

				runCommand(aliasConflicts, "-f")

				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Command `conflict-cmd` is a command/alias in plugin 'AliasCollision'.  You could try uninstalling plugin 'AliasCollision' and then install this plugin in order to invoke the `conflict-cmd` command.  However, you should first fully understand the impact of uninstalling the existing 'AliasCollision' plugin."},
					[]string{"FAILED"},
				))
			})
		})

		It("if plugin name is already taken", func() {
			pluginConfig.PluginsReturns(map[string]pluginconfig.PluginMetadata{"Test1": {}})
			runCommand(test_1, "-f")

			Expect(ui.Outputs()).To(ContainSubstrings(
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
				pluginConfig.PluginsReturns(map[string]pluginconfig.PluginMetadata{"useless": {}})
				pluginConfig.GetPluginPathReturns(curDir)

				runCommand(filepath.Join(curDir, pluginFile.Name()), "-f")
				Expect(ui.Outputs()).To(ContainSubstrings(
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

			runCommand(test_curDir, "-f")
			_, err = os.Stat(filepath.Join(pluginDir, "test_1.exe"))
			Expect(err).ToNot(HaveOccurred())

			err = os.Chdir(curDir)
			Expect(err).ToNot(HaveOccurred())
		})

		It("copies the plugin into directory <FAKE_HOME_DIR>/.cf/plugins/PLUGIN_FILE_NAME", func() {
			runCommand(test_1, "-f")

			_, err := os.Stat(test_1)
			Expect(err).ToNot(HaveOccurred())
			_, err = os.Stat(filepath.Join(pluginDir, "test_1.exe"))
			Expect(err).ToNot(HaveOccurred())
		})

		if runtime.GOOS != "windows" {
			It("Chmods the plugin so it is executable", func() {
				runCommand(test_1, "-f")

				fileInfo, err := os.Stat(filepath.Join(pluginDir, "test_1.exe"))
				Expect(err).ToNot(HaveOccurred())
				Expect(int(fileInfo.Mode())).To(Equal(0700))
			})
		}

		It("populate the configuration with plugin metadata", func() {
			runCommand(test_1, "-f")

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
			Expect(ui.Outputs()).To(ContainSubstrings(
				[]string{"Installing plugin test_1.exe"},
				[]string{"OK"},
				[]string{"Plugin", "Test1", "v1.2.4", "successfully installed"},
			))
		})

		It("installs multiple plugins with no aliases", func() {
			Expect(runCommand(test_1, "-f")).To(Equal(true))
			Expect(runCommand(test_2, "-f")).To(Equal(true))
		})
	})
})

type testOrgsCmd struct{}

func (t testOrgsCmd) MetaData() commandregistry.CommandMetadata {
	return commandregistry.CommandMetadata{
		Name:      "orgs",
		ShortName: "o",
	}
}

func (cmd testOrgsCmd) Requirements(requirementsFactory requirements.Factory, fc flags.FlagContext) ([]requirements.Requirement, error) {
	return []requirements.Requirement{}, nil
}

func (cmd testOrgsCmd) SetDependency(deps commandregistry.Dependency, pluginCall bool) (c commandregistry.Command) {
	return
}

func (cmd testOrgsCmd) Execute(c flags.FlagContext) error {
	return nil
}
