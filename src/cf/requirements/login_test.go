package requirements_test

import (
	"cf/configuration"
	. "cf/requirements"
	"testing"
	"testhelpers"
	"github.com/stretchr/testify/assert"
)

func TestLoginRequirement(t *testing.T) {
	ui := new(testhelpers.FakeUI)
	config := &configuration.Configuration{
		AccessToken: "foo bar token",
	}

	req := NewLoginRequirement(ui, config)
	err := req.Execute()
	assert.NoError(t, err)

	config = &configuration.Configuration{
		AccessToken: "",
	}

	req = NewLoginRequirement(ui, config)
	err = req.Execute()
	assert.Error(t, err)
	assert.Contains(t, ui.Outputs[0], "Not logged in.")
}
