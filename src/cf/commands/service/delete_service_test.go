package service_test

import (
	"cf"
	"cf/api"
	. "cf/commands/service"
	"cf/configuration"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testconfig "testhelpers/configuration"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestDeleteServiceCommand(t *testing.T) {
	serviceInstance := cf.ServiceInstance{Name: "my-service", Guid: "my-service-guid"}
	reqFactory := &testreq.FakeReqFactory{}
	serviceRepo := &testapi.FakeServiceRepo{FindInstanceByNameServiceInstance: serviceInstance}
	fakeUI := callDeleteService(t, []string{"my-service"}, reqFactory, serviceRepo)

	assert.Contains(t, fakeUI.Outputs[0], "Deleting service")
	assert.Contains(t, fakeUI.Outputs[0], "my-service")
	assert.Contains(t, fakeUI.Outputs[0], "my-org")
	assert.Contains(t, fakeUI.Outputs[0], "my-space")
	assert.Contains(t, fakeUI.Outputs[0], "my-user")

	assert.Equal(t, serviceRepo.DeleteServiceServiceInstance, serviceInstance)
	assert.Contains(t, fakeUI.Outputs[1], "OK")
}

func TestDeleteServiceCommandOnNonExistentService(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	serviceRepo := &testapi.FakeServiceRepo{FindInstanceByNameNotFound: true}
	fakeUI := callDeleteService(t, []string{"my-service"}, reqFactory, serviceRepo)

	assert.Contains(t, fakeUI.Outputs[0], "Deleting service")
	assert.Contains(t, fakeUI.Outputs[0], "my-service")

	assert.Contains(t, fakeUI.Outputs[1], "OK")
	assert.Contains(t, fakeUI.Outputs[2], "my-service")
	assert.Contains(t, fakeUI.Outputs[2], "not exist")
}

func TestDeleteServiceCommandFailsWithUsage(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}
	serviceRepo := &testapi.FakeServiceRepo{}

	fakeUI := callDeleteService(t, []string{}, reqFactory, serviceRepo)
	assert.True(t, fakeUI.FailedWithUsage)

	fakeUI = callDeleteService(t, []string{"my-service"}, reqFactory, serviceRepo)
	assert.False(t, fakeUI.FailedWithUsage)
}

func callDeleteService(t *testing.T, args []string, reqFactory *testreq.FakeReqFactory, serviceRepo api.ServiceRepository) (fakeUI *testterm.FakeUI) {
	fakeUI = new(testterm.FakeUI)
	ctxt := testcmd.NewContext("delete-service", args)

	token, err := testconfig.CreateAccessTokenWithTokenInfo(configuration.TokenInfo{
		Username: "my-user",
	})
	assert.NoError(t, err)

	config := &configuration.Configuration{
		Space:        cf.Space{Name: "my-space"},
		Organization: cf.Organization{Name: "my-org"},
		AccessToken:  token,
	}

	cmd := NewDeleteService(fakeUI, config, serviceRepo)
	testcmd.RunCommand(cmd, ctxt, reqFactory)
	return
}
