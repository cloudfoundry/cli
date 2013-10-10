package serviceauthtoken_test

import (
	"cf"
	. "cf/commands/serviceauthtoken"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestUpdateServiceAuthTokenFailsWithUsage(t *testing.T) {
	authTokenRepo := &testapi.FakeAuthTokenRepo{}
	reqFactory := &testreq.FakeReqFactory{}

	ui := callUpdateServiceAuthToken([]string{}, reqFactory, authTokenRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callUpdateServiceAuthToken([]string{"MY-TOKEN-LABEL"}, reqFactory, authTokenRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callUpdateServiceAuthToken([]string{"MY-TOKEN-LABEL", "my-token-abc123"}, reqFactory, authTokenRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callUpdateServiceAuthToken([]string{"MY-TOKEN-LABEL", "my-provider", "my-token-abc123"}, reqFactory, authTokenRepo)
	assert.False(t, ui.FailedWithUsage)
}

func TestUpdateServiceAuthTokenRequirements(t *testing.T) {
	authTokenRepo := &testapi.FakeAuthTokenRepo{}
	reqFactory := &testreq.FakeReqFactory{}
	args := []string{"MY-TOKEN-LABLE", "my-provider", "my-token-abc123"}

	reqFactory.LoginSuccess = true
	callUpdateServiceAuthToken(args, reqFactory, authTokenRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)

	reqFactory.LoginSuccess = false
	callUpdateServiceAuthToken(args, reqFactory, authTokenRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)
}

func TestUpdateServiceAuthToken(t *testing.T) {
	authTokenRepo := &testapi.FakeAuthTokenRepo{}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	args := []string{"a label", "a provider", "a value"}

	ui := callUpdateServiceAuthToken(args, reqFactory, authTokenRepo)
	expectedAuthToken := cf.ServiceAuthToken{
		Label:    "a label",
		Provider: "a provider",
		Token:    "a value",
	}
	assert.Equal(t, ui.Outputs[0], "Updating service auth token...")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Equal(t, authTokenRepo.UpdatedServiceAuthToken, expectedAuthToken)
}

func callUpdateServiceAuthToken(args []string, reqFactory *testreq.FakeReqFactory, authTokenRepo *testapi.FakeAuthTokenRepo) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	cmd := NewUpdateServiceAuthToken(ui, authTokenRepo)
	ctxt := testcmd.NewContext("update-service-auth-token", args)

	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
