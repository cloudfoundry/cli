package plugin_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/cli/cf/configuration/config_helpers"
	"github.com/cloudfoundry/cli/cf/configuration/plugin_config"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"

	. "github.com/cloudfoundry/cli/cf/commands/plugin"
	"github.com/cloudfoundry/cli/fileutils"
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
	)

	BeforeEach(func() {
		ui = &testterm.FakeUI{}
		requirementsFactory = &testreq.FakeReqFactory{}

		var err error
		fakePluginRepoDir, err = ioutil.TempDir(os.TempDir(), "plugins")
		Expect(err).ToNot(HaveOccurred())

		fixtureDir := filepath.Join("..", "..", "..", "fixtures", "config", "plugin-config", ".cf", "plugins")

		pluginDir = filepath.Join(fakePluginRepoDir, ".cf", "plugins")
		err = os.MkdirAll(pluginDir, 0700)
		Expect(err).NotTo(HaveOccurred())

		fileutils.CopyFile(filepath.Join(pluginDir, "test_1.exe"), filepath.Join(fixtureDir, "test_1.exe"))
		fileutils.CopyFile(filepath.Join(pluginDir, "test_2.exe"), filepath.Join(fixtureDir, "test_2.exe"))

		config_helpers.PluginRepoDir = func() string {
			return fakePluginRepoDir
		}

		pluginConfig = plugin_config.NewPluginConfig(func(err error) { Expect(err).ToNot(HaveOccurred()) })
		pluginConfig.SetPlugin("test_1.exe", filepath.Join(pluginDir, "test_1.exe"))
		pluginConfig.SetPlugin("test_2.exe", filepath.Join(pluginDir, "test_2.exe"))
	})

	AfterEach(func() {
		os.Remove(fakePluginRepoDir)
	})

	runCommand := func(args ...string) bool {
		cmd := NewPluginUninstall(ui, pluginConfig)
		return testcmd.RunCommand(cmd, args, requirementsFactory)
	}

	Describe("requirements", func() {
		It("fails with usage when not provided a path to the plugin executable", func() {
			Expect(runCommand()).ToNot(HavePassedRequirements())
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
