package coreconfig_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cli/cf/configuration"
	"code.cloudfoundry.org/cli/cf/configuration/configurationfakes"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/models"
	"github.com/blang/semver"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Configuration Repository", func() {
	var (
		config    coreconfig.Repository
		persistor *configurationfakes.FakePersistor
	)

	BeforeEach(func() {
		persistor = new(configurationfakes.FakePersistor)
		persistor.ExistsReturns(true)
		config = coreconfig.NewRepositoryFromPersistor(persistor, func(err error) { panic(err) })
	})

	It("is threadsafe", func() {
		performSaveCh := make(chan struct{})
		beginSaveCh := make(chan struct{})
		finishSaveCh := make(chan struct{})
		finishReadCh := make(chan struct{})

		persistor.SaveStub = func(configuration.DataInterface) error {
			close(beginSaveCh)
			<-performSaveCh
			close(finishSaveCh)

			return nil
		}

		go func() {
			config.SetAPIEndpoint("foo")
		}()

		<-beginSaveCh

		go func() {
			config.APIEndpoint()
			close(finishReadCh)
		}()

		Consistently(finishSaveCh).ShouldNot(BeClosed())
		Consistently(finishReadCh).ShouldNot(BeClosed())

		performSaveCh <- struct{}{}

		Eventually(finishReadCh).Should(BeClosed())
	})

	Context("when the doppler endpoint does not exist", func() {
		It("should regex the loggregator endpoint value", func() {
			config.SetLoggregatorEndpoint("http://loggregator.the-endpoint")
			Expect(config.DopplerEndpoint()).To(Equal("http://doppler.the-endpoint"))
		})
	})

	Context("when the doppler endpoint does not exist", func() {
		It("should regex the loggregator endpoint value", func() {
			config.SetLoggregatorEndpoint("http://loggregator.the-endpointffff")
			config.SetDopplerEndpoint("http://doppler.the-endpoint")
			Expect(config.DopplerEndpoint()).To(Equal("http://doppler.the-endpoint"))
		})
	})

	It("has acccessor methods for all config fields", func() {
		config.SetAPIEndpoint("http://api.the-endpoint")
		Expect(config.APIEndpoint()).To(Equal("http://api.the-endpoint"))

		config.SetAPIVersion("3")
		Expect(config.APIVersion()).To(Equal("3"))

		config.SetAuthenticationEndpoint("http://auth.the-endpoint")
		Expect(config.AuthenticationEndpoint()).To(Equal("http://auth.the-endpoint"))

		config.SetLoggregatorEndpoint("http://loggregator.the-endpoint")
		Expect(config.LoggregatorEndpoint()).To(Equal("http://loggregator.the-endpoint"))

		config.SetUaaEndpoint("http://uaa.the-endpoint")
		Expect(config.UaaEndpoint()).To(Equal("http://uaa.the-endpoint"))

		config.SetAccessToken("the-token")
		Expect(config.AccessToken()).To(Equal("the-token"))

		config.SetUAAOAuthClient("cf-oauth-client-id")
		Expect(config.UAAOAuthClient()).To(Equal("cf-oauth-client-id"))

		config.SetUAAOAuthClientSecret("cf-oauth-client-secret")
		Expect(config.UAAOAuthClientSecret()).To(Equal("cf-oauth-client-secret"))

		config.SetSSHOAuthClient("oauth-client-id")
		Expect(config.SSHOAuthClient()).To(Equal("oauth-client-id"))

		config.SetRefreshToken("the-token")
		Expect(config.RefreshToken()).To(Equal("the-token"))

		organization := models.OrganizationFields{Name: "the-org"}
		config.SetOrganizationFields(organization)
		Expect(config.OrganizationFields()).To(Equal(organization))

		space := models.SpaceFields{Name: "the-space"}
		config.SetSpaceFields(space)
		Expect(config.SpaceFields()).To(Equal(space))

		config.SetSSLDisabled(false)
		Expect(config.IsSSLDisabled()).To(BeFalse())

		config.SetLocale("en_US")
		Expect(config.Locale()).To(Equal("en_US"))

		config.SetPluginRepo(models.PluginRepo{Name: "repo", URL: "nowhere.com"})
		Expect(config.PluginRepos()[0].Name).To(Equal("repo"))
		Expect(config.PluginRepos()[0].URL).To(Equal("nowhere.com"))

		s, _ := semver.Make("3.1")
		Expect(config.IsMinAPIVersion(s)).To(Equal(false))

		config.SetMinCLIVersion("6.5.0")
		Expect(config.MinCLIVersion()).To(Equal("6.5.0"))

		config.SetMinRecommendedCLIVersion("6.9.0")
		Expect(config.MinRecommendedCLIVersion()).To(Equal("6.9.0"))
	})

	Describe("HasAPIEndpoint", func() {
		Context("when both endpoint and version are set", func() {
			BeforeEach(func() {
				config.SetAPIEndpoint("http://example.org")
				config.SetAPIVersion("42.1.2.3")
			})

			It("returns true", func() {
				Expect(config.HasAPIEndpoint()).To(BeTrue())
			})
		})

		Context("when endpoint is not set", func() {
			BeforeEach(func() {
				config.SetAPIVersion("42.1.2.3")
			})

			It("returns false", func() {
				Expect(config.HasAPIEndpoint()).To(BeFalse())
			})
		})

		Context("when version is not set", func() {
			BeforeEach(func() {
				config.SetAPIEndpoint("http://example.org")
			})

			It("returns false", func() {
				Expect(config.HasAPIEndpoint()).To(BeFalse())
			})
		})
	})

	Describe("UserGUID", func() {
		Context("with a valid access token", func() {
			BeforeEach(func() {
				config.SetAccessToken("bearer eyJhbGciOiJSUzI1NiJ9.eyJqdGkiOiJjNDE4OTllNS1kZTE1LTQ5NGQtYWFiNC04ZmNlYzUxN2UwMDUiLCJzdWIiOiI3NzJkZGEzZi02NjlmLTQyNzYtYjJiZC05MDQ4NmFiZTFmNmYiLCJzY29wZSI6WyJjbG91ZF9jb250cm9sbGVyLnJlYWQiLCJjbG91ZF9jb250cm9sbGVyLndyaXRlIiwib3BlbmlkIiwicGFzc3dvcmQud3JpdGUiXSwiY2xpZW50X2lkIjoiY2YiLCJjaWQiOiJjZiIsImdyYW50X3R5cGUiOiJwYXNzd29yZCIsInVzZXJfaWQiOiI3NzJkZGEzZi02NjlmLTQyNzYtYjJiZC05MDQ4NmFiZTFmNmYiLCJ1c2VyX25hbWUiOiJ1c2VyMUBleGFtcGxlLmNvbSIsImVtYWlsIjoidXNlcjFAZXhhbXBsZS5jb20iLCJpYXQiOjEzNzcwMjgzNTYsImV4cCI6MTM3NzAzNTU1NiwiaXNzIjoiaHR0cHM6Ly91YWEuYXJib3JnbGVuLmNmLWFwcC5jb20vb2F1dGgvdG9rZW4iLCJhdWQiOlsib3BlbmlkIiwiY2xvdWRfY29udHJvbGxlciIsInBhc3N3b3JkIl19.kjFJHi0Qir9kfqi2eyhHy6kdewhicAFu8hrPR1a5AxFvxGB45slKEjuP0_72cM_vEYICgZn3PcUUkHU9wghJO9wjZ6kiIKK1h5f2K9g-Iprv9BbTOWUODu1HoLIvg2TtGsINxcRYy_8LW1RtvQc1b4dBPoopaEH4no-BIzp0E5E")
			})

			It("returns the guid", func() {
				Expect(config.UserGUID()).To(Equal("772dda3f-669f-4276-b2bd-90486abe1f6f"))
			})
		})

		Context("with an invalid access token", func() {
			BeforeEach(func() {
				config.SetAccessToken("bearer eyJhbGciOiJSUzI1NiJ9")
			})

			It("returns an empty string", func() {
				Expect(config.UserGUID()).To(BeEmpty())
			})
		})
	})

	Describe("UserEmail", func() {
		Context("with a valid access token", func() {
			BeforeEach(func() {
				config.SetAccessToken("bearer eyJhbGciOiJSUzI1NiJ9.eyJqdGkiOiJjNDE4OTllNS1kZTE1LTQ5NGQtYWFiNC04ZmNlYzUxN2UwMDUiLCJzdWIiOiI3NzJkZGEzZi02NjlmLTQyNzYtYjJiZC05MDQ4NmFiZTFmNmYiLCJzY29wZSI6WyJjbG91ZF9jb250cm9sbGVyLnJlYWQiLCJjbG91ZF9jb250cm9sbGVyLndyaXRlIiwib3BlbmlkIiwicGFzc3dvcmQud3JpdGUiXSwiY2xpZW50X2lkIjoiY2YiLCJjaWQiOiJjZiIsImdyYW50X3R5cGUiOiJwYXNzd29yZCIsInVzZXJfaWQiOiI3NzJkZGEzZi02NjlmLTQyNzYtYjJiZC05MDQ4NmFiZTFmNmYiLCJ1c2VyX25hbWUiOiJ1c2VyMUBleGFtcGxlLmNvbSIsImVtYWlsIjoidXNlcjFAZXhhbXBsZS5jb20iLCJpYXQiOjEzNzcwMjgzNTYsImV4cCI6MTM3NzAzNTU1NiwiaXNzIjoiaHR0cHM6Ly91YWEuYXJib3JnbGVuLmNmLWFwcC5jb20vb2F1dGgvdG9rZW4iLCJhdWQiOlsib3BlbmlkIiwiY2xvdWRfY29udHJvbGxlciIsInBhc3N3b3JkIl19.kjFJHi0Qir9kfqi2eyhHy6kdewhicAFu8hrPR1a5AxFvxGB45slKEjuP0_72cM_vEYICgZn3PcUUkHU9wghJO9wjZ6kiIKK1h5f2K9g-Iprv9BbTOWUODu1HoLIvg2TtGsINxcRYy_8LW1RtvQc1b4dBPoopaEH4no-BIzp0E5E")
			})

			It("returns the email", func() {
				Expect(config.UserEmail()).To(Equal("user1@example.com"))
			})
		})

		Context("with an invalid access token", func() {
			BeforeEach(func() {
				config.SetAccessToken("bearer eyJhbGciOiJSUzI1NiJ9")
			})

			It("returns an empty string", func() {
				Expect(config.UserEmail()).To(BeEmpty())
			})
		})
	})

	Describe("NewRepositoryFromFilepath", func() {
		var configPath string

		It("returns nil repository if no error handler provided", func() {
			config = coreconfig.NewRepositoryFromFilepath(configPath, nil)
			Expect(config).To(BeNil())
		})

		Context("when the configuration file doesn't exist", func() {
			var tmpDir string

			BeforeEach(func() {
				tmpDir, err := ioutil.TempDir("", "test-config")
				if err != nil {
					Fail("Couldn't create tmp file")
				}

				configPath = filepath.Join(tmpDir, ".cf", "config.json")
			})

			AfterEach(func() {
				if tmpDir != "" {
					os.RemoveAll(tmpDir)
				}
			})

			It("has sane defaults when there is no config to read", func() {
				config = coreconfig.NewRepositoryFromFilepath(configPath, func(err error) {
					panic(err)
				})

				Expect(config.APIEndpoint()).To(Equal(""))
				Expect(config.APIVersion()).To(Equal(""))
				Expect(config.AuthenticationEndpoint()).To(Equal(""))
				Expect(config.AccessToken()).To(Equal(""))
			})
		})

		Context("when the configuration version is older than the current version", func() {
			BeforeEach(func() {
				cwd, err := os.Getwd()
				Expect(err).NotTo(HaveOccurred())
				configPath = filepath.Join(cwd, "..", "..", "..", "fixtures", "config", "outdated-config", ".cf", "config.json")
			})

			It("returns a new empty config", func() {
				config = coreconfig.NewRepositoryFromFilepath(configPath, func(err error) {
					panic(err)
				})

				Expect(config.APIEndpoint()).To(Equal(""))
			})
		})
	})

	Describe("IsMinCLIVersion", func() {
		It("returns true when the actual version is BUILT_FROM_SOURCE", func() {
			Expect(config.IsMinCLIVersion("BUILT_FROM_SOURCE")).To(BeTrue())
		})

		It("returns true when the MinCLIVersion is empty", func() {
			config.SetMinCLIVersion("")
			Expect(config.IsMinCLIVersion("1.2.3")).To(BeTrue())
		})

		It("returns false when the actual version is less than the MinCLIVersion", func() {
			actualVersion := "1.2.3+abc123"
			minCLIVersion := "1.2.4"
			config.SetMinCLIVersion(minCLIVersion)
			Expect(config.IsMinCLIVersion(actualVersion)).To(BeFalse())
		})

		It("returns true when the actual version is equal to the MinCLIVersion", func() {
			actualVersion := "1.2.3+abc123"
			minCLIVersion := "1.2.3"
			config.SetMinCLIVersion(minCLIVersion)
			Expect(config.IsMinCLIVersion(actualVersion)).To(BeTrue())
		})

		It("returns true when the actual version is greater than the MinCLIVersion", func() {
			actualVersion := "1.2.3+abc123"
			minCLIVersion := "1.2.2"
			config.SetMinCLIVersion(minCLIVersion)
			Expect(config.IsMinCLIVersion(actualVersion)).To(BeTrue())
		})
	})
})
