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

func TestDeleteServiceBrokerFailsWithUsage(t *testing.T) {
	ui, _, _ := deleteServiceBroker("y", []string{})
	assert.True(t, ui.FailedWithUsage)

	ui, _, _ = deleteServiceBroker("y", []string{"my-broker"})
	assert.False(t, ui.FailedWithUsage)
}

func TestDeleteServiceBrokerRequirements(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	repo := &testapi.FakeServiceBrokerRepo{}

	reqFactory.LoginSuccess = false
	callDeleteServiceBroker([]string{"-f", "my-broker"}, reqFactory, repo)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory.LoginSuccess = true
	callDeleteServiceBroker([]string{"-f", "my-broker"}, reqFactory, repo)
	assert.True(t, testcmd.CommandDidPassRequirements)
}

func TestDeleteConfirmingWithY(t *testing.T) {
	ui, _, repo := deleteServiceBroker("y", []string{"service-broker-to-delete"})

	assert.Equal(t, repo.FindByNameName, "service-broker-to-delete")
	assert.Equal(t, repo.DeletedServiceBroker.Name, "service-broker-to-delete")
	assert.Equal(t, len(ui.Outputs), 2)
	assert.Contains(t, ui.Prompts[0], "Really delete")
	assert.Contains(t, ui.Outputs[0], "service-broker-to-delete")
	assert.Contains(t, ui.Outputs[0], "Deleting")
	assert.Contains(t, ui.Outputs[1], "OK")
}

func TestDeleteConfirmingWithYes(t *testing.T) {
	ui, _, repo := deleteServiceBroker("Yes", []string{"service-broker-to-delete"})

	assert.Equal(t, repo.FindByNameName, "service-broker-to-delete")
	assert.Equal(t, repo.DeletedServiceBroker.Name, "service-broker-to-delete")
	assert.Equal(t, len(ui.Outputs), 2)
	assert.Contains(t, ui.Prompts[0], "Really delete")
	assert.Contains(t, ui.Outputs[0], "service-broker-to-delete")
	assert.Contains(t, ui.Outputs[0], "Deleting")
	assert.Contains(t, ui.Outputs[1], "OK")
}

func TestDeleteWithForceOption(t *testing.T) {
	serviceBroker := cf.ServiceBroker{
		Name: "service-broker-to-delete",
		Guid: "service-broker-to-delete-guid",
	}

	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	repo := &testapi.FakeServiceBrokerRepo{FindByNameServiceBroker: serviceBroker}
	ui := callDeleteServiceBroker([]string{"-f", "service-broker-to-delete"}, reqFactory, repo)

	assert.Equal(t, repo.FindByNameName, "service-broker-to-delete")
	assert.Equal(t, repo.DeletedServiceBroker, serviceBroker)
	assert.Equal(t, len(ui.Prompts), 0)
	assert.Equal(t, len(ui.Outputs), 2)
	assert.Contains(t, ui.Outputs[0], "Deleting")
	assert.Contains(t, ui.Outputs[0], "service-broker-to-delete")
	assert.Contains(t, ui.Outputs[1], "OK")
}

func TestDeleteAppThatDoesNotExist(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	repo := &testapi.FakeServiceBrokerRepo{FindByNameNotFound: true}
	ui := callDeleteServiceBroker([]string{"-f", "service-broker-to-delete"}, reqFactory, repo)

	assert.Equal(t, repo.FindByNameName, "service-broker-to-delete")
	assert.Equal(t, repo.DeletedServiceBroker.Name, "")
	assert.Contains(t, ui.Outputs[0], "Deleting")
	assert.Contains(t, ui.Outputs[0], "service-broker-to-delete")
	assert.Contains(t, ui.Outputs[1], "OK")
	assert.Contains(t, ui.Outputs[2], "service-broker-to-delete")
	assert.Contains(t, ui.Outputs[2], "does not exist")
}

func callDeleteServiceBroker(args []string, reqFactory *testreq.FakeReqFactory, repo *testapi.FakeServiceBrokerRepo) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}
	ctxt := testcmd.NewContext("delete-service-broker", args)

	cmd := NewDeleteServiceBroker(ui, repo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}

func deleteServiceBroker(confirmation string, args []string) (ui *testterm.FakeUI, reqFactory *testreq.FakeReqFactory, repo *testapi.FakeServiceBrokerRepo) {
	serviceBroker := cf.ServiceBroker{
		Name: "service-broker-to-delete",
		Guid: "service-broker-to-delete-guid",
	}

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: true}
	repo = &testapi.FakeServiceBrokerRepo{FindByNameServiceBroker: serviceBroker}
	ui = &testterm.FakeUI{
		Inputs: []string{confirmation},
	}
	ctxt := testcmd.NewContext("delete-service-broker", args)
	cmd := NewDeleteServiceBroker(ui, repo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
