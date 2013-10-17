package organization_test

import (
	"cf"
	. "cf/commands/organization"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestListOrgsRequirements(t *testing.T) {
	orgRepo := &testapi.FakeOrgRepository{}
	config := &configuration.Configuration{}

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	callListOrgs(config, reqFactory, orgRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: false}
	callListOrgs(config, reqFactory, orgRepo)
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

	tokenInfo := configuration.TokenInfo{Username: "my-user"}
	accessToken, err := testconfig.CreateAccessTokenWithTokenInfo(tokenInfo)
	assert.NoError(t, err)
	config := &configuration.Configuration{AccessToken: accessToken}

	ui := callListOrgs(config, reqFactory, orgRepo)

	assert.Contains(t, ui.Outputs[0], "Getting orgs as")
	assert.Contains(t, ui.Outputs[0], "my-user")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[3], "Organization-1")
	assert.Contains(t, ui.Outputs[4], "Organization-2")
}

func callListOrgs(config *configuration.Configuration, reqFactory *testreq.FakeReqFactory, orgRepo *testapi.FakeOrgRepository) (fakeUI *testterm.FakeUI) {
	fakeUI = &testterm.FakeUI{}
	ctxt := testcmd.NewContext("orgs", []string{})
	cmd := NewListOrgs(fakeUI, config, orgRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
