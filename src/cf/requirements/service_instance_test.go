package requirements_test

import (
	"cf"
	"cf/configuration"
	. "cf/requirements"
	"github.com/stretchr/testify/assert"
	"testhelpers"
	"testing"
)

func TestServiceInstanceReqExecute(t *testing.T) {
	instance := cf.ServiceInstance{Name: "my-service", Guid: "my-service-guid"}
	repo := &testhelpers.FakeServiceRepo{FindInstanceByNameServiceInstance: instance}
	config := configuration.Configuration{}
	ui := new(testhelpers.FakeUI)

	req := NewServiceInstanceRequirement("foo", ui, config, repo)
	success := req.Execute()

	assert.True(t, success)
	assert.Equal(t, repo.FindInstanceByNameName, "foo")
	assert.Equal(t, req.GetServiceInstance(), instance)
}
