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
		It("returns a default config", func() {
			config, err := LoadConfig()
			Expect(err).ToNot(HaveOccurred())

			Expect(config).ToNot(BeNil())
			Expect(config.Target()).To(Equal(DefaultTarget))
			Expect(config.ColorEnabled()).To(Equal(ColorEnabled))
			Expect(config.PluginHome()).To(Equal(filepath.Join(homeDir, ".cf", "plugins")))
			Expect(config.StagingTimeout()).To(Equal(DefaultStagingTimeout))
			Expect(config.StartupTimeout()).To(Equal(DefaultStartupTimeout))
		})
	})

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

		Context("when there are environment variables", func() {
			var (
				originalCFPluginHome     string
				originalCFStagingTimeout string
				originalCFStartupTimeout string
				originalHTTPSProxy       string

				config *Config
			)

			BeforeEach(func() {
				originalCFPluginHome = os.Getenv("CF_PLUGIN_HOME")
				originalCFStagingTimeout = os.Getenv("CF_STAGING_TIMEOUT")
				originalCFStartupTimeout = os.Getenv("CF_STARTUP_TIMEOUT")
				originalHTTPSProxy = os.Getenv("https_proxy")
				os.Setenv("CF_PLUGIN_HOME", "/plugins/there")
				os.Setenv("CF_STAGING_TIMEOUT", "8675")
				os.Setenv("CF_STARTUP_TIMEOUT", "309")
				os.Setenv("https_proxy", "proxy.com")

				var err error
				config, err = LoadConfig()
				Expect(err).ToNot(HaveOccurred())
				Expect(config).ToNot(BeNil())
			})

			AfterEach(func() {
				os.Setenv("CF_PLUGIN_HOME", originalCFPluginHome)
				os.Setenv("CF_STAGING_TIMEOUT", originalCFStagingTimeout)
				os.Setenv("CF_STARTUP_TIMEOUT", originalCFStartupTimeout)
				os.Setenv("https_proxy", originalHTTPSProxy)
			})

			It("overrides specific config values", func() {
				Expect(config.PluginHome()).To(Equal("/plugins/there"))
				Expect(config.StagingTimeout()).To(Equal(time.Duration(8675) * time.Minute))
				Expect(config.StartupTimeout()).To(Equal(time.Duration(309) * time.Minute))
				Expect(config.HTTPSProxy()).To(Equal("proxy.com"))
			})
		})
	})
})
