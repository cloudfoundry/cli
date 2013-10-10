package serviceauthtoken_test

import (
	"cf"
	. "cf/commands/serviceauthtoken"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestDeleteServiceAuthTokenFailsWithUsage(t *testing.T) {
	authTokenRepo := &testhelpers.FakeAuthTokenRepo{}
	reqFactory := &testhelpers.FakeReqFactory{}

	ui := callDeleteServiceAuthToken([]string{}, reqFactory, authTokenRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callDeleteServiceAuthToken([]string{"arg1"}, reqFactory, authTokenRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callDeleteServiceAuthToken([]string{"arg1", "arg2"}, reqFactory, authTokenRepo)
	assert.False(t, ui.FailedWithUsage)
}

func TestDeleteServiceAuthTokenRequirements(t *testing.T) {
	authTokenRepo := &testhelpers.FakeAuthTokenRepo{}
	reqFactory := &testhelpers.FakeReqFactory{}
	args := []string{"arg1", "arg2"}

	reqFactory.LoginSuccess = true
	callDeleteServiceAuthToken(args, reqFactory, authTokenRepo)
	assert.True(t, testhelpers.CommandDidPassRequirements)

	reqFactory.LoginSuccess = false
	callDeleteServiceAuthToken(args, reqFactory, authTokenRepo)
	assert.False(t, testhelpers.CommandDidPassRequirements)
}

func TestDeleteServiceAuthToken(t *testing.T) {
	expectedToken := cf.ServiceAuthToken{
		Label:    "a label",
		Provider: "a provider",
	}
	authTokenRepo := &testhelpers.FakeAuthTokenRepo{
		FindByNameServiceAuthToken: expectedToken,
	}
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true}
	args := []string{"a label", "a provider"}

	ui := callDeleteServiceAuthToken(args, reqFactory, authTokenRepo)
	assert.Contains(t, ui.Outputs[0], "Deleting service auth token...")

	assert.Equal(t, authTokenRepo.DeletedServiceAuthToken, expectedToken)

	assert.Contains(t, ui.Outputs[1], "OK")
}

func callDeleteServiceAuthToken(args []string, reqFactory *testhelpers.FakeReqFactory, authTokenRepo *testhelpers.FakeAuthTokenRepo) (ui *testhelpers.FakeUI) {
	ui = new(testhelpers.FakeUI)
	cmd := NewDeleteServiceAuthToken(ui, authTokenRepo)
	ctxt := testhelpers.NewContext("delete-service-auth-token", args)

	testhelpers.RunCommand(cmd, ctxt, reqFactory)
	return
}
