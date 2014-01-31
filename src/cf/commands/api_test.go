package commands_test

import (
	. "cf/commands"
	"cf/configuration"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
)

func callApi(args []string, config *configuration.Configuration, endpointRepo *testapi.FakeEndpointRepo) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)

	cmd := NewApi(ui, config, endpointRepo)
	ctxt := testcmd.NewContext("api", args)
	reqFactory := &testreq.FakeReqFactory{}
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
func init() {
	Describe("Testing with ginkgo", func() {
		It("TestApiWithoutArgument", func() {
			config := &configuration.Configuration{
				Target:     "https://api.run.pivotal.io",
				ApiVersion: "2.0",
			}
			endpointRepo := &testapi.FakeEndpointRepo{}

			ui := callApi([]string{}, config, endpointRepo)

			assert.Equal(mr.T(), len(ui.Outputs), 1)
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"https://api.run.pivotal.io", "2.0"},
			})
		})
		It("TestApiWhenChangingTheEndpoint", func() {

			endpointRepo := &testapi.FakeEndpointRepo{}
			config := &configuration.Configuration{}

			ui := callApi([]string{"http://example.com"}, config, endpointRepo)

			assert.Equal(mr.T(), endpointRepo.UpdateEndpointReceived, "http://example.com")
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Setting api endpoint to", "example.com"},
				{"OK"},
			})
		})
		It("TestApiWithTrailingSlash", func() {

			endpointRepo := &testapi.FakeEndpointRepo{}
			config := &configuration.Configuration{}

			ui := callApi([]string{"https://example.com/"}, config, endpointRepo)

			assert.Equal(mr.T(), endpointRepo.UpdateEndpointReceived, "https://example.com")
			testassert.SliceContains(mr.T(), ui.Outputs, testassert.Lines{
				{"Setting api endpoint to", "example.com"},
				{"OK"},
			})
		})
	})
}
