package commands_test

import (
	"fmt"

	"github.com/cloudfoundry/cli/cf"
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/command_registry"
	"github.com/cloudfoundry/cli/cf/configuration/core_config"
	"github.com/cloudfoundry/cli/cf/errors"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

var _ = Describe("api command", func() {
	var (
		config              core_config.Repository
		endpointRepo        *testapi.FakeEndpointRepo
		deps                command_registry.Dependency
		requirementsFactory *testreq.FakeReqFactory
		ui                  *testterm.FakeUI
	)

	updateCommandDependency := func(pluginCall bool) {
		deps.Ui = ui
		deps.Config = config
		deps.RepoLocator = deps.RepoLocator.SetEndpointRepository(endpointRepo)
		command_registry.Commands.SetCommand(command_registry.Commands.FindCommand("api").SetDependency(deps, pluginCall))
	}

	callApi := func(args []string, config core_config.Repository, endpointRepo *testapi.FakeEndpointRepo) {
		testcmd.RunCliCommand("api", args, requirementsFactory, updateCommandDependency, false)
	}

	BeforeEach(func() {
		ui = new(testterm.FakeUI)
		requirementsFactory = &testreq.FakeReqFactory{}
		config = testconfig.NewRepository()
		endpointRepo = &testapi.FakeEndpointRepo{}
		deps = command_registry.NewDependency()
	})

	Context("when the api endpoint's ssl certificate is invalid", func() {
		It("warns the user and prints out a tip", func() {
			endpointRepo.UpdateEndpointError = errors.NewInvalidSSLCert("https://buttontomatoes.org", "why? no. go away")
			callApi([]string{"https://buttontomatoes.org"}, config, endpointRepo)

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
				callApi([]string{}, config, endpointRepo)

				Expect(ui.Outputs).To(ContainSubstrings([]string{"https://api.run.pivotal.io", "2.0"}))
				Expect(config.IsSSLDisabled()).To(BeTrue())
			})

			Context("when the --unset flag is passed", func() {
				It("unsets the ApiEndpoint", func() {
					callApi([]string{"--unset"}, config, endpointRepo)

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
				callApi([]string{}, config, endpointRepo)

				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"No api endpoint set", fmt.Sprintf("Use '%s api' to set an endpoint", cf.Name())},
				))
			})
		})
	})

	Context("when the user provides the --skip-ssl-validation flag", func() {
		It("updates the SSLDisabled field in config", func() {
			config.SetSSLDisabled(false)
			callApi([]string{"--skip-ssl-validation", "https://example.com"}, config, endpointRepo)

			Expect(config.IsSSLDisabled()).To(Equal(true))
		})
	})

	Context("the user provides an endpoint", func() {
		Describe("when the user passed in the skip-ssl-validation flag", func() {
			It("disables SSL validation in the config", func() {
				callApi([]string{"--skip-ssl-validation", "https://example.com"}, config, endpointRepo)

				Expect(endpointRepo.UpdateEndpointReceived).To(Equal("https://example.com"))
				Expect(config.IsSSLDisabled()).To(BeTrue())
			})
		})

		Context("when the user passed in the unset flag", func() {
			Context("when the config.ApiEndpoint is set", func() {
				BeforeEach(func() {
					config.SetApiEndpoint("some-silly-thing")
					ui = new(testterm.FakeUI)
				})

				It("unsets the ApiEndpoint", func() {
					callApi([]string{"--unset", "https://example.com"}, config, endpointRepo)

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
					callApi([]string{"--unset", "https://example.com"}, config, endpointRepo)

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
				callApi([]string{"https://example.com"}, config, endpointRepo)
				Expect(endpointRepo.UpdateEndpointReceived).To(Equal("https://example.com"))
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Setting api endpoint to", "example.com"},
					[]string{"OK"},
				))
			})

			It("trims trailing slashes from the api endpoint", func() {
				callApi([]string{"https://example.com/"}, config, endpointRepo)
				Expect(endpointRepo.UpdateEndpointReceived).To(Equal("https://example.com"))
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Setting api endpoint to", "example.com"},
					[]string{"OK"},
				))
			})
		})

		Context("when the ssl certificate is invalid", func() {
			BeforeEach(func() {
				endpointRepo.UpdateEndpointError = errors.NewInvalidSSLCert("https://example.com", "it don't work")
			})

			It("fails and gives the user a helpful message about skipping", func() {
				callApi([]string{"https://example.com"}, config, endpointRepo)

				Expect(config.ApiEndpoint()).To(Equal(""))
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Invalid SSL Cert", "https://example.com"},
					[]string{"TIP", "api"},
				))
			})
		})

		Describe("unencrypted http endpoints", func() {
			It("warns the user", func() {
				callApi([]string{"http://example.com"}, config, endpointRepo)
				Expect(ui.Outputs).To(ContainSubstrings([]string{"Warning"}))
			})
		})
	})
})
