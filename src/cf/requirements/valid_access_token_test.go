package requirements

import (
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testterm "testhelpers/terminal"
	"testing"
)

func TestValidAccessRequirement(t *testing.T) {
	ui := new(testterm.FakeUI)
	appRepo := &testapi.FakeApplicationRepository{
		ReadAuthErr: true,
	}

	req := newValidAccessTokenRequirement(ui, appRepo)
	success := req.Execute()
	assert.False(t, success)
	testassert.SliceContains(t, ui.Outputs, testassert.Lines{{"Not logged in."}})

	appRepo.ReadAuthErr = false

	req = newValidAccessTokenRequirement(ui, appRepo)
	success = req.Execute()
	assert.True(t, success)
}
