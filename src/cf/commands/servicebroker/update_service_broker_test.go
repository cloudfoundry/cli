package servicebroker_test

import (
	"cf"
	. "cf/commands/servicebroker"
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

func TestUpdateServiceBrokerFailsWithUsage(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	repo := &testapi.FakeServiceBrokerRepo{}

	ui := callUpdateServiceBroker(t, []string{}, reqFactory, repo)
	assert.True(t, ui.FailedWithUsage)

	ui = callUpdateServiceBroker(t, []string{"arg1"}, reqFactory, repo)
	assert.True(t, ui.FailedWithUsage)

	ui = callUpdateServiceBroker(t, []string{"arg1", "arg2"}, reqFactory, repo)
	assert.True(t, ui.FailedWithUsage)

	ui = callUpdateServiceBroker(t, []string{"arg1", "arg2", "arg3"}, reqFactory, repo)
	assert.True(t, ui.FailedWithUsage)

	ui = callUpdateServiceBroker(t, []string{"arg1", "arg2", "arg3", "arg4"}, reqFactory, repo)
	assert.False(t, ui.FailedWithUsage)
}

func TestUpdateServiceBrokerRequirements(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	repo := &testapi.FakeServiceBrokerRepo{}
	args := []string{"arg1", "arg2", "arg3", "arg4"}

	reqFactory.LoginSuccess = false
	callUpdateServiceBroker(t, args, reqFactory, repo)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory.LoginSuccess = true
	callUpdateServiceBroker(t, args, reqFactory, repo)
	assert.True(t, testcmd.CommandDidPassRequirements)
}

func TestUpdateServiceBroker(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	broker := cf.ServiceBroker{}
	broker.Name = "my-found-broker"
	broker.Guid = "my-found-broker-guid"
	repo := &testapi.FakeServiceBrokerRepo{
		FindByNameServiceBroker: broker,
	}
	args := []string{"my-broker", "new-username", "new-password", "new-url"}

	ui := callUpdateServiceBroker(t, args, reqFactory, repo)

	assert.Equal(t, repo.FindByNameName, "my-broker")

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Updating service broker", "my-found-broker", "my-user"},
		{"OK"},
	})

	expectedServiceBroker := cf.ServiceBroker{}
	expectedServiceBroker.Name = "my-found-broker"
	expectedServiceBroker.Username = "new-username"
	expectedServiceBroker.Password = "new-password"
	expectedServiceBroker.Url = "new-url"
	expectedServiceBroker.Guid = "my-found-broker-guid"

	assert.Equal(t, repo.UpdatedServiceBroker, expectedServiceBroker)
}

func callUpdateServiceBroker(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory, repo *testapi.FakeServiceBrokerRepo) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}

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

	cmd := NewUpdateServiceBroker(ui, config, repo)
	ctxt := testcmd.NewContext("update-service-broker", args)
	testcmd.RunCommand(cmd, ctxt, reqFactory)

	return
}
