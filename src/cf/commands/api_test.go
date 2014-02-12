package commands_test

import (
	. "cf/commands"
	"cf/configuration"
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

var _ = Describe("Testing with ginkgo", func() {
	It("TestApiWithoutArgument", func() {
		config := testconfig.NewRepository()
		config.SetApiEndpoint("https://api.run.pivotal.io")
		config.SetApiVersion("2.0")

		endpointRepo := &testapi.FakeEndpointRepo{Config: config}

		ui := callApi([]string{}, config, endpointRepo)

		Expect(len(ui.Outputs)).To(Equal(1))
		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"https://api.run.pivotal.io", "2.0"},
		})
	})

	It("TestApiWhenChangingTheEndpoint", func() {
		config := testconfig.NewRepository()
		endpointRepo := &testapi.FakeEndpointRepo{Config: config}

		ui := callApi([]string{"http://example.com"}, config, endpointRepo)

		Expect(endpointRepo.UpdateEndpointReceived).To(Equal("http://example.com"))
		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Setting api endpoint to", "example.com"},
			{"OK"},
		})
	})

	It("TestApiWithTrailingSlash", func() {
		config := testconfig.NewRepository()
		endpointRepo := &testapi.FakeEndpointRepo{Config: config}

		ui := callApi([]string{"https://example.com/"}, config, endpointRepo)

		Expect(endpointRepo.UpdateEndpointReceived).To(Equal("https://example.com"))
		testassert.SliceContains(ui.Outputs, testassert.Lines{
			{"Setting api endpoint to", "example.com"},
			{"OK"},
		})
	})

})
