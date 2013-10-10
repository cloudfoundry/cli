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

func TestCreateServiceAuthTokenFailsWithUsage(t *testing.T) {
	authTokenRepo := &testapi.FakeAuthTokenRepo{}
	reqFactory := &testreq.FakeReqFactory{}

	ui := callCreateServiceAuthToken([]string{}, reqFactory, authTokenRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callCreateServiceAuthToken([]string{"arg1"}, reqFactory, authTokenRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callCreateServiceAuthToken([]string{"arg1", "arg2"}, reqFactory, authTokenRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callCreateServiceAuthToken([]string{"arg1", "arg2", "arg3"}, reqFactory, authTokenRepo)
	assert.False(t, ui.FailedWithUsage)
}

func TestCreateServiceAuthTokenRequirements(t *testing.T) {
	authTokenRepo := &testapi.FakeAuthTokenRepo{}
	reqFactory := &testreq.FakeReqFactory{}
	args := []string{"arg1", "arg2", "arg3"}

	reqFactory.LoginSuccess = true
	callCreateServiceAuthToken(args, reqFactory, authTokenRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)

	reqFactory.LoginSuccess = false
	callCreateServiceAuthToken(args, reqFactory, authTokenRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)
}

func TestCreateServiceAuthToken(t *testing.T) {
	authTokenRepo := &testapi.FakeAuthTokenRepo{}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	args := []string{"a label", "a provider", "a value"}

	ui := callCreateServiceAuthToken(args, reqFactory, authTokenRepo)
	assert.Contains(t, ui.Outputs[0], "Creating service auth token...")

	assert.Equal(t, authTokenRepo.CreatedServiceAuthToken, cf.ServiceAuthToken{
		Label:    "a label",
		Provider: "a provider",
		Token:    "a value",
	})

	assert.Contains(t, ui.Outputs[1], "OK")
}

func callCreateServiceAuthToken(args []string, reqFactory *testreq.FakeReqFactory, authTokenRepo *testapi.FakeAuthTokenRepo) (ui *testterm.FakeUI) {
	ui = new(testterm.FakeUI)
	cmd := NewCreateServiceAuthToken(ui, authTokenRepo)
	ctxt := testcmd.NewContext("create-service-auth-token", args)

	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
