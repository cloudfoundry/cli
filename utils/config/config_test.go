package config_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	. "code.cloudfoundry.org/cli/utils/config"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	var homeDir string

	BeforeEach(func() {
		var err error
		homeDir, err = ioutil.TempDir("", "cli-config-tests")
		Expect(err).NotTo(HaveOccurred())

		os.Setenv("CF_HOME", homeDir)
	})

	AfterEach(func() {
		if homeDir != "" {
			os.RemoveAll(homeDir)
			os.Unsetenv("CF_HOME")
		}
	})

	Context("when there isn't a config set", func() {
		var (
			oldLang  string
			oldLCAll string
		)

		BeforeEach(func() {
			oldLang = os.Getenv("LANG")
			oldLCAll = os.Getenv("LC_ALL")
			os.Unsetenv("LANG")
			os.Unsetenv("LC_ALL")
		})

		It("returns a default config", func() {
			defer os.Setenv("LANG", oldLang)
			defer os.Setenv("LC_ALL", oldLCAll)

			config, err := LoadConfig()
			Expect(err).ToNot(HaveOccurred())

			Expect(config).ToNot(BeNil())
			Expect(config.Target()).To(Equal(DefaultTarget))
			Expect(config.ColorEnabled()).To(Equal(ColorEnabled))
			Expect(config.PluginHome()).To(Equal(filepath.Join(homeDir, ".cf", "plugins")))
			Expect(config.StagingTimeout()).To(Equal(DefaultStagingTimeout))
			Expect(config.StartupTimeout()).To(Equal(DefaultStartupTimeout))
			Expect(config.Locale()).To(BeEmpty())

			pluginConfig := config.PluginConfig()
			Expect(pluginConfig).To(BeEmpty())
		})
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

			plugins := config.PluginConfig()
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

	Describe("Getter Functions", func() {
		DescribeTable("ColorEnabled",
			func(configVal string, envVal string, expected ColorSetting) {
				rawConfig := fmt.Sprintf(`{"ColorEnabled":"%s"}`, configVal)
				setConfig(homeDir, rawConfig)

				defer os.Unsetenv("CF_COLOR")
				if envVal == "" {
					os.Unsetenv("CF_COLOR")
				} else {
					os.Setenv("CF_COLOR", envVal)
				}

				config, err := LoadConfig()
				Expect(err).ToNot(HaveOccurred())
				Expect(config).ToNot(BeNil())

				Expect(config.ColorEnabled()).To(Equal(expected))
			},
			Entry("config=true  env=true  enabled", "true", "true", ColorEnabled),
			Entry("config=true  env=false disabled", "true", "false", ColorDisbled),
			Entry("config=false env=true  enabled", "false", "true", ColorEnabled),
			Entry("config=false env=false disabled", "false", "false", ColorDisbled),

			Entry("config=unset env=false disabled", "", "false", ColorDisbled),
			Entry("config=unset env=true  enabled", "", "true", ColorEnabled),
			Entry("config=false env=unset disabled", "false", "", ColorDisbled),
			Entry("config=true  env=unset disabled", "true", "", ColorEnabled),

			Entry("config=unset env=unset falls back to default", "", "", ColorEnabled),
		)

		Describe("Target", func() {
			var config *Config

			BeforeEach(func() {
				rawConfig := `{ "Target":"https://api.foo.com" }`
				setConfig(homeDir, rawConfig)

				var err error
				config, err = LoadConfig()
				Expect(err).ToNot(HaveOccurred())
				Expect(config).ToNot(BeNil())
			})

			It("returns fields directly from config", func() {
				Expect(config.Target()).To(Equal("https://api.foo.com"))
			})
		})

		DescribeTable("Locale",
			func(langVal string, lcAllVall string, configVal string, expected string) {
				rawConfig := fmt.Sprintf(`{"Locale":"%s"}`, configVal)
				setConfig(homeDir, rawConfig)

				defer os.Unsetenv("LANG")
				if langVal == "" {
					os.Unsetenv("LANG")
				} else {
					os.Setenv("LANG", langVal)
				}

				defer os.Unsetenv("LC_ALL")
				if lcAllVall == "" {
					os.Unsetenv("LC_ALL")
				} else {
					os.Setenv("LC_ALL", lcAllVall)
				}

				config, err := LoadConfig()
				Expect(err).ToNot(HaveOccurred())
				Expect(config).ToNot(BeNil())

				Expect(config.Locale()).To(Equal(expected))
			},

			Entry("LANG=ko-KO.UTF-8 LC_ALL=empty       config=empty ko-KO", "ko-KO.UTF-8", "", "", "ko-KO"),
			Entry("LANG=ko-KO.UTF-8 LC_ALL=fr_FR.UTF-8 config=empty fr-FR", "ko-KO.UTF-8", "fr_FR.UTF-8", "", "fr-FR"),
			Entry("LANG=ko-KO.UTF-8 LC_ALL=fr_FR.UTF-8 config=pt-BR pt-BR", "ko-KO.UTF-8", "fr_FR.UTF-8", "pt-BR", "pt-BR"),

			Entry("config=empty LANG=empty       LC_ALL=empty       default", "", "", "", DefaultLocale),
		)

		Describe("BinaryName", func() {
			It("returns the name used to invoke", func() {
				config, err := LoadConfig()
				Expect(err).ToNot(HaveOccurred())
				Expect(config).ToNot(BeNil())

				// Ginkgo will uses a config file as the first test argument, so that
				// will be considered the binary name
				Expect(config.BinaryName()).To(MatchRegexp("config\\.test$"))
			})
		})

		Context("when there are environment variables", func() {
			var (
				originalCFStagingTimeout string
				originalCFStartupTimeout string
				originalHTTPSProxy       string

				config *Config
			)

			BeforeEach(func() {
				originalCFStagingTimeout = os.Getenv("CF_STAGING_TIMEOUT")
				originalCFStartupTimeout = os.Getenv("CF_STARTUP_TIMEOUT")
				originalHTTPSProxy = os.Getenv("https_proxy")
				os.Setenv("CF_STAGING_TIMEOUT", "8675")
				os.Setenv("CF_STARTUP_TIMEOUT", "309")
				os.Setenv("https_proxy", "proxy.com")

				var err error
				config, err = LoadConfig()
				Expect(err).ToNot(HaveOccurred())
				Expect(config).ToNot(BeNil())
			})

			AfterEach(func() {
				os.Setenv("CF_STAGING_TIMEOUT", originalCFStagingTimeout)
				os.Setenv("CF_STARTUP_TIMEOUT", originalCFStartupTimeout)
				os.Setenv("https_proxy", originalHTTPSProxy)
			})

			It("overrides specific config values", func() {
				Expect(config.StagingTimeout()).To(Equal(time.Duration(8675) * time.Minute))
				Expect(config.StartupTimeout()).To(Equal(time.Duration(309) * time.Minute))
				Expect(config.HTTPSProxy()).To(Equal("proxy.com"))
			})
		})
	})
})
