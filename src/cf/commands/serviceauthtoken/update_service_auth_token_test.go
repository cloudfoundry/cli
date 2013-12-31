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
	foundAuthToken := cf.ServiceAuthTokenFields{}
	foundAuthToken.Guid = "found-auth-token-guid"
	foundAuthToken.Label = "found label"
	foundAuthToken.Provider = "found provider"

	authTokenRepo := &testapi.FakeAuthTokenRepo{FindByLabelAndProviderServiceAuthTokenFields: foundAuthToken}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	args := []string{"a label", "a provider", "a value"}

	ui := callUpdateServiceAuthToken(t, args, reqFactory, authTokenRepo)
	expectedAuthToken := cf.ServiceAuthTokenFields{}
	expectedAuthToken.Guid = "found-auth-token-guid"
	expectedAuthToken.Label = "found label"
	expectedAuthToken.Provider = "found provider"
	expectedAuthToken.Token = "a value"

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Updating service auth token as", "my-user"},
		{"OK"},
	})

	assert.Equal(t, authTokenRepo.FindByLabelAndProviderLabel, "a label")
	assert.Equal(t, authTokenRepo.FindByLabelAndProviderProvider, "a provider")
	assert.Equal(t, authTokenRepo.UpdatedServiceAuthTokenFields, expectedAuthToken)
	assert.Equal(t, authTokenRepo.UpdatedServiceAuthTokenFields, expectedAuthToken)
}

func callUpdateServiceAuthToken(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory, authTokenRepo *testapi.FakeAuthTokenRepo) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)
	space := cf.SpaceFields{}
	space.Name = "my-space"
	org := cf.OrganizationFields{}
	org.Name = "my-org"
	config := &configuration.Configuration{
		SpaceFields:        space,
		OrganizationFields: org,
		AccessToken:        token,
	}

	cmd := NewUpdateServiceAuthToken(ui, config, authTokenRepo)
	ctxt := testcmd.NewContext("update-service-auth-token", args)

	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
