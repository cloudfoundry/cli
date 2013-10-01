package requirements_test

import (
	"cf"
	. "cf/requirements"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestServiceInstanceReqExecute(t *testing.T) {
	instance := cf.ServiceInstance{Name: "my-service", Guid: "my-service-guid"}
	repo := &testhelpers.FakeServiceRepo{FindInstanceByNameServiceInstance: instance}
	ui := new(testhelpers.FakeUI)

	req := NewServiceInstanceRequirement("foo", ui, repo)
	success := req.Execute()

	assert.True(t, success)
	assert.Equal(t, repo.FindInstanceByNameName, "foo")
	assert.Equal(t, req.GetServiceInstance(), instance)
}

func TestServiceInstanceReqExecuteWhenServiceInstanceNotFound(t *testing.T) {
	repo := &testhelpers.FakeServiceRepo{FindInstanceByNameNotFound: true}
	ui := new(testhelpers.FakeUI)

	req := NewServiceInstanceRequirement("foo", ui, repo)
	success := req.Execute()

	assert.False(t, success)
}
