package commands_test

import (
	"cf"
	. "cf/commands"
	"cf/configuration"
	"cf/errors"
	"fmt"
	"github.com/codegangsta/cli"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

func callApi(args []string, config configuration.ReadWriter, endpointRepo *testapi.FakeEndpointRepo) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)

	cmd := NewApi(ui, config, endpointRepo)
	ctxt := testcmd.NewContext("api", args)
	reqFactory := &testreq.FakeReqFactory{}
	testcmd.RunCommand(cmd, ctxt, reqFactory)
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

			testassert.SliceContains(ui.Outputs, testassert.Lines{
				{"FAILED"},
				{"SSL cert", "https://buttontomatoes.org"},
				{"TIP", "--skip-ssl-validation"},
			})
		})
	})

	Context("when the user does not provide an endpoint", func() {
		Context("when the endpoint is set in the config", func() {
			var (
				ui         *testterm.FakeUI
				cmd        Command
				ctx        *cli.Context
				reqFactory *testreq.FakeReqFactory
			)

			BeforeEach(func() {
				config.SetApiEndpoint("https://api.run.pivotal.io")
				config.SetApiVersion("2.0")
				config.SetSSLDisabled(true)

				ui = new(testterm.FakeUI)
				cmd = NewApi(ui, config, endpointRepo)
				ctx = testcmd.NewContext("api", []string{})
				reqFactory = &testreq.FakeReqFactory{}
			})

			JustBeforeEach(func() {
				testcmd.RunCommand(cmd, ctx, reqFactory)
			})

			It("prints out the api endpoint", func() {
				testassert.SliceContains(ui.Outputs, testassert.Lines{
					{"https://api.run.pivotal.io", "2.0"},
				})
			})

			It("should not change the SSL setting in the config", func() {
				Expect(config.IsSSLDisabled()).To(BeTrue())
			})
		})

		Context("when the endpoint is not set in the config", func() {
			It("prompts the user to set an endpoint", func() {
				ui := callApi([]string{}, config, endpointRepo)

				testassert.SliceContains(ui.Outputs, testassert.Lines{
					{"No api endpoint set", fmt.Sprintf("use '%s api' to set an endpoint", cf.Name())},
				})
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
			BeforeEach(func() {
				ui = callApi([]string{"https://example.com"}, config, endpointRepo)
			})

			It("updates the api endpoint with the given url", func() {
				Expect(endpointRepo.UpdateEndpointReceived).To(Equal("https://example.com"))
				testassert.SliceContains(ui.Outputs, testassert.Lines{
					{"Setting api endpoint to", "example.com"},
					{"OK"},
				})
			})

			It("trims trailing slashes from the api endpoint", func() {
				Expect(endpointRepo.UpdateEndpointReceived).To(Equal("https://example.com"))
				testassert.SliceContains(ui.Outputs, testassert.Lines{
					{"Setting api endpoint to", "example.com"},
					{"OK"},
				})
			})
		})

		Context("when the ssl certificate is invalid", func() {
			BeforeEach(func() {
				endpointRepo.UpdateEndpointError = errors.NewInvalidSSLCert("https://example.com", "it don't work")
			})

			It("fails and gives the user a helpful message about skipping", func() {
				ui := callApi([]string{"https://example.com"}, config, endpointRepo)

				Expect(config.ApiEndpoint()).To(Equal(""))
				testassert.SliceContains(ui.Outputs, testassert.Lines{
					{"Invalid SSL Cert", "https://example.com"},
					{"TIP"},
				})
			})
		})

		Describe("unencrypted http endpoints", func() {
			It("warns the user", func() {
				ui = callApi([]string{"http://example.com"}, config, endpointRepo)
				testassert.SliceContains(ui.Outputs, testassert.Lines{
					{"Warning"},
				})
			})
		})
	})
})
