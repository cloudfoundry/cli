package commands_test

import (
	"cf"
	. "cf/commands"
	"cf/configuration"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

func callApi(args []string, config configuration.Reader, endpointRepo *testapi.FakeEndpointRepo) (ui *testterm.FakeUI) {
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
		endpointRepo = &testapi.FakeEndpointRepo{Config: config}
	})

	Context("when the user does not provide an endpoint", func() {
		Context("when the endpoint is set", func() {
			It("prints out the api endpoint", func() {
				config.SetApiEndpoint("https://api.run.pivotal.io")
				config.SetApiVersion("2.0")

				ui := callApi([]string{}, config, endpointRepo)

				testassert.SliceContains(ui.Outputs, testassert.Lines{
					{"https://api.run.pivotal.io", "2.0"},
				})
			})
		})

		Context("when the user has not set an endpoint", func() {
			It("prompts the user to set an endpoint", func() {
				ui := callApi([]string{}, config, endpointRepo)

				testassert.SliceContains(ui.Outputs, testassert.Lines{
					{"No api endpoint set", fmt.Sprintf("use '%s api' to set an endpoint", cf.Name())},
				})
			})
		})
	})

	Context("the user provides an api endpoint", func() {
		var (
			ui *testterm.FakeUI
		)

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
