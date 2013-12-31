package serviceauthtoken_test

import (
	"cf"
	. "cf/commands/serviceauthtoken"
	"cf/configuration"
	"cf/net"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestDeleteServiceAuthTokenFailsWithUsage(t *testing.T) {
	authTokenRepo := &testapi.FakeAuthTokenRepo{}
	reqFactory := &testreq.FakeReqFactory{}

	ui := callDeleteServiceAuthToken(t, []string{}, []string{"Y"}, reqFactory, authTokenRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callDeleteServiceAuthToken(t, []string{"arg1"}, []string{"Y"}, reqFactory, authTokenRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callDeleteServiceAuthToken(t, []string{"arg1", "arg2"}, []string{"Y"}, reqFactory, authTokenRepo)
	assert.False(t, ui.FailedWithUsage)
}

func TestDeleteServiceAuthTokenRequirements(t *testing.T) {
	authTokenRepo := &testapi.FakeAuthTokenRepo{}
	reqFactory := &testreq.FakeReqFactory{}
	args := []string{"arg1", "arg2"}

	reqFactory.LoginSuccess = true
	callDeleteServiceAuthToken(t, args, []string{"Y"}, reqFactory, authTokenRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)

	reqFactory.LoginSuccess = false
	callDeleteServiceAuthToken(t, args, []string{"Y"}, reqFactory, authTokenRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)
}

func TestDeleteServiceAuthToken(t *testing.T) {
	expectedToken := cf.ServiceAuthTokenFields{}
	expectedToken.Label = "a label"
	expectedToken.Provider = "a provider"

	authTokenRepo := &testapi.FakeAuthTokenRepo{
		FindByLabelAndProviderServiceAuthTokenFields: expectedToken,
	}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	args := []string{"a label", "a provider"}

	ui := callDeleteServiceAuthToken(t, args, []string{"Y"}, reqFactory, authTokenRepo)
	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Deleting service auth token as", "my-user"},
		{"OK"},
	})

	assert.Equal(t, authTokenRepo.FindByLabelAndProviderLabel, "a label")
	assert.Equal(t, authTokenRepo.FindByLabelAndProviderProvider, "a provider")
	assert.Equal(t, authTokenRepo.DeletedServiceAuthTokenFields, expectedToken)
}

func TestDeleteServiceAuthTokenWithN(t *testing.T) {
	authTokenRepo := &testapi.FakeAuthTokenRepo{}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	args := []string{"a label", "a provider"}

	ui := callDeleteServiceAuthToken(t, args, []string{"N"}, reqFactory, authTokenRepo)

	testassert.SliceContains(t, ui.Prompts, testassert.Lines{
		{"Are you sure you want to delete", "a label", "a provider"},
	})
	assert.Equal(t, len(ui.Outputs), 0)
	assert.Equal(t, authTokenRepo.DeletedServiceAuthTokenFields, cf.ServiceAuthTokenFields{})
}

func TestDeleteServiceAuthTokenWithY(t *testing.T) {
	expectedToken := cf.ServiceAuthTokenFields{}
	expectedToken.Label = "a label"
	expectedToken.Provider = "a provider"

	authTokenRepo := &testapi.FakeAuthTokenRepo{
		FindByLabelAndProviderServiceAuthTokenFields: expectedToken,
	}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	args := []string{"a label", "a provider"}

	ui := callDeleteServiceAuthToken(t, args, []string{"Y"}, reqFactory, authTokenRepo)

	testassert.SliceContains(t, ui.Prompts, testassert.Lines{
		{"Are you sure you want to delete", "a label", "a provider"},
	})
	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Deleting"},
		{"OK"},
	})

	assert.Equal(t, authTokenRepo.DeletedServiceAuthTokenFields, expectedToken)
}

func TestDeleteServiceAuthTokenWithForce(t *testing.T) {
	expectedToken := cf.ServiceAuthTokenFields{}
	expectedToken.Label = "a label"
	expectedToken.Provider = "a provider"

	authTokenRepo := &testapi.FakeAuthTokenRepo{
		FindByLabelAndProviderServiceAuthTokenFields: expectedToken,
	}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	args := []string{"-f", "a label", "a provider"}
	ui := callDeleteServiceAuthToken(t, args, []string{"Y"}, reqFactory, authTokenRepo)

	assert.Equal(t, len(ui.Prompts), 0)
	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Deleting"},
		{"OK"},
	})

	assert.Equal(t, authTokenRepo.DeletedServiceAuthTokenFields, expectedToken)
}

func TestDeleteServiceAuthTokenWhenTokenDoesNotExist(t *testing.T) {
	authTokenRepo := &testapi.FakeAuthTokenRepo{
		FindByLabelAndProviderApiResponse: net.NewNotFoundApiResponse("not found"),
	}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	args := []string{"a label", "a provider"}

	ui := callDeleteServiceAuthToken(t, args, []string{"Y"}, reqFactory, authTokenRepo)
	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Deleting service auth token as", "my-user"},
		{"OK"},
		{"does not exist"},
	})
}

func TestDeleteServiceAuthTokenFailsWithError(t *testing.T) {
	authTokenRepo := &testapi.FakeAuthTokenRepo{
		FindByLabelAndProviderApiResponse: net.NewApiResponseWithMessage("OH NOES"),
	}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	args := []string{"a label", "a provider"}

	ui := callDeleteServiceAuthToken(t, args, []string{"Y"}, reqFactory, authTokenRepo)
	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Deleting service auth token as", "my-user"},
		{"FAILED"},
		{"OH NOES"},
	})
}

func callDeleteServiceAuthToken(t *testing.T, args []string, inputs []string, reqFactory *testreq.FakeReqFactory, authTokenRepo *testapi.FakeAuthTokenRepo) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{
		Inputs: inputs,
	}

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)
	org := cf.OrganizationFields{}
	org.Name = "my-org"
	space := cf.SpaceFields{}
	space.Name = "my-space"
	config := &configuration.Configuration{
		SpaceFields:        space,
		OrganizationFields: org,
		AccessToken:        token,
	}

	cmd := NewDeleteServiceAuthToken(ui, config, authTokenRepo)
	ctxt := testcmd.NewContext("delete-service-auth-token", args)

	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
