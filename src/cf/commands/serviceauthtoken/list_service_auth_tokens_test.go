package serviceauthtoken_test

import (
	"cf"
	. "cf/commands/serviceauthtoken"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestListServiceAuthTokensRequirements(t *testing.T) {
	authTokenRepo := &testhelpers.FakeAuthTokenRepo{}
	reqFactory := &testhelpers.FakeReqFactory{}

	reqFactory.LoginSuccess = false
	callListServiceAuthTokens(reqFactory, authTokenRepo)
	assert.False(t, testhelpers.CommandDidPassRequirements)

	reqFactory.LoginSuccess = true
	callListServiceAuthTokens(reqFactory, authTokenRepo)
	assert.True(t, testhelpers.CommandDidPassRequirements)
}

func TestListServiceAuthTokens(t *testing.T) {
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true}
	authTokenRepo := &testhelpers.FakeAuthTokenRepo{}

	authTokenRepo.FindAllAuthTokens = []cf.ServiceAuthToken{
		cf.ServiceAuthToken{Label: "a label", Provider: "a provider"},
		cf.ServiceAuthToken{Label: "a second label", Provider: "a second provider"},
	}

	ui := callListServiceAuthTokens(reqFactory, authTokenRepo)
	assert.Contains(t, ui.Outputs[0], "Getting service auth tokens")
	assert.Contains(t, ui.Outputs[1], "OK")

	assert.Contains(t, ui.Outputs[3], "label")
	assert.Contains(t, ui.Outputs[3], "provider")

	assert.Contains(t, ui.Outputs[4], "a label")
	assert.Contains(t, ui.Outputs[4], "a provider")

	assert.Contains(t, ui.Outputs[5], "a second label")
	assert.Contains(t, ui.Outputs[5], "a second provider")
}

func callListServiceAuthTokens(reqFactory *testhelpers.FakeReqFactory, authTokenRepo *testhelpers.FakeAuthTokenRepo) (ui *testhelpers.FakeUI) {
	ui = &testhelpers.FakeUI{}
	cmd := NewListServiceAuthTokens(ui, authTokenRepo)
	ctxt := testhelpers.NewContext("service-auth-tokens", []string{})
	testhelpers.RunCommand(cmd, ctxt, reqFactory)

	return
}
