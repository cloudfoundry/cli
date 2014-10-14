package plugin_test

import (
	"io/ioutil"
	"net/rpc"
	"os"
	"path/filepath"
	"runtime"

	"github.com/cloudfoundry/cli/cf/command"
	testCommand "github.com/cloudfoundry/cli/cf/command/fakes"
	testconfig "github.com/cloudfoundry/cli/cf/configuration/plugin_config/fakes"
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

		test_1 string
		test_2 string
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
		FIt("if a plugin's command shares the same name as a core command", func() {
			coreCmds["help"] = &testCommand.FakeCommand{}

			x, err := os.Stat(test_1)
			println("name file: ", x.Name())
			Ω(err).ToNot(HaveOccurred())

			runCommand(test_1)

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Plugin 'test_1.exe' cannot be installed from", test_1, "at this time because the command 'cf help' already exists."},
				[]string{"FAILED"},
			))
		})

		It("plugin binary argument is a bad file path", func() {
			runCommand("path/to/not/a/thing.exe")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Binary file 'path/to/not/a/thing.exe' not found"},
				[]string{"FAILED"},
			))
		})

		It("if plugin name is already taken", func() {
			config.PluginsReturns(map[string]string{"Test1": "do/not/care"})
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
				config.PluginsReturns(map[string]string{"useless": "do/not/care"})
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
		var (
			sourceBinaryPath string
		)

		BeforeEach(func() {
			err := os.MkdirAll(pluginDir, 0700)
			Expect(err).ToNot(HaveOccurred())

			sourceBinaryPath = filepath.Join("..", "..", "..", "fixtures", "plugins", "test_1.exe")
			config.GetPluginPathReturns(pluginDir)
			runCommand(sourceBinaryPath)
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

		It("populate the configuration map with the plugin name and location", func() {
			pluginName, pluginExecutable := config.SetPluginArgsForCall(0)

			Expect(pluginName).To(Equal("Test1"))
			Expect(pluginExecutable).To(Equal(filepath.Join(pluginDir, "test_1.exe")))
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Installing plugin", sourceBinaryPath},
				[]string{"OK"},
				[]string{"Plugin", "Test1", "successfully installed"},
			))
		})

	})
})
