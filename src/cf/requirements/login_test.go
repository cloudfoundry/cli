package requirements

import (
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testassert "testhelpers/assert"
	testterm "testhelpers/terminal"
	"testing"
)

func TestLoginRequirement(t *testing.T) {
	ui := new(testterm.FakeUI)
	config := &configuration.Configuration{
		AccessToken: "foo bar token",
	}

	req := newLoginRequirement(ui, config)
	success := req.Execute()
	assert.True(t, success)

	config = &configuration.Configuration{
		AccessToken: "",
	}

	req = newLoginRequirement(ui, config)
	success = req.Execute()
	assert.False(t, success)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{{"Not logged in."}})
}
