package configv3_test

import (
	"os"
	"path/filepath"

	. "code.cloudfoundry.org/cli/util/configv3"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	var homeDir string

	BeforeEach(func() {
		homeDir = setup()
	})

	AfterEach(func() {
		teardown(homeDir)
	})

	DescribeTable("when the plugin config exists",
		func(setup func() (string, string)) {
			location, CFPluginHome := setup()
			if CFPluginHome != "" {
				os.Setenv("CF_PLUGIN_HOME", CFPluginHome)
				defer os.Unsetenv("CF_PLUGIN_HOME")
			}

			rawConfig := `
{
  "Plugins": {
    "Diego-Enabler": {
      "Location": "~/.cf/plugins/diego-enabler_darwin_amd64",
      "Version": {
        "Major": 1,
        "Minor": 0,
        "Build": 1
      },
      "Commands": [
        {
          "Name": "enable-diego",
          "Alias": "",
          "HelpText": "enable Diego support for an app",
          "UsageDetails": {
            "Usage": "cf enable-diego APP_NAME",
            "Options": null
          }
        },
        {
          "Name": "disable-diego",
          "Alias": "",
          "HelpText": "disable Diego support for an app",
          "UsageDetails": {
            "Usage": "cf disable-diego APP_NAME",
            "Options": null
          }
        }
			]
		}
	}
}`
			setPluginConfig(location, rawConfig)
			config, err := LoadConfig()
			Expect(err).ToNot(HaveOccurred())
			Expect(config).ToNot(BeNil())

			plugins := config.Plugins()
			Expect(plugins).ToNot(BeEmpty())

			plugin := plugins["Diego-Enabler"]
			Expect(plugin.Location).To(Equal("~/.cf/plugins/diego-enabler_darwin_amd64"))
			Expect(plugin.Version.Major).To(Equal(1))
			Expect(plugin.Commands).To(HaveLen(2))
			Expect(plugin.Commands).To(ContainElement(
				PluginCommand{
					Name:     "enable-diego",
					Alias:    "",
					HelpText: "enable Diego support for an app",
					UsageDetails: PluginUsageDetails{
						Usage: "cf enable-diego APP_NAME",
					},
				},
			))
		},

		Entry("standard location", func() (string, string) {
			return filepath.Join(homeDir, ".cf", "plugins"), ""
		}),

		Entry("non-standard location", func() (string, string) {
			return filepath.Join(homeDir, "foo", ".cf", "plugins"), filepath.Join(homeDir, "foo")
		}),
	)
})
