package configv3_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/integration/helpers"
	"code.cloudfoundry.org/cli/util/configv3"
	. "code.cloudfoundry.org/cli/util/configv3"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	var (
		oldLang           string
		oldLCAll          string
		oldCfExperimental string
		homeDir           string
	)

	BeforeEach(func() {
		homeDir = setup()

		oldLang = os.Getenv("LANG")
		oldLCAll = os.Getenv("LC_ALL")
		// specifically for when we run unit tests locally
		// we save and unset this variable in case it's present
		// since we want to load a default config
		oldCfExperimental = os.Getenv("CF_CLI_EXPERIMENTAL")
		Expect(os.Unsetenv("LANG")).ToNot(HaveOccurred())
		Expect(os.Unsetenv("LC_ALL")).ToNot(HaveOccurred())
		Expect(os.Unsetenv("CF_CLI_EXPERIMENTAL")).To(Succeed())
	})

	AfterEach(func() {
		os.Setenv("LANG", oldLang)
		os.Setenv("LC_ALL", oldLCAll)
		os.Setenv("CF_CLI_EXPERIMENTAL", oldCfExperimental)
		teardown(homeDir)
	})

	Describe("LoadConfig", func() {
		var (
			config  *Config
			loadErr error
			inFlags []FlagOverride
		)

		BeforeEach(func() {
			inFlags = []FlagOverride{}
		})

		JustBeforeEach(func() {
			config, loadErr = LoadConfig(inFlags...)
		})

		When("there are old temp-config* files lingering from previous failed attempts to write the config", func() {
			BeforeEach(func() {
				configDir := filepath.Join(homeDir, ".cf")
				Expect(os.MkdirAll(configDir, 0777)).To(Succeed())
				for i := 0; i < 3; i++ {
					tmpFile, fileErr := ioutil.TempFile(configDir, "temp-config")
					Expect(fileErr).ToNot(HaveOccurred())
					tmpFile.Close()
				}
			})

			It("removes the lingering temp-config* files", func() {
				Expect(loadErr).ToNot(HaveOccurred())

				oldTempFileNames, configErr := filepath.Glob(filepath.Join(homeDir, ".cf", "temp-config?*"))
				Expect(configErr).ToNot(HaveOccurred())
				Expect(oldTempFileNames).To(BeEmpty())
			})
		})

		When("there isn't a config set", func() {
			It("returns a default config", func() {
				Expect(loadErr).ToNot(HaveOccurred())

				Expect(config).ToNot(BeNil())
				Expect(config.ConfigFile).To(Equal(
					JSONConfig{
						ColorEnabled:         DefaultColorEnabled,
						ConfigVersion:        configv3.CurrentConfigVersion,
						SSHOAuthClient:       DefaultSSHOAuthClient,
						UAAOAuthClient:       DefaultUAAOAuthClient,
						UAAOAuthClientSecret: DefaultUAAOAuthClientSecret,
						PluginRepositories: []PluginRepository{{
							Name: DefaultPluginRepoName,
							URL:  DefaultPluginRepoURL,
						}},
						Target: DefaultTarget,
					},
				))
				Expect(config.ENV).To(Equal(
					EnvOverride{
						BinaryName: "configv3.test", // Ginkgo will uses a config file as the first test argument, so that will be considered the binary name
					},
				))
				Expect(config.Flags).To(Equal(FlagOverride{}))
				Expect(config.PluginHome()).To(Equal(filepath.Join(homeDir, ".cf", "plugins")))

				pluginConfig := config.Plugins()
				Expect(pluginConfig).To(BeEmpty())

				// test the plugins map is initialized
				config.AddPlugin(Plugin{})
			})
		})

		When("there is a config set", func() {
			Context("but it is empty", func() {
				BeforeEach(func() {
					setConfig(homeDir, "")
				})

				It("returns the default config and an EmptyConfigError", func() {
					Expect(loadErr).To(Equal(translatableerror.EmptyConfigError{FilePath: filepath.Join(homeDir, ".cf", "config.json")}))
					Expect(config).ToNot(BeNil())
					Expect(config).ToNot(BeNil())
					Expect(config.ConfigFile).To(Equal(
						JSONConfig{
							ColorEnabled:         DefaultColorEnabled,
							ConfigVersion:        configv3.CurrentConfigVersion,
							SSHOAuthClient:       DefaultSSHOAuthClient,
							UAAOAuthClient:       DefaultUAAOAuthClient,
							UAAOAuthClientSecret: DefaultUAAOAuthClientSecret,
							PluginRepositories: []PluginRepository{{
								Name: DefaultPluginRepoName,
								URL:  DefaultPluginRepoURL,
							}},
							Target: DefaultTarget,
						},
					))
					Expect(config.ENV).To(Equal(
						EnvOverride{
							BinaryName: "configv3.test", // Ginkgo will uses a config file as the first test argument, so that will be considered the binary name
						},
					))
					Expect(config.Flags).To(Equal(FlagOverride{}))
					Expect(config.PluginHome()).To(Equal(filepath.Join(homeDir, ".cf", "plugins")))

					pluginConfig := config.Plugins()
					Expect(pluginConfig).To(BeEmpty())
				})
			})

			When("UAAOAuthClient is not present", func() {
				BeforeEach(func() {
					setConfig(homeDir, `{}`)
				})

				It("sets UAAOAuthClient to the default", func() {
					Expect(loadErr).ToNot(HaveOccurred())
					Expect(config.UAAOAuthClient()).To(Equal(DefaultUAAOAuthClient))
				})

				It("sets UAAOAuthClientSecret to the default", func() {
					Expect(loadErr).ToNot(HaveOccurred())
					Expect(config.UAAOAuthClientSecret()).To(Equal(DefaultUAAOAuthClientSecret))
				})
			})

			When("UAAOAuthClient is empty", func() {
				BeforeEach(func() {
					rawConfig := `
					{
						"UAAOAuthClient": ""
					}`
					setConfig(homeDir, rawConfig)
				})

				It("sets UAAOAuthClient to the default", func() {
					Expect(loadErr).ToNot(HaveOccurred())
					Expect(config.UAAOAuthClient()).To(Equal(DefaultUAAOAuthClient))
				})

				It("sets UAAOAuthClientSecret to the default", func() {
					Expect(loadErr).ToNot(HaveOccurred())
					Expect(config.UAAOAuthClientSecret()).To(Equal(DefaultUAAOAuthClientSecret))
				})
			})

			When("Version checking", func() {
				When("the Config Version is < the current version", func() {
					It("clears the config", func() {
						rawConfig := fmt.Sprintf(`
							{
								"AccessToken": "bearer shazbat!",
								"ConfigVersion": %d
							}`, configv3.CurrentConfigVersion-1)
						setConfig(homeDir, rawConfig)
						config := helpers.GetConfig()
						Expect(loadErr).ToNot(HaveOccurred())
						Expect(config.ConfigFile.ConfigVersion).To(Equal(configv3.CurrentConfigVersion))
						Expect(config.ConfigFile.AccessToken).To(Equal(""))
					})
				})

				When("the Config Version is = the current version", func() {
					It("clears the config", func() {
						rawConfig := fmt.Sprintf(`
					{
						"AccessToken": "bearer shazbat!",
						"ConfigVersion": %d
					}`, configv3.CurrentConfigVersion)
						setConfig(homeDir, rawConfig)
						config := helpers.GetConfig()
						Expect(loadErr).ToNot(HaveOccurred())
						Expect(config.ConfigFile.ConfigVersion).To(Equal(configv3.CurrentConfigVersion))
						Expect(config.ConfigFile.AccessToken).To(Equal("bearer shazbat!"))
					})
				})

				When("the Config Version is > the current version", func() {
					It("clears the config", func() {
						rawConfig := fmt.Sprintf(`
					{
						"AccessToken": "bearer shazbat!",
						"ConfigVersion": %d
					}`, configv3.CurrentConfigVersion+1)
						setConfig(homeDir, rawConfig)
						config := helpers.GetConfig()
						Expect(loadErr).ToNot(HaveOccurred())
						Expect(config.ConfigFile.ConfigVersion).To(Equal(configv3.CurrentConfigVersion))
						Expect(config.ConfigFile.AccessToken).To(Equal(""))
					})
				})
			})
		})

		When("passed flag overrides", func() {
			BeforeEach(func() {
				inFlags = append(inFlags, FlagOverride{Verbose: true}, FlagOverride{Verbose: false})
			})

			It("stores the first set of flag overrides on the config", func() {
				Expect(config.Flags).To(Equal(FlagOverride{Verbose: true}))
			})
		})
	})
})
