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

func TestCreateServiceBrokerFailsWithUsage(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	serviceBrokerRepo := &testapi.FakeServiceBrokerRepo{}

	ui := callCreateServiceBroker([]string{}, reqFactory, serviceBrokerRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callCreateServiceBroker([]string{"1arg"}, reqFactory, serviceBrokerRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callCreateServiceBroker([]string{"1arg", "2arg"}, reqFactory, serviceBrokerRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callCreateServiceBroker([]string{"1arg", "2arg", "3arg"}, reqFactory, serviceBrokerRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callCreateServiceBroker([]string{"1arg", "2arg", "3arg", "4arg"}, reqFactory, serviceBrokerRepo)
	assert.False(t, ui.FailedWithUsage)

}
func TestCreateServiceBrokerRequirements(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	serviceBrokerRepo := &testapi.FakeServiceBrokerRepo{}
	args := []string{"1arg", "2arg", "3arg", "4arg"}

	reqFactory.LoginSuccess = false
	callCreateServiceBroker(args, reqFactory, serviceBrokerRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory.LoginSuccess = true
	callCreateServiceBroker(args, reqFactory, serviceBrokerRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)
}

func TestCreateServiceBroker(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	serviceBrokerRepo := &testapi.FakeServiceBrokerRepo{}
	args := []string{"my-broker", "my username", "my password", "http://example.com"}
	ui := callCreateServiceBroker(args, reqFactory, serviceBrokerRepo)

	assert.Contains(t, ui.Outputs[0], "Creating service broker")
	assert.Contains(t, ui.Outputs[0], "my-broker")

	expectedServiceBroker := cf.ServiceBroker{
		Name:     "my-broker",
		Username: "my username",
		Password: "my password",
		Url:      "http://example.com",
	}
	assert.Equal(t, serviceBrokerRepo.CreatedServiceBroker, expectedServiceBroker)

	assert.Contains(t, ui.Outputs[1], "OK")
}

func callCreateServiceBroker(args []string, reqFactory *testreq.FakeReqFactory, serviceBrokerRepo *testapi.FakeServiceBrokerRepo) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}
	ctxt := testcmd.NewContext("create-service-broker", args)

	cmd := NewCreateServiceBroker(ui, serviceBrokerRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
