package plugin_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/config_helpers"
	"github.com/cloudfoundry/cli/cf/configuration/plugin_config"
	"github.com/cloudfoundry/gofileutils/fileutils"

	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Uninstall", func() {
	var (
		ui                  *testterm.FakeUI
		requirementsFactory *testreq.FakeReqFactory
		fakePluginRepoDir   string
		pluginDir           string
		pluginConfig        *plugin_config.PluginConfig
		deps                command_registry.Dependency
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.PluginConfig = pluginConfig
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("uninstall-plugin").SetDependency(deps, pluginCall))
	}

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		requirementsFactory = &testreq.FakeReqFactory{}

		var err error
		fakePluginRepoDir, err = ioutil.TempDir("", "plugins")
		Expect(err).ToNot(HaveOccurred())

		fixtureDir := filepath.Join("..", "..", "..", "fixtures", "plugins")

		pluginDir = filepath.Join(fakePluginRepoDir, ".cf", "plugins")
		err = os.MkdirAll(pluginDir, 0700)
		Expect(err).NotTo(HaveOccurred())

		fileutils.CopyPathToPath(filepath.Join(fixtureDir, "test_1.exe"), filepath.Join(pluginDir, "test_1.exe"))
		fileutils.CopyPathToPath(filepath.Join(fixtureDir, "test_2.exe"), filepath.Join(pluginDir, "test_2.exe"))

		config_helpers.PluginRepoDir = func() string {
			return fakePluginRepoDir
		}

		pluginConfig = plugin_config.NewPluginConfig(func(err error) { Expect(err).ToNot(HaveOccurred()) })
		pluginConfig.SetPlugin("test_1.exe", plugin_config.PluginMetadata{Location: filepath.Join(pluginDir, "test_1.exe")})
		pluginConfig.SetPlugin("test_2.exe", plugin_config.PluginMetadata{Location: filepath.Join(pluginDir, "test_2.exe")})
	})

	AfterEach(func() {
		err := os.RemoveAll(fakePluginRepoDir)
		Expect(err).NotTo(HaveOccurred())
	})

	runCommand := func(args ...string) bool {
		return testcmd.RunCliCommand("uninstall-plugin", args, requirementsFactory, updateCommandDependency, false)
	}

	Describe("requirements", func() {
		It("fails with usage when not provided a path to the plugin executable", func() {
			runCommand()
			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Incorrect Usage."},
			))
		})
	})

	Describe("failures", func() {
		It("if plugin name does not exist", func() {
			runCommand("garbage")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Uninstalling plugin garbage..."},
				[]string{"FAILED"},
				[]string{"Plugin name", "garbage", "does not exist"},
			))
		})
	})

	Describe("success", func() {

		Context("notifying plugin of uninstalling", func() {
			var path2file string

			BeforeEach(func() {
				path2file = filepath.Join(os.TempDir(), "uninstall-test-file-for-test_1.exe")

				f, err := os.Create(path2file)
				Ω(err).ToNot(HaveOccurred())
				defer f.Close()
			})

			AfterEach(func() {
				os.Remove(path2file)
			})

			It("notifies the plugin upon uninstalling", func() {
				_, err := os.Stat(path2file)
				Ω(err).ToNot(HaveOccurred())

				runCommand("test_1.exe")

				_, err = os.Stat(path2file)
				Ω(err).To(HaveOccurred())
				Ω(os.IsNotExist(err)).To(BeTrue())
			})
		})

		It("removes the binary from the <FAKE_HOME_DIR>/.cf/plugins dir", func() {
			_, err := os.Stat(filepath.Join(pluginDir, "test_1.exe"))
			Expect(err).ToNot(HaveOccurred())

			runCommand("test_1.exe")

			_, err = os.Stat(filepath.Join(pluginDir, "test_1.exe"))
			Expect(err).To(HaveOccurred())
			Expect(os.IsNotExist(err)).To(BeTrue())
		})

		It("removes the entry from the config.json", func() {
			plugins := pluginConfig.Plugins()
			Expect(plugins).To(HaveKey("test_1.exe"))

			runCommand("test_1.exe")

			plugins = pluginConfig.Plugins()
			Expect(plugins).NotTo(HaveKey("test_1.exe"))
		})

		It("prints success text", func() {
			runCommand("test_1.exe")

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"Uninstalling plugin test_1.exe..."},
				[]string{"OK"},
				[]string{"Plugin", "test_1.exe", "successfully uninstalled."},
			))
		})
	})
})
