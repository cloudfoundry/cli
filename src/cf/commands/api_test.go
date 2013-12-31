package commands_test

import (
	. "cf/commands"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestApiWithoutArgument(t *testing.T) {
	config := &configuration.Configuration{
		Target:     "https://api.run.pivotal.io",
		ApiVersion: "2.0",
	}
	endpointRepo := &testapi.FakeEndpointRepo{}

	ui := callApi([]string{}, config, endpointRepo)

	assert.Equal(t, len(ui.Outputs), 1)
	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"https://api.run.pivotal.io", "2.0"},
	})
}

func TestApiWhenChangingTheEndpoint(t *testing.T) {
	endpointRepo := &testapi.FakeEndpointRepo{}
	config := &configuration.Configuration{}

	ui := callApi([]string{"http://example.com"}, config, endpointRepo)

	assert.Equal(t, endpointRepo.UpdateEndpointEndpoint, "http://example.com")
	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Setting api endpoint to", "example.com"},
		{"OK"},
	})
}

func TestApiWithTrailingSlash(t *testing.T) {
	endpointRepo := &testapi.FakeEndpointRepo{}
	config := &configuration.Configuration{}

	ui := callApi([]string{"https://example.com/"}, config, endpointRepo)

	assert.Equal(t, endpointRepo.UpdateEndpointEndpoint, "https://example.com")
	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Setting api endpoint to", "example.com"},
		{"OK"},
	})
}

func callApi(args []string, config *configuration.Configuration, endpointRepo *testapi.FakeEndpointRepo) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)

	cmd := NewApi(ui, config, endpointRepo)
	ctxt := testcmd.NewContext("api", args)
	reqFactory := &testreq.FakeReqFactory{}
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
