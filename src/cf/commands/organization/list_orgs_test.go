package organization_test

import (
	"cf"
	. "cf/commands/organization"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestListOrgsRequirements(t *testing.T) {
	orgRepo := &testapi.FakeOrgRepository{}

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	callListOrgs(reqFactory, orgRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: false}
	callListOrgs(reqFactory, orgRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)
}

func TestListOrgs(t *testing.T) {
	orgs := []cf.Organization{
		cf.Organization{Name: "Organization-1"},
		cf.Organization{Name: "Organization-2"},
	}
	orgRepo := &testapi.FakeOrgRepository{
		Organizations: orgs,
	}

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}

	ui := callListOrgs(reqFactory, orgRepo)

	assert.Contains(t, ui.Outputs[0], "Getting orgs")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[2], "Organization-1")
	assert.Contains(t, ui.Outputs[3], "Organization-2")
}

func callListOrgs(reqFactory *testreq.FakeReqFactory, orgRepo *testapi.FakeOrgRepository) (fakeUI *testterm.FakeUI) {
	fakeUI = &testterm.FakeUI{}
	ctxt := testcmd.NewContext("orgs", []string{})
	cmd := NewListOrgs(fakeUI, orgRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
