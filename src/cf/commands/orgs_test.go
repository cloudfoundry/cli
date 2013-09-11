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
	orgRepo := &testhelpers.FakeOrgRepository{
		Organizations: orgs,
	}

	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true}

	ui := callOrgs(orgRepo, reqFactory)

	assert.True(t, testhelpers.CommandDidPassRequirements)

	assert.Contains(t, ui.Outputs[0], "Getting organizations")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[2], "name")
	assert.Contains(t, ui.Outputs[3], "Organization-1")
	assert.Contains(t, ui.Outputs[4], "Organization-2")
}

func callOrgs(orgRepo *testhelpers.FakeOrgRepository, reqFactory *testhelpers.FakeReqFactory) (fakeUI *testhelpers.FakeUI) {
	fakeUI = &testhelpers.FakeUI{}
	ctxt := testhelpers.NewContext("orgs", []string{})
	cmd := NewListOrganizations(fakeUI, orgRepo)
	testhelpers.RunCommand(cmd, ctxt, reqFactory)
	return
}
