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

func TestDeleteServiceAuthTokenFailsWithUsage(t *testing.T) {
	authTokenRepo := &testapi.FakeAuthTokenRepo{}
	reqFactory := &testreq.FakeReqFactory{}

	ui := callDeleteServiceAuthToken([]string{}, reqFactory, authTokenRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callDeleteServiceAuthToken([]string{"arg1"}, reqFactory, authTokenRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callDeleteServiceAuthToken([]string{"arg1", "arg2"}, reqFactory, authTokenRepo)
	assert.False(t, ui.FailedWithUsage)
}

func TestDeleteServiceAuthTokenRequirements(t *testing.T) {
	authTokenRepo := &testapi.FakeAuthTokenRepo{}
	reqFactory := &testreq.FakeReqFactory{}
	args := []string{"arg1", "arg2"}

	reqFactory.LoginSuccess = true
	callDeleteServiceAuthToken(args, reqFactory, authTokenRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)

	reqFactory.LoginSuccess = false
	callDeleteServiceAuthToken(args, reqFactory, authTokenRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)
}

func TestDeleteServiceAuthToken(t *testing.T) {
	expectedToken := cf.ServiceAuthToken{
		Label:    "a label",
		Provider: "a provider",
	}
	authTokenRepo := &testapi.FakeAuthTokenRepo{
		FindByNameServiceAuthToken: expectedToken,
	}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	args := []string{"a label", "a provider"}

	ui := callDeleteServiceAuthToken(args, reqFactory, authTokenRepo)
	assert.Contains(t, ui.Outputs[0], "Deleting service auth token...")

	assert.Equal(t, authTokenRepo.DeletedServiceAuthToken, expectedToken)

	assert.Contains(t, ui.Outputs[1], "OK")
}

func callDeleteServiceAuthToken(args []string, reqFactory *testreq.FakeReqFactory, authTokenRepo *testapi.FakeAuthTokenRepo) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	cmd := NewDeleteServiceAuthToken(ui, authTokenRepo)
	ctxt := testcmd.NewContext("delete-service-auth-token", args)

	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
