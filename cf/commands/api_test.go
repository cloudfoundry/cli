package commands_test

import (
	"fmt"

	"github.com/cloudfoundry/cli/cf"
	"github.com/cloudfoundry/cli/cf/commandregistry"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig"
	"github.com/cloudfoundry/cli/cf/errors"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/cli/cf/api"
	"github.com/cloudfoundry/cli/cf/commands"
	"github.com/cloudfoundry/cli/cf/configuration/coreconfig/coreconfigfakes"
	"github.com/cloudfoundry/cli/flags"
	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

var _ = Describe("Api", func() {
	var (
		config              coreconfig.Repository
		endpointRepo        *coreconfigfakes.FakeEndpointRepository
		deps                commandregistry.Dependency
		requirementsFactory *testreq.FakeReqFactory
		ui                  *testterm.FakeUI
		cmd                 commands.Api
		flagContext         flags.FlagContext
		repoLocator         api.RepositoryLocator
	)

	callApi := func(args []string) {
		err := flagContext.Parse(args...)
		Expect(err).NotTo(HaveOccurred())

		cmd.Execute(flagContext)
	}

	BeforeEach(func() {
		ui = new(testterm.FakeUI)
		requirementsFactory = &testreq.FakeReqFactory{}
		config = testconfig.NewRepository()
		endpointRepo = new(coreconfigfakes.FakeEndpointRepository)

		endpointRepo.GetCCInfoStub = func(endpoint string) (*coreconfig.CCInfo, string, error) {
			return &coreconfig.CCInfo{
				ApiVersion:               config.ApiVersion(),
				AuthorizationEndpoint:    config.AuthenticationEndpoint(),
				LoggregatorEndpoint:      "log/endpoint",
				MinCliVersion:            config.MinCliVersion(),
				MinRecommendedCliVersion: config.MinRecommendedCliVersion(),
				SSHOAuthClient:           config.SSHOAuthClient(),
				RoutingApiEndpoint:       config.RoutingApiEndpoint(),
			}, endpoint, nil
		}

		repoLocator = api.RepositoryLocator{}.SetEndpointRepository(endpointRepo)

		deps = commandregistry.Dependency{
			Ui:          ui,
			Config:      config,
			RepoLocator: repoLocator,
		}

		cmd = commands.Api{}.SetDependency(deps, false).(commands.Api)
		flagContext = flags.NewFlagContext(cmd.MetaData().Flags)
	})

	Context("when the api endpoint's ssl certificate is invalid", func() {
		It("warns the user and prints out a tip", func() {
			endpointRepo.GetCCInfoReturns(nil, "", errors.NewInvalidSSLCert("https://buttontomatoes.org", "why? no. go away"))

			err := flagContext.Parse("https://buttontomatoes.org")
			Expect(err).NotTo(HaveOccurred())

			Expect(func() {
				cmd.Execute(flagContext)
			}).To(Panic())

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"FAILED"},
				[]string{"SSL Cert", "https://buttontomatoes.org"},
				[]string{"TIP", "--skip-ssl-validation"},
			))
		})
	})

	Context("when the user does not provide an endpoint", func() {
		Context("when the endpoint is set in the config", func() {
			BeforeEach(func() {
				config.SetApiEndpoint("https://api.run.pivotal.io")
				config.SetApiVersion("2.0")
				config.SetSSLDisabled(true)
			})

			It("prints out the api endpoint and appropriately sets the config", func() {
				callApi([]string{})

				Expect(ui.Outputs).To(ContainSubstrings([]string{"https://api.run.pivotal.io", "2.0"}))
				Expect(config.IsSSLDisabled()).To(BeTrue())
			})

			Context("when the --unset flag is passed", func() {
				It("unsets the ApiEndpoint", func() {
					callApi([]string{"--unset"})

					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Unsetting api endpoint..."},
						[]string{"OK"},
						[]string{"No api endpoint set."},
					))
					Expect(config.ApiEndpoint()).To(Equal(""))
				})
			})
		})

		Context("when the endpoint is not set in the config", func() {
			It("prompts the user to set an endpoint", func() {
				callApi([]string{})

				Expect(ui.Outputs).To(ContainSubstrings(
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
			Context("when the config.ApiEndpoint is set", func() {
				BeforeEach(func() {
					config.SetApiEndpoint("some-silly-thing")
				})

				It("unsets the ApiEndpoint", func() {
					callApi([]string{"--unset", "https://example.com"})

					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Unsetting api endpoint..."},
						[]string{"OK"},
						[]string{"No api endpoint set."},
					))
					Expect(config.ApiEndpoint()).To(Equal(""))
				})
			})

			Context("when the config.ApiEndpoint is empty", func() {
				It("unsets the ApiEndpoint", func() {
					callApi([]string{"--unset", "https://example.com"})

					Expect(ui.Outputs).To(ContainSubstrings(
						[]string{"Unsetting api endpoint..."},
						[]string{"OK"},
						[]string{"No api endpoint set."},
					))
					Expect(config.ApiEndpoint()).To(Equal(""))
				})
			})

		})

		Context("when the ssl certificate is valid", func() {
			It("updates the api endpoint with the given url", func() {
				callApi([]string{"https://example.com"})
				Expect(endpointRepo.GetCCInfoCallCount()).To(Equal(1))
				Expect(endpointRepo.GetCCInfoArgsForCall(0)).To(Equal("https://example.com"))
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Setting api endpoint to", "example.com"},
					[]string{"OK"},
				))
			})

			It("trims trailing slashes from the api endpoint", func() {
				callApi([]string{"https://example.com/"})
				Expect(endpointRepo.GetCCInfoCallCount()).To(Equal(1))
				Expect(endpointRepo.GetCCInfoArgsForCall(0)).To(Equal("https://example.com"))
				Expect(ui.Outputs).To(ContainSubstrings(
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
				err := flagContext.Parse("https://example.com")
				Expect(err).NotTo(HaveOccurred())

				Expect(func() {
					cmd.Execute(flagContext)
				}).To(Panic())

				Expect(config.ApiEndpoint()).To(Equal(""))
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Invalid SSL Cert", "https://example.com"},
					[]string{"TIP", "api"},
				))
			})
		})

		Describe("unencrypted http endpoints", func() {
			It("warns the user", func() {
				callApi([]string{"http://example.com"})
				Expect(ui.Outputs).To(ContainSubstrings([]string{"Warning"}))
			})
		})
	})
})
