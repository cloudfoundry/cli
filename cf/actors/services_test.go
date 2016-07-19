package actors_test

import (
	"code.cloudfoundry.org/cli/cf/actors"
	"code.cloudfoundry.org/cli/cf/actors/brokerbuilder/brokerbuilderfakes"
	"code.cloudfoundry.org/cli/cf/actors/servicebuilder/servicebuilderfakes"
	"code.cloudfoundry.org/cli/cf/api/organizations/organizationsfakes"
	"code.cloudfoundry.org/cli/cf/errors"
	"code.cloudfoundry.org/cli/cf/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Services", func() {
	var (
		actor          actors.ServiceActor
		brokerBuilder  *brokerbuilderfakes.FakeBrokerBuilder
		serviceBuilder *servicebuilderfakes.FakeServiceBuilder
		orgRepo        *organizationsfakes.FakeOrganizationRepository
		serviceBroker1 models.ServiceBroker
		service1       models.ServiceOffering
	)

	BeforeEach(func() {
		orgRepo = new(organizationsfakes.FakeOrganizationRepository)
		brokerBuilder = new(brokerbuilderfakes.FakeBrokerBuilder)
		serviceBuilder = new(servicebuilderfakes.FakeServiceBuilder)

		actor = actors.NewServiceHandler(orgRepo, brokerBuilder, serviceBuilder)

		serviceBroker1 = models.ServiceBroker{GUID: "my-service-broker-guid1", Name: "my-service-broker1"}

		service1 = models.ServiceOffering{ServiceOfferingFields: models.ServiceOfferingFields{
			Label:      "my-service1",
			GUID:       "service-guid1",
			BrokerGUID: "my-service-broker-guid1"},
		}

		org1 := models.Organization{}
		org1.Name = "org1"
		org1.GUID = "org-guid"

		org2 := models.Organization{}
		org2.Name = "org2"
		org2.GUID = "org2-guid"
	})

	Describe("FilterBrokers", func() {
		Context("when no flags are passed", func() {
			It("returns all brokers", func() {
				returnedBrokers := []models.ServiceBroker{serviceBroker1}
				brokerBuilder.GetAllServiceBrokersReturns(returnedBrokers, nil)

				brokers, err := actor.FilterBrokers("", "", "")
				Expect(err).NotTo(HaveOccurred())

				Expect(len(brokers)).To(Equal(1))
			})
		})

		Context("when the -b flag is passed", func() {
			It("returns a single broker contained in a slice with all dependencies populated", func() {
				returnedBroker := serviceBroker1
				brokerBuilder.GetBrokerWithAllServicesReturns(returnedBroker, nil)

				brokers, err := actor.FilterBrokers("my-service-broker1", "", "")
				Expect(err).NotTo(HaveOccurred())

				Expect(len(brokers)).To(Equal(1))
			})
		})

		Context("when the -e flag is passed", func() {
			It("returns a single broker containing a single service", func() {
				serviceBroker1.Services = []models.ServiceOffering{service1}
				returnedBroker := serviceBroker1
				brokerBuilder.GetBrokerWithSpecifiedServiceReturns(returnedBroker, nil)

				brokers, err := actor.FilterBrokers("", "my-service1", "")
				Expect(err).NotTo(HaveOccurred())

				Expect(len(brokers)).To(Equal(1))
				Expect(len(brokers[0].Services)).To(Equal(1))

				Expect(brokers[0].Services[0].GUID).To(Equal("service-guid1"))
			})
		})

		Context("when the -o flag is passed", func() {
			It("returns an error if the org does not actually exist", func() {
				orgRepo.FindByNameReturns(models.Organization{}, errors.NewModelNotFoundError("organization", "org-that-shall-not-be-found"))
				_, err := actor.FilterBrokers("", "", "org-that-shall-not-be-found")

				Expect(err).To(HaveOccurred())
			})

			It("returns a slice of brokers containing Services/Service Plans visible to the org", func() {
				serviceBroker1.Services = []models.ServiceOffering{service1}
				returnedBroker := []models.ServiceBroker{serviceBroker1}

				serviceBuilder.GetServicesVisibleToOrgReturns([]models.ServiceOffering{service1}, nil)
				brokerBuilder.AttachBrokersToServicesReturns(returnedBroker, nil)

				brokers, err := actor.FilterBrokers("", "", "org1")
				Expect(err).NotTo(HaveOccurred())

				orgName := serviceBuilder.GetServicesVisibleToOrgArgsForCall(0)
				Expect(orgName).To(Equal("org1"))

				Expect(len(brokers)).To(Equal(1))
				Expect(len(brokers[0].Services)).To(Equal(1))
				Expect(brokers[0].Services[0].GUID).To(Equal("service-guid1"))
			})
		})

		Context("when the -b AND the -e flags are passed", func() {
			It("returns the intersection set", func() {
				serviceBroker1.Services = []models.ServiceOffering{service1}
				returnedBroker := serviceBroker1
				brokerBuilder.GetBrokerWithSpecifiedServiceReturns(returnedBroker, nil)

				brokers, err := actor.FilterBrokers("my-service-broker1", "my-service1", "")
				Expect(err).NotTo(HaveOccurred())

				Expect(len(brokers)).To(Equal(1))
				Expect(len(brokers[0].Services)).To(Equal(1))

				Expect(brokers[0].Services[0].Label).To(Equal("my-service1"))
				Expect(brokers[0].Services[0].GUID).To(Equal("service-guid1"))
			})

			Context("when the -b AND -e intersection is the empty set", func() {
				It("returns an empty set", func() {
					brokerBuilder.GetBrokerWithSpecifiedServiceReturns(models.ServiceBroker{}, nil)
					brokers, err := actor.FilterBrokers("my-service-broker", "my-service2", "")

					Expect(len(brokers)).To(Equal(0))
					Expect(err).To(BeNil())
				})
			})
		})

		Context("when the -b AND the -o flags are passed", func() {
			It("returns the intersection set", func() {
				serviceBroker1.Services = []models.ServiceOffering{service1}
				returnedBroker := serviceBroker1

				serviceBuilder.GetServiceVisibleToOrgReturns(service1, nil)
				brokerBuilder.AttachSpecificBrokerToServicesReturns(returnedBroker, nil)

				brokers, err := actor.FilterBrokers("my-service-broker", "", "org1")
				Expect(err).NotTo(HaveOccurred())

				Expect(len(brokers)).To(Equal(1))
				Expect(len(brokers[0].Services)).To(Equal(1))

				Expect(brokers[0].Services[0].Label).To(Equal("my-service1"))
				Expect(brokers[0].Services[0].GUID).To(Equal("service-guid1"))
			})
		})

		Context("when the -e AND the -o flags are passed", func() {
			It("returns the intersection set", func() {
				serviceBroker1.Services = []models.ServiceOffering{service1}
				returnedBrokers := []models.ServiceBroker{serviceBroker1}

				serviceBuilder.GetServicesVisibleToOrgReturns([]models.ServiceOffering{service1}, nil)
				brokerBuilder.AttachBrokersToServicesReturns(returnedBrokers, nil)

				brokers, err := actor.FilterBrokers("", "my-service1", "org1")
				Expect(err).NotTo(HaveOccurred())

				Expect(len(brokers)).To(Equal(1))
				Expect(len(brokers[0].Services)).To(Equal(1))

				Expect(brokers[0].Services[0].Label).To(Equal("my-service1"))
				Expect(brokers[0].Services[0].GUID).To(Equal("service-guid1"))
			})
		})

		Context("when the -b AND -e AND the -o flags are passed", func() {
			It("returns the intersection set", func() {
				serviceBroker1.Services = []models.ServiceOffering{service1}
				returnedBroker := serviceBroker1

				serviceBuilder.GetServicesVisibleToOrgReturns([]models.ServiceOffering{service1}, nil)
				brokerBuilder.AttachSpecificBrokerToServicesReturns(returnedBroker, nil)

				brokers, err := actor.FilterBrokers("my-service-broker1", "my-service1", "org1")
				Expect(err).NotTo(HaveOccurred())

				Expect(len(brokers)).To(Equal(1))
				Expect(len(brokers[0].Services)).To(Equal(1))

				Expect(brokers[0].Services[0].Label).To(Equal("my-service1"))
				Expect(brokers[0].Services[0].GUID).To(Equal("service-guid1"))
			})
		})
	})
})
