package config_test

import (
	"encoding/json"
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
			Expect(config.PluginRepos()).To(Equal([]PluginRepos{{
				Name: "CF-Community",
				URL:  "https://plugins.cloudfoundry.org",
			}}))
			Expect(config.Experimental()).To(BeFalse())

			pluginConfig := config.Plugins()
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

	Describe("getter functions", func() {
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

		DescribeTable("Experimental",
			func(envVal string, expected bool) {
				rawConfig := fmt.Sprintf(`{}`)
				setConfig(homeDir, rawConfig)

				defer os.Unsetenv("EXPERIMENTAL")
				if envVal == "" {
					os.Unsetenv("EXPERIMENTAL")
				} else {
					os.Setenv("EXPERIMENTAL", envVal)
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

		Describe("CurrentUser", func() {
			Context("when the user token is set", func() {
				It("returns the user", func() {
					config := Config{
						ConfigFile: CFConfig{
							AccessToken: "bearer eyJhbGciOiJSUzI1NiIsImtpZCI6ImxlZ2FjeS10b2tlbi1rZXkiLCJ0eXAiOiJKV1QifQ.eyJqdGkiOiI3YzZkMDA2MjA2OTI0NmViYWI0ZjBmZjY3NGQ3Zjk4OSIsInN1YiI6Ijk1MTliZTNlLTQ0ZDktNDBkMC1hYjlhLWY0YWNlMTFkZjE1OSIsInNjb3BlIjpbIm9wZW5pZCIsInJvdXRpbmcucm91dGVyX2dyb3Vwcy53cml0ZSIsInNjaW0ucmVhZCIsImNsb3VkX2NvbnRyb2xsZXIuYWRtaW4iLCJ1YWEudXNlciIsInJvdXRpbmcucm91dGVyX2dyb3Vwcy5yZWFkIiwiY2xvdWRfY29udHJvbGxlci5yZWFkIiwicGFzc3dvcmQud3JpdGUiLCJjbG91ZF9jb250cm9sbGVyLndyaXRlIiwiZG9wcGxlci5maXJlaG9zZSIsInNjaW0ud3JpdGUiXSwiY2xpZW50X2lkIjoiY2YiLCJjaWQiOiJjZiIsImF6cCI6ImNmIiwiZ3JhbnRfdHlwZSI6InBhc3N3b3JkIiwidXNlcl9pZCI6Ijk1MTliZTNlLTQ0ZDktNDBkMC1hYjlhLWY0YWNlMTFkZjE1OSIsIm9yaWdpbiI6InVhYSIsInVzZXJfbmFtZSI6ImFkbWluIiwiZW1haWwiOiJhZG1pbiIsImF1dGhfdGltZSI6MTQ3MzI4NDU3NywicmV2X3NpZyI6IjZiMjdkYTZjIiwiaWF0IjoxNDczMjg0NTc3LCJleHAiOjE0NzMyODUxNzcsImlzcyI6Imh0dHBzOi8vdWFhLmJvc2gtbGl0ZS5jb20vb2F1dGgvdG9rZW4iLCJ6aWQiOiJ1YWEiLCJhdWQiOlsiY2YiLCJvcGVuaWQiLCJyb3V0aW5nLnJvdXRlcl9ncm91cHMiLCJzY2ltIiwiY2xvdWRfY29udHJvbGxlciIsInVhYSIsInBhc3N3b3JkIiwiZG9wcGxlciJdfQ.OcH_w9yIKJkEcTZMThIs-qJAHk3G0JwNjG-aomVH9hKye4ciFO6IMQMLKmCBrrAQVc7ST1SZZwq7gv12Dq__6Jp-hai0a2_ADJK-Vc9YXyNZKgYTWIeVNGM1JGdHgFSrBR2Lz7IIrH9HqeN8plrKV5HzU8uI9LL4lyOCjbXJ9cM",
						},
					}

					user, err := config.CurrentUser()
					Expect(err).ToNot(HaveOccurred())
					Expect(user).To(Equal(User{
						Name: "admin",
					}))
				})
			})

			Context("when the user token is blank", func() {
				It("returns the user", func() {
					var config Config
					user, err := config.CurrentUser()
					Expect(err).ToNot(HaveOccurred())
					Expect(user).To(Equal(User{}))
				})
			})
		})
	})

	Describe("Write Config", func() {
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
					"wws://loggregator.foo.com:443",
					"wws://doppler.foo.com:443",
					"https://uaa.foo.com",
					"https://api.foo.com/routing",
					true,
				)

				Expect(config.ConfigFile.Target).To(Equal("https://api.foo.com"))
				Expect(config.ConfigFile.APIVersion).To(Equal("2.59.31"))
				Expect(config.ConfigFile.AuthorizationEndpoint).To(Equal("https://login.foo.com"))
				Expect(config.ConfigFile.LoggregatorEndpoint).To(Equal("wws://loggregator.foo.com:443"))
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

		Describe("SetOrganizationInformation", func() {
			It("sets the organization GUID and name", func() {
				config := Config{}
				config.SetOrganizationInformation("guid-value-1", "my-org-name")

				Expect(config.ConfigFile.TargetedOrganization.GUID).To(Equal("guid-value-1"))
				Expect(config.ConfigFile.TargetedOrganization.Name).To(Equal("my-org-name"))
			})
		})

		Describe("SetSpaceInformation", func() {
			It("sets the organization GUID and name", func() {
				config := Config{}
				config.SetSpaceInformation("guid-value-1", "my-org-name", true)

				Expect(config.ConfigFile.TargetedSpace.GUID).To(Equal("guid-value-1"))
				Expect(config.ConfigFile.TargetedSpace.Name).To(Equal("my-org-name"))
				Expect(config.ConfigFile.TargetedSpace.AllowSSH).To(BeTrue())
			})
		})
	})
})
