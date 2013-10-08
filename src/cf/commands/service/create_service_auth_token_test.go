package service_test

import (
	"cf"
	. "cf/commands/service"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestCreateServiceAuthTokenFailsWithUsage(t *testing.T) {
	authTokenRepo := &testhelpers.FakeAuthTokenRepo{}
	reqFactory := &testhelpers.FakeReqFactory{}

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
	authTokenRepo := &testhelpers.FakeAuthTokenRepo{}
	reqFactory := &testhelpers.FakeReqFactory{}
	args := []string{"arg1", "arg2", "arg3"}

	reqFactory.LoginSuccess = true
	callCreateServiceAuthToken(args, reqFactory, authTokenRepo)
	assert.True(t, testhelpers.CommandDidPassRequirements)

	reqFactory.LoginSuccess = false
	callCreateServiceAuthToken(args, reqFactory, authTokenRepo)
	assert.False(t, testhelpers.CommandDidPassRequirements)
}

func TestCreateServiceAuthToken(t *testing.T) {
	authTokenRepo := &testhelpers.FakeAuthTokenRepo{}
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true}
	args := []string{"a label", "a provider", "a value"}

	ui := callCreateServiceAuthToken(args, reqFactory, authTokenRepo)
	assert.Contains(t, ui.Outputs[0], "Creating service auth token...")

	assert.Equal(t, authTokenRepo.CreatedServiceAuthToken, cf.ServiceAuthToken{
		Label:    "a label",
		Provider: "a provider",
		Value:    "a value",
	})

	assert.Contains(t, ui.Outputs[1], "OK")
}

func callCreateServiceAuthToken(args []string, reqFactory *testhelpers.FakeReqFactory, authTokenRepo *testhelpers.FakeAuthTokenRepo) (ui *testhelpers.FakeUI) {
	ui = new(testhelpers.FakeUI)
	cmd := NewCreateServiceAuthToken(ui, authTokenRepo)
	ctxt := testhelpers.NewContext("create-service-auth-token", args)

	testhelpers.RunCommand(cmd, ctxt, reqFactory)
	return
}
