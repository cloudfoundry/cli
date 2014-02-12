package requirements_test

import (
	"cf/models"
	. "cf/requirements"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	testapi "testhelpers/api"
	testassert "testhelpers/assert"
	testterm "testhelpers/terminal"
)

var _ = Describe("Testing with ginkgo", func() {
	It("TestServiceInstanceReqExecute", func() {

		instance := models.ServiceInstance{}
		instance.Name = "my-service"
		instance.Guid = "my-service-guid"
		repo := &testapi.FakeServiceRepo{FindInstanceByNameServiceInstance: instance}
		ui := new(testterm.FakeUI)

		req := NewServiceInstanceRequirement("foo", ui, repo)
		success := req.Execute()

		Expect(success).To(BeTrue())
		Expect(repo.FindInstanceByNameName).To(Equal("foo"))
		Expect(req.GetServiceInstance()).To(Equal(instance))
	})
	It("TestServiceInstanceReqExecuteWhenServiceInstanceNotFound", func() {

		repo := &testapi.FakeServiceRepo{FindInstanceByNameNotFound: true}
		ui := new(testterm.FakeUI)

		testassert.AssertPanic(testterm.FailedWasCalled, func() {
			NewServiceInstanceRequirement("foo", ui, repo).Execute()
		})
	})
})
