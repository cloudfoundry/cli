package configv3_test

import (
	"fmt"
	"time"

	. "code.cloudfoundry.org/cli/util/configv3"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("JSONConfig", func() {
	var homeDir string
	var config *Config

	BeforeEach(func() {
		homeDir = setup()
	})

	AfterEach(func() {
		teardown(homeDir)
	})

	Describe("AccessToken", func() {
		BeforeEach(func() {
			rawConfig := fmt.Sprintf(`{ "AccessToken":"some-token", "ConfigVersion": %d }`, CurrentConfigVersion)
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

	Describe("APIVersion", func() {
		It("returns the api version", func() {
			config = &Config{
				ConfigFile: JSONConfig{
					APIVersion: "2.59.0",
				},
			}

			Expect(config.APIVersion()).To(Equal("2.59.0"))
		})
	})

	Describe("CurrentUser", func() {
		When("using client credentials and the user token is set", func() {
			It("returns the user", func() {
				config = &Config{
					ConfigFile: JSONConfig{
						AccessToken: AccessTokenForClientUsers,
					},
				}
				user, err := config.CurrentUser()
				Expect(err).ToNot(HaveOccurred())
				Expect(user).To(Equal(User{
					Name:     "potato-face",
					GUID:     "potato-face",
					IsClient: true,
				}))
			})
		})

		When("using user/password and the user token is set", func() {
			It("returns the user", func() {
				config = &Config{
					ConfigFile: JSONConfig{
						AccessToken: AccessTokenForHumanUsers,
					},
				}

				user, err := config.CurrentUser()
				Expect(err).ToNot(HaveOccurred())
				Expect(user).To(Equal(User{
					Name:     "admin",
					GUID:     "9519be3e-44d9-40d0-ab9a-f4ace11df159",
					Origin:   "uaa",
					IsClient: false,
				}))
			})
		})

		When("the user token is blank", func() {
			It("returns the user", func() {
				config = new(Config)
				user, err := config.CurrentUser()
				Expect(err).ToNot(HaveOccurred())
				Expect(user).To(Equal(User{}))
			})
		})
	})

	Describe("CurrentUserName", func() {
		When("using client credentials and the user token is set", func() {
			It("returns the username", func() {
				config = &Config{
					ConfigFile: JSONConfig{
						AccessToken: AccessTokenForClientUsers,
					},
				}

				username, err := config.CurrentUserName()
				Expect(err).ToNot(HaveOccurred())
				Expect(username).To(Equal("potato-face"))
			})
		})

		When("using user/password and the user token is set", func() {
			It("returns the username", func() {
				config = &Config{
					ConfigFile: JSONConfig{
						AccessToken: AccessTokenForHumanUsers,
					},
				}

				username, err := config.CurrentUserName()
				Expect(err).ToNot(HaveOccurred())
				Expect(username).To(Equal("admin"))
			})
		})

		When("the user token is blank", func() {
			It("returns an empty string", func() {
				config = new(Config)
				username, err := config.CurrentUserName()
				Expect(err).ToNot(HaveOccurred())
				Expect(username).To(BeEmpty())
			})
		})
	})

	Describe("HasTargetedOrganization", func() {
		When("an organization is targeted", func() {
			It("returns true", func() {
				config = new(Config)
				config.SetOrganizationInformation("guid-value-1", "my-org-name")
				Expect(config.HasTargetedOrganization()).To(BeTrue())
			})
		})

		When("an organization is not targeted", func() {
			It("returns false", func() {
				config = new(Config)
				Expect(config.HasTargetedOrganization()).To(BeFalse())
			})
		})
	})

	Describe("HasTargetedSpace", func() {
		When("an space is targeted", func() {
			It("returns true", func() {
				config = new(Config)
				config.SetSpaceInformation("guid-value-1", "my-org-name", true)
				Expect(config.HasTargetedSpace()).To(BeTrue())
			})
		})

		When("an space is not targeted", func() {
			It("returns false", func() {
				config = new(Config)
				Expect(config.HasTargetedSpace()).To(BeFalse())
			})
		})
	})

	Describe("MinCLIVersion", func() {
		It("returns the minimum CLI version the CC requires", func() {
			config = &Config{
				ConfigFile: JSONConfig{
					MinCLIVersion: "1.0.0",
				},
			}

			Expect(config.MinCLIVersion()).To(Equal("1.0.0"))
		})
	})

	Describe("OverallPollingTimeout", func() {
		When("AsyncTimeout is set in config", func() {
			BeforeEach(func() {
				rawConfig := fmt.Sprintf(`{ "AsyncTimeout":5, "ConfigVersion": %d }`, CurrentConfigVersion)
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

	Describe("RefreshToken", func() {
		BeforeEach(func() {
			rawConfig := fmt.Sprintf(`{ "RefreshToken":"some-token", "ConfigVersion": %d }`, CurrentConfigVersion)
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

	Describe("SetAsyncTimeout", func() {
		It("sets the async timeout", func() {
			config = new(Config)
			config.SetAsyncTimeout(2)
			Expect(config.ConfigFile.AsyncTimeout).To(Equal(2))
		})
	})

	Describe("SetColorEnabled", func() {
		It("sets the color enabled field", func() {
			config = new(Config)
			config.SetColorEnabled("true")
			Expect(config.ConfigFile.ColorEnabled).To(Equal("true"))
		})
	})

	Describe("SetAccessToken", func() {
		It("sets the authentication token information", func() {
			config = new(Config)
			config.SetAccessToken("I am the access token")
			Expect(config.ConfigFile.AccessToken).To(Equal("I am the access token"))
		})
	})

	Describe("SetLocale", func() {
		It("sets the locale field", func() {
			config = new(Config)
			config.SetLocale("en-US")
			Expect(config.ConfigFile.Locale).To(Equal("en-US"))
		})

		It("clears the locale field if requested", func() {
			config = new(Config)
			config.SetLocale("CLEAR")
			Expect(config.ConfigFile.Locale).To(Equal(""))
		})
	})

	Describe("SetOrganizationInformation", func() {
		It("sets the organization GUID and name", func() {
			config = new(Config)
			config.SetOrganizationInformation("guid-value-1", "my-org-name")

			Expect(config.ConfigFile.TargetedOrganization.GUID).To(Equal("guid-value-1"))
			Expect(config.ConfigFile.TargetedOrganization.Name).To(Equal("my-org-name"))
		})
	})

	Describe("SetRefreshToken", func() {
		It("sets the refresh token information", func() {
			config = new(Config)
			config.SetRefreshToken("I am the refresh token")
			Expect(config.ConfigFile.RefreshToken).To(Equal("I am the refresh token"))
		})
	})

	Describe("SetSpaceInformation", func() {
		It("sets the space GUID, name, and AllowSSH", func() {
			config = new(Config)
			config.SetSpaceInformation("guid-value-1", "my-org-name", true)

			Expect(config.ConfigFile.TargetedSpace.GUID).To(Equal("guid-value-1"))
			Expect(config.ConfigFile.TargetedSpace.Name).To(Equal("my-org-name"))
			Expect(config.ConfigFile.TargetedSpace.AllowSSH).To(BeTrue())
		})
	})

	Describe("SetTargetInformation", func() {
		It("sets the api target and other related endpoints", func() {
			config = &Config{
				ConfigFile: JSONConfig{
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
			config = new(Config)
			config.SetTokenInformation("I am the access token", "I am the refresh token", "I am the SSH OAuth client")

			Expect(config.ConfigFile.AccessToken).To(Equal("I am the access token"))
			Expect(config.ConfigFile.RefreshToken).To(Equal("I am the refresh token"))
			Expect(config.ConfigFile.SSHOAuthClient).To(Equal("I am the SSH OAuth client"))
		})
	})

	Describe("SetTrace", func() {
		It("sets the trace field", func() {
			config = new(Config)
			config.SetTrace("true")
			Expect(config.ConfigFile.Trace).To(Equal("true"))
		})
	})

	Describe("SetUAAClientCredentials", func() {
		It("sets the UAA client credentials", func() {
			config = new(Config)
			config.SetUAAClientCredentials("some-uaa-client", "some-uaa-client-secret")
			Expect(config.ConfigFile.UAAOAuthClient).To(Equal("some-uaa-client"))
			Expect(config.ConfigFile.UAAOAuthClientSecret).To(Equal("some-uaa-client-secret"))
		})
	})

	Describe("SetUAAEndpoint", func() {
		It("sets the UAA endpoint", func() {
			config = new(Config)
			config.SetUAAEndpoint("some-uaa-endpoint.com")
			Expect(config.ConfigFile.UAAEndpoint).To(Equal("some-uaa-endpoint.com"))
		})
	})

	Describe("SetUAAGrantType", func() {
		It("sets the UAA endpoint", func() {
			config = new(Config)
			config.SetUAAGrantType("some-uaa-grant-type")
			Expect(config.ConfigFile.UAAGrantType).To(Equal("some-uaa-grant-type"))
		})
	})

	Describe("SkipSSLValidation", func() {
		BeforeEach(func() {
			rawConfig := fmt.Sprintf(`{ "SSLDisabled":true, "ConfigVersion": %d }`, CurrentConfigVersion)
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

	Describe("SSHOAuthClient", func() {
		BeforeEach(func() {
			rawConfig := fmt.Sprintf(`{ "SSHOAuthClient":"some-ssh-client", "ConfigVersion": %d }`, CurrentConfigVersion)
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

	Describe("Target", func() {
		BeforeEach(func() {
			rawConfig := fmt.Sprintf(`{ "Target":"https://api.foo.com", "ConfigVersion": %d }`, CurrentConfigVersion)
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

	Describe("TargetedOrganization", func() {
		It("returns the organization", func() {
			organization := Organization{
				GUID: "some-guid",
				Name: "some-org",
			}
			config = &Config{
				ConfigFile: JSONConfig{
					TargetedOrganization: organization,
				},
			}

			Expect(config.TargetedOrganization()).To(Equal(organization))
		})
	})

	Describe("TargetedOrganizationName", func() {
		It("returns the name of targeted organization", func() {
			organization := Organization{
				GUID: "some-guid",
				Name: "some-org",
			}
			config = &Config{
				ConfigFile: JSONConfig{
					TargetedOrganization: organization,
				},
			}

			Expect(config.TargetedOrganizationName()).To(Equal(organization.Name))
		})
	})

	Describe("TargetedSpace", func() {
		It("returns the space", func() {
			space := Space{
				GUID: "some-guid",
				Name: "some-space",
			}
			config = &Config{
				ConfigFile: JSONConfig{
					TargetedSpace: space,
				},
			}

			Expect(config.TargetedSpace()).To(Equal(space))
		})
	})

	Describe("UAAGrantType", func() {
		BeforeEach(func() {
			rawConfig := fmt.Sprintf(`{ "UAAGrantType":"some-grant-type", "ConfigVersion": %d }`, CurrentConfigVersion)
			setConfig(homeDir, rawConfig)

			var err error
			config, err = LoadConfig()
			Expect(err).ToNot(HaveOccurred())
			Expect(config).ToNot(BeNil())
		})

		It("returns the client secret", func() {
			Expect(config.UAAGrantType()).To(Equal("some-grant-type"))
		})
	})

	Describe("UAAOAuthClient", func() {
		BeforeEach(func() {
			rawConfig := fmt.Sprintf(`{ "UAAOAuthClient":"some-client", "ConfigVersion": %d }`, CurrentConfigVersion)
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
		BeforeEach(func() {
			rawConfig := fmt.Sprintf(`
					{
						"UAAOAuthClient": "some-client-id",
						"UAAOAuthClientSecret": "some-client-secret",
						"ConfigVersion": %d
					}`, CurrentConfigVersion)
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

	Describe("UnsetOrganizationAndSpaceInformation", func() {
		BeforeEach(func() {
			config = new(Config)
			config.SetOrganizationInformation("some-org-guid", "some-org")
			config.SetSpaceInformation("guid-value-1", "my-org-name", true)
		})

		It("resets the org GUID and name", func() {
			config.UnsetOrganizationAndSpaceInformation()

			Expect(config.ConfigFile.TargetedOrganization.GUID).To(BeEmpty())
			Expect(config.ConfigFile.TargetedOrganization.Name).To(BeEmpty())
			Expect(config.ConfigFile.TargetedSpace.GUID).To(BeEmpty())
			Expect(config.ConfigFile.TargetedSpace.Name).To(BeEmpty())
			Expect(config.ConfigFile.TargetedSpace.AllowSSH).To(BeFalse())
		})
	})

	Describe("UnsetSpaceInformation", func() {
		BeforeEach(func() {
			config = new(Config)
			config.SetSpaceInformation("guid-value-1", "my-org-name", true)
		})

		It("resets the space GUID, name, and AllowSSH to default values", func() {
			config.UnsetSpaceInformation()

			Expect(config.ConfigFile.TargetedSpace.GUID).To(BeEmpty())
			Expect(config.ConfigFile.TargetedSpace.Name).To(BeEmpty())
			Expect(config.ConfigFile.TargetedSpace.AllowSSH).To(BeFalse())
		})
	})

	Describe("UnsetUserInformation", func() {
		BeforeEach(func() {
			config = new(Config)
			config.SetAccessToken("some-access-token")
			config.SetRefreshToken("some-refresh-token")
			config.SetUAAGrantType("client-credentials")
			config.SetUAAClientCredentials("some-client", "some-client-secret")
			config.SetOrganizationInformation("some-org-guid", "some-org")
			config.SetSpaceInformation("guid-value-1", "my-org-name", true)
		})

		It("resets all user information", func() {
			config.UnsetUserInformation()

			Expect(config.ConfigFile.AccessToken).To(BeEmpty())
			Expect(config.ConfigFile.RefreshToken).To(BeEmpty())
			Expect(config.ConfigFile.TargetedOrganization.GUID).To(BeEmpty())
			Expect(config.ConfigFile.TargetedOrganization.Name).To(BeEmpty())
			Expect(config.ConfigFile.TargetedSpace.AllowSSH).To(BeFalse())
			Expect(config.ConfigFile.TargetedSpace.GUID).To(BeEmpty())
			Expect(config.ConfigFile.TargetedSpace.Name).To(BeEmpty())
			Expect(config.ConfigFile.UAAGrantType).To(BeEmpty())
			Expect(config.ConfigFile.UAAOAuthClient).To(Equal(DefaultUAAOAuthClient))
			Expect(config.ConfigFile.UAAOAuthClientSecret).To(Equal(DefaultUAAOAuthClientSecret))
		})
	})
})
