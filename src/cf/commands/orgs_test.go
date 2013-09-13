package commands_test

import (
	"cf"
	. "cf/commands"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestOrgsRequirements(t *testing.T) {
	orgRepo := &testhelpers.FakeOrgRepository{}

	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true}
	callOrgs(reqFactory, orgRepo)
	assert.True(t, testhelpers.CommandDidPassRequirements)

	reqFactory = &testhelpers.FakeReqFactory{LoginSuccess: false}
	callOrgs(reqFactory, orgRepo)
	assert.False(t, testhelpers.CommandDidPassRequirements)
}

func TestOrgs(t *testing.T) {
	orgs := []cf.Organization{
		cf.Organization{Name: "Organization-1"},
		cf.Organization{Name: "Organization-2"},
	}
	orgRepo := &testhelpers.FakeOrgRepository{
		Organizations: orgs,
	}

	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true}

	ui := callOrgs(reqFactory, orgRepo)

	assert.Contains(t, ui.Outputs[0], "Getting organizations")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[2], "Organization-1")
	assert.Contains(t, ui.Outputs[3], "Organization-2")
}

func callOrgs(reqFactory *testhelpers.FakeReqFactory, orgRepo *testhelpers.FakeOrgRepository) (fakeUI *testhelpers.FakeUI) {
	fakeUI = &testhelpers.FakeUI{}
	ctxt := testhelpers.NewContext("orgs", []string{})
	cmd := NewListOrganizations(fakeUI, orgRepo)
	testhelpers.RunCommand(cmd, ctxt, reqFactory)
	return
}
