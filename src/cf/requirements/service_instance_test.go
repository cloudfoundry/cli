package requirements_test

import (
	"cf"
	. "cf/requirements"
	"github.com/stretchr/testify/assert"
	testapi "testhelpers/api"
	testterm "testhelpers/terminal"
	"testing"
)

func TestServiceInstanceReqExecute(t *testing.T) {
	instance := cf.ServiceInstance{Name: "my-service", Guid: "my-service-guid"}
	repo := &testapi.FakeServiceRepo{FindInstanceByNameServiceInstance: instance}
	ui := new(testterm.FakeUI)

	req := NewServiceInstanceRequirement("foo", ui, repo)
	success := req.Execute()

	assert.True(t, success)
	assert.Equal(t, repo.FindInstanceByNameName, "foo")
	assert.Equal(t, req.GetServiceInstance(), instance)
}

func TestServiceInstanceReqExecuteWhenServiceInstanceNotFound(t *testing.T) {
	repo := &testapi.FakeServiceRepo{FindInstanceByNameNotFound: true}
	ui := new(testterm.FakeUI)

	req := NewServiceInstanceRequirement("foo", ui, repo)
	success := req.Execute()

	assert.False(t, success)
}
