package plugin_test

import (
	"io/ioutil"
	"net/rpc"
	"os"
	"path/filepath"
	"runtime"

	"github.com/cloudfoundry/cli/cf/command"
	testCommand "github.com/cloudfoundry/cli/cf/command/fakes"
	"github.com/cloudfoundry/cli/cf/configuration/plugin_config"
	testconfig "github.com/cloudfoundry/cli/cf/configuration/plugin_config/fakes"
	"github.com/cloudfoundry/cli/plugin"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/commands/plugin"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Install", func() {
	var (
		ui                  *testterm.FakeUI
		requirementsFactory *testreq.FakeReqFactory
		config              *testconfig.FakePluginConfiguration

		coreCmds   map[string]command.Command
		pluginFile *os.File
		homeDir    string
		pluginDir  string
		curDir     string

		test_1         string
		test_2         string
		test_with_help string
		test_with_push string
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		requirementsFactory = &testreq.FakeReqFactory{}
		config = &testconfig.FakePluginConfiguration{}
		coreCmds = make(map[string]command.Command)

		dir, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		test_1 = filepath.Join(dir, "..", "..", "..", "fixtures", "plugins", "test_1.exe")
		test_2 = filepath.Join(dir, "..", "..", "..", "fixtures", "plugins", "test_2.exe")
		test_with_help = filepath.Join(dir, "..", "..", "..", "fixtures", "plugins", "test_with_help.exe")
		test_with_push = filepath.Join(dir, "..", "..", "..", "fixtures", "plugins", "test_with_push.exe")

		rpc.DefaultServer = rpc.NewServer()

		homeDir, err = ioutil.TempDir(os.TempDir(), "plugins")
		Expect(err).ToNot(HaveOccurred())

		pluginDir = filepath.Join(homeDir, ".cf", "plugins")
		config.GetPluginPathReturns(pluginDir)

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
		cmd := NewPluginInstall(ui, config, coreCmds)
		return testcmd.RunCommand(cmd, args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("fails with usage when not provided a path to the plugin executable", func() {
			Expect(runCommand()).ToNot(HavePassedRequirements())
		})
	})

	Describe("failures", func() {
		Context("when the plugin contains a 'help' command", func() {
			It("fails", func() {
				runCommand(test_with_help)

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Command `help` in the plugin being installed is a native CF command.  Rename the `help` command in the plugin being installed in order to enable its installation and use."},
					[]string{"FAILED"},
				))
			})
		})

		Context("when the plugin contains a core command", func() {
			It("fails", func() {
				coreCmds["push"] = &testCommand.FakeCommand{}
				runCommand(test_with_push)

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Command `push` in the plugin being installed is a native CF command.  Rename the `push` command in the plugin being installed in order to enable its installation and use."},
					[]string{"FAILED"},
				))
			})
		})

		Context("when the plugin contains a command that another installed plugin contains", func() {
			BeforeEach(func() {
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
				config.PluginsReturns(pluginsMap)
			})
			It("fails", func() {
				runCommand(test_1)

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"`test_1_cmd1` is a command in plugin 'Test1Collision'.  You could try uninstalling plugin 'Test1Collision' and then install this plugin in order to invoke the `test_1_cmd1` command.  However, you should first fully understand the impact of uninstalling the existing 'Test1Collision' plugin."},
					[]string{"FAILED"},
				))
			})
		})

		It("plugin binary argument is a bad file path", func() {
			runCommand("path/to/not/a/thing.exe")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Binary file 'path/to/not/a/thing.exe' not found"},
				[]string{"FAILED"},
			))
		})

		It("if plugin name is already taken", func() {
			config.PluginsReturns(map[string]plugin_config.PluginMetadata{"Test1": plugin_config.PluginMetadata{}})
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
				config.PluginsReturns(map[string]plugin_config.PluginMetadata{"useless": plugin_config.PluginMetadata{}})
				config.GetPluginPathReturns(curDir)

				runCommand(filepath.Join(curDir, pluginFile.Name()))
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Installing plugin"},
					[]string{"The file", pluginFile.Name(), "already exists"},
					[]string{"FAILED"},
				))
			})
		})
	})

	Describe("success", func() {
		BeforeEach(func() {
			err := os.MkdirAll(pluginDir, 0700)
			Expect(err).ToNot(HaveOccurred())
			config.GetPluginPathReturns(pluginDir)
			runCommand(test_1)
		})

		It("copies the plugin into directory <FAKE_HOME_DIR>/.cf/plugins/PLUGIN_FILE_NAME", func() {
			_, err := os.Stat(test_1)
			Expect(err).ToNot(HaveOccurred())
			_, err = os.Stat(filepath.Join(pluginDir, "test_1.exe"))
			Expect(err).ToNot(HaveOccurred())
		})

		if runtime.GOOS != "windows" {
			It("Chmods the plugin so it is executable", func() {
				fileInfo, err := os.Stat(filepath.Join(pluginDir, "test_1.exe"))
				Expect(err).ToNot(HaveOccurred())
				Expect(int(fileInfo.Mode())).To(Equal(0700))
			})
		}

		It("populate the configuration with plugin metadata", func() {
			pluginName, pluginMetadata := config.SetPluginArgsForCall(0)

			Expect(pluginName).To(Equal("Test1"))
			Expect(pluginMetadata.Location).To(Equal(filepath.Join(pluginDir, "test_1.exe")))
			Expect(pluginMetadata.Commands[0].Name).To(Equal("test_1_cmd1"))
			Expect(pluginMetadata.Commands[0].HelpText).To(Equal("help text for test_1_cmd1"))
			Expect(pluginMetadata.Commands[1].Name).To(Equal("test_1_cmd2"))
			Expect(pluginMetadata.Commands[1].HelpText).To(Equal("help text for test_1_cmd2"))
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Installing plugin", test_1},
				[]string{"OK"},
				[]string{"Plugin", "Test1", "successfully installed"},
			))
		})
	})
})
