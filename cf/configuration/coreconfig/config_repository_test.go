package coreconfig_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cli/cf/configuration"
	"code.cloudfoundry.org/cli/cf/configuration/configurationfakes"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/models"
	"code.cloudfoundry.org/cli/version"
	"github.com/blang/semver"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Configuration Repository", func() {

	const (
		AccessTokenForHumanUsers = "bearer eyJhbGciOiJSUzI1NiIsImtpZCI6ImxlZ2FjeS10b2tlbi1rZXkiLCJ0eXAiOiJKV1QifQ.eyJqdGkiOiI3YzZkMDA2MjA2OTI0NmViYWI0ZjBmZjY3NGQ3Zjk4OSIsInN1YiI6Ijk1MTliZTNlLTQ0ZDktNDBkMC1hYjlhLWY0YWNlMTFkZjE1OSIsInNjb3BlIjpbIm9wZW5pZCIsInJvdXRpbmcucm91dGVyX2dyb3Vwcy53cml0ZSIsInNjaW0ucmVhZCIsImNsb3VkX2NvbnRyb2xsZXIuYWRtaW4iLCJ1YWEudXNlciIsInJvdXRpbmcucm91dGVyX2dyb3Vwcy5yZWFkIiwiY2xvdWRfY29udHJvbGxlci5yZWFkIiwicGFzc3dvcmQud3JpdGUiLCJjbG91ZF9jb250cm9sbGVyLndyaXRlIiwiZG9wcGxlci5maXJlaG9zZSIsInNjaW0ud3JpdGUiXSwiY2xpZW50X2lkIjoiY2YiLCJjaWQiOiJjZiIsImF6cCI6ImNmIiwiZ3JhbnRfdHlwZSI6InBhc3N3b3JkIiwidXNlcl9pZCI6Ijk1MTliZTNlLTQ0ZDktNDBkMC1hYjlhLWY0YWNlMTFkZjE1OSIsIm9yaWdpbiI6InVhYSIsInVzZXJfbmFtZSI6ImFkbWluIiwiZW1haWwiOiJhZG1pbiIsImF1dGhfdGltZSI6MTQ3MzI4NDU3NywicmV2X3NpZyI6IjZiMjdkYTZjIiwiaWF0IjoxNDczMjg0NTc3LCJleHAiOjE0NzMyODUxNzcsImlzcyI6Imh0dHBzOi8vdWFhLmJvc2gtbGl0ZS5jb20vb2F1dGgvdG9rZW4iLCJ6aWQiOiJ1YWEiLCJhdWQiOlsiY2YiLCJvcGVuaWQiLCJyb3V0aW5nLnJvdXRlcl9ncm91cHMiLCJzY2ltIiwiY2xvdWRfY29udHJvbGxlciIsInVhYSIsInBhc3N3b3JkIiwiZG9wcGxlciJdfQ.OcH_w9yIKJkEcTZMThIs-qJAHk3G0JwNjG-aomVH9hKye4ciFO6IMQMLKmCBrrAQVc7ST1SZZwq7gv12Dq__6Jp-hai0a2_ADJK-Vc9YXyNZKgYTWIeVNGM1JGdHgFSrBR2Lz7IIrH9HqeN8plrKV5HzU8uI9LL4lyOCjbXJ9cM"
		InvalidAccessToken       = "bearer eyJhbGciOiJSUzI1NiJ9"
	)

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

	It("has acccessor methods for all config fields", func() {
		config.SetAPIEndpoint("http://api.the-endpoint")
		Expect(config.APIEndpoint()).To(Equal("http://api.the-endpoint"))

		config.SetAPIVersion("3")
		Expect(config.APIVersion()).To(Equal("3"))

		config.SetAuthenticationEndpoint("http://auth.the-endpoint")
		Expect(config.AuthenticationEndpoint()).To(Equal("http://auth.the-endpoint"))

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

		config.SetDopplerEndpoint("doppler.the-endpoint")
		Expect(config.DopplerEndpoint()).To(Equal("doppler.the-endpoint"))

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

	Describe("APIEndpoint", func() {
		It("sanitizes the target URL", func() {
			config.SetAPIEndpoint("http://api.the-endpoint/")
			Expect(config.APIEndpoint()).To(Equal("http://api.the-endpoint"))
		})
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
				config.SetAccessToken(AccessTokenForHumanUsers)
			})

			It("returns the guid", func() {
				Expect(config.UserGUID()).To(Equal("9519be3e-44d9-40d0-ab9a-f4ace11df159"))
			})
		})

		Context("with an invalid access token", func() {
			BeforeEach(func() {
				config.SetAccessToken(InvalidAccessToken)
			})

			It("returns an empty string", func() {
				Expect(config.UserGUID()).To(BeEmpty())
			})
		})
	})

	Describe("Username", func() {
		Context("with a valid user access token", func() {
			BeforeEach(func() {
				config.SetAccessToken(AccessTokenForHumanUsers)
			})

			It("returns the username", func() {
				Expect(config.Username()).To(Equal("admin"))
			})
		})

		Context("with an invalid access token", func() {
			BeforeEach(func() {
				config.SetAccessToken(InvalidAccessToken)
			})

			It("returns an empty string", func() {
				Expect(config.Username()).To(BeEmpty())
			})
		})
	})

	Describe("UserEmail", func() {
		Context("with a valid access token", func() {
			BeforeEach(func() {
				config.SetAccessToken(AccessTokenForHumanUsers)
			})

			It("returns the email", func() {
				Expect(config.UserEmail()).To(Equal("admin"))
			})
		})

		Context("with an invalid access token", func() {
			BeforeEach(func() {
				config.SetAccessToken(InvalidAccessToken)
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
				var err error
				tmpDir, err = ioutil.TempDir("", "test-config")
				if err != nil {
					Fail("Couldn't create tmp file")
				}

				configPath = filepath.Join(tmpDir, ".cf", "config.json")
			})

			AfterEach(func() {
				Expect(os.RemoveAll(tmpDir)).ToNot(HaveOccurred())
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
		It("returns true when the actual version is the default version string", func() {
			Expect(config.IsMinCLIVersion(version.DefaultVersion)).To(BeTrue())
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
