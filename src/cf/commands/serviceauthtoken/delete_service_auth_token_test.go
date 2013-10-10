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
	"cf/net"
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

func TestDeleteServiceAuthTokenWithN(t *testing.T) {
	authTokenRepo := &testapi.FakeAuthTokenRepo{}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	args := []string{"a label", "a provider"}
	ui := &testterm.FakeUI{
		Inputs: []string{"N"},
	}
	ctxt := testcmd.NewContext("delete-service-auth-token", args)
	cmd := NewDeleteServiceAuthToken(ui, authTokenRepo)

	testcmd.RunCommand(cmd, ctxt, reqFactory)

	assert.Contains(t, ui.Prompts[0], "Are you sure you want to delete")
	assert.Contains(t, ui.Prompts[0], "a label a provider")
	assert.Equal(t, len(ui.Outputs), 0)
	assert.Equal(t, authTokenRepo.DeletedServiceAuthToken, cf.ServiceAuthToken{})
}

func TestDeleteServiceAuthTokenWithY(t *testing.T) {
	expectedToken := cf.ServiceAuthToken{
		Label:    "a label",
		Provider: "a provider",
	}
	authTokenRepo := &testapi.FakeAuthTokenRepo{
		FindByNameServiceAuthToken: expectedToken,
	}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	args := []string{"a label", "a provider"}
	ui := &testterm.FakeUI{
		Inputs: []string{"Y"},
	}
	ctxt := testcmd.NewContext("delete-service-auth-token", args)
	cmd := NewDeleteServiceAuthToken(ui, authTokenRepo)

	testcmd.RunCommand(cmd, ctxt, reqFactory)

	assert.Contains(t, ui.Prompts[0], "delete")
	assert.Contains(t, ui.Prompts[0], "a label")
	assert.Contains(t, ui.Prompts[0], "a provider")
	assert.Contains(t, ui.Outputs[0], "Deleting")
	assert.Equal(t, authTokenRepo.DeletedServiceAuthToken, expectedToken)
	assert.Contains(t, ui.Outputs[1], "OK")
}

func TestDeleteServiceAuthTokenWithForce(t *testing.T) {
	expectedToken := cf.ServiceAuthToken{
		Label:    "a label",
		Provider: "a provider",
	}
	authTokenRepo := &testapi.FakeAuthTokenRepo{
		FindByNameServiceAuthToken: expectedToken,
	}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	args := []string{"-f", "a label", "a provider"}
	ui := &testterm.FakeUI{}
	ctxt := testcmd.NewContext("delete-service-auth-token", args)
	cmd := NewDeleteServiceAuthToken(ui, authTokenRepo)

	testcmd.RunCommand(cmd, ctxt, reqFactory)

	assert.Equal(t, len(ui.Prompts), 0)
	assert.Contains(t, ui.Outputs[0], "Deleting")
	assert.Contains(t, ui.Outputs[1], "OK")

	assert.Equal(t, authTokenRepo.DeletedServiceAuthToken, expectedToken)
}

func TestDeleteServiceAuthTokenWhenTokenDoesNotExist(t *testing.T) {
	authTokenRepo := &testapi.FakeAuthTokenRepo{
		FindByNameServiceApiResponse: net.NewNotFoundApiResponse("not found"),
	}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	args := []string{"a label", "a provider"}

	ui := callDeleteServiceAuthToken(args, reqFactory, authTokenRepo)
	assert.Contains(t, ui.Outputs[0], "Deleting service auth token...")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[2], "does not exist")
}

func TestDeleteServiceAuthTokenFailsWithError(t *testing.T) {
	authTokenRepo := &testapi.FakeAuthTokenRepo{
		FindByNameServiceApiResponse: net.NewApiResponseWithMessage("OH NOES"),
	}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	args := []string{"a label", "a provider"}

	ui := callDeleteServiceAuthToken(args, reqFactory, authTokenRepo)
	assert.Contains(t, ui.Outputs[0], "Deleting service auth token...")
	assert.Contains(t, ui.Outputs[1], "FAILED")
	assert.Contains(t, ui.Outputs[2], "OH NOES")
}

func callDeleteServiceAuthToken(args []string, reqFactory *testreq.FakeReqFactory, authTokenRepo *testapi.FakeAuthTokenRepo) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{
		Inputs: []string{"Y"},
	}
	cmd := NewDeleteServiceAuthToken(ui, authTokenRepo)
	ctxt := testcmd.NewContext("delete-service-auth-token", args)

	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
