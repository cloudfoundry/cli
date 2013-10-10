package service_test

import (
	"cf"
	. "cf/commands/service"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testcmd "testhelpers/commands"
	testreq "testhelpers/requirements"
	testterm "testhelpers/terminal"
	"testing"
)

func TestRenameServiceFailsWithUsage(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{}

	fakeUI, _ := callRenameService([]string{}, reqFactory)
	assert.True(t, fakeUI.FailedWithUsage)

	fakeUI, _ = callRenameService([]string{"my-service"}, reqFactory)
	assert.True(t, fakeUI.FailedWithUsage)

	fakeUI, _ = callRenameService([]string{"my-service", "new-name", "extra"}, reqFactory)
	assert.True(t, fakeUI.FailedWithUsage)
}

func TestRenameServiceRequirements(t *testing.T) {
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: false, TargetedSpaceSuccess: true}
	callRenameService([]string{"my-service", "new-name"}, reqFactory)
	assert.False(t, testcmd.CommandDidPassRequirements)

	reqFactory = &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: false}
	callRenameService([]string{"my-service", "new-name"}, reqFactory)
	assert.False(t, testcmd.CommandDidPassRequirements)

	assert.Equal(t, reqFactory.ServiceInstanceName, "my-service")
}

func TestRenameService(t *testing.T) {
	serviceInstance := cf.ServiceInstance{Name: "different-name", Guid: "different-name-guid"}
	reqFactory := &testreq.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess: true, ServiceInstance: serviceInstance}
	fakeUI, fakeServiceRepo := callRenameService([]string{"my-service", "new-name"}, reqFactory)

	assert.Equal(t, fakeUI.Outputs[0], "Renaming service different-name...")
	assert.Equal(t, fakeUI.Outputs[1], "OK")

	assert.Equal(t, fakeServiceRepo.RenameServiceServiceInstance, serviceInstance)
	assert.Equal(t, fakeServiceRepo.RenameServiceNewName, "new-name")
}

func callRenameService(args []string, reqFactory *testreq.FakeReqFactory) (ui *testterm.FakeUI, serviceRepo *testapi.FakeServiceRepo) {
	ui = &testterm.FakeUI{}
	serviceRepo = &testapi.FakeServiceRepo{}
	cmd := NewRenameService(ui, serviceRepo)
	ctxt := testcmd.NewContext("rename-service", args)

	testcmd.RunCommand(cmd, ctxt, reqFactory)

	return
}
