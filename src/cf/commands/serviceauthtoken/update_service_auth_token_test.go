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

func TestUpdateServiceAuthTokenFailsWithUsage(t *testing.T) {
	authTokenRepo := &testapi.FakeAuthTokenRepo{}
	reqFactory := &testreq.FakeReqFactory{}

	ui := callUpdateServiceAuthToken(t, []string{}, reqFactory, authTokenRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callUpdateServiceAuthToken(t, []string{"MY-TOKEN-LABEL"}, reqFactory, authTokenRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callUpdateServiceAuthToken(t, []string{"MY-TOKEN-LABEL", "my-token-abc123"}, reqFactory, authTokenRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callUpdateServiceAuthToken(t, []string{"MY-TOKEN-LABEL", "my-provider", "my-token-abc123"}, reqFactory, authTokenRepo)
	assert.False(t, ui.FailedWithUsage)
}

func TestUpdateServiceAuthTokenRequirements(t *testing.T) {
	authTokenRepo := &testapi.FakeAuthTokenRepo{}
	reqFactory := &testreq.FakeReqFactory{}
	args := []string{"MY-TOKEN-LABLE", "my-provider", "my-token-abc123"}

	reqFactory.LoginSuccess = true
	callUpdateServiceAuthToken(t, args, reqFactory, authTokenRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)

	reqFactory.LoginSuccess = false
	callUpdateServiceAuthToken(t, args, reqFactory, authTokenRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)
}

func TestUpdateServiceAuthToken(t *testing.T) {
	foundAuthToken := cf.ServiceAuthToken{
		Guid:     "found-auth-token-guid",
		Label:    "found label",
		Provider: "found provider",
	}
	authTokenRepo := &testapi.FakeAuthTokenRepo{FindByLabelAndProviderServiceAuthToken: foundAuthToken}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	args := []string{"a label", "a provider", "a value"}

	ui := callUpdateServiceAuthToken(t, args, reqFactory, authTokenRepo)
	expectedAuthToken := cf.ServiceAuthToken{
		Guid:     "found-auth-token-guid",
		Label:    "found label",
		Provider: "found provider",
		Token:    "a value",
	}
	assert.Contains(t, ui.Outputs[0], "Updating service auth token as")
	assert.Contains(t, ui.Outputs[0], "my-user")
	assert.Contains(t, ui.Outputs[1], "OK")

	assert.Equal(t, authTokenRepo.FindByLabelAndProviderLabel, "a label")
	assert.Equal(t, authTokenRepo.FindByLabelAndProviderProvider, "a provider")
	assert.Equal(t, authTokenRepo.UpdatedServiceAuthToken, expectedAuthToken)
	assert.Equal(t, authTokenRepo.UpdatedServiceAuthToken, expectedAuthToken)
}

func callUpdateServiceAuthToken(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory, authTokenRepo *testapi.FakeAuthTokenRepo) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)

	config := &configuration.Configuration{
		Space:        cf.Space{Name: "my-space"},
		Organization: cf.Organization{Name: "my-org"},
		AccessToken:  token,
	}

	cmd := NewUpdateServiceAuthToken(ui, config, authTokenRepo)
	ctxt := testcmd.NewContext("update-service-auth-token", args)

	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
