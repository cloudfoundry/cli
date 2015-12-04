package requirements_test

import (
	testapi "github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/errors"
	"github.com/cloudfoundry/cli/cf/models"
	. "github.com/cloudfoundry/cli/cf/requirements"
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
			repo := &testapi.FakeServiceRepository{}
			repo.FindInstanceByNameReturns(instance, nil)

			req := NewServiceInstanceRequirement("my-service", ui, repo)

			Expect(req.Execute()).To(BeTrue())
			Expect(repo.FindInstanceByNameArgsForCall(0)).To(Equal("my-service"))
			Expect(req.GetServiceInstance()).To(Equal(instance))
		})
	})

	Context("when a service instance with the given name can't be found", func() {
		It("fails", func() {
			repo := &testapi.FakeServiceRepository{}
			repo.FindInstanceByNameReturns(models.ServiceInstance{}, errors.NewModelNotFoundError("Service instance", "my-service"))
			testassert.AssertPanic(testterm.QuietPanic, func() {
				NewServiceInstanceRequirement("foo", ui, repo).Execute()
			})
		})
	})
})
