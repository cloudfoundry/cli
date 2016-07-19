package pluginconfig_test

import (
	"fmt"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cli/cf/configuration"
	"code.cloudfoundry.org/cli/cf/configuration/confighelpers"
	. "code.cloudfoundry.org/cli/cf/configuration/pluginconfig"
	"code.cloudfoundry.org/cli/plugin"

	"code.cloudfoundry.org/cli/cf/configuration/configurationfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PluginConfig", func() {
	Describe(".ListCommands", func() {
		var (
			pluginConfig *PluginConfig
		)

		BeforeEach(func() {
			pluginConfig = NewPluginConfig(
				func(err error) { Fail(err.Error()) },
				new(configurationfakes.FakePersistor),
				"some-path",
			)

			plugin1 := PluginMetadata{
				Commands: []plugin.Command{
					{
						Name: "plugin1-command1",
					},
					{
						Name: "plugin1-command2",
					},
				},
			}

			pluginConfig.SetPlugin("plugin1", plugin1)
		})

		It("should list the plugin commands", func() {
			Expect(pluginConfig.ListCommands()).To(Equal([]string{
				"plugin1-command1",
				"plugin1-command2",
			}))
		})
	})

	Describe("Integration tests", func() {
		var (
			metadata     PluginMetadata
			commands1    []plugin.Command
			commands2    []plugin.Command
			pluginConfig *PluginConfig
			plugins      map[string]PluginMetadata
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

		JustBeforeEach(func() {
			pluginPath := filepath.Join(confighelpers.PluginRepoDir(), ".cf", "plugins")
			pluginConfig = NewPluginConfig(
				func(err error) {
					if err != nil {
						Fail(fmt.Sprintf("Config error: %s", err))
					}
				},
				configuration.NewDiskPersistor(filepath.Join(pluginPath, "config.json")),
				pluginPath,
			)
			plugins = pluginConfig.Plugins()
		})

		Describe("Reading configuration data", func() {
			BeforeEach(func() {
				confighelpers.PluginRepoDir = func() string {
					return filepath.Join("..", "..", "..", "fixtures", "config", "plugin-config")
				}
			})

			It("returns a list of plugin executables and their location", func() {
				Expect(plugins["Test1"].Location).To(Equal("../../../fixtures/plugins/test_1.exe"))
				Expect(plugins["Test1"].Commands).To(Equal(commands1))
				Expect(plugins["Test2"].Location).To(Equal("../../../fixtures/plugins/test_2.exe"))
				Expect(plugins["Test2"].Commands).To(Equal(commands2))
			})
		})

		Describe("Writing configuration data", func() {
			BeforeEach(func() {
				confighelpers.PluginRepoDir = func() string { return os.TempDir() }
			})

			AfterEach(func() {
				os.Remove(filepath.Join(os.TempDir(), ".cf", "plugins", "config.json"))
			})

			It("saves plugin location and executable information", func() {
				pluginConfig.SetPlugin("foo", metadata)
				plugins = pluginConfig.Plugins()
				Expect(plugins["foo"].Commands).To(Equal(commands1))
			})
		})

		Describe("Removing configuration data", func() {
			BeforeEach(func() {
				confighelpers.PluginRepoDir = func() string { return os.TempDir() }
			})

			AfterEach(func() {
				os.Remove(filepath.Join(os.TempDir()))
			})

			It("removes plugin location and executable information", func() {
				pluginConfig.SetPlugin("foo", metadata)
				plugins = pluginConfig.Plugins()
				Expect(plugins).To(HaveKey("foo"))

				pluginConfig.RemovePlugin("foo")
				plugins = pluginConfig.Plugins()
				Expect(plugins).NotTo(HaveKey("foo"))
			})

			It("handles when the config is not yet initialized", func() {
				pluginConfig.RemovePlugin("foo")
				plugins = pluginConfig.Plugins()
				Expect(plugins).NotTo(HaveKey("foo"))
			})
		})
	})
})
