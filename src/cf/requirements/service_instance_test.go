package requirements

import (
	"cf"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testterm "testhelpers/terminal"
	"testing"
)

func TestServiceInstanceReqExecute(t *testing.T) {
	instance := cf.ServiceInstance{}
	instance.Name = "my-service"
	instance.Guid = "my-service-guid"
	repo := &testapi.FakeServiceRepo{FindInstanceByNameServiceInstance: instance}
	ui := new(testterm.FakeUI)

	req := newServiceInstanceRequirement("foo", ui, repo)
	success := req.Execute()

	assert.True(t, success)
	assert.Equal(t, repo.FindInstanceByNameName, "foo")
	assert.Equal(t, req.GetServiceInstance(), instance)
}

func TestServiceInstanceReqExecuteWhenServiceInstanceNotFound(t *testing.T) {
	repo := &testapi.FakeServiceRepo{FindInstanceByNameNotFound: true}
	ui := new(testterm.FakeUI)

	testassert.AssertPanic(t, testterm.FailedWasCalled, func() {
		newServiceInstanceRequirement("foo", ui, repo).Execute()
	})
}
