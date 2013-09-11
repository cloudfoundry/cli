package commands_test

import (
	/*	"cf"*/
	/*	"cf/api"*/
	. "cf/commands"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestCreateOrganization(t *testing.T) {
	fakeUI := callCreateOrganization(
		[]string{"--name", "my-org"},
	)

	assert.Equal(t, len(fakeUI.Outputs), 2)
	assert.Contains(t, fakeUI.Outputs[0], "Creating organization")
	assert.Contains(t, fakeUI.Outputs[0], "my-org")
	assert.Contains(t, fakeUI.Outputs[1], "OK")
}

func callCreateOrganization(args []string) (fakeUI *testhelpers.FakeUI) {
	fakeUI = new(testhelpers.FakeUI)
	ctxt := testhelpers.NewContext("create-org", args)
	orgRepo := &testhelpers.FakeOrgRepository{}
	cmd := NewCreateOrganization(fakeUI, orgRepo)
	reqFactory := &testhelpers.FakeReqFactory{}

	testhelpers.RunCommand(cmd, ctxt, reqFactory)
	return
}
