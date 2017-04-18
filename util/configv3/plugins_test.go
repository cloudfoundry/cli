package configv3_test

import (
	"io/ioutil"
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

	Describe("Plugin", func() {
		Describe("CalculateSHA1", func() {
			var plugin Plugin

			Context("when no errors are encountered calculating the sha1 value", func() {
				var file *os.File

				BeforeEach(func() {
					var err error
					file, err = ioutil.TempFile("", "")
					defer file.Close()
					Expect(err).NotTo(HaveOccurred())

					err = ioutil.WriteFile(file.Name(), []byte("foo"), 0600)
					Expect(err).NotTo(HaveOccurred())

					plugin.Location = file.Name()
				})

				AfterEach(func() {
					err := os.Remove(file.Name())
					Expect(err).NotTo(HaveOccurred())
				})

				It("returns the sha1 value", func() {
					Expect(plugin.CalculateSHA1()).To(Equal("0beec7b5ea3f0fdbc95d0dd47f3c5bc275da8a33"))
				})
			})

			Context("when an error is encountered calculating the sha1 value", func() {
				var dirPath string

				BeforeEach(func() {
					var err error
					dirPath, err = ioutil.TempDir("", "")
					Expect(err).NotTo(HaveOccurred())

					plugin.Location = dirPath
				})

				AfterEach(func() {
					err := os.RemoveAll(dirPath)
					Expect(err).NotTo(HaveOccurred())
				})

				It("returns 'N/A'", func() {
					Expect(plugin.CalculateSHA1()).To(Equal("N/A"))
				})
			})
		})
	})

	Describe("PluginVersion", func() {
		var version PluginVersion

		Describe("String", func() {
			It("returns the version in the format x.y.z", func() {
				version = PluginVersion{
					Major: 1,
					Minor: 2,
					Build: 3,
				}
				Expect(version.String()).To(Equal("1.2.3"))
			})

			Context("when the major, minor, and build are all 0", func() {
				BeforeEach(func() {
					version = PluginVersion{
						Major: 0,
						Minor: 0,
						Build: 0,
					}
				})

				It("returns 'N/A'", func() {
					Expect(version.String()).To(Equal("N/A"))
				})
			})
		})
	})

	Describe("PluginCommand", func() {
		var cmd PluginCommand

		Describe("CommandName", func() {
			It("returns the name of the command", func() {
				cmd = PluginCommand{Name: "some-command"}
				Expect(cmd.CommandName()).To(Equal("some-command"))
			})

			Context("when the command name and command alias are not empty", func() {
				BeforeEach(func() {
					cmd = PluginCommand{Name: "some-command", Alias: "sp"}
				})

				It("returns the command name concatenated with the command alias", func() {
					Expect(cmd.CommandName()).To(Equal("some-command, sp"))
				})
			})
		})
	})

	Describe("RemovePlugin", func() {
		var (
			config *Config
			err    error
		)

		BeforeEach(func() {
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
        }
      ]
    },
    "Dora-Non-Enabler": {
      "Location": "~/.cf/plugins/diego-enabler_darwin_amd64",
      "Version": {
        "Major": 1,
        "Minor": 0,
        "Build": 1
      },
      "Commands": [
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
			setPluginConfig(filepath.Join(homeDir, ".cf", "plugins"), rawConfig)

			config, err = LoadConfig()
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when the plugin exists", func() {
			It("removes the plugin from the config", func() {
				plugins := config.Plugins()

				_, found := plugins["Diego-Enabler"]
				Expect(found).To(BeTrue())

				config.RemovePlugin("Diego-Enabler")

				_, found = plugins["Diego-Enabler"]
				Expect(found).To(BeFalse())
			})

			Context("when the plugin does not exist", func() {
				It("doesn't blow up", func() {
					config.RemovePlugin("does-not-exist")
				})
			})
		})
	})

	Describe("WritePluginConfig", func() {
		var config *Config

		BeforeEach(func() {
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
				}
			]
		}
	}
}`
			setPluginConfig(filepath.Join(homeDir, ".cf", "plugins"), rawConfig)

			var err error
			config, err = LoadConfig()
			Expect(err).ToNot(HaveOccurred())
		})

		Context("when no errors are encountered", func() {
			It("writes the plugin config to pluginHome/.cf/plugin/config.json", func() {
				plugin := config.Plugins()["Diego-Enabler"]
				plugin.Location = "BAR"
				config.Plugins()["Diego-Enabler"] = plugin

				err := config.WritePluginConfig()
				Expect(err).ToNot(HaveOccurred())

				newConfig, err := LoadConfig()
				Expect(err).ToNot(HaveOccurred())
				Expect(newConfig.Plugins()["Diego-Enabler"].Location).To(Equal("BAR"))
			})
		})

		Context("when an error is encountered", func() {
			BeforeEach(func() {
				err := os.Chmod(filepath.Join(homeDir, ".cf", "plugins", "config.json"), 0000)
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns the error", func() {
				err := config.WritePluginConfig()
				Expect(err).To(MatchError(MatchRegexp("permission denied")))
			})
		})
	})
})
