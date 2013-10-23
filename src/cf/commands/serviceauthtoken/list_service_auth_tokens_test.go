package serviceauthtoken_test

import (
	"cf"
	. "cf/commands/serviceauthtoken"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
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

	authTokenRepo.FindAllAuthTokens = []cf.ServiceAuthToken{
		cf.ServiceAuthToken{Label: "a label", Provider: "a provider"},
		cf.ServiceAuthToken{Label: "a second label", Provider: "a second provider"},
	}

	ui := callListServiceAuthTokens(t, reqFactory, authTokenRepo)
	assert.Contains(t, ui.Outputs[0], "Getting service auth tokens as")
	assert.Contains(t, ui.Outputs[0], "my-user")
	assert.Contains(t, ui.Outputs[1], "OK")

	assert.Contains(t, ui.Outputs[3], "label")
	assert.Contains(t, ui.Outputs[3], "provider")

	assert.Contains(t, ui.Outputs[4], "a label")
	assert.Contains(t, ui.Outputs[4], "a provider")

	assert.Contains(t, ui.Outputs[5], "a second label")
	assert.Contains(t, ui.Outputs[5], "a second provider")
}

func callListServiceAuthTokens(t *testing.T, reqFactory *testreq.FakeReqFactory, authTokenRepo *testapi.FakeAuthTokenRepo) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)

	config := &configuration.Configuration{
		Space:        cf.Space{Name: "my-space"},
		Organization: cf.Organization{Name: "my-org"},
		AccessToken:  token,
	}

	cmd := NewListServiceAuthTokens(ui, config, authTokenRepo)
	ctxt := testcmd.NewContext("service-auth-tokens", []string{})
	testcmd.RunCommand(cmd, ctxt, reqFactory)

	return
}
