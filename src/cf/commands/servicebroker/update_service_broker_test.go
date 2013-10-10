package servicebroker_test

import (
	"cf"
	. "cf/commands/servicebroker"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestUpdateServiceBrokerFailsWithUsage(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	repo := &testapi.FakeServiceBrokerRepo{}

	ui := callUpdateServiceBroker([]string{}, reqFactory, repo)
	assert.True(t, ui.FailedWithUsage)

	ui = callUpdateServiceBroker([]string{"arg1"}, reqFactory, repo)
	assert.True(t, ui.FailedWithUsage)

	ui = callUpdateServiceBroker([]string{"arg1", "arg2"}, reqFactory, repo)
	assert.True(t, ui.FailedWithUsage)

	ui = callUpdateServiceBroker([]string{"arg1", "arg2", "arg3"}, reqFactory, repo)
	assert.True(t, ui.FailedWithUsage)

	ui = callUpdateServiceBroker([]string{"arg1", "arg2", "arg3", "arg4"}, reqFactory, repo)
	assert.False(t, ui.FailedWithUsage)
}

func TestUpdateServiceBrokerRequirements(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	repo := &testapi.FakeServiceBrokerRepo{}
	args := []string{"arg1", "arg2", "arg3", "arg4"}

	reqFactory.LoginSuccess = false
	callUpdateServiceBroker(args, reqFactory, repo)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory.LoginSuccess = true
	callUpdateServiceBroker(args, reqFactory, repo)
	assert.True(t, testcmd.CommandDidPassRequirements)
}

func TestUpdateServiceBroker(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	repo := &testapi.FakeServiceBrokerRepo{
		FindByNameServiceBroker: cf.ServiceBroker{Name: "my-found-broker", Guid: "my-found-broker-guid"},
	}
	args := []string{"my-broker", "new-username", "new-password", "new-url"}

	ui := callUpdateServiceBroker(args, reqFactory, repo)

	assert.Equal(t, repo.FindByNameName, "my-broker")

	assert.Contains(t, ui.Outputs[0], "Updating service broker")
	assert.Contains(t, ui.Outputs[0], "my-found-broker")

	expectedServiceBroker := cf.ServiceBroker{
		Name:     "my-found-broker",
		Username: "new-username",
		Password: "new-password",
		Url:      "new-url",
		Guid:     "my-found-broker-guid",
	}

	assert.Equal(t, repo.UpdatedServiceBroker, expectedServiceBroker)

	assert.Contains(t, ui.Outputs[1], "OK")
}

func callUpdateServiceBroker(args []string, reqFactory *testreq.FakeReqFactory, repo *testapi.FakeServiceBrokerRepo) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}

	cmd := NewUpdateServiceBroker(ui, repo)
	ctxt := testcmd.NewContext("update-service-broker", args)
	testcmd.RunCommand(cmd, ctxt, reqFactory)

	return
}
