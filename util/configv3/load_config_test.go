package configv3_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cli/command/translatableerror"
	. "code.cloudfoundry.org/cli/util/configv3"

	. "github.com/onsi/ginkgo"
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

	Describe("LoadConfig", func() {
		Context("when there isn't a config set", func() {
			var (
				oldLang  string
				oldLCAll string
			)

			BeforeEach(func() {
				oldLang = os.Getenv("LANG")
				oldLCAll = os.Getenv("LC_ALL")
				Expect(os.Unsetenv("LANG")).ToNot(HaveOccurred())
				Expect(os.Unsetenv("LC_ALL")).ToNot(HaveOccurred())
			})

			It("returns a default config", func() {
				defer os.Setenv("LANG", oldLang)
				defer os.Setenv("LC_ALL", oldLCAll)

				// specifically for when we run unit tests locally
				// we save and unset this variable in case it's present
				// since we want to load a default config
				envVal := os.Getenv("CF_CLI_EXPERIMENTAL")
				Expect(os.Unsetenv("CF_CLI_EXPERIMENTAL")).ToNot(HaveOccurred())

				config, err := LoadConfig()
				Expect(err).ToNot(HaveOccurred())

				// then we reset the env variable
				err = os.Setenv("CF_CLI_EXPERIMENTAL", envVal)
				Expect(err).ToNot(HaveOccurred())

				Expect(config).ToNot(BeNil())
				Expect(config.Target()).To(Equal(DefaultTarget))
				Expect(config.SkipSSLValidation()).To(BeFalse())
				Expect(config.ColorEnabled()).To(Equal(ColorAuto))
				Expect(config.PluginHome()).To(Equal(filepath.Join(homeDir, ".cf", "plugins")))
				Expect(config.StagingTimeout()).To(Equal(DefaultStagingTimeout))
				Expect(config.StartupTimeout()).To(Equal(DefaultStartupTimeout))
				Expect(config.Locale()).To(BeEmpty())
				Expect(config.SSHOAuthClient()).To(Equal(DefaultSSHOAuthClient))
				Expect(config.UAAOAuthClient()).To(Equal(DefaultUAAOAuthClient))
				Expect(config.UAAOAuthClientSecret()).To(Equal(DefaultUAAOAuthClientSecret))
				Expect(config.OverallPollingTimeout()).To(Equal(DefaultOverallPollingTimeout))
				Expect(config.LogLevel()).To(Equal(0))
				Expect(config.DockerPassword()).To(BeEmpty())

				Expect(config.PluginRepositories()).To(Equal([]PluginRepository{{
					Name: "CF-Community",
					URL:  "https://plugins.cloudfoundry.org",
				}}))
				Expect(config.Experimental()).To(BeFalse())

				pluginConfig := config.Plugins()
				Expect(pluginConfig).To(BeEmpty())

				trace, location := config.Verbose()
				Expect(trace).To(BeFalse())
				Expect(location).To(BeEmpty())

				// test the plugins map is initialized
				config.AddPlugin(Plugin{})
			})
		})

		Context("when there is a config set", func() {
			var (
				config *Config
				err    error
			)

			Context("but it is empty", func() {
				var (
					oldLang  string
					oldLCAll string
				)

				BeforeEach(func() {
					oldLang = os.Getenv("LANG")
					oldLCAll = os.Getenv("LC_ALL")
					Expect(os.Unsetenv("LANG")).ToNot(HaveOccurred())
					Expect(os.Unsetenv("LC_ALL")).ToNot(HaveOccurred())

					setConfig(homeDir, "")
				})

				It("returns the default config with a json error", func() {
					defer os.Setenv("LANG", oldLang)
					defer os.Setenv("LC_ALL", oldLCAll)

					// specifically for when we run unit tests locally
					// we save and unset this variable in case it's present
					// since we want to load a default config
					envVal := os.Getenv("CF_CLI_EXPERIMENTAL")
					Expect(os.Unsetenv("CF_CLI_EXPERIMENTAL")).ToNot(HaveOccurred())

					config, err = LoadConfig()
					Expect(err).To(Equal(translatableerror.EmptyConfigError{FilePath: filepath.Join(homeDir, ".cf", "config.json")}))

					// then we reset the env variable
					err = os.Setenv("CF_CLI_EXPERIMENTAL", envVal)
					Expect(err).ToNot(HaveOccurred())

					Expect(config).ToNot(BeNil())
					Expect(config.Target()).To(Equal(DefaultTarget))
					Expect(config.SkipSSLValidation()).To(BeFalse())
					Expect(config.ColorEnabled()).To(Equal(ColorAuto))
					Expect(config.PluginHome()).To(Equal(filepath.Join(homeDir, ".cf", "plugins")))
					Expect(config.StagingTimeout()).To(Equal(DefaultStagingTimeout))
					Expect(config.StartupTimeout()).To(Equal(DefaultStartupTimeout))
					Expect(config.Locale()).To(BeEmpty())
					Expect(config.SSHOAuthClient()).To(Equal(DefaultSSHOAuthClient))
					Expect(config.UAAOAuthClient()).To(Equal(DefaultUAAOAuthClient))
					Expect(config.UAAOAuthClientSecret()).To(Equal(DefaultUAAOAuthClientSecret))
					Expect(config.OverallPollingTimeout()).To(Equal(DefaultOverallPollingTimeout))
					Expect(config.LogLevel()).To(Equal(0))
					Expect(config.DockerPassword()).To(BeEmpty())

					Expect(config.PluginRepositories()).To(Equal([]PluginRepository{{
						Name: "CF-Community",
						URL:  "https://plugins.cloudfoundry.org",
					}}))
					Expect(config.Experimental()).To(BeFalse())

					pluginConfig := config.Plugins()
					Expect(pluginConfig).To(BeEmpty())

					trace, location := config.Verbose()
					Expect(trace).To(BeFalse())
					Expect(location).To(BeEmpty())

					// test the plugins map is initialized
					config.AddPlugin(Plugin{})
				})
			})

			Context("and there are old temp-config* files lingering from previous failed attempts to write the config", func() {
				var (
					oldLang  string
					oldLCAll string
				)

				BeforeEach(func() {
					oldLang = os.Getenv("LANG")
					oldLCAll = os.Getenv("LC_ALL")
					Expect(os.Unsetenv("LANG")).ToNot(HaveOccurred())
					Expect(os.Unsetenv("LC_ALL")).ToNot(HaveOccurred())

					setConfig(homeDir, `{}`)
					configDir := filepath.Join(homeDir, ".cf")
					for i := 0; i < 3; i++ {
						tmpFile, fileErr := ioutil.TempFile(configDir, "temp-config")
						Expect(fileErr).ToNot(HaveOccurred())
						tmpFile.Close()
					}
				})

				It("returns the default config and removes the lingering temp-config* files", func() {
					defer os.Setenv("LANG", oldLang)
					defer os.Setenv("LC_ALL", oldLCAll)

					// specifically for when we run unit tests locally
					// we save and unset this variable in case it's present
					// since we want to load a default config
					envVal := os.Getenv("CF_CLI_EXPERIMENTAL")
					Expect(os.Unsetenv("CF_CLI_EXPERIMENTAL")).ToNot(HaveOccurred())

					config, err = LoadConfig()
					Expect(err).ToNot(HaveOccurred())

					oldTempFileNames, configErr := filepath.Glob(filepath.Join(homeDir, ".cf", "temp-config?*"))
					Expect(configErr).ToNot(HaveOccurred())
					Expect(oldTempFileNames).To(BeEmpty())

					// then we reset the env variable
					err = os.Setenv("CF_CLI_EXPERIMENTAL", envVal)
					Expect(err).ToNot(HaveOccurred())

					Expect(config).ToNot(BeNil())
					Expect(config.Target()).To(Equal(DefaultTarget))
					Expect(config.SkipSSLValidation()).To(BeFalse())
					Expect(config.ColorEnabled()).To(Equal(ColorAuto))
					Expect(config.PluginHome()).To(Equal(filepath.Join(homeDir, ".cf", "plugins")))
					Expect(config.StagingTimeout()).To(Equal(DefaultStagingTimeout))
					Expect(config.StartupTimeout()).To(Equal(DefaultStartupTimeout))
					Expect(config.Locale()).To(BeEmpty())
					Expect(config.SSHOAuthClient()).To(Equal(DefaultSSHOAuthClient))
					Expect(config.UAAOAuthClient()).To(Equal(DefaultUAAOAuthClient))
					Expect(config.UAAOAuthClientSecret()).To(Equal(DefaultUAAOAuthClientSecret))
					Expect(config.OverallPollingTimeout()).To(Equal(DefaultOverallPollingTimeout))
					Expect(config.LogLevel()).To(Equal(0))

					Expect(config.Experimental()).To(BeFalse())

					pluginConfig := config.Plugins()
					Expect(pluginConfig).To(BeEmpty())

					trace, location := config.Verbose()
					Expect(trace).To(BeFalse())
					Expect(location).To(BeEmpty())

					// test the plugins map is initialized
					config.AddPlugin(Plugin{})
				})
			})

			Context("when UAAOAuthClient is not present", func() {
				BeforeEach(func() {
					setConfig(homeDir, `{}`)

					config, err = LoadConfig()
					Expect(err).ToNot(HaveOccurred())
				})

				It("sets UAAOAuthClient to the default", func() {
					Expect(config.UAAOAuthClient()).To(Equal(DefaultUAAOAuthClient))
				})

				It("sets UAAOAuthClientSecret to the default", func() {
					Expect(config.UAAOAuthClientSecret()).To(Equal(DefaultUAAOAuthClientSecret))
				})
			})

			Context("when UAAOAuthClient is empty", func() {
				BeforeEach(func() {
					rawConfig := `
					{
						"UAAOAuthClient": ""
					}`
					setConfig(homeDir, rawConfig)

					config, err = LoadConfig()
					Expect(err).ToNot(HaveOccurred())
				})

				It("sets UAAOAuthClient to the default", func() {
					Expect(config.UAAOAuthClient()).To(Equal(DefaultUAAOAuthClient))
				})

				It("sets UAAOAuthClientSecret to the default", func() {
					Expect(config.UAAOAuthClientSecret()).To(Equal(DefaultUAAOAuthClientSecret))
				})
			})
		})
	})
})
