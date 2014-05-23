package requirements_test

import (
	"github.com/cloudfoundry/cli/cf/models"
	. "github.com/cloudfoundry/cli/cf/requirements"
	testapi "github.com/cloudfoundry/cli/testhelpers/api"
	testassert "github.com/cloudfoundry/cli/testhelpers/assert"
	testterm "github.com/cloudfoundry/cli/testhelpers/terminal"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ServiceInstanceRequirement", func() {
	var (
		ui *testterm.FakeUI
	)

	BeforeEach(func() {
		ui = new(testterm.FakeUI)
	})

	Context("when a service instance with the given name can be found", func() {
		It("succeeds", func() {
			instance := models.ServiceInstance{}
			instance.Name = "my-service"
			instance.Guid = "my-service-guid"
			repo := &testapi.FakeServiceRepo{FindInstanceByNameServiceInstance: instance}

			req := NewServiceInstanceRequirement("my-service", ui, repo)

			Expect(req.Execute()).To(BeTrue())
			Expect(repo.FindInstanceByNameName).To(Equal("my-service"))
			Expect(req.GetServiceInstance()).To(Equal(instance))
		})
	})

	Context("when a service instance with the given name can't be found", func() {
		It("fails", func() {
			repo := &testapi.FakeServiceRepo{FindInstanceByNameNotFound: true}
			testassert.AssertPanic(testterm.FailedWasCalled, func() {
				NewServiceInstanceRequirement("foo", ui, repo).Execute()
			})
		})
	})
})
