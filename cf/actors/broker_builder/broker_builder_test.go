package broker_builder_test

import (
	"errors"

	"github.com/cloudfoundry/cli/cf/actors/broker_builder"
	"github.com/cloudfoundry/cli/cf/api/fakes"
	"github.com/cloudfoundry/cli/cf/models"

	fake_service_builder "github.com/cloudfoundry/cli/cf/actors/service_builder/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Broker Builder", func() {
	var (
		brokerBuilder broker_builder.BrokerBuilder

		serviceBuilder *fake_service_builder.FakeServiceBuilder
		brokerRepo     *fakes.FakeServiceBrokerRepo

		serviceBroker1 models.ServiceBroker

		services           models.ServiceOfferings
		service1           models.ServiceOffering
		service2           models.ServiceOffering
		service3           models.ServiceOffering
		publicServicePlan  models.ServicePlanFields
		privateServicePlan models.ServicePlanFields
	)

	BeforeEach(func() {
		brokerRepo = &fakes.FakeServiceBrokerRepo{}
		serviceBuilder = &fake_service_builder.FakeServiceBuilder{}
		brokerBuilder = broker_builder.NewBuilder(brokerRepo, serviceBuilder)

		serviceBroker1 = models.ServiceBroker{Guid: "my-service-broker-guid", Name: "my-service-broker"}

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

		service1 = models.ServiceOffering{
			ServiceOfferingFields: models.ServiceOfferingFields{
				Label:      "my-public-service",
				Guid:       "my-public-service-guid",
				BrokerGuid: "my-service-broker-guid",
			},
			Plans: []models.ServicePlanFields{
				publicServicePlan,
				privateServicePlan,
			},
		}

		service2 = models.ServiceOffering{
			ServiceOfferingFields: models.ServiceOfferingFields{
				Label:      "my-other-public-service",
				Guid:       "my-other-public-service-guid",
				BrokerGuid: "my-service-broker-guid",
			},
			Plans: []models.ServicePlanFields{
				publicServicePlan,
				privateServicePlan,
			},
		}

		service3 = models.ServiceOffering{
			ServiceOfferingFields: models.ServiceOfferingFields{
				Label:      "my-other-public-service",
				Guid:       "my-other-public-service-guid",
				BrokerGuid: "my-service-broker-guid",
			},
			Plans: []models.ServicePlanFields{
				publicServicePlan,
				privateServicePlan,
			},
		}

		services = models.ServiceOfferings(
			[]models.ServiceOffering{
				service1,
				service2,
			})

		brokerRepo.FindByGuidServiceBroker = serviceBroker1
	})

	Describe(".AttachBrokersToServices", func() {
		It("attaches brokers to an array of services", func() {

			brokers, err := brokerBuilder.AttachBrokersToServices(services)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(brokers)).To(Equal(1))
			Expect(brokers[0].Name).To(Equal("my-service-broker"))
			Expect(brokers[0].Services[0].Label).To(Equal("my-public-service"))
			Expect(len(brokers[0].Services[0].Plans)).To(Equal(2))
			Expect(brokers[0].Services[1].Label).To(Equal("my-other-public-service"))
			Expect(len(brokers[0].Services[0].Plans)).To(Equal(2))
		})

		It("skips services that have no associated broker, e.g. v1 services", func() {
			brokerlessService := models.ServiceOffering{
				ServiceOfferingFields: models.ServiceOfferingFields{
					Label: "lonely-v1-service",
					Guid:  "i-am-sad-and-old",
				},
				Plans: []models.ServicePlanFields{
					publicServicePlan,
					privateServicePlan,
				},
			}
			services = models.ServiceOfferings(
				[]models.ServiceOffering{
					service1,
					service2,
					brokerlessService,
				})

			brokers, err := brokerBuilder.AttachBrokersToServices(services)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(brokers)).To(Equal(1))
			Expect(brokers[0].Name).To(Equal("my-service-broker"))
			Expect(brokers[0].Services[0].Label).To(Equal("my-public-service"))
			Expect(len(brokers[0].Services[0].Plans)).To(Equal(2))
			Expect(brokers[0].Services[1].Label).To(Equal("my-other-public-service"))
			Expect(len(brokers[0].Services[0].Plans)).To(Equal(2))
		})
	})

	Describe(".AttachSpecificBrokerToServices", func() {
		BeforeEach(func() {
			service3 = models.ServiceOffering{
				ServiceOfferingFields: models.ServiceOfferingFields{
					Label:      "my-other-public-service",
					Guid:       "my-other-public-service-guid",
					BrokerGuid: "my-other-service-broker-guid",
				},
				Plans: []models.ServicePlanFields{
					publicServicePlan,
					privateServicePlan,
				},
			}
			services = append(services, service3)
		})

		It("attaches a single broker to only services that match", func() {
			serviceBroker1.Services = models.ServiceOfferings{}
			brokerRepo.FindByNameServiceBroker = serviceBroker1
			broker, err := brokerBuilder.AttachSpecificBrokerToServices("my-service-broker", services)

			Expect(err).NotTo(HaveOccurred())
			Expect(broker.Name).To(Equal("my-service-broker"))
			Expect(broker.Services[0].Label).To(Equal("my-public-service"))
			Expect(len(broker.Services[0].Plans)).To(Equal(2))
			Expect(broker.Services[1].Label).To(Equal("my-other-public-service"))
			Expect(len(broker.Services[0].Plans)).To(Equal(2))
			Expect(len(broker.Services)).To(Equal(2))
		})
	})

	Describe(".GetAllServiceBrokers", func() {
		It("returns an error if we cannot list all brokers", func() {
			brokerRepo.ListErr = true

			_, err := brokerBuilder.GetAllServiceBrokers()
			Expect(err).To(HaveOccurred())
		})

		It("returns an error if we cannot list the services for a broker", func() {
			brokerRepo.ServiceBrokers = []models.ServiceBroker{serviceBroker1}
			serviceBuilder.GetServicesForManyBrokersReturns(nil, errors.New("Cannot find services"))

			_, err := brokerBuilder.GetAllServiceBrokers()
			Expect(err).To(HaveOccurred())
		})

		It("returns all service brokers populated with their services", func() {
			brokerRepo.ServiceBrokers = []models.ServiceBroker{serviceBroker1}
			serviceBuilder.GetServicesForManyBrokersReturns(services, nil)

			brokers, err := brokerBuilder.GetAllServiceBrokers()
			Expect(err).NotTo(HaveOccurred())
			Expect(len(brokers)).To(Equal(1))
			Expect(brokers[0].Name).To(Equal("my-service-broker"))
			Expect(brokers[0].Services[0].Label).To(Equal("my-public-service"))
			Expect(len(brokers[0].Services[0].Plans)).To(Equal(2))
			Expect(brokers[0].Services[1].Label).To(Equal("my-other-public-service"))
			Expect(len(brokers[0].Services[0].Plans)).To(Equal(2))
		})
	})

	Describe(".GetBrokerWithAllServices", func() {
		It("returns a service broker populated with their services", func() {
			brokerRepo.FindByNameServiceBroker = serviceBroker1
			serviceBuilder.GetServicesForBrokerReturns(services, nil)

			broker, err := brokerBuilder.GetBrokerWithAllServices("my-service-broker")
			Expect(err).NotTo(HaveOccurred())
			Expect(broker.Name).To(Equal("my-service-broker"))
			Expect(broker.Services[0].Label).To(Equal("my-public-service"))
			Expect(len(broker.Services[0].Plans)).To(Equal(2))
			Expect(broker.Services[1].Label).To(Equal("my-other-public-service"))
			Expect(len(broker.Services[0].Plans)).To(Equal(2))
		})
	})

	Describe(".GetBrokerWithSpecifiedService", func() {
		It("returns an error if a broker containeing the specific service cannot be found", func() {
			serviceBuilder.GetServiceByNameWithPlansWithOrgNamesReturns(models.ServiceOffering{}, errors.New("Asplosions"))
			_, err := brokerBuilder.GetBrokerWithSpecifiedService("totally-not-a-service")

			Expect(err).To(HaveOccurred())
		})

		It("returns the service broker populated with the specific service", func() {
			serviceBuilder.GetServiceByNameWithPlansWithOrgNamesReturns(service1, nil)
			brokerRepo.FindByGuidServiceBroker = serviceBroker1

			broker, err := brokerBuilder.GetBrokerWithSpecifiedService("my-public-service")
			Expect(err).NotTo(HaveOccurred())
			Expect(broker.Name).To(Equal("my-service-broker"))
			Expect(len(broker.Services)).To(Equal(1))
			Expect(broker.Services[0].Label).To(Equal("my-public-service"))
			Expect(len(broker.Services[0].Plans)).To(Equal(2))
		})
	})
})
