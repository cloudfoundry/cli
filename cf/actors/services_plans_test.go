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

		service models.ServiceOffering
	)

	BeforeEach(func() {
		serviceRepo = &fakes.FakeServiceRepo{}
		servicePlanRepo = &fakes.FakeServicePlanRepo{}
		servicePlanVisibilityRepo = &fakes.FakeServicePlanVisibilityRepository{}
		orgRepo = &fakes.FakeOrgRepository{}

		actor = actors.NewServicePlanHandler(serviceRepo, servicePlanRepo, servicePlanVisibilityRepo, orgRepo)

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

		service = models.ServiceOffering{
			ServiceOfferingFields: models.ServiceOfferingFields{
				Label: "my-service",
				Guid:  "my-service-guid",
			},
			Plans: []models.ServicePlanFields{
				publicServicePlan,
				privateServicePlan,
			},
		}

		serviceRepo.FindServiceOfferingByLabelServiceOffering = service

		servicePlanRepo.SearchReturns = map[string][]models.ServicePlanFields{
			"my-service-guid": {
				publicServicePlan,
				privateServicePlan,
			},
		}
	})

	Describe(".GetServiceWithSinglePlan", func() {
		It("Returns a single service", func() {
			serviceOffering, err := actor.GetServiceWithSinglePlan("my-service", "public-service-plan")
			Expect(err).NotTo(HaveOccurred())

			Expect(len(serviceOffering.Plans)).To(Equal(1))
			Expect(serviceOffering.Plans[0].Name).To(Equal("public-service-plan"))
			Expect(serviceOffering.Plans[0].Guid).To(Equal("public-service-plan-guid"))
		})
	})

	Describe(".UpdateServicePlanAvailability", func() {
		It("sets a service plan to public", func() {
			err := actor.UpdateServicePlanAvailability(service, true)
			Expect(err).ToNot(HaveOccurred())

			servicePlan, serviceGuid, public := servicePlanRepo.UpdateArgsForCall(0)
			Expect(servicePlan.Public).To(BeTrue())
			Expect(serviceGuid).To(Equal("my-service-guid"))
			Expect(public).To(BeTrue())
		})
	})

	Describe(".RemoveServicePlanVisabilities", func() {
		It("removes all service plan visabilites for a service plan", func() {
			err := actor.RemoveServicePlanVisabilities(service)
			Expect(err).ToNot(HaveOccurred())

			
		})
	})
})
