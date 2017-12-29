package configv3_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"code.cloudfoundry.org/cli/command/translatableerror"
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
				Expect(config.ColorEnabled()).To(Equal(ColorEnabled))
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
					Expect(config.ColorEnabled()).To(Equal(ColorEnabled))
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
					Expect(config.ColorEnabled()).To(Equal(ColorEnabled))
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

	Describe("check functions", func() {
		Describe("HasTargetedOrganization", func() {
			Context("when an organization is targeted", func() {
				It("returns true", func() {
					config := Config{}
					config.SetOrganizationInformation("guid-value-1", "my-org-name")
					Expect(config.HasTargetedOrganization()).To(BeTrue())
				})
			})

			Context("when an organization is not targeted", func() {
				It("returns false", func() {
					config := Config{}
					Expect(config.HasTargetedOrganization()).To(BeFalse())
				})
			})
		})

		Describe("HasTargetedSpace", func() {
			Context("when an space is targeted", func() {
				It("returns true", func() {
					config := Config{}
					config.SetSpaceInformation("guid-value-1", "my-org-name", true)
					Expect(config.HasTargetedSpace()).To(BeTrue())
				})
			})

			Context("when an space is not targeted", func() {
				It("returns false", func() {
					config := Config{}
					Expect(config.HasTargetedSpace()).To(BeFalse())
				})
			})
		})
	})

	Describe("getter functions", func() {
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

			It("returns the target", func() {
				Expect(config.Target()).To(Equal("https://api.foo.com"))
			})
		})

		Describe("OverallPollingTimeout", func() {
			var config *Config

			Context("when AsyncTimeout is set in config", func() {
				BeforeEach(func() {
					rawConfig := `{ "AsyncTimeout":5 }`
					setConfig(homeDir, rawConfig)

					var err error
					config, err = LoadConfig()
					Expect(err).ToNot(HaveOccurred())
					Expect(config).ToNot(BeNil())
				})

				It("returns the timeout in duration form", func() {
					Expect(config.OverallPollingTimeout()).To(Equal(5 * time.Minute))
				})
			})
		})

		Describe("SkipSSLValidation", func() {
			var config *Config

			BeforeEach(func() {
				rawConfig := `{ "SSLDisabled":true }`
				setConfig(homeDir, rawConfig)

				var err error
				config, err = LoadConfig()
				Expect(err).ToNot(HaveOccurred())
				Expect(config).ToNot(BeNil())
			})

			It("returns fields directly from config", func() {
				Expect(config.SkipSSLValidation()).To(BeTrue())
			})
		})

		Describe("AccessToken", func() {
			var config *Config

			BeforeEach(func() {
				rawConfig := `{ "AccessToken":"some-token" }`
				setConfig(homeDir, rawConfig)

				var err error
				config, err = LoadConfig()
				Expect(err).ToNot(HaveOccurred())
				Expect(config).ToNot(BeNil())
			})

			It("returns fields directly from config", func() {
				Expect(config.AccessToken()).To(Equal("some-token"))
			})
		})

		Describe("RefreshToken", func() {
			var config *Config

			BeforeEach(func() {
				rawConfig := `{ "RefreshToken":"some-token" }`
				setConfig(homeDir, rawConfig)

				var err error
				config, err = LoadConfig()
				Expect(err).ToNot(HaveOccurred())
				Expect(config).ToNot(BeNil())
			})

			It("returns fields directly from config", func() {
				Expect(config.RefreshToken()).To(Equal("some-token"))
			})
		})

		Describe("SSHOAuthClient", func() {
			var config *Config

			BeforeEach(func() {
				rawConfig := `{ "SSHOAuthClient":"some-ssh-client" }`
				setConfig(homeDir, rawConfig)

				var err error
				config, err = LoadConfig()
				Expect(err).ToNot(HaveOccurred())
				Expect(config).ToNot(BeNil())
			})

			It("returns the client ID", func() {
				Expect(config.SSHOAuthClient()).To(Equal("some-ssh-client"))
			})
		})

		Describe("UAAOAuthClient", func() {
			var config *Config

			BeforeEach(func() {
				rawConfig := `{ "UAAOAuthClient":"some-client" }`
				setConfig(homeDir, rawConfig)

				var err error
				config, err = LoadConfig()
				Expect(err).ToNot(HaveOccurred())
				Expect(config).ToNot(BeNil())
			})

			It("returns the client ID", func() {
				Expect(config.UAAOAuthClient()).To(Equal("some-client"))
			})
		})

		Describe("UAAOAuthClientSecret", func() {
			var config *Config

			BeforeEach(func() {
				rawConfig := `
					{
						"UAAOAuthClient": "some-client-id",
						"UAAOAuthClientSecret": "some-client-secret"
					}`
				setConfig(homeDir, rawConfig)

				var err error
				config, err = LoadConfig()
				Expect(err).ToNot(HaveOccurred())
				Expect(config).ToNot(BeNil())
			})

			It("returns the client secret", func() {
				Expect(config.UAAOAuthClientSecret()).To(Equal("some-client-secret"))
			})
		})

		DescribeTable("Experimental",
			func(envVal string, expected bool) {
				setConfig(homeDir, `{}`)

				defer os.Unsetenv("CF_CLI_EXPERIMENTAL")
				if envVal == "" {
					Expect(os.Unsetenv("CF_CLI_EXPERIMENTAL")).ToNot(HaveOccurred())
				} else {
					Expect(os.Setenv("CF_CLI_EXPERIMENTAL", envVal)).ToNot(HaveOccurred())
				}

				config, err := LoadConfig()
				Expect(err).ToNot(HaveOccurred())
				Expect(config).ToNot(BeNil())

				Expect(config.Experimental()).To(Equal(expected))
			},

			Entry("uses default value of false if environment value is not set", "", false),
			Entry("uses environment value if a valid environment value is set", "true", true),
			Entry("uses default value of false if an invalid environment value is set", "something-invalid", false),
		)

		Describe("BinaryName", func() {
			It("returns the name used to invoke", func() {
				config, err := LoadConfig()
				Expect(err).ToNot(HaveOccurred())
				Expect(config).ToNot(BeNil())

				// Ginkgo will uses a config file as the first test argument, so that
				// will be considered the binary name
				Expect(config.BinaryName()).To(Equal("configv3.test"))
			})
		})

		Context("when there are environment variables", func() {
			var (
				originalCFStagingTimeout string
				originalCFStartupTimeout string
				originalHTTPSProxy       string
				originalForceTTY         string
				originalDockerPassword   string

				config *Config
			)

			BeforeEach(func() {
				originalCFStagingTimeout = os.Getenv("CF_STAGING_TIMEOUT")
				originalCFStartupTimeout = os.Getenv("CF_STARTUP_TIMEOUT")
				originalHTTPSProxy = os.Getenv("https_proxy")
				originalForceTTY = os.Getenv("FORCE_TTY")
				originalDockerPassword = os.Getenv("CF_DOCKER_PASSWORD")
				Expect(os.Setenv("CF_STAGING_TIMEOUT", "8675")).ToNot(HaveOccurred())
				Expect(os.Setenv("CF_STARTUP_TIMEOUT", "309")).ToNot(HaveOccurred())
				Expect(os.Setenv("https_proxy", "proxy.com")).ToNot(HaveOccurred())
				Expect(os.Setenv("FORCE_TTY", "true")).ToNot(HaveOccurred())
				Expect(os.Setenv("CF_DOCKER_PASSWORD", "banana")).ToNot(HaveOccurred())

				var err error
				config, err = LoadConfig()
				Expect(err).ToNot(HaveOccurred())
				Expect(config).ToNot(BeNil())
			})

			AfterEach(func() {
				Expect(os.Setenv("CF_STAGING_TIMEOUT", originalCFStagingTimeout)).ToNot(HaveOccurred())
				Expect(os.Setenv("CF_STARTUP_TIMEOUT", originalCFStartupTimeout)).ToNot(HaveOccurred())
				Expect(os.Setenv("https_proxy", originalHTTPSProxy)).ToNot(HaveOccurred())
				Expect(os.Setenv("FORCE_TTY", originalForceTTY)).ToNot(HaveOccurred())
				Expect(os.Setenv("CF_DOCKER_PASSWORD", originalDockerPassword)).ToNot(HaveOccurred())
			})

			It("overrides specific config values", func() {
				Expect(config.StagingTimeout()).To(Equal(time.Duration(8675) * time.Minute))
				Expect(config.StartupTimeout()).To(Equal(time.Duration(309) * time.Minute))
				Expect(config.HTTPSProxy()).To(Equal("proxy.com"))
				Expect(config.IsTTY()).To(BeTrue())
				Expect(config.DockerPassword()).To(Equal("banana"))
			})
		})

		Describe("APIVersion", func() {
			It("returns the api version", func() {
				config := Config{
					ConfigFile: CFConfig{
						APIVersion: "2.59.0",
					},
				}

				Expect(config.APIVersion()).To(Equal("2.59.0"))
			})
		})

		Describe("MinCLIVersion", func() {
			It("returns the minimum CLI version the CC requires", func() {
				config := Config{
					ConfigFile: CFConfig{
						MinCLIVersion: "1.0.0",
					},
				}

				Expect(config.MinCLIVersion()).To(Equal("1.0.0"))
			})
		})

		Describe("TargetedOrganization", func() {
			It("returns the organization", func() {
				organization := Organization{
					GUID: "some-guid",
					Name: "some-org",
				}
				config := Config{
					ConfigFile: CFConfig{
						TargetedOrganization: organization,
					},
				}

				Expect(config.TargetedOrganization()).To(Equal(organization))
			})
		})

		Describe("TargetedSpace", func() {
			It("returns the space", func() {
				space := Space{
					GUID: "some-guid",
					Name: "some-space",
				}
				config := Config{
					ConfigFile: CFConfig{
						TargetedSpace: space,
					},
				}

				Expect(config.TargetedSpace()).To(Equal(space))
			})
		})

		Describe("DialTimeout", func() {
			var (
				originalDialTimeout string

				config *Config
			)

			BeforeEach(func() {
				originalDialTimeout = os.Getenv("CF_DIAL_TIMEOUT")
				Expect(os.Setenv("CF_DIAL_TIMEOUT", "1234")).ToNot(HaveOccurred())

				var err error
				config, err = LoadConfig()
				Expect(err).ToNot(HaveOccurred())
				Expect(config).ToNot(BeNil())
			})

			AfterEach(func() {
				Expect(os.Setenv("CF_DIAL_TIMEOUT", originalDialTimeout)).ToNot(HaveOccurred())
			})

			It("returns the dial timeout", func() {
				Expect(config.DialTimeout()).To(Equal(1234 * time.Second))
			})
		})

		Describe("BinaryVersion", func() {
			It("returns back version.BinaryVersion", func() {
				conf := Config{}
				Expect(conf.BinaryVersion()).To(Equal("0.0.0-unknown-version"))
			})
		})

		DescribeTable("LogLevel",
			func(envVal string, expectedLevel int) {
				config := Config{ENV: EnvOverride{CFLogLevel: envVal}}
				Expect(config.LogLevel()).To(Equal(expectedLevel))
			},

			Entry("Default to 0", "", 0),
			Entry("panic returns 0", "panic", 0),
			Entry("fatal returns 1", "fatal", 1),
			Entry("error returns 2", "error", 2),
			Entry("warn returns 3", "warn", 3),
			Entry("info returns 4", "info", 4),
			Entry("debug returns 5", "debug", 5),
			Entry("dEbUg returns 5", "dEbUg", 5),
		)

		Describe("RequestRetryCount", func() {
			It("returns the number of request retries", func() {
				conf := Config{}
				Expect(conf.RequestRetryCount()).To(Equal(2))
			})
		})

		Describe("NOAARequestRetryCount", func() {
			It("returns the number of request retries", func() {
				conf := Config{}
				Expect(conf.NOAARequestRetryCount()).To(Equal(5))
			})
		})
	})

	Describe("WriteConfig", func() {
		var config *Config

		BeforeEach(func() {
			config = &Config{
				ConfigFile: CFConfig{
					ConfigVersion: 3,
					Target:        "foo.com",
					ColorEnabled:  "true",
				},
				ENV: EnvOverride{
					CFColor: "false",
				},
			}
		})

		Context("when no errors are encountered", func() {
			It("writes ConfigFile to homeDir/.cf/config.json", func() {
				err := WriteConfig(config)
				Expect(err).ToNot(HaveOccurred())

				file, err := ioutil.ReadFile(filepath.Join(homeDir, ".cf", "config.json"))
				Expect(err).ToNot(HaveOccurred())

				var writtenCFConfig CFConfig
				err = json.Unmarshal(file, &writtenCFConfig)
				Expect(err).ToNot(HaveOccurred())

				Expect(writtenCFConfig.ConfigVersion).To(Equal(config.ConfigFile.ConfigVersion))
				Expect(writtenCFConfig.Target).To(Equal(config.ConfigFile.Target))
				Expect(writtenCFConfig.ColorEnabled).To(Equal(config.ConfigFile.ColorEnabled))
			})
		})
	})

	Describe("setter functions", func() {
		Describe("SetTargetInformation", func() {
			It("sets the api target and other related endpoints", func() {
				config := Config{
					ConfigFile: CFConfig{
						TargetedOrganization: Organization{
							GUID: "this-is-a-guid",
							Name: "jo bobo jim boo",
						},
						TargetedSpace: Space{
							GUID:     "this-is-a-guid",
							Name:     "jo bobo jim boo",
							AllowSSH: true,
						},
					},
				}
				config.SetTargetInformation(
					"https://api.foo.com",
					"2.59.31",
					"https://login.foo.com",
					"2.0.0",
					"wws://doppler.foo.com:443",
					"https://api.foo.com/routing",
					true,
				)

				Expect(config.ConfigFile.Target).To(Equal("https://api.foo.com"))
				Expect(config.ConfigFile.APIVersion).To(Equal("2.59.31"))
				Expect(config.ConfigFile.AuthorizationEndpoint).To(Equal("https://login.foo.com"))
				Expect(config.ConfigFile.MinCLIVersion).To(Equal("2.0.0"))
				Expect(config.ConfigFile.DopplerEndpoint).To(Equal("wws://doppler.foo.com:443"))
				Expect(config.ConfigFile.RoutingEndpoint).To(Equal("https://api.foo.com/routing"))
				Expect(config.ConfigFile.SkipSSLValidation).To(BeTrue())

				Expect(config.ConfigFile.TargetedOrganization.GUID).To(BeEmpty())
				Expect(config.ConfigFile.TargetedOrganization.Name).To(BeEmpty())
				Expect(config.ConfigFile.TargetedSpace.GUID).To(BeEmpty())
				Expect(config.ConfigFile.TargetedSpace.Name).To(BeEmpty())
				Expect(config.ConfigFile.TargetedSpace.AllowSSH).To(BeFalse())
			})
		})

		Describe("SetTokenInformation", func() {
			It("sets the authentication token information", func() {
				var config Config
				config.SetTokenInformation("I am the access token", "I am the refresh token", "I am the SSH OAuth client")

				Expect(config.ConfigFile.AccessToken).To(Equal("I am the access token"))
				Expect(config.ConfigFile.RefreshToken).To(Equal("I am the refresh token"))
				Expect(config.ConfigFile.SSHOAuthClient).To(Equal("I am the SSH OAuth client"))
			})
		})

		Describe("SetAccessToken", func() {
			It("sets the authentication token information", func() {
				var config Config
				config.SetAccessToken("I am the access token")
				Expect(config.ConfigFile.AccessToken).To(Equal("I am the access token"))
			})
		})

		Describe("SetRefreshToken", func() {
			It("sets the refresh token information", func() {
				var config Config
				config.SetRefreshToken("I am the refresh token")
				Expect(config.ConfigFile.RefreshToken).To(Equal("I am the refresh token"))
			})
		})

		Describe("SetOrganizationInformation", func() {
			It("sets the organization GUID and name", func() {
				config := Config{}
				config.SetOrganizationInformation("guid-value-1", "my-org-name")

				Expect(config.ConfigFile.TargetedOrganization.GUID).To(Equal("guid-value-1"))
				Expect(config.ConfigFile.TargetedOrganization.Name).To(Equal("my-org-name"))
			})
		})

		Describe("SetSpaceInformation", func() {
			It("sets the space GUID, name, and AllowSSH", func() {
				config := Config{}
				config.SetSpaceInformation("guid-value-1", "my-org-name", true)

				Expect(config.ConfigFile.TargetedSpace.GUID).To(Equal("guid-value-1"))
				Expect(config.ConfigFile.TargetedSpace.Name).To(Equal("my-org-name"))
				Expect(config.ConfigFile.TargetedSpace.AllowSSH).To(BeTrue())
			})
		})

		Describe("SetUAAEndpoint", func() {
			It("sets the UAA endpoint", func() {
				var config Config
				config.SetUAAEndpoint("some-uaa-endpoint.com")
				Expect(config.ConfigFile.UAAEndpoint).To(Equal("some-uaa-endpoint.com"))
			})
		})

		Describe("UnsetOrganizationInformation", func() {
			config := Config{}
			BeforeEach(func() {
				config.SetOrganizationInformation("some-org-guid", "some-org")
			})

			It("resets the org GUID and name", func() {
				config.UnsetOrganizationInformation()

				Expect(config.ConfigFile.TargetedOrganization.GUID).To(Equal(""))
				Expect(config.ConfigFile.TargetedOrganization.Name).To(Equal(""))
			})
		})

		Describe("UnsetSpaceInformation", func() {
			config := Config{}
			BeforeEach(func() {
				config.SetSpaceInformation("guid-value-1", "my-org-name", true)
			})

			It("resets the space GUID, name, and AllowSSH to default values", func() {
				config.UnsetSpaceInformation()

				Expect(config.ConfigFile.TargetedSpace.GUID).To(Equal(""))
				Expect(config.ConfigFile.TargetedSpace.Name).To(Equal(""))
				Expect(config.ConfigFile.TargetedSpace.AllowSSH).To(BeFalse())
			})
		})
	})
})
