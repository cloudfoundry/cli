package plugin_config_test

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/cli/cf/configuration/config_helpers"
	. "github.com/cloudfoundry/cli/cf/configuration/plugin_config"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("PluginConfig", func() {

	Describe("Reading configuration data", func() {
		BeforeEach(func() {
			curDir, err := os.Getwd()
			Expect(err).ToNot(HaveOccurred())

			config_helpers.PluginRepoDir = func() string {
				return filepath.Join(curDir, "..", "..", "..", "fixtures", "config", "plugin-config")
			}
		})

		It("returns a list of plugin executables and their location", func() {
			pluginConfig := NewPluginConfig(func(err error) {
				if err != nil {
					panic(fmt.Sprintf("Config error: %s", err))
				}
			})
			plugins := pluginConfig.Plugins()

			Expect(plugins["test_1"]).To(Equal("../../../fixtures/config/plugin-config/.cf/plugins/test_1.exe"))
			Expect(plugins["test_2"]).To(Equal("../../../fixtures/config/plugin-config/.cf/plugins/test_2.exe"))
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

			pluginConfig.SetPlugin("foo", "bar")
			plugins := pluginConfig.Plugins()
			Expect(plugins["foo"]).To(Equal("bar"))
		})
	})
})
