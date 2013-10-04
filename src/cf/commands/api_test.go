package commands_test

import (
	. "cf/commands"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestApiWithoutArgument(t *testing.T) {
	config := &configuration.Configuration{
		Target:     "https://api.run.pivotal.io",
		ApiVersion: "2.0",
	}
	endpointRepo := &testhelpers.FakeEndpointRepo{}

	ui := callApi([]string{}, config, endpointRepo)

	assert.Equal(t, len(ui.Outputs), 1)
	assert.Contains(t, ui.Outputs[0], "https://api.run.pivotal.io")
	assert.Contains(t, ui.Outputs[0], "2.0")
}

func TestApiWhenChangingTheEndpoint(t *testing.T) {
	endpointRepo := &testhelpers.FakeEndpointRepo{}
	config := &configuration.Configuration{}

	ui := callApi([]string{"http://example.com"}, config, endpointRepo)

	assert.Contains(t, ui.Outputs[0], "Setting api endpoint to")
	assert.Equal(t, endpointRepo.UpdateEndpointEndpoint, "http://example.com")
	assert.Contains(t, ui.Outputs[1], "OK")
}

func callApi(args []string, config *configuration.Configuration, endpointRepo *testhelpers.FakeEndpointRepo) (ui *testhelpers.FakeUI) {
	ui = new(testhelpers.FakeUI)

	cmd := NewApi(ui, config, endpointRepo)
	ctxt := testhelpers.NewContext("api", args)
	reqFactory := &testhelpers.FakeReqFactory{}
	testhelpers.RunCommand(cmd, ctxt, reqFactory)
	return
}
