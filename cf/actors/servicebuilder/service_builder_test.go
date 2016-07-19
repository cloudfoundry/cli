package servicebuilder_test

import (
	"errors"

	"code.cloudfoundry.org/cli/cf/actors/planbuilder/planbuilderfakes"
	"code.cloudfoundry.org/cli/cf/actors/servicebuilder"
	"code.cloudfoundry.org/cli/cf/api/apifakes"

	"code.cloudfoundry.org/cli/cf/models"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Service Builder", func() {
	var (
		planBuilder     *planbuilderfakes.FakePlanBuilder
		serviceBuilder  servicebuilder.ServiceBuilder
		serviceRepo     *apifakes.FakeServiceRepository
		service1        models.ServiceOffering
		service2        models.ServiceOffering
		v1Service       models.ServiceOffering
		planWithoutOrgs models.ServicePlanFields
		plan1           models.ServicePlanFields
		plan2           models.ServicePlanFields
		plan3           models.ServicePlanFields
	)

	BeforeEach(func() {
		serviceRepo = new(apifakes.FakeServiceRepository)
		planBuilder = new(planbuilderfakes.FakePlanBuilder)

		serviceBuilder = servicebuilder.NewBuilder(serviceRepo, planBuilder)
		service1 = models.ServiceOffering{
			ServiceOfferingFields: models.ServiceOfferingFields{
				Label:      "my-service1",
				GUID:       "service-guid1",
				BrokerGUID: "my-service-broker-guid1",
			},
		}

		service2 = models.ServiceOffering{
			ServiceOfferingFields: models.ServiceOfferingFields{
				Label:      "my-service2",
				GUID:       "service-guid2",
				BrokerGUID: "my-service-broker-guid2",
			},
		}

		v1Service = models.ServiceOffering{
			ServiceOfferingFields: models.ServiceOfferingFields{
				Label:      "v1Service",
				GUID:       "v1Service-guid",
				BrokerGUID: "my-service-broker-guid1",
				Provider:   "IAmV1",
			},
		}

		serviceOfferings := models.ServiceOfferings([]models.ServiceOffering{service1, v1Service})
		serviceRepo.FindServiceOfferingsByLabelReturns(serviceOfferings, nil)
		serviceRepo.GetServiceOfferingByGUIDReturns(service1, nil)
		serviceRepo.ListServicesFromBrokerReturns([]models.ServiceOffering{service1}, nil)
		serviceRepo.ListServicesFromManyBrokersReturns([]models.ServiceOffering{service1, service2}, nil)

		plan1 = models.ServicePlanFields{
			Name:                "service-plan1",
			GUID:                "service-plan1-guid",
			ServiceOfferingGUID: "service-guid1",
			OrgNames:            []string{"org1", "org2"},
		}

		plan2 = models.ServicePlanFields{
			Name:                "service-plan2",
			GUID:                "service-plan2-guid",
			ServiceOfferingGUID: "service-guid1",
		}

		plan3 = models.ServicePlanFields{
			Name:                "service-plan3",
			GUID:                "service-plan3-guid",
			ServiceOfferingGUID: "service-guid2",
		}

		planWithoutOrgs = models.ServicePlanFields{
			Name:                "service-plan-without-orgs",
			GUID:                "service-plan-without-orgs-guid",
			ServiceOfferingGUID: "service-guid1",
		}

		planBuilder.GetPlansVisibleToOrgReturns([]models.ServicePlanFields{plan1, plan2}, nil)
		planBuilder.GetPlansForServiceWithOrgsReturns([]models.ServicePlanFields{plan1, plan2}, nil)
		planBuilder.GetPlansForManyServicesWithOrgsReturns([]models.ServicePlanFields{plan1, plan2, plan3}, nil)
		planBuilder.GetPlansForServiceForOrgReturns([]models.ServicePlanFields{plan1, plan2}, nil)
	})

	Describe(".GetServicesForSpace", func() {
		BeforeEach(func() {
			serviceRepo.GetServiceOfferingsForSpaceReturns([]models.ServiceOffering{service1, service1}, nil)
		})

		It("returns the services for the space", func() {
			services, err := serviceBuilder.GetServicesForSpace("spaceGUID")
			Expect(err).NotTo(HaveOccurred())

			Expect(len(services)).To(Equal(2))
		})
	})

	Describe(".GetServicesForSpaceWithPlans", func() {
		BeforeEach(func() {
			serviceRepo.GetServiceOfferingsForSpaceReturns([]models.ServiceOffering{service1, service1}, nil)
			planBuilder.GetPlansForServiceReturns([]models.ServicePlanFields{planWithoutOrgs}, nil)
		})

		It("returns the services for the space, populated with plans", func() {
			services, err := serviceBuilder.GetServicesForSpaceWithPlans("spaceGUID")
			Expect(err).NotTo(HaveOccurred())

			Expect(len(services)).To(Equal(2))
			Expect(services[0].Plans[0]).To(Equal(planWithoutOrgs))
			Expect(services[1].Plans[0]).To(Equal(planWithoutOrgs))
		})
	})

	Describe(".GetAllServices", func() {
		BeforeEach(func() {
			serviceRepo.GetAllServiceOfferingsReturns([]models.ServiceOffering{service1, service1}, nil)
		})

		It("returns the named service, populated with plans", func() {
			services, err := serviceBuilder.GetAllServices()
			Expect(err).NotTo(HaveOccurred())

			Expect(len(services)).To(Equal(2))
		})
	})

	Describe(".GetAllServicesWithPlans", func() {
		BeforeEach(func() {
			serviceRepo.GetAllServiceOfferingsReturns([]models.ServiceOffering{service1, service1}, nil)
			planBuilder.GetPlansForServiceReturns([]models.ServicePlanFields{plan1}, nil)
		})

		It("returns the named service, populated with plans", func() {
			services, err := serviceBuilder.GetAllServicesWithPlans()
			Expect(err).NotTo(HaveOccurred())

			Expect(len(services)).To(Equal(2))
			Expect(services[0].Plans[0]).To(Equal(plan1))
		})
	})

	Describe(".GetServiceByNameWithPlans", func() {
		BeforeEach(func() {
			planBuilder.GetPlansForServiceReturns([]models.ServicePlanFields{plan2}, nil)
		})

		It("returns the named service, populated with plans", func() {
			service, err := serviceBuilder.GetServiceByNameWithPlans("my-service1")
			Expect(err).NotTo(HaveOccurred())

			Expect(len(service.Plans)).To(Equal(1))
			Expect(service.Plans[0].Name).To(Equal("service-plan2"))
			Expect(service.Plans[0].OrgNames).To(BeNil())
		})
	})

	Describe(".GetServiceByNameWithPlansWithOrgNames", func() {
		It("returns the named service, populated with plans", func() {
			service, err := serviceBuilder.GetServiceByNameWithPlansWithOrgNames("my-service1")
			Expect(err).NotTo(HaveOccurred())

			Expect(len(service.Plans)).To(Equal(2))
			Expect(service.Plans[0].Name).To(Equal("service-plan1"))
			Expect(service.Plans[1].Name).To(Equal("service-plan2"))
			Expect(service.Plans[0].OrgNames).To(Equal([]string{"org1", "org2"}))
		})
	})

	Describe(".GetServiceByNameForSpace", func() {
		Context("mixed v2 and v1 services", func() {
			BeforeEach(func() {
				service2 := models.ServiceOffering{
					ServiceOfferingFields: models.ServiceOfferingFields{
						Label: "service",
						GUID:  "service-guid-v2",
					},
				}

				service1 := models.ServiceOffering{
					ServiceOfferingFields: models.ServiceOfferingFields{
						Label:    "service",
						GUID:     "service-guid",
						Provider: "a provider",
					},
				}

				serviceRepo.FindServiceOfferingsForSpaceByLabelReturns(
					models.ServiceOfferings([]models.ServiceOffering{service1, service2}),
					nil,
				)
			})

			It("returns the nv2 service", func() {
				service, err := serviceBuilder.GetServiceByNameForSpace("service", "spaceGUID")
				Expect(err).NotTo(HaveOccurred())

				Expect(len(service.Plans)).To(Equal(0))
				Expect(service.GUID).To(Equal("service-guid-v2"))
			})
		})

		Context("v2 services", func() {
			BeforeEach(func() {
				service := models.ServiceOffering{
					ServiceOfferingFields: models.ServiceOfferingFields{
						Label: "service",
						GUID:  "service-guid",
					},
				}

				serviceRepo.FindServiceOfferingsForSpaceByLabelReturns(
					models.ServiceOfferings([]models.ServiceOffering{service}),
					nil,
				)
			})

			It("returns the named service", func() {
				service, err := serviceBuilder.GetServiceByNameForSpace("service", "spaceGUID")
				Expect(err).NotTo(HaveOccurred())

				Expect(len(service.Plans)).To(Equal(0))
				Expect(service.GUID).To(Equal("service-guid"))
			})
		})

		Context("v1 services", func() {
			BeforeEach(func() {
				service := models.ServiceOffering{
					ServiceOfferingFields: models.ServiceOfferingFields{
						Label:    "service",
						GUID:     "service-guid",
						Provider: "a provider",
					},
				}

				serviceRepo.FindServiceOfferingsForSpaceByLabelReturns(
					models.ServiceOfferings([]models.ServiceOffering{service}),
					nil,
				)
			})

			It("returns the an error", func() {
				service, err := serviceBuilder.GetServiceByNameForSpace("service", "spaceGUID")
				Expect(service).To(Equal(models.ServiceOffering{}))
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe(".GetServiceByNameForSpaceWithPlans", func() {
		BeforeEach(func() {
			service := models.ServiceOffering{
				ServiceOfferingFields: models.ServiceOfferingFields{
					Label: "serviceWithPlans",
				},
			}

			serviceRepo.FindServiceOfferingsForSpaceByLabelReturns(
				models.ServiceOfferings([]models.ServiceOffering{service}),
				nil,
			)
			planBuilder.GetPlansForServiceReturns([]models.ServicePlanFields{planWithoutOrgs}, nil)
		})

		It("returns the named service", func() {
			service, err := serviceBuilder.GetServiceByNameForSpaceWithPlans("serviceWithPlans", "spaceGUID")
			Expect(err).NotTo(HaveOccurred())

			Expect(len(service.Plans)).To(Equal(1))
			Expect(service.Plans[0].Name).To(Equal("service-plan-without-orgs"))
			Expect(service.Plans[0].OrgNames).To(BeNil())
		})
	})

	Describe(".GetServicesByNameForSpaceWithPlans", func() {
		BeforeEach(func() {
			serviceRepo.FindServiceOfferingsForSpaceByLabelReturns(
				models.ServiceOfferings([]models.ServiceOffering{service1, v1Service}),
				nil,
			)

			planBuilder.GetPlansForServiceReturns([]models.ServicePlanFields{planWithoutOrgs}, nil)
		})

		It("returns the named service", func() {
			services, err := serviceBuilder.GetServicesByNameForSpaceWithPlans("serviceWithPlans", "spaceGUID")
			Expect(err).NotTo(HaveOccurred())

			Expect(len(services)).To(Equal(2))
			Expect(services[0].Label).To(Equal("my-service1"))
			Expect(services[0].Plans[0].Name).To(Equal("service-plan-without-orgs"))
			Expect(services[0].Plans[0].OrgNames).To(BeNil())
			Expect(services[1].Label).To(Equal("v1Service"))
			Expect(services[1].Plans[0].Name).To(Equal("service-plan-without-orgs"))
			Expect(services[1].Plans[0].OrgNames).To(BeNil())
		})
	})

	Describe(".GetServiceByNameForOrg", func() {
		It("returns the named service, populated with plans", func() {
			service, err := serviceBuilder.GetServiceByNameForOrg("my-service1", "org1")
			Expect(err).NotTo(HaveOccurred())

			Expect(planBuilder.GetPlansForServiceForOrgCallCount()).To(Equal(1))
			servName, orgName := planBuilder.GetPlansForServiceForOrgArgsForCall(0)
			Expect(servName).To(Equal("service-guid1"))
			Expect(orgName).To(Equal("org1"))

			Expect(len(service.Plans)).To(Equal(2))
			Expect(service.Plans[0].Name).To(Equal("service-plan1"))
			Expect(service.Plans[1].Name).To(Equal("service-plan2"))
			Expect(service.Plans[0].OrgNames).To(Equal([]string{"org1", "org2"}))
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
			Expect(service.Plans[0].OrgNames).To(Equal([]string{"org1", "org2"}))
		})
	})

	Describe(".GetServicesForManyBrokers", func() {
		It("returns all the services for an array of broker guids, fully populated", func() {
			brokerGUIDs := []string{"my-service-broker-guid1", "my-service-broker-guid2"}
			services, err := serviceBuilder.GetServicesForManyBrokers(brokerGUIDs)
			Expect(err).NotTo(HaveOccurred())

			Expect(services).To(HaveLen(2))

			broker_service := services[0]
			Expect(broker_service.Label).To(Equal("my-service1"))
			Expect(len(broker_service.Plans)).To(Equal(2))
			Expect(broker_service.Plans[0].Name).To(Equal("service-plan1"))
			Expect(broker_service.Plans[1].Name).To(Equal("service-plan2"))
			Expect(broker_service.Plans[0].OrgNames).To(Equal([]string{"org1", "org2"}))

			broker_service2 := services[1]
			Expect(broker_service2.Label).To(Equal("my-service2"))
			Expect(len(broker_service2.Plans)).To(Equal(1))
			Expect(broker_service2.Plans[0].Name).To(Equal("service-plan3"))
		})

		It("raises errors from the service repo", func() {
			serviceRepo.ListServicesFromManyBrokersReturns([]models.ServiceOffering{}, errors.New("error"))
			brokerGUIDs := []string{"my-service-broker-guid1", "my-service-broker-guid2"}
			_, err := serviceBuilder.GetServicesForManyBrokers(brokerGUIDs)
			Expect(err).To(HaveOccurred())
		})

		It("raises errors from the plan builder", func() {
			planBuilder.GetPlansForManyServicesWithOrgsReturns(nil, errors.New("error"))
			brokerGUIDs := []string{"my-service-broker-guid1", "my-service-broker-guid2"}
			_, err := serviceBuilder.GetServicesForManyBrokers(brokerGUIDs)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe(".GetServiceVisibleToOrg", func() {
		It("Returns a service populated with plans visible to the provided org", func() {
			service, err := serviceBuilder.GetServiceVisibleToOrg("my-service1", "org1")
			Expect(err).NotTo(HaveOccurred())

			Expect(service.Label).To(Equal("my-service1"))
			Expect(len(service.Plans)).To(Equal(2))
			Expect(service.Plans[0].Name).To(Equal("service-plan1"))
			Expect(service.Plans[0].OrgNames).To(Equal([]string{"org1", "org2"}))
		})

		Context("When no plans are visible to the provided org", func() {
			It("Returns nil", func() {
				planBuilder.GetPlansVisibleToOrgReturns(nil, nil)
				service, err := serviceBuilder.GetServiceVisibleToOrg("my-service1", "org3")
				Expect(err).NotTo(HaveOccurred())

				Expect(service).To(Equal(models.ServiceOffering{}))
			})
		})
	})

	Describe(".GetServicesVisibleToOrg", func() {
		It("Returns services with plans visible to the provided org", func() {
			planBuilder.GetPlansVisibleToOrgReturns([]models.ServicePlanFields{plan1, plan2}, nil)
			services, err := serviceBuilder.GetServicesVisibleToOrg("org1")
			Expect(err).NotTo(HaveOccurred())

			service := services[0]
			Expect(service.Label).To(Equal("my-service1"))
			Expect(len(service.Plans)).To(Equal(2))
			Expect(service.Plans[0].Name).To(Equal("service-plan1"))
			Expect(service.Plans[0].OrgNames).To(Equal([]string{"org1", "org2"}))
		})

		Context("When no plans are visible to the provided org", func() {
			It("Returns nil", func() {
				planBuilder.GetPlansVisibleToOrgReturns(nil, nil)
				services, err := serviceBuilder.GetServicesVisibleToOrg("org3")
				Expect(err).NotTo(HaveOccurred())

				Expect(services).To(BeNil())
			})
		})
	})
})
