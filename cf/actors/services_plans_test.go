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

		publicServicePlan  models.ServicePlanFields
		privateServicePlan models.ServicePlanFields
	)

	BeforeEach(func() {
		serviceRepo = &fakes.FakeServiceRepo{}
		servicePlanRepo = &fakes.FakeServicePlanRepo{}
		servicePlanVisibilityRepo = &fakes.FakeServicePlanVisibilityRepository{}
		orgRepo = &fakes.FakeOrgRepository{}

		actor = actors.NewServicePlanHandler(serviceRepo, servicePlanRepo, servicePlanVisibilityRepo, orgRepo)

		service := models.ServiceOffering{ServiceOfferingFields: models.ServiceOfferingFields{Label: "my-service", Guid: "my-service-guid"}}

		serviceRepo.FindServiceOfferingByLabelServiceOffering = service

		publicServicePlan = models.ServicePlanFields{
			Name:   "public-service-plan",
			Guid:   "public-service-plan-guid",
			Public: true,
		}

		privateServicePlan = models.ServicePlanFields{
			Name:   "private-service-plan",
			Guid:   "private-service-plan-guid",
			Public: false,
			OrgNames: []string{
				"org-1",
				"org-2",
			},
		}

		servicePlanRepo.SearchReturns = map[string][]models.ServicePlanFields{
			"my-service-guid": {
				publicServicePlan,
				privateServicePlan,
			},
		}
	})

	Describe(".GetSingleServicePlan", func() {
		It("Returns a single service plan", func() {
			servicePlan, err := actor.GetSingleServicePlanForService("my-service", "public-service-plan")
			Expect(err).NotTo(HaveOccurred())

			Expect(servicePlan.Name).To(Equal("public-service-plan"))
			Expect(servicePlan.Guid).To(Equal("public-service-plan-guid"))
		})
	})

	Describe(".SetServicePlanPublic", func() {
		It("sets a service plan to public", func() {
			err := actor.SetServicePlanPublic(privateServicePlan)
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
