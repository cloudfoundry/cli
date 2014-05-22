package commands_test

import (
	"fmt"

	"github.com/cloudfoundry/cli/cf"
	. "github.com/cloudfoundry/cli/cf/commands"
	"github.com/cloudfoundry/cli/cf/configuration"
	"github.com/cloudfoundry/cli/cf/errors"
	testapi "github.com/cloudfoundry/cli/testhelpers/api"
	testcmd "github.com/cloudfoundry/cli/testhelpers/commands"
	testconfig "github.com/cloudfoundry/cli/testhelpers/configuration"
	testreq "github.com/cloudfoundry/cli/testhelpers/requirements"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/cloudfoundry/cli/testhelpers/matchers"
)

func callApi(args []string, config configuration.ReadWriter, endpointRepo *testapi.FakeEndpointRepo) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)

	cmd := NewApi(ui, config, endpointRepo)
	requirementsFactory := &testreq.FakeReqFactory{}
	testcmd.RunCommand(cmd, args, requirementsFactory)
	return
}

var _ = Describe("api command", func() {
	var (
		config       configuration.ReadWriter
		endpointRepo *testapi.FakeEndpointRepo
	)

	BeforeEach(func() {
		config = testconfig.NewRepository()
		endpointRepo = &testapi.FakeEndpointRepo{}
	})

	Context("when the api endpoint's ssl certificate is invalid", func() {
		It("warns the user and prints out a tip", func() {
			endpointRepo.UpdateEndpointError = errors.NewInvalidSSLCert("https://buttontomatoes.org", "why? no. go away")
			ui := callApi([]string{"https://buttontomatoes.org"}, config, endpointRepo)

			Expect(ui.Outputs).To(ContainSubstrings(
				[]string{"FAILED"},
				[]string{"SSL Cert", "https://buttontomatoes.org"},
				[]string{"TIP", "--skip-ssl-validation"},
			))
		})
	})

	Context("when the user does not provide an endpoint", func() {
		Context("when the endpoint is set in the config", func() {
			var (
				ui                  *testterm.FakeUI
				requirementsFactory *testreq.FakeReqFactory
			)

			BeforeEach(func() {
				config.SetApiEndpoint("https://api.run.pivotal.io")
				config.SetApiVersion("2.0")
				config.SetSSLDisabled(true)

				ui = new(testterm.FakeUI)
				requirementsFactory = &testreq.FakeReqFactory{}
			})

			JustBeforeEach(func() {
				testcmd.RunCommand(NewApi(ui, config, endpointRepo), []string{}, requirementsFactory)
			})

			It("prints out the api endpoint", func() {
				Expect(ui.Outputs).To(ContainSubstrings([]string{"https://api.run.pivotal.io", "2.0"}))
			})

			It("should not change the SSL setting in the config", func() {
				Expect(config.IsSSLDisabled()).To(BeTrue())
			})
		})

		Context("when the endpoint is not set in the config", func() {
			It("prompts the user to set an endpoint", func() {
				ui := callApi([]string{}, config, endpointRepo)

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
		var (
			ui *testterm.FakeUI
		)

		Describe("when the user passed in the skip-ssl-validation flag", func() {
			It("disables SSL validation in the config", func() {
				ui = callApi([]string{"--skip-ssl-validation", "https://example.com"}, config, endpointRepo)

				Expect(endpointRepo.UpdateEndpointReceived).To(Equal("https://example.com"))
				Expect(config.IsSSLDisabled()).To(BeTrue())
			})
		})

		Context("when the ssl certificate is valid", func() {
			It("updates the api endpoint with the given url", func() {
				ui = callApi([]string{"https://example.com"}, config, endpointRepo)
				Expect(endpointRepo.UpdateEndpointReceived).To(Equal("https://example.com"))
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Setting api endpoint to", "example.com"},
					[]string{"OK"},
				))
			})

			It("trims trailing slashes from the api endpoint", func() {
				ui = callApi([]string{"https://example.com/"}, config, endpointRepo)
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
				ui := callApi([]string{"https://example.com"}, config, endpointRepo)

				Expect(config.ApiEndpoint()).To(Equal(""))
				Expect(ui.Outputs).To(ContainSubstrings(
					[]string{"Invalid SSL Cert", "https://example.com"},
					[]string{"TIP"},
				))
			})
		})

		Describe("unencrypted http endpoints", func() {
			It("warns the user", func() {
				ui = callApi([]string{"http://example.com"}, config, endpointRepo)
				Expect(ui.Outputs).To(ContainSubstrings([]string{"Warning"}))
			})
		})
	})
})
