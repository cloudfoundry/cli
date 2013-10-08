package service_test

import (
	. "cf/commands/service"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
	"cf"
)

func TestUpdateServiceAuthTokenFailsWithUsage(t *testing.T) {
	authTokenRepo := &testhelpers.FakeAuthTokenRepo{}
	reqFactory := &testhelpers.FakeReqFactory{}

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
	authTokenRepo := &testhelpers.FakeAuthTokenRepo{}
	reqFactory := &testhelpers.FakeReqFactory{}
	args := []string{"MY-TOKEN-LABLE", "my-provider", "my-token-abc123"}

	reqFactory.LoginSuccess = true
	callUpdateServiceAuthToken(args, reqFactory, authTokenRepo)
	assert.True(t, testhelpers.CommandDidPassRequirements)

	reqFactory.LoginSuccess = false
	callUpdateServiceAuthToken(args, reqFactory, authTokenRepo)
	assert.False(t, testhelpers.CommandDidPassRequirements)
}

func TestUpdateServiceAuthToken(t *testing.T) {
	authTokenRepo := &testhelpers.FakeAuthTokenRepo{}
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true}
	args := []string{"a label", "a provider", "a value"}

	ui := callUpdateServiceAuthToken(args, reqFactory, authTokenRepo)
	expectedAuthToken := cf.ServiceAuthToken{
		Label:    "a label",
		Provider: "a provider",
		Value:    "a value",
	}
	assert.Equal(t, ui.Outputs[0], "Updating service auth token...")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Equal(t, authTokenRepo.UpdatedServiceAuthToken, expectedAuthToken)
}

func callUpdateServiceAuthToken(args []string, reqFactory *testhelpers.FakeReqFactory, authTokenRepo *testhelpers.FakeAuthTokenRepo) (ui *testhelpers.FakeUI) {
	ui = new(testhelpers.FakeUI)
	cmd := NewUpdateServiceAuthToken(ui, authTokenRepo)
	ctxt := testhelpers.NewContext("update-service-auth-token", args)

	testhelpers.RunCommand(cmd, ctxt, reqFactory)
	return
}
