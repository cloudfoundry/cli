package serviceauthtoken_test

import (
	"cf"
	. "cf/commands/serviceauthtoken"
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

func TestListServiceAuthTokensRequirements(t *testing.T) {
	authTokenRepo := &testapi.FakeAuthTokenRepo{}
	reqFactory := &testreq.FakeReqFactory{}

	reqFactory.LoginSuccess = false
	callListServiceAuthTokens(t, reqFactory, authTokenRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory.LoginSuccess = true
	callListServiceAuthTokens(t, reqFactory, authTokenRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)
}

func TestListServiceAuthTokens(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	authTokenRepo := &testapi.FakeAuthTokenRepo{}
	authToken := cf.ServiceAuthTokenFields{}
	authToken.Label = "a label"
	authToken.Provider = "a provider"
	authToken2 := cf.ServiceAuthTokenFields{}
	authToken2.Label = "a second label"
	authToken2.Provider = "a second provider"
	authTokenRepo.FindAllAuthTokens = []cf.ServiceAuthTokenFields{authToken, authToken2}

	ui := callListServiceAuthTokens(t, reqFactory, authTokenRepo)
	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Getting service auth tokens as", "my-user"},
		{"OK"},
		{"label", "provider"},
		{"a label", "a provider"},
		{"a second label", "a second provider"},
	})
}

func callListServiceAuthTokens(t *testing.T, reqFactory *testreq.FakeReqFactory, authTokenRepo *testapi.FakeAuthTokenRepo) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)
	org := cf.OrganizationFields{}
	org.Name = "my-org"
	space := cf.SpaceFields{}
	space.Name = "my-space"
	config := &configuration.Configuration{
		SpaceFields:        space,
		OrganizationFields: org,
		AccessToken:        token,
	}

	cmd := NewListServiceAuthTokens(ui, config, authTokenRepo)
	ctxt := testcmd.NewContext("service-auth-tokens", []string{})
	testcmd.RunCommand(cmd, ctxt, reqFactory)

	return
}
