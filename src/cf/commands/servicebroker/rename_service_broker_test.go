package servicebroker_test

import (
	"cf"
	. "cf/commands/servicebroker"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestRenameServiceBrokerFailsWithUsage(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	repo := &testapi.FakeServiceBrokerRepo{}

	ui := callRenameServiceBroker(t, []string{}, reqFactory, repo)
	assert.True(t, ui.FailedWithUsage)

	ui = callRenameServiceBroker(t, []string{"arg1"}, reqFactory, repo)
	assert.True(t, ui.FailedWithUsage)

	ui = callRenameServiceBroker(t, []string{"arg1", "arg2"}, reqFactory, repo)
	assert.False(t, ui.FailedWithUsage)
}

func TestRenameServiceBrokerRequirements(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	repo := &testapi.FakeServiceBrokerRepo{}
	args := []string{"arg1", "arg2"}

	reqFactory.LoginSuccess = false
	callRenameServiceBroker(t, args, reqFactory, repo)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory.LoginSuccess = true
	callRenameServiceBroker(t, args, reqFactory, repo)
	assert.True(t, testcmd.CommandDidPassRequirements)
}

func TestRenameServiceBroker(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	broker := cf.ServiceBroker{}
	broker.Name = "my-found-broker"
	broker.Guid = "my-found-broker-guid"
	repo := &testapi.FakeServiceBrokerRepo{
		FindByNameServiceBroker: broker,
	}
	args := []string{"my-broker", "my-new-broker"}

	ui := callRenameServiceBroker(t, args, reqFactory, repo)

	assert.Equal(t, repo.FindByNameName, "my-broker")

	assert.Contains(t, ui.Outputs[0], "Renaming service broker")
	assert.Contains(t, ui.Outputs[0], "my-found-broker")
	assert.Contains(t, ui.Outputs[0], "my-new-broker")
	assert.Contains(t, ui.Outputs[0], "my-user")

	assert.Equal(t, repo.RenamedServiceBrokerGuid, "my-found-broker-guid")
	assert.Equal(t, repo.RenamedServiceBrokerName, "my-new-broker")

	assert.Contains(t, ui.Outputs[1], "OK")
}

func callRenameServiceBroker(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory, repo *testapi.FakeServiceBrokerRepo) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}

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

	cmd := NewRenameServiceBroker(ui, config, repo)
	ctxt := testcmd.NewContext("rename-service-broker", args)
	testcmd.RunCommand(cmd, ctxt, reqFactory)

	return
}
