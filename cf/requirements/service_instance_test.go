package requirements_test

import (
	"code.cloudfoundry.org/cli/cf/api/apifakes"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/models"
	. "code.cloudfoundry.org/cli/cf/requirements"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ServiceInstanceRequirement", func() {
	var repo *apifakes.FakeServiceRepository

	BeforeEach(func() {
		repo = new(apifakes.FakeServiceRepository)
	})

	Context("when a service instance with the given name can be found", func() {
		It("succeeds", func() {
			instance := models.ServiceInstance{}
			instance.Name = "my-service"
			instance.GUID = "my-service-guid"
			repo.FindInstanceByNameReturns(instance, nil)

			req := NewServiceInstanceRequirement("my-service", repo)

			err := req.Execute()
			Expect(err).NotTo(HaveOccurred())
			Expect(repo.FindInstanceByNameArgsForCall(0)).To(Equal("my-service"))
			Expect(req.GetServiceInstance()).To(Equal(instance))
		})
	})

	Context("when a service instance with the given name can't be found", func() {
		It("errors", func() {
			repo.FindInstanceByNameReturns(models.ServiceInstance{}, errors.NewModelNotFoundError("Service instance", "my-service"))
			err := NewServiceInstanceRequirement("foo", repo).Execute()
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("Service instance my-service not found"))
		})
	})
})
