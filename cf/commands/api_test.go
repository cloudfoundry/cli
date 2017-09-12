package commands_test

import (
	"fmt"

	"code.cloudfoundry.org/cli/cf"
	"code.cloudfoundry.org/cli/cf/commandregistry"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig"
	"code.cloudfoundry.org/cli/cf/errors"
	testconfig "code.cloudfoundry.org/cli/util/testhelpers/configuration"
	testterm "code.cloudfoundry.org/cli/util/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/cli/cf/api"
	"code.cloudfoundry.org/cli/cf/commands"
	"code.cloudfoundry.org/cli/cf/configuration/coreconfig/coreconfigfakes"
	"code.cloudfoundry.org/cli/cf/flags"
	. "code.cloudfoundry.org/cli/util/testhelpers/matchers"
)

var _ = Describe("Api", func() {
	var (
		config       coreconfig.Repository
		endpointRepo *coreconfigfakes.FakeEndpointRepository
		deps         commandregistry.Dependency
		ui           *testterm.FakeUI
		cmd          commands.API
		flagContext  flags.FlagContext
		repoLocator  api.RepositoryLocator
		runCLIErr    error
	)

	callApi := func(args []string) {
		err := flagContext.Parse(args...)
		Expect(err).NotTo(HaveOccurred())

		runCLIErr = cmd.Execute(flagContext)
	}

	BeforeEach(func() {
		ui = new(testterm.FakeUI)
		config = testconfig.NewRepository()
		endpointRepo = new(coreconfigfakes.FakeEndpointRepository)

		endpointRepo.GetCCInfoStub = func(endpoint string) (*coreconfig.CCInfo, string, error) {
			return &coreconfig.CCInfo{
				APIVersion:               config.APIVersion(),
				AuthorizationEndpoint:    config.AuthenticationEndpoint(),
				MinCLIVersion:            config.MinCLIVersion(),
				MinRecommendedCLIVersion: config.MinRecommendedCLIVersion(),
				SSHOAuthClient:           config.SSHOAuthClient(),
				RoutingAPIEndpoint:       config.RoutingAPIEndpoint(),
			}, endpoint, nil
		}

		repoLocator = api.RepositoryLocator{}.SetEndpointRepository(endpointRepo)

		deps = commandregistry.Dependency{
			UI:          ui,
			Config:      config,
			RepoLocator: repoLocator,
		}

		cmd = commands.API{}.SetDependency(deps, false).(commands.API)
		flagContext = flags.NewFlagContext(cmd.MetaData().Flags)
	})

	Context("when the api endpoint's ssl certificate is invalid", func() {
		It("warns the user and prints out a tip", func() {
			endpointRepo.GetCCInfoReturns(nil, "", errors.NewInvalidSSLCert("https://buttontomatoes.org", "why? no. go away"))

			callApi([]string{"https://buttontomatoes.org"})
			Expect(runCLIErr).To(HaveOccurred())
			Expect(runCLIErr.Error()).To(ContainSubstring("Invalid SSL Cert for https://buttontomatoes.org"))
			Expect(runCLIErr.Error()).To(ContainSubstring("TIP"))
			Expect(runCLIErr.Error()).To(ContainSubstring("--skip-ssl-validation"))
		})
	})

	Context("when the user does not provide an endpoint", func() {
		Context("when the endpoint is set in the config", func() {
			BeforeEach(func() {
				config.SetAPIEndpoint("https://api.run.pivotal.io")
				config.SetAPIVersion("2.0")
				config.SetSSLDisabled(true)
			})

			It("prints out the api endpoint and appropriately sets the config", func() {
				callApi([]string{})

				Expect(ui.Outputs()).To(ContainSubstrings([]string{"https://api.run.pivotal.io", "2.0"}))
				Expect(config.IsSSLDisabled()).To(BeTrue())
			})

			Context("when the --unset flag is passed", func() {
				It("unsets the APIEndpoint", func() {
					callApi([]string{"--unset"})

					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"Unsetting api endpoint..."},
						[]string{"OK"},
						[]string{"No api endpoint set."},
					))
					Expect(config.APIEndpoint()).To(Equal(""))
				})
			})
		})

		Context("when the endpoint is not set in the config", func() {
			It("prompts the user to set an endpoint", func() {
				callApi([]string{})

				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"No api endpoint set", fmt.Sprintf("Use '%s api' to set an endpoint", cf.Name)},
				))
			})
		})
	})

	Context("when the user provides the --skip-ssl-validation flag", func() {
		It("updates the SSLDisabled field in config", func() {
			config.SetSSLDisabled(false)
			callApi([]string{"--skip-ssl-validation", "https://example.com"})

			Expect(config.IsSSLDisabled()).To(Equal(true))
		})
	})

	Context("the user provides an endpoint", func() {
		Describe("when the user passed in the skip-ssl-validation flag", func() {
			It("disables SSL validation in the config", func() {
				callApi([]string{"--skip-ssl-validation", "https://example.com"})

				Expect(endpointRepo.GetCCInfoCallCount()).To(Equal(1))
				Expect(endpointRepo.GetCCInfoArgsForCall(0)).To(Equal("https://example.com"))
				Expect(config.IsSSLDisabled()).To(BeTrue())
			})
		})

		Context("when the user passed in the unset flag", func() {
			Context("when the config.APIEndpoint is set", func() {
				BeforeEach(func() {
					config.SetAPIEndpoint("some-silly-thing")
				})

				It("unsets the APIEndpoint", func() {
					callApi([]string{"--unset", "https://example.com"})

					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"Unsetting api endpoint..."},
						[]string{"OK"},
						[]string{"No api endpoint set."},
					))
					Expect(config.APIEndpoint()).To(Equal(""))
				})
			})

			Context("when the config.APIEndpoint is empty", func() {
				It("unsets the APIEndpoint", func() {
					callApi([]string{"--unset", "https://example.com"})

					Expect(ui.Outputs()).To(ContainSubstrings(
						[]string{"Unsetting api endpoint..."},
						[]string{"OK"},
						[]string{"No api endpoint set."},
					))
					Expect(config.APIEndpoint()).To(Equal(""))
				})
			})

		})

		Context("when the ssl certificate is valid", func() {
			It("updates the api endpoint with the given url", func() {
				callApi([]string{"https://example.com"})
				Expect(endpointRepo.GetCCInfoCallCount()).To(Equal(1))
				Expect(endpointRepo.GetCCInfoArgsForCall(0)).To(Equal("https://example.com"))
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Setting api endpoint to", "example.com"},
					[]string{"OK"},
				))
			})

			It("trims trailing slashes from the api endpoint", func() {
				callApi([]string{"https://example.com/"})
				Expect(endpointRepo.GetCCInfoCallCount()).To(Equal(1))
				Expect(endpointRepo.GetCCInfoArgsForCall(0)).To(Equal("https://example.com"))
				Expect(ui.Outputs()).To(ContainSubstrings(
					[]string{"Setting api endpoint to", "example.com"},
					[]string{"OK"},
				))
			})
		})

		Context("when the ssl certificate is invalid", func() {
			BeforeEach(func() {
				endpointRepo.GetCCInfoReturns(nil, "", errors.NewInvalidSSLCert("https://example.com", "it don't work"))
			})

			It("fails and gives the user a helpful message about skipping", func() {
				callApi([]string{"https://example.com"})
				Expect(runCLIErr).To(HaveOccurred())
				Expect(runCLIErr.Error()).To(ContainSubstring("Invalid SSL Cert"))
				Expect(runCLIErr.Error()).To(ContainSubstring("https://example.com"))
				Expect(runCLIErr.Error()).To(ContainSubstring("TIP"))

				Expect(config.APIEndpoint()).To(Equal(""))
			})
		})

		Describe("unencrypted http endpoints", func() {
			It("warns the user", func() {
				callApi([]string{"http://example.com"})
				Expect(ui.Outputs()).To(ContainSubstrings([]string{"Warning"}))
			})
		})
	})
})
