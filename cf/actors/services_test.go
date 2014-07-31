package actors_test

import (
	"github.com/cloudfoundry/cli/cf/actors"
	"github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Services", func() {
	var (
		actor                     actors.ServiceActor
		brokerRepo                *fakes.FakeServiceBrokerRepo
		serviceRepo               *fakes.FakeServiceRepo
		servicePlanRepo           *fakes.FakeServicePlanRepo
		servicePlanVisibilityRepo *fakes.FakeServicePlanVisibilityRepository
		orgRepo                   *fakes.FakeOrgRepository
		serviceBroker1            models.ServiceBroker
		serviceBroker2            models.ServiceBroker
	)

	BeforeEach(func() {
		brokerRepo = &fakes.FakeServiceBrokerRepo{}
		serviceRepo = &fakes.FakeServiceRepo{}
		servicePlanRepo = &fakes.FakeServicePlanRepo{}
		servicePlanVisibilityRepo = &fakes.FakeServicePlanVisibilityRepository{}
		orgRepo = &fakes.FakeOrgRepository{}

		actor = actors.NewServiceHandler(brokerRepo, serviceRepo, servicePlanRepo, servicePlanVisibilityRepo, orgRepo)

		serviceBroker1 = models.ServiceBroker{Guid: "my-service-broker-guid", Name: "my-service-broker"}
		serviceBroker2 = models.ServiceBroker{Guid: "my-service-broker-guid2", Name: "my-service-broker2"}

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
	})

	Describe("FilterBrokers", func() {
		Context("when no flags are passed", func() {
			It("returns all brokers/services/plans", func() {
				brokers, err := actor.FilterBrokers("", "", "")
				Expect(err).NotTo(HaveOccurred())

				Expect(len(brokers)).To(Equal(2))
				Expect(len(brokers[0].Services)).To(Equal(1))
				Expect(len(brokers[1].Services)).To(Equal(2))

				Expect(brokers[1].Services[0].Guid).To(Equal("service-guid"))
				Expect(brokers[1].Services[0].Plans[0].Name).To(Equal("service-plan"))
				Expect(brokers[1].Services[0].Plans[1].Name).To(Equal("other-plan"))
				Expect(brokers[1].Services[1].Guid).To(Equal("service-guid2"))
				Expect(brokers[1].Services[1].Plans[0].Name).To(Equal("service-plan2"))
				Expect(brokers[1].Services[1].Plans[0].OrgNames).To(Equal([]string{"org1", "org2"}))

			})

		})

		Context("when the -b flag is passed", func() {
			It("returns a single broker contained in a slice with all dependencies populated", func() {
				brokers, err := actor.FilterBrokers("my-service-broker2", "", "")
				Expect(err).NotTo(HaveOccurred())

				Expect(len(brokers)).To(Equal(1))
				Expect(len(brokers[0].Services)).To(Equal(2))

				Expect(brokers[0].Services[0].Guid).To(Equal("service-guid"))
				Expect(brokers[0].Services[0].Plans[0].Name).To(Equal("service-plan"))
				Expect(brokers[0].Services[0].Plans[1].Name).To(Equal("other-plan"))
				Expect(brokers[0].Services[1].Guid).To(Equal("service-guid2"))
				Expect(brokers[0].Services[1].Plans[0].Name).To(Equal("service-plan2"))
				Expect(brokers[0].Services[1].Plans[0].OrgNames).To(Equal([]string{"org1", "org2"}))
			})
		})

		Context("when the -e flag is passed", func() {
			It("returns a single broker containing a single service with all dependencies populated", func() {
				brokers, err := actor.FilterBrokers("", "my-service2", "")
				Expect(err).NotTo(HaveOccurred())

				Expect(len(brokers)).To(Equal(1))
				Expect(len(brokers[0].Services)).To(Equal(1))

				Expect(brokers[0].Services[0].Guid).To(Equal("service-guid2"))
				Expect(brokers[0].Services[0].Plans[0].Name).To(Equal("service-plan2"))
				Expect(brokers[0].Services[0].Plans[0].OrgNames).To(Equal([]string{"org1", "org2"}))
			})
		})

		Context("when the -o flag is passed", func() {
			It("returns a slice of brokers containing Services/Service Plans visible to the org", func() {
				servicePlanRepo.SearchReturns = map[string][]models.ServicePlanFields{
					"service-guid": {
						{Name: "service-plan", Guid: "service-plan-guid", ServiceOfferingGuid: "service-guid"},
						{Name: "other-plan", Guid: "other-plan-guid", ServiceOfferingGuid: "service-guid", Public: true},
						{Name: "private-plan", Guid: "private-plan-guid", ServiceOfferingGuid: "service-guid", Public: false},
					},
				}
				serviceRepo.GetServiceOfferingByGuidReturns.ServiceOffering = models.ServiceOffering{
					ServiceOfferingFields: models.ServiceOfferingFields{Guid: "service-guid", Label: "my-service", BrokerGuid: "my-service-broker-guid"},
				}
				serviceRepo.GetServiceOfferingByGuidReturns.Error = nil
				brokerRepo.FindByGuidServiceBroker = models.ServiceBroker{Guid: "my-service-broker-guid", Name: "my-service-broker"}

				brokers, err := actor.FilterBrokers("", "", "org1")
				Expect(err).NotTo(HaveOccurred())

				Expect(len(brokers)).To(Equal(1))
				Expect(len(brokers[0].Services)).To(Equal(1))
				Expect(len(brokers[0].Services[0].Plans)).To(Equal(2))

				Expect(brokers[0].Services[0].Guid).To(Equal("service-guid"))
				Expect(brokers[0].Services[0].Plans[0].Name).To(Equal("service-plan"))
				Expect(brokers[0].Services[0].Plans[0].OrgNames).To(Equal([]string{"org1", "org2"}))
				Expect(brokers[0].Services[0].Plans[1].Name).To(Equal("other-plan"))
			})

			It("ignores any service that does not have a a broker guid attached", func() {
				//The service offering fixtures we use don't have broker guids attached, and thus we ignore them.
				brokers, err := actor.FilterBrokers("", "", "org1")
				Expect(err).NotTo(HaveOccurred())

				Expect(len(brokers)).To(Equal(0))
			})
		})

		Context("when the -b AND the -e flags are passed", func() {
			It("returns the intersection set", func() {
				brokers, err := actor.FilterBrokers("my-service-broker2", "my-service2", "")
				Expect(err).To(BeNil())
				Expect(len(brokers)).To(Equal(1))
				Expect(len(brokers[0].Services)).To(Equal(1))

				Expect(brokers[0].Services[0].Guid).To(Equal("service-guid2"))
				Expect(brokers[0].Services[0].Plans[0].Name).To(Equal("service-plan2"))
				Expect(brokers[0].Services[0].Plans[0].OrgNames).To(Equal([]string{"org1", "org2"}))
			})

			Context("when the -b AND -e intersection is the empty set", func() {
				It("returns an empty set", func() {
					brokerRepo.FindByNameServiceBroker = serviceBroker1
					brokers, err := actor.FilterBrokers("my-service-broker", "my-service2", "")

					Expect(len(brokers)).To(Equal(0))
					Expect(err).To(BeNil())
				})
			})
		})

		Context("when the -b AND the -o flags are passed", func() {
			It("returns the intersection set", func() {
				servicePlanRepo.SearchReturns = map[string][]models.ServicePlanFields{
					"service-guid": {
						{Name: "service-plan", Guid: "service-plan-guid", ServiceOfferingGuid: "service-guid"},
						{Name: "other-plan", Guid: "other-plan-guid", ServiceOfferingGuid: "service-guid", Public: true},
						{Name: "private-plan", Guid: "private-plan-guid", ServiceOfferingGuid: "service-guid", Public: false},
					},
				}
				serviceRepo.GetServiceOfferingByGuidReturns.ServiceOffering = models.ServiceOffering{
					ServiceOfferingFields: models.ServiceOfferingFields{Guid: "service-guid", Label: "my-service", BrokerGuid: "my-service-broker-guid"},
				}
				serviceRepo.GetServiceOfferingByGuidReturns.Error = nil
				brokerRepo.FindByGuidServiceBroker = models.ServiceBroker{Guid: "my-service-broker-guid", Name: "my-service-broker"}

				brokers, err := actor.FilterBrokers("my-service-broker", "", "org1")
				Expect(err).NotTo(HaveOccurred())

				Expect(len(brokers)).To(Equal(1))
				Expect(len(brokers[0].Services)).To(Equal(1))
				Expect(len(brokers[0].Services[0].Plans)).To(Equal(2))

				Expect(brokers[0].Services[0].Guid).To(Equal("service-guid"))
				Expect(brokers[0].Services[0].Plans[0].Name).To(Equal("service-plan"))
				Expect(brokers[0].Services[0].Plans[0].OrgNames).To(Equal([]string{"org1", "org2"}))
				Expect(brokers[0].Services[0].Plans[1].Name).To(Equal("other-plan"))
			})
		})

		Context("when the -e AND the -o flags are passed", func() {
			It("returns the intersection set", func() {
				servicePlanRepo.SearchReturns = map[string][]models.ServicePlanFields{
					"service-guid": {
						{Name: "service-plan", Guid: "service-plan-guid", ServiceOfferingGuid: "service-guid"},
						{Name: "other-plan", Guid: "other-plan-guid", ServiceOfferingGuid: "service-guid", Public: true},
						{Name: "another-plan", Guid: "another-plan-guid", ServiceOfferingGuid: "service-guid2", Public: true},
						{Name: "private-plan", Guid: "private-plan-guid", ServiceOfferingGuid: "service-guid", Public: false},
					},
				}

				service := models.ServiceOffering{
					ServiceOfferingFields: models.ServiceOfferingFields{Guid: "service-guid", Label: "my-service", BrokerGuid: "my-service-broker-guid"},
				}
				serviceRepo.GetServiceOfferingByGuidReturns.ServiceOffering = service
				serviceRepo.GetServiceOfferingByGuidReturns.Error = nil
				brokerRepo.FindByGuidServiceBroker = models.ServiceBroker{Guid: "my-service-broker-guid", Name: "my-service-broker"}
				serviceRepo.FindServiceOfferingByLabelServiceOffering = service

				brokers, err := actor.FilterBrokers("", "my-service", "org1")
				Expect(err).NotTo(HaveOccurred())

				Expect(len(brokers)).To(Equal(1))
				Expect(len(brokers[0].Services)).To(Equal(1))
				Expect(len(brokers[0].Services[0].Plans)).To(Equal(2))

				Expect(brokers[0].Services[0].Guid).To(Equal("service-guid"))
				Expect(brokers[0].Services[0].Plans[0].Name).To(Equal("service-plan"))
				Expect(brokers[0].Services[0].Plans[0].OrgNames).To(Equal([]string{"org1", "org2"}))
				Expect(brokers[0].Services[0].Plans[1].Name).To(Equal("other-plan"))
			})
		})

		Context("when the -b AND -e AND the -o flags are passed", func() {
		})
	})
})
