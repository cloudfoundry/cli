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

func TestRenameServiceBrokerFailsWithUsage(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	repo := &testapi.FakeServiceBrokerRepo{}

	ui := callRenameServiceBroker([]string{}, reqFactory, repo)
	assert.True(t, ui.FailedWithUsage)

	ui = callRenameServiceBroker([]string{"arg1"}, reqFactory, repo)
	assert.True(t, ui.FailedWithUsage)

	ui = callRenameServiceBroker([]string{"arg1", "arg2"}, reqFactory, repo)
	assert.False(t, ui.FailedWithUsage)
}

func TestRenameServiceBrokerRequirements(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	repo := &testapi.FakeServiceBrokerRepo{}
	args := []string{"arg1", "arg2"}

	reqFactory.LoginSuccess = false
	callRenameServiceBroker(args, reqFactory, repo)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory.LoginSuccess = true
	callRenameServiceBroker(args, reqFactory, repo)
	assert.True(t, testcmd.CommandDidPassRequirements)
}

func TestRenameServiceBroker(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	repo := &testapi.FakeServiceBrokerRepo{
		FindByNameServiceBroker: cf.ServiceBroker{Name: "my-found-broker", Guid: "my-found-broker-guid"},
	}
	args := []string{"my-broker", "my-new-broker"}

	ui := callRenameServiceBroker(args, reqFactory, repo)

	assert.Equal(t, repo.FindByNameName, "my-broker")

	assert.Contains(t, ui.Outputs[0], "Renaming service broker")
	assert.Contains(t, ui.Outputs[0], "my-found-broker")

	expectedServiceBroker := cf.ServiceBroker{
		Name: "my-new-broker",
		Guid: "my-found-broker-guid",
	}

	assert.Equal(t, repo.RenamedServiceBroker, expectedServiceBroker)

	assert.Contains(t, ui.Outputs[1], "OK")
}

func callRenameServiceBroker(args []string, reqFactory *testreq.FakeReqFactory, repo *testapi.FakeServiceBrokerRepo) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}

	cmd := NewRenameServiceBroker(ui, repo)
	ctxt := testcmd.NewContext("rename-service-broker", args)
	testcmd.RunCommand(cmd, ctxt, reqFactory)

	return
}
