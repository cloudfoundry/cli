package organization_test

import (
	"cf"
	. "cf/commands/organization"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
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

func TestListAllPagesOfOrgs(t *testing.T) {
	orgs := []cf.Organization{
		cf.Organization{Name: "Organization-1"},
		cf.Organization{Name: "Organization-2"},
		cf.Organization{Name: "Organization-3"},
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

	testassert.SliceContains(t, ui.Outputs, []string{
		"Getting orgs as my-user",
		"Organization-1",
		"Organization-2",
		"Organization-3",
	})
}

func TestListNoOrgs(t *testing.T) {
	orgs := []cf.Organization{}
	orgRepo := &testapi.FakeOrgRepository{
		Organizations: orgs,
	}

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}

	tokenInfo := configuration.TokenInfo{Username: "my-user"}
	accessToken, err := testconfig.CreateAccessTokenWithTokenInfo(tokenInfo)
	assert.NoError(t, err)
	config := &configuration.Configuration{AccessToken: accessToken}

	ui := callListOrgs(config, reqFactory, orgRepo)

	testassert.SliceContains(t, ui.Outputs, []string{
		"Getting orgs as my-user",
		"No orgs found",
	})
}

func callListOrgs(config *configuration.Configuration, reqFactory *testreq.FakeReqFactory, orgRepo *testapi.FakeOrgRepository) (fakeUI *testterm.FakeUI) {
	fakeUI = &testterm.FakeUI{}
	ctxt := testcmd.NewContext("orgs", []string{})
	cmd := NewListOrgs(fakeUI, config, orgRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
