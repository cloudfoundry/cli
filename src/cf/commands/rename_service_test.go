package commands_test

import (
	"testing"
	. "cf/commands"
	"testhelpers"
	"github.com/stretchr/testify/assert"
	"cf"
)

func TestRenameServiceFailsWithUsage(t *testing.T) {
	reqFactory := &testhelpers.FakeReqFactory{}

	fakeUI, _ := callRenameService([]string{}, reqFactory)
	assert.True(t, fakeUI.FailedWithUsage)

	fakeUI, _ = callRenameService([]string{"my-service"}, reqFactory)
	assert.True(t, fakeUI.FailedWithUsage)

	fakeUI, _ = callRenameService([]string{"my-service", "new-name", "extra"}, reqFactory)
	assert.True(t, fakeUI.FailedWithUsage)
}

func TestRenameServiceRequirements(t *testing.T) {
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: false, TargetedSpaceSuccess : true}
	callRenameService([]string{"my-service", "new-name"}, reqFactory)
	assert.False(t, testhelpers.CommandDidPassRequirements)

	reqFactory = &testhelpers.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess : false}
	callRenameService([]string{"my-service", "new-name"}, reqFactory)
	assert.False(t, testhelpers.CommandDidPassRequirements)

	assert.Equal(t, reqFactory.ServiceInstanceName, "my-service")
}

func TestRenameService(t *testing.T) {
	serviceInstance := cf.ServiceInstance{Name: "different-name", Guid:"different-name-guid"}
	reqFactory := &testhelpers.FakeReqFactory{LoginSuccess: true, TargetedSpaceSuccess : true, ServiceInstance: serviceInstance}
	fakeUI, fakeServiceRepo := callRenameService([]string{"my-service", "new-name"}, reqFactory)

	assert.Equal(t, fakeUI.Outputs[0], "Renaming service different-name...")
	assert.Equal(t, fakeUI.Outputs[1], "OK")

	assert.Equal(t, fakeServiceRepo.RenameServiceServiceInstance, serviceInstance)
	assert.Equal(t, fakeServiceRepo.RenameServiceNewName, "new-name")
}

func callRenameService(args []string, reqFactory *testhelpers.FakeReqFactory) (ui *testhelpers.FakeUI, serviceRepo *testhelpers.FakeServiceRepo) {
	ui = &testhelpers.FakeUI{}
	serviceRepo = &testhelpers.FakeServiceRepo{}
	cmd := NewRenameService(ui, serviceRepo)
	ctxt := testhelpers.NewContext("rename-service", args)

	testhelpers.RunCommand(cmd, ctxt, reqFactory)

	return
}
