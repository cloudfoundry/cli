package plugin_config_test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/cli/cf/configuration/config_helpers"
	. "github.com/cloudfoundry/cli/cf/configuration/plugin_config"
	"github.com/cloudfoundry/cli/plugin"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PluginConfig", func() {
	var (
		metadata  PluginMetadata
		commands1 []plugin.Command
		commands2 []plugin.Command
	)

	BeforeEach(func() {
		commands1 = []plugin.Command{
			{
				Name:     "test_1_cmd1",
				HelpText: "help text for test1 cmd1",
			},
			{
				Name:     "test_1_cmd2",
				HelpText: "help text for test1 cmd2",
			},
		}

		commands2 = []plugin.Command{
			{
				Name:     "test_2_cmd1",
				HelpText: "help text for test2 cmd1",
			},
			{
				Name:     "test_2_cmd2",
				HelpText: "help text for test2 cmd2",
			},
		}

		metadata = PluginMetadata{
			Location: "../../../fixtures/plugins/test_1.exe",
			Commands: commands1,
		}
	})

	Describe("Reading configuration data", func() {
		BeforeEach(func() {

			config_helpers.PluginRepoDir = func() string {
				return filepath.Join("..", "..", "..", "fixtures", "config", "plugin-config")
			}
		})

		It("returns a list of plugin executables and their location", func() {
			pluginConfig := NewPluginConfig(func(err error) {
				if err != nil {
					panic(fmt.Sprintf("Config error: %s", err))
				}
			})
			plugins := pluginConfig.Plugins()

			Expect(plugins["Test1"].Location).To(Equal("../../../fixtures/plugins/test_1.exe"))
			Expect(plugins["Test1"].Commands).To(Equal(commands1))
			Expect(plugins["Test2"].Location).To(Equal("../../../fixtures/plugins/test_2.exe"))
			Expect(plugins["Test2"].Commands).To(Equal(commands2))
		})
	})

	Describe("Writing configuration data", func() {
		BeforeEach(func() {
			config_helpers.PluginRepoDir = func() string { return os.TempDir() }
		})

		AfterEach(func() {
			os.Remove(filepath.Join(os.TempDir(), ".cf", "plugins", "config.json"))
		})

		It("saves plugin location and executable information", func() {
			pluginConfig := NewPluginConfig(func(err error) {
				if err != nil {
					panic(fmt.Sprintf("Config error: %s", err))
				}
			})

			pluginConfig.SetPlugin("foo", metadata)
			plugins := pluginConfig.Plugins()
			Expect(plugins["foo"].Commands).To(Equal(commands1))
		})
	})

	Describe("Removing configuration data", func() {
		var (
			pluginConfig *PluginConfig
		)

		BeforeEach(func() {
			config_helpers.PluginRepoDir = func() string { return os.TempDir() }
			pluginConfig = NewPluginConfig(func(err error) {
				if err != nil {
					panic(fmt.Sprintf("Config error: %s", err))
				}
			})
		})

		AfterEach(func() {
			os.Remove(filepath.Join(os.TempDir()))
		})

		It("removes plugin location and executable information", func() {
			pluginConfig.SetPlugin("foo", metadata)

			plugins := pluginConfig.Plugins()
			Expect(plugins).To(HaveKey("foo"))

			pluginConfig.RemovePlugin("foo")

			plugins = pluginConfig.Plugins()
			Expect(plugins).NotTo(HaveKey("foo"))
		})

		It("handles when the config is not yet initialized", func() {
			pluginConfig.RemovePlugin("foo")

			plugins := pluginConfig.Plugins()
			Expect(plugins).NotTo(HaveKey("foo"))
		})
	})
})
