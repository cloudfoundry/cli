package requirements

import (
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testterm "testhelpers/terminal"
	"testing"
)

func TestValidAccessRequirement(t *testing.T) {
	ui := new(testterm.FakeUI)
	appRepo := &testapi.FakeApplicationRepository{
		FindByNameAuthErr: true,
	}

	req := newValidAccessTokenRequirement(ui, appRepo)
	success := req.Execute()
	assert.False(t, success)
	assert.Contains(t, ui.Outputs[0], "Not logged in.")

	appRepo.FindByNameAuthErr = false

	req = newValidAccessTokenRequirement(ui, appRepo)
	success = req.Execute()
	assert.True(t, success)
}
