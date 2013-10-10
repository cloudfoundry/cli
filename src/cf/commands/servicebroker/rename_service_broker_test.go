package servicebroker_test

import (
	"cf"
	. "cf/commands/servicebroker"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestRenameServiceBrokerFailsWithUsage(t *testing.T) {
	reqFactory := &testhelpers.FakeReqFactory{}
	repo := &testhelpers.FakeServiceBrokerRepo{}

	ui := callRenameServiceBroker([]string{}, reqFactory, repo)
	assert.True(t, ui.FailedWithUsage)

	ui = callRenameServiceBroker([]string{"arg1"}, reqFactory, repo)
	assert.True(t, ui.FailedWithUsage)

	ui = callRenameServiceBroker([]string{"arg1", "arg2"}, reqFactory, repo)
	assert.False(t, ui.FailedWithUsage)
}

func TestRenameServiceBrokerRequirements(t *testing.T) {
	reqFactory := &testhelpers.FakeReqFactory{}
	repo := &testhelpers.FakeServiceBrokerRepo{}
	args := []string{"arg1", "arg2"}

	reqFactory.LoginSuccess = false
	callRenameServiceBroker(args, reqFactory, repo)
	assert.False(t, testhelpers.CommandDidPassRequirements)

	reqFactory.LoginSuccess = true
	callRenameServiceBroker(args, reqFactory, repo)
	assert.True(t, testhelpers.CommandDidPassRequirements)
}

func TestRenameServiceBroker(t *testing.T) {
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true}
	repo := &testhelpers.FakeServiceBrokerRepo{
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

func callRenameServiceBroker(args []string, reqFactory *testhelpers.FakeReqFactory, repo *testhelpers.FakeServiceBrokerRepo) (ui *testhelpers.FakeUI) {
	ui = &testhelpers.FakeUI{}

	cmd := NewRenameServiceBroker(ui, repo)
	ctxt := testhelpers.NewContext("rename-service-broker", args)
	testhelpers.RunCommand(cmd, ctxt, reqFactory)

	return
}
