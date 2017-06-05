package configv3_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	. "code.cloudfoundry.org/cli/util/configv3"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Config", func() {
	var homeDir string

	BeforeEach(func() {
		homeDir = createAndSetHomeDir()
	})

	AfterEach(func() {
		removeAndUnsetHomeDir(homeDir)
	})

	Context("when there isn't a config set", func() {
		It("returns a default config", func() {
			config, err := LoadConfig()

			Expect(err).ToNot(HaveOccurred())

			Expect(config).ToNot(BeNil())
			Expect(config.Target()).To(Equal(DefaultTarget))
			Expect(config.SkipSSLValidation()).To(BeFalse())
			Expect(config.ColorEnabled()).To(Equal(ColorEnabled))
			Expect(config.PluginHome()).To(Equal(filepath.Join(HomeDirectory(false), ".cf", "plugins")))
			Expect(config.StagingTimeout()).To(Equal(DefaultStagingTimeout))
			Expect(config.StartupTimeout()).To(Equal(DefaultStartupTimeout))
			Expect(config.Locale()).To(BeEmpty())
			Expect(config.UAAOAuthClient()).To(Equal(DefaultUAAOAuthClient))
			Expect(config.UAAOAuthClientSecret()).To(Equal(DefaultUAAOAuthClientSecret))
			Expect(config.OverallPollingTimeout()).To(Equal(DefaultOverallPollingTimeout))
			Expect(config.LogLevel()).To(Equal(0))

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

		Context("when UAAOAuthClient is not present", func() {
			BeforeEach(func() {
				rawConfig := `{}`
				writeConfig(homeDir, rawConfig)

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
				writeConfig(homeDir, rawConfig)

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
				writeConfig(homeDir, rawConfig)

				var err error
				config, err = LoadConfig()
				Expect(err).ToNot(HaveOccurred())
				Expect(config).ToNot(BeNil())
			})

			It("returns fields directly from config", func() {
				Expect(config.Target()).To(Equal("https://api.foo.com"))
			})
		})

		Describe("OverallPollingTimeout", func() {
			var config *Config

			Context("when AsyncTimeout is set in config", func() {
				BeforeEach(func() {
					rawConfig := `{ "AsyncTimeout":5 }`
					writeConfig(homeDir, rawConfig)

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
				writeConfig(homeDir, rawConfig)

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
				writeConfig(homeDir, rawConfig)

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
				writeConfig(homeDir, rawConfig)

				var err error
				config, err = LoadConfig()
				Expect(err).ToNot(HaveOccurred())
				Expect(config).ToNot(BeNil())
			})

			It("returns fields directly from config", func() {
				Expect(config.RefreshToken()).To(Equal("some-token"))
			})
		})

		Describe("UAAOAuthClient", func() {
			var config *Config

			BeforeEach(func() {
				rawConfig := `{ "UAAOAuthClient":"some-client" }`
				writeConfig(homeDir, rawConfig)

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
				writeConfig(homeDir, rawConfig)

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
				rawConfig := fmt.Sprintf(`{}`)
				writeConfig(homeDir, rawConfig)

				defer os.Unsetenv("CF_CLI_EXPERIMENTAL")
				if envVal == "" {
					os.Unsetenv("CF_CLI_EXPERIMENTAL")
				} else {
					os.Setenv("CF_CLI_EXPERIMENTAL", envVal)
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

				config *Config
			)

			BeforeEach(func() {
				originalCFStagingTimeout = os.Getenv("CF_STAGING_TIMEOUT")
				originalCFStartupTimeout = os.Getenv("CF_STARTUP_TIMEOUT")
				originalHTTPSProxy = os.Getenv("https_proxy")
				originalForceTTY = os.Getenv("FORCE_TTY")
				os.Setenv("CF_STAGING_TIMEOUT", "8675")
				os.Setenv("CF_STARTUP_TIMEOUT", "309")
				os.Setenv("https_proxy", "proxy.com")
				os.Setenv("FORCE_TTY", "true")

				var err error
				config, err = LoadConfig()
				Expect(err).ToNot(HaveOccurred())
				Expect(config).ToNot(BeNil())
			})

			AfterEach(func() {
				os.Setenv("CF_STAGING_TIMEOUT", originalCFStagingTimeout)
				os.Setenv("CF_STARTUP_TIMEOUT", originalCFStartupTimeout)
				os.Setenv("https_proxy", originalHTTPSProxy)
				os.Setenv("FORCE_TTY", originalForceTTY)
			})

			It("overrides specific config values", func() {
				Expect(config.StagingTimeout()).To(Equal(time.Duration(8675) * time.Minute))
				Expect(config.StartupTimeout()).To(Equal(time.Duration(309) * time.Minute))
				Expect(config.HTTPSProxy()).To(Equal("proxy.com"))
				Expect(config.IsTTY()).To(BeTrue())
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

		DescribeTable("Verbose",
			func(env string, configTrace string, flag bool, expected bool, location []string) {
				rawConfig := fmt.Sprintf(`{ "Trace":"%s" }`, configTrace)
				writeConfig(homeDir, rawConfig)

				defer os.Unsetenv("CF_TRACE")
				if env == "" {
					os.Unsetenv("CF_TRACE")
				} else {
					os.Setenv("CF_TRACE", env)
				}

				config, err := LoadConfig(FlagOverride{
					Verbose: flag,
				})
				Expect(err).ToNot(HaveOccurred())
				Expect(config).ToNot(BeNil())

				verbose, parsedLocation := config.Verbose()
				Expect(verbose).To(Equal(expected))
				Expect(parsedLocation).To(Equal(location))
			},

			Entry("CF_TRACE true: enables verbose", "true", "", false, true, nil),
			Entry("CF_Trace true, config trace false: enables verbose", "true", "false", false, true, nil),
			Entry("CF_Trace true, config trace file path: enables verbose AND logging to file", "true", "/foo/bar", false, true, []string{"/foo/bar"}),

			Entry("CF_TRACE false: disables verbose", "false", "", false, false, nil),
			Entry("CF_TRACE false, '-v': enables verbose", "false", "", true, true, nil),
			Entry("CF_TRACE false, config trace true: disables verbose", "false", "true", false, false, nil),
			Entry("CF_TRACE false, config trace file path: enables logging to file", "false", "/foo/bar", false, false, []string{"/foo/bar"}),
			Entry("CF_TRACE false, config trace file path, '-v': enables verbose AND logging to file", "false", "/foo/bar", true, true, []string{"/foo/bar"}),

			Entry("CF_TRACE empty: disables verbose", "", "", false, false, nil),
			Entry("CF_TRACE empty:, '-v': enables verbose", "", "", true, true, nil),
			Entry("CF_TRACE empty, config trace true: enables verbose", "", "true", false, true, nil),
			Entry("CF_TRACE empty, config trace file path: enables logging to file", "", "/foo/bar", false, false, []string{"/foo/bar"}),
			Entry("CF_TRACE empty, config trace file path, '-v': enables verbose AND logging to file", "", "/foo/bar", true, true, []string{"/foo/bar"}),

			Entry("CF_TRACE filepath: enables logging to file", "/foo/bar", "", false, false, []string{"/foo/bar"}),
			Entry("CF_TRACE filepath, '-v': enables logging to file", "/foo/bar", "", true, true, []string{"/foo/bar"}),
			Entry("CF_TRACE filepath, config trace true: enables verbose AND logging to file", "/foo/bar", "true", false, true, []string{"/foo/bar"}),
			Entry("CF_TRACE filepath, config trace filepath: enables logging to file for BOTH paths", "/foo/bar", "/baz", false, false, []string{"/foo/bar", "/baz"}),
			Entry("CF_TRACE filepath, config trace filepath, '-v': enables verbose AND logging to file for BOTH paths", "/foo/bar", "/baz", true, true, []string{"/foo/bar", "/baz"}),
		)

		Describe("DialTimeout", func() {
			var (
				originalDialTimeout string

				config *Config
			)

			BeforeEach(func() {
				originalDialTimeout = os.Getenv("CF_DIAL_TIMEOUT")
				os.Setenv("CF_DIAL_TIMEOUT", "1234")

				var err error
				config, err = LoadConfig()
				Expect(err).ToNot(HaveOccurred())
				Expect(config).ToNot(BeNil())
			})

			AfterEach(func() {
				os.Setenv("CF_DIAL_TIMEOUT", originalDialTimeout)
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

		Context("when an error is encountered", func() {
			BeforeEach(func() {
				err := WriteConfig(config)
				Expect(err).ToNot(HaveOccurred())

				configFilePath := filepath.Join(homeDir, ".cf", "config.json")
				err = os.Remove(configFilePath)
				Expect(err).ToNot(HaveOccurred())
				err = os.Mkdir(configFilePath, 0700)
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns the error", func() {
				err := WriteConfig(config)
				_, ok := err.(*os.PathError)
				Expect(ok).To(BeTrue())
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
					"https://uaa.foo.com",
					"https://api.foo.com/routing",
					true,
				)

				Expect(config.ConfigFile.Target).To(Equal("https://api.foo.com"))
				Expect(config.ConfigFile.APIVersion).To(Equal("2.59.31"))
				Expect(config.ConfigFile.AuthorizationEndpoint).To(Equal("https://login.foo.com"))
				Expect(config.ConfigFile.MinCLIVersion).To(Equal("2.0.0"))
				Expect(config.ConfigFile.DopplerEndpoint).To(Equal("wws://doppler.foo.com:443"))
				Expect(config.ConfigFile.UAAEndpoint).To(Equal("https://uaa.foo.com"))
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
