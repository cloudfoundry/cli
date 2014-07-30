package actors_test

import (
	"github.com/cloudfoundry/cli/cf/actors"
	"github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Service Plans", func() {
	var (
		actor                     actors.ServicePlanActor
		serviceRepo               *fakes.FakeServiceRepo
		servicePlanRepo           *fakes.FakeServicePlanRepo
		servicePlanVisibilityRepo *fakes.FakeServicePlanVisibilityRepository
		orgRepo                   *fakes.FakeOrgRepository
	)

	BeforeEach(func() {
		serviceRepo = &fakes.FakeServiceRepo{}
		servicePlanRepo = &fakes.FakeServicePlanRepo{}
		servicePlanVisibilityRepo = &fakes.FakeServicePlanVisibilityRepository{}
		orgRepo = &fakes.FakeOrgRepository{}

		actor = actors.NewServicePlanHandler(serviceRepo, servicePlanRepo, servicePlanVisibilityRepo, orgRepo)

		service := models.ServiceOffering{ServiceOfferingFields: models.ServiceOfferingFields{Label: "my-service", Guid: "my-service-guid"}}

		serviceRepo.FindServiceOfferingByLabelServiceOffering = service

		servicePlanRepo.SearchReturns = map[string][]models.ServicePlanFields{
			"my-service-guid": {
				{Name: "small-service-plan", Guid: "small-service-plan-guid"},
				{Name: "large-service-plan", Guid: "large-service-plan-guid"},
			},
		}
	})

	Describe(".GetServicePlanForService", func() {
		It("Returns a single service plan", func() {
			servicePlan, err := actor.GetSingleServicePlanForService("my-service", "small-service-plan")
			Expect(err).NotTo(HaveOccurred())

			Expect(servicePlan.Name).To(Equal("small-service-plan"))
			Expect(servicePlan.Guid).To(Equal("small-service-plan-guid"))
		})
	})
})
