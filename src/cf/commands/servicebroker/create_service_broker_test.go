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

func TestCreateServiceBrokerFailsWithUsage(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	serviceBrokerRepo := &testapi.FakeServiceBrokerRepo{}

	ui := callCreateServiceBroker(t, []string{}, reqFactory, serviceBrokerRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callCreateServiceBroker(t, []string{"1arg"}, reqFactory, serviceBrokerRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callCreateServiceBroker(t, []string{"1arg", "2arg"}, reqFactory, serviceBrokerRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callCreateServiceBroker(t, []string{"1arg", "2arg", "3arg"}, reqFactory, serviceBrokerRepo)
	assert.True(t, ui.FailedWithUsage)

	ui = callCreateServiceBroker(t, []string{"1arg", "2arg", "3arg", "4arg"}, reqFactory, serviceBrokerRepo)
	assert.False(t, ui.FailedWithUsage)

}
func TestCreateServiceBrokerRequirements(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	serviceBrokerRepo := &testapi.FakeServiceBrokerRepo{}
	args := []string{"1arg", "2arg", "3arg", "4arg"}

	reqFactory.LoginSuccess = false
	callCreateServiceBroker(t, args, reqFactory, serviceBrokerRepo)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory.LoginSuccess = true
	callCreateServiceBroker(t, args, reqFactory, serviceBrokerRepo)
	assert.True(t, testcmd.CommandDidPassRequirements)
}

func TestCreateServiceBroker(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true}
	serviceBrokerRepo := &testapi.FakeServiceBrokerRepo{}
	args := []string{"my-broker", "my username", "my password", "http://example.com"}
	ui := callCreateServiceBroker(t, args, reqFactory, serviceBrokerRepo)

	testassert.SliceContains(t, ui.Outputs, testassert.Lines{
		{"Creating service broker", "my-broker", "my-user"},
		{"OK"},
	})

	assert.Equal(t, serviceBrokerRepo.CreateName, "my-broker")
	assert.Equal(t, serviceBrokerRepo.CreateUrl, "http://example.com")
	assert.Equal(t, serviceBrokerRepo.CreateUsername, "my username")
	assert.Equal(t, serviceBrokerRepo.CreatePassword, "my password")
}

func callCreateServiceBroker(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory, serviceBrokerRepo *testapi.FakeServiceBrokerRepo) (ui *testterm.FakeUI) {
	ui = &testterm.FakeUI{}
	ctxt := testcmd.NewContext("create-service-broker", args)

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

	cmd := NewCreateServiceBroker(ui, config, serviceBrokerRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
