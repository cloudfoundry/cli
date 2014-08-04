package service_builder_test

import (
	"github.com/cloudfoundry/cli/cf/actors/service_builder"
	"github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Service Builder", func() {
	var (
		serviceBuilder service_builder.ServiceBuilder
		serviceRepo    *fakes.FakeServiceRepo
		planRepo       *fakes.FakeServicePlanRepo
		visibilityRepo *fakes.FakeServicePlanVisibilityRepository
		orgRepo        *fakes.FakeOrgRepository
		service1       models.ServiceOffering
		plan1          models.ServicePlanFields
		plan2          models.ServicePlanFields
	)

	BeforeEach(func() {
		serviceRepo = &fakes.FakeServiceRepo{}
		planRepo = &fakes.FakeServicePlanRepo{}
		visibilityRepo = &fakes.FakeServicePlanVisibilityRepository{}
		orgRepo = &fakes.FakeOrgRepository{}

		serviceBuilder = service_builder.NewBuilder(serviceRepo, planRepo, visibilityRepo, orgRepo)

		service1 = models.ServiceOffering{
			ServiceOfferingFields: models.ServiceOfferingFields{
				Label:      "my-service1",
				Guid:       "service-guid1",
				BrokerGuid: "my-service-broker-guid1",
			},
		}

		serviceRepo.FindServiceOfferingByLabelName = "my-service1"
		serviceRepo.FindServiceOfferingByLabelServiceOffering = service1

		serviceRepo.ListServicesFromBrokerReturns = map[string][]models.ServiceOffering{
			"my-service-broker-guid1": []models.ServiceOffering{service1},
		}

		plan1 = models.ServicePlanFields{
			Name:                "service-plan1",
			Guid:                "service-plan1-guid",
			ServiceOfferingGuid: "service-guid1",
		}

		plan2 = models.ServicePlanFields{
			Name:                "service-plan2",
			Guid:                "service-plan2-guid",
			ServiceOfferingGuid: "service-guid1",
		}
		planRepo.SearchReturns = map[string][]models.ServicePlanFields{
			"service-guid1": []models.ServicePlanFields{plan1, plan2},
		}
		org1 := models.Organization{}
		org1.Name = "org1"
		org1.Guid = "org1-guid"

		org2 := models.Organization{}
		org2.Name = "org2"
		org2.Guid = "org2-guid"

		orgRepo.Organizations = []models.Organization{
			org1,
			org2,
		}
		visibilityRepo.ListReturns([]models.ServicePlanVisibilityFields{
			{ServicePlanGuid: "service-plan1-guid", OrganizationGuid: "org1-guid"},
			{ServicePlanGuid: "service-plan2-guid", OrganizationGuid: "org1-guid"},
		}, nil)
	})

	Describe(".AttachPlansToService", func() {
		It("returns the service, populated with plans", func() {
			service, err := serviceBuilder.AttachPlansToService(service1)
			Expect(err).NotTo(HaveOccurred())

			Expect(len(service.Plans)).To(Equal(2))
			Expect(service.Plans[0].Name).To(Equal("service-plan1"))
			Expect(service.Plans[1].Name).To(Equal("service-plan2"))
			Expect(service.Plans[0].OrgNames).To(Equal([]string{"org1"}))
		})
	})

	Describe(".GetServiceByName", func() {
		It("returns the named service, populated with plans", func() {
			services, err := serviceBuilder.GetServiceByName("my-cool-service")
			Expect(err).NotTo(HaveOccurred())

			service := services[0]
			Expect(len(service.Plans)).To(Equal(2))
			Expect(service.Plans[0].Name).To(Equal("service-plan1"))
			Expect(service.Plans[1].Name).To(Equal("service-plan2"))
			Expect(service.Plans[0].OrgNames).To(Equal([]string{"org1"}))
		})
	})

	Describe(".GetServicesForBroker", func() {
		It("returns all the services for a broker, fully populated", func() {
			services, err := serviceBuilder.GetServicesForBroker("my-service-broker-guid1")
			Expect(err).NotTo(HaveOccurred())

			service := services[0]
			Expect(service.Label).To(Equal("my-service1"))
			Expect(len(service.Plans)).To(Equal(2))
			Expect(service.Plans[0].Name).To(Equal("service-plan1"))
			Expect(service.Plans[1].Name).To(Equal("service-plan2"))
			Expect(service.Plans[0].OrgNames).To(Equal([]string{"org1"}))
		})
	})

	Describe(".GetServiceVisibleToOrg", func() {
		BeforeEach(func() {
			visibilityRepo.ListReturns([]models.ServicePlanVisibilityFields{
				{ServicePlanGuid: "service-plan1-guid", OrganizationGuid: "org1-guid"},
				{ServicePlanGuid: "service-plan2-guid", OrganizationGuid: "org2-guid"},
			}, nil)
		})

		It("Returns a service populated with plans visible to the provided org", func() {
			services, err := serviceBuilder.GetServiceVisibleToOrg("my-service1", "org1")
			Expect(err).NotTo(HaveOccurred())

			service := services[0]
			Expect(service.Label).To(Equal("my-service1"))
			Expect(len(service.Plans)).To(Equal(1))
			Expect(service.Plans[0].Name).To(Equal("service-plan1"))
			Expect(service.Plans[0].OrgNames).To(Equal([]string{"org1"}))
		})

		Context("When no plans are visible to the provided org", func() {
			It("Returns nil", func() {
				services, err := serviceBuilder.GetServiceVisibleToOrg("my-service1", "org3")
				Expect(err).NotTo(HaveOccurred())

				Expect(services).To(BeNil())
			})
		})
	})

	Describe(".GetServicesVisibleToOrg", func() {
		BeforeEach(func() {
			visibilityRepo.ListReturns([]models.ServicePlanVisibilityFields{
				{ServicePlanGuid: "service-plan1-guid", OrganizationGuid: "org1-guid"},
				{ServicePlanGuid: "service-plan2-guid", OrganizationGuid: "org2-guid"},
			}, nil)
		})

		It("Returns services with plans visible to the provided org", func() {
			services, err := serviceBuilder.GetServiceVisibleToOrg("my-service1", "org1")
			Expect(err).NotTo(HaveOccurred())

			service := services[0]
			Expect(service.Label).To(Equal("my-service1"))
			Expect(len(service.Plans)).To(Equal(1))
			Expect(service.Plans[0].Name).To(Equal("service-plan1"))
			Expect(service.Plans[0].OrgNames).To(Equal([]string{"org1"}))
		})

		Context("When no plans are visible to the provided org", func() {
			It("Returns nil", func() {
				services, err := serviceBuilder.GetServiceVisibleToOrg("my-service1", "org3")
				Expect(err).NotTo(HaveOccurred())

				Expect(services).To(BeNil())
			})
		})
	})
})

/*
	publicServicePlanVisibilityFields = models.ServicePlanVisibilityFields{
		Guid:            "public-service-plan-visibility-guid",
		ServicePlanGuid: "public-service-plan-guid",
	}

	privateServicePlanVisibilityFields = models.ServicePlanVisibilityFields{
		Guid:            "private-service-plan-visibility-guid",
		ServicePlanGuid: "private-service-plan-guid",
	}

*/
/*
	brokerRepo.FindByNameServiceBroker = serviceBroker2

	brokerRepo.ServiceBrokers = []models.ServiceBroker{
		serviceBroker1,
		serviceBroker2,
	}

	serviceRepo.ListServicesFromBrokerReturns = map[string][]models.ServiceOffering{
		"my-service-broker-guid": {
			{ServiceOfferingFields: models.ServiceOfferingFields{Guid: "a-guid", Label: "a-label"}},
		},
		"my-service-broker-guid2": {
			{ServiceOfferingFields: models.ServiceOfferingFields{Guid: "service-guid", Label: "my-service"}},
			{ServiceOfferingFields: models.ServiceOfferingFields{Guid: "service-guid2", Label: "my-service2"}},
		},
	}

	service2 := models.ServiceOffering{ServiceOfferingFields: models.ServiceOfferingFields{
		Label:      "my-service2",
		Guid:       "service-guid2",
		BrokerGuid: "my-service-broker-guid2"},
	}

	serviceRepo.FindServiceOfferingByLabelServiceOffering = service2

	servicePlanRepo.SearchReturns = map[string][]models.ServicePlanFields{
		"service-guid": {{Name: "service-plan", Guid: "service-plan-guid", ServiceOfferingGuid: "service-guid"},
			{Name: "other-plan", Guid: "other-plan-guid", ServiceOfferingGuid: "service-guid", Public: true}},
		"service-guid2": {{Name: "service-plan2", Guid: "service-plan2-guid", ServiceOfferingGuid: "service-guid2"}},
	}

	servicePlanVisibilityRepo.ListReturns([]models.ServicePlanVisibilityFields{
		{ServicePlanGuid: "service-plan2-guid", OrganizationGuid: "org-guid"},
		{ServicePlanGuid: "service-plan-guid", OrganizationGuid: "org-guid"},
		{ServicePlanGuid: "service-plan-guid", OrganizationGuid: "org2-guid"},
		{ServicePlanGuid: "service-plan2-guid", OrganizationGuid: "org2-guid"},
		{ServicePlanGuid: "other-plan-guid", OrganizationGuid: "org-guid"},
	}, nil)

	org1 := models.Organization{}
	org1.Name = "org1"
	org1.Guid = "org-guid"

	org2 := models.Organization{}
	org2.Name = "org2"
	org2.Guid = "org2-guid"

	orgRepo.Organizations = []models.Organization{
		org1,
		org2,
	}
*/
