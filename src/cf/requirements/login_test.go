package requirements

import (
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testterm "testhelpers/terminal"
	"testing"
)

func TestLoginRequirement(t *testing.T) {
	ui := new(testterm.FakeUI)
	config := &configuration.Configuration{
		AccessToken: "foo bar token",
	}

	req := NewLoginRequirement(ui, config)
	success := req.Execute()
	assert.True(t, success)

	config = &configuration.Configuration{
		AccessToken: "",
	}

	req = NewLoginRequirement(ui, config)
	success = req.Execute()
	assert.False(t, success)
	assert.Contains(t, ui.Outputs[0], "Not logged in.")
}
