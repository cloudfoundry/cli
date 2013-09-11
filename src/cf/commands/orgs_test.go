package commands_test

import (
	"cf"
	. "cf/commands"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestOrgs(t *testing.T) {
	orgs := []cf.Organization{
		cf.Organization{Name: "Organization-1"},
		cf.Organization{Name: "Organization-2"},
	}

	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true}

	ui := callOrgs(reqFactory)

	assert.True(t, testhelpers.CommandDidPassRequirements)

	assert.Contains(t, ui.Outputs[0], "Getting organizations")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[2], "Name")
	assert.Contains(t, ui.Outputs[3], "Organization-1")
	assert.Contains(t, ui.Outputs[4], "Organization-2")
}

func callOrgs(reqFactory *testhelpers.FakeReqFactory) (ui *testhelpers.FakeUI) {
	ui = &testhelpers.FakeUI{}
	ctxt := testhelpers.NewContext("orgs", []string{})
	cmd := NewCreateOrganization(fakeUI)
	testhelpers.RunCommand(cmd, ctxt, reqFactory)
	return
}
