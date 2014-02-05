package requirements_test

import (
	"cf"
	. "cf/requirements"
	. "github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	mr "github.com/tjarratt/mr_t"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testterm "testhelpers/terminal"
)

func init() {
	Describe("Testing with ginkgo", func() {
		It("TestServiceInstanceReqExecute", func() {

			instance := cf.ServiceInstance{}
			instance.Name = "my-service"
			instance.Guid = "my-service-guid"
			repo := &testapi.FakeServiceRepo{FindInstanceByNameServiceInstance: instance}
			ui := new(testterm.FakeUI)

			req := NewServiceInstanceRequirement("foo", ui, repo)
			success := req.Execute()

			assert.True(mr.T(), success)
			assert.Equal(mr.T(), repo.FindInstanceByNameName, "foo")
			assert.Equal(mr.T(), req.GetServiceInstance(), instance)
		})
		It("TestServiceInstanceReqExecuteWhenServiceInstanceNotFound", func() {

			repo := &testapi.FakeServiceRepo{FindInstanceByNameNotFound: true}
			ui := new(testterm.FakeUI)

			testassert.AssertPanic(mr.T(), testterm.FailedWasCalled, func() {
				NewServiceInstanceRequirement("foo", ui, repo).Execute()
			})
		})
	})
}
